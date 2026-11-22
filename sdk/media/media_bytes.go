package media

import (
	"bytes"
	"context"
	"encoding/base64"

	"github.com/AlonaBot/AlonaBot/sdk/errcode"
	"github.com/AlonaBot/AlonaBot/sdk/message"
)

const maxInlineMediaBytes = 1 << 20 // 1 MiB 以内直接内联。

// Image 将非空图片字节内联，超过 1 MiB 时先上传。
func Image(ctx context.Context, data []byte) (message.Segment, error) {
	source, err := bytesSource(ctx, data, "")
	if err != nil {
		return nil, err
	}
	return message.Image(source), nil
}

// Video 将非空视频字节内联，超过 1 MiB 时先上传。
func Video(ctx context.Context, data []byte) (message.Segment, error) {
	source, err := bytesSource(ctx, data, "")
	if err != nil {
		return nil, err
	}
	return message.Video(source), nil
}

// File 将非空文件字节内联或上传，保留展示名称与 MIME。
func File(ctx context.Context, data []byte, filename, mime string) (message.Segment, error) {
	source, err := bytesSource(ctx, data, mime)
	if err != nil {
		return nil, err
	}
	file := message.File(source, filename).(message.FileSegment)
	file.MIME = mime
	return file, nil
}

func bytesSource(ctx context.Context, data []byte, mime string) (message.MediaSource, error) {
	if len(data) == 0 {
		return message.MediaSource{}, errcode.InvalidArgument.New("media bytes are empty")
	}
	if len(data) <= maxInlineMediaBytes {
		return message.MediaSource{InlineBase64: base64.StdEncoding.EncodeToString(data)}, nil
	}
	ref, err := Upload(ctx, bytes.NewReader(data), mime)
	if err != nil {
		return message.MediaSource{}, err
	}
	return ref.Source(), nil
}
