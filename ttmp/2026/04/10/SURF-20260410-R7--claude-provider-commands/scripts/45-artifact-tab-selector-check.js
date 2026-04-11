function normalize(v){return String(v||'').replace(/\s+/g,' ').trim();}
const button = document.querySelector('body > div:nth-child(3) > button');
return {
  href: location.href,
  title: document.title,
  readyState: document.readyState,
  hasButton: !!button,
  buttonText: normalize(button?.innerText||button?.textContent),
  buttonAria: normalize(button?.getAttribute('aria-label')),
  bodyChildren: Array.from(document.body.children).map((el,i)=>({i,tag:el.tagName.toLowerCase(),id:el.id,className:(el.className||'').toString().slice(0,160),text:normalize(el.innerText||el.textContent).slice(0,120)})),
};
