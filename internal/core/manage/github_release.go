package manage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/release"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
)

type githubPublication struct {
	release.GitHubRelease
	Body string `json:"body"`
}

// githubRelease 让版本预览与安装共用来源校验、镜像基址和 HTTP 客户端。
func (s *Service) githubRelease(ctx context.Context, source, apiBase string) (githubPublication, error) {
	var publication githubPublication
	if !repositoryName.MatchString(source) {
		return publication, invalid("source must be owner/repo")
	}
	for _, part := range strings.Split(source, "/") {
		if part == "." || part == ".." {
			return publication, invalid("source must be owner/repo")
		}
	}
	address := strings.TrimRight(apiBase, "/") + "/repos/" + source + "/releases/latest"
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return publication, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	response, err := s.opt.HTTPClient.Do(request)
	if err != nil {
		return publication, err
	}
	raw, readErr := io.ReadAll(response.Body)
	if err := errors.Join(readErr, response.Body.Close()); err != nil {
		return publication, err
	}
	if response.StatusCode != http.StatusOK {
		err := fmt.Errorf("download %s: %s: %s", address, response.Status, raw)
		if response.StatusCode == http.StatusNotFound {
			return publication, &validate.Error{Code: pb.ErrorCode_ERROR_CODE_NOT_FOUND, Message: err.Error()}
		}
		return publication, err
	}
	if err := json.Unmarshal(raw, &publication); err != nil {
		return publication, fmt.Errorf("parse GitHub release: %w", err)
	}
	if err := publication.Validate(source); err != nil {
		return publication, invalid(err.Error())
	}
	return publication, nil
}

func (s *Service) githubReleasePreview(w http.ResponseWriter, r *http.Request) {
	base, _, err := s.githubEndpoints(InstallRequest{})
	var publication githubPublication
	source := r.URL.Query().Get("source")
	if err == nil {
		publication, err = s.githubRelease(r.Context(), source, base)
	}
	if err != nil {
		respond(w, nil, downloadFailure{err})
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	respond(w, struct {
		Repo       string `json:"repo"`
		Version    string `json:"version"`
		ReleaseURL string `json:"release_url"`
		Markdown   string `json:"markdown"`
		FetchedAt  string `json:"fetched_at"`
	}{source, strings.TrimPrefix(publication.Tag, "v"), publication.URL, publication.Body, time.Now().Format(time.RFC3339Nano)}, nil)
}
