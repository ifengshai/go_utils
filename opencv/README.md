# go_opencv 工具包说明

本目录为 **已编译的屏幕模板匹配工具** 及其 **运行时依赖（`lib\` 下 DLL）**。可在任意 **64 位 Windows** 机器上拷贝整夹使用（无需安装 Go/OpenCV 开发环境）。

---

## 一、如何调用（命令行）

依赖 DLL 放在 **`lib\`** 子目录中。Windows 默认不会自动从 `lib` 加载 DLL，需先将其加入 `PATH`：

```bat
cd /d D:\路径\windows
set PATH=%CD%\lib;%PATH%
go_opencv.exe <模板图片路径> [可选参数...]
```

PowerShell 写法：

```powershell
Set-Location D:\路径\windows
$env:PATH = "$PWD\lib;$env:PATH"
.\go_opencv.exe <模板图片路径> [可选参数...]
```

### PowerShell 与中文路径

为含空格或中文的路径 **加引号**。若仍异常，可让 CMD 代为解析：

```powershell
cmd /c 'go_opencv.exe "D:\game\...\加号.png"'
```

---

## 二、程序做什么（简要）

1. 读取你提供的 **模板小图**（PNG/JPEG 等常见格式，路径支持中文）。
2. 截取当前 **所有显示器** 的画面。
3. 在灰度图上用 OpenCV **多尺度模板匹配**（`TM_CCOEFF_NORMED`），找出与模板相似的区域；多块屏、同屏多实例可有多条结果。
4. 在 **标准输出（stdout）** 打印 **JSON 数组**；错误与提示在 **标准错误（stderr）**。

---

## 三、输入说明

### 3.1 位置参数（必填）

| 顺序 | 含义 |
|------|------|
| 第 1 个 | 模板图片路径（相对路径或绝对路径均可） |

### 3.2 可选参数（flag）

| 参数 | 默认值 | 含义 |
|------|--------|------|
| `-min` | 0.65 | 模板相对屏幕的 **最小** 缩放比例（应对 DPI/分辨率差异） |
| `-max` | 1.45 | 模板相对屏幕的 **最大** 缩放比例 |
| `-steps` | 33 | 在 `[min,max]` 之间线性采样的缩放档位数，越大越慢、越细 |
| `-thresh` | 0.72 | 匹配得分 **峰值检测** 下限（用于在得分图上找候选） |
| `-minscore` | 0.80 | **仅输出** 得分 ≥ 该值的命中；若高于 `-thresh`，内部检测也会用二者较大者 |

说明：`-thresh` 与 `-minscore` 同时存在时，内部会用 **`max(-thresh, -minscore)`** 作为峰值门槛，减少算了又丢弃的低分候选。

---

## 四、输出说明

### 4.1 标准输出（stdout）

成功时输出 **一个 JSON 数组**（可能只有 1 个元素），元素字段如下：

| 字段 | 类型 | 含义 |
|------|------|------|
| `displayIndex` | 整数 | 命中所在显示器序号（从 0 开始） |
| `scale` | 浮点 | 本次命中使用的模板相对原图的缩放比例 |
| `score` | 浮点 | 相似度（归一化相关系数，约 -1～1，**越大越像**） |
| `left` / `top` / `right` / `bottom` | 整数 | 命中矩形在 **虚拟桌面坐标** 下的像素边界（含边界） |

数组按 **`score` 从高到低** 排序。

### 4.2 标准错误（stderr）

模板无法读取、未找到匹配、参数错误等信息输出到 stderr，便于脚本把「人读日志」与「机器读 JSON」分开。

### 4.3 退出码（简要）

| 退出码 | 常见情况 |
|--------|----------|
| 0 | 成功，stdout 有合法 JSON |
| 1 | 运行失败或未找到满足条件的匹配 |
| 2 | 用法错误（例如未提供模板路径） |

---

## 五、使用示例

### 5.1 基本用法（默认只输出 score ≥ 0.8）

```bat
cd /d D:\你的路径\windows
set PATH=%CD%\lib;%PATH%
go_opencv.exe "D:\截图\加号.png"
```

### 5.2 调整灵敏度与缩放范围

```bat
go_opencv.exe "D:\截图\按钮.png" -min 0.75 -max 1.35 -steps 40 -thresh 0.70 -minscore 0.85
```

### 5.3 在 PowerShell 里把 JSON 存成文件

```powershell
Set-Location D:\你的路径\windows
$env:PATH = "$PWD\lib;$env:PATH"
.\go_opencv.exe "D:\截图\模板.png" | Out-File -Encoding utf8 result.json
```

注意：若程序在 stderr 打印了警告，**不要**把 stderr 重定向进同一个 JSON 文件，否则会破坏格式。仅重定向 stdout 即可。

### 5.4 成功时 stdout 示例

```json
[
  {
    "displayIndex": 0,
    "scale": 1.1,
    "score": 0.9907309412956238,
    "left": 510,
    "top": 1107,
    "right": 539,
    "bottom": 1142
  }
]
```

---

## 六、重新打包本目录

若你在开发机更新了源码或依赖，可在仓库根目录执行（路径按本机修改）：

```powershell
.\opencv\package_go_opencv.ps1 -OpenCVBin "C:\opencv\build\install\x64\mingw\bin" -MingwBin "C:\msys64\mingw64\bin"
```

脚本会重新生成 `opencv\windows\go_opencv.exe` 及 `lib\` 下 DLL。

---

## 七、注意事项

1. **本包为 64 位**：请在 **64 位 Windows** 上使用；与编译所用 MinGW/OpenCV 位数一致。
2. **杀毒软件** 可能误报便携 exe/dll，必要时加入白名单。
3. **模板图** 需自行准备；路径含中文时，请用 **引号** 包住路径。
4. 程序会 **实时截屏**，涉及隐私与权限，请在合规场景下使用。

如有问题，请对照源码仓库中的 `opencv/main.go` 与打包脚本 `opencv/package_go_opencv.ps1`。
