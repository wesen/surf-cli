const hits = Array.from(document.querySelectorAll('main *')).map((el, i) => ({
  i,
  tag: el.tagName.toLowerCase(),
  role: el.getAttribute('role'),
  dataTestid: el.getAttribute('data-testid'),
  messageId: el.getAttribute('data-message-id'),
  ariaLabel: el.getAttribute('aria-label'),
  className: (el.className || '').toString().slice(0, 180),
  text: (el.textContent || '').trim().replace(/\s+/g, ' ').slice(0, 300),
})).filter((x) => /hello/i.test(x.text));
return { href: location.href, count: hits.length, hits: hits.slice(0, 80) };
