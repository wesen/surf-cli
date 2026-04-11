const controls = Array.from(document.querySelectorAll('button, [role="button"], input, textarea, [contenteditable="true"]')).map((el, i) => ({
  i,
  tag: el.tagName.toLowerCase(),
  type: el.getAttribute('type'),
  role: el.getAttribute('role'),
  ariaLabel: el.getAttribute('aria-label'),
  dataTestid: el.getAttribute('data-testid'),
  disabled: !!el.disabled || el.getAttribute('aria-disabled') === 'true',
  className: (el.className || '').toString().slice(0, 180),
  text: (el.textContent || '').trim().replace(/\s+/g, ' ').slice(0, 200),
})).filter((x) => /send|submit|stop|upload|attach|claude|model|extended|artifact/i.test([x.ariaLabel, x.dataTestid, x.className, x.text].join(' ')));
return { href: location.href, controls };
