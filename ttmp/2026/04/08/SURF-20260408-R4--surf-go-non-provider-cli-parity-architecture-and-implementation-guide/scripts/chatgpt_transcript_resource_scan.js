const resources = performance.getEntriesByType('resource')
  .map((entry) => entry.name)
  .filter((name) => /conversation|backend-api|share|export|download/i.test(name));
const nav = {
  href: location.href,
  pathname: location.pathname,
  conversationId: location.pathname.split('/').filter(Boolean).pop() || null,
};
const buttons = Array.from(document.querySelectorAll('button, a, [role="menuitem"]'))
  .map((el) => ({
    text: (el.textContent || '').trim().replace(/\s+/g, ' ').slice(0, 120),
    href: el.getAttribute('href'),
    testid: el.getAttribute('data-testid'),
    aria: el.getAttribute('aria-label'),
  }))
  .filter((item) => /share|export|download|copy|transcript/i.test((item.text || '') + ' ' + (item.aria || '')));
return { nav, resources, buttons };
