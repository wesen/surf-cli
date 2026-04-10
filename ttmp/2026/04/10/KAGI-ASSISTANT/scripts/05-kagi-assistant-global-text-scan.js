const bodyTextMatches = Array.from(document.querySelectorAll('body *')).map((el) => ({
  tag: el.tagName.toLowerCase(),
  text: (el.textContent || '').trim().replace(/\s+/g, ' '),
  className: (el.className || '').toString().slice(0, 180),
  title: el.getAttribute('title'),
  aria: el.getAttribute('aria-label'),
})).filter((x) => x.text && /model|assistant|claude|gpt|gemini|kimi|sonar|o3|4o|haiku|opus|flash|grok|selection/i.test(x.text)).slice(0,80);
return { matches: bodyTextMatches };
