const sections = Array.from(document.querySelectorAll('section[data-testid^="conversation-turn-"]'));
const summary = sections.map((section, sectionIndex) => {
  const messageNode = section.querySelector('[data-message-author-role="assistant"]');
  const role = messageNode?.getAttribute('data-message-author-role') || null;
  const messageId = messageNode?.getAttribute('data-message-id') || null;
  const model = messageNode?.getAttribute('data-message-model-slug') || null;
  const thoughtButtons = Array.from(section.querySelectorAll('button'))
    .map((button, buttonIndex) => ({
      buttonIndex,
      text: (button.textContent || '').trim().replace(/\s+/g, ' '),
      aria: button.getAttribute('aria-label'),
      expanded: button.getAttribute('aria-expanded'),
      controls: button.getAttribute('aria-controls'),
      className: (button.className || '').toString().slice(0, 240),
    }))
    .filter((item) => /Thought for|Thinking|Reasoned/i.test((item.text || '') + ' ' + (item.aria || '')));
  return {
    sectionIndex,
    sectionTestId: section.getAttribute('data-testid'),
    role,
    messageId,
    model,
    thoughtButtons,
  };
}).filter((item) => item.thoughtButtons.length > 0);

return {
  href: location.href,
  title: document.title,
  sections: summary,
};
