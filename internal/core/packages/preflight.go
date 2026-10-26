// Package packages verifies signed release metadata and safely stages package archives.
package packages

import (
	"archive/tar"
	"compress/gzip"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/AlonaBot/AlonaBot/contracts/manifest"
	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/release"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
)

type Limits struct {
	MaxArchiveBytes, MaxExpandedBytes int64
	MaxEntries                        int
}

var DefaultLimits = Limits{MaxArchiveBytes: 256 << 20, MaxExpandedBytes: 1 << 30, MaxEntries: 10000}

type Input struct {
	ReleaseJSON []byte
	Signature   []byte
	PublicKey   ed25519.PublicKey
	Source      string
	OS, Arch    string
	ArchivePath string
}
type Candidate struct {
	Dir      string
	Manifest *manifest.Manifest
	Release  *release.Release
	Asset    release.Asset
}

func failure(code pb.ErrorCode, msg string) error { return &validate.Error{Code: code, Message: msg} }

func Preflight(in Input, stagingRoot string, lim Limits) (*Candidate, error) {
	if lim.MaxArchiveBytes <= 0 || lim.MaxExpandedBytes <= 0 || lim.MaxEntries <= 0 {
		return nil, errors.New("package limits must be positive")
	}
	r, err := release.Verify(in.ReleaseJSON, in.Signature, in.PublicKey)
	if err != nil {
		return nil, failure(pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, "release signature or metadata is invalid")
	}
	if r.Source != in.Source {
		return nil, failure(pb.ErrorCode_ERROR_CODE_PERMISSION_DENIED, "release source does not match the requested source")
	}
	asset, ok := r.Asset(in.OS, in.Arch)
	if !ok {
		return nil, failure(pb.ErrorCode_ERROR_CODE_UNSUPPORTED, "release has no asset for this platform")
	}
	f, err := os.Open(in.ArchivePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if info.Size() > lim.MaxArchiveBytes {
		return nil, failure(pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, "archive exceeds the configured size limit")
	}
	if info.Size() != asset.Size {
		return nil, failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "archive size does not match signed release metadata")
	}
	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		return nil, err
	}
	if hex.EncodeToString(hash.Sum(nil)) != asset.SHA256 {
		return nil, failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "archive digest does not match signed release metadata")
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(stagingRoot, 0o700); err != nil {
		return nil, err
	}
	dir, err := os.MkdirTemp(stagingRoot, "alonabot-candidate-")
	if err != nil {
		return nil, err
	}
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.RemoveAll(dir)
		}
	}()
	if err := extractReader(f, dir, lim); err != nil {
		return nil, err
	}
	manifestPath := filepath.Join(dir, "plugin.yaml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "archive does not contain a readable plugin.yaml")
	}
	parsed, err := manifest.Parse(data)
	if err != nil {
		return nil, failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "plugin.yaml is invalid")
	}
	m := parsed.Manifest
	switch {
	case m.ID != r.PluginID:
		return nil, failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "manifest plugin ID does not match release")
	case m.Version != r.Version:
		return nil, failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "manifest version does not match release")
	case m.Source != r.Source:
		return nil, failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "manifest source does not match release")
	case m.OS != in.OS || m.Arch != in.Arch:
		return nil, failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "manifest platform does not match the current platform")
	}
	entry := filepath.Join(dir, filepath.FromSlash(m.Entrypoint))
	st, err := os.Lstat(entry)
	if err != nil || !st.Mode().IsRegular() {
		return nil, failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "manifest entrypoint is not a regular file in the archive")
	}
	cleanup = false
	return &Candidate{Dir: dir, Manifest: m, Release: r, Asset: asset}, nil
}

func Extract(archivePath, dest string, lim Limits) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return extractReader(f, dest, lim)
}
func extractReader(r io.Reader, dest string, lim Limits) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "archive is not valid gzip data")
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	seen := map[string]struct{}{}
	entries := 0
	var expanded int64
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "archive tar stream is invalid")
		}
		entries++
		if entries > lim.MaxEntries {
			return failure(pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, "archive contains too many entries")
		}
		raw := h.Name
		if raw == "" || strings.Contains(raw, "\\") || strings.ContainsRune(raw, 0) || strings.HasPrefix(raw, "/") || path.IsAbs(raw) {
			return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "archive contains an invalid path")
		}
		name := raw
		if h.Typeflag == tar.TypeDir {
			name = strings.TrimSuffix(name, "/")
		}
		if name == "" || strings.Contains(name, ":") || path.Clean(name) != name {
			return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "archive path is not canonical")
		}
		for _, part := range strings.Split(name, "/") {
			if part == ".." || part == "." || part == "" {
				return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "archive path contains an unsafe component")
			}
		}
		if _, ok := seen[strings.ToLower(name)]; ok {
			return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "archive contains duplicate or case-colliding paths")
		}
		seen[strings.ToLower(name)] = struct{}{}
		target := filepath.Join(dest, filepath.FromSlash(name))
		switch h.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o700); err != nil {
				return err
			}
			if err := os.Chmod(target, 0o700); err != nil {
				return err
			}
		case tar.TypeReg, tar.TypeRegA:
			if h.Size < 0 || h.Size > lim.MaxExpandedBytes-expanded {
				return failure(pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, "archive expanded size exceeds the configured limit")
			}
			if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
				return err
			}
			mode := os.FileMode(0o600)
			if h.Mode&0o111 != 0 {
				mode = 0o700
			}
			out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
			if err != nil {
				return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "archive contains a duplicate output path")
			}
			n, copyErr := io.CopyN(out, tr, h.Size)
			closeErr := out.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
			expanded += n
			if err := os.Chmod(target, mode); err != nil {
				return err
			}
		default:
			return failure(pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, fmt.Sprintf("archive entry type %d is not allowed", h.Typeflag))
		}
	}
}
func (c *Candidate) MoveTo(packagesRoot string) (string, error) {
	if c == nil || c.Dir == "" || c.Manifest == nil || c.Release == nil {
		return "", errors.New("candidate is invalid")
	}
	digest := c.Asset.SHA256
	if len(digest) < 12 {
		return "", errors.New("candidate archive digest is invalid")
	}
	parent := filepath.Join(packagesRoot, c.Manifest.ID)
	if err := os.MkdirAll(parent, 0o700); err != nil {
		return "", err
	}
	final := filepath.Join(parent, c.Manifest.Version+"-"+digest[:12])
	if _, err := os.Lstat(final); err == nil {
		return "", failure(pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS, "immutable package directory already exists")
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	if err := os.Rename(c.Dir, final); err != nil {
		return "", err
	}
	c.Dir = ""
	return final, nil
}
func (c *Candidate) Discard() error {
	if c == nil || c.Dir == "" {
		return nil
	}
	err := os.RemoveAll(c.Dir)
	if err == nil {
		c.Dir = ""
	}
	return err
}
