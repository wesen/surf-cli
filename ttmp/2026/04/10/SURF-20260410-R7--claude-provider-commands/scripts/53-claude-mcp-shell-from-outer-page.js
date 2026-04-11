function normalize(v){return String(v||'').replace(/\s+/g,' ').trim();}
const frame = document.querySelectorAll('iframe')[0];
if (!frame) return {ok:false,error:'no first iframe',href:location.href,title:document.title};
const out = {
  ok: true,
  href: location.href,
  title: document.title,
  frameSrc: frame.src || '',
  frameTitle: frame.title || '',
};
try {
  const doc = frame.contentDocument || frame.contentWindow?.document;
  out.sameOrigin = !!doc;
  out.innerHref = frame.contentWindow?.location?.href || '';
  out.innerTitle = doc?.title || '';
  out.innerReadyState = doc?.readyState || '';
  out.innerBodyChildren = Array.from(doc?.body?.children || []).slice(0,20).map((el,i)=>({
    i,
    tag: el.tagName.toLowerCase(),
    id: el.id || '',
    className: (el.className || '').toString().slice(0,200),
    text: normalize(el.innerText || el.textContent).slice(0,180),
  }));
  const innerFrame = doc?.querySelector('iframe');
  out.hasInnerFrame = !!innerFrame;
  out.innerFrameSrc = innerFrame?.getAttribute('src') || '';
  out.innerFrameTitle = innerFrame?.getAttribute('title') || '';
  out.hasMoreBtn = !!doc?.querySelector('#more-btn');
  out.moreBtnHtml = doc?.querySelector('#more-btn')?.outerHTML || '';
  if (innerFrame) {
    try {
      const innerDoc = innerFrame.contentDocument || innerFrame.contentWindow?.document;
      out.inner2SameOrigin = !!innerDoc;
      out.inner2Href = innerFrame.contentWindow?.location?.href || '';
      out.inner2Title = innerDoc?.title || '';
      out.inner2HasMoreBtn = !!innerDoc?.querySelector('#more-btn');
      out.inner2MoreBtnHtml = innerDoc?.querySelector('#more-btn')?.outerHTML || '';
      out.inner2Buttons = Array.from(innerDoc?.querySelectorAll('button') || []).slice(0,30).map((el,i)=>({
        i,
        id: el.id || '',
        aria: normalize(el.getAttribute('aria-label')),
        text: normalize(el.textContent).slice(0,160),
      }));
      out.inner2BodyChildren = Array.from(innerDoc?.body?.children || []).slice(0,20).map((el,i)=>({
        i,
        tag: el.tagName.toLowerCase(),
        id: el.id || '',
        className: (el.className || '').toString().slice(0,200),
        text: normalize(el.innerText || el.textContent).slice(0,180),
      }));
    } catch (e) {
      out.inner2Error = String(e && e.stack || e);
    }
  }
} catch (e) {
  out.error = String(e && e.stack || e);
}
return out;
