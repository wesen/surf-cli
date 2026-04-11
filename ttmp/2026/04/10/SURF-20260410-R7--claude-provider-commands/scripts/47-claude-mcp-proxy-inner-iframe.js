function normalize(v){return String(v||'').replace(/\s+/g,' ').trim();}
const frame = document.querySelector('iframe');
const out = {
  href: location.href,
  title: document.title,
  hasFrame: !!frame,
  frameSrc: frame?.getAttribute('src') || '',
  frameTitle: frame?.getAttribute('title') || '',
};
if (frame) {
  try {
    const doc = frame.contentDocument || frame.contentWindow?.document;
    out.sameOrigin = !!doc;
    out.innerHref = frame.contentWindow?.location?.href || '';
    out.innerReadyState = doc?.readyState || '';
    out.innerTitle = doc?.title || '';
    out.innerHasMoreBtn = !!doc?.querySelector('#more-btn');
    out.innerMoreBtnHtml = doc?.querySelector('#more-btn')?.outerHTML || '';
    out.innerBodyChildren = Array.from(doc?.body?.children || []).map((el, i) => ({
      i,
      tag: el.tagName.toLowerCase(),
      id: el.id || '',
      className: (el.className || '').toString().slice(0,160),
      text: normalize(el.innerText || el.textContent).slice(0,120),
    })).slice(0, 20);
    out.innerButtons = Array.from(doc?.querySelectorAll('button') || []).map((el, i) => ({
      i,
      id: el.id || '',
      aria: normalize(el.getAttribute('aria-label')),
      text: normalize(el.textContent).slice(0,120),
    })).slice(0, 20);
  } catch (e) {
    out.sameOrigin = false;
    out.error = String(e && e.stack || e);
  }
}
return out;
