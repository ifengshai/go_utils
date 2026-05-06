package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"os"

	"github.com/golang/freetype"
	"golang.org/x/image/font"
)

func main() {

	//背景图片
	backgroundImageFile, err := os.Open("images/a.jpg")
	if err != nil {
		fmt.Println("打开图片出错")
		fmt.Println(err)
		os.Exit(-1)
	}
	defer backgroundImageFile.Close()
	backgroundImage, err := jpeg.Decode(backgroundImageFile)
	if err != nil {
		fmt.Println("把图片解码为结构体时出错")
		fmt.Println(backgroundImage)
		os.Exit(-1)
	}
	backgroundRectangle := backgroundImage.Bounds()
	// 1.创建画布。根据背景图片的大小新建一个画布
	newCanvas := image.NewRGBA(backgroundRectangle)

	// 2.创建第一个图层。画上背景图片
	//draw.Src源图像透过遮罩后，替换掉目标图像
	//draw.Over源图像透过遮罩后，覆盖在目标图像上
	draw.Draw(newCanvas, backgroundRectangle, backgroundImage, image.ZP, draw.Over)

	// 文字水印
	fontbyte, err := ioutil.ReadFile("./images/msyh.ttf")
	if err != nil {
		fmt.Println("ioutil.ReadFile error : ", err)
		return
	}
	fonttype, err := freetype.ParseFont(fontbyte)
	if err != nil {
		fmt.Println("freetype.ParseFont error : ", err)
		return
	}
	// 创建一个新的上下文
	context := freetype.NewContext()
	context.SetDPI(70)                                             // 设置屏幕分辨率，单位为每英寸点数。
	context.SetClip(newCanvas.Bounds())                                 //设置新画布的大小。
	context.SetDst(newCanvas)                                           //设置绘制操作的目标图像。
	context.SetFont(fonttype)                                          //设置文字的字体
	context.SetFontSize(50)                                        //设置字体大小
	context.SetSrc(image.NewUniform(color.RGBA{255, 0, 0, 1})) //设置字体颜色
	context.SetHinting(font.HintingFull)
	pt := freetype.Pt(10, 90)                                     //文字坐标

	// 3.创建第三个图层。画上文字水印
	context.DrawString("文字水印", pt)

	//水印图片
	waterFile, err := os.Open("images/b.png")
	if err != nil {
		fmt.Println("打开水印图片出错")
		fmt.Println(err)
		os.Exit(-1)
	}
	defer waterFile.Close()
	waterImage, err := png.Decode(waterFile)
	if err != nil {
		fmt.Println("把水印图片解码为结构体时出错")
		fmt.Println(err)
		os.Exit(-1)
	}
	//把水印写在右下角，并向0坐标偏移10个像素
	offset := image.Pt(backgroundImage.Bounds().Dx()-waterImage.Bounds().Dx()-100, backgroundImage.Bounds().Dy()-waterImage.Bounds().Dy()-10)
	// 4.创建第四个图层。画上图片水印
	//image.ZP代表Point结构体，目标的源点，即(0,0)
	draw.Draw(newCanvas, waterImage.Bounds().Add(offset), waterImage, image.ZP, draw.Over)

	//创建文件
	file, err := os.Create("./images/new.jpg")
	if err != nil {
		fmt.Println("os.Open error : ", err)
		return
	}

	// 5.输出图片。将图像写入file
	//&jpeg.Options{100} 取值范围[1,100]，越大图像编码质量越高
	jpeg.Encode(file, newCanvas, &jpeg.Options{100})
	defer file.Close()

	fmt.Println("添加文字，图片两种水印结束请查看")
}