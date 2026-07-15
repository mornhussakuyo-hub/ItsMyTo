# ItsMyTo

<div align="center">

<pre>
 ______  __               __       __           ________         __ 
/      |/  |             /  \     /  |         /        |       /  |
$$$$$$/_$$ |_    _______ $$  \   /$$ | __    __$$$$$$$$/______  $$ |
  $$ |/ $$   |  /       |$$$  \ /$$$ |/  |  /  |  $$ | /      \ $$ |
  $$ |$$$$$$/  /$$$$$$$/ $$$$  /$$$$ |$$ |  $$ |  $$ |/$$$$$$  |$$ |
  $$ |  $$ | __$$      \ $$ $$ $$/$$ |$$ |  $$ |  $$ |$$ |  $$ |$$/ 
 _$$ |_ $$ |/  |$$$$$$  |$$ |$$$/ $$ |$$ \__$$ |  $$ |$$ \__$$ | __ 
/ $$   |$$  $$//     $$/ $$ | $/  $$ |$$    $$ |  $$ |$$    $$/ /  |
$$$$$$/  $$$$/ $$$$$$$/  $$/      $$/  $$$$$$$ |  $$/  $$$$$$/  $$/ 
                                      /  \__$$ |                    
                                      $$    $$/                     
                                       $$$$$$/                     
</pre>

</div>

ItsMyTo (IMT) 是一个本地 API Key 管理桌面应用。它使用 Go 提供本地服务和加密存储，使用内嵌前端运行在轻量 WebView 窗口中，不依赖浏览器作为主界面。

名称全写为： It 's my token !

## 功能

- 卡片式管理 API Key：名称、BaseURL、APIKEY、详细介绍、文档地址。
- API Key 和 Embedding API Key 均使用 AES-256-GCM 加密落盘。
- 支持编辑、归档、取消归档、删除、显示和复制 Key。
- 搜索由点击搜索按钮或按 Enter 触发，不会边输入边请求。
- 后端并行执行关键词、正则、向量三路搜索，并通过 SSE 流式推送结果。
- 前端按卡片 ID 去重展示多路搜索结果。
- 支持深色、浅色、跟随系统主题。
- 支持开机启动。
- 支持配置 OpenAI-compatible Embedding 接口、模型名、API Key、批次大小和单批最大 Token 数。

## 运行

下载 Release 中与你系统对应的产物后运行：

```bash
./itsmyto-linux-amd64
```

Windows 运行：

```powershell
.\itsmyto-windows-amd64.exe
```

默认会打开独立桌面窗口。调试本地 HTTP 接口时可以使用：

```bash
./itsmyto-linux-amd64 -serve-only -addr 127.0.0.1:39007
```

Windows 版本使用系统 WebView2。Windows 11 通常已内置；如果无法打开窗口，请安装 Microsoft Edge WebView2 Runtime。

## 本地构建

Linux：

```bash
go build -mod=vendor -ldflags="-s -w" -o dist/itsmyto-linux-amd64 .
```

Windows 发布版使用 GUI 子系统构建，避免启动时显示额外的控制台窗口：

```powershell
$env:CC = "C:\msys64\mingw64\bin\gcc.exe"
$env:CXX = "C:\msys64\mingw64\bin\g++.exe"
.\scripts\build-windows.ps1
```

构建脚本会检查生成文件的 PE 头，只有 Subsystem 为 `IMAGE_SUBSYSTEM_WINDOWS_GUI` 时才成功。推送 `v*` 标签后，GitHub Actions 会构建 Windows 和 Linux 产物、生成 SHA256 校验文件并发布 Release。

## Linux 依赖

Linux 桌面窗口依赖系统 WebKitGTK。当前项目 vendor 了 `webview_go`，并将 Linux pkg-config 目标修正为 `webkit2gtk-4.1`。如果重新执行 `go mod vendor`，需要确认：

```text
vendor/github.com/webview/webview_go/webview.go
```

中仍使用 `webkit2gtk-4.1`。

## 存储位置

应用数据默认保存在用户配置目录：

- Linux: `~/.config/ItsMyTo`
- Windows: `%AppData%\ItsMyTo`
- macOS: `~/Library/Application Support/ItsMyTo`

关键文件：

- `master.key`：本机随机主密钥，权限为 `0600`。
- `cards.json`：卡片数据，API Key 为密文。
- `settings.json`：设置数据，Embedding API Key 为密文。

也可以通过 `ITSMYTO_MASTER_KEY` 接管主密钥。该环境变量必须是 base64 编码的 32 字节密钥。

## 搜索机制

后端同时执行三路搜索：

- 关键词硬匹配：名称、APIKEY、BaseURL、详细介绍。
- 正则匹配：名称、APIKEY、详细介绍。
- 向量相似度匹配：名称、详细介绍。

搜索请求通过 SSE 流式返回。三路搜索分别在 goroutine 中执行，任何一路先命中都会先推送卡片，前端负责去重。

未配置 Embedding API URL 和模型名时，向量搜索会自动跳过。

## 向量缓存

配置 Embedding API URL 和模型名后，应用会立即为已有卡片预处理向量。新增或编辑卡片后，也会刷新对应卡片的向量缓存。

每张卡片会保存：

- `embeddingModel`：生成该向量时使用的模型名称。
- `embeddingHash`：由 `Name + 详细介绍` 计算的文本哈希。
- `embeddingVector`：缓存向量。

搜索时只使用当前模型名称一致、文本哈希一致的缓存向量。模型变更或卡片名称/介绍变更后，旧缓存自动失效并重新生成。

设置页可以配置批次大小和单批最大处理 Token 数。后端会按两个限制同时切批；如果 Embedding 服务拒绝某个批次，会自动把该批次二分成更小批次重试。单条文本仍失败时只跳过该条缓存，不阻塞其他卡片。

默认语义召回阈值为 `0.35`，适配短文本 API 卡片描述和 `text-embedding-v4` 等模型的常见分数区间。

## 安全说明

- API Key 不会明文写入 `cards.json`。
- Embedding API Key 不会明文写入 `settings.json`。
- 列表接口只返回 API Key 掩码。
- 只有显示或复制时，后端才会解密指定卡片的 API Key。
- 搜索 API 会在内存中临时解密 API Key 以支持关键词和正则匹配。

## 开发检查

```bash
go test ./...
go build -mod=vendor -ldflags="-s -w" -o itsmyto .
```
