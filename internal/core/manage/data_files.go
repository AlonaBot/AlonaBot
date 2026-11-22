package manage

import (
	"archive/zip"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type dataDownload struct {
	file      *os.File
	name      string
	mediaType string
	cleanup   func() error // 导出临时资源在响应结束后删除。
}

func (download *dataDownload) serve(w http.ResponseWriter, r *http.Request) {
	defer func() {
		err := download.file.Close()
		if download.cleanup != nil {
			err = errors.Join(err, download.cleanup())
		}
		if err != nil {
			slog.ErrorContext(r.Context(), "close plugin data download failed", "error", err)
		}
	}()
	info, err := download.file.Stat()
	if err != nil {
		respond(w, nil, err)
		return
	}
	w.Header().Set("Content-Type", download.mediaType)
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": download.name}))
	w.Header().Set("Content-Length", fmt.Sprint(info.Size()))
	if r.Method == "HEAD" {
		return
	}
	if _, err := io.Copy(w, io.NewSectionReader(download.file, 0, info.Size())); err != nil {
		slog.ErrorContext(r.Context(), "write plugin data download failed", "error", err)
	}
}

func dataFilePath(name string, allowRoot bool) error {
	if name == "" && allowRoot {
		return nil
	}
	if !fs.ValidPath(name) || name == "." || strings.ContainsAny(name, "\\:\x00") || filepath.IsAbs(name) {
		return invalid("invalid relative file path: " + name)
	}
	return nil
}

// dataFileStat 检查已固定父目录中的单个名称。
func dataFileStat(root *os.Root, name string) (fs.FileInfo, error) {
	info, err := root.Lstat(name)
	if err == nil && info.Mode()&os.ModeSymlink != 0 {
		return nil, invalid("symbolic link is not allowed: " + name)
	}
	return info, err
}

// dataOpenDirectory 将检查结果与打开后的目录身份绑定，拒绝检查后替换的链接或目录。
func dataOpenDirectory(root *os.Root, name string, checked fs.FileInfo) (*os.Root, error) {
	directory, err := root.OpenRoot(name)
	if err != nil {
		return nil, err
	}
	opened, err := directory.Stat(".")
	if err == nil && !os.SameFile(checked, opened) {
		err = invalid("directory changed while opening: " + name)
	}
	if err != nil {
		return nil, errors.Join(err, directory.Close())
	}
	return directory, nil
}

// dataDirectory 逐段固定目录，后续操作只相对最后一个句柄执行。
func dataDirectory(root *os.Root, name string) (*os.Root, error) {
	directory, err := root.OpenRoot(".")
	if err != nil {
		return nil, err
	}
	if name == "" || name == "." {
		return directory, nil
	}
	for _, part := range strings.Split(name, "/") {
		info, err := dataFileStat(directory, part)
		if err != nil {
			return nil, errors.Join(err, directory.Close())
		}
		next, openErr := dataOpenDirectory(directory, part, info)
		err = errors.Join(openErr, directory.Close())
		if err != nil {
			if next != nil {
				err = errors.Join(err, next.Close())
			}
			return nil, err
		}
		directory = next
	}
	return directory, nil
}

func dataOpenFile(root *os.Root, name string, checked fs.FileInfo) (*os.File, error) {
	if !checked.Mode().IsRegular() {
		return nil, invalid("file is not a regular file: " + name)
	}
	file, err := root.Open(name)
	if err != nil {
		return nil, err
	}
	opened, err := file.Stat()
	if err == nil && !os.SameFile(checked, opened) {
		err = invalid("file changed while opening: " + name)
	}
	if err != nil {
		return nil, errors.Join(err, file.Close())
	}
	return file, nil
}

func dataFilesRoot(selected *dataSelection) (*os.Root, error) {
	root, err := os.OpenRoot(filepath.Dir(filepath.Dir(selected.directory)))
	if err != nil {
		return nil, err
	}
	directory, err := dataDirectory(root, "plugin_data/"+selected.id)
	if err = errors.Join(err, root.Close()); err != nil {
		if directory != nil {
			err = errors.Join(err, directory.Close())
		}
		return nil, err
	}
	return directory, nil
}

func (s *Service) dataFiles(r *http.Request, selected *dataSelection) (result any, err error) {
	name := r.URL.Query().Get("path")
	if err := dataFilePath(name, true); err != nil {
		return nil, err
	}
	root, err := dataFilesRoot(selected)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, root.Close()) }()
	directory, err := dataDirectory(root, name)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, directory.Close()) }()
	entries, err := fs.ReadDir(directory.FS(), ".")
	if err != nil {
		return nil, err
	}
	type entry struct {
		Name       string    `json:"name"`
		Dir        bool      `json:"dir"`
		Size       int64     `json:"size"`
		ModifiedAt time.Time `json:"modified_at"`
	}
	listed := make([]entry, 0, len(entries))
	for _, item := range entries {
		info, err := item.Info()
		if err != nil {
			return nil, err
		}
		listed = append(listed, entry{Name: item.Name(), Dir: info.IsDir(), Size: info.Size(), ModifiedAt: info.ModTime().UTC()})
	}
	slices.SortFunc(listed, func(a, b entry) int {
		if a.Dir != b.Dir {
			if a.Dir {
				return -1
			}
			return 1
		}
		return strings.Compare(a.Name, b.Name)
	})
	return struct {
		Entries []entry `json:"entries"`
	}{listed}, nil
}

func (s *Service) dataContent(r *http.Request, selected *dataSelection) (result any, err error) {
	name := r.URL.Query().Get("path")
	if err := dataFilePath(name, false); err != nil {
		return nil, err
	}
	root, err := dataFilesRoot(selected)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, root.Close()) }()
	if r.Method == "GET" || r.Method == "HEAD" {
		directory, openErr := dataDirectory(root, path.Dir(name))
		if openErr != nil {
			return nil, openErr
		}
		defer func() { err = errors.Join(err, directory.Close()) }()
		info, statErr := dataFileStat(directory, path.Base(name))
		if statErr != nil {
			return nil, statErr
		}
		file, openErr := dataOpenFile(directory, path.Base(name), info)
		if openErr != nil {
			return nil, openErr
		}
		return &dataDownload{file: file, name: path.Base(name), mediaType: "application/octet-stream"}, nil
	}
	if err := dataUpload(root, name, r.Body, r.URL.Query().Get("overwrite") == "1"); err != nil {
		return nil, err
	}
	selected.audit = []any{"plugin_id", selected.id, "path", name}
	return struct{}{}, nil
}

func dataUpload(root *os.Root, name string, body io.Reader, overwrite bool) (err error) {
	// 1. 核对父目录与目标，之后的临时文件和发布共用父目录句柄。
	parent := path.Dir(name)
	directory, err := dataDirectory(root, parent)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, directory.Close()) }()
	target := path.Base(name)
	info, err := dataFileStat(directory, target)
	if err == nil {
		if !overwrite {
			return os.ErrExist
		}
		if !info.Mode().IsRegular() {
			return invalid("upload target is not a regular file")
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	// 2. 正文完整写入并关闭后发布，失败只清理本次临时文件。
	temporary := ".upload-" + rand.Text()
	file, err := directory.OpenFile(temporary, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer func() {
		if cleanup := directory.Remove(temporary); !errors.Is(cleanup, os.ErrNotExist) {
			err = errors.Join(err, cleanup)
		}
	}()
	_, copyErr := io.Copy(file, body)
	if err := errors.Join(copyErr, file.Close()); err != nil {
		return err
	}
	if overwrite {
		return directory.Rename(temporary, target)
	}
	// 硬链接发布具有独占语义，避免检查后新出现的文件被 rename 覆盖。
	return directory.Link(temporary, target)
}

func (s *Service) dataFileAction(r *http.Request, selected *dataSelection) (result any, err error) {
	action := path.Base(r.URL.Path)
	var request struct {
		Path      string `json:"path"`
		From      string `json:"from"`
		To        string `json:"to"`
		Recursive bool   `json:"recursive"`
	}
	if err := decode(r, &request); err != nil {
		return nil, err
	}
	names := []string{request.Path}
	if action == "rename" {
		names = []string{request.From, request.To}
	}
	for _, name := range names {
		if err := dataFilePath(name, false); err != nil {
			return nil, err
		}
	}
	root, err := dataFilesRoot(selected)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, root.Close()) }()
	switch action {
	case "mkdir":
		directory, openErr := dataDirectory(root, path.Dir(request.Path))
		if openErr != nil {
			return nil, openErr
		}
		defer func() { err = errors.Join(err, directory.Close()) }()
		name := path.Base(request.Path)
		if _, err := dataFileStat(directory, name); err == nil {
			return nil, os.ErrExist
		} else if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		if err := directory.Mkdir(name, 0o700); err != nil {
			return nil, err
		}
	case "rename":
		from, openErr := dataDirectory(root, path.Dir(request.From))
		if openErr != nil {
			return nil, openErr
		}
		defer func() { err = errors.Join(err, from.Close()) }()
		if _, err := dataFileStat(from, path.Base(request.From)); err != nil {
			return nil, err
		}
		to, openErr := dataDirectory(root, path.Dir(request.To))
		if openErr != nil {
			return nil, openErr
		}
		defer func() { err = errors.Join(err, to.Close()) }()
		if _, err := dataFileStat(to, path.Base(request.To)); err == nil {
			return nil, os.ErrExist
		} else if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		if selected.beforeDataRename != nil {
			selected.beforeDataRename()
		}
		if err := dataRename(from, path.Base(request.From), to, path.Base(request.To)); err != nil {
			return nil, err
		}
	case "delete":
		directory, openErr := dataDirectory(root, path.Dir(request.Path))
		if openErr != nil {
			return nil, openErr
		}
		defer func() { err = errors.Join(err, directory.Close()) }()
		name := path.Base(request.Path)
		if request.Recursive {
			if err := dataWalk(directory, name, func(*os.Root, string, fs.FileInfo) error { return nil }); err != nil {
				return nil, err
			}
		}
		if err := dataRemove(directory, name, request.Recursive); err != nil {
			return nil, err
		}
	}
	selected.audit = []any{"plugin_id", selected.id, "operation", action}
	for i, name := range names {
		label := "path"
		if action == "rename" {
			label = []string{"from", "to"}[i]
		}
		selected.audit = append(selected.audit, label, name)
	}
	return struct{}{}, nil
}

func dataRename(from *os.Root, fromName string, to *os.Root, toName string) (err error) {
	source, err := from.Open(".")
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, source.Close()) }()
	target, err := to.Open(".")
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, target.Close()) }()
	if err := dataRenameAt(source, fromName, target, toName); err != nil {
		return &os.LinkError{Op: "rename-noreplace", Old: fromName, New: toName, Err: err}
	}
	return nil
}

// dataWalk 的回调对文件收到父目录句柄，对目录收到已核对身份的自身句柄。
func dataWalk(root *os.Root, name string, visit func(*os.Root, string, fs.FileInfo) error) (err error) {
	info, err := dataFileStat(root, name)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return visit(root, name, info)
	}
	directory, err := dataOpenDirectory(root, name, info)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, directory.Close()) }()
	if err := visit(directory, name, info); err != nil {
		return err
	}
	entries, err := fs.ReadDir(directory.FS(), ".")
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := dataWalk(directory, entry.Name(), func(parent *os.Root, child string, info fs.FileInfo) error {
			return visit(parent, path.Join(name, child), info)
		}); err != nil {
			return err
		}
	}
	return nil
}

func dataArchive(ctx context.Context, root *os.Root, snapshot string, destination io.Writer) (err error) {
	archive := zip.NewWriter(destination)
	defer func() { err = errors.Join(err, archive.Close()) }()
	database, err := os.Open(snapshot)
	if err != nil {
		return err
	}
	entry, err := archive.Create("db.sqlite")
	if err != nil {
		return errors.Join(err, database.Close())
	}
	_, copyErr := io.Copy(entry, database)
	if err := errors.Join(copyErr, database.Close()); err != nil {
		return err
	}
	return dataWalk(root, ".", func(directory *os.Root, name string, info fs.FileInfo) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if !info.IsDir() && !info.Mode().IsRegular() {
			return invalid("file is not a regular file: " + name)
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = path.Join("plugin_data", name)
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}
		entry, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		file, err := dataOpenFile(directory, path.Base(name), info)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(entry, file)
		return errors.Join(copyErr, file.Close())
	})
}

func dataClearFiles(root *os.Root) error {
	if err := dataWalk(root, ".", func(*os.Root, string, fs.FileInfo) error { return nil }); err != nil {
		return err
	}
	entries, err := fs.ReadDir(root.FS(), ".")
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := dataRemove(root, entry.Name(), true); err != nil {
			return err
		}
	}
	return nil
}

func dataRemove(root *os.Root, name string, recursive bool) (err error) {
	info, err := dataFileStat(root, name)
	if err != nil {
		return err
	}
	if !info.IsDir() || !recursive {
		return root.Remove(name)
	}
	directory, err := dataOpenDirectory(root, name, info)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, directory.Close()) }()
	entries, err := fs.ReadDir(directory.FS(), ".")
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := dataRemove(directory, entry.Name(), true); err != nil {
			return err
		}
	}
	return root.Remove(name)
}
