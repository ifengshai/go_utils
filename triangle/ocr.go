package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"github.com/otiai10/gosseract/v2"
)

// OCRClient OCR 客户端，封装 gosseract
type OCRClient struct {
	client *gosseract.Client
}

// NewOCRClient 创建 OCR 客户端
// langs: 识别语言列表，如 []string{"chi_sim", "eng"}
func NewOCRClient(langs []string) (*OCRClient, error) {
	client := gosseract.NewClient()
	if err := client.SetLanguage(langs...); err != nil {
		client.Close()
		return nil, fmt.Errorf("设置语言失败: %w", err)
	}
	return &OCRClient{client: client}, nil
}

// Close 释放资源
func (o *OCRClient) Close() {
	if o.client != nil {
		o.client.Close()
	}
}

// RecognizeFile 识别图片文件中的文字
func (o *OCRClient) RecognizeFile(imagePath string) (string, error) {
	if err := o.client.SetImage(imagePath); err != nil {
		return "", fmt.Errorf("设置图片失败: %w", err)
	}
	text, err := o.client.Text()
	if err != nil {
		return "", fmt.Errorf("识别失败: %w", err)
	}
	return text, nil
}

// RecognizeImage 识别 image.Image 中的文字（先保存为临时文件）
func (o *OCRClient) RecognizeImage(img image.Image) (string, error) {
	// 写入临时文件
	tmpPath := filepath.Join(os.TempDir(), "triangle_ocr_tmp.png")
	f, err := os.Create(tmpPath)
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %w", err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		return "", fmt.Errorf("编码图片失败: %w", err)
	}
	f.Close()
	defer os.Remove(tmpPath)

	return o.RecognizeFile(tmpPath)
}

// RecognizeBytes 识别图片字节数据中的文字
func (o *OCRClient) RecognizeBytes(data []byte) (string, error) {
	if err := o.client.SetImageFromBytes(data); err != nil {
		return "", fmt.Errorf("设置图片数据失败: %w", err)
	}
	text, err := o.client.Text()
	if err != nil {
		return "", fmt.Errorf("识别失败: %w", err)
	}
	return text, nil
}

// RecognizeRegion 截取屏幕指定区域并识别文字
func (o *OCRClient) RecognizeRegion(x, y, w, h int) (string, error) {
	img, err := ScreenCaptureRegion(x, y, w, h)
	if err != nil {
		return "", fmt.Errorf("截取区域失败: %w", err)
	}
	return o.RecognizeImage(img)
}

// RecognizeScreen 截取全屏并识别文字
func (o *OCRClient) RecognizeScreen() (string, error) {
	img, err := ScreenCapture()
	if err != nil {
		return "", fmt.Errorf("截屏失败: %w", err)
	}
	return o.RecognizeImage(img)
}

// SetWhitelist 设置只识别指定字符（提高精度）
// 例如只识别数字: "0123456789"
func (o *OCRClient) SetWhitelist(chars string) error {
	return o.client.SetWhitelist(chars)
}

// SetPageSegMode 设置页面分割模式
// gosseract.PSM_AUTO = 3 (默认), PSM_SINGLE_LINE = 7, PSM_SINGLE_WORD = 8
func (o *OCRClient) SetPageSegMode(mode gosseract.PageSegMode) error {
	return o.client.SetPageSegMode(mode)
}
