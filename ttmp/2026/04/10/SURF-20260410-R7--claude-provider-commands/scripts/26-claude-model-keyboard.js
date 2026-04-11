function sleep(ms){return new Promise(r=>setTimeout(r,ms));}
function text(v){return String(v||'').replace(/\s+/g,' ').trim();}
const button=document.querySelector('[data-testid="model-selector-dropdown"]');
if(!button) return {ok:false,error:'no model button'};
button.click();
await sleep(150);
const more=Array.from(document.querySelectorAll('[role="menuitem"]')).find(el=>text(el.innerText||el.textContent)==='More models');
if(!more) return {ok:false,error:'no more item',items:Array.from(document.querySelectorAll('[role="menuitem"]')).map(el=>text(el.innerText||el.textContent))};
more.focus();
more.dispatchEvent(new KeyboardEvent('keydown',{key:'ArrowRight',code:'ArrowRight',bubbles:true}));
await sleep(500);
const all=Array.from(document.querySelectorAll('[role="menuitem"], [role="menu"]')).map((el,i)=>({i,role:el.getAttribute('role'),text:text(el.innerText||el.textContent),className:(el.className||'').toString().slice(0,180)}));
return {ok:true,all};
