package manage_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/release"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/manage"
	"github.com/AlonaBot/AlonaBot/internal/core/packages"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/AlonaBot/AlonaBot/internal/tune"
)

// TestReleaseAssetsInstallation 通过模拟 Release 服务、真实状态库与归档验证安装和更新。
func TestReleaseAssetsInstallation(t *testing.T) {
	const repo = "example/webui"
	const id = "alonabot-webui-fixture"
	dir := t.TempDir()
	store, err := state.Open(t.Context(), dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	})
	var mu sync.Mutex
	var publication release.GitHubRelease
	var archive []byte
	var requests []string
	publish := func(version string) {
		t.Helper()
		var out bytes.Buffer
		gz := gzip.NewWriter(&out)
		tarred := tar.NewWriter(gz)
		for name, raw := range map[string]string{
			"plugin.yaml":       fmt.Sprintf("manifest_version: 1\nid: %s\nname: Release fixture\nversion: %s\nsource: github.com/%s\nrole: webui\nos: any\narch: any\nassets: public\nadmin_api_min_major: 1\nadmin_api_max_major: 1\n", id, version, repo),
			"public/index.html": "<!doctype html><title>" + version + "</title>",
		} {
			if err := tarred.WriteHeader(&tar.Header{Name: name, Mode: 0600, Size: int64(len(raw))}); err != nil {
				t.Fatal(err)
			}
			if _, err := tarred.Write([]byte(raw)); err != nil {
				t.Fatal(err)
			}
		}
		if err := errors.Join(tarred.Close(), gz.Close()); err != nil {
			t.Fatal(err)
		}
		name := id + "-" + version + "-any.tar.gz"
		mu.Lock()
		defer mu.Unlock()
		archive = out.Bytes()
		publication = release.GitHubRelease{Tag: "v" + version, URL: "https://github.com/" + repo + "/releases/tag/v" + version,
			Assets: []release.Download{{Name: name, URL: "https://github.com/" + repo + "/releases/download/v" + version + "/" + name, Size: int64(len(archive)), Digest: fmt.Sprintf("sha256:%x", sha256.Sum256(archive))}}}
	}
	client := &http.Client{Transport: releaseTransport(func(r *http.Request) (*http.Response, error) {
		address := r.URL
		if r.URL.Host == "mirror.example" {
			var err error
			address, err = url.Parse(r.URL.Query().Get("url"))
			if err != nil {
				return nil, err
			}
		}
		if address.Scheme != "https" || (address.Host != "api.github.com" && address.Host != "github.com") {
			return nil, fmt.Errorf("unexpected download: %s", r.URL)
		}
		mu.Lock()
		defer mu.Unlock()
		w := httptest.NewRecorder()
		requests = append(requests, address.Path)
		switch address.Path {
		case "/repos/" + repo + "/releases/latest":
			if err := json.NewEncoder(w).Encode(publication); err != nil {
				t.Error(err)
			}
		case "/" + repo + "/releases/download/" + publication.Tag + "/" + publication.Assets[0].Name:
			if _, err := w.Write(archive); err != nil {
				t.Error(err)
			}
		default:
			t.Errorf("unexpected download: %s", r.URL)
			http.NotFound(w, r)
		}
		return w.Result(), nil
	})}
	configuration, err := settings.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	service := manage.New(manage.Options{Store: store, Settings: configuration, Packages: filepath.Join(dir, "packages"),
		Limits: packages.NewLimits(tune.Defaults()), HTTPClient: client, State: func(string) string { return "" }})
	publish("1.0.0")
	request := manage.InstallRequest{Source: repo}
	if _, err := service.Install(t.Context(), request); err == nil {
		t.Fatal("first source needs confirmation")
	} else {
		var confirm *manage.ErrNeedsConfirm
		if !errors.As(err, &confirm) {
			t.Fatal(err)
		}
	}
	request.Confirm = true
	installed, err := service.Install(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}
	if installed.ID != id || installed.Version != "1.0.0" || installed.Source != "github.com/"+repo {
		t.Fatalf("installed: %+v", installed)
	}
	if err := store.SetPublisherKeyID(t.Context(), id, "historical-publisher"); err != nil {
		t.Fatal(err)
	}
	publish("1.1.0")
	mu.Lock()
	archive[len(archive)-1] ^= 1
	mu.Unlock()
	var denied *validate.Error
	prefix := "https://mirror.example/?url="
	if _, err := service.Update(t.Context(), id, manage.InstallRequest{GitHubDownloadPrefix: &prefix}, false); !errors.As(err, &denied) || denied.Code != pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED {
		t.Fatalf("equal-size archive with changed bytes: %v", err)
	}
	mu.Lock()
	archive[len(archive)-1] ^= 1
	mu.Unlock()
	if _, err := service.Update(t.Context(), id, manage.InstallRequest{}, false); err != nil {
		t.Fatal(err)
	}
	updated, err := store.GetInstall(t.Context(), id)
	if err != nil || updated.Version != "1.1.0" || updated.PublisherKeyID != "" {
		t.Fatalf("updated: %+v %v", updated, err)
	}
	if err := store.SetPublisherKeyID(t.Context(), id, "historical-publisher"); err != nil {
		t.Fatal(err)
	}
	if err := service.Uninstall(t.Context(), id, manage.PurgeNone, true); err != nil {
		t.Fatal(err)
	}
	publish("1.2.0")
	if _, err := service.Install(t.Context(), request); err != nil {
		t.Fatal(err)
	}
	reinstalled, err := store.GetInstall(t.Context(), id)
	if err != nil || reinstalled.Version != "1.2.0" || reinstalled.PublisherKeyID != "" || reinstalled.Uninstalled {
		t.Fatalf("reinstalled: %+v %v", reinstalled, err)
	}
	mu.Lock()
	defer mu.Unlock()
	var downloaded []string
	for _, path := range requests {
		if strings.HasSuffix(path, ".tar.gz") {
			downloaded = append(downloaded, filepath.Base(path))
		}
	}
	want := []string{id + "-1.0.0-any.tar.gz", id + "-1.0.0-any.tar.gz", id + "-1.1.0-any.tar.gz", id + "-1.1.0-any.tar.gz", id + "-1.2.0-any.tar.gz"}
	if !reflect.DeepEqual(downloaded, want) {
		t.Fatalf("downloaded: %v", downloaded)
	}
}

type releaseTransport func(*http.Request) (*http.Response, error)

func (t releaseTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return t(request)
}
