return {
  href: location.href,
  items: Array.from(document.querySelectorAll('[data-testid]')).map((el, i) => ({
    i,
    tag: el.tagName.toLowerCase(),
    dataTestid: el.getAttribute('data-testid'),
    text: (el.textContent || '').trim().replace(/\s+/g, ' ').slice(0, 250),
    className: (el.className || '').toString().slice(0, 160),
  })).filter((x) => /message|chat|response|assistant|user|title|menu|input|share/i.test(x.dataTestid || '')).slice(0, 200),
};
