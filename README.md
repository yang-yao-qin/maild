# maild

基于 Resend API 的本地 Markdown 邮件发送客户端。

maild 不是邮件客户端。它不收信、不存邮件、不实现 IMAP / SMTP。
它只做一件事：把 Markdown 写成邮件，调用 API 发出。

## 项目结构

```
cmd/maild/         入口
internal/
  config/          解析 config.toml
  mail/            邮件模型
  provider/        MailProvider 接口与 Resend 实现
  markdown/        Markdown → HTML
  server/          HTTP API 与静态文件服务
web/               WebUI（HTML + JS + CSS）
```

## 快速开始

```bash
cp config.example.toml config.toml
```

编辑 `config.toml`，填写 Resend API Key 和发件身份。

```bash
make build
./maild
```

浏览器打开 `http://127.0.0.1:8080`。

## 配置

`config.toml` 不在版本控制中（`.gitignore` 已忽略）。
仓库提供 `config.example.toml` 作为模板。

```toml
[server]
address = "127.0.0.1:8080"

[resend]
api_key = "re_xxxx"

[senders]
me = "me@example.org"
contact = "contact@example.org"
```

`[senders]` 中的身份会出现在 WebUI 的发件人下拉框中。前端只能选择已配置的身份，不能任意填写 `From` 地址。

## 使用

1. 打开 `http://127.0.0.1:8080`
2. 选择发件身份，填写收件人和主题
3. 用 Markdown 编写正文
4. 点击发送

邮件通过 Resend API 发出，使用你在 `config.toml` 中配置的域名身份。

## 工作流程

```
Markdown
  ↓  goldmark
HTML
  ↓  Resend API (HTTPS)
收件人
```

maild 不保存邮件副本。不提供发件历史。发送结果通过 WebUI 即时反馈。

## 作为服务运行

maild 设计为长期运行的本地服务。推荐使用 systemd 用户单元：

```ini
# ~/.config/systemd/user/maild.service
[Unit]
Description=maild

[Service]
ExecStart=%h/src/maild/maild
WorkingDirectory=%h/src/maild
Restart=on-failure

[Install]
WantedBy=default.target
```

```bash
systemctl --user daemon-reload
systemctl --user enable maild
systemctl --user start maild

# 查看日志
journalctl --user -u maild -f
```

## Provider

目前仅实现 Resend。`MailProvider` 接口定义在 `internal/provider/provider.go`：

```go
type MailProvider interface {
    Send(mail.Mail) error
}
```

添加新 Provider（SES、Mailgun、Postmark 等）只需实现该接口。

## 构建

```bash
make build        # 构建
make run          # 构建并运行
make build-static # 静态链接（CGO_ENABLED=0）
make clean        # 清理
```

## Roadmap

- 草稿保存
- HTML 邮件模板自定义
- 多 Provider 支持（SES、Mailgun）
- 附件支持

## License

GPL-3.0
