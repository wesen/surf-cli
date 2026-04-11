function normalize(v){return String(v||'').replace(/\s+/g,' ').trim();}
const controls = Array.from(document.querySelectorAll('button, a, [role="button"]')).map((el,i)=>({
  i,
  tag: el.tagName.toLowerCase(),
  text: normalize(el.innerText||el.textContent),
  ariaLabel: normalize(el.getAttribute('aria-label')),
  title: normalize(el.getAttribute('title')),
  dataTestid: el.getAttribute('data-testid'),
  href: el.tagName.toLowerCase()==='a' ? el.href : null,
  className: (el.className||'').toString().slice(0,220),
  visible: !!el.getClientRects().length,
})).filter(x => x.visible).filter(x => /download|copy|clipboard|artifact|save|open|preview|html|code|share/i.test([x.text,x.ariaLabel,x.title,x.dataTestid,x.className,x.href].join(' ')));
return {href:location.href,title:document.title,count:controls.length,controls};
