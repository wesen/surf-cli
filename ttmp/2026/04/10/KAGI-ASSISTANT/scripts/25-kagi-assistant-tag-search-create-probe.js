const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

function normalizeText(value) {
  return (value || '').replace(/\s+/g, ' ').trim();
}

function realClick(el) {
  if (!el) return;
  el.focus();
  for (const type of ['pointerdown', 'mousedown', 'pointerup', 'mouseup', 'click']) {
    el.dispatchEvent(new MouseEvent(type, { bubbles: true, cancelable: true, view: window }));
  }
}

const button = document.querySelector('#tags-add');
if (!button) {
  return { ok: false, error: 'tags-add button not found', href: location.href };
}
realClick(button);
await sleep(500);
const root = document.querySelector('#tags dialog.promptOptionsSelector');
const input = root?.querySelector('input[type="search"][data-selector="input"]');
const createButton = root?.querySelector('button.create-new[data-selector="createNew"]');
if (!root || !input || !createButton) {
  return { ok: false, error: 'tag search controls not found', href: location.href };
}
input.focus();
input.value = 'codex-probe-tag';
input.dispatchEvent(new Event('input', { bubbles: true }));
input.dispatchEvent(new Event('change', { bubbles: true }));
await sleep(500);
const checkboxes = Array.from(root.querySelectorAll('input[type="checkbox"][name="tag"]')).map((el) => ({
  value: el.value,
  checked: !!el.checked,
  label: normalizeText(el.closest('label')?.textContent || ''),
}));
return {
  ok: true,
  href: location.href,
  inputValue: input.value,
  createButton: {
    text: normalizeText(createButton.textContent || ''),
    disabled: !!createButton.disabled,
  },
  checkboxes,
};
