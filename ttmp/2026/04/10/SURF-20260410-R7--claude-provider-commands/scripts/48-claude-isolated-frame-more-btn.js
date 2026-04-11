const normalize = (v) => String(v || '').replace(/\s+/g, ' ').trim();
const b = document.querySelector('#more-btn');
return {
  href: location.href,
  title: document.title,
  hasMoreBtn: !!b,
  moreBtnAria: b?.getAttribute('aria-label') || '',
  moreBtnHtml: b?.outerHTML || '',
  buttons: Array.from(document.querySelectorAll('button')).slice(0, 50).map((el, i) => ({
    i,
    id: el.id || '',
    aria: normalize(el.getAttribute('aria-label')),
    text: normalize(el.textContent).slice(0, 120),
  })),
  bodyChildren: Array.from(document.body.children).slice(0, 20).map((el, i) => ({
    i,
    tag: el.tagName.toLowerCase(),
    id: el.id || '',
    className: (el.className || '').toString().slice(0, 200),
    text: normalize(el.innerText || el.textContent).slice(0, 160),
  })),
};
