const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

function normalizeText(value) {
  return (value || '').replace(/\s+/g, ' ').trim();
}

async function waitFor(fn, timeoutMs = 4000, intervalMs = 100) {
  const started = Date.now();
  while (Date.now() - started < timeoutMs) {
    const value = fn();
    if (value) return { value, waitedMs: Date.now() - started };
    await sleep(intervalMs);
  }
  return null;
}

const button = document.querySelector('#lens-select');
if (!button) {
  return { ok: false, error: 'Lens button not found', href: location.href };
}
button.click();
const ready = await waitFor(() => Array.from(document.querySelectorAll('dialog[open], [role="dialog"], [role="listbox"]')).find((el) => /lens/i.test(normalizeText(el.textContent || ''))), 4000, 100);
const popup = ready?.value || null;
return {
  ok: !!popup,
  href: location.href,
  button: {
    text: normalizeText(button.textContent || ''),
    title: button.getAttribute('title'),
    ariaExpanded: button.getAttribute('aria-expanded'),
  },
  popupText: popup ? normalizeText(popup.textContent || '').slice(0, 3000) : null,
  options: popup ? Array.from(popup.querySelectorAll('button, li, [role="option"], label')).map((el) => normalizeText(el.textContent || '')).filter(Boolean).slice(0, 50) : [],
};
