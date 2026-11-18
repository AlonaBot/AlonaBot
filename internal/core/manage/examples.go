package manage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"slices"

	"github.com/AlonaBot/AlonaBot/contracts/configschema"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/core/state"
)

func (s *Service) example(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	name := "examples/" + r.PathValue("path")
	if _, err := validate.PluginID(id); err != nil || !validStaticPath(name) {
		http.NotFound(w, r)
		return
	}
	file, err := s.exampleContent(r.Context(), id, name)
	if errors.Is(err, errStaticNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		staticError(w, err)
		return
	}
	defer file.close()
	contentType := ""
	switch path.Ext(name) {
	case ".png":
		contentType = "image/png"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".webp":
		contentType = "image/webp"
	case ".gif":
		contentType = "image/gif"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Security-Policy", iconCSP)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(w, r, file.info.Name(), file.info.ModTime(), file.content)
}

func (s *Service) exampleContent(ctx context.Context, id, name string) (*staticFile, error) {
	unlock := s.lock(id)
	defer unlock()
	in, err := s.opt.Store.GetInstall(ctx, id)
	if errors.Is(err, state.ErrNotFound) || err == nil && (in.Uninstalled || in.Source == "dev") {
		return nil, errStaticNotFound
	}
	if err != nil {
		return nil, err
	}
	// 1. 用 Active 对应的已提交 Journal 选包，更新中的候选与 Pending 不参与图片选择。
	active, err := s.opt.Store.Active(ctx, id)
	if err != nil {
		return nil, err
	}
	if active == "" {
		return nil, errStaticNotFound
	}
	records, err := s.opt.Store.Journals(ctx, id)
	if err != nil {
		return nil, err
	}
	directory := ""
	for _, record := range records {
		if record.NewGeneration == active && record.Phase == "committed" {
			directory = record.PackageDir
			break
		}
	}
	if directory == "" {
		return nil, errors.New("active generation has no committed package directory")
	}
	root, err := os.OpenRoot(directory)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	// 2. schema 决定可读取的图片白名单，打开后的文件句柄保持到响应结束。
	schemaFile, err := openStaticFile(root.FS(), "config_schema.json")
	if errors.Is(err, fs.ErrNotExist) {
		return nil, errStaticNotFound
	}
	if err != nil {
		return nil, err
	}
	defer schemaFile.close()
	raw, err := io.ReadAll(schemaFile.content)
	if err != nil {
		return nil, err
	}
	schema, problems := configschema.Parse(raw)
	if len(problems) != 0 {
		return nil, fmt.Errorf("invalid active config schema: %v", problems)
	}
	if !slices.Contains(schema.ExampleImages(), name) {
		return nil, errStaticNotFound
	}
	info, err := staticStat(root.FS(), name)
	if errors.Is(err, fs.ErrNotExist) || errors.Is(err, errStaticSymlink) || err == nil && !info.Mode().IsRegular() {
		return nil, errStaticNotFound
	}
	if err != nil {
		return nil, err
	}
	return openStaticFile(root.FS(), name)
}
