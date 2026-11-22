package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/app"
	"github.com/AlonaBot/AlonaBot/internal/core/logging"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
	"github.com/AlonaBot/AlonaBot/internal/tune"
	"github.com/spf13/cobra"
)

// reportedError 标记已由 Core Logger 记录的运行错误，避免命令层重复输出。
type reportedError struct{ error }

func execute(args []string) int {
	// 1. 参数与命令解析失败使用 usage 退出码；进入业务执行后切换为 failure。
	code := exitUsage
	var inputs []string
	var values tune.Values
	root := &cobra.Command{Use: "alonabot", Short: "Run and configure AlonaBot Core", SilenceErrors: true, SilenceUsage: true}
	root.Args = cobra.ArbitraryArgs
	root.CompletionOptions.DisableDefaultCmd = true
	root.PersistentFlags().StringArrayVar(&inputs, "args", nil, "tuning overrides (KEY=VALUE,... or a text file; repeatable)")
	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		var err error
		overrides := inputs
		if env := os.Getenv("ALONABOT_ARGS"); env != "" {
			overrides = append([]string{env}, overrides...)
		}
		values, _, err = tune.Parse(overrides)
		return err
	}
	var dataDir, format, listen string
	var reload, stdio, console bool
	levelText := "info"
	runCommand := &cobra.Command{Use: "run [directories...]", Short: "Build development components and run AlonaBot Core", Args: cobra.ArbitraryArgs}
	for _, command := range []*cobra.Command{root, runCommand} {
		flags := command.Flags()
		flags.StringVar(&dataDir, "data-dir", "data", "Core data directory (only one Core per directory)")
		flags.StringVar(&listen, "listen", "", "Core HTTP listen address (overrides settings for this run)")
		flags.StringVar(&levelText, "log-level", "info", "global log level: debug, info, warn, error or critical")
		flags.StringVar(&format, "log-format", string(logging.FormatText), "stderr log format: text or json")
		flags.BoolVar(&reload, "reload", true, "rebuild and restart changed development components")
		flags.BoolVar(&stdio, "stdio", false, "use stdio for all Native component processes in this run")
		flags.BoolVar(&console, "console", false, "chat in this terminal (default: terminal stdin and no configured adapter instances)")
	}
	runCommand.RunE = func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()
		options := app.Options{DataDir: dataDir, Args: inputs}
		if flags.Changed("listen") {
			options.Listen = &listen
		}
		if flags.Changed("log-level") {
			level := logging.LevelCritical
			if !strings.EqualFold(levelText, "critical") {
				if err := level.UnmarshalText([]byte(levelText)); err != nil {
					return err
				}
			}
			switch level {
			case slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError, logging.LevelCritical:
				options.LogLevel = &level
			default:
				return fmt.Errorf("--log-level must be debug, info, warn, error or critical")
			}
		}
		if flags.Changed("log-format") {
			options.LogFormat = &format
		}
		enabled, err := consoleEnabled(dataDir, flags.Changed("console"), console, isTerminal(os.Stdin))
		if err != nil {
			return err
		}
		if enabled {
			level := slog.LevelWarn
			options.DefaultStderrLevel = &level
			options.RuntimeAdapters = map[string]settings.AdapterSettings{consoleAdapterID: {
				Package: consoleComponentID, Instances: map[string]json.RawMessage{consoleInstanceID: json.RawMessage(`{}`)},
			}}
			options.RuntimeAdmins = []settings.Admin{{AdapterID: consoleAdapterID, UserID: consoleUserID}}
		}
		code = exitFailure
		if err := runDevelopment(args, dataDir, reload, stdio, values.DevReloadPoll, options); err != nil {
			var detail *validate.Error
			if errors.As(err, &detail) && detail.Code == pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT {
				code = exitUsage
			}
			return err
		}
		return nil
	}
	root.RunE = runCommand.RunE
	tuneCommand := &cobra.Command{Use: "tune", Short: "Inspect tuning parameters", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error { return cmd.Help() }}
	listCommand := &cobra.Command{Use: "list", Short: "List default and effective tuning parameters", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			code = exitFailure
			return tune.WriteList(os.Stdout, values)
		}}
	tuneCommand.AddCommand(listCommand)
	root.AddCommand(runCommand, tuneCommand, adminTokenCommand(&code))
	root.AddCommand(initCommand(&code), buildCommand(&code), checkCommand(&code), testCommand(&code, &values), packCommand(&code))
	root.AddCommand(managementCommands(&code)...)
	root.AddCommand(logCommands(&code)...)
	root.AddCommand(diagnoseCommand(&code))
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
