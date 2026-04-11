function sleep(ms){return new Promise(r=>setTimeout(r,ms));}
function text(v){return String(v||'').replace(/\s+/g,' ').trim();}
const button=document.querySelector('[data-testid="model-selector-dropdown"]');
if(!button) return {ok:false,error:'no model button'};
const before=text(button.innerText||button.textContent);
button.click();
await sleep(200);
const ext=Array.from(document.querySelectorAll('[role="menuitem"]')).find(el=>text(el.innerText||el.textContent).startsWith('Extended thinking'));
if(!ext) return {ok:false,error:'no extended toggle',before};
ext.click();
await sleep(700);
return {ok:true,before,after:text(button.innerText||button.textContent)};
