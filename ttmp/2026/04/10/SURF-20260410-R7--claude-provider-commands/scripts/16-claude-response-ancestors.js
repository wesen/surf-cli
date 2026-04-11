function pathFor(el) {
  const out = [];
  let cur = el;
  for (let depth = 0; cur && depth < 8; depth += 1, cur = cur.parentElement) {
    out.push({
      depth,
      tag: cur.tagName.toLowerCase(),
      role: cur.getAttribute('role'),
      dataTestid: cur.getAttribute('data-testid'),
      messageId: cur.getAttribute('data-message-id'),
      ariaLabel: cur.getAttribute('aria-label'),
      className: (cur.className || '').toString().slice(0, 180),
      text: (cur.textContent || '').trim().replace(/\s+/g, ' ').slice(0, 300),
    });
  }
  return out;
}
const all = Array.from(document.querySelectorAll('main *'));
const hits = all.filter((el) => /hello/i.test((el.textContent || '').trim())).slice(0, 20);
return {
  href: location.href,
  hitCount: hits.length,
  hits: hits.map((el, i) => ({
    i,
    self: {
      tag: el.tagName.toLowerCase(),
      role: el.getAttribute('role'),
      dataTestid: el.getAttribute('data-testid'),
      className: (el.className || '').toString().slice(0, 180),
      text: (el.textContent || '').trim().replace(/\s+/g, ' ').slice(0, 300),
    },
    ancestors: pathFor(el),
  })),
};
