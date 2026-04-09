const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

const ACTIVITY_SELECTORS = [
  'body > div:nth-child(5) > div > div.relative.z-0.flex.min-h-0.w-full.flex-1 > div.bg-token-bg-elevated-secondary.relative.z-1.shrink-0.overflow-x-hidden.max-lg\\:w-0\\!.stage-thread-flyout-preset-default > div > div > section',
  '[class*="stage-thread-flyout"] section',
];

const ACTIVITY_HEADER_RE = /Activity\s*[^\w\n]?\s*[^\n]+/i;
const SOURCES_HEADER_RE = /Sources\s*[^\w\n]?\s*\d+/i;

function getActivitySection() {
  for (const selector of ACTIVITY_SELECTORS) {
    for (const node of document.querySelectorAll(selector)) {
      const text = (node.innerText || '').trim();
      if (ACTIVITY_HEADER_RE.test(text) || text.includes('Thinking\n') || SOURCES_HEADER_RE.test(text)) {
        return { selector, node, text };
      }
    }
  }
  return null;
}

async function waitForCondition(predicate, timeoutMs, intervalMs = 250) {
  const started = Date.now();
  while (Date.now() - started < timeoutMs) {
    const value = predicate();
    if (value) {
      return { value, waitedMs: Date.now() - started };
    }
    await sleep(intervalMs);
  }
  return null;
}

function durationFromButtonText(text) {
  return (text || '').replace(/^Thought for\s*/i, '').trim();
}

function activityMatchesDuration(activityText, duration) {
  if (!activityText || !duration) return false;
  const escaped = duration.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  return new RegExp(`Activity\\s*[^\\w\\n]?\\s*${escaped}`, 'i').test(activityText);
}

async function normalizeAndOpen(button) {
  const buttonText = (button.textContent || '').trim().replace(/\s+/g, ' ');
  const duration = durationFromButtonText(buttonText);
  button.scrollIntoView({ block: 'center', inline: 'center' });
  await sleep(250);

  for (let attempt = 1; attempt <= 2; attempt += 1) {
    button.click();
    const hit = await waitForCondition(() => {
      const activity = getActivitySection();
      if (!activity) return null;
      return activityMatchesDuration(activity.text, duration) ? activity : null;
    }, 7000);
    if (hit) {
      return { attempt, duration, buttonText, ...hit.value, waitedMs: hit.waitedMs };
    }
    await sleep(400);
  }

  return { attempt: 2, duration, buttonText, selector: null, node: null, text: null, waitedMs: null };
}

const sections = Array.from(document.querySelectorAll('section[data-testid^="conversation-turn-"]'));
const targetSection = sections.find((section) =>
  Array.from(section.querySelectorAll('button')).some((node) => /Thought for/i.test((node.textContent || '').trim()))
);
if (!targetSection) {
  return { ok: false, error: 'No conversation turn with thought button found', href: location.href };
}
const assistantNode = targetSection.querySelector('[data-message-author-role="assistant"]');
const button = Array.from(targetSection.querySelectorAll('button')).find((node) => /Thought for/i.test((node.textContent || '').trim()));
if (!button) {
  return { ok: false, error: 'No thought button found in target section', href: location.href };
}

const activity = await normalizeAndOpen(button);
return {
  ok: true,
  href: location.href,
  sectionTestId: targetSection.getAttribute('data-testid'),
  messageId: assistantNode ? assistantNode.getAttribute('data-message-id') : null,
  model: assistantNode ? assistantNode.getAttribute('data-message-model-slug') : null,
  buttonText: activity.buttonText,
  duration: activity.duration,
  foundActivity: !!activity.text,
  attempts: activity.attempt,
  waitedMs: activity.waitedMs,
  selector: activity.selector,
  textLength: activity.text ? activity.text.length : 0,
  fullText: activity.text,
};
