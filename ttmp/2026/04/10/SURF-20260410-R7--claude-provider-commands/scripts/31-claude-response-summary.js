function normalize(v){return String(v||'').replace(/\s+/g,' ').trim();}
const assistants = Array.from(document.querySelectorAll('div.font-claude-response')).filter(el => !el.parentElement?.closest('div.font-claude-response'));
const latest = assistants[assistants.length-1];
if (!latest) return {ok:false,error:'no assistant node',href:location.href,title:document.title};
return {
  ok:true,
  href:location.href,
  title:document.title,
  assistantCount:assistants.length,
  latestText: normalize(latest.innerText||latest.textContent),
  latestLength: normalize(latest.innerText||latest.textContent).length,
  sendVisible: !!Array.from(document.querySelectorAll('button')).find(el => normalize(el.getAttribute('aria-label')) === 'Send message'),
};
