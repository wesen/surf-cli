const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));
const btn = document.querySelector('#profile-select');
if (!btn) return { ok: false, error: 'profile-select not found' };
btn.click();
await sleep(800);
const listboxes = Array.from(document.querySelectorAll('[role="listbox"], [role="option"], .listbox, .dropdown, .popover, .floating-ui-portal, .radix-select-content, [data-radix-popper-content-wrapper]')).map((el, i) => ({
  i,
  tag: el.tagName.toLowerCase(),
  role: el.getAttribute('role'),
  id: el.id || null,
  className: (el.className || '').toString().slice(0, 240),
  text: (el.textContent || '').trim().replace(/\s+/g, ' ').slice(0, 2000),
  html: el.outerHTML.slice(0, 3000),
})).filter((x) => x.text);
const options = Array.from(document.querySelectorAll('[role="option"], li, button, a, label, div')).map((el, i) => ({
  i,
  tag: el.tagName.toLowerCase(),
  role: el.getAttribute('role'),
  text: (el.textContent || '').trim().replace(/\s+/g, ' '),
  className: (el.className || '').toString().slice(0, 240),
  ariaSelected: el.getAttribute('aria-selected'),
  dataSelected: el.getAttribute('data-selected'),
})).filter((x) => /Claude|GPT|Gemini|Grok|GLM|Kagi Research|ChatGPT|assistant|research|mini|sonnet|opus|haiku|reasoning|custom/i.test(x.text)).slice(0, 200);
return {
  ok: true,
  ariaExpanded: btn.getAttribute('aria-expanded'),
  listboxes,
  options,
};
