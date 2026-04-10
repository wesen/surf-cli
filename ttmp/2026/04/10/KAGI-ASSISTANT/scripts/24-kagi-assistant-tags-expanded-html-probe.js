const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

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
const tags = document.querySelector('#tags');
return {
  ok: !!tags,
  href: location.href,
  tagsHtml: tags ? tags.outerHTML.slice(0, 6000) : null,
};
