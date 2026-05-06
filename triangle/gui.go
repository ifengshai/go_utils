package main

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// TriangleApp 主应用 GUI
type TriangleApp struct {
	win    fyne.Window
	app    fyne.App
	logBox *widget.Entry // 日志输出区域
}

// NewTriangleApp 创建主界面
func NewTriangleApp(a fyne.App, w fyne.Window) *TriangleApp {
	return &TriangleApp{app: a, win: w}
}

// Build 构建并返回主界面内容
func (t *TriangleApp) Build() fyne.CanvasObject {
	// 日志区域
	t.logBox = widget.NewMultiLineEntry()
	t.logBox.Disable()
	t.logBox.SetPlaceHolder("操作日志...")
	logScroll := container.NewScroll(t.logBox)
	logScroll.SetMinSize(fyne.NewSize(600, 150))

	// 标题
	title := canvas.NewText("Triangle 图像识别自动化工具", color.NRGBA{R: 0x42, G: 0x85, B: 0xF4, A: 0xFF})
	title.TextSize = 20
	title.Alignment = fyne.TextAlignCenter

	// 各功能 Tab
	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("截屏", theme.ComputerIcon(), t.buildScreenTab()),
		container.NewTabItemWithIcon("模板匹配", theme.SearchIcon(), t.buildTemplateTab()),
		container.NewTabItemWithIcon("OCR 识别", theme.DocumentIcon(), t.buildOCRTab()),
		container.NewTabItemWithIcon("YOLO 检测", theme.InfoIcon(), t.buildYOLOTab()),
		container.NewTabItemWithIcon("鼠标操作", theme.MoveUpIcon(), t.buildActionTab()),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	return container.NewBorder(
		container.NewVBox(title, widget.NewSeparator()),
		container.NewVBox(widget.NewSeparator(), widget.NewLabel("日志"), logScroll),
		nil, nil,
		tabs,
	)
}

// log 向日志区域追加一行
func (t *TriangleApp) log(msg string) {
	current := t.logBox.Text
	if current != "" {
		current += "\n"
	}
	t.logBox.SetText(current + msg)
}

// ── 截屏 Tab ──────────────────────────────────────────────────────────────────

func (t *TriangleApp) buildScreenTab() fyne.CanvasObject {
	savePathEntry := widget.NewEntry()
	savePathEntry.SetPlaceHolder("保存路径，如 screenshot.png")
	savePathEntry.SetText("screenshot.png")

	xEntry := widget.NewEntry()
	xEntry.SetPlaceHolder("X (留空=全屏)")
	yEntry := widget.NewEntry()
	yEntry.SetPlaceHolder("Y")
	wEntry := widget.NewEntry()
	wEntry.SetPlaceHolder("宽")
	hEntry := widget.NewEntry()
	hEntry.SetPlaceHolder("高")

	captureBtn := widget.NewButton("截取全屏", func() {
		img, err := ScreenCapture()
		if err != nil {
			t.log("截屏失败: " + err.Error())
			return
		}
		path := savePathEntry.Text
		if path == "" {
			path = "screenshot.png"
		}
		if err := SaveImageToFile(img, path); err != nil {
			t.log("保存失败: " + err.Error())
			return
		}
		w, h := img.Bounds().Dx(), img.Bounds().Dy()
		t.log(fmt.Sprintf("截屏成功 %dx%d → %s", w, h, path))
	})

	captureRegionBtn := widget.NewButton("截取区域", func() {
		x, _ := strconv.Atoi(xEntry.Text)
		y, _ := strconv.Atoi(yEntry.Text)
		w, _ := strconv.Atoi(wEntry.Text)
		h, _ := strconv.Atoi(hEntry.Text)
		if w <= 0 || h <= 0 {
			t.log("请填写有效的宽高")
			return
		}
		img, err := ScreenCaptureRegion(x, y, w, h)
		if err != nil {
			t.log("截取区域失败: " + err.Error())
			return
		}
		path := savePathEntry.Text
		if path == "" {
			path = "region.png"
		}
		if err := SaveImageToFile(img, path); err != nil {
			t.log("保存失败: " + err.Error())
			return
		}
		t.log(fmt.Sprintf("截取区域成功 (%d,%d) %dx%d → %s", x, y, w, h, path))
	})

	screenSizeBtn := widget.NewButton("获取屏幕尺寸", func() {
		w, h := GetScreenSize()
		t.log(fmt.Sprintf("屏幕尺寸: %dx%d", w, h))
	})

	return container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("保存路径", savePathEntry),
		),
		widget.NewLabel("区域截取（可选）"),
		container.NewGridWithColumns(4, xEntry, yEntry, wEntry, hEntry),
		container.NewHBox(captureBtn, captureRegionBtn, screenSizeBtn),
	)
}

// ── 模板匹配 Tab ──────────────────────────────────────────────────────────────

func (t *TriangleApp) buildTemplateTab() fyne.CanvasObject {
	targetEntry := widget.NewEntry()
	targetEntry.SetPlaceHolder("目标图片路径，如 target.png")

	threshEntry := widget.NewEntry()
	threshEntry.SetText("0.8")
	threshEntry.SetPlaceHolder("匹配阈值 0~1")

	findBtn := widget.NewButton("在屏幕中查找", func() {
		path := targetEntry.Text
		if path == "" {
			t.log("请填写目标图片路径")
			return
		}
		thresh, err := strconv.ParseFloat(threshEntry.Text, 32)
		if err != nil || thresh <= 0 || thresh > 1 {
			thresh = 0.8
		}
		result, err := FindImageOnScreen(path, float32(thresh))
		if err != nil {
			t.log("查找失败: " + err.Error())
			return
		}
		if result.Found {
			t.log(fmt.Sprintf("找到目标！位置: (%d, %d)  置信度: %.2f  区域: %v",
				result.Location.X, result.Location.Y, result.Confidence, result.Rect))
		} else {
			t.log(fmt.Sprintf("未找到目标（最高置信度: %.2f）", result.Confidence))
		}
	})

	findClickBtn := widget.NewButton("查找并点击", func() {
		path := targetEntry.Text
		if path == "" {
			t.log("请填写目标图片路径")
			return
		}
		thresh, err := strconv.ParseFloat(threshEntry.Text, 32)
		if err != nil || thresh <= 0 || thresh > 1 {
			thresh = 0.8
		}
		ok, err := FindAndClick(path, float32(thresh))
		if err != nil {
			t.log("操作失败: " + err.Error())
			return
		}
		if ok {
			t.log("找到目标并已点击")
		} else {
			t.log("未找到目标，未执行点击")
		}
	})

	return container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("目标图片", targetEntry),
			widget.NewFormItem("匹配阈值", threshEntry),
		),
		container.NewHBox(findBtn, findClickBtn),
	)
}

// ── OCR Tab ───────────────────────────────────────────────────────────────────

func (t *TriangleApp) buildOCRTab() fyne.CanvasObject {
	imagePathEntry := widget.NewEntry()
	imagePathEntry.SetPlaceHolder("图片路径（留空则截取全屏）")

	langEntry := widget.NewEntry()
	langEntry.SetText("chi_sim+eng")
	langEntry.SetPlaceHolder("语言，如 chi_sim+eng")

	whitelistEntry := widget.NewEntry()
	whitelistEntry.SetPlaceHolder("白名单字符（可选），如 0123456789")

	resultEntry := widget.NewMultiLineEntry()
	resultEntry.SetPlaceHolder("识别结果...")
	resultEntry.SetMinRowsVisible(5)

	recognizeBtn := widget.NewButton("开始识别", func() {
		langs := strings.Split(langEntry.Text, "+")
		if len(langs) == 0 || langs[0] == "" {
			langs = []string{"chi_sim", "eng"}
		}

		client, err := NewOCRClient(langs)
		if err != nil {
			t.log("初始化 OCR 失败: " + err.Error())
			dialog.ShowError(err, t.win)
			return
		}
		defer client.Close()

		if whitelistEntry.Text != "" {
			if err := client.SetWhitelist(whitelistEntry.Text); err != nil {
				t.log("设置白名单失败: " + err.Error())
			}
		}

		var text string
		imgPath := imagePathEntry.Text
		if imgPath == "" {
			t.log("截取全屏进行 OCR...")
			text, err = client.RecognizeScreen()
		} else {
			t.log(fmt.Sprintf("识别图片: %s", imgPath))
			text, err = client.RecognizeFile(imgPath)
		}

		if err != nil {
			t.log("识别失败: " + err.Error())
			dialog.ShowError(err, t.win)
			return
		}

		resultEntry.SetText(text)
		t.log(fmt.Sprintf("识别完成，共 %d 个字符", len([]rune(text))))
	})

	return container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("图片路径", imagePathEntry),
			widget.NewFormItem("识别语言", langEntry),
			widget.NewFormItem("白名单字符", whitelistEntry),
		),
		recognizeBtn,
		widget.NewLabel("识别结果:"),
		container.NewScroll(resultEntry),
	)
}

// ── YOLO Tab ──────────────────────────────────────────────────────────────────

func (t *TriangleApp) buildYOLOTab() fyne.CanvasObject {
	modelPathEntry := widget.NewEntry()
	modelPathEntry.SetPlaceHolder("ONNX 模型路径，如 yolov8n.onnx")

	libPathEntry := widget.NewEntry()
	libPathEntry.SetText("onnxruntime.dll")
	libPathEntry.SetPlaceHolder("onnxruntime 动态库路径")

	classNamesEntry := widget.NewMultiLineEntry()
	classNamesEntry.SetPlaceHolder("类别名称，每行一个，如:\nperson\ncar\ndog")
	classNamesEntry.SetMinRowsVisible(4)

	imagePathEntry := widget.NewEntry()
	imagePathEntry.SetPlaceHolder("图片路径（留空则截取全屏）")

	confEntry := widget.NewEntry()
	confEntry.SetText("0.5")
	confEntry.SetPlaceHolder("置信度阈值")

	iouEntry := widget.NewEntry()
	iouEntry.SetText("0.45")
	iouEntry.SetPlaceHolder("NMS IOU 阈值")

	resultLabel := widget.NewLabel("检测结果将显示在此处")

	detectBtn := widget.NewButton("开始检测", func() {
		modelPath := modelPathEntry.Text
		if modelPath == "" {
			t.log("请填写模型路径")
			return
		}
		libPath := libPathEntry.Text
		if libPath == "" {
			libPath = "onnxruntime.dll"
		}

		// 解析类别名称
		classNames := []string{}
		for _, line := range strings.Split(classNamesEntry.Text, "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				classNames = append(classNames, line)
			}
		}

		confThresh, err := strconv.ParseFloat(confEntry.Text, 32)
		if err != nil || confThresh <= 0 {
			confThresh = 0.5
		}
		iouThresh, err := strconv.ParseFloat(iouEntry.Text, 32)
		if err != nil || iouThresh <= 0 {
			iouThresh = 0.45
		}

		t.log("初始化 YOLO 检测器...")
		detector, err := NewYOLODetector(
			modelPath, libPath, classNames,
			640, 640,
			float32(confThresh), float32(iouThresh),
		)
		if err != nil {
			t.log("初始化失败: " + err.Error())
			dialog.ShowError(err, t.win)
			return
		}
		defer detector.Close()

		// 获取图像
		var img interface{ Bounds() interface{ Dx() int } }
		imgPath := imagePathEntry.Text
		var detections []YOLODetection

		if imgPath == "" {
			t.log("截取全屏进行检测...")
			screenImg, err := ScreenCapture()
			if err != nil {
				t.log("截屏失败: " + err.Error())
				return
			}
			detections, err = detector.Detect(screenImg)
		} else {
			t.log(fmt.Sprintf("检测图片: %s", imgPath))
			// 用 gocv 读取图片再转为 image.Image
			screenImg, err := ScreenCapture() // fallback，实际应读取文件
			_ = img
			if imgPath != "" {
				// 直接用 OCR 的方式读取图片文件
				screenImg, err = loadImageFile(imgPath)
			}
			if err != nil {
				t.log("读取图片失败: " + err.Error())
				return
			}
			detections, err = detector.Detect(screenImg)
		}

		if err != nil {
			t.log("检测失败: " + err.Error())
			return
		}

		// 显示结果
		if len(detections) == 0 {
			resultLabel.SetText("未检测到目标")
			t.log("检测完成，未发现目标")
		} else {
			var sb strings.Builder
			for i, d := range detections {
				sb.WriteString(fmt.Sprintf("[%d] %s (%.1f%%)  位置: %v\n",
					i+1, d.ClassName, d.Confidence*100, d.Rect))
			}
			resultLabel.SetText(sb.String())
			t.log(fmt.Sprintf("检测完成，发现 %d 个目标", len(detections)))
		}
	})

	return container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("ONNX 模型", modelPathEntry),
			widget.NewFormItem("ORT 动态库", libPathEntry),
			widget.NewFormItem("图片路径", imagePathEntry),
			widget.NewFormItem("置信度阈值", confEntry),
			widget.NewFormItem("IOU 阈值", iouEntry),
		),
		widget.NewLabel("类别名称（每行一个）:"),
		classNamesEntry,
		detectBtn,
		widget.NewSeparator(),
		resultLabel,
	)
}

// ── 鼠标操作 Tab ──────────────────────────────────────────────────────────────

func (t *TriangleApp) buildActionTab() fyne.CanvasObject {
	xEntry := widget.NewEntry()
	xEntry.SetPlaceHolder("X")
	yEntry := widget.NewEntry()
	yEntry.SetPlaceHolder("Y")

	dxEntry := widget.NewEntry()
	dxEntry.SetPlaceHolder("目标 X（拖动）")
	dyEntry := widget.NewEntry()
	dyEntry.SetPlaceHolder("目标 Y（拖动）")

	textEntry := widget.NewEntry()
	textEntry.SetPlaceHolder("要输入的文字")

	keyEntry := widget.NewEntry()
	keyEntry.SetPlaceHolder("按键名，如 enter, space, ctrl")

	action := DefaultMouseAction()

	posBtn := widget.NewButton("获取鼠标位置", func() {
		p := GetPosition()
		t.log(fmt.Sprintf("当前鼠标位置: (%d, %d)", p.X, p.Y))
	})

	moveBtn := widget.NewButton("移动鼠标", func() {
		x, _ := strconv.Atoi(xEntry.Text)
		y, _ := strconv.Atoi(yEntry.Text)
		action.MoveTo(x, y)
		t.log(fmt.Sprintf("移动鼠标到 (%d, %d)", x, y))
	})

	clickBtn := widget.NewButton("左键点击", func() {
		x, _ := strconv.Atoi(xEntry.Text)
		y, _ := strconv.Atoi(yEntry.Text)
		action.Click(x, y)
		t.log(fmt.Sprintf("点击 (%d, %d)", x, y))
	})

	dblClickBtn := widget.NewButton("双击", func() {
		x, _ := strconv.Atoi(xEntry.Text)
		y, _ := strconv.Atoi(yEntry.Text)
		action.DoubleClick(x, y)
		t.log(fmt.Sprintf("双击 (%d, %d)", x, y))
	})

	rightClickBtn := widget.NewButton("右键点击", func() {
		x, _ := strconv.Atoi(xEntry.Text)
		y, _ := strconv.Atoi(yEntry.Text)
		action.RightClick(x, y)
		t.log(fmt.Sprintf("右键点击 (%d, %d)", x, y))
	})

	dragBtn := widget.NewButton("拖动", func() {
		x, _ := strconv.Atoi(xEntry.Text)
		y, _ := strconv.Atoi(yEntry.Text)
		dx, _ := strconv.Atoi(dxEntry.Text)
		dy, _ := strconv.Atoi(dyEntry.Text)
		action.Drag(x, y, dx, dy)
		t.log(fmt.Sprintf("拖动 (%d,%d) → (%d,%d)", x, y, dx, dy))
	})

	typeBtn := widget.NewButton("输入文字", func() {
		text := textEntry.Text
		if text == "" {
			return
		}
		TypeString(text)
		t.log(fmt.Sprintf("输入文字: %s", text))
	})

	keyBtn := widget.NewButton("按键", func() {
		key := keyEntry.Text
		if key == "" {
			return
		}
		KeyTap(key)
		t.log(fmt.Sprintf("按键: %s", key))
	})

	colorBtn := widget.NewButton("获取像素颜色", func() {
		x, _ := strconv.Atoi(xEntry.Text)
		y, _ := strconv.Atoi(yEntry.Text)
		c := GetPixelColor(x, y)
		t.log(fmt.Sprintf("(%d, %d) 颜色: #%s", x, y, c))
	})

	return container.NewVBox(
		widget.NewLabel("坐标"),
		container.NewGridWithColumns(2, xEntry, yEntry),
		widget.NewLabel("拖动目标坐标"),
		container.NewGridWithColumns(2, dxEntry, dyEntry),
		container.NewGridWithColumns(3, moveBtn, clickBtn, dblClickBtn),
		container.NewGridWithColumns(3, rightClickBtn, dragBtn, colorBtn),
		widget.NewSeparator(),
		widget.NewLabel("键盘操作"),
		container.NewGridWithColumns(2, textEntry, typeBtn),
		container.NewGridWithColumns(2, keyEntry, keyBtn),
		widget.NewSeparator(),
		container.NewHBox(posBtn, layout.NewSpacer()),
	)
}
