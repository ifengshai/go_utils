package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"strings"

	ort "github.com/yalue/onnxruntime_go"
)

// YOLODetection 单个检测结果
type YOLODetection struct {
	ClassID    int
	ClassName  string
	Confidence float32
	Rect       image.Rectangle
}

func main() {
	modelPath := "../model/yolo26n.onnx"
	libPath := "../onnxruntime.dll"
	imagePath := "./file/1.jpg"
	outputPath := "./file/1_result.png"

	ort.SetSharedLibraryPath(libPath)
	if err := ort.InitializeEnvironment(); err != nil {
		fmt.Printf("初始化 onnxruntime 失败: %v\n", err)
		os.Exit(1)
	}
	defer ort.DestroyEnvironment()

	inputs, outputs, err := ort.GetInputOutputInfo(modelPath)
	if err != nil {
		fmt.Printf("读取模型信息失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== 模型信息 ===")
	for _, v := range inputs {
		fmt.Printf("输入: name=%s  shape=%v  type=%v\n", v.Name, v.Dimensions, v.DataType)
	}
	for _, v := range outputs {
		fmt.Printf("输出: name=%s  shape=%v  type=%v\n", v.Name, v.Dimensions, v.DataType)
	}

	inputNames := make([]string, len(inputs))
	outputNames := make([]string, len(outputs))
	for i, v := range inputs {
		inputNames[i] = v.Name
	}
	for i, v := range outputs {
		outputNames[i] = v.Name
	}

	session, err := ort.NewDynamicAdvancedSession(modelPath, inputNames, outputNames, nil)
	if err != nil {
		fmt.Printf("创建 ONNX session 失败: %v\n", err)
		os.Exit(1)
	}
	defer session.Destroy()

	img, err := loadImageFile(imagePath)
	if err != nil {
		fmt.Printf("加载图片失败: %v\n", err)
		os.Exit(1)
	}
	origW := img.Bounds().Dx()
	origH := img.Bounds().Dy()
	fmt.Printf("\n图片尺寸: %dx%d\n", origW, origH)

	inputW, inputH := 640, 640
	inputData := preprocessImage(img, inputW, inputH)
	shape := ort.NewShape(1, 3, int64(inputH), int64(inputW))
	inputTensor, err := ort.NewTensor(shape, inputData)
	if err != nil {
		fmt.Printf("创建输入 tensor 失败: %v\n", err)
		os.Exit(1)
	}
	defer inputTensor.Destroy()

	outValues := make([]ort.Value, len(outputs))
	if err := session.Run([]ort.Value{inputTensor}, outValues); err != nil {
		fmt.Printf("推理失败: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		for _, o := range outValues {
			if o != nil {
				o.Destroy()
			}
		}
	}()

	outTensor, ok := outValues[0].(*ort.Tensor[float32])
	if !ok {
		fmt.Println("输出类型不是 float32 tensor")
		os.Exit(1)
	}
	outData := outTensor.GetData()
	outShape := outTensor.GetShape()
	fmt.Printf("输出 tensor shape: %v\n\n", outShape)

	var detections []YOLODetection

	if len(outShape) == 3 && outShape[2] == 6 {
		// NMS 后处理格式: [1, numDetections, 6] -> [x1, y1, x2, y2, conf, classID]
		fmt.Println("输出格式: NMS后处理 [batch, detections, 6]")
		numDets := int(outShape[1])

		// 调试：打印前10行原始数据
		fmt.Println("--- 原始输出前10行 [x1, y1, x2, y2, conf, classID] ---")
		for i := 0; i < 10 && i < numDets; i++ {
			base := i * 6
			fmt.Printf("  [%d] x1=%.2f y1=%.2f x2=%.2f y2=%.2f conf=%.4f cls=%.0f\n",
				i, outData[base], outData[base+1], outData[base+2], outData[base+3],
				outData[base+4], outData[base+5])
		}
		fmt.Println()

		detections = parseYOLONMSOutput(outData, numDets, origW, origH, inputW, inputH, 0.25)

	} else if len(outShape) == 3 && outShape[1] > 4 {
		// 标准 YOLOv8 格式: [1, 4+numClasses, numAnchors]
		numFeatures := int(outShape[1])
		numAnchors := int(outShape[2])
		numClasses := numFeatures - 4
		fmt.Printf("输出格式: YOLOv8标准 [1, %d, %d], 类别数: %d\n\n", numFeatures, numAnchors, numClasses)

		// 调试：打印前5个 anchor 的原始数据
		fmt.Println("--- 原始输出前5个anchor [cx, cy, w, h, maxConf, classID] ---")
		for i := 0; i < 5 && i < numAnchors; i++ {
			cx := outData[0*numAnchors+i]
			cy := outData[1*numAnchors+i]
			bw := outData[2*numAnchors+i]
			bh := outData[3*numAnchors+i]
			maxConf := float32(0)
			cls := 0
			for c := 0; c < numClasses; c++ {
				v := outData[(4+c)*numAnchors+i]
				if v > maxConf {
					maxConf = v
					cls = c
				}
			}
			fmt.Printf("  [%d] cx=%.2f cy=%.2f w=%.2f h=%.2f conf=%.4f cls=%d\n",
				i, cx, cy, bw, bh, maxConf, cls)
		}
		fmt.Println()

		classNames := getClassNames(numClasses)
		detections = parseYOLOv8Output(outData, outShape, origW, origH, inputW, inputH, 0.25, classNames)
		detections = nms(detections, 0.45)
	} else {
		fmt.Printf("未知输出格式: %v\n", outShape)
		os.Exit(1)
	}

	fmt.Printf("=== 检测结果 (置信度阈值: 0.25) ===\n")
	if len(detections) == 0 {
		fmt.Println("未检测到目标")
	} else {
		for i, d := range detections {
			fmt.Printf("[%d] 类别: %-15s  置信度: %.2f%%  位置: (%d,%d)-(%d,%d)\n",
				i+1, d.ClassName, d.Confidence*100,
				d.Rect.Min.X, d.Rect.Min.Y, d.Rect.Max.X, d.Rect.Max.Y)
		}
	}

	resultImg := drawDetections(img, detections)
	if err := saveImageFile(resultImg, outputPath); err != nil {
		fmt.Printf("保存结果图失败: %v\n", err)
	} else {
		fmt.Printf("\n结果图已保存到: %s\n", outputPath)
	}
}

// getClassNames 根据类别数返回类别名称列表
func getClassNames(numClasses int) []string {
	coco80 := []string{
		"person", "bicycle", "car", "motorcycle", "airplane", "bus", "train", "truck", "boat",
		"traffic light", "fire hydrant", "stop sign", "parking meter", "bench", "bird", "cat",
		"dog", "horse", "sheep", "cow", "elephant", "bear", "zebra", "giraffe", "backpack",
		"umbrella", "handbag", "tie", "suitcase", "frisbee", "skis", "snowboard", "sports ball",
		"kite", "baseball bat", "baseball glove", "skateboard", "surfboard", "tennis racket",
		"bottle", "wine glass", "cup", "fork", "knife", "spoon", "bowl", "banana", "apple",
		"sandwich", "orange", "broccoli", "carrot", "hot dog", "pizza", "donut", "cake",
		"chair", "couch", "potted plant", "bed", "dining table", "toilet", "tv", "laptop",
		"mouse", "remote", "keyboard", "cell phone", "microwave", "oven", "toaster", "sink",
		"refrigerator", "book", "clock", "vase", "scissors", "teddy bear", "hair drier",
		"toothbrush",
	}
	if numClasses == 80 {
		return coco80
	}
	names := make([]string, numClasses)
	for i := range names {
		if i < len(coco80) {
			names[i] = coco80[i]
		} else {
			names[i] = fmt.Sprintf("class_%d", i)
		}
	}
	return names
}

// parseYOLONMSOutput 解析 NMS 后处理格式: [1, numDets, 6] -> [x1,y1,x2,y2,conf,classID]
// 坐标是模型输入空间（640x640）下的绝对像素坐标
func parseYOLONMSOutput(data []float32, numDets, origW, origH, inputW, inputH int, confThresh float32) []YOLODetection {
	scaleX := float32(origW) / float32(inputW)
	scaleY := float32(origH) / float32(inputH)
	classNames := getClassNames(300)

	var detections []YOLODetection
	for i := 0; i < numDets; i++ {
		base := i * 6
		x1 := data[base+0]
		y1 := data[base+1]
		x2 := data[base+2]
		y2 := data[base+3]
		conf := data[base+4]
		classID := int(data[base+5])

		if conf < confThresh {
			continue
		}

		rx1 := clampInt(int(x1*scaleX), 0, origW)
		ry1 := clampInt(int(y1*scaleY), 0, origH)
		rx2 := clampInt(int(x2*scaleX), 0, origW)
		ry2 := clampInt(int(y2*scaleY), 0, origH)

		className := fmt.Sprintf("class_%d", classID)
		if classID < len(classNames) {
			className = classNames[classID]
		}
		detections = append(detections, YOLODetection{
			ClassID:    classID,
			ClassName:  className,
			Confidence: conf,
			Rect:       image.Rect(rx1, ry1, rx2, ry2),
		})
	}
	return detections
}

// parseYOLOv8Output 解析标准 YOLOv8 输出: [1, 4+numClasses, numAnchors]
func parseYOLOv8Output(data []float32, shape ort.Shape, origW, origH, inputW, inputH int, confThresh float32, classNames []string) []YOLODetection {
	if len(shape) < 3 {
		return nil
	}
	numFeatures := int(shape[1])
	numAnchors := int(shape[2])
	numClasses := numFeatures - 4

	scaleX := float32(origW) / float32(inputW)
	scaleY := float32(origH) / float32(inputH)

	var detections []YOLODetection
	for i := 0; i < numAnchors; i++ {
		cx := data[0*numAnchors+i]
		cy := data[1*numAnchors+i]
		bw := data[2*numAnchors+i]
		bh := data[3*numAnchors+i]

		maxConf := float32(0)
		classID := 0
		for c := 0; c < numClasses; c++ {
			conf := data[(4+c)*numAnchors+i]
			if conf > maxConf {
				maxConf = conf
				classID = c
			}
		}
		if maxConf < confThresh {
			continue
		}

		x1 := clampInt(int((cx-bw/2)*scaleX), 0, origW)
		y1 := clampInt(int((cy-bh/2)*scaleY), 0, origH)
		x2 := clampInt(int((cx+bw/2)*scaleX), 0, origW)
		y2 := clampInt(int((cy+bh/2)*scaleY), 0, origH)

		className := fmt.Sprintf("class_%d", classID)
		if classID < len(classNames) {
			className = classNames[classID]
		}
		detections = append(detections, YOLODetection{
			ClassID:    classID,
			ClassName:  className,
			Confidence: maxConf,
			Rect:       image.Rect(x1, y1, x2, y2),
		})
	}
	return detections
}

// preprocessImage 将图片 resize 并归一化为 float32 NCHW
func preprocessImage(img image.Image, targetW, targetH int) []float32 {
	bounds := img.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()
	data := make([]float32, 3*targetH*targetW)
	for y := 0; y < targetH; y++ {
		for x := 0; x < targetW; x++ {
			srcX := int(float64(x) * float64(srcW) / float64(targetW))
			srcY := int(float64(y) * float64(srcH) / float64(targetH))
			r, g, b, _ := img.At(bounds.Min.X+srcX, bounds.Min.Y+srcY).RGBA()
			data[0*targetH*targetW+y*targetW+x] = float32(r>>8) / 255.0
			data[1*targetH*targetW+y*targetW+x] = float32(g>>8) / 255.0
			data[2*targetH*targetW+y*targetW+x] = float32(b>>8) / 255.0
		}
	}
	return data
}

// nms 非极大值抑制
func nms(detections []YOLODetection, iouThresh float32) []YOLODetection {
	if len(detections) == 0 {
		return detections
	}
	for i := 0; i < len(detections)-1; i++ {
		for j := i + 1; j < len(detections); j++ {
			if detections[j].Confidence > detections[i].Confidence {
				detections[i], detections[j] = detections[j], detections[i]
			}
		}
	}
	kept := make([]bool, len(detections))
	for i := range kept {
		kept[i] = true
	}
	for i := 0; i < len(detections); i++ {
		if !kept[i] {
			continue
		}
		for j := i + 1; j < len(detections); j++ {
			if !kept[j] {
				continue
			}
			if iou(detections[i].Rect, detections[j].Rect) > iouThresh {
				kept[j] = false
			}
		}
	}
	var result []YOLODetection
	for i, d := range detections {
		if kept[i] {
			result = append(result, d)
		}
	}
	return result
}

func iou(a, b image.Rectangle) float32 {
	inter := a.Intersect(b)
	if inter.Empty() {
		return 0
	}
	interArea := float64(inter.Dx() * inter.Dy())
	aArea := float64(a.Dx() * a.Dy())
	bArea := float64(b.Dx() * b.Dy())
	unionArea := aArea + bArea - interArea
	if unionArea <= 0 {
		return 0
	}
	return float32(interArea / unionArea)
}

func clampInt(v, min, max int) int {
	return int(math.Min(math.Max(float64(v), float64(min)), float64(max)))
}

// loadImageFile 加载图片文件
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
		img, _, err := image.Decode(f)
		return img, err
	}
}

// drawDetections 在图片上绘制检测框
func drawDetections(src image.Image, detections []YOLODetection) image.Image {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)
	draw.Draw(dst, bounds, src, bounds.Min, draw.Src)

	boxColors := []color.RGBA{
		{255, 0, 0, 255},
		{0, 255, 0, 255},
		{0, 0, 255, 255},
		{255, 255, 0, 255},
		{255, 0, 255, 255},
		{0, 255, 255, 255},
	}
	for i, d := range detections {
		c := boxColors[i%len(boxColors)]
		drawRect(dst, d.Rect, c, 3)
	}
	return dst
}

// drawRect 绘制矩形框
func drawRect(img *image.RGBA, rect image.Rectangle, c color.RGBA, thickness int) {
	for t := 0; t < thickness; t++ {
		for x := rect.Min.X; x <= rect.Max.X; x++ {
			img.SetRGBA(x, rect.Min.Y+t, c)
			img.SetRGBA(x, rect.Max.Y-t, c)
		}
		for y := rect.Min.Y; y <= rect.Max.Y; y++ {
			img.SetRGBA(rect.Min.X+t, y, c)
			img.SetRGBA(rect.Max.X-t, y, c)
		}
	}
}

// saveImageFile 保存图片为 PNG
func saveImageFile(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
