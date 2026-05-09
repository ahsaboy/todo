package models

import "encoding/json"

type UserReminderConfig struct {
	ID                  int64             `json:"id"`
	UserID              int64             `json:"user_id"`
	Name                string            `json:"name"`
	ChannelType         string            `json:"channel_type"`
	WebhookURL          string            `json:"webhook_url"`
	WebhookMethod       string            `json:"webhook_method"`
	WebhookHeaders      map[string]string `json:"webhook_headers"`
	WebhookBodyTemplate string            `json:"webhook_body_template"`
	MaxRetries          int               `json:"max_retries"`
	RetryDelaySeconds   int               `json:"retry_delay_seconds"`
	Enabled             bool              `json:"enabled"`
	CreatedAt           string            `json:"created_at"`
	UpdatedAt           string            `json:"updated_at"`
}

type CreateReminderConfigRequest struct {
	Name                string            `json:"name" binding:"required,max=64"`
	ChannelType         string            `json:"channel_type" binding:"required,oneof=webhook feishu dingtalk wecom slack"`
	WebhookURL          string            `json:"webhook_url" binding:"required,url"`
	WebhookMethod       string            `json:"webhook_method" binding:"omitempty,oneof=GET POST PUT"`
	WebhookHeaders      map[string]string `json:"webhook_headers"`
	WebhookBodyTemplate string            `json:"webhook_body_template"`
	MaxRetries          *int              `json:"max_retries" binding:"omitempty,min=0,max=10"`
	RetryDelaySeconds   *int              `json:"retry_delay_seconds" binding:"omitempty,min=1,max=300"`
	Enabled             *bool             `json:"enabled"`
}

type UpdateReminderConfigRequest struct {
	Name                *string           `json:"name" binding:"omitempty,max=64"`
	ChannelType         *string           `json:"channel_type" binding:"omitempty,oneof=webhook feishu dingtalk wecom slack"`
	WebhookURL          *string           `json:"webhook_url" binding:"omitempty,url"`
	WebhookMethod       *string           `json:"webhook_method" binding:"omitempty,oneof=GET POST PUT"`
	WebhookHeaders      map[string]string `json:"webhook_headers"`
	WebhookBodyTemplate *string           `json:"webhook_body_template"`
	MaxRetries          *int              `json:"max_retries" binding:"omitempty,min=0,max=10"`
	RetryDelaySeconds   *int              `json:"retry_delay_seconds" binding:"omitempty,min=1,max=300"`
	Enabled             *bool             `json:"enabled"`
}

func (c *UserReminderConfig) GetWebhookHeadersJSON() string {
	if c.WebhookHeaders == nil {
		return "{}"
	}
	b, _ := json.Marshal(c.WebhookHeaders)
	return string(b)
}

func ParseWebhookHeaders(raw string) map[string]string {
	if raw == "" {
		return nil
	}
	var h map[string]string
	json.Unmarshal([]byte(raw), &h)
	return h
}
