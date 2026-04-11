const summarize = (el, i) => ({
  i,
  tag: el.tagName.toLowerCase(),
  type: el.getAttribute('type'),
  role: el.getAttribute('role'),
  ariaLabel: el.getAttribute('aria-label'),
  placeholder: el.getAttribute('placeholder'),
  contenteditable: el.getAttribute('contenteditable'),
  dataTestid: el.getAttribute('data-testid'),
  className: (el.className || '').toString().slice(0, 160),
  text: (el.textContent || '').trim().slice(0, 200),
});
return {
  href: location.href,
  inputs: Array.from(document.querySelectorAll('input, textarea, [contenteditable="true"], button')).map(summarize).filter((x) => {
    const text = [x.ariaLabel, x.placeholder, x.text, x.dataTestid, x.className].join(' ');
    return /chat|message|send|model|claude|assistant|attach|upload|artifact|project|search/i.test(text);
  }).slice(0, 200),
};
