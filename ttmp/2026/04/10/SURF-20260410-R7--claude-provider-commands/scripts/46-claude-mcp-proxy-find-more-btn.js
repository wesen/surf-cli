function normalize(v){return String(v||'').replace(/\s+/g,' ').trim();}
const btn = document.querySelector('#more-btn');
return {
  href: location.href,
  title: document.title,
  readyState: document.readyState,
  hasMoreBtn: !!btn,
  moreBtnAria: normalize(btn?.getAttribute('aria-label')),
  moreBtnHtml: btn?.outerHTML || '',
  bodyChildren: Array.from(document.body.children).map((el,i)=>({i,tag:el.tagName.toLowerCase(),id:el.id,className:(el.className||'').toString().slice(0,160),text:normalize(el.innerText||el.textContent).slice(0,120)})).slice(0,10),
};
