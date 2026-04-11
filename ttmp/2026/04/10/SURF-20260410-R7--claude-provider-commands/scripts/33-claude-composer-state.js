function normalize(v){return String(v||'').replace(/\s+/g,' ').trim();}
const input = document.querySelector('[data-testid="chat-input"]');
const formish = input?.closest('form, div');
const sendCandidates = Array.from(document.querySelectorAll('button')).map((el,i)=>({
  i,
  ariaLabel: normalize(el.getAttribute('aria-label')),
  text: normalize(el.innerText||el.textContent),
  disabled: el.disabled,
  visible: !!el.getClientRects().length,
  className:(el.className||'').toString().slice(0,200),
})).filter(x=>x.visible && /send|stop|arrow|voice|microphone|up|submit|continue/i.test([x.ariaLabel,x.text,x.className].join(' ')));
return {
  href:location.href,
  title:document.title,
  inputExists: !!input,
  inputText: normalize(input?.innerText||input?.textContent),
  formText: normalize(formish?.innerText||formish?.textContent).slice(0,800),
  sendCandidates,
};
