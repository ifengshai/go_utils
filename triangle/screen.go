package main

import (
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/go-vgo/robotgo"
	"gocv.io/x/gocv"
)

// ScreenCapture 截取全屏，返回 image.Image
func ScreenCapture() (image.Image, error) {
	w, h := robotgo.GetScreenSize()
	bitmap := robotgo.CaptureScreen(0, 0, w, h)
	if bitmap == nil {
		return nil, fmt.Errorf("截屏失败")
	}
	defer robotgo.FreeBitmap(bitmap)
	img := robotgo.ToImage(bitmap)
	return img, nil
}

// ScreenCaptureRegion 截取指定区域
func ScreenCaptureRegion(x, y, w, h int) (image.Image, error) {
	bitmap := robotgo.CaptureScreen(x, y, w, h)
	if bitmap == nil {
		return nil, fmt.Errorf("截取区域失败")
	}
	defer robotgo.FreeBitmap(bitmap)
	img := robotgo.ToImage(bitmap)
	return img, nil
}

// SaveImageToFile 将 image.Image 保存为 PNG 文件
func SaveImageToFile(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer f.Close()
	return png.Encode(f, img)
}

// TemplateMatchResult 模板匹配结果
type TemplateMatchResult struct {
	Found      bool
	Confidence float32     // 匹配置信度 0~1
	Location   image.Point // 目标中心点（屏幕坐标）
	Rect       image.Rectangle
}

// FindImageOnScreen 在屏幕上查找目标图片（模板匹配）
// targetPath: 要查找的目标图片路径
// threshold: 匹配阈值，建议 0.8
func FindImageOnScreen(targetPath string, threshold float32) (*TemplateMatchResult, error) {
	// 截屏
	screenImg, err := ScreenCapture()
	if err != nil {
		return nil, err
	}
	return FindImageInImage(screenImg, targetPath, threshold)
}

// FindImageInImage 在给定图像中查找目标图片
func FindImageInImage(src image.Image, targetPath string, threshold float32) (*TemplateMatchResult, error) {
	// 转换 source 为 Mat
	srcMat, err := gocv.ImageToMatRGB(src)
	if err != nil {
		return nil, fmt.Errorf("源图像转换失败: %w", err)
	}
	defer srcMat.Close()

	// 读取目标模板
	tplMat := gocv.IMRead(targetPath, gocv.IMReadColor)
	if tplMat.Empty() {
		return nil, fmt.Errorf("无法读取目标图片: %s", targetPath)
	}
	defer tplMat.Close()

	return matchTemplate(&srcMat, &tplMat, threshold)
}

// FindImageInMat 在 Mat 中查找目标 Mat（内部复用）
func FindImageInMat(srcMat, tplMat *gocv.Mat, threshold float32) (*TemplateMatchResult, error) {
	return matchTemplate(srcMat, tplMat, threshold)
}

func matchTemplate(srcMat, tplMat *gocv.Mat, threshold float32) (*TemplateMatchResult, error) {
	result := gocv.NewMat()
	defer result.Close()

	mask := gocv.NewMat()
	defer mask.Close()

	gocv.MatchTemplate(*srcMat, *tplMat, &result, gocv.TmCcoeffNormed, mask)

	_, maxVal, _, maxLoc := gocv.MinMaxLoc(result)

	res := &TemplateMatchResult{
		Confidence: float32(maxVal),
	}

	if float32(maxVal) >= threshold {
		w, h := tplMat.Cols(), tplMat.Rows()
		res.Found = true
		res.Rect = image.Rect(maxLoc.X, maxLoc.Y, maxLoc.X+w, maxLoc.Y+h)
		res.Location = image.Point{
			X: maxLoc.X + w/2,
			Y: maxLoc.Y + h/2,
		}
	}

	return res, nil
}
