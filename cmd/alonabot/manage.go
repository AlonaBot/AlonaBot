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
	"runtime"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/manage"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
	"github.com/spf13/cobra"
)

type managementOptions struct{ dataDir, adminURL string }
type managementClient struct {
	host  string
	base  *url.URL
	token string
}

func (o *managementOptions) flags(command *cobra.Command) {
	flags := command.PersistentFlags()
	flags.StringVar(&o.dataDir, "data-dir", "data", "Core data directory (settings and persistent credentials)")
	flags.StringVar(&o.adminURL, "admin-url", "", "Local Core connection URL (default: listen in core_config.json)")
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
	adapterDebugCommands(adapter, instance, adapterOptions, code)
	adapter.AddCommand(instance)
	commands := []*cobra.Command{list, admin, adapter, profileCommand(code), sleepPolicyCommand(code)}
	for _, action := range []string{"install", "enable", "disable", "restart", "uninstall", "pin", "reload", "update", "rollback"} {
		options := &managementOptions{}
		var yes, replace, purgeData, purgeAll, manualMigrationApproved bool
		var source, archive string
		command := &cobra.Command{Use: action + " <id>", Short: "Manage installed components", Args: cobra.ExactArgs(1)}
		options.flags(command)
		command.Flags().BoolVar(&yes, "yes", false, "confirm the operation and its affected objects")
		if action == "install" || action == "update" {
			installMirrorFlags(command)
			command.Flags().StringVar(&source, "source", "", "GitHub repository owner/repo")
			command.Flags().StringVar(&archive, "archive", "", "local package archive")
			if action == "install" {
				command.Use = "install [archive|component-id]"
				command.Short = "Install a local archive, a bundled component ID, or the unique package in the current directory"
				command.Args = cobra.MaximumNArgs(1)
				command.Flags().BoolVar(&replace, "replace", false, "replace an installed component of the same kind")
			}
		}
		if action == "reload" || action == "update" || action == "rollback" {
			command.Flags().BoolVar(&manualMigrationApproved, "manual-migration-approved", false, "confirm external data has been backed up for this candidate (disable the component first)")
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
			if manualMigrationApproved {
				request["manual_migration_approved"] = true
			}
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
				selectedArchive := archive
				var expectedID string
				if action == "install" {
					if len(args) != 0 && (archive != "" || source != "") {
						return errors.New("install argument, --archive and --source are mutually exclusive")
					}
					if source == "" && archive == "" {
						var argument string
						if len(args) != 0 {
							argument = args[0]
						}
						var err error
						selectedArchive, expectedID, err = discoverInstallArchive(argument)
						if err != nil {
							return err
						}
					}
				}
				candidate, err := localInstallRequest(selectedArchive, source)
				if err != nil {
					return err
				}
				candidate.Confirm, candidate.Replace, candidate.ID = yes, replace, expectedID
				applyInstallMirrorFlags(cmd, &candidate)
				if action == "install" {
					body = candidate
				} else {
					request["install"] = candidate
				}
			}
			if action == "reload" {
				client, err := options.client(true)
				if err != nil {
					return err
				}
				if _, err := client.request(cmd.Context(), http.MethodPost, "/api/v1/plugins/"+url.PathEscape(args[0])+"/config/import", nil); err != nil {
					return err
				}
			}
			return managementCall(cmd, code, options, http.MethodPost, "/api/v1/packages/"+action, body)
		}
		commands = append(commands, command)
	}
	return commands
}

// discoverInstallArchive 只读定位安装输入，包内身份与归档预检交给管理服务。
func discoverInstallArchive(argument string) (string, string, error) {
	if argument != "" {
		info, err := os.Stat(argument)
		if err == nil && info.Mode().IsRegular() {
			return argument, "", nil
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", "", err
		}
		if _, err := validate.PluginID(argument); err != nil {
			return "", "", err
		}
		executable, err := os.Executable()
		if err != nil {
			return "", "", err
		}
		archive, err := bundledInstallArchive(filepath.Join(filepath.Dir(executable), "packages"), argument)
		return archive, argument, err
	}
	directory, err := os.Getwd()
	if err != nil {
		return "", "", err
	}
	raw, err := os.ReadFile(filepath.Join(directory, "plugin.yaml"))
	if err == nil {
		if _, projectErr := manifest.ParseProject(raw); projectErr == nil {
			directory = filepath.Join(directory, "bin")
		} else if _, packageErr := manifest.Parse(raw); packageErr != nil {
			return "", "", projectErr
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", "", err
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return "", "", fmt.Errorf("search packages in %s; candidates: []: %w", directory, err)
	}
	directories := []string{directory}
	for _, entry := range entries {
		if entry.IsDir() {
			directories = append(directories, filepath.Join(directory, entry.Name()))
		}
	}
	var candidates []string
	for _, dir := range directories {
		archives, err := platformArchives(dir)
		if err != nil {
			return "", "", err
		}
		candidates = append(candidates, archives...)
	}
	if len(candidates) != 1 {
		return "", "", fmt.Errorf("expected one package for %s/%s in %s; candidates: %v", runtime.GOOS, runtime.GOARCH, directory, candidates)
	}
	return candidates[0], "", nil
}

func bundledInstallArchive(directory, id string) (string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return "", fmt.Errorf("find bundled component %q in %s; available official components: []: %w", id, directory, err)
	}
	var available []string
	for _, entry := range entries {
		if entry.IsDir() {
			available = append(available, entry.Name())
		}
	}
	req, err := manage.BundledPackage(directory, id)
	if err != nil {
		return "", fmt.Errorf("find bundled component %q in %s; available official components: %v: %w", id, directory, available, err)
	}
	return req.Archive, nil
}

func platformArchives(directory string) ([]string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	var archives []string
	for _, entry := range entries {
		if entry.Type().IsRegular() && (strings.HasSuffix(entry.Name(), "-any.tar.gz") || strings.HasSuffix(entry.Name(), "-"+runtime.GOOS+"-"+runtime.GOARCH+".tar.gz")) {
			archives = append(archives, filepath.Join(directory, entry.Name()))
		}
	}
	return archives, nil
}

func localInstallRequest(archive, source string) (manage.InstallRequest, error) {
	req := manage.InstallRequest{Source: source}
	if source != "" && archive != "" {
		return req, errors.New("--source and --archive are mutually exclusive")
	}
	if archive == "" {
		return req, nil
	}
	var err error
	req.Archive, err = filepath.Abs(archive)
	return req, err
}

func managementCall(cmd *cobra.Command, code *int, opt *managementOptions, method, path string, body any) error {
	client, err := opt.client(true)
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

func (o *managementOptions) client(required bool) (*managementClient, error) {
	defaults := settings.Defaults()
	configuration := struct {
		Listen  string `json:"listen"`
		BaseURL string `json:"base_url"`
	}{defaults.Listen, defaults.BaseURL}
	content, err := os.ReadFile(filepath.Join(o.dataDir, settings.FileName))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("read Core settings: %w", err)
	}
	if err == nil {
		if err := json.Unmarshal(content, &configuration); err != nil {
			return nil, fmt.Errorf("read Core settings: %w", err)
		}
	}
	expand := func(value string) (string, error) {
		if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
			name := strings.TrimSuffix(strings.TrimPrefix(value, "${"), "}")
			result, ok := os.LookupEnv(name)
			if !ok {
				return "", fmt.Errorf("environment variable %s is not set", name)
			}
			return result, nil
		}
		return value, nil
	}
	controlled, err := expand(configuration.BaseURL)
	if err != nil {
		return nil, err
	}
	controlledURL, err := url.Parse(controlled)
	if err != nil || controlledURL.Host == "" {
		return nil, errors.New("base_url must be an absolute URL")
	}
	address := o.adminURL
	if address == "" {
		listen, err := expand(configuration.Listen)
		if err != nil {
			return nil, err
		}
		host, port, err := net.SplitHostPort(listen)
		if err != nil {
			return nil, fmt.Errorf("read local management listen address: %w", err)
		}
		if host == "" || net.ParseIP(host).IsUnspecified() {
			host = "127.0.0.1"
		}
		address = "http://" + net.JoinHostPort(host, port)
	}
	base, err := url.Parse(address)
	if err != nil || base.Host == "" || (base.Scheme != "http" && base.Scheme != "https") || base.User != nil || base.RawQuery != "" || base.Fragment != "" {
		return nil, errors.New("admin URL must be an absolute HTTP(S) URL without credentials, query or fragment")
	}
	hostname := strings.TrimSuffix(base.Hostname(), ".")
	if !strings.EqualFold(hostname, "localhost") && !net.ParseIP(hostname).IsLoopback() {
		return nil, errors.New("admin URL must be local; use ssh or docker exec for remote Core")
	}
	content, err = os.ReadFile(filepath.Join(o.dataDir, "admin_token"))
	if err != nil && (required || !errors.Is(err, os.ErrNotExist)) {
		return nil, fmt.Errorf("read admin token; run admin-token reset if missing: %w", err)
	}
	token := strings.TrimSpace(string(content))
	if required && token == "" {
		return nil, errors.New("admin token file is empty; run admin-token reset")
	}
	return &managementClient{base: base, host: controlledURL.Host, token: token}, nil
}

func (c *managementClient) request(ctx context.Context, method, path string, body any) (json.RawMessage, error) {
	response, err := c.open(ctx, method, path, body)
	if err != nil {
		return nil, err
	}
	raw, readErr := io.ReadAll(response.Body)
	if err := errors.Join(readErr, response.Body.Close()); err != nil {
		return nil, fmt.Errorf("read management response: %w", err)
	}
	if !json.Valid(raw) {
		return nil, errors.New("management API returned invalid JSON")
	}
	return json.RawMessage(bytes.TrimSpace(raw)), nil
}

func (c *managementClient) open(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encode management request: %w", err)
		}
		reader = bytes.NewReader(raw)
	}
	address := *c.base
	endpoint, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("parse management path: %w", err)
	}
	address.Path = strings.TrimSuffix(address.Path, "/") + endpoint.Path
	address.RawQuery = endpoint.RawQuery
	request, err := http.NewRequestWithContext(ctx, method, address.String(), reader)
	if err != nil {
		return nil, err
	}
	request.Host = c.host
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
	if response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices {
		return response, nil
	}
	raw, readErr := io.ReadAll(response.Body)
	if err := errors.Join(readErr, response.Body.Close()); err != nil {
		return nil, fmt.Errorf("read management response: %w", err)
	}
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
