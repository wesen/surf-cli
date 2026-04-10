const textarea = document.querySelector('textarea[placeholder="Ask Assistant"]');
const controls = Array.from(document.querySelectorAll('button,[role="button"],[role="combobox"],select')).map((el, i) => ({
  i,
  tag: el.tagName.toLowerCase(),
  role: el.getAttribute('role'),
  text: (el.textContent || '').trim().replace(/\s+/g, ' ').slice(0, 160),
  aria: el.getAttribute('aria-label'),
  title: el.getAttribute('title'),
  id: el.id || null,
  className: (el.className || '').toString().slice(0, 200),
  dataState: el.getAttribute('data-state'),
  ariaExpanded: el.getAttribute('aria-expanded'),
})).filter((x) => x.text || x.aria || x.title);
const nearby = textarea ? Array.from(textarea.closest('form,div')?.querySelectorAll('*') || []).slice(0,80).map((el, i) => ({
  i,
  tag: el.tagName.toLowerCase(),
  role: el.getAttribute('role'),
  text: (el.textContent || '').trim().replace(/\s+/g, ' ').slice(0, 120),
  aria: el.getAttribute('aria-label'),
  title: el.getAttribute('title'),
  className: (el.className || '').toString().slice(0, 180),
})).filter((x) => x.text || x.aria || x.title) : [];
return { href: location.href, controls: controls.slice(0,120), nearby };
