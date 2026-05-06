package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"strings"
)

// loadImageFile 从文件加载 image.Image，支持 png/jpg
func loadImageFile(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer f.Close()

	ext := strings.ToLower(path)
	switch {
	case strings.HasSuffix(ext, ".png"):
		return png.Decode(f)
	case strings.HasSuffix(ext, ".jpg"), strings.HasSuffix(ext, ".jpeg"):
		return jpeg.Decode(f)
	default:
		// 尝试自动检测
		img, _, err := image.Decode(f)
		if err != nil {
			return nil, fmt.Errorf("不支持的图片格式: %w", err)
		}
		return img, nil
	}
}
