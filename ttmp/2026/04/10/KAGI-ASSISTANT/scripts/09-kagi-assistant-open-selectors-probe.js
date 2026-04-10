const candidateButtons = Array.from(document.querySelectorAll('button')).map((el, i) => ({
  i,
  text: (el.textContent || '').trim().replace(/\s+/g, ' '),
  title: el.getAttribute('title'),
  ariaExpanded: el.getAttribute('aria-expanded'),
  className: (el.className || '').toString().slice(0, 200),
})).filter((x) => /assistant|model|chatgpt|claude|gemini|grok|kimi|research|web access/i.test((x.text || '') + ' ' + (x.title || '')));
return { candidateButtons };
