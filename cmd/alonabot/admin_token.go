package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlonaBot/AlonaBot/internal/core/adminauth"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/spf13/cobra"
)

func adminTokenCommand(code *int) *cobra.Command {
	var dataDir string
	group := &cobra.Command{Use: "admin-token", Short: "Show, initialize or reset the admin token", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error { return cmd.Help() }}
	group.PersistentFlags().StringVar(&dataDir, "data-dir", "data", "Core data directory")
	for _, action := range []string{"init", "reset"} {
		var ifMissing bool
		command := &cobra.Command{Use: action, Short: action + " the offline admin token (Core must be stopped)", Args: cobra.NoArgs}
		if action == "init" {
			command.Flags().BoolVar(&ifMissing, "if-missing", false, "Initialize only if no admin credential exists")
		}
		command.RunE = func(cmd *cobra.Command, args []string) error {
			if dataDir == "" {
				return errors.New("--data-dir is required")
			}
			*code = exitFailure
			return writeAdminToken(cmd.Context(), dataDir, action == "reset", ifMissing)
		}
		group.AddCommand(command)
	}
	show := &cobra.Command{Use: "show", Short: "Show the persistent admin token file", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if dataDir == "" {
				return errors.New("--data-dir is required")
			}
			*code = exitFailure
			token, err := os.ReadFile(filepath.Join(dataDir, adminauth.FirstTokenFile))
			if err != nil {
				return fmt.Errorf("read admin token file: %w", err)
			}
			if _, err := cmd.OutOrStdout().Write(token); err != nil {
				return fmt.Errorf("write admin token: %w", err)
			}
			return nil
		}}
	group.AddCommand(show)
	return group
}

func writeAdminToken(ctx context.Context, dataDir string, reset, ifMissing bool) (err error) {
	// 1. 先持有 core.lock，和 Core 使用相同的状态库入口。
	store, err := state.Open(ctx, dataDir)
	if err != nil {
		if errors.Is(err, state.ErrDataDirLocked) {
			return fmt.Errorf("stop Core before changing the admin token: %w", err)
		}
		return err
	}
	defer func() { err = errors.Join(err, store.Close()) }()

	if ifMissing {
		// 持锁检查，避免与正在运行的 Core 或另一次初始化交错。
		if _, err := store.AdminCredential(ctx); err == nil {
			return nil
		} else if !errors.Is(err, state.ErrNotFound) {
			return err
		}
	}

	// 2. 凭据持久化成功后只输出文件位置。

	if reset {
		_, err = adminauth.ResetToken(ctx, store, dataDir)
	} else {
		_, err = adminauth.InitToken(ctx, store, dataDir)
	}
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintln(os.Stderr, "Admin token saved:", filepath.Join(dataDir, adminauth.FirstTokenFile)); err != nil {
		return fmt.Errorf("write admin token: %w", err)
	}
	return nil
}
