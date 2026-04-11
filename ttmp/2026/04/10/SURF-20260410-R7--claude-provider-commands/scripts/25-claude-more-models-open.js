function sleep(ms){return new Promise(r=>setTimeout(r,ms));}
function text(v){return String(v||'').replace(/\s+/g,' ').trim();}
const button=document.querySelector('[data-testid="model-selector-dropdown"]');
if(!button) return {ok:false,error:'no model button',href:location.href};
button.click();
await sleep(200);
const items=Array.from(document.querySelectorAll('[role="menuitem"]'));
const more=items.find(el=>text(el.innerText||el.textContent)==='More models');
if(!more) return {ok:false,error:'no more models item',topLevel:items.map(el=>text(el.innerText||el.textContent))};
more.click();
await sleep(500);
const all=Array.from(document.querySelectorAll('[role="menuitem"], [role="menu"]')).map((el,i)=>({i,role:el.getAttribute('role'),text:text(el.innerText||el.textContent),ariaExpanded:el.getAttribute('aria-expanded'),ariaHaspopup:el.getAttribute('aria-haspopup'),className:(el.className||'').toString().slice(0,180)}));
return {ok:true,href:location.href,topLevel:items.map(el=>text(el.innerText||el.textContent)),all};
