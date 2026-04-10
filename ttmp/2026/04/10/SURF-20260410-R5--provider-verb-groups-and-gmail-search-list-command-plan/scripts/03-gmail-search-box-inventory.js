const inputs = Array.from(document.querySelectorAll('input, textarea')).filter(el => {
  const label = [el.getAttribute('aria-label'), el.getAttribute('placeholder'), el.name, el.id].filter(Boolean).join(' | ');
  return /search/i.test(label);
});
return {
  href: location.href,
  title: document.title,
  searchInputs: inputs.map((el, i) => ({
    i,
    tag: el.tagName,
    type: el.getAttribute('type'),
    name: el.getAttribute('name'),
    id: el.id || null,
    ariaLabel: el.getAttribute('aria-label'),
    placeholder: el.getAttribute('placeholder'),
    formRole: el.closest('form')?.getAttribute('role') || null,
  })),
  buttons: Array.from(document.querySelectorAll('button, [role="button"]')).filter(el => /search/i.test((el.getAttribute('aria-label') || el.textContent || '').trim())).slice(0, 10).map((el, i) => ({
    i,
    tag: el.tagName,
    ariaLabel: el.getAttribute('aria-label'),
    text: (el.textContent || '').trim(),
  })),
};
