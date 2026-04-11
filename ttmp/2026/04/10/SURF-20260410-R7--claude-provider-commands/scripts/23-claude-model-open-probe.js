function sleep(ms){return new Promise(r=>setTimeout(r,ms));}
function text(v){return String(v||'').replace(/\s+/g,' ').trim();}
const button=document.querySelector('[data-testid="model-selector-dropdown"]');
if(!button) return {ok:false,error:'no model button',href:location.href};
const before={
  text:text(button.innerText||button.textContent),
  ariaExpanded:button.getAttribute('aria-expanded'),
  ariaHaspopup:button.getAttribute('aria-haspopup'),
  dataState:button.getAttribute('data-state'),
};
button.focus();
button.click();
await sleep(500);
const menuitems=Array.from(document.querySelectorAll('[role="menuitem"], [role="option"], [role="menu"], [role="listbox"]')).map((el,i)=>({i,tag:el.tagName.toLowerCase(),role:el.getAttribute('role'),text:text(el.innerText||el.textContent),className:(el.className||'').toString().slice(0,180),dataTestid:el.getAttribute('data-testid')})).filter(x=>x.text||x.role==='menu'||x.role==='listbox');
return {
  ok:true,
  href:location.href,
  before,
  after:{
    ariaExpanded:button.getAttribute('aria-expanded'),
    ariaHaspopup:button.getAttribute('aria-haspopup'),
    dataState:button.getAttribute('data-state'),
  },
  activeTag:document.activeElement?.tagName,
  activeText:text(document.activeElement?.innerText||document.activeElement?.textContent),
  menuitems,
};
