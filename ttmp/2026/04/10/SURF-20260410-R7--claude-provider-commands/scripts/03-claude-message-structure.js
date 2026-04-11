const candidates = Array.from(document.querySelectorAll('[data-testid], [data-message-id], article, main section, [role="article"]')).map((el, i) => ({
  i,
  tag: el.tagName.toLowerCase(),
  dataTestid: el.getAttribute('data-testid'),
  messageId: el.getAttribute('data-message-id'),
  role: el.getAttribute('role'),
  className: (el.className || '').toString().slice(0, 180),
  text: (el.textContent || '').trim().replace(/\s+/g, ' ').slice(0, 300),
})).filter((x) => /message|assistant|user|claude|chat|artifact|thinking|response/i.test([x.dataTestid, x.className, x.text].join(' ')));
return {
  href: location.href,
  count: candidates.length,
  candidates: candidates.slice(0, 120),
};
