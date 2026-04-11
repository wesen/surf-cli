function normalize(v){return String(v||'').replace(/\s+/g,' ').trim();}
const button = document.querySelector('body > div:nth-child(3) > button');
if (!button) return {ok:false,error:'floating button not found',href:location.href,title:document.title};
return {
  ok:true,
  href:location.href,
  title:document.title,
  outerHTML: button.outerHTML,
  text: normalize(button.innerText||button.textContent),
  ariaLabel: normalize(button.getAttribute('aria-label')),
  titleAttr: normalize(button.getAttribute('title')),
  className: (button.className||'').toString(),
};
