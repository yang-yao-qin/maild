// Package server implements the HTTP server for the maild composer.
//
// It serves the web UI and exposes a JSON API for sending emails.
// The server listens only on 127.0.0.1 — it is not designed for network exposure.
package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"maild/internal/config"
	"maild/internal/mail"
	"maild/internal/markdown"
	"maild/internal/provider"
)

// Server is the HTTP server for the mail composer.
type Server struct {
	cfg      *config.Config
	provider provider.MailProvider
	mux      *http.ServeMux
}

// sendRequest is the JSON body expected by POST /api/send.
type sendRequest struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Subject  string `json:"subject"`
	Markdown string `json:"markdown"`
}

// sendResponse is the JSON body returned by POST /api/send.
type sendResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// New creates a new http.Server with the given configuration and mail provider.
// webDir is the path to the directory containing static web assets (index.html, etc.).
func New(cfg *config.Config, p provider.MailProvider, webDir string) *http.Server {
	s := &Server{
		cfg:      cfg,
		provider: p,
		mux:      http.NewServeMux(),
	}

	s.registerRoutes(webDir)

	return &http.Server{
		Addr:    cfg.Server.Address,
		Handler: s.mux,
	}
}

// registerRoutes sets up all HTTP routes.
func (s *Server) registerRoutes(webDir string) {
	// Serve static web assets from the filesystem.
	s.mux.Handle("/", http.FileServer(http.Dir(webDir)))

	// API routes.
	s.mux.HandleFunc("/api/send", s.handleSend)
	s.mux.HandleFunc("/api/senders", s.handleSenders)
}

// handleSend processes POST /api/send requests.
// Accepts both application/json (no attachments) and multipart/form-data (with attachments).
func (s *Server) handleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeJSON(w, http.StatusMethodNotAllowed, sendResponse{
			Success: false,
			Message: "Method not allowed. Use POST.",
		})
		return
	}

	var (
		req         sendRequest
		attachments []mail.Attachment
	)

	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "multipart/form-data") {
		// Multipart: parse form fields and file uploads.
		maxMem := s.cfg.Server.MaxAttachmentSize.Bytes()
		if maxMem <= 0 {
			maxMem = 10 << 20
		}
		if err := r.ParseMultipartForm(maxMem); err != nil {
			s.writeJSON(w, http.StatusBadRequest, sendResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to parse form data: %v", err),
			})
			return
		}
		defer r.Body.Close()

		req.From = strings.TrimSpace(r.FormValue("from"))
		req.To = strings.TrimSpace(r.FormValue("to"))
		req.Subject = strings.TrimSpace(r.FormValue("subject"))
		req.Markdown = strings.TrimSpace(r.FormValue("markdown"))

		// Process uploaded files.
		files := r.MultipartForm.File["attachments"]
		var totalSize int64
		for _, fh := range files {
			totalSize += fh.Size
		}
		if totalSize > s.cfg.Server.MaxAttachmentSize.Bytes() {
			s.writeJSON(w, http.StatusBadRequest, sendResponse{
				Success: false,
				Message: fmt.Sprintf("Total attachment size (%d bytes) exceeds limit (%d bytes).",
					totalSize, s.cfg.Server.MaxAttachmentSize.Bytes()),
			})
			return
		}

		for _, fh := range files {
			file, err := fh.Open()
			if err != nil {
				s.writeJSON(w, http.StatusInternalServerError, sendResponse{
					Success: false,
					Message: fmt.Sprintf("Failed to read attachment %q: %v", fh.Filename, err),
				})
				return
			}
			content, err := io.ReadAll(file)
			file.Close()
			if err != nil {
				s.writeJSON(w, http.StatusInternalServerError, sendResponse{
					Success: false,
					Message: fmt.Sprintf("Failed to read attachment %q: %v", fh.Filename, err),
				})
				return
			}

			ct := mime.TypeByExtension(filepath.Ext(fh.Filename))
			if ct == "" {
				ct = http.DetectContentType(content)
			}

			attachments = append(attachments, mail.Attachment{
				Filename:    fh.Filename,
				ContentType: ct,
				Content:     content,
			})
		}
	} else {
		// JSON: original behavior, no attachments.
		body, err := io.ReadAll(r.Body)
		if err != nil {
			s.writeJSON(w, http.StatusBadRequest, sendResponse{
				Success: false,
				Message: "Failed to read request body.",
			})
			return
		}
		defer r.Body.Close()

		if err := json.Unmarshal(body, &req); err != nil {
			s.writeJSON(w, http.StatusBadRequest, sendResponse{
				Success: false,
				Message: "Invalid JSON. Expected: from, to, subject, markdown.",
			})
			return
		}

		req.From = strings.TrimSpace(req.From)
		req.To = strings.TrimSpace(req.To)
		req.Subject = strings.TrimSpace(req.Subject)
		req.Markdown = strings.TrimSpace(req.Markdown)
	}

	// --- validation (shared by both paths) ---

	if req.From == "" || req.To == "" || req.Subject == "" || req.Markdown == "" {
		s.writeJSON(w, http.StatusBadRequest, sendResponse{
			Success: false,
			Message: "All fields are required: from, to, subject, markdown.",
		})
		return
	}

	// Validate sender identity.
	fromAddr, ok := s.cfg.Senders[req.From]
	if !ok {
		valid := false
		for _, addr := range s.cfg.Senders {
			if addr == req.From {
				valid = true
				fromAddr = addr
				break
			}
		}
		if !valid {
			log.Printf("WARN  sender rejected: %q", req.From)
			s.writeJSON(w, http.StatusForbidden, sendResponse{
				Success: false,
				Message: fmt.Sprintf("Sender %q is not allowed. Use one of the configured sender labels.", req.From),
			})
			return
		}
	}

	// Prevent batch sending.
	if strings.Contains(req.To, ",") {
		s.writeJSON(w, http.StatusBadRequest, sendResponse{
			Success: false,
			Message: "Only one recipient is allowed. Multiple recipients are not supported.",
		})
		return
	}

	// Render Markdown to HTML.
	html, err := markdown.RenderHTML(req.Markdown)
	if err != nil {
		log.Printf("ERROR markdown render: %v", err)
		s.writeJSON(w, http.StatusInternalServerError, sendResponse{
			Success: false,
			Message: "Failed to render markdown to HTML.",
		})
		return
	}

	// Send the email via the provider.
	m := mail.Mail{
		From:        fromAddr,
		To:          req.To,
		Subject:     req.Subject,
		HTMLBody:    html,
		Attachments: attachments,
	}

	if err := s.provider.Send(m); err != nil {
		log.Printf("ERROR send failed: %v", err)
		s.writeJSON(w, http.StatusInternalServerError, sendResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to send email: %v", err),
		})
		return
	}

	if len(attachments) > 0 {
		log.Printf("INFO  sent: from=%s to=%s subject=%q attachments=%d", fromAddr, req.To, req.Subject, len(attachments))
	} else {
		log.Printf("INFO  sent: from=%s to=%s subject=%q", fromAddr, req.To, req.Subject)
	}

	s.writeJSON(w, http.StatusOK, sendResponse{
		Success: true,
		Message: "Email sent successfully!",
	})
}

// handleSenders returns the list of configured sender labels and addresses.
// Only the labels and addresses are exposed — never the API key.
func (s *Server) handleSenders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeJSON(w, http.StatusMethodNotAllowed, sendResponse{
			Success: false,
			Message: "Method not allowed. Use GET.",
		})
		return
	}

	// Return label→address map so the frontend can populate the dropdown.
	// The frontend sends back the label; the backend resolves it.
	type senderEntry struct {
		Label   string `json:"label"`
		Address string `json:"address"`
	}

	senders := make([]senderEntry, 0, len(s.cfg.Senders))
	for label, addr := range s.cfg.Senders {
		senders = append(senders, senderEntry{Label: label, Address: addr})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(senders)
}

// writeJSON writes a JSON response with the given status code.
func (s *Server) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

