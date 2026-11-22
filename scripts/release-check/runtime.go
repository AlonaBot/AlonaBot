package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// runtimeHost 持有唯一候选进程；done 关闭后才能读取 waitErr 和完整日志。
type runtimeHost struct {
	ctx                                          context.Context
	binary, cwd, data, evidence, adminURL, token string
	command                                      *exec.Cmd
	done                                         chan struct{}
	waitErr                                      error
	client                                       *http.Client
	sequence                                     int
}

// cliEvidenceError 区分证据保存失败，防止预期的命令拒绝掩盖它。
type cliEvidenceError struct{ err error }

func (e *cliEvidenceError) Error() string { return e.err.Error() }
func (e *cliEvidenceError) Unwrap() error { return e.err }

func startRuntime(ctx context.Context, binary, cwd, data, evidence string, bundle bool) (*runtimeHost, error) {
	h := &runtimeHost{ctx: ctx, binary: binary, cwd: cwd, data: data, evidence: evidence, done: make(chan struct{})}
	// 环境只保留平台运行所需变量，排除调用者的 token、调参及开发注入。
	h.command = exec.Command(binary, "run", "--data-dir", data, "--listen", "127.0.0.1:0", "--console=false", "--log-format", "json")
	h.command.Dir, h.command.Env = cwd, runtimeEnvironment()
	prepareRuntime(h.command)
	output, err := os.OpenFile(filepath.Join(evidence, "core.stdout.log"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, fileMode)
	if err != nil {
		return nil, err
	}
	log, err := os.OpenFile(filepath.Join(evidence, "core.stderr.log"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, fileMode)
	if err != nil {
		return nil, errors.Join(err, output.Close())
	}
	h.command.Stdout = output
	stderr, err := h.command.StderrPipe()
	if err != nil {
		return nil, errors.Join(err, output.Close(), log.Close())
	}
	if err := h.command.Start(); err != nil {
		return nil, errors.Join(err, stderr.Close(), output.Close(), log.Close())
	}
	ready := make(chan string, 1)
	go func() {
		reader := bufio.NewReader(io.TeeReader(stderr, log))
		var address string
		var readErr error
		started, announced := false, false
		activated := map[string]bool{}
		for {
			line, err := reader.ReadBytes('\n')
			var event struct {
				Message  string `json:"msg"`
				Listen   string `json:"listen"`
				PluginID string `json:"plugin_id"`
			}
			if json.Unmarshal(line, &event) == nil {
				if event.Message == "HTTP listener started" {
					address = event.Listen
				}
				if event.Message == "Core started" {
					started = true
				}
				if event.Message == "plugin activated" {
					activated[event.PluginID] = true
				}
				// 整合包的 HTTP 监听先于 WebUI 握手；Core 模式启动后即可读取快照。
				if !announced && started && (!bundle || activated["alonabot-webui-admin"]) {
					ready <- address
					announced = true
				}
			}
			if err != nil {
				if !errors.Is(err, io.EOF) {
					readErr = err
				}
				break
			}
		}
		h.waitErr = errors.Join(readErr, h.command.Wait(), output.Close(), log.Close())
		close(h.done)
	}()
	select {
	case address := <-ready:
		if _, _, err := net.SplitHostPort(address); err != nil {
			return h, err
		}
		h.adminURL = "http://" + address
		return h, nil
	case <-h.done:
		return h, fmt.Errorf("Core exited before readiness: %w", errors.Join(errors.New("missing Core started event"), h.waitErr))
	case <-ctx.Done():
		return h, ctx.Err()
	}
}

func runtimeEnvironment() []string {
	var result []string
	for _, name := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "SystemRoot", "WINDIR"} {
		if value, ok := os.LookupEnv(name); ok {
			result = append(result, name+"="+value)
		}
	}
	return result
}

func (h *runtimeHost) stop() error {
	select {
	case <-h.done:
		return errors.Join(errors.New("Core exited before requested shutdown"), h.waitErr)
	default:
	}
	signalErr := interruptRuntime(h.command.Process)
	if signalErr != nil {
		killErr := h.command.Process.Kill()
		<-h.done
		return errors.Join(signalErr, killErr, h.waitErr)
	}
	select {
	case <-h.done:
		return errors.Join(signalErr, h.waitErr)
	case <-h.ctx.Done():
		killErr := h.command.Process.Kill()
		<-h.done
		return errors.Join(signalErr, h.ctx.Err(), killErr, h.waitErr)
	}
}

// authenticate 只在内存接收 show 输出；证据仅记录权限和成功状态。
func (h *runtimeHost) authenticate() error {
	tokenFile := filepath.Join(h.data, "admin_token")
	if err := checkTokenPermissions(tokenFile); err != nil {
		return err
	}
	command := exec.CommandContext(h.ctx, h.binary, "admin-token", "show", "--data-dir", h.data)
	command.Dir, command.Env = h.cwd, runtimeEnvironment()
	var stderr bytes.Buffer
	command.Stderr = &stderr
	raw, err := command.Output()
	if err != nil {
		return fmt.Errorf("admin-token show: %w: %s", err, stderr.String())
	}
	h.token = strings.TrimSpace(string(raw))
	if h.token == "" {
		return errors.New("admin-token show returned an empty token")
	}
	stored, err := os.ReadFile(tokenFile)
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(stored)) != h.token {
		return errors.New("admin-token show differs from the first token file")
	}
	return nil
}

// serveManagement 复用候选的随机监听端口，Host 与 Origin 仍遵守实际 base_url。
func (h *runtimeHost) serveManagement() (func() error, error) {
	raw, err := os.ReadFile(filepath.Join(h.data, "core_config.json"))
	if err != nil {
		return nil, err
	}
	var settings struct {
		BaseURL string `json:"base_url"`
	}
	if err := json.Unmarshal(raw, &settings); err != nil {
		return nil, err
	}
	base, err := url.Parse(settings.BaseURL)
	if err != nil {
		return nil, err
	}
	if base.Scheme != "http" || base.Host == "" {
		return nil, fmt.Errorf("unexpected initial base_url %q", settings.BaseURL)
	}
	target, err := url.Parse(h.adminURL)
	if err != nil {
		return nil, err
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = &http.Transport{Proxy: nil}
	director := proxy.Director
	proxy.Director = func(r *http.Request) { director(r); r.Host = base.Host; r.Header.Set("Origin", base.String()) }
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	server := &http.Server{Handler: proxy}
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	h.adminURL = "http://" + listener.Addr().String()
	transport := &http.Transport{Proxy: nil}
	h.client = &http.Client{Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	return func() error {
		closeErr := server.Close()
		serveErr := <-done
		if errors.Is(serveErr, http.ErrServerClosed) {
			serveErr = nil
		}
		transport.CloseIdleConnections()
		proxy.Transport.(*http.Transport).CloseIdleConnections()
		return errors.Join(closeErr, serveErr)
	}, nil
}

func (h *runtimeHost) cli(name string, args ...string) ([]byte, error) {
	args = append(args, "--data-dir", h.data, "--admin-url", h.adminURL)
	command := exec.CommandContext(h.ctx, h.binary, args...)
	command.Dir, command.Env = h.cwd, runtimeEnvironment()
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	h.sequence++
	prefix := filepath.Join(h.evidence, fmt.Sprintf("%02d-%s", h.sequence, name))
	writeErr := errors.Join(os.WriteFile(prefix+".stdout.json", stdout.Bytes(), fileMode), os.WriteFile(prefix+".stderr.log", stderr.Bytes(), fileMode))
	receipt := struct {
		Args     []string `json:"args"`
		ExitCode int      `json:"exit_code"`
		Error    string   `json:"error,omitempty"`
	}{Args: args, ExitCode: -1}
	if command.ProcessState != nil {
		receipt.ExitCode = command.ProcessState.ExitCode()
	}
	if err != nil {
		receipt.Error = err.Error()
	}
	writeErr = errors.Join(writeErr, writeJSON(prefix+".command.json", receipt))
	if writeErr != nil {
		writeErr = &cliEvidenceError{err: writeErr}
	}
	var commandErr error
	if err != nil {
		commandErr = fmt.Errorf("%s: %w; stderr: %s; stdout: %s", name, err, stderr.String(), stdout.String())
	}
	return stdout.Bytes(), errors.Join(commandErr, writeErr)
}

func (h *runtimeHost) api(name, method, path, body string) ([]byte, error) {
	req, err := http.NewRequestWithContext(h.ctx, method, h.adminURL+path, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+h.token)
	req.Header.Set("Content-Type", "application/json")
	response, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	raw, readErr := io.ReadAll(response.Body)
	err = errors.Join(readErr, response.Body.Close(), os.WriteFile(filepath.Join(h.evidence, name+".response"), raw, fileMode))
	if response.StatusCode != http.StatusOK {
		err = errors.Join(err, fmt.Errorf("%s %s: HTTP %d: %s", method, path, response.StatusCode, raw))
	}
	return raw, err
}

func writeJSON(name string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(name, append(raw, '\n'), fileMode)
}
