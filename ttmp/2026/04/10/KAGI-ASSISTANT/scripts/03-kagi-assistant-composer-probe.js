const textarea = document.querySelector('textarea[placeholder="Ask Assistant"]');
if (!textarea) {
  return { ok: false, error: 'textarea not found', href: location.href, title: document.title };
}
let root = textarea;
for (let i = 0; i < 6 && root.parentElement; i += 1) {
  root = root.parentElement;
}
const siblings = Array.from(root.parentElement?.children || []).map((el, i) => ({
  i,
  tag: el.tagName.toLowerCase(),
  text: (el.textContent || '').trim().replace(/\s+/g, ' ').slice(0, 200),
  className: (el.className || '').toString().slice(0, 240),
  html: el.outerHTML.slice(0, 800),
}));
return {
  ok: true,
  href: location.href,
  title: document.title,
  composerHtml: root.outerHTML.slice(0, 4000),
  siblings,
};
