function sleep(ms){return new Promise(r=>setTimeout(r,ms));}
function normalize(v){return String(v||'').replace(/\s+/g,' ').trim();}
const assistantNodes = Array.from(document.querySelectorAll('div.font-claude-response')).filter(el => !el.parentElement?.closest('div.font-claude-response'));
const latest = assistantNodes[assistantNodes.length - 1];
if (!latest) return {ok:false,error:'no assistant node',href:location.href,title:document.title};
const button = Array.from(latest.querySelectorAll('button')).find(el => /searched the web/i.test(normalize(el.innerText || el.textContent) + ' ' + normalize(el.getAttribute('aria-label'))));
if (!button) return {ok:false,error:'no searched the web button',href:location.href,title:document.title};
const beforeExpanded = button.getAttribute('aria-expanded');
button.click();
await sleep(600);
const afterExpanded = button.getAttribute('aria-expanded');
const nearby = [];
let cur = button.parentElement;
for (let depth = 0; cur && depth < 5; depth += 1, cur = cur.parentElement) {
  nearby.push({
    depth,
    tag: cur.tagName.toLowerCase(),
    className: (cur.className || '').toString().slice(0, 200),
    text: normalize(cur.innerText || cur.textContent).slice(0, 1500),
  });
}
const details = Array.from(document.querySelectorAll('a[href], button, [role="button"], div, section')).map((el, i) => ({
  i,
  tag: el.tagName.toLowerCase(),
  role: el.getAttribute('role'),
  text: normalize(el.innerText || el.textContent).slice(0, 300),
  href: el.tagName.toLowerCase() === 'a' ? el.href : null,
  className: (el.className || '').toString().slice(0, 200),
})).filter(x => /searched the web|search query|web search|result|source|visit|mit press|oapen|ocw/i.test([x.text, x.className, x.role].join(' '))).slice(0, 100);
return {ok:true,href:location.href,title:document.title,beforeExpanded,afterExpanded,nearby,details};
