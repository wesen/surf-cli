function normalize(v){return String(v||'').replace(/\s+/g,' ').trim();}
return {
  href: location.href,
  title: document.title,
  bodyChildren: Array.from(document.body.children).map((el,i)=>({
    i,
    tag: el.tagName.toLowerCase(),
    id: el.id,
    className: (el.className||'').toString().slice(0,240),
    dataTestid: el.getAttribute('data-testid'),
    text: normalize(el.innerText||el.textContent).slice(0,300),
    childCount: el.children.length,
  })),
};
