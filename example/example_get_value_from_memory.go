package main

import (
	"fmt"
	"strconv"
	"unsafe"
)

func main() {
	// 定义起始和结束的十六进制地址
	startAddress := uint64(0x0)
	endAddress := uint64(0xFFFFFFFFFFFFFFFF)

	// 循环遍历所有可能的地址
	for address := startAddress; address <= endAddress; address++ {
		// 将 uint64 类型的地址转换为十六进制字符串
		hexAddress := fmt.Sprintf("0x%X", address)

		// 在这里可以调用你之前编写的方法，获取内存中的值
		// 例如：value := getValueFromMemory(hexAddress)
		value := getValueFromMemory(hexAddress)
		if value == nil {
			fmt.Println("指针为空")
		} else {
			fmt.Println(value)
		}

		//// 这里打印地址，你可以在这里做一些其他操作
		//fmt.Println(hexAddress)

		// 为了演示，限制打印的数量
		if address-startAddress >= 100000 {
			break
		}
	}
}

// getAddressBinary 获取变量的内存地址的二进制表示
func getAddressBinary(ptr *int) string {
	address := uintptr(unsafe.Pointer(ptr))
	addressBinary := fmt.Sprintf("%b", address)
	return addressBinary
}

func getValueFromMemory(hexAddress string) interface{} {

	defer func() {
		recover()
	}()
	// 将十六进制地址转换为 uint64 类型
	address, err := strconv.ParseUint(hexAddress, 0, 64)
	if err != nil {
		fmt.Println("解析十六进制地址时出错:", err)
		return nil
	}

	// 将 uint64 类型的地址转换为指针类型
	ptr := unsafe.Pointer(uintptr(address))

	if ptr == nil {
		return nil
	}

	// 尝试解释为不同类型的数据
	// 这里尝试解释为 int32 类型
	var valueInt32 int32
	valueInt32Ptr := (*int32)(ptr)
	valueInt32 = *valueInt32Ptr
	if valueInt32 != 0 {
		return valueInt32
	}

	// 尝试解释为 uint32 类型
	var valueUint32 uint32
	valueUint32Ptr := (*uint32)(ptr)
	valueUint32 = *valueUint32Ptr
	if valueUint32 != 0 {
		return valueUint32
	}

	// 尝试解释为 float32 类型
	var valueFloat32 float32
	valueFloat32Ptr := (*float32)(ptr)
	valueFloat32 = *valueFloat32Ptr
	if valueFloat32 != 0.0 {
		return valueFloat32
	}

	// 尝试解释为字符串类型
	var valueString string
	valueStringPtr := (*string)(ptr)
	valueString = *valueStringPtr
	if valueString != "" {
		return valueString
	}

	// 如果无法解释为以上任何一种类型，则返回 nil
	return nil
}
