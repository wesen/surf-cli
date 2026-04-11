function normalize(v){return String(v||'').replace(/\s+/g,' ').trim();}
const assistantNodes = Array.from(document.querySelectorAll('div.font-claude-response')).filter(el => !el.parentElement?.closest('div.font-claude-response'));
const latest = assistantNodes[assistantNodes.length-1];
if (!latest) return {ok:false,error:'no assistant node',href:location.href,title:document.title};
const markers = Array.from(latest.querySelectorAll('*')).map((el,i)=>({
  i,
  tag:el.tagName.toLowerCase(),
  role:el.getAttribute('role'),
  dataTestid:el.getAttribute('data-testid'),
  ariaLabel:el.getAttribute('aria-label'),
  className:(el.className||'').toString().slice(0,200),
  text:normalize(el.innerText||el.textContent).slice(0,200),
})).filter(x=>/cite|source|footnote|reference|copy|retry|stop|thinking|edit|web|artifact|branch/i.test([x.role,x.dataTestid,x.ariaLabel,x.className,x.text].join(' ')));
return {ok:true,href:location.href,title:document.title,nodeCount:assistantNodes.length,markers};
