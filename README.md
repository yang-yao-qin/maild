# maild

maild 是一个本地邮件发送器。

它不是邮件客户端，不负责收件，也不管理邮件。

它仅负责：

- 编写 Markdown 邮件
- 渲染 HTML 邮件
- 调用邮件 API（目前支持 Resend）
- 使用自己的域名身份发送邮件

## 配置

项目不会包含 `config.toml`。

请复制：

```bash
cp config.example.toml config.toml
```

然后填写自己的配置。

其中：

```toml
api_key = "你的 Resend API Key"
```

请不要将 `config.toml` 提交到 Git 仓库。

项目已默认在 `.gitignore` 中忽略该文件。

## 启动

```bash
go run ./cmd/maild
```

浏览器打开：

```
http://127.0.0.1:8080
```

即可开始编写和发送邮件。
