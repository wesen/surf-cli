const normalize = (v) => String(v || '').replace(/\s+/g, ' ').trim();
return {
  href: location.href,
  title: document.title,
  readyState: document.readyState,
  hasMoreBtn: !!document.querySelector('#more-btn'),
  moreBtnHtml: document.querySelector('#more-btn')?.outerHTML || '',
  buttons: Array.from(document.querySelectorAll('button')).slice(0, 80).map((el, i) => ({
    i,
    id: el.id || '',
    aria: normalize(el.getAttribute('aria-label')),
    text: normalize(el.textContent).slice(0, 160),
    className: (el.className || '').toString().slice(0, 200),
  })),
  bodyChildren: Array.from(document.body.children).slice(0, 40).map((el, i) => ({
    i,
    tag: el.tagName.toLowerCase(),
    id: el.id || '',
    className: (el.className || '').toString().slice(0, 200),
    text: normalize(el.innerText || el.textContent).slice(0, 180),
  })),
};
