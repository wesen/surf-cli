function normalize(v){return String(v||'').replace(/\s+/g,' ').trim();}
const iframe = Array.from(document.querySelectorAll('iframe')).find(el => /isolated-segment/.test(el.src || ''));
if (!iframe) return {ok:false,error:'no artifact iframe',href:location.href,title:document.title};
const ancestors=[];
let cur=iframe;
for (let depth=0; cur && depth<8; depth+=1, cur=cur.parentElement) {
  const buttons = Array.from(cur.querySelectorAll('button, a, [role="button"]')).map((el,i)=>({
    i,
    tag:el.tagName.toLowerCase(),
    text:normalize(el.innerText||el.textContent),
    ariaLabel:normalize(el.getAttribute('aria-label')),
    title:normalize(el.getAttribute('title')),
    dataTestid:el.getAttribute('data-testid'),
    href: el.tagName.toLowerCase()==='a' ? el.href : null,
  })).filter(x=>/download|copy|clipboard|artifact|save|open|preview|html|code|share/i.test([x.text,x.ariaLabel,x.title,x.dataTestid,x.href].join(' ')));
  ancestors.push({depth, tag:cur.tagName.toLowerCase(), className:(cur.className||'').toString().slice(0,220), buttons});
}
return {ok:true,href:location.href,title:document.title,iframeSrc:iframe.src,ancestors};
