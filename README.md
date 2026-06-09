# ItsMyTo

ItsMyTo 是一个本地 API Key 管理软件：Go 后端、内嵌前端、轻量桌面 WebView 窗口。

## 运行

```bash
go run -mod=vendor .
```

构建精简桌面二进制：

```bash
go build -mod=vendor -ldflags="-s -w" -o itsmyto .
./itsmyto
```

默认会打开独立应用窗口。调试接口时可使用：

```bash
./itsmyto -serve-only -addr 127.0.0.1:39007
```

## Linux 依赖

当前项目 vendor 了 `webview_go`，并将 Linux pkg-config 目标修正为 `webkit2gtk-4.1`。如果重新执行 `go mod vendor`，需要确认 `vendor/github.com/webview/webview_go/webview.go` 仍使用 `webkit2gtk-4.1`。

## 安全存储

- 业务 API Key 使用 AES-256-GCM 加密保存。
- Embedding API Key 同样加密保存，设置接口只返回 `hasEmbeddingApiKey`。
- 本地随机主密钥保存在用户配置目录下的 `ItsMyTo/master.key`，权限为 `0600`。
- 可用 `ITSMYTO_MASTER_KEY` 接管主密钥，值必须是 base64 编码的 32 字节密钥。

## 搜索

前端只有一个搜索框，后端同时尝试：

- 关键词硬匹配：Name、APIKEY、BaseURL、详细介绍。
- 正则匹配：Name、APIKEY、详细介绍。
- 余弦相似度匹配：Name、详细介绍。

未配置 Embedding API URL 和模型名称时，余弦相似度匹配自动跳过。
