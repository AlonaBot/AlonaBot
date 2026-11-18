package main

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/AlonaBot/AlonaBot/internal/core/diagnostics"
	"github.com/spf13/cobra"
)

const winsockConnectionRefused syscall.Errno = 10061 // Windows 的 WSAECONNREFUSED。

func diagnoseCommand(code *int) *cobra.Command {
	options := &managementOptions{}
	var since time.Duration
	var noLogs, noSettings, noComponents, noRedact bool
	var output string
	command := &cobra.Command{Use: "diagnose", Short: "Write a diagnostic zip from Core or a stopped Core data directory", Args: cobra.NoArgs}
	options.flags(command)
	flags := command.Flags()
	flags.DurationVar(&since, "since", 24*time.Hour, "include logs from this positive duration")
	flags.BoolVar(&noLogs, "no-logs", false, "exclude logs")
	flags.BoolVar(&noSettings, "no-settings", false, "exclude settings and plugin configurations")
	flags.BoolVar(&noComponents, "no-components", false, "exclude component and environment details")
	flags.BoolVar(&noRedact, "no-redact", false, "include credentials without redaction")
	flags.StringVarP(&output, "output", "o", "", "output zip file")
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		if since <= 0 {
			return errors.New("since must be a positive duration")
		}
		if noLogs && noSettings && noComponents {
			return errors.New("select at least one diagnostic group")
		}
		*code = exitFailure
		if output == "" {
			output = "alonabot-diagnostics-" + time.Now().UTC().Format("20060102T150405.000000000Z") + ".zip"
		}
		if options.tokenFile == "" && os.Getenv("ALONABOT_ADMIN_TOKEN") == "" {
			name := filepath.Join(options.dataDir, "admin_token")
			if _, err := os.Stat(name); err == nil {
				options.tokenFile = name
			} else if !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}
		client, err := options.client(false)
		if err != nil {
			return err
		}
		opts := diagnostics.Options{Since: since, Logs: !noLogs, Settings: !noSettings, Components: !noComponents, Redact: !noRedact}
		version := "(devel)"
		if build, ok := debug.ReadBuildInfo(); ok && build.Main.Version != "" {
			version = build.Main.Version
		}
		source := diagnostics.Directory{DataDir: options.dataDir, Version: version}
		if err := writeDiagnosticFile(cmd.Context(), output, client, source, opts); err != nil {
			return err
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout(), output)
		return err
	}
	return command
}

func writeDiagnosticFile(ctx context.Context, name string, client *managementClient, source diagnostics.Directory, opts diagnostics.Options) (err error) {
	// 1. 临时文件与目标位于同一目录，失败时保留已有目标。
	f, err := os.CreateTemp(filepath.Dir(name), ".alonabot-diagnostics-*.zip")
	if err != nil {
		return err
	}
	closed := false
	defer func() {
		if !closed {
			err = errors.Join(err, f.Close())
		}
		if err != nil {
			err = errors.Join(err, os.Remove(f.Name()))
		}
	}()
	request := struct {
		Since      string `json:"since"`
		Logs       bool   `json:"logs"`
		Settings   bool   `json:"settings"`
		Components bool   `json:"components"`
		Redact     bool   `json:"redact"`
	}{opts.Since.String(), opts.Logs, opts.Settings, opts.Components, opts.Redact}
	response, requestErr := client.open(ctx, http.MethodPost, "/api/v1/diagnostics", request)
	if requestErr != nil {
		if !errors.Is(requestErr, syscall.ECONNREFUSED) && !(runtime.GOOS == "windows" && errors.Is(requestErr, winsockConnectionRefused)) {
			return requestErr
		}
		if err := diagnostics.Write(ctx, f, source, opts); err != nil {
			return err
		}
	} else {
		_, copyErr := io.Copy(f, response.Body)
		if err := errors.Join(copyErr, response.Body.Close()); err != nil {
			return err
		}
	}
	// 2. 核对完整 zip 后再持久化并替换目标；截断的 HTTP 下载也不会留下最终文件。
	info, err := f.Stat()
	if err != nil {
		return err
	}
	archive, err := zip.NewReader(f, info.Size())
	if err != nil {
		return fmt.Errorf("invalid diagnostic archive: %w", err)
	}
	for _, entry := range archive.File {
		reader, err := entry.Open()
		if err != nil {
			return err
		}
		_, readErr := io.Copy(io.Discard, reader)
		if err := errors.Join(readErr, reader.Close()); err != nil {
			return err
		}
	}
	if err := f.Sync(); err != nil {
		return err
	}
	closeErr := f.Close()
	closed = true
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(f.Name(), name)
}
