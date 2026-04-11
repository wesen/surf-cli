function normalize(v){return String(v||'').replace(/\s+/g,' ').trim();}
const candidates = Array.from(document.querySelectorAll('button')).map((el,i)=>({
  i,
  el,
  aria: normalize(el.getAttribute('aria-label')),
  testid: el.getAttribute('data-testid') || '',
  rect: el.getBoundingClientRect(),
})).filter(x => /More options for Art of insight landing page/i.test(x.aria));
if (!candidates.length) return {ok:false,error:'no matching button',href:location.href,title:document.title};
const target = candidates.sort((a,b)=> b.rect.y - a.rect.y)[0];
target.el.click();
await new Promise(r => setTimeout(r, 1000));
const menus = Array.from(document.querySelectorAll('[role="menu"], [data-radix-popper-content-wrapper], [cmdk-root], [data-side], [role="dialog"]')).map((el,i)=>({
  i,
  tag: el.tagName.toLowerCase(),
  role: el.getAttribute('role') || '',
  text: normalize(el.innerText || el.textContent).slice(0,1200),
})).filter(x => x.text);
const items = Array.from(document.querySelectorAll('[role="menuitem"], [role="menuitemradio"], [role="option"], button, a')).map((el,i)=>({
  i,
  tag: el.tagName.toLowerCase(),
  text: normalize(el.innerText || el.textContent).slice(0,200),
  aria: normalize(el.getAttribute('aria-label')),
  role: el.getAttribute('role') || '',
  href: el.tagName.toLowerCase()==='a' ? el.href : '',
})).filter(x => /download|copy|clipboard|artifact|save|html|open|share|duplicate|delete|rename|export/i.test([x.text,x.aria,x.role,x.href].join(' ')));
return {
  ok:true,
  href:location.href,
  title:document.title,
  clickedAria: target.aria,
  clickedY: target.rect.y,
  menus,
  items,
};
