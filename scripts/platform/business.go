package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

const (
	telegramPackageID   = "alonabot-adapter-telegram"
	showcasePackageID   = "showcase"
	echoPackageID       = "echo"
	treePackageID       = "platform-tree"
	coreConfigFile      = "core_config.json"
	firstTokenFile      = "admin_token"
	csrfHeader          = "X-Alonabot-CSRF"
	verificationBaseURL = "http://127.0.0.1:7650" // 受控 Host/Origin；实际 Listener 使用独立动态端口。
)

type httpObservation struct {
	Method       string `json:"method"`
	Path         string `json:"path"`
	RequestBody  string `json:"request_body,omitempty"`
	Status       int    `json:"status"`
	ResponseBody string `json:"response_body"`
}

type adminSession struct {
	report  *report
	core    *coreProcess
	base    *url.URL
	cookies []*http.Cookie
	csrf    string
}

func (r *report) business(ctx context.Context, opt options) (err error) {
	// 1. 归档由调用者构建；先读取其中的真实 build-info，再经正式 CLI 安装。
	ids := []string{telegramPackageID, echoPackageID, showcasePackageID, treePackageID}
	archives := make(map[string]string)
	if err := r.step("example_packages", func(_ *stepResult) error {
		for _, id := range ids {
			for _, asset := range r.Assets {
				if strings.HasPrefix(filepath.Base(asset.Path), id+"-") && strings.HasSuffix(asset.Path, "-"+runtime.GOOS+"-"+runtime.GOARCH+".tar.gz") {
					if archives[id] != "" {
						return fmt.Errorf("multiple candidate archives for %s: %s and %s", id, archives[id], asset.Path)
					}
					archives[id] = asset.Path
				}
			}
			if archives[id] == "" {
				return fmt.Errorf("missing current-platform archive for %s in %s", id, opt.packages)
			}
			build, err := packageBuild(archives[id])
			if err != nil {
				return err
			}
			r.Components = append(r.Components, build)
		}
		return nil
	}); err != nil {
		return err
	}
	bot := newBotServer()
	defer func() {
		bot.server.Close()
		r.BotCalls = bot.snapshot()
		for _, call := range r.BotCalls {
			if call.Error != "" {
				err = errors.Join(err, fmt.Errorf("Bot API %s: %s", call.Method, call.Error))
			}
		}
	}()
	config := map[string]any{"base_url": verificationBaseURL, "adapters": map[string]any{"telegram": map[string]any{
		"package": telegramPackageID, "default_instance": testBotName,
		"instances": map[string]any{testBotName: map[string]any{"token": testBotToken, "api_base_url": bot.server.URL, "proxy_mode": "none"}},
	}}}
	raw, err := json.Marshal(config)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(r.DataDir, coreConfigFile), raw, privateFileMode); err != nil {
		return err
	}
	core, startErr := r.startCore(ctx, opt)
	if core != nil {
		defer func() { err = errors.Join(err, core.stop()) }()
	}
	if startErr != nil {
		return startErr
	}
	// 2. 读取非终端首次凭证并登录；凭证本身只在内存与私有文件中保留。
	var token string
	if err := r.step("first_token_show", func(step *stepResult) error {
		path := filepath.Join(r.DataDir, firstTokenFile)
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if runtime.GOOS != "windows" && info.Mode().Perm() != privateFileMode {
			return fmt.Errorf("first token permissions %v, want %o", info.Mode(), privateFileMode)
		}
		argv := []string{opt.core, "admin-token", "show", "--data-dir", r.DataDir}
		command := exec.CommandContext(ctx, argv[0], argv[1:]...)
		command.Dir = r.Directory
		step.Command, step.Directory = argv, command.Dir
		output, err := command.CombinedOutput()
		if command.ProcessState != nil {
			code := command.ProcessState.ExitCode()
			step.ExitCode = &code
		}
		if err != nil {
			step.Output = string(output)
			return fmt.Errorf("show first token: %w", err)
		}
		step.OutputRedacted = true
		token = strings.TrimSpace(string(output))
		fileToken, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if token == "" || token != strings.TrimSpace(string(fileToken)) {
			return errors.New("admin-token show differs from the first token file")
		}
		return nil
	}); err != nil {
		return err
	}
	var settings struct {
		BaseURL string `json:"base_url"`
	}
	raw, err = os.ReadFile(filepath.Join(r.DataDir, coreConfigFile))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, &settings); err != nil {
		return err
	}
	base, err := url.Parse(settings.BaseURL)
	if err != nil {
		return err
	}
	session := &adminSession{report: r, core: core, base: base}
	if _, err := session.request(ctx, "unauthenticated_packages", http.MethodGet, "/api/v1/packages", nil, http.StatusUnauthorized, false); err != nil {
		return err
	}
	raw, err = session.request(ctx, "first_login", http.MethodPost, "/api/auth/login", map[string]string{"token": token}, http.StatusOK, true)
	if err != nil {
		return err
	}
	var login struct {
		CSRFToken string `json:"csrf_token"`
	}
	if err := json.Unmarshal(raw, &login); err != nil {
		return err
	}
	if len(session.cookies) != 1 || login.CSRFToken == "" {
		return errors.New("login did not establish one cookie and a CSRF token")
	}
	session.csrf = login.CSRFToken
	if _, err := os.Stat(filepath.Join(r.DataDir, firstTokenFile)); !errors.Is(err, os.ErrNotExist) {
		return errors.Join(errors.New("first token file remains after login"), err)
	}
	cliToken := filepath.Join(r.Workspace, "cli-token")
	if err := os.WriteFile(cliToken, []byte(token), privateFileMode); err != nil {
		return err
	}
	// 3. 动态监听地址沿用集成测试的 Host 转发器，认证仍由真实 Core 完成。
	target, err := url.Parse("http://" + core.run.Address)
	if err != nil {
		return err
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	director := proxy.Director
	proxy.Director = func(request *http.Request) { director(request); request.Host = base.Host }
	cli := httptest.NewServer(proxy)
	defer cli.Close()
	for _, id := range ids {
		if err := r.command(ctx, "install_"+id, opt.core, "install", archives[id], "--yes", "--admin-url", cli.URL, "--token-file", cliToken); err != nil {
			return err
		}
		if _, err := session.request(ctx, "enable_"+id, http.MethodPost, "/api/v1/packages/enable", map[string]string{"id": id}, http.StatusOK, false); err != nil {
			return err
		}
	}
	raw, err = session.request(ctx, "activated_packages", http.MethodGet, "/api/v1/packages", nil, http.StatusOK, false)
	if err != nil {
		return err
	}
	var installed []struct {
		ID, Active, State string
		Enabled           bool
	}
	if err := json.Unmarshal(raw, &installed); err != nil {
		return err
	}
	for _, id := range ids {
		found := false
		for _, item := range installed {
			if item.ID == id && item.Enabled && item.Active != "" && item.State == "Activated" {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("component %s is not activated: %s", id, raw)
		}
	}
	// 4. 真实 Adapter 的 long poll 入站与 sendMessage 出站覆盖一次 echo 往返。
	if err := r.step("echo_roundtrip", func(step *stepResult) error {
		update, err := json.Marshal(map[string]any{"update_id": testMessageID, "message": map[string]any{
			"message_id": testMessageID, "date": time.Now().Unix(), "text": "/echo " + echoText,
			"from": map[string]any{"id": testUserID, "first_name": "Platform", "is_bot": false},
			"chat": map[string]any{"id": testUserID, "type": "private"},
		}})
		if err != nil {
			return err
		}
		step.Output = string(update)
		ctx, cancel := context.WithTimeout(ctx, observationDeadline)
		defer cancel()
		select {
		case bot.updates <- update:
		case <-core.done:
			return errors.Join(errors.New("Core exited before echo input"), core.err)
		case <-ctx.Done():
			return ctx.Err()
		}
		select {
		case reply := <-bot.replies:
			raw, err := json.Marshal(reply)
			if err != nil {
				return err
			}
			step.Output += "\n" + string(raw)
			if reply["text"] != echoText || fmt.Sprint(reply["chat_id"]) != fmt.Sprint(testUserID) {
				return fmt.Errorf("unexpected echo reply: %s", raw)
			}
			return nil
		case <-core.done:
			return errors.Join(errors.New("Core exited before echo reply"), core.err)
		case <-ctx.Done():
			return ctx.Err()
		}
	}); err != nil {
		return err
	}
	if raw, err := session.request(ctx, "showcase_status", http.MethodGet, "/plugins/showcase/status", nil, http.StatusOK, false); err != nil {
		return err
	} else if string(raw) != "ok" {
		return fmt.Errorf("unexpected showcase status: %s", raw)
	}
	if raw, err := session.request(ctx, "showcase_page", http.MethodGet, "/plugins/showcase/page", nil, http.StatusOK, false); err != nil {
		return err
	} else if !bytes.Contains(raw, []byte("window.AlonaBot")) {
		return errors.New("showcase page is missing the SDK bridge")
	}
	if raw, err := session.request(ctx, "showcase_bridge_script", http.MethodGet, "/plugins/showcase/kit.js", nil, http.StatusOK, false); err != nil {
		return err
	} else if !bytes.Contains(raw, []byte("postMessage")) {
		return errors.New("showcase bridge script is missing postMessage")
	}
	return r.lifecycle(ctx, opt, session, ids)
}

func (s *adminSession) request(ctx context.Context, name, method, path string, input any, expected int, credential bool) ([]byte, error) {
	var result []byte
	err := s.report.step(name, func(step *stepResult) error {
		ctx, cancel := context.WithTimeout(ctx, observationDeadline)
		defer cancel()
		var body []byte
		if input != nil {
			var err error
			body, err = json.Marshal(input)
			if err != nil {
				return err
			}
		}
		evidence := &httpObservation{Method: method, Path: path}
		step.HTTP = evidence
		step.OutputRedacted = credential
		if !credential {
			evidence.RequestBody = string(body)
		}
		request, err := http.NewRequestWithContext(ctx, method, "http://"+s.core.run.Address+path, bytes.NewReader(body))
		if err != nil {
			return err
		}
		request.Host = s.base.Host
		request.Header.Set("Origin", s.base.String())
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set(csrfHeader, s.csrf)
		for _, cookie := range s.cookies {
			request.AddCookie(cookie)
		}
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			return err
		}
		evidence.Status = response.StatusCode
		result, err = io.ReadAll(response.Body)
		if !credential || response.StatusCode != expected {
			evidence.ResponseBody = string(result)
		}
		if err := errors.Join(err, response.Body.Close()); err != nil {
			return err
		}
		if response.StatusCode != expected {
			return fmt.Errorf("%s %s: got HTTP %d, want %d", method, path, response.StatusCode, expected)
		}
		if credential {
			s.cookies = response.Cookies()
		}
		return nil
	})
	return result, err
}

func packageBuild(path string) (result candidate, err error) {
	file, err := os.Open(path)
	if err != nil {
		return result, err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	compressed, err := gzip.NewReader(file)
	if err != nil {
		return result, err
	}
	defer func() { err = errors.Join(err, compressed.Close()) }()
	archive := tar.NewReader(compressed)
	for {
		header, err := archive.Next()
		if errors.Is(err, io.EOF) {
			return result, fmt.Errorf("missing build-info in %s", path)
		}
		if err != nil {
			return result, err
		}
		if header.Name != "build-info" {
			continue
		}
		raw, err := io.ReadAll(archive)
		if err != nil {
			return result, err
		}
		build, err := debug.ParseBuildInfo(string(raw))
		if err != nil {
			return result, err
		}
		result = candidate{Path: path, GoVersion: build.GoVersion, Program: build.Path, Module: build.Main, Settings: build.Settings}
		// ParseBuildInfo 保留模块与设置，Go 版本需从 String 输出的首行单独读取。
		firstLine, _, _ := strings.Cut(string(raw), "\n")
		if version, ok := strings.CutPrefix(firstLine, "go\t"); ok {
			result.GoVersion = version
		} else {
			return result, fmt.Errorf("missing Go version in %s build-info", path)
		}
		for _, setting := range build.Settings {
			switch setting.Key {
			case "GOOS":
				result.OS = setting.Value
			case "GOARCH":
				result.Arch = setting.Value
			}
		}
		if result.OS != runtime.GOOS || result.Arch != runtime.GOARCH {
			return result, fmt.Errorf("component %s target %s/%s differs from verifier %s/%s", path, result.OS, result.Arch, runtime.GOOS, runtime.GOARCH)
		}
		return result, nil
	}
}
