package main

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"
)

// 素材校验配置
const (
	// 品牌名称最大长度限制
	MaxBrandNameLength = 50

	// 图片URL最大长度限制
	MaxImageURLLength = 500

	// 标题最大长度限制
	MaxHeadlineLength = 100
)

// validateMaterial 校验素材是否符合要求
func validateMaterial(ctx context.Context, item EntityItem) bool {
	// 校验品牌名称
	if !validateBrandName(item.AssetInfo.ElementInfo.BrandName) {
		resource.LoggerService.Warning(ctx, "brand name validation failed",
			logit.AutoField("brandName", item.AssetInfo.ElementInfo.BrandName),
			logit.AutoField("assetID", item.AssetInfo.AssetID))
		return false
	}

	// 校验图片URL
	if !validateImageURL(item.AssetInfo.ElementInfo.ImageURL) {
		resource.LoggerService.Warning(ctx, "image URL validation failed",
			logit.AutoField("imageURL", item.AssetInfo.ElementInfo.ImageURL),
			logit.AutoField("assetID", item.AssetInfo.AssetID))
		return false
	}

	// 校验标题
	if !validateHeadline(item.AssetInfo.ElementInfo.Headline) {
		resource.LoggerService.Warning(ctx, "headline validation failed",
			logit.AutoField("headline", item.AssetInfo.ElementInfo.Headline),
			logit.AutoField("assetID", item.AssetInfo.AssetID))
		return false
	}

	return true
}

// validateBrandName 校验品牌名称
func validateBrandName(brandName string) bool {
	if brandName == "" {
		return false
	}

	// 检查长度限制
	if utf8.RuneCountInString(brandName) > MaxBrandNameLength {
		return false
	}

	// 检查是否包含特殊字符或敏感词
	if strings.ContainsAny(brandName, "<>\"'&") {
		return false
	}

	return true
}

// validateImageURL 校验图片URL
func validateImageURL(imageURL string) bool {
	if imageURL == "" {
		return false
	}

	// 检查长度限制
	if len(imageURL) > MaxImageURLLength {
		return false
	}

	// 检查是否为有效的HTTP/HTTPS URL
	if !strings.HasPrefix(imageURL, "http://") && !strings.HasPrefix(imageURL, "https://") {
		return false
	}

	// 检查是否为支持的图片格式
	supportedFormats := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	isSupported := false
	for _, format := range supportedFormats {
		if strings.HasSuffix(strings.ToLower(imageURL), format) {
			isSupported = true
			break
		}
	}

	return isSupported
}

// validateHeadline 校验标题
func validateHeadline(headline string) bool {
	if headline == "" {
		return false
	}

	// 检查长度限制
	if utf8.RuneCountInString(headline) > MaxHeadlineLength {
		return false
	}

	return true
}

// getValidationErrorMessage 获取校验失败的错误信息
func getValidationErrorMessage(item EntityItem) string {
	var errors []string

	if !validateBrandName(item.AssetInfo.ElementInfo.BrandName) {
		errors = append(errors, fmt.Sprintf("品牌名称不符合要求(长度限制%d字符)", MaxBrandNameLength))
	}

	if !validateImageURL(item.AssetInfo.ElementInfo.ImageURL) {
		errors = append(errors, fmt.Sprintf("图片URL不符合要求(长度限制%d字符，仅支持jpg/png/gif/webp格式)", MaxImageURLLength))
	}

	if !validateHeadline(item.AssetInfo.ElementInfo.Headline) {
		errors = append(errors, fmt.Sprintf("标题不符合要求(长度限制%d字符)", MaxHeadlineLength))
	}

	return strings.Join(errors, "; ")
}
