package main

import (
	"fmt"
	"image"
	"math"

	ort "github.com/yalue/onnxruntime_go"
)

// YOLODetection 单个检测结果
type YOLODetection struct {
	ClassID    int
	ClassName  string
	Confidence float32
	Rect       image.Rectangle // 在原图中的边界框
}

// YOLODetector YOLO 目标检测器
type YOLODetector struct {
	session    *ort.DynamicAdvancedSession
	classNames []string
	inputW     int // 模型输入宽度（通常 640）
	inputH     int // 模型输入高度（通常 640）
	confThresh float32
	iouThresh  float32
}

// NewYOLODetector 创建 YOLO 检测器
// modelPath: .onnx 模型文件路径
// libPath:   onnxruntime 动态库路径（Windows: onnxruntime.dll）
// classNames: 类别名称列表
// inputW/H:  模型输入尺寸，YOLOv8 默认 640x640
func NewYOLODetector(modelPath, libPath string, classNames []string, inputW, inputH int, confThresh, iouThresh float32) (*YOLODetector, error) {
	// 设置 onnxruntime 动态库路径
	ort.SetSharedLibraryPath(libPath)

	// 初始化 onnxruntime 环境
	if err := ort.InitializeEnvironment(); err != nil {
		return nil, fmt.Errorf("初始化 onnxruntime 失败: %w", err)
	}

	// 查询模型输入输出信息
	inputs, outputs, err := ort.GetInputOutputInfo(modelPath)
	if err != nil {
		return nil, fmt.Errorf("读取模型信息失败: %w", err)
	}
	if len(inputs) == 0 || len(outputs) == 0 {
		return nil, fmt.Errorf("模型输入/输出为空")
	}

	inputNames := make([]string, len(inputs))
	outputNames := make([]string, len(outputs))
	for i, v := range inputs {
		inputNames[i] = v.Name
	}
	for i, v := range outputs {
		outputNames[i] = v.Name
	}

	// 创建动态 session（输入输出 tensor 在 Run 时传入）
	session, err := ort.NewDynamicAdvancedSession(modelPath, inputNames, outputNames, nil)
	if err != nil {
		return nil, fmt.Errorf("创建 ONNX session 失败: %w", err)
	}

	return &YOLODetector{
		session:    session,
		classNames: classNames,
		inputW:     inputW,
		inputH:     inputH,
		confThresh: confThresh,
		iouThresh:  iouThresh,
	}, nil
}

// Close 释放资源
func (d *YOLODetector) Close() {
	if d.session != nil {
		d.session.Destroy()
	}
	ort.DestroyEnvironment()
}

// Detect 对 image.Image 进行目标检测
// 返回检测结果列表（已经过 NMS 过滤）
func (d *YOLODetector) Detect(img image.Image) ([]YOLODetection, error) {
	origW := img.Bounds().Dx()
	origH := img.Bounds().Dy()

	// 预处理：resize + normalize → float32 NCHW
	inputData := preprocessImage(img, d.inputW, d.inputH)

	// 创建输入 tensor: shape [1, 3, inputH, inputW]
	shape := ort.NewShape(1, 3, int64(d.inputH), int64(d.inputW))
	inputTensor, err := ort.NewTensor(shape, inputData)
	if err != nil {
		return nil, fmt.Errorf("创建输入 tensor 失败: %w", err)
	}
	defer inputTensor.Destroy()

	// 创建输出 tensor（动态 session 会自动分配，传 nil 即可）
	outputs := make([]ort.Value, 1)

	// 运行推理
	if err := d.session.Run([]ort.Value{inputTensor}, outputs); err != nil {
		return nil, fmt.Errorf("推理失败: %w", err)
	}
	defer func() {
		for _, o := range outputs {
			if o != nil {
				o.Destroy()
			}
		}
	}()

	if len(outputs) == 0 || outputs[0] == nil {
		return nil, fmt.Errorf("模型无输出")
	}

	// 将输出转为 float32 tensor
	outTensor, ok := outputs[0].(*ort.Tensor[float32])
	if !ok {
		return nil, fmt.Errorf("输出类型不是 float32 tensor")
	}

	outData := outTensor.GetData()
	outShape := outTensor.GetShape()

	detections := d.parseYOLOv8Output(outData, outShape, origW, origH)
	return nms(detections, d.iouThresh), nil
}

// parseYOLOv8Output 解析 YOLOv8 输出
// outShape: [1, 4+numClasses, numAnchors]
func (d *YOLODetector) parseYOLOv8Output(data []float32, shape ort.Shape, origW, origH int) []YOLODetection {
	if len(shape) < 3 {
		return nil
	}

	numFeatures := int(shape[1]) // 4 + numClasses
	numAnchors := int(shape[2])
	numClasses := numFeatures - 4

	scaleX := float32(origW) / float32(d.inputW)
	scaleY := float32(origH) / float32(d.inputH)

	var detections []YOLODetection

	for i := 0; i < numAnchors; i++ {
		// 取 cx, cy, w, h
		cx := data[0*numAnchors+i]
		cy := data[1*numAnchors+i]
		bw := data[2*numAnchors+i]
		bh := data[3*numAnchors+i]

		// 找最高置信度的类别
		maxConf := float32(0)
		classID := 0
		for c := 0; c < numClasses; c++ {
			conf := data[(4+c)*numAnchors+i]
			if conf > maxConf {
				maxConf = conf
				classID = c
			}
		}

		if maxConf < d.confThresh {
			continue
		}

		// 还原到原图坐标
		x1 := int((cx - bw/2) * scaleX)
		y1 := int((cy - bh/2) * scaleY)
		x2 := int((cx + bw/2) * scaleX)
		y2 := int((cy + bh/2) * scaleY)

		// 边界裁剪
		x1 = clampInt(x1, 0, origW)
		y1 = clampInt(y1, 0, origH)
		x2 = clampInt(x2, 0, origW)
		y2 = clampInt(y2, 0, origH)

		className := fmt.Sprintf("class_%d", classID)
		if classID < len(d.classNames) {
			className = d.classNames[classID]
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

// preprocessImage 将 image.Image resize 并归一化为 float32 NCHW 格式
func preprocessImage(img image.Image, targetW, targetH int) []float32 {
	bounds := img.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	data := make([]float32, 3*targetH*targetW)

	for y := 0; y < targetH; y++ {
		for x := 0; x < targetW; x++ {
			// 双线性插值采样坐标
			srcX := int(float64(x) * float64(srcW) / float64(targetW))
			srcY := int(float64(y) * float64(srcH) / float64(targetH))

			r, g, b, _ := img.At(bounds.Min.X+srcX, bounds.Min.Y+srcY).RGBA()

			// 归一化到 [0, 1]，NCHW 排列
			data[0*targetH*targetW+y*targetW+x] = float32(r>>8) / 255.0
			data[1*targetH*targetW+y*targetW+x] = float32(g>>8) / 255.0
			data[2*targetH*targetW+y*targetW+x] = float32(b>>8) / 255.0
		}
	}

	return data
}

// nms 非极大值抑制（NMS）
func nms(detections []YOLODetection, iouThresh float32) []YOLODetection {
	if len(detections) == 0 {
		return detections
	}

	// 按置信度降序排列
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

// iou 计算两个矩形的交并比
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
