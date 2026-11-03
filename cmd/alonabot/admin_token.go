package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/AlonaBot/AlonaBot/internal/core/adminauth"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/spf13/cobra"
)

// terminal 可由本地自测替换；生产使用平台的终端查询，避免把 /dev/null 当作终端。
var terminal = isTerminal

func adminTokenCommand(code *int) *cobra.Command {
	var dataDir string
	group := &cobra.Command{Use: "admin-token", Short: "Initialize or reset the offline admin token", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error { return cmd.Help() }}
	group.PersistentFlags().StringVar(&dataDir, "data-dir", "", "Core data directory (Core must be stopped)")
	for _, action := range []string{"init", "reset"} {
		command := &cobra.Command{Use: action, Short: action + " the admin token", Args: cobra.NoArgs}
		command.RunE = func(cmd *cobra.Command, args []string) error {
			if dataDir == "" {
				return errors.New("--data-dir is required")
			}
			if !terminal(os.Stdin) || !terminal(os.Stderr) {
				return errors.New("admin-token requires stdin and stderr to be terminals")
			}
			*code = exitFailure
			return writeAdminToken(cmd.Context(), dataDir, action == "reset")
		}
		group.AddCommand(command)
	}
	return group
}

func writeAdminToken(ctx context.Context, dataDir string, reset bool) (err error) {
	// 1. 先持有 core.lock，和 Core 使用相同的状态库入口。
	store, err := state.Open(ctx, dataDir)
	if err != nil {
		if errors.Is(err, state.ErrDataDirLocked) {
			return fmt.Errorf("stop Core before changing the admin token: %w", err)
		}
		return err
	}
	defer func() { err = errors.Join(err, store.Close()) }()

	// 2. 凭证事务成功后只写一次本机终端，不经过 Logger。
	var token string
	if reset {
		token, err = adminauth.ResetToken(ctx, store)
	} else {
		token, err = adminauth.InitToken(ctx, store)
	}
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintln(os.Stderr, token); err != nil {
		return fmt.Errorf("write admin token: %w", err)
	}
	return nil
}
