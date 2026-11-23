package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/AlonaBot/AlonaBot/sdk/debug"
	"github.com/spf13/cobra"
)

func adapterDebugCommands(command, instance *cobra.Command, options *managementOptions, code *int) {
	instance.AddCommand(&cobra.Command{Use: "list", Short: "List Adapter instances and runtime states", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		return managementCall(cmd, code, options, http.MethodGet, "/api/v1/adapter/instances", nil)
	}})
	connect := func() (*debug.Client, error) {
		c, err := options.client(true)
		if err != nil {
			return nil, err
		}
		return &debug.Client{URL: c.base.String(), Host: c.host, Token: c.token}, nil
	}
	command.AddCommand(&cobra.Command{Use: "actions <instance> [operation]", Short: "Discover live Adapter operations and schemas", Args: cobra.RangeArgs(1, 2), RunE: func(cmd *cobra.Command, args []string) error {
		*code = exitFailure
		c, err := connect()
		if err != nil {
			return err
		}
		name := ""
		if len(args) > 1 {
			name = args[1]
		}
		operations, err := c.Operations(cmd.Context(), args[0], name)
		if err != nil {
			return err
		}
		return json.NewEncoder(cmd.OutOrStdout()).Encode(operations)
	}})
	var params, file, to string
	var version uint32
	call := &cobra.Command{Use: "call <instance> <operation>", Short: "Call a live Adapter operation", Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
		raw := json.RawMessage(params)
		if file != "" {
			data, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("read params file: %w", err)
			}
			raw = data
		}
		if !json.Valid(raw) {
			return errors.New("params must contain one JSON value")
		}
		var target json.RawMessage
		if to != "" {
			if !json.Valid([]byte(to)) {
				return errors.New("--to must contain a Target JSON object")
			}
			target = json.RawMessage(to)
		}
		*code = exitFailure
		c, err := connect()
		if err != nil {
			return err
		}
		result, callErr := c.Invoke(cmd.Context(), debug.Call{Instance: args[0], Name: args[1], Version: version, To: target, Params: raw})
		if result.TraceID != "" {
			if err := json.NewEncoder(cmd.OutOrStdout()).Encode(result); err != nil {
				return errors.Join(callErr, err)
			}
		}
		return callErr
	}}
	call.Flags().Uint32Var(&version, "version", 1, "operation version")
	call.Flags().StringVar(&params, "params", "{}", "operation params JSON")
	call.Flags().StringVar(&file, "params-file", "", "file containing operation params JSON")
	call.Flags().StringVar(&to, "to", "", "Target JSON object, for example {\"chatId\":\"private:alice\"}")
	call.MarkFlagsMutuallyExclusive("params", "params-file")
	command.AddCommand(call)
}
