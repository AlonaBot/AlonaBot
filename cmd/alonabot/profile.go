package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
	"github.com/spf13/cobra"
)

const defaultProfileID = "default"
const defaultProfileColor = "teal"

func profileCommand(code *int) *cobra.Command {
	options := &managementOptions{}
	command := &cobra.Command{Use: "profile", Short: "Manage conversation profiles", Args: cobra.NoArgs}
	options.flags(command)
	list := func(cmd *cobra.Command, _ []string) error {
		return managementCall(cmd, code, options, http.MethodGet, "/api/v1/profiles", nil)
	}
	command.RunE = list
	command.AddCommand(&cobra.Command{Use: "list", Short: "List profiles and session counts", Args: cobra.NoArgs, RunE: list})
	var from string
	create := &cobra.Command{Use: "create <name>", Short: "Create a profile", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		return managementCall(cmd, code, options, http.MethodPost, "/api/v1/profiles", struct {
			Name     string `json:"name"`
			Color    string `json:"color"`
			CopyFrom string `json:"copy_from,omitempty"`
		}{args[0], defaultProfileColor, from})
	}}
	create.Flags().StringVar(&from, "from", "", "profile ID to copy")
	command.AddCommand(create)
	command.AddCommand(&cobra.Command{Use: "rename <id> <name>", Short: "Rename a profile", Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
		return managementCall(cmd, code, options, http.MethodPatch, "/api/v1/profiles/"+url.PathEscape(args[0]), struct {
			Name string `json:"name"`
		}{args[1]})
	}})
	command.AddCommand(&cobra.Command{Use: "delete <id>", Short: "Delete a profile", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		return managementCall(cmd, code, options, http.MethodDelete, "/api/v1/profiles/"+url.PathEscape(args[0]), nil)
	}})
	for _, assign := range []bool{true, false} {
		use, short := "unassign <instance> [session]", "Remove a profile assignment"
		args := cobra.RangeArgs(1, 2)
		if assign {
			use, short = "assign <id> <instance> [session]", "Assign a profile to an instance or session"
			args = cobra.RangeArgs(2, 3)
		}
		command.AddCommand(&cobra.Command{Use: use, Short: short, Args: args, RunE: func(cmd *cobra.Command, args []string) error {
			var id *string
			if assign {
				id = &args[0]
				args = args[1:]
			}
			if len(args) == 1 {
				return managementCall(cmd, code, options, http.MethodPut, "/api/v1/profiles/assignments/instances/"+url.PathEscape(args[0]), struct {
					Profile *string `json:"profile"`
				}{id})
			}
			return managementCall(cmd, code, options, http.MethodPut, "/api/v1/profiles/assignments/sessions", struct {
				Instance string  `json:"instance"`
				Session  string  `json:"session"`
				Profile  *string `json:"profile"`
			}{args[0], args[1], id})
		}})
	}
	for _, unset := range []bool{false, true} {
		use, short := "set <id> <plugin-id> <field-path> <json-value>", "Set a profile field override"
		args := cobra.ExactArgs(4)
		if unset {
			use, short = "unset <id> <plugin-id> <field-path>", "Remove a profile field override"
			args = cobra.ExactArgs(3)
		}
		command.AddCommand(&cobra.Command{Use: use, Short: short, Args: args, RunE: func(cmd *cobra.Command, args []string) error {
			value := json.RawMessage("null")
			if !unset {
				value = json.RawMessage(args[3])
				if !json.Valid(value) {
					return errors.New("field value must contain one valid JSON value")
				}
			}
			if args[0] == defaultProfileID {
				return profilePending(cmd, code, options, args[1], args[2], value)
			}
			return managementCall(cmd, code, options, http.MethodPatch, "/api/v1/profiles/"+url.PathEscape(args[0])+"/plugins/"+url.PathEscape(args[1]), struct {
				Overrides map[string]json.RawMessage `json:"overrides"`
			}{map[string]json.RawMessage{args[2]: value}})
		}})
	}
	return command
}

func profilePending(cmd *cobra.Command, code *int, options *managementOptions, pluginID, path string, value json.RawMessage) error {
	client, err := options.client()
	if err != nil {
		return err
	}
	*code = exitFailure
	endpoint := "/api/v1/plugins/" + url.PathEscape(pluginID) + "/config"
	raw, err := client.request(cmd.Context(), http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	var current struct {
		ActiveGeneration string          `json:"active_generation"`
		Schema           json.RawMessage `json:"schema"`
	}
	if err := json.Unmarshal(raw, &current); err != nil {
		return err
	}
	schema, problems := configschema.Parse(current.Schema)
	if len(problems) != 0 {
		return fmt.Errorf("invalid active config schema: %v", problems)
	}
	if !slices.Contains(schema.ProfilePaths(), path) {
		return fmt.Errorf("profile validation failed at %s: field is not marked for profiles", path)
	}
	if strings.TrimSpace(string(value)) != "null" {
		if err := schema.ValidateProfileValue(path, value); err != nil {
			return err
		}
	}
	// 按 object 路径构造 Merge Patch，沿用配置 API 的 Pending 与秘密字段处理。
	var patch any = value
	parts := strings.Split(path, ".")
	for i := len(parts) - 1; i >= 0; i-- {
		patch = map[string]any{parts[i]: patch}
	}
	result, err := client.request(cmd.Context(), http.MethodPatch, endpoint+"/pending", struct {
		BaseGeneration string `json:"base_generation"`
		Patch          any    `json:"patch"`
	}{current.ActiveGeneration, patch})
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), string(result))
	return err
}
