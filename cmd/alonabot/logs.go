package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/AlonaBot/AlonaBot/internal/core/logging"
	"github.com/spf13/cobra"
)

const logsPath = "/api/v1/logs/"

type logLevelSettings struct {
	Global  string            `json:"global"`
	Sources map[string]string `json:"sources"`
}

func logCommands(code *int) []*cobra.Command {
	options := &managementOptions{}
	var source, level, trace, date string
	var follow bool
	logs := &cobra.Command{Use: "logs", Short: "Read and follow Core logs", Args: cobra.NoArgs}
	options.flags(logs)
	logs.Flags().StringVar(&source, "source", "", "component ID (history: displayed source)")
	logs.Flags().StringVar(&level, "level", "", "minimum record level: debug, info, warn, error or critical")
	logs.Flags().StringVar(&trace, "trace", "", "trace ID (history: first eight digits)")
	logs.Flags().StringVar(&date, "date", "", "log file date: YYYY-MM-DD")
	logs.Flags().BoolVar(&follow, "follow", false, "follow live logs, including the initial snapshot")
	logs.RunE = func(cmd *cobra.Command, _ []string) error {
		if date != "" && follow {
			return errors.New("--date and --follow are mutually exclusive")
		}
		*code = exitFailure
		client, err := options.client()
		if err != nil {
			return err
		}
		query := url.Values{}
		for key, value := range map[string]string{"source": source, "level": level, "trace": trace, "date": date} {
			if value != "" {
				query.Set(key, value)
			}
		}
		endpoint := "recent"
		if date != "" {
			endpoint = "history"
		}
		if follow {
			endpoint = "stream"
		}
		ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer cancel()
		output, err := logging.Open(logging.Options{Format: logging.FormatText, Level: slog.LevelDebug, Stderr: cmd.OutOrStdout(), BufferRecords: 1})
		if err != nil {
			return err
		}
		defer output.Close()
		path := logsPath + endpoint + "?" + query.Encode()
		if follow {
			return followLogs(ctx, client, path, func(record logging.Record) error { return printLogRecord(ctx, output, record) }, cmd.ErrOrStderr())
		}
		raw, err := client.request(ctx, http.MethodGet, path, nil)
		if err != nil {
			return err
		}
		var records []logging.Record
		if err := json.Unmarshal(raw, &records); err != nil {
			return fmt.Errorf("decode logs: %w", err)
		}
		for _, record := range records {
			if date != "" {
				if err := printHistoryRecord(cmd.OutOrStdout(), record); err != nil {
					return err
				}
			} else if err := printLogRecord(ctx, output, record); err != nil {
				return err
			}
		}
		return nil
	}
	levelOptions := &managementOptions{}
	change := &cobra.Command{Use: "log-level [ID] [LEVEL]", Short: "Inspect or update output levels (ID global selects Core)", Args: cobra.MaximumNArgs(2)}
	levelOptions.flags(change)
	change.RunE = func(cmd *cobra.Command, args []string) error {
		*code = exitFailure
		client, err := levelOptions.client()
		if err != nil {
			return err
		}
		method := http.MethodGet
		var body any
		if len(args) == 2 {
			method = http.MethodPut
			if args[0] == "global" {
				body = map[string]any{"global": args[1]}
			} else {
				body = map[string]any{"sources": map[string]string{args[0]: args[1]}}
			}
		}
		raw, err := client.request(cmd.Context(), method, logsPath+"levels", body)
		if err != nil {
			return err
		}
		if len(args) != 1 {
			_, err = fmt.Fprintln(cmd.OutOrStdout(), string(raw))
			return err
		}
		var levels logLevelSettings
		if err := json.Unmarshal(raw, &levels); err != nil {
			return err
		}
		value := levels.Global
		if args[0] != "global" {
			value = levels.Sources[args[0]]
			if value == "" {
				value = "inherit"
			}
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout(), value)
		return err
	}
	return []*cobra.Command{logs, change}
}

func followLogs(ctx context.Context, client *managementClient, path string, print func(logging.Record) error, diagnostic io.Writer) (err error) {
	response, err := client.open(ctx, http.MethodGet, path, nil)
	if err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return err
	}
	defer func() { err = errors.Join(err, response.Body.Close()) }()
	if content, _, _ := strings.Cut(response.Header.Get("Content-Type"), ";"); content != "text/event-stream" {
		return fmt.Errorf("expected text/event-stream, got %q", content)
	}
	reader := bufio.NewReader(response.Body)
	var event string
	var data []string
	for {
		line, readErr := reader.ReadString('\n')
		if readErr != nil {
			if ctx.Err() != nil {
				return nil
			}
			if errors.Is(readErr, io.EOF) {
				return errors.New("log stream closed")
			}
			return fmt.Errorf("read log stream: %w", readErr)
		}
		line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
		if line == "" {
			switch event {
			case "log":
				var record logging.Record
				if err := json.Unmarshal([]byte(strings.Join(data, "\n")), &record); err != nil {
					return fmt.Errorf("decode log stream record: %w", err)
				}
				if err := print(record); err != nil {
					return err
				}
			case "overflow":
				if _, err := fmt.Fprintln(diagnostic, "Log stream overflow; reconnect and read recent logs to catch up."); err != nil {
					return err
				}
				return errors.New("log stream overflow")
			}
			event = ""
			data = nil
			continue
		}
		key, value, _ := strings.Cut(line, ":")
		value = strings.TrimPrefix(value, " ")
		switch key {
		case "event":
			event = value
		case "data":
			data = append(data, value)
		}
	}
}

func printLogRecord(ctx context.Context, output *logging.Output, record logging.Record) error {
	attrs := make([]slog.Attr, 0, len(record.Attrs))
	for key, value := range record.Attrs {
		if key == slog.SourceKey {
			raw, err := json.Marshal(value)
			if err != nil {
				return err
			}
			source := &slog.Source{}
			if err := json.Unmarshal(raw, source); err != nil {
				return err
			}
			value = source
		}
		attrs = append(attrs, slog.Any(key, value))
	}
	var handler slog.Handler
	if record.Source == logging.ComponentCore {
		handler = output.Logger().Handler().WithAttrs(attrs)
	} else {
		handler = output.Component(record.Source, "", attrs...)
	}
	r := slog.NewRecord(record.Time, record.Level, record.Message, 0)
	return handler.Handle(ctx, r)
}

func printHistoryRecord(writer io.Writer, record logging.Record) error {
	tag := "Core"
	switch record.Attrs[logging.KeyComponent] {
	case logging.ComponentPlugin:
		tag = "Plug:" + record.Source
	case logging.ComponentAdapter:
		tag = "Adpt:" + record.Source
	}
	_, err := fmt.Fprintf(writer, "[%s] [%s]%s\n", record.Time.Local().Format("15:04:05.000"), tag, record.Message)
	return err
}
