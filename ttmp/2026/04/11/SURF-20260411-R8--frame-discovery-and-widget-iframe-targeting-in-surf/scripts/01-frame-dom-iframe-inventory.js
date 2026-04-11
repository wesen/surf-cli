function normalize(v){return String(v || '').replace(/\s+/g, ' ').trim();}
const iframes = Array.from(document.querySelectorAll('iframe')).map((el, domIndex) => ({
  domIndex,
  src: el.src || '',
  title: el.title || '',
  name: el.getAttribute('name') || '',
  id: el.id || '',
  sandbox: el.getAttribute('sandbox') || '',
  allow: el.getAttribute('allow') || '',
  className: (el.className || '').toString().slice(0, 200),
  rect: (() => {
    const r = el.getBoundingClientRect();
    return { x: r.x, y: r.y, width: r.width, height: r.height };
  })(),
}));
return {
  href: location.href,
  title: document.title,
  iframeCount: iframes.length,
  iframes,
  notes: [
    'This script runs in the main page context only.',
    'It inventories visible iframe elements but does not attempt cross-origin DOM access.'
  ],
};
