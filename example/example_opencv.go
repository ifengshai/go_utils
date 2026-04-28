//使用gocv，参考文档https://gocv.io/getting-started/windows/
// 
// 第一步：安装MinGW 编译器
// 安装MSYS2：https://www.msys2.org/  打开 「MSYS2 MINGW64」
// 执行命令：pacman -Syu
// 执行命令：pacman -S --needed mingw-w64-x86_64-gcc mingw-w64-x86_64-make
// 把编译好的exe加入到path中 C:\msys64\mingw64\bin
//
// 第二步：安装Cmake https://cmake.org/download/#previous
//
// 第三步：下载gocv git clone https://github.com/hybridgroup/gocv.git
//
// 第四步：编译gocv
// 执行命令：cd gocv
// 执行命令：chdir gocv
//   .\win_download_opencv.cmd
//   .\win_build_opencv.cmd
//  
//
//	go run -tags cgo example/example_opencv.go template.png
//	go run -tags cgo example/example_opencv.go template.png -min 0.75 -max 1.35 -steps 25 -thresh 0.72 -minscore 0.8
//	成功时在 stdout 输出缩进 JSON 数组（默认仅含 score>=0.8）；错误信息在 stderr。

//打包命令：.\scripts\package_go_opencv.ps1 -OpenCVBin "C:\opencv\build\install\x64\mingw\bin" -MingwBin "C:\msys64\mingw64\bin"

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"os"
	"sort"

	"github.com/kbinani/screenshot"
	"gocv.io/x/gocv"
)

type matchResult struct {
	DisplayIndex int     `json:"displayIndex"`
	Scale        float64 `json:"scale"`
	Score        float64 `json:"score"`
	// 虚拟桌面坐标系下的轴对齐矩形：左、上、右、下（像素，含边界）
	Left   int `json:"left"`
	Top    int `json:"top"`
	Right  int `json:"right"`
	Bottom int `json:"bottom"`
}

func main() {
	minScale := flag.Float64("min", 0.65, "模板相对屏幕截图的最小缩放比例（应对 DPI/分辨率差异）")
	maxScale := flag.Float64("max", 1.45, "模板相对屏幕截图的最大缩放比例")
	steps := flag.Int("steps", 33, "在 [min,max] 之间线性采样的缩放档位数")
	threshold := flag.Float64("thresh", 0.72, "TM_CCOEFF_NORMED 峰值检测最低分（用于找候选）")
	minScore := flag.Float64("minscore", 0.8, "只输出分数>=该值的命中；默认 0.8")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "用法: example_opencv <模板图片路径> [flags]")
		os.Exit(2)
	}
	tplPath := args[0]

	tpl, err := imReadFile(tplPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "无法读取模板: %v\n", err)
		os.Exit(1)
	}
	defer tpl.Close()

	tplGray := gocv.NewMat()
	gocv.CvtColor(tpl, &tplGray, gocv.ColorBGRToGray)
	defer tplGray.Close()

	tw := tplGray.Cols()
	th := tplGray.Rows()
	if tw < 4 || th < 4 {
		fmt.Fprintln(os.Stderr, "模板过小")
		os.Exit(1)
	}

	var all []matchResult

	n := screenshot.NumActiveDisplays()
	for i := 0; i < n; i++ {
		b := screenshot.GetDisplayBounds(i)
		img, err := screenshot.CaptureRect(b)
		if err != nil {
			fmt.Fprintf(os.Stderr, "截取显示器 %d 失败: %v\n", i, err)
			continue
		}

		screenGray, err := rgbaToGrayMat(img)
		if err != nil {
			fmt.Fprintf(os.Stderr, "转换屏幕图像失败: %v\n", err)
			continue
		}

		sw := screenGray.Cols()
		sh := screenGray.Rows()
		if sw < tw || sh < th {
			screenGray.Close()
			continue
		}

		peakTh := *threshold
		if *minScore > peakTh {
			peakTh = *minScore
		}
		found := matchMultiScaleAll(screenGray, tplGray, b.Min.X, b.Min.Y, i, *minScale, *maxScale, *steps, peakTh)
		all = append(all, found...)
		screenGray.Close()
	}

	if len(all) == 0 {
		fmt.Fprintln(os.Stderr, "未在任一显示器上找到足够相似的匹配（可调低 -thresh/-minscore 或扩大 -min/-max）")
		os.Exit(1)
	}

	sort.Slice(all, func(i, j int) bool { return all[i].Score > all[j].Score })

	out := all[:0]
	for _, m := range all {
		if m.Score >= *minScore {
			out = append(out, m)
		}
	}
	if len(out) == 0 {
		fmt.Fprintf(os.Stderr, "无分数>=%.2f 的命中（当前最高 %.4f，可调低 -minscore 或 -thresh）\n", *minScore, all[0].Score)
		os.Exit(1)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(out)
}

// imReadFile 用 Go 打开文件再 IMDecode，避免 Windows 上 cv::imread 对 UTF-8 路径乱码导致中文文件名失败。
func imReadFile(path string) (gocv.Mat, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return gocv.Mat{}, err
	}
	mat, err := gocv.IMDecode(data, gocv.IMReadColor)
	if err != nil {
		return gocv.Mat{}, err
	}
	if mat.Empty() {
		mat.Close()
		return gocv.Mat{}, fmt.Errorf("无法解码图片（格式不支持或内容损坏）: %s", path)
	}
	return mat, nil
}

// rgbaToGrayMat 从截图 RGBA 直接建 OpenCV 灰度图，避免 PNG 编解码与 BGR 中间缓冲（全屏时收益很大）。
func rgbaToGrayMat(img *image.RGBA) (gocv.Mat, error) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return gocv.Mat{}, fmt.Errorf("空图像")
	}
	rowBytes := w * 4

	var rgba gocv.Mat
	var err error
	if img.Stride == rowBytes {
		if len(img.Pix) < h*rowBytes {
			return gocv.Mat{}, fmt.Errorf("RGBA 数据长度不足")
		}
		rgba, err = gocv.NewMatFromBytes(h, w, gocv.MatTypeCV8UC4, img.Pix[:h*rowBytes])
	} else {
		pix := make([]byte, rowBytes*h)
		for y := 0; y < h; y++ {
			rowOff := y * img.Stride
			copy(pix[y*rowBytes:], img.Pix[rowOff:rowOff+rowBytes])
		}
		rgba, err = gocv.NewMatFromBytes(h, w, gocv.MatTypeCV8UC4, pix)
	}
	if err != nil {
		return gocv.Mat{}, err
	}

	gray := gocv.NewMat()
	gocv.CvtColor(rgba, &gray, gocv.ColorRGBAToGray)
	rgba.Close()
	if gray.Empty() {
		gray.Close()
		return gocv.Mat{}, fmt.Errorf("灰度转换失败")
	}
	return gray, nil
}

const (
	maxPeaksPerScale = 48
	nmsIoU           = 0.45
)

// matchMultiScaleAll 在同一显示器上多尺度匹配，每尺度取多个峰值并做 NMS，避免相邻尺度重复框。
func matchMultiScaleAll(
	screenGray gocv.Mat,
	tplGray gocv.Mat,
	offsetX, offsetY int,
	displayIndex int,
	minScale, maxScale float64,
	steps int,
	threshold float64,
) []matchResult {
	sw := screenGray.Cols()
	sh := screenGray.Rows()
	tw0 := tplGray.Cols()
	th0 := tplGray.Rows()

	if steps < 2 {
		steps = 2
	}

	result := gocv.NewMat()
	defer result.Close()
	resized := gocv.NewMat()
	defer resized.Close()
	mask := gocv.NewMat()
	defer mask.Close()

	var cand []matchResult

	for s := 0; s < steps; s++ {
		t := float64(s) / float64(steps-1)
		scale := minScale + t*(maxScale-minScale)
		if scale <= 0 {
			continue
		}
		nw := int(float64(tw0)*scale + 0.5)
		nh := int(float64(th0)*scale + 0.5)
		if nw < 4 || nh < 4 || nw >= sw || nh >= sh {
			continue
		}
		interp := gocv.InterpolationLinear
		if nw < tw0 || nh < th0 {
			interp = gocv.InterpolationArea
		}
		gocv.Resize(tplGray, &resized, image.Pt(nw, nh), 0, 0, interp)

		gocv.MatchTemplate(screenGray, resized, &result, gocv.TmCcoeffNormed, mask)

		for p := 0; p < maxPeaksPerScale; p++ {
			_, maxVal, _, maxLoc := gocv.MinMaxLoc(result)
			mv := float64(maxVal)
			if mv < threshold {
				break
			}
			left := offsetX + maxLoc.X
			top := offsetY + maxLoc.Y
			cand = append(cand, matchResult{
				DisplayIndex: displayIndex,
				Scale:        scale,
				Score:        mv,
				Left:         left,
				Top:          top,
				Right:        left + nw - 1,
				Bottom:       top + nh - 1,
			})
			suppressMatchRegion(&result, maxLoc, nw, nh)
		}
	}

	return nmsByIoU(cand, nmsIoU)
}

func suppressMatchRegion(m *gocv.Mat, maxLoc image.Point, nw, nh int) {
	r := image.Rect(maxLoc.X, maxLoc.Y, maxLoc.X+nw, maxLoc.Y+nh).Intersect(
		image.Rect(0, 0, m.Cols(), m.Rows()))
	if r.Empty() {
		return
	}
	roi := m.Region(r)
	roi.SetTo(gocv.NewScalar(-1, 0, 0, 0))
	_ = roi.Close()
}

func rectArea(left, top, right, bottom int) float64 {
	w := float64(right - left + 1)
	h := float64(bottom - top + 1)
	if w <= 0 || h <= 0 {
		return 0
	}
	return w * h
}

func rectIoU(a, b matchResult) float64 {
	x0 := maxInt(a.Left, b.Left)
	y0 := maxInt(a.Top, b.Top)
	x1 := minInt(a.Right, b.Right)
	y1 := minInt(a.Bottom, b.Bottom)
	if x1 < x0 || y1 < y0 {
		return 0
	}
	inter := float64(x1-x0+1) * float64(y1-y0+1)
	union := rectArea(a.Left, a.Top, a.Right, a.Bottom) + rectArea(b.Left, b.Top, b.Right, b.Bottom) - inter
	if union <= 0 {
		return 0
	}
	return inter / union
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// nmsByIoU 按 score 从高保留；与已保留框 IoU >= th 的丢弃（用于跨尺度去重）。
func nmsByIoU(in []matchResult, iouTh float64) []matchResult {
	if len(in) == 0 {
		return nil
	}
	sort.Slice(in, func(i, j int) bool { return in[i].Score > in[j].Score })
	var keep []matchResult
	for _, m := range in {
		ok := true
		for _, k := range keep {
			if rectIoU(m, k) >= iouTh {
				ok = false
				break
			}
		}
		if ok {
			keep = append(keep, m)
		}
	}
	return keep
}
