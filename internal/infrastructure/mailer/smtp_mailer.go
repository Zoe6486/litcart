package mailer

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"html/template"
	"net"
	"net/smtp"
	"strconv"
	"time"

	"litcart/internal/user/domain"
)

// SMTPConfig 是 SMTP 连接配置。
//
// 常见提供商:
//
//	Gmail:     smtp.gmail.com:587  (StartTLS) 或 465 (SMTPS)
//	             需要"应用专用密码",普通密码用不了
//	QQ 邮箱:   smtp.qq.com:587 / 465
//	             密码是"授权码",不是邮箱密码
//	阿里云:    smtpdm.aliyun.com:80 / 465
//	SendGrid:  smtp.sendgrid.net:587  username 写死 "apikey",password 是 API Key
type SMTPConfig struct {
	Host        string        // smtp.example.com
	Port        int           // 587 (StartTLS) 或 465 (SMTPS)
	Username    string        // 登录用户名(通常是邮箱)
	Password    string        // 登录密码 / 授权码 / API Key
	FromAddr    string        // 发件人地址(必须与登录账号匹配,否则被拒)
	FromName    string        // 发件人显示名,如 "LitCart"
	UseSMTPS    bool          // true 用 465 端口的隐式 TLS,false 用 587 的 StartTLS
	DialTimeout time.Duration // 连接 + 发送超时,默认 10s
}

// SMTPMailer 实现 domain.Mailer,用 net/smtp 发邮件。
//
// 设计要点:
//   - 模板预编译:NewSMTPMailer 时把 HTML 模板 Parse 一次,后续每次发邮件复用。
//     模板写错会在启动时直接 panic,不会拖到运行时才发现。
//   - 超时控制:用 net.DialTimeout + ctx,避免 SMTP 服务器卡住把请求堆爆。
//   - 邮件发送是阻塞 IO,实际生产建议:
//     (a) 异步发送(放消息队列),让 HTTP 请求立刻返回
//     (b) 或至少加 worker pool 限并发
//     这里保持简单,同步发送。Service 层已经把"邮件失败不影响主流程"做了。
type SMTPMailer struct {
	cfg           SMTPConfig
	verifyURLBase string
	resetURLBase  string
	verifyTmpl    *template.Template
	resetTmpl     *template.Template
}

var _ domain.Mailer = (*SMTPMailer)(nil)

func NewSMTPMailer(cfg SMTPConfig, verifyURLBase, resetURLBase string) *SMTPMailer {
	if cfg.DialTimeout <= 0 {
		cfg.DialTimeout = 10 * time.Second
	}
	return &SMTPMailer{
		cfg:           cfg,
		verifyURLBase: verifyURLBase,
		resetURLBase:  resetURLBase,
		verifyTmpl:    template.Must(template.New("verify").Parse(verifyEmailTmpl)),
		resetTmpl:     template.Must(template.New("reset").Parse(resetEmailTmpl)),
	}
}

func (m *SMTPMailer) SendVerifyEmail(ctx context.Context, to domain.Email, token string) error {
	body, err := renderHTML(m.verifyTmpl, map[string]string{
		"Link": m.verifyURLBase + "?token=" + token,
	})
	if err != nil {
		return fmt.Errorf("render verify mail: %w", err)
	}
	return m.send(ctx, to, "Please verify your email", body)
}

func (m *SMTPMailer) SendPasswordResetEmail(ctx context.Context, to domain.Email, token string) error {
	body, err := renderHTML(m.resetTmpl, map[string]string{
		"Link": m.resetURLBase + "?token=" + token,
	})
	if err != nil {
		return fmt.Errorf("render reset mail: %w", err)
	}
	return m.send(ctx, to, "Reset your password", body)
}

// send 是真正发邮件的逻辑。
//
// 流程:
//  1. 拨号(StartTLS 或 SMTPS)
//  2. SMTP 握手 + 认证
//  3. 设置 from / to,写邮件 DATA
//  4. 关闭连接
//
// 用 ctx.Done() + 单独 goroutine 实现"软超时":context 取消时主动 close 连接,
// 否则 net/smtp 自己不感知 ctx,会一直阻塞到 TCP 超时。
func (m *SMTPMailer) send(ctx context.Context, to domain.Email, subject, htmlBody string) error {
	addr := net.JoinHostPort(m.cfg.Host, strconv.Itoa(m.cfg.Port))

	dialer := &net.Dialer{Timeout: m.cfg.DialTimeout}
	var conn net.Conn
	var err error

	if m.cfg.UseSMTPS {
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{ServerName: m.cfg.Host})
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}

	// ctx 取消时强制断开,跳出可能卡住的 IO
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			_ = conn.Close()
		case <-done:
		}
	}()

	client, err := smtp.NewClient(conn, m.cfg.Host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer func() { _ = client.Quit() }()

	// StartTLS 模式(587):明文连进去后升级到 TLS
	if !m.cfg.UseSMTPS {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(&tls.Config{ServerName: m.cfg.Host}); err != nil {
				return fmt.Errorf("smtp starttls: %w", err)
			}
		} else {
			return errors.New("smtp: server does not support STARTTLS; use UseSMTPS for port 465")
		}
	}

	// AUTH:大多数主流 SMTP 服务都支持 PLAIN
	auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}

	if err := client.Mail(m.cfg.FromAddr); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err := client.Rcpt(to.String()); err != nil {
		return fmt.Errorf("smtp rcpt to: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}

	msg := buildMessage(m.cfg.FromName, m.cfg.FromAddr, to.String(), subject, htmlBody)
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp close data: %w", err)
	}
	return nil
}

// buildMessage 拼出符合 RFC 5322 的邮件原文。
//
// 注意点:
//   - From 用 "Name <addr>" 格式,显示名才能正确显示
//   - MIME-Version + Content-Type 让客户端按 HTML 渲染而不是显示原始代码
//   - 头部之间用 CRLF(\r\n),头部和正文之间空一行
func buildMessage(fromName, fromAddr, to, subject, htmlBody string) []byte {
	var buf bytes.Buffer
	if fromName != "" {
		fmt.Fprintf(&buf, "From: %s <%s>\r\n", fromName, fromAddr)
	} else {
		fmt.Fprintf(&buf, "From: %s\r\n", fromAddr)
	}
	fmt.Fprintf(&buf, "To: %s\r\n", to)
	fmt.Fprintf(&buf, "Subject: %s\r\n", subject)
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(htmlBody)
	return buf.Bytes()
}

func renderHTML(t *template.Template, data any) (string, error) {
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// ---- 邮件 HTML 模板 ----
// 简单内联样式,跨客户端兼容性最好(Gmail、Outlook 都不支持 <style> 块的部分特性)

const verifyEmailTmpl = `<!doctype html>
<html><body style="font-family:Arial,sans-serif;line-height:1.6;color:#333">
  <h2>Verify your email</h2>
  <p>Welcome! Please click the link below to verify your email address:</p>
  <p><a href="{{.Link}}" style="display:inline-block;padding:10px 20px;background:#1677ff;color:#fff;text-decoration:none;border-radius:4px">Verify Email</a></p>
  <p>Or copy this URL into your browser:</p>
  <p style="color:#666;word-break:break-all">{{.Link}}</p>
  <p style="color:#999;font-size:12px">If you did not create an account, you can safely ignore this email.</p>
</body></html>`

const resetEmailTmpl = `<!doctype html>
<html><body style="font-family:Arial,sans-serif;line-height:1.6;color:#333">
  <h2>Reset your password</h2>
  <p>We received a request to reset your password. Click the link below to set a new one:</p>
  <p><a href="{{.Link}}" style="display:inline-block;padding:10px 20px;background:#1677ff;color:#fff;text-decoration:none;border-radius:4px">Reset Password</a></p>
  <p>Or copy this URL into your browser:</p>
  <p style="color:#666;word-break:break-all">{{.Link}}</p>
  <p style="color:#999;font-size:12px">If you did not request this, you can safely ignore this email — your password will not change.</p>
</body></html>`
