// Maild — compose form logic. No framework, vanilla JS only.

const form = document.getElementById('compose-form');
const fromSelect = document.getElementById('from');
const toInput = document.getElementById('to');
const subjectInput = document.getElementById('subject');
const markdownTextarea = document.getElementById('markdown');
const sendBtn = document.getElementById('send-btn');
const previewBtn = document.getElementById('preview-btn');
const previewPanel = document.getElementById('preview-panel');
const previewContent = document.getElementById('preview-content');
const closePreviewBtn = document.getElementById('close-preview');
const statusEl = document.getElementById('status');

// Load sender identities from the server.
async function loadSenders() {
  try {
    const resp = await fetch('/api/senders');
    if (!resp.ok) throw new Error('Failed to load senders');
    const senders = await resp.json();

    fromSelect.innerHTML = '<option value="">Choose sender…</option>';
    for (const s of senders) {
      const opt = document.createElement('option');
      opt.value = s.label;
      opt.textContent = `${s.label} — ${s.address}`;
      fromSelect.appendChild(opt);
    }
  } catch (err) {
    fromSelect.innerHTML = '<option value="">Failed to load senders</option>';
    showStatus('error', 'Could not load sender identities. Is the server running?');
  }
}

// Preview: render markdown client-side for a quick preview.
previewBtn.addEventListener('click', () => {
  const md = markdownTextarea.value.trim();
  if (!md) {
    previewContent.innerHTML = '<p style="color: var(--muted)">Nothing to preview.</p>';
  } else {
    // Simple client-side Markdown-to-HTML using a minimal converter.
    // For a faithful preview we call the server, but for speed we do a basic render.
    previewContent.innerHTML = simpleMarkdownToHTML(md);
  }
  form.classList.add('hidden');
  previewPanel.classList.remove('hidden');
});

closePreviewBtn.addEventListener('click', () => {
  previewPanel.classList.add('hidden');
  form.classList.remove('hidden');
});

// Form submission.
form.addEventListener('submit', async (e) => {
  e.preventDefault();

  const from = fromSelect.value;
  const to = toInput.value.trim();
  const subject = subjectInput.value.trim();
  const markdown = markdownTextarea.value.trim();

  if (!from || !to || !subject || !markdown) {
    showStatus('error', 'All fields are required.');
    return;
  }

  // Basic email validation.
  if (!to.includes('@')) {
    showStatus('error', 'Please enter a valid recipient email address.');
    return;
  }

  sendBtn.disabled = true;
  sendBtn.textContent = 'Sending…';
  hideStatus();

  try {
    const resp = await fetch('/api/send', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ from, to, subject, markdown }),
    });

    const data = await resp.json();

    if (data.success) {
      showStatus('success', data.message);
      // Clear the form on success for quick follow-up sends.
      // Keep the sender selection, clear the rest.
      toInput.value = '';
      subjectInput.value = '';
      markdownTextarea.value = '';
    } else {
      showStatus('error', data.message);
    }
  } catch (err) {
    showStatus('error', 'Network error. Could not reach the server.');
  } finally {
    sendBtn.disabled = false;
    sendBtn.textContent = 'Send';
  }
});

// --- Helpers ---

function showStatus(type, message) {
  statusEl.className = type;
  statusEl.textContent = message;
  statusEl.classList.remove('hidden');
}

function hideStatus() {
  statusEl.classList.add('hidden');
}

// Minimal client-side Markdown renderer for the preview panel.
// This is intentionally simple — the server does the real rendering for email.
function simpleMarkdownToHTML(md) {
  let html = md;

  // Escape HTML entities
  html = html.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');

  // Headers
  html = html.replace(/^#### (.+)$/gm, '<h4>$1</h4>');
  html = html.replace(/^### (.+)$/gm, '<h3>$1</h3>');
  html = html.replace(/^## (.+)$/gm, '<h2>$1</h2>');
  html = html.replace(/^# (.+)$/gm, '<h1>$1</h1>');

  // Bold and italic
  html = html.replace(/\*\*\*(.+?)\*\*\*/g, '<strong><em>$1</em></strong>');
  html = html.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');
  html = html.replace(/\*(.+?)\*/g, '<em>$1</em>');

  // Inline code
  html = html.replace(/`([^`]+)`/g, '<code>$1</code>');

  // Links
  html = html.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2">$1</a>');

  // Blockquotes
  html = html.replace(/^&gt; (.+)$/gm, '<blockquote>$1</blockquote>');

  // Unordered lists — group consecutive list items
  html = html.replace(/((?:^- .+\n?)+)/gm, (match) => {
    const items = match.trim().split('\n').map(line => `<li>${line.replace(/^- /, '')}</li>`).join('');
    return `<ul>${items}</ul>`;
  });

  // Ordered lists
  html = html.replace(/((?:^\d+\. .+\n?)+)/gm, (match) => {
    const items = match.trim().split('\n').map(line => `<li>${line.replace(/^\d+\. /, '')}</li>`).join('');
    return `<ol>${items}</ol>`;
  });

  // Paragraphs — wrap remaining text blocks
  html = html.replace(/\n\n/g, '</p><p>');
  html = '<p>' + html + '</p>';
  html = html.replace(/<p>\s*<\/p>/g, '');

  // Line breaks
  html = html.replace(/\n/g, '<br>');

  return html;
}

// Initialize on page load.
loadSenders();
