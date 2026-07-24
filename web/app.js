// Maild — compose form logic.

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
const attachmentInput = document.getElementById('attachment-input');
const attachmentList = document.getElementById('attachment-list');

// Track selected files so we can rebuild the FileList after modification.
let selectedFiles = [];

// Load sender identities — display email addresses only, no labels.
async function loadSenders() {
  try {
    const resp = await fetch('/api/senders');
    if (!resp.ok) throw new Error('Failed to load senders');
    const senders = await resp.json();

    fromSelect.innerHTML = '';
    for (const s of senders) {
      const opt = document.createElement('option');
      opt.value = s.address;
      opt.textContent = s.address;
      fromSelect.appendChild(opt);
    }
  } catch (err) {
    fromSelect.innerHTML = '<option value="">加载失败</option>';
    showStatus('error', '无法加载发件人列表，请确认服务是否运行。');
  }
}

// --- Attachments ---

attachmentInput.addEventListener('change', () => {
  // Merge newly selected files into our tracked array.
  for (const f of attachmentInput.files) {
    // Avoid duplicate filenames.
    if (!selectedFiles.some(s => s.name === f.name && s.size === f.size)) {
      selectedFiles.push(f);
    }
  }
  attachmentInput.value = '';
  renderAttachmentList();
});

function renderAttachmentList() {
  attachmentList.innerHTML = '';
  selectedFiles.forEach((f, i) => {
    const li = document.createElement('li');
    li.textContent = f.name;

    const rm = document.createElement('button');
    rm.textContent = '×';
    rm.title = '移除';
    rm.addEventListener('click', () => {
      selectedFiles.splice(i, 1);
      renderAttachmentList();
    });

    li.appendChild(rm);
    attachmentList.appendChild(li);
  });

  attachmentList.style.display = selectedFiles.length ? 'flex' : 'none';
}

// Preview toggle.
previewBtn.addEventListener('click', () => {
  const md = markdownTextarea.value.trim();
  previewContent.innerHTML = md ? simpleMarkdownToHTML(md) : '<p style="color: var(--ink-faint)">暂无内容。</p>';
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

  const from = fromSelect.value.trim();
  const to = toInput.value.trim();
  const subject = subjectInput.value.trim();
  const markdown = markdownTextarea.value.trim();

  if (!from || !to || !subject || !markdown) {
    showStatus('error', '请填写所有字段。');
    return;
  }

  if (!to.includes('@')) {
    showStatus('error', '请填写有效的收件人邮箱。');
    return;
  }

  sendBtn.disabled = true;
  sendBtn.textContent = '发送中…';
  hideStatus();

  try {
    let resp;
    if (selectedFiles.length > 0) {
      // Multipart: send fields + attachments.
      const fd = new FormData();
      fd.append('from', from);
      fd.append('to', to);
      fd.append('subject', subject);
      fd.append('markdown', markdown);
      for (const f of selectedFiles) {
        fd.append('attachments', f);
      }
      resp = await fetch('/api/send', { method: 'POST', body: fd });
    } else {
      // JSON: no attachments (backward compatible).
      resp = await fetch('/api/send', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ from, to, subject, markdown }),
      });
    }

    const data = await resp.json();

    if (data.success) {
      showStatus('success', '已发送');
      toInput.value = '';
      subjectInput.value = '';
      markdownTextarea.value = '';
      selectedFiles = [];
      renderAttachmentList();
    } else {
      showStatus('error', data.message);
    }
  } catch (err) {
    showStatus('error', '网络异常，无法连接服务。');
  } finally {
    sendBtn.disabled = false;
    sendBtn.textContent = '发送';
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

// Simple client-side Markdown renderer for preview.
function simpleMarkdownToHTML(md) {
  let html = md;

  // Escape HTML.
  html = html.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');

  // Headers.
  html = html.replace(/^#### (.+)$/gm, '<h4>$1</h4>');
  html = html.replace(/^### (.+)$/gm, '<h3>$1</h3>');
  html = html.replace(/^## (.+)$/gm, '<h2>$1</h2>');
  html = html.replace(/^# (.+)$/gm, '<h1>$1</h1>');

  // Bold / italic.
  html = html.replace(/\*\*\*(.+?)\*\*\*/g, '<strong><em>$1</em></strong>');
  html = html.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');
  html = html.replace(/\*(.+?)\*/g, '<em>$1</em>');

  // Inline code.
  html = html.replace(/`([^`]+)`/g, '<code>$1</code>');

  // Links.
  html = html.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2">$1</a>');

  // Blockquotes.
  html = html.replace(/^&gt; (.+)$/gm, '<blockquote>$1</blockquote>');

  // Lists.
  html = html.replace(/((?:^- .+\n?)+)/gm, (match) => {
    const items = match.trim().split('\n').map(line => `<li>${line.replace(/^- /, '')}</li>`).join('');
    return `<ul>${items}</ul>`;
  });
  html = html.replace(/((?:^\d+\. .+\n?)+)/gm, (match) => {
    const items = match.trim().split('\n').map(line => `<li>${line.replace(/^\d+\. /, '')}</li>`).join('');
    return `<ol>${items}</ol>`;
  });

  // Paragraphs.
  html = html.replace(/\n\n/g, '</p><p>');
  html = '<p>' + html + '</p>';
  html = html.replace(/<p>\s*<\/p>/g, '');
  html = html.replace(/\n/g, '<br>');

  return html;
}

// Auto-focus the editor on load, so the user can start writing immediately.
loadSenders();
markdownTextarea.focus();
