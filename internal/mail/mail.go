// Package mail defines the core email model used throughout the application.
package mail

// Mail represents an email message ready to be sent.
type Mail struct {
	From     string
	To       string
	Subject  string
	HTMLBody string
}
