package main

import (
	"fmt"
	"github.com/agnivade/levenshtein"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {

	//// 调用tesseract命令行工具进行文字识别
	//cmd := exec.Command("tesseract", "./../file/images/e224334d10f995f82dbac1f1703230c8.jpg", "stdout")
	//output, err := cmd.CombinedOutput()
	//if err != nil {
	//	fmt.Println("执行命令出错:", err)
	//}
	//fmt.Println("执行到：" + string(output))
	//return

	var images []string
	//获取文件夹下图片
	dirPath := "./../file/images" // 当前文件夹路径
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		fmt.Println("无法读取文件夹:", err)
		os.Exit(1)
	}
	for _, file := range files {
		if !file.IsDir() && isImageFile(file.Name()) {
			images = append(images, file.Name())
		}
	}

	//提取图片文字
	imageText := make(map[string]string, 0)
	for _, v := range images {
		fmt.Println("执行到：" + v)
		// 调用tesseract命令行工具进行文字识别
		cmd := exec.Command("tesseract", "./../file/images/"+v, "stdout")
		output, err := cmd.CombinedOutput()
		if err != nil {
			//log.Fatal(err)
			fmt.Println("图片翻译为文字失败：" + v)
			continue
		}
		imageText[v] = string(output)
	}
	fmt.Println("图片翻译为文字已结束。")
	fmt.Println("开始进行相似度匹配loading...")

	keysToDelete := make(map[string]bool, 0)
	for k, v := range imageText {
		for i, it := range imageText {
			if k == i {
				continue
			}
			_, ok := keysToDelete[i]
			if ok {
				continue
			}
			distance := levenshtein.ComputeDistance(v, it)
			similarity := 1 - float64(distance)/float64(len(v)+len(it))
			if similarity > 0.6 {

				fmt.Println("相似度大于0.6")
				copyFile("./../file/images/"+k, "./../file/images/similar/"+k+"/"+k)
				copyFile("./../file/images/"+i, "./../file/images/similar/"+k+"/"+i)
				keysToDelete[i] = true
			}
		}
	}
	fmt.Println("执行结束")

}

// 判断文件是否是图片文件
func isImageFile(filename string) bool {
	extensions := []string{".jpg", ".jpeg", ".png", ".gif"}
	for _, ext := range extensions {
		if strings.HasSuffix(strings.ToLower(filename), ext) {
			return true
		}
	}
	return false
}

// 复制文件
func copyFile(sourceFile, destinationFile string) error {

	dirPath := filepath.Dir(destinationFile)
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		fmt.Println("创建文件夹失败:", err)
	} else {
		fmt.Println("文件夹创建成功")
	}

	// 打开源文件
	source, err := os.Open(sourceFile)
	if err != nil {
		return err
	}
	defer source.Close()

	// 创建目标文件
	destination, err := os.Create(destinationFile)
	if err != nil {
		return err
	}
	defer destination.Close()

	// 复制文件内容
	_, err = io.Copy(destination, source)
	if err != nil {
		return err
	}

	return nil
}
