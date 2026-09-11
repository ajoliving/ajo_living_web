/*
 * Media processing and malware scanning adapters.
 * 1. Keep video processing and content scanning behind narrow interfaces.
 * 2. Provide disabled, mock, ffmpeg, and ClamAV command implementations.
 * 3. Execute configured commands without a shell and with bounded timeouts.
 */
package service

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path"
	"strings"
	"time"

	"ajoliving_web/http_service/internal/config"
	"ajoliving_web/http_service/internal/model"
)

var (
	// 1. ErrMediaProcessorDisabled identifies an intentionally disabled processor.
	ErrMediaProcessorDisabled = errors.New("media processor is disabled")
	// 2. ErrMediaScannerDisabled identifies an intentionally disabled scanner.
	ErrMediaScannerDisabled = errors.New("media scanner is disabled")
)

// 3. MediaProcessingInput carries the source asset and deterministic output keys.
type MediaProcessingInput struct {
	Asset         model.MediaAsset
	ThumbnailKey  string
	PlaybackKey   string
	SourcePath    string
	ThumbnailPath string
	PlaybackPath  string
}

// 4. MediaProcessingResult reports generated derivative object keys.
type MediaProcessingResult struct {
	ThumbnailObjectKey string
	PlaybackObjectKey  string
}

// 5. MediaProcessor transforms a video into optional thumbnail and playback objects.
type MediaProcessor interface {
	Process(ctx context.Context, input MediaProcessingInput) (MediaProcessingResult, error)
}

// 6. MediaScanInput carries the source asset to a malware scanner.
type MediaScanInput struct {
	Asset      model.MediaAsset
	SourcePath string
}

// 7. MediaScanResult reports scanner verdict and optional reason.
type MediaScanResult struct {
	Passed bool
	Reason string
}

// 8. MediaScanner checks uploaded media for malware or policy violations.
type MediaScanner interface {
	Scan(ctx context.Context, input MediaScanInput) (MediaScanResult, error)
}

// 9. disabledMediaProcessor fails closed when no processor is configured.
type disabledMediaProcessor struct{}

// 10. Process rejects processing instead of exposing an unprocessed video.
func (disabledMediaProcessor) Process(context.Context, MediaProcessingInput) (MediaProcessingResult, error) {
	return MediaProcessingResult{}, ErrMediaProcessorDisabled
}

// 11. disabledMediaScanner fails closed when no scanner is configured.
type disabledMediaScanner struct{}

// 12. Scan rejects scanning instead of marking content as trusted.
func (disabledMediaScanner) Scan(context.Context, MediaScanInput) (MediaScanResult, error) {
	return MediaScanResult{}, ErrMediaScannerDisabled
}

// 13. mockMediaProcessor is deterministic and only intended for tests.
type mockMediaProcessor struct{}

// 14. Process reports successful no-op processing without claiming derivative objects exist.
func (mockMediaProcessor) Process(context.Context, MediaProcessingInput) (MediaProcessingResult, error) {
	return MediaProcessingResult{}, nil
}

// 15. mockMediaScanner is deterministic and only intended for tests.
type mockMediaScanner struct{}

// 16. Scan reports a clean verdict without external side effects.
func (mockMediaScanner) Scan(context.Context, MediaScanInput) (MediaScanResult, error) {
	return MediaScanResult{Passed: true}, nil
}

// 16.1 requiresLocalMediaFile identifies adapters that must receive a downloaded source file.
type requiresLocalMediaFile interface {
	requiresLocalMediaFile() bool
}

// 17. commandMediaProcessor invokes a configured ffmpeg-compatible wrapper.
type commandMediaProcessor struct {
	binary  string
	args    []string
	timeout time.Duration
}

// 18. Process executes ffmpeg arguments with deterministic key placeholders.
func (p commandMediaProcessor) Process(ctx context.Context, input MediaProcessingInput) (MediaProcessingResult, error) {
	if strings.TrimSpace(p.binary) == "" {
		return MediaProcessingResult{}, fmt.Errorf("media processor binary is empty")
	}
	commandCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()
	args := expandMediaCommandArgs(p.args, input.Asset, input.ThumbnailKey, input.PlaybackKey, input.SourcePath, input.ThumbnailPath, input.PlaybackPath)
	if len(args) == 0 {
		return MediaProcessingResult{}, fmt.Errorf("media processor command arguments are empty")
	}
	if output, err := exec.CommandContext(commandCtx, p.binary, args...).CombinedOutput(); err != nil {
		return MediaProcessingResult{}, fmt.Errorf("media processor command failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return MediaProcessingResult{ThumbnailObjectKey: input.ThumbnailKey, PlaybackObjectKey: input.PlaybackKey}, nil
}

// 18.1 requiresLocalMediaFile makes the command receive temporary local file paths instead of OSS keys.
func (commandMediaProcessor) requiresLocalMediaFile() bool { return true }

// 19. commandMediaScanner invokes a configured clamscan-compatible wrapper.
type commandMediaScanner struct {
	binary  string
	args    []string
	timeout time.Duration
}

// 20. Scan treats ClamAV exit code 1 as an infected verdict and other failures as unavailable.
func (p commandMediaScanner) Scan(ctx context.Context, input MediaScanInput) (MediaScanResult, error) {
	if strings.TrimSpace(p.binary) == "" {
		return MediaScanResult{}, fmt.Errorf("media scanner binary is empty")
	}
	commandCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()
	args := expandMediaCommandArgs(p.args, input.Asset, "", "", input.SourcePath, "", "")
	if len(args) == 0 {
		return MediaScanResult{}, fmt.Errorf("media scanner command arguments are empty")
	}
	output, err := exec.CommandContext(commandCtx, p.binary, args...).CombinedOutput()
	if err == nil {
		return MediaScanResult{Passed: true}, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return MediaScanResult{Passed: false, Reason: "malware detected"}, nil
	}
	return MediaScanResult{}, fmt.Errorf("media scanner command failed: %w: %s", err, strings.TrimSpace(string(output)))
}

// 20.1 requiresLocalMediaFile makes the scanner inspect the downloaded source bytes.
func (commandMediaScanner) requiresLocalMediaFile() bool { return true }

// 21. NewMediaProcessor creates the configured processor adapter.
func NewMediaProcessor(cfg *config.Config) (MediaProcessor, error) {
	if cfg == nil {
		return disabledMediaProcessor{}, nil
	}
	switch strings.ToLower(strings.TrimSpace(cfg.MediaProcessingProvider)) {
	case "", "disabled", "off":
		return disabledMediaProcessor{}, nil
	case "mock":
		return mockMediaProcessor{}, nil
	case "ffmpeg", "command":
		args := splitMediaCommandArgs(cfg.MediaFFmpegArgs)
		if strings.TrimSpace(cfg.MediaFFmpegBinary) == "" || len(args) == 0 {
			return nil, fmt.Errorf("MEDIA_FFMPEG_BINARY and MEDIA_FFMPEG_ARGS are required")
		}
		return commandMediaProcessor{binary: cfg.MediaFFmpegBinary, args: args, timeout: positiveMediaTimeout(cfg.MediaProcessingTimeout, 10*time.Minute)}, nil
	default:
		return nil, fmt.Errorf("unsupported MEDIA_PROCESSING_PROVIDER %q", cfg.MediaProcessingProvider)
	}
}

// 22. NewMediaScanner creates the configured scanner adapter.
func NewMediaScanner(cfg *config.Config) (MediaScanner, error) {
	if cfg == nil {
		return disabledMediaScanner{}, nil
	}
	switch strings.ToLower(strings.TrimSpace(cfg.MediaScanProvider)) {
	case "", "disabled", "off":
		return disabledMediaScanner{}, nil
	case "mock":
		return mockMediaScanner{}, nil
	case "clamav", "clamscan", "command":
		args := splitMediaCommandArgs(cfg.MediaClamAVArgs)
		if strings.TrimSpace(cfg.MediaClamAVBinary) == "" || len(args) == 0 {
			return nil, fmt.Errorf("MEDIA_CLAMAV_BINARY and MEDIA_CLAMAV_ARGS are required")
		}
		return commandMediaScanner{binary: cfg.MediaClamAVBinary, args: args, timeout: positiveMediaTimeout(cfg.MediaScanTimeout, 2*time.Minute)}, nil
	default:
		return nil, fmt.Errorf("unsupported MEDIA_SCAN_PROVIDER %q", cfg.MediaScanProvider)
	}
}

// 23. splitMediaCommandArgs tokenizes administrator-provided arguments without shell evaluation.
func splitMediaCommandArgs(raw string) []string {
	return strings.Fields(strings.TrimSpace(raw))
}

// 24. expandMediaCommandArgs replaces supported object-key and local-file placeholders.
func expandMediaCommandArgs(args []string, asset model.MediaAsset, thumbnailKey string, playbackKey string, sourcePath string, thumbnailPath string, playbackPath string) []string {
	values := map[string]string{
		"{object_key}":     asset.ObjectKey,
		"{mime_type}":      asset.MimeType,
		"{thumbnail_key}":  thumbnailKey,
		"{playback_key}":   playbackKey,
		"{source_path}":    sourcePath,
		"{thumbnail_path}": thumbnailPath,
		"{playback_path}":  playbackPath,
	}
	result := make([]string, 0, len(args))
	for _, arg := range args {
		for placeholder, value := range values {
			arg = strings.ReplaceAll(arg, placeholder, value)
		}
		result = append(result, arg)
	}
	return result
}

// 25. mediaDerivativeKeys derives deterministic derivative object keys for a video asset.
func mediaDerivativeKeys(objectKey string) (string, string) {
	base := strings.TrimSuffix(objectKey, path.Ext(objectKey))
	return base + ".thumb.jpg", base + ".playback.mp4"
}

// 26. positiveMediaTimeout applies a safe fallback for invalid configuration.
func positiveMediaTimeout(value time.Duration, fallback time.Duration) time.Duration {
	if value <= 0 {
		return fallback
	}
	return value
}
