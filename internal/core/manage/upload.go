package manage

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
)

func multipartUpload(r *http.Request) bool {
	mediaType, _, _ := strings.Cut(r.Header.Get("Content-Type"), ";")
	return strings.EqualFold(strings.TrimSpace(mediaType), "multipart/form-data")
}

func (s *Service) uploadPackage(r *http.Request, update bool) (result any, err error) {
	// 1. 本请求独占上传目录，清理在服务调用结束后、写响应前完成。
	staging := filepath.Join(filepath.Dir(s.opt.Packages), "staging")
	if err := os.MkdirAll(staging, 0o700); err != nil {
		return nil, fmt.Errorf("create upload staging root: %w", err)
	}
	directory, err := os.MkdirTemp(staging, "upload-*")
	if err != nil {
		return nil, fmt.Errorf("create upload directory: %w", err)
	}
	defer func() {
		if cleanupErr := os.RemoveAll(directory); cleanupErr != nil {
			cleanupErr = fmt.Errorf("remove upload directory: %w", cleanupErr)
			slog.Error("clean management upload", "err", cleanupErr)
			err = errors.Join(err, cleanupErr)
		}
	}()

	// 2. 解析只构造本地候选，信任、确认与安装副作用交给同一管理服务。
	request, id, err := s.readUpload(r, directory, update)
	if err != nil {
		return nil, err
	}
	if update {
		return struct{}{}, s.Update(r.Context(), id, request)
	}
	return s.Install(r.Context(), request)
}

func (s *Service) readUpload(r *http.Request, directory string, update bool) (request InstallRequest, id string, err error) {
	// multipart 会把新 part 头部读到的裸 EOF 也当作正常结束。用独立错误区分它；
	// 补一行结尾仍允许合法的、末尾没有 CRLF 的最终边界，文件 part 的字节保持原样。
	framed := *r
	framed.Body = io.NopCloser(uploadEOFReader{io.MultiReader(r.Body, strings.NewReader("\r\n"))})
	reader, err := framed.MultipartReader()
	if err != nil {
		return request, id, fmt.Errorf("%w: %w", invalid("invalid multipart request"), err)
	}
	seen := make(map[string]bool)
	for {
		// NextRawPart 保留文件原始字节，避免按 Content-Transfer-Encoding 改写签名材料。
		part, err := reader.NextRawPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return request, id, fmt.Errorf("%w: %w", invalid("read multipart part"), err)
		}
		disposition, parameters, err := mime.ParseMediaType(part.Header.Get("Content-Disposition"))
		if err != nil {
			return request, id, fmt.Errorf("%w: %w", invalid("invalid multipart disposition"), err)
		}
		if disposition != "form-data" {
			return request, id, invalid("part must have a form-data disposition")
		}
		name := parameters["name"]
		file := name == "archive" || name == "release" || name == "signature"
		text := name == "public_key" || name == "confirm" || (!update && name == "replace") || (update && name == "id")
		if !file && !text {
			return request, id, invalid("unknown multipart part: " + name)
		}
		if seen[name] {
			return request, id, invalid("duplicate multipart part: " + name)
		}
		seen[name] = true
		if _, hasFilename := parameters["filename"]; hasFilename != file {
			return request, id, invalid("incorrect multipart file or text part: " + name)
		}
		if name == "archive" {
			request.Archive, err = writeUploadArchive(part, directory, s.opt.Limits.MaxArchiveBytes)
			if err != nil {
				return request, id, err
			}
			continue
		}
		raw, err := io.ReadAll(part)
		if err != nil {
			return request, id, fmt.Errorf("%w: %w", invalid("read multipart "+name), err)
		}
		value := string(raw)
		switch name {
		case "release":
			request.ReleaseJSON = value
		case "signature":
			request.Signature = value
		case "public_key":
			request.PublicKey = value
		case "id":
			id = value
		case "confirm", "replace":
			if value != "true" && value != "false" {
				return request, id, invalid(name + " must be true or false")
			}
			if name == "confirm" {
				request.Confirm = value == "true"
			} else {
				request.Replace = value == "true"
			}
		}
	}
	if !seen["archive"] || !seen["release"] || (update && id == "") {
		return request, id, invalid("archive, release and an update id are required")
	}
	return request, id, nil
}

type uploadEOFReader struct{ io.Reader }

// Read 让 multipart 的底层输入耗尽与最终边界产生的 io.EOF 保持可区分。
func (r uploadEOFReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	if err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return n, err
}

func writeUploadArchive(content io.Reader, directory string, limit int64) (name string, err error) {
	file, err := os.CreateTemp(directory, "archive-*.tar.gz")
	if err != nil {
		return "", fmt.Errorf("create upload archive: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			closeErr = fmt.Errorf("close upload archive: %w", closeErr)
			slog.Error("close management upload", "err", closeErr)
			err = errors.Join(err, closeErr)
		}
	}()
	n, err := io.Copy(file, io.LimitReader(content, limit+1))
	if err != nil {
		var pathError *os.PathError
		if errors.As(err, &pathError) {
			return "", fmt.Errorf("write upload archive: %w", err)
		}
		return "", fmt.Errorf("%w: %w", invalid("read upload archive"), err)
	}
	if n > limit {
		return "", &validate.Error{Code: pb.ErrorCode_ERROR_CODE_RESOURCE_EXHAUSTED, Message: "archive exceeds the size limit"}
	}
	return file.Name(), nil
}
