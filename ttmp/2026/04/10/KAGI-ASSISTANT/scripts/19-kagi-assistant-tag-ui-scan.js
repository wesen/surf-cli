function normalizeText(value) {
  return (value || '').replace(/\s+/g, ' ').trim();
}

const matches = Array.from(document.querySelectorAll('button, a, label, div, span, li')).map((el, i) => ({
  i,
  tag: el.tagName.toLowerCase(),
  id: el.id || null,
  role: el.getAttribute('role'),
  text: normalizeText(el.textContent || '').slice(0, 240),
  title: el.getAttribute('title'),
  aria: el.getAttribute('aria-label'),
  cls: (el.className || '').toString().slice(0, 200),
})).filter((x) => /tag|untagged|edit tags|rename|customize|delete/i.test((x.text || '') + ' ' + (x.title || '') + ' ' + (x.aria || '')));

return {
  href: location.href,
  title: document.title,
  matches: matches.slice(0, 200),
};
