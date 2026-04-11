function normalize(v){return String(v||'').replace(/\s+/g,' ').trim();}
const assistantNodes = Array.from(document.querySelectorAll('div.font-claude-response')).filter(el => !el.parentElement?.closest('div.font-claude-response'));
const latest = assistantNodes[assistantNodes.length-1];
if (!latest) return {ok:false,error:'no assistant node',href:location.href,title:document.title};
const links = Array.from(latest.querySelectorAll('a[href]')).map((a,i)=>({
  i,
  href:a.href,
  text:normalize(a.innerText||a.textContent),
  rel:a.getAttribute('rel'),
  target:a.getAttribute('target'),
  dataTestid:a.getAttribute('data-testid'),
  className:(a.className||'').toString().slice(0,200),
  parentText:normalize(a.parentElement?.innerText||a.parentElement?.textContent).slice(0,300),
}));
const buttons = Array.from(latest.querySelectorAll('button')).map((b,i)=>({
  i,
  text:normalize(b.innerText||b.textContent),
  ariaLabel:b.getAttribute('aria-label'),
  dataTestid:b.getAttribute('data-testid'),
  className:(b.className||'').toString().slice(0,200),
})).filter(x=>/cite|source|link|footnote|reference|web/i.test([x.text,x.ariaLabel,x.dataTestid,x.className].join(' ')));
return {
  ok:true,
  href:location.href,
  title:document.title,
  assistantText: normalize(latest.innerText||latest.textContent).slice(0,1500),
  linkCount: links.length,
  links,
  buttons,
};
