// Package mail defines the core email model used throughout the application.
package mail

// Attachment represents a file attached to an email.
type Attachment struct {
	Filename    string
	ContentType string
	Content     []byte
}

// Mail represents an email message ready to be sent.
type Mail struct {
	From        string
	To          string
	Subject     string
	HTMLBody    string
	Attachments []Attachment
}
