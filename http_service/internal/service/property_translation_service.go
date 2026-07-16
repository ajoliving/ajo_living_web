/*
 * Property content translation service.
 * 1. Validate property title and description translation requests.
 * 2. Translate Traditional Chinese property content to English through DeepL.
 * 3. Keep provider credentials and upstream failures inside the backend.
 */
package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	"ajoliving_web/http_service/internal/errcode"
)

const maxTranslationResponseBytes = 1024 * 1024

// 1. PropertyContentTranslationInput defines translatable property content.
type PropertyContentTranslationInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// 2. PropertyContentTranslationResult defines translated English content.
type PropertyContentTranslationResult struct {
	TitleEn       string `json:"title_en"`
	DescriptionEn string `json:"description_en"`
}

type deepLTranslationResponse struct {
	Translations []struct {
		Text string `json:"text"`
	} `json:"translations"`
}

// 3. TranslatePropertyContent translates property content to English.
func (s *PropertyService) TranslatePropertyContent(
	ctx context.Context,
	input PropertyContentTranslationInput,
) (*PropertyContentTranslationResult, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	if input.Title == "" && input.Description == "" {
		return nil, errcode.New(errcode.CodeValidationError, "property title or description is required")
	}
	if utf8.RuneCountInString(input.Title) > 40 || utf8.RuneCountInString(input.Description) > 1000 {
		return nil, errcode.New(errcode.CodeValidationError, "property translation source is too long")
	}

	provider := strings.ToLower(strings.TrimSpace(s.runtime.Config.TranslationProvider))
	if provider != "deepl" {
		return nil, errcode.New(errcode.CodeInternalError, "translation service is unavailable")
	}

	return s.translatePropertyContentWithDeepL(ctx, input)
}

// 4. translatePropertyContentWithDeepL calls the configured DeepL endpoint.
func (s *PropertyService) translatePropertyContentWithDeepL(
	ctx context.Context,
	input PropertyContentTranslationInput,
) (*PropertyContentTranslationResult, error) {
	config := s.runtime.Config
	if config.DeepLAPIBaseURL == "" || config.DeepLAuthKey == "" {
		return nil, errcode.New(errcode.CodeInternalError, "translation service is unavailable")
	}

	form := url.Values{
		"source_lang": {"ZH"},
		"target_lang": {"EN-US"},
	}
	fields := make([]string, 0, 2)
	if input.Title != "" {
		form.Add("text", input.Title)
		fields = append(fields, "title")
	}
	if input.Description != "" {
		form.Add("text", input.Description)
		fields = append(fields, "description")
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		config.DeepLAPIBaseURL+"/v2/translate",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "translation service is unavailable")
	}
	request.Header.Set("Authorization", "DeepL-Auth-Key "+config.DeepLAuthKey)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: config.TranslationRequestTimeout}
	response, err := client.Do(request)
	if err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "translation service is unavailable")
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, errcode.New(errcode.CodeInternalError, "translation service is unavailable")
	}

	var payload deepLTranslationResponse
	if err := json.NewDecoder(io.LimitReader(response.Body, maxTranslationResponseBytes)).Decode(&payload); err != nil {
		return nil, errcode.New(errcode.CodeInternalError, "translation service returned an invalid response")
	}
	if len(payload.Translations) != len(fields) {
		return nil, errcode.New(errcode.CodeInternalError, "translation service returned an incomplete response")
	}

	result := &PropertyContentTranslationResult{}
	for index, field := range fields {
		translated := strings.TrimSpace(payload.Translations[index].Text)
		if translated == "" {
			return nil, errcode.New(errcode.CodeInternalError, "translation service returned an empty response")
		}
		if field == "title" {
			result.TitleEn = translated
		} else {
			result.DescriptionEn = translated
		}
	}
	if utf8.RuneCountInString(result.TitleEn) > 100 || utf8.RuneCountInString(result.DescriptionEn) > 2000 {
		return nil, errcode.New(errcode.CodeValidationError, "translated property content is too long")
	}

	return result, nil
}
