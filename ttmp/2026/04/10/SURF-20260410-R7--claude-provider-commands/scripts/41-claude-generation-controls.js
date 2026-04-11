function normalize(v){return String(v||'').replace(/\s+/g,' ').trim();}
const buttons = Array.from(document.querySelectorAll('button, [role="button"]')).map((el,i)=>({
  i,
  text: normalize(el.innerText||el.textContent),
  ariaLabel: normalize(el.getAttribute('aria-label')),
  dataTestid: el.getAttribute('data-testid'),
  className: (el.className||'').toString().slice(0,220),
  visible: !!el.getClientRects().length,
})).filter(x=>x.visible);
return {
  href: location.href,
  title: document.title,
  controls: buttons.filter(x=>/stop|send|retry|edit|copy|vote|thumb|regenerate|voice|search/i.test([x.text,x.ariaLabel,x.dataTestid,x.className].join(' '))),
  allVisibleCount: buttons.length,
};
