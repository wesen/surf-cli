const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

const getActivitySection = () => {
  const selectors = [
    'body > div:nth-child(5) > div > div.relative.z-0.flex.min-h-0.w-full.flex-1 > div.bg-token-bg-elevated-secondary.relative.z-1.shrink-0.overflow-x-hidden.max-lg\\:w-0\\!.stage-thread-flyout-preset-default > div > div > section',
    '[class*="stage-thread-flyout"] section',
  ];
  for (const selector of selectors) {
    const nodes = Array.from(document.querySelectorAll(selector));
    for (const node of nodes) {
      const text = (node.innerText || '').trim();
      if (/Activity\s*·|Activity\s*:|Thinking\n|Sources\s*·/i.test(text)) {
        return node;
      }
    }
  }
  return null;
};

const clickLikeUser = async (button) => {
  button.scrollIntoView({ block: 'center', inline: 'center' });
  await sleep(150);
  for (const type of ['pointerdown', 'mousedown', 'pointerup', 'mouseup', 'click']) {
    button.dispatchEvent(new MouseEvent(type, { bubbles: true, cancelable: true, view: window }));
  }
};

const sections = Array.from(document.querySelectorAll('section[data-testid^="conversation-turn-"]'));
const thoughtTargets = [];
for (const section of sections) {
  const assistant = section.querySelector('[data-message-author-role="assistant"]');
  if (!assistant) continue;
  const buttons = Array.from(section.querySelectorAll('button')).filter((button) => /Thought for/i.test((button.textContent || '').trim()));
  if (buttons.length === 0) continue;
  thoughtTargets.push({
    sectionTestId: section.getAttribute('data-testid'),
    messageId: assistant.getAttribute('data-message-id'),
    model: assistant.getAttribute('data-message-model-slug'),
    buttons,
  });
}

const results = [];
for (const target of thoughtTargets) {
  for (let i = 0; i < Math.min(target.buttons.length, 2); i += 1) {
    const button = target.buttons[i];
    await clickLikeUser(button);
    await sleep(1200);
    const section = getActivitySection();
    const text = section ? (section.innerText || '').trim() : '';
    results.push({
      messageId: target.messageId,
      model: target.model,
      sectionTestId: target.sectionTestId,
      buttonIndex: i,
      buttonText: (button.textContent || '').trim(),
      foundActivity: !!section,
      activityPreview: text.slice(0, 1200),
      activityLength: text.length,
    });
  }
}

return {
  href: location.href,
  title: document.title,
  resultCount: results.length,
  results,
};
