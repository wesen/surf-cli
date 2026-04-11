function sleep(ms){return new Promise(r=>setTimeout(r,ms));}
function normalize(v){return String(v||'').replace(/\s+/g,' ').trim();}
const button = document.querySelector('body > div:nth-child(3) > button');
if (!button) return {ok:false,error:'floating button not found',href:location.href,title:document.title};
button.click();
await sleep(500);
const controls = Array.from(document.querySelectorAll('body > div:nth-child(3) button, body > div:nth-child(3) a, body > div:nth-child(3) [role="menuitem"], [role="menuitem"], [role="dialog"] button, [role="dialog"] a')).map((el,i)=>({
  i,
  tag:el.tagName.toLowerCase(),
  text: normalize(el.innerText||el.textContent),
  ariaLabel: normalize(el.getAttribute('aria-label')),
  titleAttr: normalize(el.getAttribute('title')),
  dataTestid: el.getAttribute('data-testid'),
  href: el.tagName.toLowerCase()==='a' ? el.href : null,
  className: (el.className||'').toString().slice(0,220),
  visible: !!el.getClientRects().length,
})).filter(x=>x.visible);
return {ok:true,href:location.href,title:document.title,controls};
