package manage

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/AlonaBot/AlonaBot/internal/core/diagnostics"
	"github.com/AlonaBot/AlonaBot/internal/core/settings"
)

type downloadFailure struct {
	cause  error
	prefix string
}

func (e downloadFailure) Error() string {
	return diagnostics.RedactText(redactDownloadQuery(diagnostics.OmitURLQueries(e.cause.Error()), e.prefix))
}
func (e downloadFailure) Unwrap() error { return e.cause }

func githubClient(base *http.Client) *http.Client {
	client := *base
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if base.CheckRedirect != nil {
			if err := base.CheckRedirect(req, via); err != nil {
				return err
			}
		} else if len(via) >= 10 { // 保留 net/http 默认的重定向次数上限。
			return errors.New("stopped after 10 redirects")
		}
		if req.URL.Scheme != "https" {
			return errors.New("GitHub redirects require HTTPS")
		}
		if !strings.EqualFold(req.URL.Host, via[len(via)-1].URL.Host) {
			req.Header.Del("Referer")
		}
		return nil
	}
	return &client
}

func (s *Service) githubHTTPError(address string, response *http.Response) error {
	raw, err := io.ReadAll(io.LimitReader(response.Body, s.opt.Tuning.InstallMaxHTTPError))
	return errors.Join(fmt.Errorf("download %s: %s: %s", address, response.Status, raw), err, response.Body.Close())
}

// redactDownloadQuery 同时遮盖镜像在响应正文中单独回显的凭据值。
func redactDownloadQuery(text, prefix string) string {
	_, query, _ := strings.Cut(prefix, "?")
	if query == "" {
		return text
	}
	var replacements []string
	for field := range strings.SplitSeq(query, "&") {
		_, value, hasValue := strings.Cut(field, "=")
		if !hasValue {
			value = field
		}
		if value != "" {
			replacements = append(replacements, value, "[REDACTED]")
			if decoded, err := url.QueryUnescape(value); err == nil && decoded != value {
				replacements = append(replacements, decoded, "[REDACTED]")
			}
		}
	}
	return strings.NewReplacer(replacements...).Replace(text)
}

// githubEndpoints 冻结本次下载所用设置，逐字段应用显式覆盖。
func (s *Service) githubEndpoints(req InstallRequest) (string, string, error) {
	current := s.opt.Settings.Settings()
	base, prefix := current.GitHubAPIBase, current.GitHubDownloadPrefix
	if req.GitHubAPIBase != nil {
		base = *req.GitHubAPIBase
		if err := settings.ValidateGitHubEndpoint(base, false); err != nil {
			return "", "", invalid("github_api_base: " + err.Error())
		}
	}
	if req.GitHubDownloadPrefix != nil {
		prefix = *req.GitHubDownloadPrefix
		if err := settings.ValidateGitHubEndpoint(prefix, true); err != nil {
			return "", "", invalid("github_download_prefix: " + err.Error())
		}
	}
	if base == "" {
		base = s.opt.ReleaseBaseURL
	}
	return base, prefix, nil
}

type githubSettings struct {
	APIBase        string `json:"github_api_base"`
	DownloadPrefix string `json:"github_download_prefix"`
}

func (s *Service) githubRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/core/github/release", s.githubReleasePreview)
	mux.HandleFunc("GET /api/v1/core/github", func(w http.ResponseWriter, r *http.Request) {
		current := s.opt.Settings.Settings()
		w.Header().Set("Cache-Control", "no-store")
		respond(w, githubSettings{current.GitHubAPIBase, current.GitHubDownloadPrefix}, nil)
	})
	mux.HandleFunc("PUT /api/v1/core/github", func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			APIBase        *string `json:"github_api_base"`
			DownloadPrefix *string `json:"github_download_prefix"`
		}
		if err := decode(r, &request); err != nil {
			respond(w, nil, err)
			return
		}
		if request.APIBase == nil || request.DownloadPrefix == nil {
			respond(w, nil, invalid("github_api_base and github_download_prefix are required"))
			return
		}
		if err := settings.ValidateGitHubEndpoint(*request.APIBase, false); err != nil {
			respond(w, nil, invalid("github_api_base: "+err.Error()))
			return
		}
		if err := settings.ValidateGitHubEndpoint(*request.DownloadPrefix, true); err != nil {
			respond(w, nil, invalid("github_download_prefix: "+err.Error()))
			return
		}
		err := s.opt.Settings.Update(func(doc *settings.Document) error {
			if err := doc.Set([]string{"github_api_base"}, *request.APIBase); err != nil {
				return err
			}
			return doc.Set([]string{"github_download_prefix"}, *request.DownloadPrefix)
		})
		respond(w, struct{}{}, err)
	})
}
