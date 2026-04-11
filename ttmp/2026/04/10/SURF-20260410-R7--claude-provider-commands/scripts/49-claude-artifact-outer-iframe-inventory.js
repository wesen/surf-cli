function normalize(v){return String(v||'').replace(/\s+/g,' ').trim();}
const iframes = Array.from(document.querySelectorAll('iframe')).map((el,i)=>({
  i,
  src: el.src || '',
  title: el.title || '',
  sandbox: el.getAttribute('sandbox') || '',
  allow: el.getAttribute('allow') || '',
  className: (el.className || '').toString().slice(0,200),
  width: el.style.width || '',
  height: el.style.height || '',
}));
const artifactish = Array.from(document.querySelectorAll('button, a, [role="button"]')).map((el,i)=>({
  i,
  tag: el.tagName.toLowerCase(),
  text: normalize(el.textContent).slice(0,140),
  aria: normalize(el.getAttribute('aria-label')),
  title: normalize(el.getAttribute('title')),
  testid: el.getAttribute('data-testid') || '',
  href: el.tagName.toLowerCase()==='a' ? el.href : '',
  outer: el.outerHTML.slice(0,300),
})).filter(x => /artifact|preview|html|download|copy|save|widget|open|fullscreen|expand/i.test([x.text,x.aria,x.title,x.testid,x.href,x.outer].join(' ')));
return {href:location.href,title:document.title,iframes,artifactish};
