package provider

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"maild/internal/mail"
)

const resendAPIURL = "https://api.resend.com/emails"

// ResendProvider sends emails via the Resend API.
type ResendProvider struct {
	apiKey     string
	httpClient *http.Client
}

// NewResendProvider creates a new ResendProvider with the given API key.
// The HTTP client is configured to respect the standard proxy environment variables
// (HTTP_PROXY, HTTPS_PROXY, NO_PROXY) via http.ProxyFromEnvironment.
func NewResendProvider(apiKey string) *ResendProvider {
	return &ResendProvider{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				Proxy:                 http.ProxyFromEnvironment,
				DialContext:           (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
	}
}

// resendRequest is the JSON payload sent to the Resend API.
type resendRequest struct {
	From        string             `json:"from"`
	To          string             `json:"to"`
	Subject     string             `json:"subject"`
	HTML        string             `json:"html"`
	Attachments []resendAttachment `json:"attachments,omitempty"`
}

// resendAttachment is a single attachment in the Resend API format.
type resendAttachment struct {
	Filename string `json:"filename"`
	Content  string `json:"content"` // base64-encoded
}

// Send delivers an email via the Resend API.
func (p *ResendProvider) Send(m mail.Mail) error {
	payload := resendRequest{
		From:    m.From,
		To:      m.To,
		Subject: m.Subject,
		HTML:    m.HTMLBody,
	}

	if len(m.Attachments) > 0 {
		payload.Attachments = make([]resendAttachment, len(m.Attachments))
		for i, a := range m.Attachments {
			payload.Attachments[i] = resendAttachment{
				Filename: a.Filename,
				Content:  base64.StdEncoding.EncodeToString(a.Content),
			}
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, resendAPIURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	// Drain the response body to allow HTTP connection reuse.
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("resend API returned status %d", resp.StatusCode)
	}

	return nil
}
