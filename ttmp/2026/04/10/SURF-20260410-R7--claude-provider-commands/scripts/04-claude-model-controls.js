return {
  href: location.href,
  controls: Array.from(document.querySelectorAll('button, [role="button"], [role="combobox"], [aria-haspopup], [data-testid]')).map((el, i) => ({
    i,
    tag: el.tagName.toLowerCase(),
    role: el.getAttribute('role'),
    ariaLabel: el.getAttribute('aria-label'),
    ariaHaspopup: el.getAttribute('aria-haspopup'),
    dataTestid: el.getAttribute('data-testid'),
    className: (el.className || '').toString().slice(0, 180),
    text: (el.textContent || '').trim().replace(/\s+/g, ' ').slice(0, 200),
  })).filter((x) => /model|sonnet|opus|haiku|claude|chat style|customize|project/i.test([x.ariaLabel, x.dataTestid, x.className, x.text].join(' '))).slice(0, 200),
};
