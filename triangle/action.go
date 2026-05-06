package main

import (
	"image"
	"time"

	"github.com/go-vgo/robotgo"
)

// MouseAction 鼠标操作封装
type MouseAction struct {
	// MoveDelay 每次移动后的等待时间
	MoveDelay time.Duration
	// ClickDelay 点击前后的等待时间
	ClickDelay time.Duration
}

// DefaultMouseAction 默认鼠标操作配置
func DefaultMouseAction() *MouseAction {
	return &MouseAction{
		MoveDelay:  300 * time.Millisecond,
		ClickDelay: 200 * time.Millisecond,
	}
}

// MoveTo 平滑移动鼠标到指定位置
func (m *MouseAction) MoveTo(x, y int) {
	robotgo.MoveSmooth(x, y, 0.5, 0.5)
	time.Sleep(m.MoveDelay)
}

// MoveToPoint 移动到 image.Point
func (m *MouseAction) MoveToPoint(p image.Point) {
	m.MoveTo(p.X, p.Y)
}

// Click 左键单击
func (m *MouseAction) Click(x, y int) {
	m.MoveTo(x, y)
	time.Sleep(m.ClickDelay)
	robotgo.Click()
	time.Sleep(m.ClickDelay)
}

// ClickPoint 点击 image.Point
func (m *MouseAction) ClickPoint(p image.Point) {
	m.Click(p.X, p.Y)
}

// DoubleClick 左键双击
func (m *MouseAction) DoubleClick(x, y int) {
	m.MoveTo(x, y)
	time.Sleep(m.ClickDelay)
	robotgo.Click("left", true)
	time.Sleep(m.ClickDelay)
}

// RightClick 右键单击
func (m *MouseAction) RightClick(x, y int) {
	m.MoveTo(x, y)
	time.Sleep(m.ClickDelay)
	robotgo.Click("right")
	time.Sleep(m.ClickDelay)
}

// Drag 从 (startX, startY) 拖动到 (endX, endY)
func (m *MouseAction) Drag(startX, startY, endX, endY int) {
	robotgo.Move(startX, startY)
	time.Sleep(m.MoveDelay)
	robotgo.Toggle("left", "down")
	time.Sleep(m.ClickDelay)
	robotgo.MoveSmooth(endX, endY, 0.5, 0.5)
	time.Sleep(m.MoveDelay)
	robotgo.Toggle("left", "up")
	time.Sleep(m.ClickDelay)
}

// DragPoint 从 src 拖动到 dst
func (m *MouseAction) DragPoint(src, dst image.Point) {
	m.Drag(src.X, src.Y, dst.X, dst.Y)
}

// Scroll 滚动鼠标滚轮，x/y 为滚动量（正数向下/右，负数向上/左）
func (m *MouseAction) Scroll(x, y int) {
	robotgo.Scroll(x, y)
	time.Sleep(m.ClickDelay)
}

// GetPosition 获取当前鼠标位置
func GetPosition() image.Point {
	x, y := robotgo.Location()
	return image.Point{X: x, Y: y}
}

// GetScreenSize 获取屏幕尺寸
func GetScreenSize() (width, height int) {
	return robotgo.GetScreenSize()
}

// GetPixelColor 获取指定坐标的像素颜色（返回十六进制字符串）
func GetPixelColor(x, y int) string {
	return robotgo.GetPixelColor(x, y)
}

// KeyTap 按下并释放一个键
func KeyTap(key string, modifiers ...string) {
	robotgo.KeyTap(key, modifiers)
}

// TypeString 输入字符串
func TypeString(s string) {
	robotgo.TypeStr(s)
}

// FindAndClick 在屏幕上找到目标图片并点击
// 返回是否找到并点击成功
func FindAndClick(targetPath string, threshold float32) (bool, error) {
	result, err := FindImageOnScreen(targetPath, threshold)
	if err != nil {
		return false, err
	}
	if !result.Found {
		return false, nil
	}
	action := DefaultMouseAction()
	action.ClickPoint(result.Location)
	return true, nil
}

// FindAndDrag 在屏幕上找到目标图片并拖动到指定偏移位置
func FindAndDrag(targetPath string, threshold float32, offsetX, offsetY int) (bool, error) {
	result, err := FindImageOnScreen(targetPath, threshold)
	if err != nil {
		return false, err
	}
	if !result.Found {
		return false, nil
	}
	action := DefaultMouseAction()
	src := result.Location
	dst := image.Point{X: src.X + offsetX, Y: src.Y + offsetY}
	action.DragPoint(src, dst)
	return true, nil
}
