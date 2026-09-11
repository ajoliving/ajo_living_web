/*
 * 媒體配置安全邊界測試。
 * 1. 驗證開發與測試環境使用本地 mock 預設。
 * 2. 驗證生產與預發布環境強制要求真實掃描器。
 */
package config

import (
	"os"
	"testing"
)

// 1. TestDefaultMediaProvider verifies development and test attachment defaults.
func TestDefaultMediaProvider(t *testing.T) {
	testCases := []struct {
		name   string
		appEnv string
		want   string
	}{
		{name: "development", appEnv: "development", want: "mock"},
		{name: "test", appEnv: "test", want: "mock"},
		{name: "staging", appEnv: "staging", want: "disabled"},
		{name: "production", appEnv: "production", want: "disabled"},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := defaultMediaProvider(testCase.appEnv); got != testCase.want {
				t.Fatalf("defaultMediaProvider(%q) = %q, want %q", testCase.appEnv, got, testCase.want)
			}
		})
	}
}

// 2. TestValidateRequiresRealScannerInProduction verifies production media cannot use mock or disabled scanning.
func TestValidateRequiresRealScannerInProduction(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatalf("resolve test executable: %v", err)
	}
	testCases := []struct {
		name      string
		config    Config
		wantError bool
	}{
		{name: "development mock", config: Config{AppEnv: "development", MediaProcessingProvider: "mock", MediaScanProvider: "mock"}},
		{name: "production mock scanner", config: Config{AppEnv: "production", MediaProcessingProvider: "ffmpeg", MediaFFmpegBinary: executable, MediaFFmpegArgs: "-i {source_path} {playback_path}", MediaScanProvider: "mock"}, wantError: true},
		{name: "staging disabled scanner", config: Config{AppEnv: "staging", MediaProcessingProvider: "disabled", MediaScanProvider: "disabled"}, wantError: true},
		{name: "production missing scanner command", config: Config{AppEnv: "production", MediaProcessingProvider: "ffmpeg", MediaFFmpegBinary: executable, MediaFFmpegArgs: "-i {source_path} {playback_path}", MediaScanProvider: "clamav"}, wantError: true},
		{name: "production mock processor", config: Config{AppEnv: "production", MediaProcessingProvider: "mock", MediaScanProvider: "clamav", MediaClamAVBinary: executable, MediaClamAVArgs: "--no-summary {source_path}"}, wantError: true},
		{name: "production missing ffmpeg command", config: Config{AppEnv: "production", MediaProcessingProvider: "ffmpeg", MediaScanProvider: "clamav", MediaClamAVBinary: executable, MediaClamAVArgs: "--no-summary {source_path}"}, wantError: true},
		{name: "production unavailable executables", config: Config{AppEnv: "production", MediaProcessingProvider: "ffmpeg", MediaFFmpegBinary: "missing-ffmpeg", MediaFFmpegArgs: "-i {source_path} {playback_path}", MediaScanProvider: "clamav", MediaClamAVBinary: executable, MediaClamAVArgs: "--no-summary {source_path}"}, wantError: true},
		{name: "production ffmpeg and clamav", config: Config{AppEnv: "production", MediaProcessingProvider: "ffmpeg", MediaFFmpegBinary: executable, MediaFFmpegArgs: "-i {source_path} {playback_path}", MediaScanProvider: "clamav", MediaClamAVBinary: executable, MediaClamAVArgs: "--no-summary {source_path}"}},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.config.Validate()
			if (err != nil) != testCase.wantError {
				t.Fatalf("Validate() error = %v, wantError %t", err, testCase.wantError)
			}
		})
	}
}
