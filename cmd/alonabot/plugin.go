package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlonaBot/AlonaBot/sdk/generate"
	"github.com/spf13/cobra"
)

const generatedFileMode = 0o644

func pluginCommand(code *int) *cobra.Command {
	var options generate.Options
	plugin := &cobra.Command{Use: "plugin", Short: "Generate and check plugin registrations", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error { return cmd.Help() }}
	flags := plugin.PersistentFlags()
	flags.StringVar(&options.Dir, "dir", ".", "plugin build entry directory")
	flags.StringVar(&options.GOOS, "goos", "", "target operating system (default: current environment)")
	flags.StringVar(&options.GOARCH, "goarch", "", "target architecture (default: current environment)")
	flags.StringSliceVar(&options.Tags, "tags", nil, "comma-separated build tags")
	for _, name := range []string{"generate", "check"} {
		command := &cobra.Command{Use: name, Args: cobra.NoArgs}
		if name == "generate" {
			command.Short = "Generate plugin registration code"
		} else {
			command.Short = "Check that plugin registration code is current"
		}
		command.RunE = func(cmd *cobra.Command, args []string) error {
			*code = exitFailure
			return pluginGeneration(cmd, options, name == "check")
		}
		plugin.AddCommand(command)
	}
	return plugin
}

func pluginGeneration(cmd *cobra.Command, options generate.Options, check bool) error {
	// 1. 扫描源码；逐行输出诊断后沿用已报告错误的退出路径。
	result, err := generate.Scan(cmd.Context(), options)
	if err != nil {
		return err
	}
	if len(result.Diagnostics) > 0 {
		for _, diagnostic := range result.Diagnostics {
			if _, err := fmt.Fprintf(cmd.ErrOrStderr(), "%s: %s\n", diagnostic.Pos, diagnostic.Message); err != nil {
				return fmt.Errorf("write diagnostic: %w", err)
			}
		}
		return &reportedError{errors.New("plugin scan reported diagnostics")}
	}
	if result.RootType == "" && result.AdapterType == "" {
		// 只有实例声明的旧目录已由 Scan 校验，check 无需写生成代码。
		if check {
			return nil
		}
		return errors.New("expected exactly one root type, found 0")
	}
	// 2. 先比较完整内容；check 只读，generate 仅在内容改变时写入。
	path := filepath.Join(options.Dir, generate.FileName)
	current, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read generated registration: %w", err)
	}
	if err == nil && bytes.Equal(current, result.Source) {
		return nil
	}
	if check {
		return fmt.Errorf("%s is out of date; run alonabot plugin generate", generate.FileName)
	}
	if err := os.WriteFile(path, result.Source, generatedFileMode); err != nil {
		return fmt.Errorf("write generated registration: %w", err)
	}
	return nil
}
