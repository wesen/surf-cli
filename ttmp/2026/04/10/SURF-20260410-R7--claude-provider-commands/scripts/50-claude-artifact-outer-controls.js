function normalize(v){return String(v||'').replace(/\s+/g,' ').trim();}
const rows = Array.from(document.querySelectorAll('button, a, [role="button"]')).map((el,i)=>({
  i,
  tag: el.tagName.toLowerCase(),
  text: normalize(el.textContent).slice(0,140),
  aria: normalize(el.getAttribute('aria-label')),
  title: normalize(el.getAttribute('title')),
  testid: el.getAttribute('data-testid') || '',
  href: el.tagName.toLowerCase()==='a' ? el.href : '',
  className: (el.className || '').toString().slice(0,200),
  rect: (() => { const r = el.getBoundingClientRect(); return {x:r.x,y:r.y,w:r.width,h:r.height}; })(),
})).filter(x => x.rect.w > 0 && x.rect.h > 0).filter(x => /artifact|preview|html|download|copy|save|widget|open|fullscreen|expand|more/i.test([x.text,x.aria,x.title,x.testid,x.href,x.className].join(' ')));
return {href:location.href,title:document.title,rows};
