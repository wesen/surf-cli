function normalize(v){return String(v||'').replace(/\s+/g,' ').trim();}
const node = Array.from(document.querySelectorAll('div.font-claude-response')).filter(el => !el.parentElement?.closest('div.font-claude-response')).slice(-1)[0];
if (!node) return {ok:false,error:'no assistant node'};
const ancestors = [];
let cur = node;
for (let depth=0; cur && depth<8; depth += 1, cur = cur.parentElement) {
  const buttons = Array.from(cur.querySelectorAll('button')).map(el => ({
    text: normalize(el.innerText||el.textContent),
    aria: normalize(el.getAttribute('aria-label')),
    dataTestid: el.getAttribute('data-testid'),
  })).filter(x => /copy|retry|edit|search/i.test([x.text,x.aria,x.dataTestid].join(' ')));
  ancestors.push({
    depth,
    tag: cur.tagName.toLowerCase(),
    className:(cur.className||'').toString().slice(0,200),
    buttonCount: buttons.length,
    buttons,
  });
}
return {href:location.href,title:document.title,ancestors};
