const textMatches = Array.from(document.querySelectorAll('*'))
  .map((node) => {
    const text = (node.textContent || '').trim().replace(/\s+/g, ' ');
    if (!text || !/Thought for|Thinking|Reasoned for/i.test(text)) {
      return null;
    }
    return {
      tag: node.tagName,
      testid: node.getAttribute('data-testid'),
      aria: node.getAttribute('aria-label'),
      role: node.getAttribute('role'),
      className: (node.className || '').toString().slice(0, 200),
      text: text.slice(0, 300),
    };
  })
  .filter(Boolean)
  .slice(0, 100);

const likelyButtons = Array.from(document.querySelectorAll('button,[role="button"],[role="menuitem"]'))
  .map((node) => ({
    text: (node.textContent || '').trim().replace(/\s+/g, ' ').slice(0, 200),
    aria: node.getAttribute('aria-label'),
    testid: node.getAttribute('data-testid'),
    expanded: node.getAttribute('aria-expanded'),
    controls: node.getAttribute('aria-controls'),
  }))
  .filter((item) => /Thought for|Thinking|Reasoned|Show thinking|Hide thinking/i.test((item.text || '') + ' ' + (item.aria || '')));

return {
  href: location.href,
  title: document.title,
  textMatches,
  likelyButtons,
};
