return {
  href: location.href,
  title: document.title,
  sections: Array.from(document.querySelectorAll('main section, article, [data-testid], [data-message-id], [role="article"]')).map((el, i) => ({
    i,
    tag: el.tagName.toLowerCase(),
    role: el.getAttribute('role'),
    dataTestid: el.getAttribute('data-testid'),
    messageId: el.getAttribute('data-message-id'),
    ariaLabel: el.getAttribute('aria-label'),
    className: (el.className || '').toString().slice(0, 180),
    text: (el.textContent || '').trim().replace(/\s+/g, ' ').slice(0, 500),
  })).filter((x) => x.text && !/^New chatSearch/.test(x.text)).slice(0, 200),
};
