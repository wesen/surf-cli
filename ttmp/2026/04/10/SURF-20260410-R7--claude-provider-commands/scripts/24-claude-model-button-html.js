const button=document.querySelector('[data-testid="model-selector-dropdown"]');
if(!button) return {ok:false,error:'no model button',href:location.href};
return {
 ok:true,
 href:location.href,
 outerHTML: button.outerHTML,
 parentHTML: button.parentElement?.outerHTML?.slice(0,3000),
};
