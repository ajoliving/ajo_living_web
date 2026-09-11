/*
 * OSS media HTTP handlers.
 * 1. Bind OSS upload and media asset requests.
 * 2. Delegate media persistence and query logic to the service layer.
 */
package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/service"
)

// 1. UploadHandler handles OSS media endpoints.
type UploadHandler struct {
	uploadService *service.UploadService
}

// 2. presignRequest defines the presign payload.
type presignRequest struct {
	FileName     string `json:"file_name" binding:"required"`
	MimeType     string `json:"mime_type" binding:"required"`
	FileSize     int64  `json:"file_size" binding:"required"`
	ObjectPrefix string `json:"object_prefix"`
	Purpose      string `json:"purpose"`
}

// 3. completeUploadRequest defines the upload completion payload.
type completeUploadRequest struct {
	ObjectKey      string `json:"object_key" binding:"required"`
	UploadToken    string `json:"upload_token" binding:"required"`
	MimeType       string `json:"mime_type" binding:"required"`
	FileSize       int64  `json:"file_size" binding:"required"`
	Width          *int   `json:"width"`
	Height         *int   `json:"height"`
	ChecksumSHA256 string `json:"checksum_sha256"`
}

// 4. uploadCallbackRequest defines the OSS callback payload.
type uploadCallbackRequest struct {
	BucketName    string `json:"bucket_name" binding:"required"`
	ObjectKey     string `json:"object_key" binding:"required"`
	MimeType      string `json:"mime_type" binding:"required"`
	FileSize      int64  `json:"file_size" binding:"required"`
	ETag          string `json:"etag"`
	CallbackToken string `json:"callback_token"`
}

// 5.1 multipartInitiateRequest defines a resumable upload request.
type multipartInitiateRequest struct {
	FileName     string `json:"file_name" binding:"required"`
	MimeType     string `json:"mime_type" binding:"required"`
	FileSize     int64  `json:"file_size" binding:"required"`
	ObjectPrefix string `json:"object_prefix"`
	Purpose      string `json:"purpose"`
}

// 5.2 multipartPartRequest defines one signed part URL request.
type multipartPartRequest struct {
	ObjectKey   string `json:"object_key" binding:"required"`
	UploadID    string `json:"upload_id" binding:"required"`
	UploadToken string `json:"upload_token" binding:"required"`
	MimeType    string `json:"mime_type" binding:"required"`
	FileSize    int64  `json:"file_size" binding:"required"`
	PartNumber  int32  `json:"part_number" binding:"required"`
}

// 5.3 multipartCompleteRequest defines multipart completion metadata.
type multipartCompleteRequest struct {
	ObjectKey      string                  `json:"object_key" binding:"required"`
	UploadID       string                  `json:"upload_id" binding:"required"`
	UploadToken    string                  `json:"upload_token" binding:"required"`
	MimeType       string                  `json:"mime_type" binding:"required"`
	FileSize       int64                   `json:"file_size" binding:"required"`
	Parts          []service.MultipartPart `json:"parts" binding:"required"`
	Width          *int                    `json:"width"`
	Height         *int                    `json:"height"`
	ChecksumSHA256 string                  `json:"checksum_sha256"`
}

// 5.4 multipartAbortRequest defines multipart cancellation metadata.
type multipartAbortRequest struct {
	ObjectKey   string `json:"object_key" binding:"required"`
	UploadID    string `json:"upload_id" binding:"required"`
	UploadToken string `json:"upload_token" binding:"required"`
	MimeType    string `json:"mime_type" binding:"required"`
	FileSize    int64  `json:"file_size" binding:"required"`
}

// 5.5 multipartListRequest defines resumable parts lookup metadata.
type multipartListRequest struct {
	ObjectKey   string `json:"object_key" binding:"required"`
	UploadID    string `json:"upload_id" binding:"required"`
	UploadToken string `json:"upload_token" binding:"required"`
	MimeType    string `json:"mime_type" binding:"required"`
	FileSize    int64  `json:"file_size" binding:"required"`
}

// 5. NewUploadHandler creates an upload handler instance.
func NewUploadHandler(uploadService *service.UploadService) *UploadHandler {
	return &UploadHandler{uploadService: uploadService}
}

// 6. Presign creates an upload target.
func (h *UploadHandler) Presign(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var request presignRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}

	objectPrefix := strings.TrimSpace(request.ObjectPrefix)
	prefix, agencyPurpose := service.ResolveAgencyProfileUploadPrefix(strings.TrimSpace(request.Purpose))
	if strings.TrimSpace(request.Purpose) == "chat_attachment" {
		objectPrefix = "ajo_living/chat/"
	}
	if user.MemberStatus != "active" && (!agencyPurpose || !service.AgencyProfilePurposeMatchesAccount(request.Purpose, user.AccountType)) {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthForbidden, "only agency profile uploads are available before approval"))
		return
	}
	if agencyPurpose {
		objectPrefix = prefix
	}

	result, err := h.uploadService.Presign(c.Request.Context(), service.PresignParams{
		UserID:       user.UserID,
		FileName:     strings.TrimSpace(request.FileName),
		MimeType:     strings.TrimSpace(request.MimeType),
		FileSize:     request.FileSize,
		ObjectPrefix: objectPrefix,
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 7. InitiateMultipart creates a resumable OSS upload session.
func (h *UploadHandler) InitiateMultipart(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	var request multipartInitiateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}
	objectPrefix := strings.TrimSpace(request.ObjectPrefix)
	prefix, agencyPurpose := service.ResolveAgencyProfileUploadPrefix(strings.TrimSpace(request.Purpose))
	if strings.TrimSpace(request.Purpose) == "chat_attachment" {
		objectPrefix = "ajo_living/chat/"
	}
	if user.MemberStatus != "active" && (!agencyPurpose || !service.AgencyProfilePurposeMatchesAccount(request.Purpose, user.AccountType)) {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthForbidden, "only agency profile uploads are available before approval"))
		return
	}
	if agencyPurpose {
		objectPrefix = prefix
	}
	result, err := h.uploadService.InitiateMultipart(c.Request.Context(), service.MultipartInitiateParams{UserID: user.UserID, FileName: strings.TrimSpace(request.FileName), MimeType: strings.TrimSpace(request.MimeType), FileSize: request.FileSize, ObjectPrefix: objectPrefix})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 8. PresignMultipartPart returns a signed URL for one upload part.
func (h *UploadHandler) PresignMultipartPart(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	var request multipartPartRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}
	result, err := h.uploadService.PresignMultipartPart(c.Request.Context(), service.MultipartPartParams{UserID: user.UserID, ObjectKey: strings.TrimSpace(request.ObjectKey), UploadID: strings.TrimSpace(request.UploadID), UploadToken: strings.TrimSpace(request.UploadToken), MimeType: strings.TrimSpace(request.MimeType), FileSize: request.FileSize, PartNumber: request.PartNumber})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 9. CompleteMultipart commits uploaded parts and stores media metadata.
func (h *UploadHandler) CompleteMultipart(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	var request multipartCompleteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}
	if user.MemberStatus != "active" && !service.IsAgencyProfileObjectKey(request.ObjectKey) {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthForbidden, "only agency profile uploads are available before approval"))
		return
	}
	result, err := h.uploadService.CompleteMultipart(c.Request.Context(), service.MultipartCompleteParams{UserID: user.UserID, ObjectKey: strings.TrimSpace(request.ObjectKey), UploadID: strings.TrimSpace(request.UploadID), UploadToken: strings.TrimSpace(request.UploadToken), MimeType: strings.TrimSpace(request.MimeType), FileSize: request.FileSize, Parts: request.Parts, Width: request.Width, Height: request.Height, ChecksumSHA256: strings.TrimSpace(request.ChecksumSHA256)})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 10. ListMultipartParts returns uploaded parts for resume.
func (h *UploadHandler) ListMultipartParts(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	var request multipartListRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}
	result, err := h.uploadService.ListMultipartParts(c.Request.Context(), service.MultipartListParams{UserID: user.UserID, ObjectKey: strings.TrimSpace(request.ObjectKey), UploadID: strings.TrimSpace(request.UploadID), UploadToken: strings.TrimSpace(request.UploadToken), MimeType: strings.TrimSpace(request.MimeType), FileSize: request.FileSize})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, gin.H{"parts": result})
}

// 11. AbortMultipart cancels an in-progress multipart upload.
func (h *UploadHandler) AbortMultipart(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}
	var request multipartAbortRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}
	if err := h.uploadService.AbortMultipart(c.Request.Context(), service.MultipartAbortParams{UserID: user.UserID, ObjectKey: strings.TrimSpace(request.ObjectKey), UploadID: strings.TrimSpace(request.UploadID), UploadToken: strings.TrimSpace(request.UploadToken), MimeType: strings.TrimSpace(request.MimeType), FileSize: request.FileSize}); err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, gin.H{"aborted": true})
}

// 11. CompleteUpload persists uploaded media metadata.
func (h *UploadHandler) CompleteUpload(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var request completeUploadRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}
	if user.MemberStatus != "active" && !service.IsAgencyProfileObjectKey(request.ObjectKey) {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthForbidden, "only agency profile uploads are available before approval"))
		return
	}

	result, err := h.uploadService.CompleteUpload(c.Request.Context(), service.CompleteUploadParams{
		UserID:         user.UserID,
		ObjectKey:      strings.TrimSpace(request.ObjectKey),
		UploadToken:    strings.TrimSpace(request.UploadToken),
		MimeType:       strings.TrimSpace(request.MimeType),
		FileSize:       request.FileSize,
		Width:          request.Width,
		Height:         request.Height,
		ChecksumSHA256: strings.TrimSpace(request.ChecksumSHA256),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 8. UploadCallback accepts an OSS upload callback acknowledgement.
func (h *UploadHandler) UploadCallback(c *gin.Context) {
	var request uploadCallbackRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid callback payload"))
		return
	}

	result, err := h.uploadService.AcceptUploadCallback(c.Request.Context(), service.UploadCallbackParams{
		BucketName:    strings.TrimSpace(request.BucketName),
		ObjectKey:     strings.TrimSpace(request.ObjectKey),
		MimeType:      strings.TrimSpace(request.MimeType),
		FileSize:      request.FileSize,
		ETag:          strings.TrimSpace(request.ETag),
		CallbackToken: strings.TrimSpace(request.CallbackToken),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 9. ListAssets returns paginated media assets owned by the current user.
func (h *UploadHandler) ListAssets(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	page, pageSize := parsePagination(c)
	result, pagination, err := h.uploadService.ListMediaAssets(c.Request.Context(), user.UserID, service.MediaAssetListFilters{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"items": result, "pagination": pagination})
}

// 10. GetAsset returns a single media asset owned by the current user.
func (h *UploadHandler) GetAsset(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.uploadService.GetMediaAsset(c.Request.Context(), user.UserID, strings.TrimSpace(c.Param("mediaAssetId")))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 11. DeleteAsset removes an unused media asset owned by the current user.
func (h *UploadHandler) DeleteAsset(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.uploadService.DeleteMediaAsset(c.Request.Context(), user.UserID, strings.TrimSpace(c.Param("mediaAssetId")))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}
