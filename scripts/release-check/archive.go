package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode/utf8"
)

type archiveEntry struct {
	name      string
	directory bool
	header    bool // 隐式父目录尚未出现归档头。
}

// extractArchive 只写入调用者新建的目录，保留失败现场，拒绝链接及不安全路径。
func extractArchive(archive, destination string) ([]string, error) {
	if err := os.Mkdir(destination, directoryMode); err != nil {
		return nil, err
	}
	seen := map[string]archiveEntry{}
	var files []string
	write := func(name string, mode fs.FileMode, body io.Reader) error {
		directory := mode.IsDir()
		if directory {
			name = strings.TrimSuffix(name, "/")
		}
		if !utf8.ValidString(name) || !fs.ValidPath(name) || name == "." || strings.ContainsAny(name, "\\:\x00") {
			return fmt.Errorf("unsafe archive path %q", name)
		}
		if !directory && !mode.IsRegular() {
			return fmt.Errorf("unsupported archive entry %q: %s", name, mode)
		}
		// 父目录也登记大小写，避免不同平台对同一资产作出不同解释。
		parts := strings.Split(name, "/")
		for i := range parts {
			prefix := strings.Join(parts[:i+1], "/")
			key := strings.ToLower(prefix)
			last := i == len(parts)-1
			isDir := !last || directory
			previous, exists := seen[key]
			if exists && (previous.name != prefix || previous.directory != isDir || last && previous.header) {
				return fmt.Errorf("duplicate or conflicting archive path %q (previous %q)", name, previous.name)
			}
			seen[key] = archiveEntry{prefix, isDir, previous.header || last}
		}
		dest := filepath.Join(destination, filepath.FromSlash(name))
		if directory {
			return os.MkdirAll(dest, directoryMode)
		}
		if err := os.MkdirAll(filepath.Dir(dest), directoryMode); err != nil {
			return err
		}
		permissions := os.FileMode(fileMode)
		if mode.Perm()&0o111 != 0 {
			permissions = executableMode
		}
		file, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_EXCL, permissions)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(file, body)
		if err := errors.Join(copyErr, file.Close()); err != nil {
			return err
		}
		files = append(files, name)
		return nil
	}
	if strings.HasSuffix(archive, ".zip") {
		reader, err := zip.OpenReader(archive)
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		for _, file := range reader.File {
			body, err := file.Open()
			if err != nil {
				return nil, err
			}
			err = errors.Join(write(file.Name, file.Mode(), body), body.Close())
			if err != nil {
				return nil, err
			}
		}
	} else {
		file, err := os.Open(archive)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		compressed, err := gzip.NewReader(file)
		if err != nil {
			return nil, err
		}
		defer compressed.Close()
		reader := tar.NewReader(compressed)
		for {
			header, err := reader.Next()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				return nil, err
			}
			if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeDir {
				return nil, fmt.Errorf("unsupported tar entry %q: type %d", header.Name, header.Typeflag)
			}
			if err := write(header.Name, header.FileInfo().Mode(), reader); err != nil {
				return nil, err
			}
		}
		// tar EOF 可以早于 gzip 尾部，继续读完以检查压缩流的 CRC。
		if _, err := io.Copy(io.Discard, compressed); err != nil {
			return nil, err
		}
	}
	slices.Sort(files)
	return files, nil
}

func sameFiles(actual, expected []string) error {
	slices.Sort(expected)
	if !slices.Equal(actual, expected) {
		return fmt.Errorf("archive files = %q; want %q", actual, expected)
	}
	return nil
}
