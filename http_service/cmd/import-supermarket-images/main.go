/*
 * 超市商品圖片導入命令。
 * 1. 掃描本地商品圖片並解析商品代號與實際媒體格式。
 * 2. 按固定 object_key 上傳至 OSS 或本地 storage provider。
 * 3. 幂等新增或顯式替換 media_assets，供超市優惠接口按商品代號回填圖片。
 */
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"ajoliving_web/http_service/internal/config"
	"ajoliving_web/http_service/internal/database"
	"ajoliving_web/http_service/internal/logger"
	"ajoliving_web/http_service/internal/model"
	"ajoliving_web/http_service/internal/service"
	"ajoliving_web/http_service/internal/utils"
)

const supermarketImageObjectPrefix = "ajo_living/supermarket/products/"

type importStats struct {
	Scanned  int
	Uploaded int
	Replaced int
	Skipped  int
	Failed   int
}

// 1. main imports local supermarket product images into configured storage.
func main() {
	dir := flag.String("dir", "../pns_images", "local supermarket image directory")
	operatorUserID := flag.Int64("user", 0, "media asset created_by user id")
	replace := flag.Bool("replace", false, "replace existing objects and media asset metadata")
	dryRun := flag.Bool("dry-run", false, "scan files without uploading or writing database")
	flag.Parse()

	cfg := config.Load()
	logg := logger.New(cfg)
	db, err := database.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatal(err)
	}

	storageProvider, err := service.NewStorageProvider(cfg)
	if err != nil {
		log.Fatal(err)
	}
	runtime := &service.Runtime{
		Config:          cfg,
		DB:              db,
		Logger:          logg,
		StorageProvider: storageProvider,
		Now:             time.Now,
	}

	userID := *operatorUserID
	if userID <= 0 {
		userID = cfg.SystemUserID
	}
	if userID <= 0 {
		userID = 1
	}

	stats, err := importSupermarketImages(context.Background(), runtime, strings.TrimSpace(*dir), userID, *replace, *dryRun)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("scanned=%d uploaded=%d replaced=%d skipped=%d failed=%d\n", stats.Scanned, stats.Uploaded, stats.Replaced, stats.Skipped, stats.Failed)
}

// 2. importSupermarketImages scans and imports all supported product images.
func importSupermarketImages(ctx context.Context, runtime *service.Runtime, dir string, userID int64, replace bool, dryRun bool) (importStats, error) {
	stats := importStats{}
	resolvedDir, err := resolveSupermarketImageDir(dir)
	if err != nil {
		return stats, err
	}

	entries, err := os.ReadDir(resolvedDir)
	if err != nil {
		return stats, fmt.Errorf("read image directory: %w", err)
	}

	sort.Slice(entries, func(left int, right int) bool {
		return entries[left].Name() < entries[right].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() || !isSupportedSupermarketImageFile(entry.Name()) {
			continue
		}
		stats.Scanned++
		result, err := importSupermarketImage(ctx, runtime, filepath.Join(resolvedDir, entry.Name()), userID, replace, dryRun)
		if err != nil {
			stats.Failed++
			fmt.Fprintf(os.Stderr, "skip %s: %v\n", entry.Name(), err)
			continue
		}
		switch result {
		case "uploaded":
			stats.Uploaded++
		case "replaced":
			stats.Replaced++
		default:
			stats.Skipped++
		}
	}

	return stats, nil
}

// 3. importSupermarketImage uploads and registers one local product image.
func importSupermarketImage(ctx context.Context, runtime *service.Runtime, localPath string, userID int64, replace bool, dryRun bool) (string, error) {
	code, err := supermarketImageProductCode(localPath)
	if err != nil {
		return "", err
	}

	body, err := os.ReadFile(localPath)
	if err != nil {
		return "", fmt.Errorf("read image: %w", err)
	}
	mimeType := detectSupermarketImageMimeType(body)
	if !strings.HasPrefix(mimeType, "image/") {
		return "", fmt.Errorf("unsupported image content type %s", mimeType)
	}

	objectKey := supermarketProductImageObjectKey(code)
	if dryRun {
		fmt.Printf("%s -> %s (%s)\n", filepath.Base(localPath), objectKey, mimeType)
		return "skipped", nil
	}

	var existing model.MediaAsset
	err = runtime.DB.WithContext(ctx).
		Where("bucket_name = ? AND object_key = ?", runtime.Config.StorageBucket, objectKey).
		Limit(1).
		Find(&existing).Error
	if err != nil {
		return "", fmt.Errorf("load media asset: %w", err)
	}
	if existing.ID > 0 && !replace {
		fmt.Printf("exists %s\n", objectKey)
		return "skipped", nil
	}

	if _, err := runtime.StorageProvider.PutObject(ctx, service.PutObjectInput{
		ObjectKey: objectKey,
		MimeType:  mimeType,
		Body:      body,
	}); err != nil {
		return "", fmt.Errorf("upload image: %w", err)
	}

	checksum := sha256.Sum256(body)
	createdBy := userID
	if existing.ID > 0 {
		existing.StorageProvider = runtime.Config.StorageProvider
		existing.MimeType = mimeType
		existing.FileSize = int64(len(body))
		existing.ChecksumSHA256 = hex.EncodeToString(checksum[:])
		existing.CreatedBy = &createdBy
		if err := runtime.DB.WithContext(ctx).Save(&existing).Error; err != nil {
			return "", fmt.Errorf("update media asset: %w", err)
		}
		fmt.Printf("replaced %s\n", objectKey)
		return "replaced", nil
	}

	asset := model.MediaAsset{
		PublicID:        utils.NewPublicID(),
		StorageProvider: runtime.Config.StorageProvider,
		BucketName:      runtime.Config.StorageBucket,
		ObjectKey:       objectKey,
		MimeType:        mimeType,
		FileSize:        int64(len(body)),
		ChecksumSHA256:  hex.EncodeToString(checksum[:]),
		CreatedBy:       &createdBy,
	}
	if err := runtime.DB.WithContext(ctx).Create(&asset).Error; err != nil {
		_ = runtime.StorageProvider.DeleteObject(ctx, objectKey)
		return "", fmt.Errorf("save media asset: %w", err)
	}

	fmt.Printf("uploaded %s\n", objectKey)
	return "uploaded", nil
}

// 4. supermarketImageProductCode extracts the product code from the filename.
func supermarketImageProductCode(localPath string) (string, error) {
	name := filepath.Base(localPath)
	if strings.EqualFold(name, "desktop.ini") {
		return "", fmt.Errorf("system file")
	}
	if !isSupportedSupermarketImageFile(name) {
		return "", fmt.Errorf("unsupported extension")
	}

	prefix := strings.TrimSpace(strings.SplitN(strings.TrimSuffix(name, filepath.Ext(name)), "_", 2)[0])
	if prefix == "" {
		return "", fmt.Errorf("missing product code")
	}

	return strings.ToUpper(prefix), nil
}

// 5. isSupportedSupermarketImageFile checks supported source filename extensions.
func isSupportedSupermarketImageFile(name string) bool {
	switch strings.ToLower(filepath.Ext(strings.TrimSpace(name))) {
	case ".jpg", ".jpeg", ".png", ".webp":
		return true
	default:
		return false
	}
}

// 6. detectSupermarketImageMimeType detects the real image format instead of trusting the extension.
func detectSupermarketImageMimeType(body []byte) string {
	if len(body) >= 12 && string(body[:4]) == "RIFF" && string(body[8:12]) == "WEBP" {
		return "image/webp"
	}
	return strings.ToLower(strings.TrimSpace(http.DetectContentType(body)))
}

// 7. supermarketProductImageObjectKey builds the stable OSS object key.
func supermarketProductImageObjectKey(code string) string {
	return supermarketImageObjectPrefix + strings.ToUpper(strings.TrimSpace(code)) + ".jpg"
}

// 8. resolveSupermarketImageDir finds a usable image directory path.
func resolveSupermarketImageDir(dir string) (string, error) {
	candidates := []string{}
	if strings.TrimSpace(dir) != "" {
		candidates = append(candidates, strings.TrimSpace(dir))
	}
	candidates = append(candidates, "pns_images", "../pns_images", "../../pns_images")

	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		cleaned := filepath.Clean(candidate)
		if _, ok := seen[cleaned]; ok {
			continue
		}
		seen[cleaned] = struct{}{}
		info, err := os.Stat(cleaned)
		if err == nil && info.IsDir() {
			return cleaned, nil
		}
	}

	return "", fmt.Errorf("image directory not found: %s", strings.Join(candidates, ", "))
}
