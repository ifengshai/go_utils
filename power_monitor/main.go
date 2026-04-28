// power_monitor - Windows 电脑功率监控工具
// 获取 CPU、GPU、内存、磁盘等各组件功率，并估算电费
// 运行方式: go run ./power_monitor/
package main

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// ============================================================
// 电价配置（中国居民阶梯电价，以华北电网为参考）
// ============================================================

// TierPrice 阶梯电价
type TierPrice struct {
	Name      string  // 阶梯名称
	UnitPrice float64 // 元/度
}

// 中国居民阶梯电价（华北/华东通用参考值）
// 第一档：0.5469 元/kWh（年用电量 ≤2160 度）
// 第二档：0.5969 元/kWh（年用电量 2161~4200 度）
// 第三档：0.8469 元/kWh（年用电量 >4200 度）
var tierPrices = []TierPrice{
	{"第一档 (≤2160度/年)", 0.5469},
	{"第二档 (2161~4200度/年)", 0.5969},
	{"第三档 (>4200度/年)", 0.8469},
}

// ============================================================
// 功率数据结构
// ============================================================

// ComponentPower 单个组件功率
type ComponentPower struct {
	Name   string  // 组件名称
	PowerW float64 // 功率（瓦）
	Source string  // 数据来源说明
}

// SystemPower 整机功率汇总
type SystemPower struct {
	Components []ComponentPower
	TotalW     float64
	MeasuredAt time.Time
}

// ============================================================
// WMI 查询工具（通过 PowerShell 执行）
// ============================================================

// runPS 执行 PowerShell 命令并返回输出
func runPS(script string) (string, error) {
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// ============================================================
// CPU 功率采集
// ============================================================

// getCPUPower 通过 WMI 获取 CPU 负载和 TDP，估算实际功率
func getCPUPower() ComponentPower {
	// 获取 CPU 负载百分比
	loadStr, err := runPS(`(Get-WmiObject Win32_Processor | Measure-Object -Property LoadPercentage -Average).Average`)
	if err != nil {
		return ComponentPower{"CPU", 35.0, "默认估算(WMI失败)"}
	}
	load, err := strconv.ParseFloat(strings.TrimSpace(loadStr), 64)
	if err != nil || load < 0 {
		load = 30.0
	}

	// 获取 CPU 名称，用于估算 TDP
	nameStr, _ := runPS(`(Get-WmiObject Win32_Processor | Select-Object -First 1).Name`)
	tdp := estimateCPUTdp(nameStr)

	// 实际功率 = 空载功率 + (TDP - 空载功率) * 负载率
	// 空载约为 TDP 的 15%
	idlePower := tdp * 0.15
	actualPower := idlePower + (tdp-idlePower)*(load/100.0)

	source := fmt.Sprintf("WMI负载%.0f%% × TDP%.0fW (%s)", load, tdp, strings.TrimSpace(nameStr))
	return ComponentPower{"CPU", math.Round(actualPower*10) / 10, source}
}

// estimateCPUTdp 根据 CPU 名称估算 TDP（瓦）
func estimateCPUTdp(name string) float64 {
	name = strings.ToUpper(name)
	// Intel 桌面
	if strings.Contains(name, "I9") {
		return 125.0
	}
	if strings.Contains(name, "I7") {
		if strings.Contains(name, "K") {
			return 125.0
		}
		return 65.0
	}
	if strings.Contains(name, "I5") {
		return 65.0
	}
	if strings.Contains(name, "I3") {
		return 58.0
	}
	// Intel 移动端
	if strings.Contains(name, "INTEL") && strings.Contains(name, "U") {
		return 15.0
	}
	if strings.Contains(name, "INTEL") && strings.Contains(name, "H") {
		return 45.0
	}
	// AMD Ryzen 桌面
	if strings.Contains(name, "RYZEN 9") {
		return 105.0
	}
	if strings.Contains(name, "RYZEN 7") {
		return 65.0
	}
	if strings.Contains(name, "RYZEN 5") {
		return 65.0
	}
	if strings.Contains(name, "RYZEN 3") {
		return 65.0
	}
	// AMD 移动端
	if strings.Contains(name, "AMD") && strings.Contains(name, "U") {
		return 15.0
	}
	// 默认
	return 65.0
}

// ============================================================
// GPU 功率采集
// ============================================================

// getGPUPower 尝试通过 nvidia-smi 获取 NVIDIA GPU 功率，失败则估算
func getGPUPower() ComponentPower {
	// 尝试 nvidia-smi（NVIDIA 独显）
	out, err := runPS(`nvidia-smi --query-gpu=power.draw,name --format=csv,noheader,nounits 2>$null`)
	if err == nil && out != "" && !strings.Contains(out, "error") {
		lines := strings.Split(out, "\n")
		var totalPower float64
		var gpuNames []string
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			parts := strings.SplitN(line, ",", 2)
			if len(parts) == 2 {
				pw, e := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
				if e == nil {
					totalPower += pw
					gpuNames = append(gpuNames, strings.TrimSpace(parts[1]))
				}
			}
		}
		if totalPower > 0 {
			return ComponentPower{
				"GPU (NVIDIA)",
				math.Round(totalPower*10) / 10,
				"nvidia-smi 实测: " + strings.Join(gpuNames, ", "),
			}
		}
	}

	// 尝试获取 GPU 名称估算
	gpuName, _ := runPS(`(Get-WmiObject Win32_VideoController | Where-Object {$_.AdapterRAM -gt 0} | Select-Object -First 1).Name`)
	gpuName = strings.TrimSpace(gpuName)
	power := estimateGPUPower(gpuName)
	return ComponentPower{"GPU", power, fmt.Sprintf("估算 (%s)", gpuName)}
}

// estimateGPUPower 根据 GPU 名称估算功率（瓦）
func estimateGPUPower(name string) float64 {
	name = strings.ToUpper(name)
	// NVIDIA 独显
	if strings.Contains(name, "RTX 4090") {
		return 200.0
	}
	if strings.Contains(name, "RTX 4080") {
		return 150.0
	}
	if strings.Contains(name, "RTX 4070") {
		return 100.0
	}
	if strings.Contains(name, "RTX 4060") {
		return 70.0
	}
	if strings.Contains(name, "RTX 3090") {
		return 200.0
	}
	if strings.Contains(name, "RTX 3080") {
		return 150.0
	}
	if strings.Contains(name, "RTX 3070") {
		return 100.0
	}
	if strings.Contains(name, "RTX 3060") {
		return 70.0
	}
	if strings.Contains(name, "RTX 30") || strings.Contains(name, "RTX 20") {
		return 80.0
	}
	if strings.Contains(name, "GTX 1080") {
		return 80.0
	}
	if strings.Contains(name, "GTX 1070") {
		return 60.0
	}
	if strings.Contains(name, "GTX 1060") {
		return 40.0
	}
	if strings.Contains(name, "GTX") {
		return 50.0
	}
	// AMD 独显
	if strings.Contains(name, "RX 7900") {
		return 150.0
	}
	if strings.Contains(name, "RX 7800") || strings.Contains(name, "RX 7700") {
		return 100.0
	}
	if strings.Contains(name, "RX 6900") || strings.Contains(name, "RX 6800") {
		return 120.0
	}
	if strings.Contains(name, "RX 6700") || strings.Contains(name, "RX 6600") {
		return 80.0
	}
	if strings.Contains(name, "RADEON") {
		return 60.0
	}
	// 核显
	if strings.Contains(name, "INTEL") || strings.Contains(name, "UHD") || strings.Contains(name, "IRIS") {
		return 5.0
	}
	return 30.0
}

// ============================================================
// 内存功率
// ============================================================

// getMemoryPower 根据内存条数量和类型估算功率
func getMemoryPower() ComponentPower {
	// 获取内存条数量和类型
	out, _ := runPS(`Get-WmiObject Win32_PhysicalMemory | Select-Object Capacity,SMBIOSMemoryType | ConvertTo-Csv -NoTypeInformation`)
	lines := strings.Split(out, "\n")
	count := 0
	isDDR5 := false
	for _, line := range lines {
		if strings.Contains(line, "Capacity") {
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		count++
		// SMBIOSMemoryType 34 = DDR5, 26 = DDR4
		if strings.Contains(line, "34") {
			isDDR5 = true
		}
	}
	if count == 0 {
		count = 2
	}

	// DDR5 每条约 3W，DDR4 每条约 2.5W，DDR3 约 2W
	perStick := 2.5
	memType := "DDR4"
	if isDDR5 {
		perStick = 3.0
		memType = "DDR5"
	}
	power := float64(count) * perStick
	return ComponentPower{
		"内存 (RAM)",
		math.Round(power*10) / 10,
		fmt.Sprintf("%d条 %s，每条约%.1fW", count, memType, perStick),
	}
}

// ============================================================
// 磁盘功率
// ============================================================

// getDiskPower 估算磁盘功率
func getDiskPower() ComponentPower {
	out, _ := runPS(`Get-WmiObject Win32_DiskDrive | Select-Object MediaType,Model | ConvertTo-Csv -NoTypeInformation`)
	lines := strings.Split(out, "\n")

	ssdCount := 0
	hddCount := 0
	for _, line := range lines {
		if strings.Contains(line, "MediaType") || strings.TrimSpace(line) == "" {
			continue
		}
		lineUp := strings.ToUpper(line)
		if strings.Contains(lineUp, "SSD") || strings.Contains(lineUp, "SOLID") || strings.Contains(lineUp, "NVME") {
			ssdCount++
		} else if strings.Contains(lineUp, "HDD") || strings.Contains(lineUp, "FIXED") || strings.Contains(lineUp, "HARD") {
			hddCount++
		} else {
			ssdCount++ // 默认当 SSD
		}
	}
	if ssdCount == 0 && hddCount == 0 {
		ssdCount = 1
	}

	// SSD 约 2W，HDD 约 6W（活跃时）
	power := float64(ssdCount)*2.0 + float64(hddCount)*6.0
	desc := fmt.Sprintf("SSD×%d(2W) + HDD×%d(6W)", ssdCount, hddCount)
	return ComponentPower{"磁盘", math.Round(power*10) / 10, desc}
}

// ============================================================
// 主板功率
// ============================================================

func getMotherboardPower() ComponentPower {
	return ComponentPower{"主板", 25.0, "典型主板待机功耗估算"}
}

// ============================================================
// 音频设备功率
// ============================================================

func getAudioPower() ComponentPower {
	// 检查是否有音频输出设备
	out, _ := runPS(`Get-WmiObject Win32_SoundDevice | Where-Object {$_.Status -eq 'OK'} | Measure-Object | Select-Object -ExpandProperty Count`)
	count := 0
	if n, err := strconv.Atoi(strings.TrimSpace(out)); err == nil {
		count = n
	}
	if count == 0 {
		return ComponentPower{"音频", 0.5, "无活跃音频设备"}
	}
	// 板载声卡约 1~3W，独立声卡约 5W
	return ComponentPower{"音频", 2.0, fmt.Sprintf("%d个音频设备，板载声卡估算", count)}
}

// ============================================================
// 显示器功率（外接显示器）
// ============================================================

func getMonitorPower() ComponentPower {
	out, _ := runPS(`Get-WmiObject Win32_DesktopMonitor | Where-Object {$_.MonitorType -ne $null} | Measure-Object | Select-Object -ExpandProperty Count`)
	count := 1
	if n, err := strconv.Atoi(strings.TrimSpace(out)); err == nil && n > 0 {
		count = n
	}
	// 24寸显示器约 25W，27寸约 35W，默认按 30W 估算
	power := float64(count) * 30.0
	return ComponentPower{"显示器", power, fmt.Sprintf("%d台显示器，每台约30W估算", count)}
}

// ============================================================
// 网卡/USB 等其他设备
// ============================================================

func getOtherPower() ComponentPower {
	return ComponentPower{"网卡/USB/其他", 5.0, "网卡、USB设备等综合估算"}
}

// ============================================================
// 电源效率修正（PSU 效率损耗）
// ============================================================

// applyPSUEfficiency 考虑电源转换效率（80 PLUS Bronze 约 85%）
func applyPSUEfficiency(wallPower float64) float64 {
	const efficiency = 0.85
	return wallPower / efficiency
}

// ============================================================
// 主采集逻辑
// ============================================================

// collectPower 采集所有组件功率
func collectPower() SystemPower {
	fmt.Println("  正在采集 CPU 数据...")
	cpu := getCPUPower()

	fmt.Println("  正在采集 GPU 数据...")
	gpu := getGPUPower()

	fmt.Println("  正在采集内存数据...")
	mem := getMemoryPower()

	fmt.Println("  正在采集磁盘数据...")
	disk := getDiskPower()

	fmt.Println("  正在采集音频数据...")
	audio := getAudioPower()

	fmt.Println("  正在采集显示器数据...")
	monitor := getMonitorPower()

	mb := getMotherboardPower()
	other := getOtherPower()

	components := []ComponentPower{cpu, gpu, mem, disk, mb, audio, monitor, other}

	var total float64
	for _, c := range components {
		total += c.PowerW
	}

	// 加上电源效率损耗后的实际墙插功率
	wallTotal := applyPSUEfficiency(total)

	return SystemPower{
		Components: components,
		TotalW:     math.Round(wallTotal*10) / 10,
		MeasuredAt: time.Now(),
	}
}

// ============================================================
// 电费计算
// ============================================================

// PowerCost 功率消耗和费用
type PowerCost struct {
	Hours    float64   // 小时数
	KWh      float64   // 度数（千瓦时）
	CostTier []float64 // 各阶梯费用（元）
}

// calcCost 计算各时间梯度的用电量和费用
func calcCost(totalW float64) []PowerCost {
	hours := []float64{1, 5, 12, 24}
	var results []PowerCost

	for _, h := range hours {
		kwh := totalW / 1000.0 * h
		var costs []float64
		for _, tier := range tierPrices {
			costs = append(costs, math.Round(kwh*tier.UnitPrice*100)/100)
		}
		results = append(results, PowerCost{
			Hours:    h,
			KWh:      math.Round(kwh*1000) / 1000,
			CostTier: costs,
		})
	}
	return results
}

// ============================================================
// 输出格式化
// ============================================================

func printReport(sp SystemPower) {
	sep60 := strings.Repeat("=", 60)
	sep60d := strings.Repeat("-", 60)

	fmt.Println(sep60)
	fmt.Printf("[电脑功率监控]  采集时间: %s\n", sp.MeasuredAt.Format("2006-01-02 15:04:05"))
	fmt.Println(sep60)

	// 各组件功率
	fmt.Printf("\n%-14s  %7s  %s\n", "组件", "功率(W)", "说明")
	fmt.Println(sep60d)

	for _, c := range sp.Components {
		bar := powerBar(c.PowerW, 200)
		fmt.Printf("%-14s  %5.1f W  %s\n", c.Name, c.PowerW, bar)
		fmt.Printf("%-14s           %s\n", "", c.Source)
	}

	fmt.Println(sep60d)

	// 计算组件合计（不含效率损耗）
	var componentTotal float64
	for _, c := range sp.Components {
		componentTotal += c.PowerW
	}
	fmt.Printf("%-14s  %5.1f W  (组件合计)\n", "组件总计", componentTotal)
	fmt.Printf("%-14s  %5.1f W  (含电源转换损耗 / 85%%效率)\n", "墙插实际功率", sp.TotalW)

	// 电费估算
	fmt.Printf("\n")
	fmt.Println(sep60)
	fmt.Println("[用电量 & 电费估算]  基于墙插实际功率")
	fmt.Println(sep60)
	fmt.Printf("\n%-6s  %-10s  %-10s  %-10s  %-10s\n",
		"时长", "用电(度)", "一档(元)", "二档(元)", "三档(元)")
	fmt.Println(sep60d)

	costs := calcCost(sp.TotalW)
	for _, c := range costs {
		fmt.Printf("%-6s  %-10.4f  %-10.4f  %-10.4f  %-10.4f\n",
			fmt.Sprintf("%2.0f小时", c.Hours),
			c.KWh,
			c.CostTier[0],
			c.CostTier[1],
			c.CostTier[2],
		)
	}

	// 电价说明
	fmt.Printf("\n")
	fmt.Println(sep60d)
	fmt.Println("[中国居民阶梯电价参考 - 华北/华东电网]")
	for _, t := range tierPrices {
		fmt.Printf("  %s: %.4f 元/度\n", t.Name, t.UnitPrice)
	}
	fmt.Println()
	fmt.Println("注: 电价因地区和用电量档次不同而有所差异，以当地电网为准")
	fmt.Println("注: GPU如检测到nvidia-smi则为实测值，其余均为基于硬件规格的估算值")
	fmt.Println(sep60)
}

// powerBar 生成简单的功率条形图
func powerBar(power, maxPower float64) string {
	const barWidth = 8
	ratio := power / maxPower
	if ratio > 1 {
		ratio = 1
	}
	filled := int(ratio * barWidth)
	bar := "[" + strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled) + "]"
	return bar
}

// ============================================================
// 主函数
// ============================================================

func main() {
	fmt.Println()
	fmt.Println("  正在采集系统功率数据，请稍候...")
	fmt.Println()

	sp := collectPower()

	// 清屏（Windows）
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run()

	printReport(sp)
	fmt.Println()
}
