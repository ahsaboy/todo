package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"database/sql"
	"encoding/hex"
	"fmt"
	"math/big"
	"net"
	"net/smtp"
	"strings"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	"todo/internal/config"
)

const (
	codeExpiryMinutes = 10
	maxCodeAttempts   = 5
	cooldownSeconds   = 60
)

type EmailService struct {
	db      *sql.DB
	cfg     *config.Config
	logger  *zap.Logger
	enabled atomic.Bool
}

func NewEmailService(db *sql.DB, cfg *config.Config, logger *zap.Logger) *EmailService {
	s := &EmailService{
		db:     db,
		cfg:    cfg,
		logger: logger,
	}
	s.enabled.Store(cfg.Email.Enabled)
	return s
}

func (s *EmailService) IsEnabled() bool {
	return s.enabled.Load() && s.cfg.Email.SMTPHost != "" && s.cfg.Email.FromAddress != ""
}

func (s *EmailService) SetEnabled(b bool) {
	s.enabled.Store(b)
}

func (s *EmailService) smtpAddr() string {
	return fmt.Sprintf("%s:%d", s.cfg.Email.SMTPHost, s.cfg.Email.SMTPPort)
}

func (s *EmailService) sendEmail(to, subject, body string) error {
	cfg := s.cfg.Email
	addr := s.smtpAddr()

	var auth smtp.Auth
	if cfg.SMTPUsername != "" {
		auth = smtp.PlainAuth("", cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPHost)
	}

	from := cfg.FromAddress
	if cfg.FromName != "" {
		from = fmt.Sprintf("%s <%s>", cfg.FromName, cfg.FromAddress)
	}

	msg := strings.Join([]string{
		"From: " + from,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	if cfg.SMTPPort == 465 {
		// 465 端口：隐式 SSL
		tlsConn, err := tls.Dial("tcp", addr, &tls.Config{
			ServerName: cfg.SMTPHost,
		})
		if err != nil {
			return fmt.Errorf("TLS dial: %w", err)
		}

		client, err := smtp.NewClient(tlsConn, cfg.SMTPHost)
		if err != nil {
			tlsConn.Close()
			return fmt.Errorf("smtp client: %w", err)
		}
		defer client.Close()

		if auth != nil {
			if err := client.Auth(auth); err != nil {
				return fmt.Errorf("auth: %w", err)
			}
		}

		if err := client.Mail(cfg.FromAddress); err != nil {
			return fmt.Errorf("mail: %w", err)
		}
		if err := client.Rcpt(to); err != nil {
			return fmt.Errorf("rcpt: %w", err)
		}

		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("data: %w", err)
		}
		if _, err := w.Write([]byte(msg)); err != nil {
			return fmt.Errorf("write: %w", err)
		}
		if err := w.Close(); err != nil {
			return fmt.Errorf("close: %w", err)
		}

		return client.Quit()
	}

	// 其他端口：使用标准 SendMail（支持 STARTTLS）
	return smtp.SendMail(addr, auth, cfg.FromAddress, []string{to}, []byte(msg))
}

func (s *EmailService) generateCode() (string, error) {
	code := make([]byte, 6)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", fmt.Errorf("generate random digit: %w", err)
		}
		code[i] = byte('0' + n.Int64())
	}
	return string(code), nil
}

func hashCode(code string) string {
	h := sha256.Sum256([]byte(code))
	return hex.EncodeToString(h[:])
}

func (s *EmailService) SendVerificationCode(ctx context.Context, email string) error {
	if !s.IsEnabled() {
		return ErrEmailNotConfigured
	}

	// 统一 email 小写
	email = strings.ToLower(email)

	now := time.Now().UTC()
	cooldownThreshold := now.Add(-time.Duration(cooldownSeconds) * time.Second).Format(time.RFC3339)

	// 冷却检查：同 email 60 秒内不能重复发送
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM email_verification_codes
		 WHERE email = ? AND created_at > ?`,
		email, cooldownThreshold,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("check cooldown: %w", err)
	}
	if count > 0 {
		return ErrCodeCooldown
	}

	code, err := s.generateCode()
	if err != nil {
		return err
	}

	expiresAt := now.Add(time.Duration(codeExpiryMinutes) * time.Minute).Format(time.RFC3339)
	codeHash := hashCode(code)

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO email_verification_codes (email, code_hash, max_attempts, expires_at, created_at)
		 VALUES (?, ?, ?, ?, ?)`,
		email, codeHash, maxCodeAttempts, expiresAt, now.Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("insert code: %w", err)
	}

	subject := "验证码"
	body := fmt.Sprintf(`<p>您的验证码是：<strong>%s</strong></p><p>%d 分钟内有效。</p>`, code, codeExpiryMinutes)

	if err := s.sendEmail(email, subject, body); err != nil {
		s.logger.Error("发送验证码邮件失败", zap.String("email", email), zap.Error(err))
		return fmt.Errorf("send email: %w", err)
	}

	return nil
}

func (s *EmailService) VerifyCode(ctx context.Context, email, code string) error {
	// 统一 email 小写
	email = strings.ToLower(email)

	var id int
	var codeHash string
	var attempts, maxAttempts int
	var used bool
	var expiresAt string

	err := s.db.QueryRowContext(ctx,
		`SELECT id, code_hash, attempts, max_attempts, used, expires_at
		 FROM email_verification_codes
		 WHERE email = ?
		 ORDER BY id DESC LIMIT 1`,
		email,
	).Scan(&id, &codeHash, &attempts, &maxAttempts, &used, &expiresAt)

	if err == sql.ErrNoRows {
		return ErrCodeNotFound
	}
	if err != nil {
		return fmt.Errorf("query code: %w", err)
	}

	if used {
		return ErrCodeUsed
	}

	expires, err := time.Parse(time.RFC3339, expiresAt)
	if err != nil {
		return fmt.Errorf("parse expires_at: %w", err)
	}
	if time.Now().UTC().After(expires) {
		return ErrCodeExpired
	}

	if attempts >= maxAttempts {
		return ErrCodeAttemptsExceeded
	}

	// 递增尝试次数
	_, err = s.db.ExecContext(ctx,
		`UPDATE email_verification_codes SET attempts = attempts + 1 WHERE id = ?`, id,
	)
	if err != nil {
		return fmt.Errorf("update attempts: %w", err)
	}

	if hashCode(code) != codeHash {
		return ErrCodeInvalid
	}

	_, err = s.db.ExecContext(ctx,
		`UPDATE email_verification_codes SET used = 1 WHERE id = ?`, id,
	)
	if err != nil {
		return fmt.Errorf("mark used: %w", err)
	}

	return nil
}

func (s *EmailService) TestConnection(_ context.Context) error {
	addr := s.smtpAddr()
	host := s.cfg.Email.SMTPHost

	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("无法连接到 %s: %w", addr, err)
	}
	conn.Close()

	var client *smtp.Client
	if s.cfg.Email.SMTPPort == 465 {
		// 465 端口：隐式 SSL，直接建立 TLS 连接
		tlsConn, err := tls.Dial("tcp", addr, &tls.Config{
			ServerName: host,
		})
		if err != nil {
			return fmt.Errorf("TLS 握手失败: %w", err)
		}
		client, err = smtp.NewClient(tlsConn, host)
		if err != nil {
			tlsConn.Close()
			return fmt.Errorf("SMTP 客户端创建失败: %w", err)
		}
	} else {
		// 其他端口：明文连接，后续 STARTTLS
		client, err = smtp.Dial(addr)
		if err != nil {
			return fmt.Errorf("SMTP 握手失败: %w", err)
		}
	}
	defer client.Close()

	if err := client.Hello("localhost"); err != nil {
		return fmt.Errorf("HELO 失败: %w", err)
	}

	if s.cfg.Email.SMTPUsername != "" {
		auth := smtp.PlainAuth("", s.cfg.Email.SMTPUsername, s.cfg.Email.SMTPPassword, host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("认证失败: %w", err)
		}
	}

	return nil
}
