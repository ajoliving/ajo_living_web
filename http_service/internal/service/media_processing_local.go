/*
 * 聊天媒體本地檔案處理。
 * 1. 將需命令處理的 OSS 對象下載到受控臨時目錄。
 * 2. 限制本地讀取與衍生檔大小，避免 worker 被大檔案耗盡。
 * 3. 將真實 ffmpeg 產物回傳 OSS 後才允許寫入衍生對象鍵。
 */
package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ajoliving_web/http_service/internal/model"
)

const mediaLocalFileLimit = maxChatAttachmentSize

type localMediaInput struct {
	sourcePath    string
	thumbnailPath string
	playbackPath  string
}

// 1. downloadChatMediaInput retrieves the object through a short-lived signed URL.
func downloadChatMediaInput(ctx context.Context, storage StorageProvider, asset model.MediaAsset) (localMediaInput, func(), error) {
	if storage == nil {
		return localMediaInput{}, func() {}, fmt.Errorf("media storage is not configured")
	}
	directory, err := os.MkdirTemp("", "ajo-chat-media-")
	if err != nil {
		return localMediaInput{}, func() {}, fmt.Errorf("create media temporary directory: %w", err)
	}
	cleanup := func() { _ = os.RemoveAll(directory) }
	sourcePath := filepath.Join(directory, "source"+safeMediaExtension(asset.ObjectKey))
	downloadURL, err := storage.PresignDownload(ctx, asset.ObjectKey, 5*time.Minute)
	if err != nil {
		cleanup()
		return localMediaInput{}, func() {}, fmt.Errorf("create media download url: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		cleanup()
		return localMediaInput{}, func() {}, fmt.Errorf("build media download request: %w", err)
	}
	response, err := (&http.Client{Timeout: 2 * time.Minute}).Do(request)
	if err != nil {
		cleanup()
		return localMediaInput{}, func() {}, fmt.Errorf("download media source: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		cleanup()
		return localMediaInput{}, func() {}, fmt.Errorf("download media source returned status %d", response.StatusCode)
	}
	file, err := os.OpenFile(sourcePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		cleanup()
		return localMediaInput{}, func() {}, fmt.Errorf("create media source file: %w", err)
	}
	written, copyErr := io.Copy(file, io.LimitReader(response.Body, mediaLocalFileLimit+1))
	closeErr := file.Close()
	if copyErr != nil || closeErr != nil {
		cleanup()
		return localMediaInput{}, func() {}, fmt.Errorf("write media source file")
	}
	if written == 0 || written > mediaLocalFileLimit {
		cleanup()
		return localMediaInput{}, func() {}, fmt.Errorf("media source exceeds local processing limit")
	}
	return localMediaInput{
		sourcePath:    sourcePath,
		thumbnailPath: filepath.Join(directory, "thumbnail.jpg"),
		playbackPath:  filepath.Join(directory, "playback.mp4"),
	}, cleanup, nil
}

// 2. uploadMediaFile uploads a bounded derivative and confirms its object exists.
func uploadMediaFile(ctx context.Context, storage StorageProvider, objectKey string, mimeType string, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read generated media: %w", err)
	}
	if len(data) == 0 || int64(len(data)) > mediaLocalFileLimit {
		return fmt.Errorf("generated media exceeds local processing limit")
	}
	if _, err := storage.PutObject(ctx, PutObjectInput{ObjectKey: objectKey, MimeType: mimeType, Body: data}); err != nil {
		return fmt.Errorf("upload generated media: %w", err)
	}
	if _, err := storage.HeadObject(ctx, objectKey); err != nil {
		return fmt.Errorf("verify generated media: %w", err)
	}
	return nil
}

// 3. safeMediaExtension retains only a simple source extension for local ffmpeg type detection.
func safeMediaExtension(objectKey string) string {
	extension := strings.ToLower(filepath.Ext(objectKey))
	if len(extension) > 10 {
		return ".bin"
	}
	for _, character := range extension {
		if character != '.' && (character < 'a' || character > 'z') && (character < '0' || character > '9') {
			return ".bin"
		}
	}
	if extension == "" {
		return ".bin"
	}
	return extension
}
