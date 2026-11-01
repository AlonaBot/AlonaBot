package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/AlonaBot/AlonaBot/internal/core/logging"
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"github.com/spf13/cobra"
)

// reportedError 标记已由 Core Logger 记录的运行错误，避免命令层重复输出。
type reportedError struct{ error }

// usageError 标记启动设置或显式覆盖错误，由 CLI 使用 usage 退出码报告。
type usageError struct{ error }

// startupOverrides 只记录命令行显式设置的值，零值沿用设置文件。
type startupOverrides struct {
	listen *string
	level  *slog.Level
	format *string
}

func execute(args []string) int {
	// 1. 参数与命令解析失败使用 usage 退出码；进入业务执行后切换为 failure。
	code := exitUsage
	var inputs []string
	var values tune.Values
	var overrides []tune.Override
	root := &cobra.Command{Use: "alonabot", Short: "Run and configure AlonaBot Core", SilenceErrors: true, SilenceUsage: true}
	root.Args = cobra.NoArgs
	root.RunE = func(cmd *cobra.Command, args []string) error { return cmd.Help() }
	root.CompletionOptions.DisableDefaultCmd = true
	root.PersistentFlags().StringArrayVar(&inputs, "args", nil, "tuning overrides (KEY=VALUE,... or a text file; repeatable)")
	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		var err error
		values, overrides, err = tune.Parse(inputs)
		return err
	}
	var dataDir, pluginID, entry, format, listen string
	var stdio bool
	level := slog.LevelInfo
	runCommand := &cobra.Command{Use: "run", Short: "Run a local plugin in development mode", Args: cobra.NoArgs}
	flags := runCommand.Flags()
	flags.StringVar(&dataDir, "data-dir", "", "Core data directory (only one Core per directory)")
	flags.StringVar(&pluginID, "plugin-id", "", "plugin ID")
	flags.StringVar(&entry, "entry", "", "plugin executable path (development mode: a local build)")
	flags.BoolVar(&stdio, "stdio", false, "use stdio transport (default: TCP loopback)")
	flags.StringVar(&listen, "listen", "", "Core HTTP listen address (overrides settings for this run)")
	flags.TextVar(&level, "log-level", slog.LevelInfo, "global log level: debug, info, warn or error")
	flags.StringVar(&format, "log-format", string(logging.FormatText), "stderr log format: text or json")
	runCommand.RunE = func(cmd *cobra.Command, args []string) error {
		if dataDir == "" || pluginID == "" || entry == "" {
			return fmt.Errorf("--data-dir, --plugin-id and --entry are required")
		}
		selected := startupOverrides{}
		if flags.Changed("listen") {
			selected.listen = &listen
		}
		if flags.Changed("log-level") {
			switch level {
			case slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError:
				selected.level = &level
			default:
				return fmt.Errorf("--log-level must be debug, info, warn or error")
			}
		}
		if flags.Changed("log-format") {
			selected.format = &format
		}
		code = exitFailure
		if err := run(dataDir, pluginID, entry, stdio, values, overrides, selected); err != nil {
			var usage *usageError
			if errors.As(err, &usage) {
				code = exitUsage
			}
			return err
		}
		return nil
	}
	tuneCommand := &cobra.Command{Use: "tune", Short: "Inspect tuning parameters", Args: cobra.NoArgs, RunE: root.RunE}
	listCommand := &cobra.Command{Use: "list", Short: "List default and effective tuning parameters", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			code = exitFailure
			return tune.WriteList(os.Stdout, values)
		}}
	tuneCommand.AddCommand(listCommand)
	root.AddCommand(runCommand, tuneCommand, pluginCommand(&code))
	// 2. 由 Cobra 生成帮助，错误只在这里打印一次。
	var helpErr error
	help := root.HelpFunc()
	root.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		// Cobra 在 PreRun 之前处理 --help，帮助路径也要校验全局覆盖。
		_, _, helpErr = tune.Parse(inputs)
		if helpErr == nil {
			help(cmd, args)
		}
	})
	root.SetArgs(args)
	err := root.Execute()
	if err == nil {
		err = helpErr
	}
	if err != nil {
		var reported *reportedError
		if !errors.As(err, &reported) {
			fmt.Fprintln(os.Stderr, "alonabot:", err)
		}
		return code
	}
	return 0
}
