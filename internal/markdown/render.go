// Package markdown renders Markdown content to HTML suitable for email.
package markdown

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/yuin/goldmark"
)

// emailTemplate wraps rendered Markdown HTML in an email-friendly document.
const emailTemplate = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>
body {
	font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
	line-height: 1.6;
	color: #1a1a1a;
	max-width: 640px;
	margin: 0 auto;
	padding: 16px;
}
pre {
	background: #f5f5f5;
	padding: 12px;
	border-radius: 4px;
	overflow-x: auto;
}
code {
	background: #f5f5f5;
	padding: 2px 6px;
	border-radius: 3px;
	font-size: 0.9em;
}
blockquote {
	border-left: 3px solid #ccc;
	margin-left: 0;
	padding-left: 16px;
	color: #666;
}
a { color: #2563eb; }
table { border-collapse: collapse; width: 100%%; }
th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
img { max-width: 100%%; height: auto; }
</style>
</head>
<body>
%s
</body>
</html>`

// markdownRenderer is the goldmark instance used for rendering.
var markdownRenderer = goldmark.New()

// RenderHTML converts a Markdown string to a full HTML email document.
func RenderHTML(md string) (string, error) {
	var buf bytes.Buffer
	if err := markdownRenderer.Convert([]byte(md), &buf); err != nil {
		return "", fmt.Errorf("rendering markdown: %w", err)
	}

	html := fmt.Sprintf(emailTemplate, buf.String())

	// Validate that the template produces valid output.
	// This is a best-effort check; goldmark already produces safe output.
	if _, err := template.New("check").Parse(html); err != nil {
		return "", fmt.Errorf("invalid email template output: %w", err)
	}

	return html, nil
}
