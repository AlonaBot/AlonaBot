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

// githubRelease 让预览与安装共用来源校验；apiBase 决定资产摘要的信任根。
func (s *Service) githubRelease(ctx context.Context, source, apiBase string) (githubPublication, error) {
	ctx, cancel := context.WithTimeout(ctx, s.opt.Tuning.InstallDownloadTimeout)
	defer cancel()
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
	if request.URL.Scheme != "https" {
		return publication, invalid("GitHub API requires HTTPS")
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	response, err := s.opt.HTTPClient.Do(request)
	if err != nil {
		return publication, err
	}
	if response.StatusCode != http.StatusOK {
		err := s.githubHTTPError(address, response)
		if response.StatusCode == http.StatusNotFound {
			return publication, errors.Join(&validate.Error{Code: pb.ErrorCode_ERROR_CODE_NOT_FOUND, Message: "GitHub release not found"}, err)
		}
		return publication, err
	}
	limited := &io.LimitedReader{R: response.Body, N: s.opt.Tuning.InstallMaxReleaseJSON}
	raw, readErr := io.ReadAll(limited)
	if readErr == nil && limited.N == 0 {
		var extra [1]byte
		_, readErr = io.ReadFull(response.Body, extra[:])
		if readErr == nil {
			readErr = &validate.Error{Code: pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, Message: "GitHub release metadata exceeds the size limit"}
		} else if readErr == io.EOF {
			readErr = nil
		}
	}
	if err := errors.Join(readErr, response.Body.Close()); err != nil {
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
