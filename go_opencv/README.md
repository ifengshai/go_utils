# example_opencv 工具包说明

本目录为 **已编译的屏幕模板匹配工具** 及其 **运行时依赖（`lib\` 下 DLL）**。可在任意 **64 位 Windows** 机器上拷贝整夹使用（无需安装 Go/OpenCV 开发环境）。

---

## 一、为什么用 `run_example_opencv.bat` 启动

依赖 DLL 放在 **`lib\`** 子目录中。Windows 默认 **不会** 自动从 `lib` 加载 DLL，因此：

- **推荐**：双击或在命令行运行 **`run_example_opencv.bat`**，脚本会把 `lib` 临时加入 `PATH` 再启动 `example_opencv.exe`。
- **不推荐**：直接双击 `example_opencv.exe`（未设置 PATH 时通常会提示缺少 DLL）。

---

## 二、如何调用（命令行）

在 **`go_opencv` 目录** 下打开 PowerShell 或 CMD，执行：

```bat
run_example_opencv.bat <模板图片路径> [可选参数...]
```

所有传给 bat 的参数会原样传给 `example_opencv.exe`。

**也可**先设置环境变量再运行 exe（效果与 bat 相同）：

```bat
cd /d D:\路径\go_opencv
set PATH=%CD%\lib;%PATH%
example_opencv.exe "D:\图片\模板.png"
```

### PowerShell 与中文路径

在 PowerShell 中请为含空格或中文的路径 **加引号**。若仍异常，可让 CMD 代为解析参数：

```powershell
cmd /c 'run_example_opencv.bat "D:\game\...\加号.png"'
```

### 若出现「`...\go_opencv\' 不是内部或外部命令」

说明 **`run_example_opencv.bat` 内容损坏**（启动 exe 的那一行被拆成多行）。请用仓库里最新版 bat 覆盖，或在源码仓库根目录重新执行 **`scripts\package_go_opencv.ps1`** 以重新生成该文件。

---

## 三、程序做什么（简要）

1. 读取你提供的 **模板小图**（PNG/JPEG 等常见格式，路径支持中文）。
2. 截取当前 **所有显示器** 的画面。
3. 在灰度图上用 OpenCV **多尺度模板匹配**（`TM_CCOEFF_NORMED`），找出与模板相似的区域；多块屏、同屏多实例可有多条结果。
4. 在 **标准输出（stdout）** 打印 **JSON 数组**；错误与提示在 **标准错误（stderr）**。

---

## 四、输入说明

### 4.1 位置参数（必填）

| 顺序 | 含义 |
|------|------|
| 第 1 个 | 模板图片路径（相对路径或绝对路径均可） |

### 4.2 可选参数（flag）

| 参数 | 默认值 | 含义 |
|------|--------|------|
| `-min` | 0.65 | 模板相对屏幕的 **最小** 缩放比例（应对 DPI/分辨率差异） |
| `-max` | 1.45 | 模板相对屏幕的 **最大** 缩放比例 |
| `-steps` | 33 | 在 `[min,max]` 之间线性采样的缩放档位数，越大越慢、越细 |
| `-thresh` | 0.72 | 匹配得分 **峰值检测** 下限（用于在得分图上找候选） |
| `-minscore` | 0.80 | **仅输出** 得分 ≥ 该值的命中；若高于 `-thresh`，内部检测也会用二者较大者 |

说明：`-thresh` 与 `-minscore` 同时存在时，内部会用 **`max(-thresh, -minscore)`** 作为峰值门槛，减少算了又丢弃的低分候选。

---

## 五、输出说明

### 5.1 标准输出（stdout）

成功时输出 **一个 JSON 数组**（可能只有 1 个元素），元素字段如下：

| 字段 | 类型 | 含义 |
|------|------|------|
| `displayIndex` | 整数 | 命中所在显示器序号（从 0 开始） |
| `scale` | 浮点 | 本次命中使用的模板相对原图的缩放比例 |
| `score` | 浮点 | 相似度（归一化相关系数，约 -1～1，**越大越像**） |
| `left` / `top` / `right` / `bottom` | 整数 | 命中矩形在 **虚拟桌面坐标** 下的像素边界（含边界） |

数组按 **`score` 从高到低** 排序。

### 5.2 标准错误（stderr）

模板无法读取、未找到匹配、参数错误等信息输出到 stderr，便于脚本把「人读日志」与「机器读 JSON」分开。

### 5.3 退出码（简要）

| 退出码 | 常见情况 |
|--------|----------|
| 0 | 成功，stdout 有合法 JSON |
| 1 | 运行失败或未找到满足条件的匹配 |
| 2 | 用法错误（例如未提供模板路径） |

---

## 六、使用示例

### 6.1 基本用法（默认只输出 score ≥ 0.8）

```bat
cd /d D:\你的路径\go_opencv
run_example_opencv.bat "D:\截图\加号.png"
```

### 6.2 调整灵敏度与缩放范围

```bat
run_example_opencv.bat "D:\截图\按钮.png" -min 0.75 -max 1.35 -steps 40 -thresh 0.70 -minscore 0.85
```

### 6.3 在 PowerShell 里把 JSON 存成文件

```powershell
Set-Location D:\你的路径\go_opencv
.\run_example_opencv.bat "D:\截图\模板.png" | Out-File -Encoding utf8 result.json
```

注意：若程序在 stderr 打印了警告，**不要**把 stderr 重定向进同一个 JSON 文件，否则会破坏格式。仅重定向 stdout 即可。

### 6.4 成功时 stdout 示例

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

## 七、重新打包本目录

若你在开发机更新了源码或依赖，可在仓库根目录执行（路径按本机修改）：

```powershell
.\scripts\package_go_opencv.ps1 -OpenCVBin "C:\opencv\build\install\x64\mingw\bin" -MingwBin "C:\msys64\mingw64\bin"
```

脚本会重新生成 `example_opencv.exe`、`lib\` 下 DLL，并刷新 `run_example_opencv.bat`。

---

## 八、注意事项

1. **本包为 64 位**：请在 **64 位 Windows** 上使用；与编译所用 MinGW/OpenCV 位数一致。
2. **杀毒软件** 可能误报便携 exe/dll，必要时加入白名单。
3. **模板图** 需自行准备；路径含中文时，请用 **引号** 包住路径。
4. 程序会 **实时截屏**，涉及隐私与权限，请在合规场景下使用。

如有问题，请对照源码仓库中的 `example/example_opencv.go` 与打包脚本 `scripts/package_go_opencv.ps1`。
