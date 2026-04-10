const buttons = Array.from(document.querySelectorAll('button'));
const target = buttons.find((el) => /Edit custom assistant|GLM|Claude|GPT|Gemini|Grok|Kagi Research|ChatGPT/i.test((el.textContent || '').trim()));
if (!target) return { ok: false, count: buttons.length };
return {
  ok: true,
  text: (target.textContent || '').trim().replace(/\s+/g, ' '),
  title: target.getAttribute('title'),
  ariaExpanded: target.getAttribute('aria-expanded'),
  className: (target.className || '').toString(),
  html: target.outerHTML,
  parentHtml: target.parentElement ? target.parentElement.outerHTML.slice(0, 4000) : null,
};
