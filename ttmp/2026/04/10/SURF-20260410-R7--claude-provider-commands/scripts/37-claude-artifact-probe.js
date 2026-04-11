function normalize(v){return String(v||'').replace(/\s+/g,' ').trim();}
const markers = Array.from(document.querySelectorAll('iframe, [data-testid], button, [role="button"], a')).map((el,i)=>({
  i,
  tag:el.tagName.toLowerCase(),
  dataTestid:el.getAttribute('data-testid'),
  ariaLabel:normalize(el.getAttribute('aria-label')),
  text:normalize(el.innerText||el.textContent).slice(0,200),
  href:el.tagName.toLowerCase()==='a'?el.href:null,
  className:(el.className||'').toString().slice(0,200),
})).filter(x => /artifact|preview|html|code|sandbox|open|download|share|copy|retry|edit|render/i.test([x.dataTestid,x.ariaLabel,x.text,x.className].join(' ')));
const assistants = Array.from(document.querySelectorAll('div.font-claude-response')).filter(el => !el.parentElement?.closest('div.font-claude-response'));
return {
  href: location.href,
  title: document.title,
  assistantCount: assistants.length,
  latestText: normalize((assistants[assistants.length-1]?.innerText||assistants[assistants.length-1]?.textContent||'')).slice(0,2000),
  markers,
};
