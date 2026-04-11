function sleep(ms){return new Promise(r=>setTimeout(r,ms));}
function text(v){return String(v||'').replace(/\s+/g,' ').trim();}
async function openMenu(){const b=document.querySelector('[data-testid="model-selector-dropdown"]'); if(!b) throw new Error('no model button'); b.click(); await sleep(200); return b;}
const button=await openMenu();
const more=Array.from(document.querySelectorAll('[role="menuitem"]')).find(el=>text(el.innerText||el.textContent)==='More models');
if(!more) return {ok:false,error:'no more models'};
more.click();
await sleep(200);
const opus=Array.from(document.querySelectorAll('[role="menuitem"]')).find(el=>text(el.innerText||el.textContent)==='Opus 4.6');
if(!opus) return {ok:false,error:'no opus 4.6'};
opus.click();
await sleep(700);
return {ok:true,current:text(button.innerText||button.textContent),href:location.href};
