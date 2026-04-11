function normalize(v){return String(v||'').replace(/\s+/g,' ').trim();}
const artifactish = Array.from(document.querySelectorAll('iframe, webview, [data-testid], button, a, div')).map((el,i)=>({
  i,
  tag: el.tagName.toLowerCase(),
  dataTestid: el.getAttribute('data-testid'),
  ariaLabel: normalize(el.getAttribute('aria-label')),
  text: normalize(el.innerText||el.textContent).slice(0,300),
  href: el.tagName.toLowerCase()==='a' ? el.href : null,
  src: el.tagName.toLowerCase()==='iframe' ? el.src : null,
  className: (el.className||'').toString().slice(0,200),
})).filter(x => /artifact|preview|html|sandbox|open preview|open artifact|download|code|render/i.test([x.dataTestid,x.ariaLabel,x.text,x.className,x.href,x.src].join(' '))).slice(0,120);
const assistants = Array.from(document.querySelectorAll('div.font-claude-response')).filter(el => !el.parentElement?.closest('div.font-claude-response'));
return {href:location.href,title:document.title,assistantTexts: assistants.map((el,i)=>({i,text:normalize(el.innerText||el.textContent).slice(0,1000)})), artifactish};
