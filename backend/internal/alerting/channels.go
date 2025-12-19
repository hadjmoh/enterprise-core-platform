package alerting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"time"
)

// SlackChannel sends alerts to Slack
type SlackChannel struct {
	webhookURL string
	client     *http.Client
}

func NewSlackChannel(webhookURL string) *SlackChannel {
	return &SlackChannel{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *SlackChannel) Send(alert *Alert) error {
	color := s.getSeverityColor(alert.Severity)
	
	payload := map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"color":      color,
				"title":      alert.Title,
				"text":       alert.Description,
				"fields": []map[string]interface{}{
					{"title": "Severity", "value": alert.Severity, "short": true},
					{"title": "Source", "value": alert.Source, "short": true},
				},
				"footer":    "Enterprise Core Platform",
				"ts":        alert.Timestamp.Unix(),
			},
		},
	}

	data, _ := json.Marshal(payload)
	resp, err := s.client.Post(s.webhookURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("slack returned status %d", resp.StatusCode)
	}

	return nil
}

func (s *SlackChannel) Name() string {
	return "Slack"
}

func (s *SlackChannel) getSeverityColor(severity string) string {
	switch severity {
	case "critical":
		return "#ff0000"
	case "high":
		return "#ff6600"
	case "medium":
		return "#ffcc00"
	default:
		return "#36a64f"
	}
}

// PagerDutyChannel sends alerts to PagerDuty
type PagerDutyChannel struct {
	integrationKey string
	client         *http.Client
}

func NewPagerDutyChannel(integrationKey string) *PagerDutyChannel {
	return &PagerDutyChannel{
		integrationKey: integrationKey,
		client:         &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *PagerDutyChannel) Send(alert *Alert) error {
	payload := map[string]interface{}{
		"routing_key":  p.integrationKey,
		"event_action": "trigger",
		"payload": map[string]interface{}{
			"summary":   alert.Title,
			"severity":  alert.Severity,
			"source":    alert.Source,
			"timestamp": alert.Timestamp.Format(time.RFC3339),
		},
	}

	data, _ := json.Marshal(payload)
	resp, err := p.client.Post("https://events.pagerduty.com/v2/enqueue", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 202 {
		return fmt.Errorf("pagerduty returned status %d", resp.StatusCode)
	}

	return nil
}

func (p *PagerDutyChannel) Name() string {
	return "PagerDuty"
}

// EmailChannel sends alerts via email
type EmailChannel struct {
	smtpHost string
	smtpPort string
	from     string
	to       []string
	auth     smtp.Auth
}

func NewEmailChannel(smtpHost, smtpPort, username, password, from string, to []string) *EmailChannel {
	return &EmailChannel{
		smtpHost: smtpHost,
		smtpPort: smtpPort,
		from:     from,
		to:       to,
		auth:     smtp.PlainAuth("", username, password, smtpHost),
	}
}

func (e *EmailChannel) Send(alert *Alert) error {
	subject := fmt.Sprintf("[%s] %s", alert.Severity, alert.Title)
	body := fmt.Sprintf("Severity: %s\nSource: %s\nTime: %s\n\n%s",
		alert.Severity,
		alert.Source,
		alert.Timestamp.Format(time.RFC3339),
		alert.Description,
	)

	msg := []byte(fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body))

	addr := e.smtpHost + ":" + e.smtpPort
	return smtp.SendMail(addr, e.auth, e.from, e.to, msg)
}

func (e *EmailChannel) Name() string {
	return "Email"
}

// WebhookChannel sends alerts to generic webhooks
type WebhookChannel struct {
	url    string
	client *http.Client
}

func NewWebhookChannel(url string) *WebhookChannel {
	return &WebhookChannel{
		url:    url,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (w *WebhookChannel) Send(alert *Alert) error {
	data, _ := json.Marshal(alert)
	resp, err := w.client.Post(w.url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

func (w *WebhookChannel) Name() string {
	return "Webhook"
}
