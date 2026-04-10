const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));
const options = typeof SURF_OPTIONS === 'object' && SURF_OPTIONS !== null ? SURF_OPTIONS : {};
const prompt = options.prompt || 'hello';

function normalizeText(value) {
  return (value || '').replace(/\s+/g, ' ').trim();
}

async function waitFor(fn, timeoutMs = 30000, intervalMs = 200) {
  const started = Date.now();
  while (Date.now() - started < timeoutMs) {
    const value = fn();
    if (value) return { value, waitedMs: Date.now() - started };
    await sleep(intervalMs);
  }
  return null;
}

function messageNodes() {
  return Array.from(document.querySelectorAll('main [data-query-id], main article, main .message, main .assistant-response, main .chat-response, main [data-message-author-role]'));
}

function summarizeMessages() {
  return messageNodes().map((el, i) => ({
    i,
    tag: el.tagName.toLowerCase(),
    cls: (el.className || '').toString().slice(0, 200),
    text: normalizeText(el.textContent || '').slice(0, 500),
    attrs: {
      queryId: el.getAttribute('data-query-id'),
      role: el.getAttribute('data-message-author-role'),
    },
  })).filter((x) => x.text);
}

const textarea = document.querySelector('#promptBox, textarea[placeholder="Ask Assistant"]');
const form = document.querySelector('form#form');
const submit = document.querySelector('button#submit.submit[type="submit"], button[aria-label="Submit"]');
if (!textarea || !form || !submit) {
  return { ok: false, error: 'Prompt controls not found', href: location.href, controls: { textarea: !!textarea, form: !!form, submit: !!submit } };
}

textarea.focus();
textarea.value = prompt;
textarea.dispatchEvent(new Event('input', { bubbles: true }));
textarea.dispatchEvent(new Event('change', { bubbles: true }));
await sleep(100);
submit.click();

const waited = await waitFor(() => {
  const summaries = summarizeMessages();
  return summaries.find((m) => /hello|hey|hi|need|assist|help/i.test(m.text));
}, 30000, 300);

return {
  ok: !!waited,
  href: location.href,
  prompt,
  waitedMs: waited?.waitedMs || null,
  matched: waited?.value || null,
  messages: summarizeMessages().slice(-12),
};
