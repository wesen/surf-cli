const blocks = Array.from(document.querySelectorAll('main div, main p, main article, main section, main span')).map((el, i) => ({
  i,
  tag: el.tagName.toLowerCase(),
  role: el.getAttribute('role'),
  dataTestid: el.getAttribute('data-testid'),
  className: (el.className || '').toString().slice(0, 180),
  text: (el.textContent || '').trim().replace(/\s+/g, ' '),
})).filter((x) => x.text.length >= 8).slice(0, 500);
return {
  href: location.href,
  blocks: blocks.filter((x) => /hello|clarifying|double-check|please just say hello|simple greeting request/i.test(x.text)).slice(0, 120),
};
