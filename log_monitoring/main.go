// log_monitoring - 日志监控工具（Fyne GUI）
// 功能：监控指定文件夹或指定文件，实现 tail -f 效果
// go build -ldflags "-H windowsgui" -o log_monitoring/windows/log_monitoring.exe ./log_monitoring/ 2>&1

package main

import (
	"bufio"
	"fmt"
	"image/color"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ============================================================
// 常量 & 配置
// ============================================================

const (
	maxLines     = 2000                   // 监控区最多保留行数
	pollInterval = 500 * time.Millisecond // 轮询间隔
	windowTitle  = "日志监控"
)

// ============================================================
// 自定义主题：覆盖 Success/Warning/Error 颜色用于日志着色
// ============================================================

type logTheme struct{ fyne.Theme }

func (t logTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameSuccess: // INFO → 绿
		return color.RGBA{R: 80, G: 200, B: 100, A: 255}
	case theme.ColorNameWarning: // NOTICE → 橙黄
		return color.RGBA{R: 255, G: 180, B: 0, A: 255}
	case theme.ColorNameError: // ERROR → 红
		return color.RGBA{R: 255, G: 80, B: 80, A: 255}
	}
	return t.Theme.Color(name, variant)
}

// levelColorName 根据行内容返回对应的主题颜色名
func levelColorName(line string) fyne.ThemeColorName {
	u := strings.ToUpper(line)
	switch {
	case strings.Contains(u, "[ERROR]"):
		return theme.ColorNameError
	case strings.Contains(u, "[NOTICE]"):
		return theme.ColorNameWarning
	case strings.Contains(u, "[INFO]"):
		return theme.ColorNameSuccess
	default:
		return theme.ColorNameForeground
	}
}

// ============================================================
// Windows "浮于所有窗口之上" API
// ============================================================

var (
	user32           = syscall.NewLazyDLL("user32.dll")
	procFindWindowW  = user32.NewProc("FindWindowW")
	procSetWindowPos = user32.NewProc("SetWindowPos")
)

const (
	hwndTopmost   = ^uintptr(0)
	hwndNoTopmost = ^uintptr(1)
	swpNoMove     = 0x0002
	swpNoSize     = 0x0001
	swpNoActivate = 0x0010
)

func findWindow(title string) uintptr {
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	hwnd, _, _ := procFindWindowW.Call(0, uintptr(unsafe.Pointer(titlePtr)))
	return hwnd
}

func setWindowFloat(title string, float bool) {
	hwnd := findWindow(title)
	if hwnd == 0 {
		return
	}
	insertAfter := hwndNoTopmost
	if float {
		insertAfter = hwndTopmost
	}
	procSetWindowPos.Call(hwnd, insertAfter, 0, 0, 0, 0,
		swpNoMove|swpNoSize|swpNoActivate)
}

// ============================================================
// 文件监控器
// ============================================================

type FileWatcher struct {
	path    string
	offset  int64
	mu      sync.Mutex
	stopCh  chan struct{}
	running bool
	onLine  func(path, line string)
}

func NewFileWatcher(path string, onLine func(path, line string)) *FileWatcher {
	return &FileWatcher{path: path, onLine: onLine}
}

func (fw *FileWatcher) Start() {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	if fw.running {
		return
	}
	fw.running = true
	fw.stopCh = make(chan struct{})
	if f, err := os.Open(fw.path); err == nil {
		if info, err := f.Stat(); err == nil {
			fw.offset = info.Size()
		}
		f.Close()
	}
	go fw.watch()
}

func (fw *FileWatcher) Stop() {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	if !fw.running {
		return
	}
	fw.running = false
	close(fw.stopCh)
}

func (fw *FileWatcher) IsRunning() bool {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	return fw.running
}

func (fw *FileWatcher) watch() {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-fw.stopCh:
			return
		case <-ticker.C:
			fw.readNew()
		}
	}
}

func (fw *FileWatcher) readNew() {
	f, err := os.Open(fw.path)
	if err != nil {
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return
	}
	if info.Size() < fw.offset {
		fw.offset = 0
	}
	if info.Size() == fw.offset {
		return
	}
	if _, err := f.Seek(fw.offset, io.SeekStart); err != nil {
		return
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fw.onLine(fw.path, scanner.Text())
	}
	if pos, err := f.Seek(0, io.SeekCurrent); err == nil {
		fw.offset = pos
	}
}

// ============================================================
// 文件夹监控器
// ============================================================

type DirWatcher struct {
	dir      string
	pattern  string
	watchers map[string]*FileWatcher
	mu       sync.Mutex
	stopCh   chan struct{}
	running  bool
	onLine   func(path, line string)
}

func NewDirWatcher(dir, pattern string, onLine func(path, line string)) *DirWatcher {
	return &DirWatcher{
		dir:      dir,
		pattern:  pattern,
		watchers: make(map[string]*FileWatcher),
		onLine:   onLine,
	}
}

func (dw *DirWatcher) Start() {
	dw.mu.Lock()
	defer dw.mu.Unlock()
	if dw.running {
		return
	}
	dw.running = true
	dw.stopCh = make(chan struct{})
	go dw.scan()
}

func (dw *DirWatcher) Stop() {
	dw.mu.Lock()
	defer dw.mu.Unlock()
	if !dw.running {
		return
	}
	dw.running = false
	close(dw.stopCh)
	for _, fw := range dw.watchers {
		fw.Stop()
	}
	dw.watchers = make(map[string]*FileWatcher)
}

func (dw *DirWatcher) IsRunning() bool {
	dw.mu.Lock()
	defer dw.mu.Unlock()
	return dw.running
}

func (dw *DirWatcher) scan() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	dw.discover()
	for {
		select {
		case <-dw.stopCh:
			return
		case <-ticker.C:
			dw.discover()
		}
	}
}

func (dw *DirWatcher) discover() {
	pat := dw.pattern
	if pat == "" {
		pat = "*"
	}
	matches, err := filepath.Glob(filepath.Join(dw.dir, pat))
	if err != nil {
		return
	}
	dw.mu.Lock()
	defer dw.mu.Unlock()
	for _, path := range matches {
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			continue
		}
		if _, exists := dw.watchers[path]; !exists {
			fw := NewFileWatcher(path, dw.onLine)
			fw.Start()
			dw.watchers[path] = fw
		}
	}
}

// ============================================================
// 日志缓冲区
// ============================================================

type LogLine struct {
	text      string
	colorName fyne.ThemeColorName
}

type LogBuffer struct {
	mu    sync.Mutex
	lines []LogLine
}

func (lb *LogBuffer) Append(line LogLine) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.lines = append(lb.lines, line)
	if len(lb.lines) > maxLines {
		lb.lines = lb.lines[len(lb.lines)-maxLines:]
	}
}

func (lb *LogBuffer) Clear() {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.lines = nil
}

// Segments 构建 RichText 所需的 segments，每行一个 TextSegment（Inline=true + "\n"）
func (lb *LogBuffer) Segments() []widget.RichTextSegment {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	segs := make([]widget.RichTextSegment, 0, len(lb.lines))
	for _, l := range lb.lines {
		segs = append(segs, &widget.TextSegment{
			Text: l.text + "\n",
			Style: widget.RichTextStyle{
				ColorName: l.colorName,
				Inline:    true,
				SizeName:  theme.SizeNameText,
				TextStyle: fyne.TextStyle{Monospace: true},
			},
		})
	}
	return segs
}

// ============================================================
// 主 UI
// ============================================================

func main() {
	a := app.NewWithID("com.goutils.logmonitor")
	a.Settings().SetTheme(logTheme{theme.DefaultTheme()})
	a.SetIcon(iconResource)
	w := a.NewWindow(windowTitle)
	w.SetMaster()
	w.SetIcon(iconResource)

	var (
		fileWatcher *FileWatcher
		dirWatcher  *DirWatcher
		logBuf      = &LogBuffer{}
		monMu       sync.Mutex
		alwaysOnTop = false
		isDirMode   = false
	)

	// ---- RichText 日志区（支持自动换行 + 颜色）----
	richLog := widget.NewRichText()
	richLog.Wrapping = fyne.TextWrapWord

	logScroll := container.NewVScroll(richLog)

	refreshLog := func() {
		segs := logBuf.Segments()
		richLog.Segments = segs
		richLog.Refresh()
		// 滚动到底部
		logScroll.ScrollToBottom()
	}

	appendLine := func(path, line string) {
		text := line
		if isDirMode {
			text = fmt.Sprintf("[%s] %s", filepath.Base(path), line)
		}
		logBuf.Append(LogLine{text: text, colorName: levelColorName(line)})
		fyne.Do(refreshLog)
	}

	// ---- 状态栏 ----
	statusLabel := widget.NewLabel("就绪")
	setStatus := func(msg string) {
		fyne.Do(func() { statusLabel.SetText(msg) })
	}

	// ---- 停止所有监控 ----
	stopAll := func() {
		monMu.Lock()
		defer monMu.Unlock()
		if fileWatcher != nil {
			fileWatcher.Stop()
			fileWatcher = nil
		}
		if dirWatcher != nil {
			dirWatcher.Stop()
			dirWatcher = nil
		}
	}

	isMonitoring := func() bool {
		monMu.Lock()
		defer monMu.Unlock()
		return (fileWatcher != nil && fileWatcher.IsRunning()) ||
			(dirWatcher != nil && dirWatcher.IsRunning())
	}

	// ---- 控制按钮（纯文字，避免 emoji 导致文字偏移）----
	startStopBtn := widget.NewButton("开始监控", nil)
	startStopBtn.Importance = widget.HighImportance

	clearBtn := widget.NewButton("清空内容", func() {
		logBuf.Clear()
		fyne.Do(refreshLog)
	})

	floatBtn := widget.NewButton("浮于所有窗口: 关", nil)
	floatBtn.OnTapped = func() {
		alwaysOnTop = !alwaysOnTop
		setWindowFloat(windowTitle, alwaysOnTop)
		if alwaysOnTop {
			floatBtn.SetText("浮于所有窗口: 开")
			floatBtn.Importance = widget.WarningImportance
		} else {
			floatBtn.SetText("浮于所有窗口: 关")
			floatBtn.Importance = widget.MediumImportance
		}
		floatBtn.Refresh()
	}

	// ---- 监控目标配置区 ----
	modeSelect := widget.NewSelect([]string{"监控文件", "监控文件夹"}, nil)
	modeSelect.SetSelected("监控文件")

	pathEntry := widget.NewEntry()
	pathEntry.SetPlaceHolder("输入文件或文件夹路径...")

	patternEntry := widget.NewEntry()
	patternEntry.SetPlaceHolder("文件过滤（如 *.log，留空监控所有文件）")
	patternEntry.Hide()

	modeSelect.OnChanged = func(mode string) {
		if mode == "监控文件夹" {
			patternEntry.Show()
		} else {
			patternEntry.Hide()
		}
	}

	browseBtn := widget.NewButton("浏览...", func() {
		if modeSelect.Selected == "监控文件" {
			dialog.ShowFileOpen(func(uc fyne.URIReadCloser, err error) {
				if err != nil || uc == nil {
					return
				}
				defer uc.Close()
				pathEntry.SetText(uc.URI().Path())
			}, w)
		} else {
			dialog.ShowFolderOpen(func(lu fyne.ListableURI, err error) {
				if err != nil || lu == nil {
					return
				}
				pathEntry.SetText(lu.Path())
			}, w)
		}
	})

	// ---- 开始/停止逻辑 ----
	startStopBtn.OnTapped = func() {
		if isMonitoring() {
			stopAll()
			startStopBtn.SetText("开始监控")
			startStopBtn.Importance = widget.HighImportance
			startStopBtn.Refresh()
			setStatus("监控已停止")
			return
		}

		target := strings.TrimSpace(pathEntry.Text)
		if target == "" {
			dialog.ShowInformation("提示", "请先输入或选择监控路径", w)
			return
		}

		info, err := os.Stat(target)
		if err != nil {
			dialog.ShowError(fmt.Errorf("路径无效: %w", err), w)
			return
		}

		monMu.Lock()
		if modeSelect.Selected == "监控文件夹" || info.IsDir() {
			isDirMode = true
			pattern := strings.TrimSpace(patternEntry.Text)
			dw := NewDirWatcher(target, pattern, appendLine)
			dw.Start()
			dirWatcher = dw
			patStr := pattern
			if patStr == "" {
				patStr = "*"
			}
			monMu.Unlock()
			setStatus(fmt.Sprintf("监控文件夹: %s  过滤: %s", target, patStr))
		} else {
			isDirMode = false
			fw := NewFileWatcher(target, appendLine)
			fw.Start()
			fileWatcher = fw
			monMu.Unlock()
			setStatus(fmt.Sprintf("监控文件: %s", target))
		}

		startStopBtn.SetText("停止监控")
		startStopBtn.Importance = widget.DangerImportance
		startStopBtn.Refresh()
	}

	// ---- 布局 ----
	configRow := container.NewBorder(
		nil, nil,
		container.NewHBox(modeSelect, browseBtn),
		nil,
		pathEntry,
	)

	toolbar := container.NewHBox(startStopBtn, clearBtn, floatBtn)

	topSection := container.NewVBox(
		configRow,
		patternEntry,
		toolbar,
		widget.NewSeparator(),
	)

	bottomBar := container.NewVBox(
		widget.NewSeparator(),
		statusLabel,
	)

	content := container.NewBorder(topSection, bottomBar, nil, nil, logScroll)

	w.SetContent(content)
	w.Resize(fyne.NewSize(900, 600))
	w.ShowAndRun()
}
