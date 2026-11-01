package telegram

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/sdk"
)

// Bot API 公共服务的 getFile 下载上限，见 CONTRACTS §1.4。
const maxDownloadSize = 20 * 1024 * 1024

type telegramFile struct {
	ID       string  `json:"file_id"`
	Size     *uint64 `json:"file_size"`
	MIME     *string `json:"mime_type"`
	Filename *string `json:"file_name"`
	Width    uint64  `json:"width"`
	Height   uint64  `json:"height"`
}

func (m *telegramMessage) mediaSegment() *pb.Segment {
	var file *telegramFile
	switch {
	case len(m.Photo) > 0:
		file = &m.Photo[0]
		for n := range m.Photo {
			candidate := &m.Photo[n]
			if candidate.Width*candidate.Height > file.Width*file.Height {
				file = candidate
			}
		}
	case m.Video != nil:
		file = m.Video
	case m.Animation != nil:
		file = m.Animation
	case m.Document != nil:
		file = m.Document
	case m.Audio != nil:
		file = m.Audio
	case m.Voice != nil:
		file = m.Voice
	default:
		return nil
	}
	incoming := &pb.IncomingMedia{Token: file.ID, Size: file.Size, Mime: file.MIME}
	source := &pb.MediaSource{Source: &pb.MediaSource_IncomingRef{IncomingRef: incoming}}
	switch {
	case len(m.Photo) > 0:
		return &pb.Segment{Data: &pb.Segment_Image{Image: &pb.ImageSegment{Source: source, Caption: m.Caption}}}
	case m.Video != nil || m.Animation != nil:
		return &pb.Segment{Data: &pb.Segment_Video{Video: &pb.VideoSegment{Source: source, Caption: m.Caption}}}
	default:
		return &pb.Segment{Data: &pb.Segment_File{File: &pb.FileSegment{Source: source, Filename: file.Filename, Mime: file.MIME, Caption: m.Caption}}}
	}
}

// ReadPlatformFile 经实例的 HTTP 客户端按需读取平台文件；内容缓存由 Core 持有。
func (i *instance) ReadPlatformFile(ctx context.Context, fileID string, offset uint64, length uint32) ([]byte, uint64, bool, error) {
	// 1. 只接受 getFile 返回的路径，并在下载前检查平台大小上限。
	var file struct {
		Path string  `json:"file_path"`
		Size *uint64 `json:"file_size"`
	}
	if err := i.client.call(ctx, "getFile", struct {
		ID string `json:"file_id"`
	}{fileID}, &file); err != nil {
		var api *apiError
		if errors.As(err, &api) && api.status == http.StatusBadRequest && strings.Contains(api.description, "file is too big") {
			return nil, offset, false, &sdk.Error{Code: pb.ErrorCode_ERROR_CODE_UNSUPPORTED, Message: api.Error()}
		}
		return nil, offset, false, err
	}
	if file.Size != nil && *file.Size > maxDownloadSize {
		return nil, offset, false, downloadUnsupported()
	}
	if file.Path == "" {
		return nil, offset, false, fmt.Errorf("telegram getFile returned no file_path")
	}
	if offset > maxDownloadSize {
		return nil, offset, false, &sdk.Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: "media offset exceeds download limit"}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, i.client.fileBase+file.Path, nil)
	if err != nil {
		return nil, offset, false, i.client.safeError(err)
	}
	response, err := i.client.http.Do(req)
	if err != nil {
		return nil, offset, false, i.client.safeError(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, offset, false, i.client.safeError(err)
		}
		return nil, offset, false, i.client.safeError(fmt.Errorf("telegram download status %d: %s", response.StatusCode, body))
	}
	if response.ContentLength > maxDownloadSize {
		return nil, offset, false, downloadUnsupported()
	}
	// 2. 流式跳过前缀，只保留请求分块和一个 EOF 探测字节，避免 Adapter 再缓存完整对象。
	// shortcut: 分块重新下载前缀；需要优化带宽且平台支持 Range 时改用范围读取。
	if _, err := io.CopyN(io.Discard, response.Body, int64(offset)); err != nil {
		if err == io.EOF {
			return nil, offset, false, &sdk.Error{Code: pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, Message: "media offset exceeds file size"}
		}
		return nil, offset, false, i.client.safeError(err)
	}
	requested := min(uint64(length), maxDownloadSize-offset)
	data, err := io.ReadAll(io.LimitReader(response.Body, int64(requested)+1))
	if err != nil {
		return nil, offset, false, i.client.safeError(err)
	}
	if offset+uint64(len(data)) > maxDownloadSize {
		return nil, offset, false, downloadUnsupported()
	}
	eof := uint64(len(data)) <= requested
	if !eof {
		data = data[:requested]
	}
	return data, offset + uint64(len(data)), eof, nil
}

func downloadUnsupported() error {
	return &sdk.Error{Code: pb.ErrorCode_ERROR_CODE_UNSUPPORTED, Message: "telegram file exceeds 20 MiB download limit"}
}
