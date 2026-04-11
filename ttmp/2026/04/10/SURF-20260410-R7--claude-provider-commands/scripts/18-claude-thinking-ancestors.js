function pathFor(el) {
  const out = [];
  let cur = el;
  for (let depth = 0; cur && depth < 10; depth += 1, cur = cur.parentElement) {
    out.push({
      depth,
      tag: cur.tagName.toLowerCase(),
      role: cur.getAttribute('role'),
      dataTestid: cur.getAttribute('data-testid'),
      ariaLabel: cur.getAttribute('aria-label'),
      className: (cur.className || '').toString().slice(0, 180),
      text: (cur.textContent || '').trim().replace(/\s+/g, ' ').slice(0, 300),
    });
  }
  return out;
}
const needles = [/Thinking about clarifying/i, /double-check responses/i, /Hello!/i, /Please just say hello\./i];
const hits = Array.from(document.querySelectorAll('main *')).filter((el) => {
  const text = (el.textContent || '').trim();
  return text && needles.some((re) => re.test(text));
}).slice(0, 50);
return {
  href: location.href,
  hitCount: hits.length,
  hits: hits.map((el, i) => ({
    i,
    self: {
      tag: el.tagName.toLowerCase(),
      role: el.getAttribute('role'),
      dataTestid: el.getAttribute('data-testid'),
      ariaLabel: el.getAttribute('aria-label'),
      className: (el.className || '').toString().slice(0, 180),
      text: (el.textContent || '').trim().replace(/\s+/g, ' ').slice(0, 300),
    },
    ancestors: pathFor(el),
  })),
};
