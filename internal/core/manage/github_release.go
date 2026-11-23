package manage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

// githubRelease 让预览与安装共用来源校验；apiBase 决定资产摘要的信任根。
func (s *Service) githubRelease(ctx context.Context, source, apiBase string) (githubPublication, error) {
	var publication githubPublication
	if err := checkRepository(source); err != nil {
		return publication, err
	}
	address := strings.TrimRight(apiBase, "/") + "/repos/" + source + "/releases/latest"
	raw, status, err := s.readGitHubAPI(ctx, address, "application/vnd.github+json")
	if err != nil {
		if status == http.StatusNotFound {
			return publication, errors.Join(&validate.Error{Code: pb.ErrorCode_ERROR_CODE_NOT_FOUND, Message: "GitHub release not found"}, err)
		}
		return publication, err
	}
	if err := json.Unmarshal(raw, &publication); err != nil {
		var field *json.UnmarshalTypeError
		if errors.As(err, &field) && field.Field == "assets.digest" {
			return publication, denied("invalid GitHub asset digest")
		}
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
		respond(w, nil, downloadFailure{cause: err})
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
