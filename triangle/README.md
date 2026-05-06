# Triangle - 图像识别自动化操作工具

基于 **gosseract**、**robotgo**、**fyne**、**onnxruntime_go** 实现的图像识别与自动化操作工具。

## 功能特性

### 1. 截屏
- 全屏截取
- 区域截取
- 保存为 PNG 文件

### 2. 模板匹配（OpenCV）
- 在屏幕中查找目标图片
- 返回匹配位置、置信度、边界框
- 支持查找后自动点击

### 3. OCR 文字识别（Tesseract）
- 支持中文、英文等多语言识别
- 可识别图片文件或屏幕截图
- 支持白名单字符过滤（提高精度）
- 支持多种页面分割模式

### 4. YOLO 目标检测（ONNX Runtime）
- 加载 YOLOv8 ONNX 模型
- 实时目标检测
- 支持自定义类别名称
- NMS 非极大值抑制

### 5. 鼠标/键盘自动化（robotgo）
- 鼠标移动、点击、双击、右键
- 拖动操作
- 滚轮滚动
- 键盘输入、按键
- 获取鼠标位置、屏幕尺寸、像素颜色

## 依赖安装

### 1. Go 依赖

```bash
go get fyne.io/fyne/v2@latest
go get github.com/go-vgo/robotgo@latest
go get github.com/otiai10/gosseract/v2@latest
go get github.com/yalue/onnxruntime_go@latest
go get gocv.io/x/gocv@latest
```

### 2. 系统依赖

#### Tesseract OCR

**Windows:**
- 下载安装：https://github.com/UB-Mannheim/tesseract/wiki
- 添加到 PATH 环境变量
- 下载语言包（如 `chi_sim.traineddata`）放到 `tessdata` 目录

**Linux:**
```bash
sudo apt install tesseract-ocr tesseract-ocr-chi-sim
```

**macOS:**
```bash
brew install tesseract tesseract-lang
```

#### ONNX Runtime

**Windows:**
- 下载 `onnxruntime.dll`：https://github.com/microsoft/onnxruntime/releases
- 放到项目目录或系统 PATH

**Linux/macOS:**
- 下载对应平台的 `libonnxruntime.so` / `libonnxruntime.dylib`

#### OpenCV（gocv）

参考 [gocv 安装文档](https://gocv.io/getting-started/)

**Windows:**
- 下载 OpenCV：https://opencv.org/releases/
- 设置环境变量 `CGO_CPPFLAGS`, `CGO_LDFLAGS`

## 使用方式

### 运行 GUI

```bash
go run ./triangle
```

### 代码示例

#### 截屏并保存

```go
img, _ := ScreenCapture()
SaveImageToFile(img, "screenshot.png")
```

#### 模板匹配

```go
result, _ := FindImageOnScreen("target.png", 0.8)
if result.Found {
    fmt.Printf("找到目标: (%d, %d)\n", result.Location.X, result.Location.Y)
}
```

#### OCR 识别

```go
client, _ := NewOCRClient([]string{"chi_sim", "eng"})
defer client.Close()

text, _ := client.RecognizeFile("image.png")
fmt.Println(text)
```

#### YOLO 检测

```go
detector, _ := NewYOLODetector(
    "yolov8n.onnx",
    "onnxruntime.dll",
    []string{"person", "car", "dog"},
    640, 640,
    0.5, 0.45,
)
defer detector.Close()

img, _ := ScreenCapture()
detections, _ := detector.Detect(img)
for _, d := range detections {
    fmt.Printf("%s: %.2f%% at %v\n", d.ClassName, d.Confidence*100, d.Rect)
}
```

#### 鼠标操作

```go
action := DefaultMouseAction()
action.MoveTo(100, 200)
action.Click(100, 200)
action.Drag(100, 200, 300, 400)
```

## 文件结构

```
triangle/
├── main.go       # GUI 主入口
├── gui.go        # Fyne 界面
├── screen.go     # 截屏 + 模板匹配
├── ocr.go        # OCR 文字识别
├── yolo.go       # YOLO 目标检测
├── action.go     # 鼠标/键盘自动化
├── util.go       # 工具函数
└── README.md
```

## 注意事项

1. **Tesseract 语言包**：首次使用 OCR 需要下载对应语言的 `.traineddata` 文件
2. **ONNX 模型**：YOLO 检测需要先导出 YOLOv8 ONNX 模型（参考 [Ultralytics 文档](https://docs.ultralytics.com/modes/export/)）
3. **权限**：某些操作（如截屏、鼠标控制）可能需要管理员权限
4. **跨平台**：robotgo 在不同平台上的行为可能略有差异

## 常见问题

### Q: OCR 识别不准确？
A: 
- 确保图片清晰、对比度高
- 使用 `SetWhitelist` 限制识别字符范围
- 调整 `SetPageSegMode` 页面分割模式

### Q: YOLO 检测失败？
A: 
- 检查模型路径和 onnxruntime.dll 路径是否正确
- 确认模型输入尺寸（通常 640x640）
- 调整置信度阈值 `confThresh`

### Q: 模板匹配找不到目标？
A: 
- 降低阈值（如 0.7）
- 确保目标图片与屏幕显示完全一致（分辨率、缩放）
- 使用更大的目标图片区域

## 参考资料

- [gosseract](https://github.com/otiai10/gosseract)
- [robotgo](https://github.com/go-vgo/robotgo)
- [fyne](https://fyne.io/)
- [onnxruntime_go](https://github.com/yalue/onnxruntime_go)
- [gocv](https://gocv.io/)
