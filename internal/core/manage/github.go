package manage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
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

// copyGitHubArchive 共用 Release 资产与源码归档的请求期限和大小上限。
func (s *Service) copyGitHubArchive(ctx context.Context, prefix, original string, dest io.Writer) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, s.opt.Tuning.InstallDownloadTimeout)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, prefix+original, nil)
	if err != nil {
		return 0, err
	}
	response, err := s.opt.HTTPClient.Do(request)
	if err != nil {
		return 0, err
	}
	if response.StatusCode != http.StatusOK {
		return 0, s.githubHTTPError(request.URL.String(), response)
	}
	n, err := io.Copy(dest, io.LimitReader(response.Body, s.opt.Limits.MaxArchiveBytes+1))
	if err := errors.Join(err, response.Body.Close()); err != nil {
		return n, err
	}
	if n > s.opt.Limits.MaxArchiveBytes {
		return n, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, Message: "archive exceeds the size limit"}
	}
	return n, nil
}

// readGitHubAPI 读取受限的 GitHub API 响应，保留 HTTP 状态供调用方解释 404。
func (s *Service) readGitHubAPI(ctx context.Context, address, accept string) ([]byte, int, error) {
	ctx, cancel := context.WithTimeout(ctx, s.opt.Tuning.InstallDownloadTimeout)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return nil, 0, err
	}
	if request.URL.Scheme != "https" {
		return nil, 0, invalid("GitHub API requires HTTPS")
	}
	request.Header.Set("Accept", accept)
	response, err := s.opt.HTTPClient.Do(request)
	if err != nil {
		return nil, 0, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, response.StatusCode, s.githubHTTPError(address, response)
	}
	limited := &io.LimitedReader{R: response.Body, N: s.opt.Tuning.InstallMaxReleaseJSON}
	raw, readErr := io.ReadAll(limited)
	if readErr == nil && limited.N == 0 {
		var extra [1]byte
		_, readErr = io.ReadFull(response.Body, extra[:])
		if readErr == nil {
			readErr = &validate.Error{Code: pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, Message: "GitHub API metadata exceeds the size limit"}
		} else if readErr == io.EOF {
			readErr = nil
		}
	}
	return raw, response.StatusCode, errors.Join(readErr, response.Body.Close())
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
