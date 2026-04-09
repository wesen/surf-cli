const buttons = Array.from(document.querySelectorAll('button'));
const target = buttons.find((node) => /Thought for/i.test((node.textContent || '').trim()));
if (!target) {
  return { ok: false, error: 'No thought button found', href: location.href };
}
const before = document.body.innerText;
target.click();
await new Promise((resolve) => setTimeout(resolve, 1200));
const dialog = document.querySelector('[role="dialog"], [data-state="open"], [aria-modal="true"]');
const openPanels = Array.from(document.querySelectorAll('*'))
  .map((node) => ({
    tag: node.tagName,
    role: node.getAttribute('role'),
    aria: node.getAttribute('aria-label'),
    text: (node.textContent || '').trim().replace(/\s+/g, ' ').slice(0, 500),
  }))
  .filter((item) => /Thought for|Thinking|Reasoned|analysis|scratchpad|chain|step/i.test(item.text + ' ' + (item.aria || '')))
  .slice(0, 50);
return {
  ok: true,
  clickedText: (target.textContent || '').trim(),
  activeElement: document.activeElement ? {
    tag: document.activeElement.tagName,
    role: document.activeElement.getAttribute('role'),
    aria: document.activeElement.getAttribute('aria-label'),
    text: (document.activeElement.textContent || '').trim().slice(0, 300),
  } : null,
  dialog: dialog ? {
    tag: dialog.tagName,
    text: (dialog.textContent || '').trim().replace(/\s+/g, ' ').slice(0, 1000),
    aria: dialog.getAttribute('aria-label'),
  } : null,
  openPanels,
  bodyChanged: before !== document.body.innerText,
};
