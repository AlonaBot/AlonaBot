package manage

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strconv"
	"strings"
)

type dataColumn struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	PK      bool   `json:"pk"`
	NotNull bool   `json:"notnull"`
	hidden  int
}

type dataTable struct {
	Name     string       `json:"name"`
	Rows     int64        `json:"rows"`
	Writable bool         `json:"writable"`
	Columns  []dataColumn `json:"columns"`
	rowid    string       // 选择未被真实列遮蔽的 SQLite 隐含主键名。
}

func dataQuote(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}

func dataSchema(r *http.Request, tx *sql.Tx, readonly bool, name string) ([]dataTable, error) {
	rows, err := tx.QueryContext(r.Context(), `SELECT name, type, wr FROM pragma_table_list WHERE schema='main' AND substr(name,1,7) != 'sqlite_' AND (? = '' OR name = ?) ORDER BY name`, name, name)
	if err != nil {
		return nil, err
	}
	tables := make([]dataTable, 0)
	for rows.Next() {
		var table dataTable
		var category string
		var withoutRowID int
		if err := rows.Scan(&table.Name, &category, &withoutRowID); err != nil {
			return nil, closeDataRows(rows, err)
		}
		table.Writable = !readonly && category == "table" && withoutRowID == 0 && !strings.HasPrefix(strings.ToLower(table.Name), "_alonabot_")
		if category == "table" && withoutRowID == 0 {
			table.rowid = "rowid"
		}
		tables = append(tables, table)
	}
	if err := closeDataRows(rows, rows.Err()); err != nil {
		return nil, err
	}
	for i := range tables {
		table := &tables[i]
		columns, err := tx.QueryContext(r.Context(), `SELECT name, type, pk != 0, "notnull", hidden FROM pragma_table_xinfo(?) ORDER BY cid`, table.Name)
		if err != nil {
			return nil, err
		}
		table.Columns = make([]dataColumn, 0)
		names := make([]string, 0)
		for columns.Next() {
			var column dataColumn
			if err := columns.Scan(&column.Name, &column.Type, &column.PK, &column.NotNull, &column.hidden); err != nil {
				return nil, closeDataRows(columns, err)
			}
			table.Columns = append(table.Columns, column)
			names = append(names, strings.ToLower(column.Name))
		}
		if err := closeDataRows(columns, columns.Err()); err != nil {
			return nil, err
		}
		if table.rowid != "" {
			table.rowid = ""
			for _, alias := range []string{"rowid", "_rowid_", "oid"} {
				if !slices.Contains(names, alias) {
					table.rowid = alias
					break
				}
			}
			if table.rowid == "" {
				table.Writable = false
			}
		}
	}
	return tables, nil
}

func closeDataRows(rows *sql.Rows, err error) error {
	return errors.Join(err, rows.Close())
}

func (s *Service) dataTables(r *http.Request, selected *dataSelection) (any, error) {
	return dataTransaction(r, selected, false, func(tx *sql.Tx) (any, error) {
		tables, err := dataSchema(r, tx, selected.readonly, "")
		if err != nil {
			return nil, err
		}
		for i := range tables {
			if err := tx.QueryRowContext(r.Context(), "SELECT count(*) FROM "+dataQuote(tables[i].Name)).Scan(&tables[i].Rows); err != nil {
				return nil, err
			}
		}
		return struct {
			Tables []dataTable `json:"tables"`
		}{tables}, err
	})
}

func (s *Service) dataRows(r *http.Request, selected *dataSelection) (any, error) {
	write := r.Method != "GET" && r.Method != "HEAD"
	return dataTransaction(r, selected, write, func(tx *sql.Tx) (any, error) {
		tables, err := dataSchema(r, tx, selected.readonly, r.PathValue("table"))
		if err != nil {
			return nil, err
		}
		for _, table := range tables {
			if table.Name != r.PathValue("table") {
				continue
			}
			if !write {
				return dataReadRows(r, tx, table)
			}
			if !table.Writable {
				return nil, denied("table is read-only")
			}
			result, rowid, err := dataWriteRow(r, tx, table)
			if err == nil {
				selected.audit = []any{"plugin_id", selected.id, "table", table.Name, "rowid", rowid}
			}
			return result, err
		}
		return nil, sql.ErrNoRows
	})
}

func dataReadRows(r *http.Request, tx *sql.Tx, table dataTable) (any, error) {
	offset, limit, err := dataPage(r.URL.Query())
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(table.Columns))
	for _, column := range table.Columns {
		if column.hidden != 1 {
			names = append(names, column.Name)
		}
	}
	query := r.URL.Query()
	var conditions []string
	var arguments []any
	for _, key := range slices.Sorted(maps.Keys(query)) {
		if !strings.HasPrefix(key, "where.") {
			continue
		}
		column := strings.TrimPrefix(key, "where.")
		if !slices.Contains(names, column) {
			return nil, invalid("unknown column: " + column)
		}
		for _, value := range query[key] {
			conditions = append(conditions, dataQuote(column)+" = ?")
			arguments = append(arguments, value)
		}
	}
	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}
	var total int64
	if err := tx.QueryRowContext(r.Context(), "SELECT count(*) FROM "+dataQuote(table.Name)+where, arguments...).Scan(&total); err != nil {
		return nil, err
	}
	order := ""
	if column := query.Get("order"); column != "" {
		if !slices.Contains(names, column) {
			return nil, invalid("unknown order column: " + column)
		}
		order = " ORDER BY " + dataQuote(column)
		if query.Get("desc") == "1" {
			order += " DESC"
		}
	} else if table.rowid != "" {
		order = " ORDER BY " + dataQuote(table.rowid)
	}
	columns := make([]string, 0, len(names)+1)
	if table.rowid != "" {
		columns = append(columns, dataQuote(table.rowid))
	} else {
		columns = append(columns, "NULL")
	}
	for _, name := range names {
		columns = append(columns, dataQuote(name))
	}
	arguments = append(arguments, limit, offset)
	rows, err := tx.QueryContext(r.Context(), "SELECT "+strings.Join(columns, ",")+" FROM "+dataQuote(table.Name)+where+order+" LIMIT ? OFFSET ?", arguments...)
	if err != nil {
		return nil, err
	}
	type row struct {
		RowID  any            `json:"rowid"`
		Values map[string]any `json:"values"`
	}
	result := make([]row, 0)
	for rows.Next() {
		values := make([]any, len(names)+1)
		destinations := make([]any, len(values))
		for i := range values {
			destinations[i] = &values[i]
		}
		if err := rows.Scan(destinations...); err != nil {
			return nil, closeDataRows(rows, err)
		}
		entry := row{RowID: values[0], Values: make(map[string]any, len(names))}
		for i, name := range names {
			value := values[i+1]
			if blob, ok := value.([]byte); ok {
				value = map[string]string{"$base64": base64.StdEncoding.EncodeToString(blob)}
			}
			entry.Values[name] = value
		}
		result = append(result, entry)
	}
	return struct {
		Rows  []row `json:"rows"`
		Total int64 `json:"total"`
	}{result, total}, closeDataRows(rows, rows.Err())
}

func dataWriteRow(r *http.Request, tx *sql.Tx, table dataTable) (any, int64, error) {
	var rowid int64
	var err error
	if r.Method != "POST" {
		rowid, err = strconv.ParseInt(r.PathValue("rowid"), 10, 64)
		if err != nil {
			return nil, 0, invalid("invalid rowid")
		}
	}
	if r.Method == "DELETE" {
		result, err := tx.ExecContext(r.Context(), "DELETE FROM "+dataQuote(table.Name)+" WHERE "+dataQuote(table.rowid)+" = ?", rowid)
		return struct{}{}, rowid, dataAffected(result, err)
	}
	var request struct {
		Values map[string]json.RawMessage `json:"values"`
	}
	if err := decode(r, &request); err != nil {
		return nil, 0, err
	}
	if request.Values == nil {
		return nil, 0, invalid("values object is required")
	}
	names := slices.Sorted(maps.Keys(request.Values))
	var columns []string
	var arguments []any
	for _, name := range names {
		index := slices.IndexFunc(table.Columns, func(c dataColumn) bool { return c.Name == name && c.hidden == 0 })
		if index < 0 {
			return nil, 0, invalid("unknown or generated column: " + name)
		}
		value, err := dataValue(request.Values[name])
		if err != nil {
			return nil, 0, err
		}
		columns = append(columns, dataQuote(name))
		arguments = append(arguments, value)
	}
	if r.Method == "POST" {
		statement := "INSERT INTO " + dataQuote(table.Name) + " DEFAULT VALUES"
		if len(columns) > 0 {
			statement = "INSERT INTO " + dataQuote(table.Name) + " (" + strings.Join(columns, ",") + ") VALUES (" + strings.TrimSuffix(strings.Repeat("?,", len(columns)), ",") + ")"
		}
		err := tx.QueryRowContext(r.Context(), statement+" RETURNING "+dataQuote(table.rowid), arguments...).Scan(&rowid)
		return struct {
			RowID int64 `json:"rowid"`
		}{rowid}, rowid, err
	}
	if len(columns) == 0 {
		return nil, 0, invalid("at least one column is required")
	}
	for i := range columns {
		columns[i] += " = ?"
	}
	arguments = append(arguments, rowid)
	result, err := tx.ExecContext(r.Context(), "UPDATE "+dataQuote(table.Name)+" SET "+strings.Join(columns, ",")+" WHERE "+dataQuote(table.rowid)+" = ?", arguments...)
	return struct{}{}, rowid, dataAffected(result, err)
}

func dataAffected(result sql.Result, err error) error {
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func dataValue(raw json.RawMessage) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, invalid(err.Error())
	}
	switch v := value.(type) {
	case map[string]any:
		encoded, ok := v["$base64"].(string)
		if !ok || len(v) != 1 {
			return nil, invalid("BLOB must contain only $base64")
		}
		blob, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, invalid(err.Error())
		}
		return blob, nil
	case json.Number:
		if integer, err := v.Int64(); err == nil {
			return integer, nil
		}
		number, err := v.Float64()
		if err != nil {
			return nil, invalid(err.Error())
		}
		return number, nil
	case []any:
		return nil, invalid("array is not a SQLite value")
	case nil, string, bool:
		return value, nil
	default:
		return nil, fmt.Errorf("unsupported SQLite value %T", value)
	}
}
