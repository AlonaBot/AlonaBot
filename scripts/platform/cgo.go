package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
)

const cgoPackageID = "platform-cgo"

// cgoBuildRecord 接收操作者保留的真实构建输入与完整命令结果；其中的命令只作为证据读取。
type cgoBuildRecord struct {
	Source          string            `json:"source"`
	Binary          string            `json:"binary"`
	Archive         string            `json:"archive"`
	Environment     map[string]string `json:"environment"`
	CompilerVersion stepResult        `json:"compiler_version"`
	Build           stepResult        `json:"build"`
}

type cgoEvidence struct {
	Record    string         `json:"record"`
	Build     cgoBuildRecord `json:"build"`
	Binary    candidate      `json:"binary"`
	Libraries []string       `json:"linked_libraries"` // 从已核对归档字节的实际二进制读取直接动态依赖。
	Status    string         `json:"status"`
}

func (r *report) prepareCGO(path string) error {
	evidence := &cgoEvidence{Record: path, Status: statusIncomplete}
	r.CGO = evidence
	return r.step("cgo_inputs", func(step *stepResult) error {
		raw, err := os.ReadFile(path)
		step.Output = string(raw)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(raw, &evidence.Build); err != nil {
			return err
		}
		build := &evidence.Build
		for _, command := range []stepResult{build.CompilerVersion, build.Build} {
			if command.Status != statusPassed || (command.ExitCode != nil && *command.ExitCode != 0) {
				return fmt.Errorf("CGO input command failed: %v: %s\n%s", command.Command, command.Error, command.Output)
			}
			if len(command.Command) == 0 || command.Directory == "" || command.ExitCode == nil {
				return errors.New("CGO evidence requires compiler and build commands, directories and exit codes")
			}
		}
		if build.CompilerVersion.Output == "" {
			return errors.New("CGO evidence is missing the C compiler version")
		}
		for _, name := range []string{build.Source, build.Binary, build.Archive} {
			if !filepath.IsAbs(name) {
				return fmt.Errorf("CGO source, binary and archive paths must be absolute: %q", name)
			}
			if _, err := os.Stat(name); err != nil {
				return err
			}
		}
		if build.Environment["CGO_ENABLED"] != "1" || build.Environment["GOOS"] != runtime.GOOS || build.Environment["GOARCH"] != runtime.GOARCH {
			return fmt.Errorf("CGO build environment does not match this platform: %v", build.Environment)
		}
		evidence.Binary, err = executableCandidate(build.Binary)
		if err != nil {
			return err
		}
		var enabled bool
		for _, setting := range evidence.Binary.Settings {
			if setting.Key == "CGO_ENABLED" && setting.Value == "1" {
				enabled = true
			}
		}
		if !enabled {
			return errors.New("CGO fixture binary was not built with CGO_ENABLED=1")
		}
		component, err := packageBuild(build.Archive)
		if err != nil {
			return err
		}
		r.Components = append(r.Components, component)
		if err := cgoArchiveBinary(build.Archive, build.Binary); err != nil {
			return err
		}
		evidence.Libraries, err = linkedLibraries(build.Binary)
		return err
	})
}

// cgoArchiveBinary 直接比对安装归档的入口字节，确保链接库与构建身份来自实际待运行的程序。
func cgoArchiveBinary(path, binary string) (err error) {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	compressed, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, compressed.Close()) }()
	archive := tar.NewReader(compressed)
	entrypoint := "bin/component"
	if runtime.GOOS == "windows" {
		entrypoint += ".exe"
	}
	for {
		header, err := archive.Next()
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("CGO archive is missing %s", entrypoint)
		}
		if err != nil {
			return err
		}
		if header.Name != entrypoint {
			continue
		}
		packed, err := io.ReadAll(archive)
		if err != nil {
			return err
		}
		built, err := os.ReadFile(binary)
		if err != nil {
			return err
		}
		if !bytes.Equal(packed, built) {
			return fmt.Errorf("CGO archive %s differs from built binary %s", path, binary)
		}
		return nil
	}
}

func linkedLibraries(path string) ([]string, error) {
	switch runtime.GOOS {
	case "darwin":
		file, err := macho.Open(path)
		if err != nil {
			return nil, err
		}
		libraries, err := file.ImportedLibraries()
		return libraries, errors.Join(err, file.Close())
	case "linux":
		file, err := elf.Open(path)
		if err != nil {
			return nil, err
		}
		libraries, err := file.ImportedLibraries()
		return libraries, errors.Join(err, file.Close())
	case "windows":
		file, err := pe.Open(path)
		if err != nil {
			return nil, err
		}
		// debug/pe.ImportedLibraries 尚未实现；ImportedSymbols 的后缀保留对应 DLL 名。
		symbols, err := file.ImportedSymbols()
		var libraries []string
		for _, symbol := range symbols {
			if _, library, ok := strings.Cut(symbol, ":"); ok {
				libraries = append(libraries, library)
			}
		}
		slices.Sort(libraries)
		return slices.Compact(libraries), errors.Join(err, file.Close())
	default:
		return nil, fmt.Errorf("unsupported CGO binary format on %s", runtime.GOOS)
	}
}

func (r *report) runCGO(ctx context.Context, opt options, original *adminSession, ids []string, cliToken string) (err error) {
	// 百次启停与强杀恢复结束后，独立一次运行验证 CGO；原有四组件继续参与资源检查。
	core, startErr := r.startCore(ctx, opt)
	if core != nil {
		defer func() { err = errors.Join(err, core.stop()) }()
	}
	if startErr != nil {
		return startErr
	}
	core.run.Purpose = "cgo"
	session := *original
	session.core = core
	cli, err := session.cliProxy()
	if err != nil {
		return err
	}
	defer cli.Close()
	if err := r.command(ctx, "install_"+cgoPackageID, opt.core, "install", r.CGO.Build.Archive, "--yes", "--admin-url", cli.URL, "--token-file", cliToken); err != nil {
		return err
	}
	if _, err := session.request(ctx, "enable_"+cgoPackageID, http.MethodPost, "/api/v1/packages/enable", map[string]string{"id": cgoPackageID}, http.StatusOK, false); err != nil {
		return err
	}
	ids = append(slices.Clone(ids), cgoPackageID)
	if err := session.ready(ctx, ids, make(map[string]string)); err != nil {
		return err
	}
	raw, err := session.request(ctx, "cgo_native_call", http.MethodGet, "/plugins/"+cgoPackageID+"/native?value=81", nil, http.StatusOK, false)
	if err != nil {
		return err
	}
	if err := r.step("cgo_native_identity", func(*stepResult) error {
		var response struct {
			Value    float64 `json:"value"`
			Compiler string  `json:"compiler"`
			OS       string  `json:"os"`
			Arch     string  `json:"arch"`
			PID      int     `json:"pid"`
		}
		if err := json.Unmarshal(raw, &response); err != nil {
			return err
		}
		if response.Value != 9 || response.Compiler == "" || response.OS != runtime.GOOS || response.Arch != runtime.GOARCH {
			return fmt.Errorf("unexpected native C result or platform: %s", raw)
		}
		for _, activation := range core.run.Activations {
			if activation.ID == cgoPackageID && activation.PID == response.PID {
				return nil
			}
		}
		return fmt.Errorf("CGO response PID differs from activated plugin: %s", raw)
	}); err != nil {
		return err
	}
	if err := r.stopAndRelease(ctx, core, ids); err != nil {
		return err
	}
	r.CGO.Status = statusPassed
	return nil
}
