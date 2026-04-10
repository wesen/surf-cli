const sleep = ms => new Promise(r => setTimeout(r, ms));
const query = 'from:github newer_than:30d';
const input = document.querySelector('input[name="q"], input[aria-label*="Search mail"]');
if (!input) {
  return { ok: false, error: 'search input not found', href: location.href, title: document.title };
}
input.focus();
input.value = query;
input.dispatchEvent(new InputEvent('input', { bubbles: true, data: query, inputType: 'insertText' }));
input.dispatchEvent(new Event('change', { bubbles: true }));
const form = input.closest('form');
if (form) {
  form.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));
}
input.dispatchEvent(new KeyboardEvent('keydown', { bubbles: true, cancelable: true, key: 'Enter', code: 'Enter' }));
input.dispatchEvent(new KeyboardEvent('keypress', { bubbles: true, cancelable: true, key: 'Enter', code: 'Enter' }));
input.dispatchEvent(new KeyboardEvent('keyup', { bubbles: true, cancelable: true, key: 'Enter', code: 'Enter' }));
const button = document.querySelector('button[aria-label="Search mail"]');
button?.click();
await sleep(3000);
const rows = Array.from(document.querySelectorAll('tr.zA')).slice(0, 5);
return {
  ok: true,
  href: location.href,
  title: document.title,
  query,
  inputValue: input.value,
  rowCount: Array.from(document.querySelectorAll('tr.zA')).length,
  sample: rows.map((row, i) => ({ i, className: row.className, text: (row.innerText || '').trim().slice(0, 300) })),
};
