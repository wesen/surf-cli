return {
  href: location.href,
  title: document.title,
  ready: document.readyState,
  bodyPreview: (document.body?.innerText || '').slice(0, 1200),
  landmarkSummary: Array.from(document.querySelectorAll('main, nav, aside, header, footer, form')).map((el, i) => ({
    i,
    tag: el.tagName.toLowerCase(),
    role: el.getAttribute('role'),
    ariaLabel: el.getAttribute('aria-label'),
    className: (el.className || '').toString().slice(0, 160),
    text: (el.textContent || '').trim().slice(0, 200),
  })),
};
