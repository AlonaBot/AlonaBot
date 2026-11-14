//go:build unix

package alonabot

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/AlonaBot/AlonaBot/contracts/release"
	"github.com/AlonaBot/AlonaBot/internal/core/manage"
)

// TestAdminSessionAndUploads 经真实 Core、SQLite、签名插件进程与文件系统验证浏览器管理流程。
func TestAdminSessionAndUploads(t *testing.T) {
	h := newWebUIHarness(t)
	h.start(nil)
	h.expect("GET", "/api/auth/session", nil, false, 401, "unauthorized")
	h.expect("GET", "/api/v1/core", nil, false, 401, "unauthorized")
	login, raw := adminRequest(t, h, "POST", "/api/auth/login", "application/json", strings.NewReader(`{"token":"`+h.token+`"}`), nil, "")
	var logged struct {
		CSRF string `json:"csrf_token"`
	}
	if login.StatusCode != 200 {
		t.Fatalf("login=%d %s", login.StatusCode, raw)
	}
	if err := json.Unmarshal(raw, &logged); err != nil {
		t.Fatal(err)
	}
	cookies := login.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies=%v", cookies)
	}
	cookie := cookies[0]
	saved, err := h.core.store.AdminSession(context.Background(), cookie.Value)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		response, raw := adminRequest(t, h, "GET", "/api/auth/session", "", nil, cookie, "")
		var session struct {
			CSRF      string `json:"csrf_token"`
			ExpiresAt int64  `json:"expires_at_unix_ms"`
		}
		if err := json.Unmarshal(raw, &session); err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != 200 || session.CSRF != logged.CSRF || session.ExpiresAt != saved.ExpiresAt || response.Header.Get("Cache-Control") != "no-store" || len(response.Cookies()) != 0 {
			t.Fatalf("session=%d %s %v", response.StatusCode, raw, response.Header)
		}
		current, err := h.core.store.AdminSession(context.Background(), cookie.Value)
		if err != nil || current != saved {
			t.Fatalf("session changed: %+v %v", current, err)
		}
	}
	response, raw := adminRequest(t, h, "GET", "/api/v1/core", "", nil, cookie, "")
	var core map[string]any
	if err := json.Unmarshal(raw, &core); err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != 200 || len(core) != 3 || core["listen"] != h.core.http.Addr || strings.HasSuffix(core["listen"].(string), ":0") || core["version"] != coreVersion() || core["admin_api_major"] != float64(1) {
		t.Fatalf("core=%d %s", response.StatusCode, raw)
	}
	_, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	first := adminSignedBundle(t, h, "1.0.0", key)
	upload := func(action string, request manage.InstallRequest, omit string, want int, contains string) {
		t.Helper()
		body, content := adminMultipart(t, request, action == "update", omit)
		response, raw := adminRequest(t, h, "POST", "/api/v1/packages/"+action, content, body, cookie, logged.CSRF)
		if response.StatusCode != want || !strings.Contains(string(raw), contains) || (action == "update" && want == 200 && strings.TrimSpace(string(raw)) != "{}") {
			t.Fatalf("%s=%d %s, want %d %q", action, response.StatusCode, raw, want, contains)
		}
		adminUploadsClean(t, h)
	}
	first.Confirm = false
	upload("install", first, "", 409, "trust publisher")
	first.Confirm = true
	upload("install", first, "", 200, `"ID":"upload_demo"`)
	h.action("enable", "upload_demo", 200)
	before := adminInstalled(t, h)
	if before.State != "Activated" || !before.Enabled || !before.Signed || before.Version != "1.0.0" {
		t.Fatalf("installed=%+v", before)
	}
	plugin, ok := h.core.assembly.router.Plugin("upload_demo")
	if !ok || !plugin.Activated() {
		t.Fatal("real plugin process was not activated")
	}
	dataPath := filepath.Join(h.dir, "plugins", "upload_demo", "upload-sentinel")
	if err := os.WriteFile(dataPath, []byte("keep original user data"), 0o600); err != nil {
		t.Fatal(err)
	}
	original, err := h.core.store.GetInstall(context.Background(), "upload_demo")
	if err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", filepath.Join(h.dir, "plugins", "upload_demo", "generations", before.Active, "db.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	_, writeErr := db.Exec("UPDATE upload_data SET value = 'keep original database data'")
	if closeErr := db.Close(); writeErr != nil || closeErr != nil {
		t.Fatalf("write user database: %v close: %v", writeErr, closeErr)
	}
	preserved := func() {
		t.Helper()
		after := adminInstalled(t, h)
		current, err := h.core.store.GetInstall(context.Background(), "upload_demo")
		if err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(dataPath)
		if err != nil {
			t.Fatal(err)
		}
		same, ok := h.core.assembly.router.Plugin("upload_demo")
		if after != before || !reflect.DeepEqual(current, original) || string(data) != "keep original user data" || !ok || same != plugin || !same.Activated() {
			t.Fatalf("old component changed: before=%+v after=%+v data=%q", before, after, data)
		}
		adminDatabase(t, h, before.Active, "1.0.0")
	}
	next := adminSignedBundle(t, h, "1.1.0", key)
	upload("update", next, "release", 400, "INVALID_ARGUMENT")
	preserved()
	bad := next
	bad.Signature = base64.StdEncoding.EncodeToString(make([]byte, ed25519.SignatureSize))
	upload("update", bad, "", 403, "PERMISSION_DENIED")
	preserved()
	upload("update", next, "", 200, "{}")
	after := adminInstalled(t, h)
	updated, ok := h.core.assembly.router.Plugin("upload_demo")
	if after.Version != "1.1.0" || after.State != "Activated" || !after.Enabled || !after.Signed || after.Active == before.Active || !ok || !updated.Activated() {
		t.Fatalf("updated component=%+v activated=%v", after, ok)
	}
	data, err := os.ReadFile(dataPath)
	if err != nil || string(data) != "keep original user data" {
		t.Fatalf("updated user data=%q err=%v", data, err)
	}
	// 更新迁移只写候选副本，旧代次的数据库保留原值。
	adminDatabase(t, h, before.Active, "1.0.0")
	adminDatabase(t, h, after.Active, "1.1.0")
	t.Log("V-40/V-47 API: real Core Listener, SQLite, signed plugin processes; 409 replay, rejected uploads preserve old state, successful 1.1.0 update preserves user data through a migrated database copy, upload cleanup passed")
}

func adminDatabase(t *testing.T, h *webUIHarness, generation, version string) {
	t.Helper()
	db, err := sql.Open("sqlite", filepath.Join(h.dir, "plugins", "upload_demo", "generations", generation, "db.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	var data, migrated string
	readErr := db.QueryRow("SELECT value, version FROM upload_data").Scan(&data, &migrated)
	if closeErr := db.Close(); readErr != nil || closeErr != nil || data != "keep original database data" || migrated != version {
		t.Fatalf("database generation=%s data=%q version=%q want=%q read=%v close=%v", generation, data, migrated, version, readErr, closeErr)
	}
}

func adminRequest(t *testing.T, h *webUIHarness, method, path, content string, body io.Reader, cookie *http.Cookie, csrf string) (*http.Response, []byte) {
	t.Helper()
	r, err := http.NewRequest(method, "http://"+h.core.http.Addr+path, body)
	if err != nil {
		t.Fatal(err)
	}
	if content != "" {
		r.Header.Set("Content-Type", content)
	}
	if cookie != nil {
		r.AddCookie(cookie)
	}
	r.Header.Set("Origin", h.origin)
	if csrf != "" {
		r.Header.Set("X-Alonabot-CSRF", csrf)
	}
	response, err := h.client.Do(r)
	if err != nil {
		t.Fatal(err)
	}
	raw, readErr := io.ReadAll(response.Body)
	if err := response.Body.Close(); err != nil {
		t.Fatal(err)
	}
	if readErr != nil {
		t.Fatal(readErr)
	}
	return response, raw
}

func adminMultipart(t *testing.T, request manage.InstallRequest, update bool, omit string) (*bytes.Buffer, string) {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for _, file := range []struct{ name, filename, content string }{{"archive", "../../client-name.tar.gz", ""}, {"release", "release.json", request.ReleaseJSON}, {"signature", "release.json.sig", request.Signature}} {
		if file.name == omit || (file.name == "signature" && file.content == "") {
			continue
		}
		out, err := writer.CreateFormFile(file.name, file.filename)
		if err != nil {
			t.Fatal(err)
		}
		if file.name == "archive" {
			in, err := os.Open(request.Archive)
			if err != nil {
				t.Fatal(err)
			}
			_, copyErr := io.Copy(out, in)
			closeErr := in.Close()
			if copyErr != nil || closeErr != nil {
				t.Fatalf("copy archive: %v %v", copyErr, closeErr)
			}
		} else if _, err := io.WriteString(out, file.content); err != nil {
			t.Fatal(err)
		}
	}
	fields := map[string]string{"public_key": request.PublicKey, "confirm": fmt.Sprint(request.Confirm)}
	if update {
		fields["id"] = "upload_demo"
	} else {
		fields["replace"] = fmt.Sprint(request.Replace)
	}
	for name, value := range fields {
		if name != omit {
			if err := writer.WriteField(name, value); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return &body, writer.FormDataContentType()
}

func adminUploadsClean(t *testing.T, h *webUIHarness) {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(h.dir, "staging"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "upload-") {
			t.Fatalf("upload leaked: %s", entry.Name())
		}
	}
}

func adminInstalled(t *testing.T, h *webUIHarness) manage.Package {
	t.Helper()
	_, raw := h.expect("GET", "/api/v1/packages", nil, true, 200, "upload_demo")
	var items []manage.Package
	if err := json.Unmarshal(raw, &items); err != nil {
		t.Fatal(err)
	}
	for _, item := range items {
		if item.ID == "upload_demo" {
			return item
		}
	}
	t.Fatal("upload_demo missing from installed packages")
	return manage.Package{}
}

func adminSignedBundle(t *testing.T, h *webUIHarness, version string, key ed25519.PrivateKey) manage.InstallRequest {
	t.Helper()
	// 复用 WebUI 测试的归档/清单装配，只替换真实插件二进制、版本和签名。
	request := h.bundle("upload_demo", "plugin", "", 1)
	project := t.TempDir()
	source := `package main
import("context";"database/sql";"fmt";"log/slog";"os";"github.com/AlonaBot/AlonaBot/sdk";"github.com/AlonaBot/AlonaBot/sdk/sqlitedb";"github.com/AlonaBot/AlonaBot/sdk/genimpl")
func init(){ genimpl.Register("main",genimpl.Resources{Manifest:[]byte(fmt.Sprintf("id: upload_demo\nname: WebUI Fixture\nversion: %s\nsource: github.com/example/upload_demo\n",` + fmt.Sprintf("%q", version) + `))}) }
func New()*sdk.Plugin{
 p:=sdk.NewPlugin()
 sqlitedb.Migrate(p,func(ctx context.Context,db *sql.DB)error{
  if _,err:=db.ExecContext(ctx,"CREATE TABLE IF NOT EXISTS upload_data (value TEXT, version TEXT); INSERT INTO upload_data SELECT 'initial database data', '' WHERE NOT EXISTS (SELECT 1 FROM upload_data)");err!=nil{return err}
  _,err:=db.ExecContext(ctx,"UPDATE upload_data SET version = ?",` + fmt.Sprintf("%q", version) + `)
  return err
 })
 return p
}
func main(){if err:=sdk.Run(New());err!=nil{slog.Error("fixture failed","error",err);os.Exit(1)}}
`
	name := filepath.Join(project, "main.go")
	if err := os.WriteFile(name, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	executable := filepath.Join(project, "component")
	if output, err := exec.Command("go", "build", "-o", executable, name).CombinedOutput(); err != nil {
		t.Fatalf("build upload plugin: %v\n%s", err, output)
	}
	binary, err := os.ReadFile(executable)
	if err != nil {
		t.Fatal(err)
	}
	input, err := os.Open(request.Archive)
	if err != nil {
		t.Fatal(err)
	}
	gz, err := gzip.NewReader(input)
	if err != nil {
		t.Fatal(err)
	}
	reader := tar.NewReader(gz)
	var archive bytes.Buffer
	output := gzip.NewWriter(&archive)
	writer := tar.NewWriter(output)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		raw, err := io.ReadAll(reader)
		if err != nil {
			t.Fatal(err)
		}
		if header.Name == "plugin.yaml" {
			raw = bytes.ReplaceAll(raw, []byte("version: 1.0.0\n"), []byte("version: "+version+"\n"))
			raw = bytes.ReplaceAll(raw, []byte("migration_mode: none\nprepare_effects: none\n"), []byte("migration_mode: self\nprepare_effects: managed_db\n"))
		}
		if header.Name == "bin/component" {
			raw = binary
			header.Mode = 0o700
		}
		header.Size = int64(len(raw))
		if err := writer.WriteHeader(header); err != nil {
			t.Fatal(err)
		}
		if _, err := writer.Write(raw); err != nil {
			t.Fatal(err)
		}
	}
	for _, err := range []error{gz.Close(), input.Close(), writer.Close(), output.Close()} {
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(request.Archive, archive.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	// release 契约要求整个归档的摘要，仅计算安装验签所需的一份。
	digest := sha256.Sum256(archive.Bytes())
	var manifest release.Release
	if err := json.Unmarshal([]byte(request.ReleaseJSON), &manifest); err != nil {
		t.Fatal(err)
	}
	public := key.Public().(ed25519.PublicKey)
	manifest.Version = version
	manifest.PublisherKeyID = release.KeyID(public)
	manifest.Assets[0].Size = int64(archive.Len())
	manifest.Assets[0].SHA256 = hex.EncodeToString(digest[:])
	raw, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	raw = append(raw, '\n')
	request.ReleaseJSON = string(raw)
	request.Signature = base64.StdEncoding.EncodeToString(ed25519.Sign(key, raw)) + "\n"
	request.PublicKey = base64.StdEncoding.EncodeToString(public) + "\n"
	return request
}
