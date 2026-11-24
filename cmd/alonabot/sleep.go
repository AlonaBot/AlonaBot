package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/AlonaBot/AlonaBot/internal/core/manage"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
	"github.com/spf13/cobra"
)

type sleepPlugin struct {
	ID      string          `json:"id"`
	Policy  json.RawMessage `json:"policy"`
	Runtime json.RawMessage `json:"runtime"`
}

func sleepPolicyCommand(code *int) *cobra.Command {
	options := &managementOptions{}
	command := &cobra.Command{Use: "sleep-policy [ID] [never|auto|duration]", Short: "Read or change plugin sleep policies", Long: "Read or change plugin/agent sleep policies. Omitted sleep_capable defaults to true; explicit false requires residency. Automatic admission also checks routes and activity. A positive per-component duration overrides resident-route restrictions with warnings.", Args: cobra.RangeArgs(0, 2)}
	options.flags(command)
	command.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) == 2 {
			path := "/api/v1/plugins/" + url.PathEscape(args[0]) + "/sleep-policy"
			if args[1] == "auto" {
				return managementCall(cmd, code, options, http.MethodDelete, path, nil)
			}
			return managementCall(cmd, code, options, http.MethodPut, path, struct {
				Policy string `json:"policy"`
			}{args[1]})
		}
		client, err := options.client(true)
		if err != nil {
			return err
		}
		*code = exitFailure
		if len(args) == 1 {
			plugin, err := client.sleepPlugin(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			return json.NewEncoder(cmd.OutOrStdout()).Encode(plugin)
		}
		// 全部请求成功后才输出，避免部分列表被误读为完整结果。
		sleep, err := client.request(cmd.Context(), http.MethodGet, "/api/v1/core/sleep", nil)
		if err != nil {
			return err
		}
		raw, err := client.request(cmd.Context(), http.MethodGet, "/api/v1/packages", nil)
		if err != nil {
			return err
		}
		var packages []manage.Package
		if err := json.Unmarshal(raw, &packages); err != nil {
			return err
		}
		plugins := make([]sleepPlugin, 0)
		for _, in := range packages {
			if in.Uninstalled || in.Role != string(state.RolePlugin) && in.Role != string(state.RoleAgent) {
				continue
			}
			plugin, err := client.sleepPlugin(cmd.Context(), in.ID)
			if err != nil {
				return err
			}
			plugins = append(plugins, plugin)
		}
		return json.NewEncoder(cmd.OutOrStdout()).Encode(struct {
			Sleep   json.RawMessage `json:"sleep"`
			Plugins []sleepPlugin   `json:"plugins"`
		}{sleep, plugins})
	}
	return command
}

func (c *managementClient) sleepPlugin(ctx context.Context, id string) (sleepPlugin, error) {
	path := "/api/v1/plugins/" + url.PathEscape(id)
	policy, err := c.request(ctx, http.MethodGet, path+"/sleep-policy", nil)
	if err != nil {
		return sleepPlugin{}, err
	}
	runtime, err := c.request(ctx, http.MethodGet, path+"/runtime", nil)
	if err != nil {
		return sleepPlugin{}, err
	}
	return sleepPlugin{ID: id, Policy: policy, Runtime: runtime}, nil
}
