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
	"strings"
	"time"
)

var archiveTime = time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC)

func archiveDirectory(root, output string) (err error) {
	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	var zipped *zip.Writer
	var tarred *tar.Writer
	if strings.HasSuffix(output, ".zip") {
		zipped = zip.NewWriter(file)
		defer func() { err = errors.Join(err, zipped.Close()) }()
	} else {
		compressed := gzip.NewWriter(file)
		defer func() { err = errors.Join(err, compressed.Close()) }()
		tarred = tar.NewWriter(compressed)
		defer func() { err = errors.Join(err, tarred.Close()) }()
	}
	// WalkDir 按名称遍历；归档权限和时间不受本机 umask 与文件创建时间影响。
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.IsDir() && !info.Mode().IsRegular() {
			return fmt.Errorf("unsupported distribution file %q", path)
		}
		name, err := filepath.Rel(filepath.Dir(root), path)
		if err != nil {
			return err
		}
		name = filepath.ToSlash(name)
		mode := fs.FileMode(fileMode)
		if entry.IsDir() {
			name += "/"
			mode = directoryMode | fs.ModeDir
		} else if entry.Name() == executableName("linux") || entry.Name() == executableName("windows") {
			mode = directoryMode
		}
		var writer io.Writer
		if zipped != nil {
			header := &zip.FileHeader{Name: name, Method: zip.Deflate, Modified: archiveTime}
			header.SetMode(mode)
			writer, err = zipped.CreateHeader(header)
		} else {
			header := &tar.Header{Name: name, Mode: int64(mode.Perm()), Size: info.Size(), ModTime: archiveTime, Format: tar.FormatPAX, Typeflag: tar.TypeReg}
			if entry.IsDir() {
				header.Typeflag, header.Size = tar.TypeDir, 0
			}
			err = tarred.WriteHeader(header)
			writer = tarred
		}
		if err != nil || entry.IsDir() {
			return err
		}
		return copyTo(writer, path)
	})
}

func copyTo(writer io.Writer, path string) (err error) {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	_, err = io.Copy(writer, file)
	return err
}

func copyFile(source, destination string, mode fs.FileMode) (err error) {
	file, err := os.OpenFile(destination, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	return copyTo(file, source)
}
