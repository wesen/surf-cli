const nodes = Array.from(document.querySelectorAll('body *')).map((el, i) => ({
  i,
  tag: el.tagName.toLowerCase(),
  role: el.getAttribute('role'),
  dataTestid: el.getAttribute('data-testid'),
  ariaLabel: el.getAttribute('aria-label'),
  className: (el.className || '').toString().slice(0, 180),
  text: (el.textContent || '').trim().replace(/\s+/g, ' '),
})).filter((x) => x.text.length >= 8);
return {
  href: location.href,
  count: nodes.length,
  blocks: nodes.filter((x) => /hello|clarifying|double-check|please just say hello|simple greeting request/i.test(x.text)).slice(0, 150),
};
