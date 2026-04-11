function normalize(v){return String(v||'').replace(/\s+/g,' ').trim();}
const buttons = Array.from(document.querySelectorAll('button, [role="button"]')).map((el,i)=>({
  i,
  tag:el.tagName.toLowerCase(),
  text:normalize(el.innerText||el.textContent),
  ariaLabel:normalize(el.getAttribute('aria-label')),
  dataTestid:el.getAttribute('data-testid'),
  className:(el.className||'').toString().slice(0,200),
  visible: !!el.getClientRects().length,
})).filter(x => x.visible).filter(x => /send|stop|copy|retry|edit|continue|web|prompt|chat|assistant|voice|regenerate|branch|share/i.test([x.text,x.ariaLabel,x.dataTestid,x.className].join(' ')));
return {href:location.href,title:document.title,buttons};
