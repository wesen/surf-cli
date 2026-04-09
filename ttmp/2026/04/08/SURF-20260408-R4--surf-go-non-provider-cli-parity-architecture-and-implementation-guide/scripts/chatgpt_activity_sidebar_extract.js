const selectors = [
  'body > div:nth-child(5) > div > div.relative.z-0.flex.min-h-0.w-full.flex-1 > div.bg-token-bg-elevated-secondary.relative.z-1.shrink-0.overflow-x-hidden.max-lg\\:w-0\\!.stage-thread-flyout-preset-default > div > div > section',
  'section:has(h2), section:has(h1)',
  '[class*="stage-thread-flyout"] section',
];

const matches = [];
for (const selector of selectors) {
  try {
    const nodes = Array.from(document.querySelectorAll(selector));
    for (const node of nodes) {
      const text = (node.innerText || '').trim();
      if (!text) continue;
      matches.push({
        selector,
        tag: node.tagName,
        textLength: text.length,
        preview: text.slice(0, 2000),
      });
    }
  } catch (error) {
    matches.push({ selector, error: String(error) });
  }
}

const activityLike = Array.from(document.querySelectorAll('section,aside,div'))
  .map((node) => ({
    tag: node.tagName,
    className: (node.className || '').toString().slice(0, 300),
    text: (node.innerText || '').trim(),
  }))
  .filter((item) => /Activity\s*·|Activity\s*:|Thinking\n|Sources\s*·|Thought for \d|Determining when to browse/i.test(item.text))
  .map((item) => ({
    tag: item.tag,
    className: item.className,
    textLength: item.text.length,
    preview: item.text.slice(0, 2500),
  }))
  .slice(0, 20);

return {
  href: location.href,
  title: document.title,
  matches,
  activityLike,
};
