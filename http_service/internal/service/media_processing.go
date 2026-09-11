/*
 * 媒體處理與安全狀態 worker。
 * 1. 為聊天媒體保存 processing、scan 與 rejection 狀態。
 * 2. 對已完成上傳的圖片、文件與視頻執行存在性校驗，避免消息引用失效物件。
 * 3. 將真正的轉碼、病毒掃描接入點保持在獨立 worker 邊界。
 */
package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/utils"
	"gorm.io/gorm"
)

const (
	mediaProcessingReady      = "ready"
	mediaProcessingProcessing = "processing"
	mediaProcessingRejected   = "rejected"
	mediaScanPending          = "pending"
	mediaScanPassed           = "passed"
	mediaScanRejected         = "rejected"
)

// 1. normalizedProcessingStatus returns a stable status for legacy rows.
func normalizedProcessingStatus(asset model.MediaAsset) string {
	return normalizedProcessingStatusValues(asset.ProcessingStatus, asset.MimeType)
}

// 2. normalizedProcessingStatusValues fills status defaults for old media rows.
func normalizedProcessingStatusValues(status string, mimeType string) string {
	if strings.TrimSpace(status) != "" {
		return strings.TrimSpace(status)
	}
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(mimeType)), "video/") {
		return mediaProcessingProcessing
	}
	return mediaProcessingReady
}

// 3. normalizedScanStatus returns a stable scan status for legacy rows.
func normalizedScanStatus(asset model.MediaAsset) string {
	if strings.HasPrefix(strings.TrimSpace(asset.ObjectKey), chatMediaObjectPrefix) && strings.TrimSpace(asset.ScanStatus) == "" {
		return mediaScanPending
	}
	return normalizedScanStatusValue(asset.ScanStatus)
}

// 4. normalizedScanStatusValue fills status defaults without hiding rejected rows.
func normalizedScanStatusValue(status string) string {
	if strings.TrimSpace(status) == "" {
		return mediaScanPassed
	}
	return strings.TrimSpace(status)
}

// 5. MediaProcessingWorker validates pending chat media on a bounded interval.
type MediaProcessingWorker struct {
	runtime        *Runtime
	mediaProcessor MediaProcessor
	mediaScanner   MediaScanner
	setupError     error
}

// 6. NewMediaProcessingWorker creates the media status worker.
func NewMediaProcessingWorker(runtime *Runtime) *MediaProcessingWorker {
	worker := &MediaProcessingWorker{runtime: runtime}
	if runtime == nil {
		return worker
	}
	worker.mediaProcessor = runtime.MediaProcessor
	worker.mediaScanner = runtime.MediaScanner
	if worker.mediaProcessor == nil {
		worker.mediaProcessor, worker.setupError = NewMediaProcessor(runtime.Config)
	}
	if worker.mediaScanner == nil {
		var err error
		worker.mediaScanner, err = NewMediaScanner(runtime.Config)
		if worker.setupError == nil {
			worker.setupError = err
		}
	}
	return worker
}

// 7. Run validates pending media until cancellation.
func (w *MediaProcessingWorker) Run(ctx context.Context) {
	if w == nil || w.runtime == nil || w.runtime.DB == nil || w.runtime.StorageProvider == nil {
		return
	}
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		w.process(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// 8. process verifies storage objects and advances safe media states.
func (w *MediaProcessingWorker) process(ctx context.Context) {
	if !w.runtime.DB.Migrator().HasTable(&model.MediaAsset{}) {
		return
	}
	claimExpiredAt := w.runtime.Now().Add(-w.mediaClaimLease())
	var assets []model.MediaAsset
	if err := w.runtime.DB.WithContext(ctx).Where("object_key LIKE ? AND (processing_status = ? OR scan_status = ? OR processing_status = '' OR scan_status = '') AND (worker_claimed_at IS NULL OR worker_claimed_at <= ?)", chatMediaObjectPrefix+"%", mediaProcessingProcessing, mediaScanPending, claimExpiredAt).Order("id asc").Limit(50).Find(&assets).Error; err != nil {
		return
	}
	for _, asset := range assets {
		claimToken, claimed := w.claimAsset(ctx, asset.ID, claimExpiredAt)
		if !claimed {
			continue
		}
		updates := w.processAsset(ctx, asset)
		updates["worker_claim_token"] = ""
		updates["worker_claimed_at"] = nil
		if err := w.persistAssetUpdates(ctx, asset.ID, claimToken, updates); err != nil && w.runtime.Logger != nil {
			w.runtime.Logger.Warn("persist media processing result", "asset_id", asset.ID, "error", err)
		}
	}
}

// 8.1 claimAsset makes one worker responsible for an asset until its bounded lease expires.
func (w *MediaProcessingWorker) claimAsset(ctx context.Context, assetID int64, claimExpiredAt time.Time) (string, bool) {
	claimToken := utils.NewPublicID()
	result := w.runtime.DB.WithContext(ctx).Model(&model.MediaAsset{}).Where("id = ? AND (worker_claimed_at IS NULL OR worker_claimed_at <= ?)", assetID, claimExpiredAt).Updates(map[string]any{"worker_claim_token": claimToken, "worker_claimed_at": w.runtime.Now()})
	return claimToken, result.Error == nil && result.RowsAffected == 1
}

// 8.2 mediaClaimLease allows a replacement worker only after the configured command windows have elapsed.
func (w *MediaProcessingWorker) mediaClaimLease() time.Duration {
	lease := 15 * time.Minute
	if w == nil || w.runtime == nil || w.runtime.Config == nil {
		return lease
	}
	configured := w.runtime.Config.MediaProcessingTimeout + w.runtime.Config.MediaScanTimeout + 5*time.Minute
	if configured > lease {
		return configured
	}
	return lease
}

// 8.3 persistAssetUpdates commits a media state and its realtime events together.
func (w *MediaProcessingWorker) persistAssetUpdates(ctx context.Context, assetID int64, claimToken string, updates map[string]any) error {
	return w.runtime.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&model.MediaAsset{}).Where("id = ? AND worker_claim_token = ?", assetID, claimToken).Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}
		return w.enqueueAttachmentUpdates(ctx, tx, assetID)
	})
}

// 9. processAsset runs storage verification, video processing, and malware scanning in order.
func (w *MediaProcessingWorker) processAsset(ctx context.Context, asset model.MediaAsset) map[string]any {
	updates := map[string]any{}
	if w.setupError != nil {
		return mediaRejectedUpdates("media adapter configuration error: " + w.setupError.Error())
	}
	if _, err := w.runtime.StorageProvider.HeadObject(ctx, asset.ObjectKey); err != nil {
		return mediaRejectedUpdates("media object is unavailable")
	}
	localInput, cleanup, err := w.prepareLocalMediaInput(ctx, asset)
	if err != nil {
		return mediaRejectedUpdates("media source preparation failed: " + err.Error())
	}
	defer cleanup()
	processingStatus := normalizedProcessingStatus(asset)
	if isVideoMedia(asset.MimeType) && processingStatus == mediaProcessingProcessing {
		thumbnailKey, playbackKey := mediaDerivativeKeys(asset.ObjectKey)
		result, err := w.mediaProcessor.Process(ctx, MediaProcessingInput{Asset: asset, ThumbnailKey: thumbnailKey, PlaybackKey: playbackKey, SourcePath: localInput.sourcePath, ThumbnailPath: localInput.thumbnailPath, PlaybackPath: localInput.playbackPath})
		if err != nil {
			return mediaRejectedUpdates("video processing failed: " + err.Error())
		}
		if err := w.uploadVideoDerivatives(ctx, result, localInput); err != nil {
			return mediaRejectedUpdates("video derivative upload failed: " + err.Error())
		}
		if result.ThumbnailObjectKey != "" {
			if _, err := w.runtime.StorageProvider.HeadObject(ctx, result.ThumbnailObjectKey); err != nil {
				return mediaRejectedUpdates("video thumbnail is unavailable")
			}
			updates["thumbnail_object_key"] = result.ThumbnailObjectKey
		}
		if result.PlaybackObjectKey != "" {
			if _, err := w.runtime.StorageProvider.HeadObject(ctx, result.PlaybackObjectKey); err != nil {
				return mediaRejectedUpdates("video playback version is unavailable")
			}
			updates["playback_object_key"] = result.PlaybackObjectKey
		}
		updates["processing_status"] = mediaProcessingReady
	}
	scanStatus := normalizedScanStatus(asset)
	if strings.TrimSpace(asset.ScanStatus) == "" {
		scanStatus = mediaScanPending
	}
	if scanStatus == mediaScanPending {
		result, err := w.mediaScanner.Scan(ctx, MediaScanInput{Asset: asset, SourcePath: localInput.sourcePath})
		if err != nil {
			return mediaRejectedUpdates("media scan failed: " + err.Error())
		}
		if !result.Passed {
			reason := result.Reason
			if strings.TrimSpace(reason) == "" {
				reason = "media scan rejected"
			}
			return mediaRejectedUpdates(reason)
		}
		updates["scan_status"] = mediaScanPassed
	}
	if len(updates) == 0 {
		return updates
	}
	if _, exists := updates["processing_status"]; !exists {
		updates["processing_status"] = processingStatus
	}
	if _, exists := updates["scan_status"]; !exists {
		updates["scan_status"] = scanStatus
	}
	return updates
}

// 9.1 prepareLocalMediaInput downloads media only when a command adapter needs local bytes.
func (w *MediaProcessingWorker) prepareLocalMediaInput(ctx context.Context, asset model.MediaAsset) (localMediaInput, func(), error) {
	if !mediaAdapterRequiresLocalFile(w.mediaProcessor) && !mediaAdapterRequiresLocalFile(w.mediaScanner) {
		return localMediaInput{}, func() {}, nil
	}
	return downloadChatMediaInput(ctx, w.runtime.StorageProvider, asset)
}

// 9.2 uploadVideoDerivatives persists actual ffmpeg output before exposing derivative keys.
func (w *MediaProcessingWorker) uploadVideoDerivatives(ctx context.Context, result MediaProcessingResult, input localMediaInput) error {
	uploadedKeys := make([]string, 0, 2)
	if result.ThumbnailObjectKey != "" {
		if err := uploadMediaFile(ctx, w.runtime.StorageProvider, result.ThumbnailObjectKey, "image/jpeg", input.thumbnailPath); err != nil {
			return err
		}
		uploadedKeys = append(uploadedKeys, result.ThumbnailObjectKey)
	}
	if result.PlaybackObjectKey != "" {
		if err := uploadMediaFile(ctx, w.runtime.StorageProvider, result.PlaybackObjectKey, "video/mp4", input.playbackPath); err != nil {
			for _, objectKey := range uploadedKeys {
				_ = w.runtime.StorageProvider.DeleteObject(ctx, objectKey)
			}
			return err
		}
	}
	return nil
}

// 9.3 mediaAdapterRequiresLocalFile keeps mock and disabled adapters free from storage downloads.
func mediaAdapterRequiresLocalFile(adapter any) bool {
	requirement, ok := adapter.(requiresLocalMediaFile)
	return ok && requirement.requiresLocalMediaFile()
}

// 10. mediaRejectedUpdates builds a consistent fail-closed state update.
func mediaRejectedUpdates(reason string) map[string]any {
	return map[string]any{"processing_status": mediaProcessingRejected, "scan_status": mediaScanRejected, "rejection_reason": truncateMediaReason(reason)}
}

// 11.1 truncateMediaReason bounds command output persisted in the varchar(500) field.
func truncateMediaReason(reason string) string {
	cleaned := strings.TrimSpace(reason)
	if len([]rune(cleaned)) <= 500 {
		return cleaned
	}
	return string([]rune(cleaned)[:500])
}

// 12. isVideoMedia reports whether derivatives are required.
func isVideoMedia(mimeType string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(mimeType)), "video/")
}

// 13. enqueueAttachmentUpdates persists status events for the realtime outbox worker.
func (w *MediaProcessingWorker) enqueueAttachmentUpdates(ctx context.Context, tx *gorm.DB, assetID int64) error {
	if !tx.Migrator().HasTable(&model.RealtimeOutbox{}) || !tx.Migrator().HasTable(&model.MessageAttachment{}) {
		return nil
	}
	var rows []struct {
		MessageID    int64
		ChatPublicID string
	}
	if err := tx.WithContext(ctx).Table("message_attachments").Select("message_attachments.message_id, chats.public_id AS chat_public_id").Joins("JOIN messages ON messages.id = message_attachments.message_id").Joins("JOIN chats ON chats.id = messages.chat_id").Where("message_attachments.media_asset_id = ?", assetID).Scan(&rows).Error; err != nil {
		return err
	}
	var asset model.MediaAsset
	if err := tx.WithContext(ctx).First(&asset, assetID).Error; err != nil {
		return err
	}
	runtimeCopy := *w.runtime
	runtimeCopy.DB = tx
	chatService := &ChatService{runtime: &runtimeCopy}
	for _, row := range rows {
		var message model.Message
		if err := tx.WithContext(ctx).First(&message, row.MessageID).Error; err != nil {
			return err
		}
		items, err := chatService.buildMessageResponses(ctx, []model.Message{message})
		if err != nil {
			return err
		}
		payload, err := buildRealtimeMessagePayload(row.ChatPublicID, items[0], "")
		if err != nil {
			return err
		}
		mediaStatus := normalizedProcessingStatus(asset) + "." + normalizedScanStatus(asset)
		aggregateID := fmt.Sprintf("attachment:%d:%d:%s", asset.ID, message.ID, mediaStatus)
		now := w.runtime.Now()
		if err := tx.Create(&model.RealtimeOutbox{EventType: "attachment", AggregateID: aggregateID, Payload: payload, Status: "pending", NextAttemptAt: now, CreatedAt: now}).Error; err != nil {
			return err
		}
	}
	return nil
}
