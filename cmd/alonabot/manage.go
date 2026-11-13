package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlonaBot/AlonaBot/internal/core/manage"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
	"github.com/spf13/cobra"
)

type managementOptions struct{ dataDir, adminURL, tokenFile string }
type managementClient struct {
	base  *url.URL
	token string
}

func (o *managementOptions) flags(command *cobra.Command) {
	flags := command.PersistentFlags()
	flags.StringVar(&o.dataDir, "data-dir", "data", "Core data directory (used to discover base_url)")
	flags.StringVar(&o.adminURL, "admin-url", "", "Core management URL (default: base_url in core_config.json)")
	flags.StringVar(&o.tokenFile, "token-file", "", "file containing the Core management Bearer token")
}

func managementCommands(code *int) []*cobra.Command {
	listOptions := &managementOptions{}
	list := &cobra.Command{Use: "list", Short: "List installed components", Args: cobra.NoArgs}
	listOptions.flags(list)
	list.RunE = func(cmd *cobra.Command, _ []string) error {
		return managementCall(cmd, code, listOptions, http.MethodGet, "/api/v1/packages", nil)
	}
	adminOptions := &managementOptions{}
	admin := &cobra.Command{Use: "admin", Short: "Manage Core administrators", Args: cobra.NoArgs}
	adminOptions.flags(admin)
	for _, action := range []string{"add", "remove"} {
		command := &cobra.Command{Use: action + " <adapter_id>:<external_user_id>", Short: "Update Core administrators", Args: cobra.ExactArgs(1)}
		command.RunE = func(cmd *cobra.Command, args []string) error {
			return managementCall(cmd, code, adminOptions, http.MethodPost, "/api/v1/admins/"+action, struct {
				User string `json:"user"`
			}{args[0]})
		}
		admin.AddCommand(command)
	}
	admin.AddCommand(&cobra.Command{Use: "list", Short: "List Core administrators", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		return managementCall(cmd, code, adminOptions, http.MethodGet, "/api/v1/admins", nil)
	}})
	adapterOptions := &managementOptions{}
	adapter := &cobra.Command{Use: "adapter", Short: "Manage Adapter instance configuration", Args: cobra.NoArgs}
	adapterOptions.flags(adapter)
	instance := &cobra.Command{Use: "instance", Short: "Create or delete Adapter instances", Args: cobra.NoArgs}
	var configFile string
	var confirmInstance bool
	create := &cobra.Command{Use: "create <adapter-package-id> <instance-id>", Short: "Save an Adapter instance configuration", Args: cobra.ExactArgs(2)}
	create.Flags().StringVar(&configFile, "config", "", "file containing the instance JSON configuration (default: {})")
	create.Flags().BoolVar(&confirmInstance, "yes", false, "confirm discarding instances owned by the previous Adapter component")
	create.RunE = func(cmd *cobra.Command, args []string) error {
		config := json.RawMessage("{}")
		if configFile != "" {
			raw, err := os.ReadFile(configFile)
			if err != nil {
				return fmt.Errorf("read instance config: %w", err)
			}
			config = raw
		}
		if !json.Valid(config) {
			return errors.New("instance config must contain valid JSON")
		}
		return managementCall(cmd, code, adapterOptions, http.MethodPost, "/api/v1/adapter/instances/create", struct {
			PackageID  string          `json:"package_id"`
			InstanceID string          `json:"instance_id"`
			Config     json.RawMessage `json:"config"`
			Confirm    bool            `json:"confirm"`
		}{args[0], args[1], config, confirmInstance})
	}
	instance.AddCommand(create)
	for _, action := range []string{"delete", "start", "stop"} {
		command := &cobra.Command{Use: action + " <instance-id>", Short: "Manage an Adapter instance", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
			return managementCall(cmd, code, adapterOptions, http.MethodPost, "/api/v1/adapter/instances/"+action, struct {
				InstanceID string `json:"instance_id"`
			}{args[0]})
		}}
		instance.AddCommand(command)
	}
	adapter.AddCommand(instance)
	commands := []*cobra.Command{list, admin, adapter}
	for _, action := range []string{"install", "enable", "disable", "restart", "uninstall", "pin", "reload", "update", "rollback"} {
		options := &managementOptions{}
		var yes, replace, purgeData, purgeAll bool
		var source, archive, releaseFile, signatureFile, keyFile string
		command := &cobra.Command{Use: action + " <id>", Short: "Manage installed components", Args: cobra.ExactArgs(1)}
		options.flags(command)
		command.Flags().BoolVar(&yes, "yes", false, "confirm the operation and its affected objects")
		if action == "install" || action == "update" {
			command.Flags().StringVar(&source, "source", "", "GitHub repository owner/repo")
			command.Flags().StringVar(&archive, "archive", "", "local package archive")
			command.Flags().StringVar(&releaseFile, "release", "", "release.json file (default: beside archive)")
			command.Flags().StringVar(&signatureFile, "signature", "", "release.json.sig file (optional)")
			command.Flags().StringVar(&keyFile, "public-key", "", "release.pub file (optional)")
			if action == "install" {
				command.Use = "install [archive]"
				command.Args = cobra.MaximumNArgs(1)
				command.Flags().BoolVar(&replace, "replace", false, "replace an installed component of the same kind")
			}
		}
		if action == "uninstall" {
			command.Flags().BoolVar(&purgeData, "purge-data", false, "also delete managed data")
			command.Flags().BoolVar(&purgeAll, "purge", false, "also delete managed data, configuration and installation binding")
			command.MarkFlagsMutuallyExclusive("purge-data", "purge")
		}
		if action == "pin" {
			command.Use = "pin <id> [version]"
			command.Args = cobra.RangeArgs(1, 2)
		}
		if action == "rollback" {
			command.Use = "rollback <id> <generation>"
			command.Args = cobra.ExactArgs(2)
		}
		command.RunE = func(cmd *cobra.Command, args []string) error {
			request := map[string]any{"confirm": yes}
			if action != "install" {
				request["id"] = args[0]
			}
			if action == "pin" {
				constraint := ""
				if len(args) > 1 {
					constraint = args[1]
				}
				request["constraint"] = constraint
			}
			if action == "rollback" {
				request["generation"] = args[1]
			}
			if action == "uninstall" {
				purge := manage.PurgeNone
				if purgeData {
					purge = manage.PurgeData
				}
				if purgeAll {
					purge = manage.PurgeAll
				}
				request["purge"] = purge
			}
			var body any = request
			if action == "install" || action == "update" {
				if action == "install" && len(args) != 0 {
					archive = args[0]
				}
				candidate, err := localInstallRequest(archive, releaseFile, signatureFile, keyFile, source)
				if err != nil {
					return err
				}
				candidate.Confirm, candidate.Replace = yes, replace
				if action == "install" {
					body = candidate
				} else {
					request["install"] = candidate
				}
			}
			return managementCall(cmd, code, options, http.MethodPost, "/api/v1/packages/"+action, body)
		}
		commands = append(commands, command)
	}
	return commands
}

func localInstallRequest(archive, releaseFile, signatureFile, keyFile, source string) (manage.InstallRequest, error) {
	req := manage.InstallRequest{Source: source}
	if source != "" {
		if archive != "" || releaseFile != "" || signatureFile != "" || keyFile != "" {
			return req, errors.New("--source and local package files are mutually exclusive")
		}
		return req, nil
	}
	if archive == "" {
		return req, nil
	} // update 沿用已绑定来源；install 由 API 报告缺少输入。
	var err error
	req.Archive, err = filepath.Abs(archive)
	if err != nil {
		return req, err
	}
	if releaseFile == "" {
		releaseFile = filepath.Join(filepath.Dir(req.Archive), "release.json")
	}
	raw, err := os.ReadFile(releaseFile)
	if err != nil {
		return req, fmt.Errorf("read release: %w", err)
	}
	req.ReleaseJSON = string(raw)
	for _, file := range []struct {
		explicit string
		name     string
		target   *string
	}{
		{signatureFile, "release.json.sig", &req.Signature}, {keyFile, "release.pub", &req.PublicKey},
	} {
		name := file.explicit
		if name == "" {
			name = filepath.Join(filepath.Dir(releaseFile), file.name)
		}
		raw, err := os.ReadFile(name)
		if file.explicit == "" && errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return req, fmt.Errorf("read %s: %w", file.name, err)
		}
		*file.target = string(raw)
	}
	return req, nil
}

func managementCall(cmd *cobra.Command, code *int, opt *managementOptions, method, path string, body any) error {
	client, err := opt.client()
	if err != nil {
		return err
	}
	*code = exitFailure
	result, err := client.request(cmd.Context(), method, path, body)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), string(result))
	return err
}

func (o *managementOptions) client() (*managementClient, error) {
	address := o.adminURL
	if address == "" {
		address = settings.Defaults().BaseURL
		// 客户端只读发现地址；settings.Load 还会创建或迁移文件，交由持有 core.lock 的 Core 调用。
		content, err := os.ReadFile(filepath.Join(o.dataDir, settings.FileName))
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("read Core settings: %w", err)
		}
		if err == nil {
			var document struct {
				BaseURL string `json:"base_url"`
			}
			if err := json.Unmarshal(content, &document); err != nil {
				return nil, fmt.Errorf("read base_url: %w", err)
			}
			if document.BaseURL != "" {
				address = document.BaseURL
			}
		}
		if strings.HasPrefix(address, "${") && strings.HasSuffix(address, "}") {
			name := strings.TrimSuffix(strings.TrimPrefix(address, "${"), "}")
			value, exists := os.LookupEnv(name)
			if !exists {
				return nil, fmt.Errorf("environment variable %s is not set", name)
			}
			address = value
		}
	}
	base, err := url.Parse(address)
	if err != nil || base.Host == "" || (base.Scheme != "http" && base.Scheme != "https") || base.User != nil || base.RawQuery != "" || base.Fragment != "" {
		return nil, errors.New("admin URL must be an absolute HTTP(S) URL without credentials, query or fragment")
	}
	hostname := strings.TrimSuffix(base.Hostname(), ".")
	if base.Scheme == "http" && !strings.EqualFold(hostname, "localhost") && !net.ParseIP(hostname).IsLoopback() {
		return nil, errors.New("non-loopback admin URL requires HTTPS")
	}
	token := os.Getenv("ALONABOT_ADMIN_TOKEN")
	if o.tokenFile != "" {
		content, err := os.ReadFile(o.tokenFile)
		if err != nil {
			return nil, fmt.Errorf("read admin token file: %w", err)
		}
		token = strings.TrimSpace(string(content))
	}
	if token == "" {
		return nil, errors.New("ALONABOT_ADMIN_TOKEN or --token-file is required")
	}
	return &managementClient{base: base, token: token}, nil
}

func (c *managementClient) request(ctx context.Context, method, path string, body any) (json.RawMessage, error) {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encode management request: %w", err)
		}
		reader = bytes.NewReader(raw)
	}
	address := *c.base
	address.Path = strings.TrimSuffix(address.Path, "/") + path
	request, err := http.NewRequestWithContext(ctx, method, address.String(), reader)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	client := *http.DefaultClient
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request management API: %w", err)
	}
	raw, readErr := io.ReadAll(response.Body)
	if err := errors.Join(readErr, response.Body.Close()); err != nil {
		return nil, fmt.Errorf("read management response: %w", err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		if response.StatusCode >= http.StatusMultipleChoices && response.StatusCode < http.StatusBadRequest {
			return nil, fmt.Errorf("management API %s: redirect to %q refused", response.Status, response.Header.Get("Location"))
		}
		var detail struct {
			Code        string   `json:"code"`
			SafeMessage string   `json:"safe_message"`
			Affected    []string `json:"affected"`
		}
		if err := json.Unmarshal(raw, &detail); err != nil {
			return nil, fmt.Errorf("management API %s: %s", response.Status, string(raw))
		}
		if detail.Affected != nil {
			return nil, fmt.Errorf("confirmation required: %s; retry with --yes", strings.Join(detail.Affected, ", "))
		}
		return nil, fmt.Errorf("%s: %s", detail.Code, detail.SafeMessage)
	}
	if !json.Valid(raw) {
		return nil, errors.New("management API returned invalid JSON")
	}
	return json.RawMessage(bytes.TrimSpace(raw)), nil
}
