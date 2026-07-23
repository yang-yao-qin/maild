// Package provider defines the MailProvider interface for sending emails.
// This abstraction allows swapping the underlying email API (Resend, SES, etc.)
// without changing the rest of the application.
package provider

import "maild/internal/mail"

// MailProvider is the interface that wraps the Send method.
//
// Send delivers a single email. It returns an error if the send fails.
// Implementations are responsible for authentication and API communication.
type MailProvider interface {
	Send(mail.Mail) error
}
