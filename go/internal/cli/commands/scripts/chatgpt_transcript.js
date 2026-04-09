const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

const options = typeof SURF_OPTIONS === 'object' && SURF_OPTIONS !== null ? SURF_OPTIONS : {};
const withActivity = !!options.withActivity;
const activityLimit = Number.isFinite(options.activityLimit) ? Number(options.activityLimit) : 0;

const ACTIVITY_SELECTORS = [
  'body > div:nth-child(5) > div > div.relative.z-0.flex.min-h-0.w-full.flex-1 > div.bg-token-bg-elevated-secondary.relative.z-1.shrink-0.overflow-x-hidden.max-lg\\:w-0\\!.stage-thread-flyout-preset-default > div > div > section',
  '[class*="stage-thread-flyout"] section',
];

const ACTIVITY_HEADER_RE = /Activity\s*[^\w\n]?\s*[^\n]+/i;
const SOURCES_HEADER_RE = /Sources\s*[^\w\n]?\s*\d+/i;

function durationFromButtonText(text) {
  return (text || '').replace(/^Thought for\s*/i, '').trim();
}

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

function activityMatchesDuration(activityText, duration) {
  if (!activityText || !duration) return false;
  const escaped = duration.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  return new RegExp(`Activity\\s*[^\\w\\n]?\\s*${escaped}`, 'i').test(activityText);
}

async function normalizeAndOpen(button) {
  const buttonText = (button.textContent || '').trim().replace(/\s+/g, ' ');
  const duration = durationFromButtonText(buttonText);
  const previousActivity = getActivitySection();
  const previousText = previousActivity ? previousActivity.text : null;

  button.scrollIntoView({ block: 'center', inline: 'center' });
  await sleep(250);

  for (let attempt = 1; attempt <= 2; attempt += 1) {
    button.click();
    const hit = await waitForCondition(() => {
      const activity = getActivitySection();
      if (!activity) return null;
      if (!activityMatchesDuration(activity.text, duration)) return null;
      if (previousText && activity.text === previousText && attempt === 1) return null;
      return activity;
    }, 7000);
    if (hit) {
      return { attempt, duration, buttonText, ...hit.value, waitedMs: hit.waitedMs };
    }
    await sleep(400);
  }

  return { attempt: 2, duration, buttonText, selector: null, node: null, text: null, waitedMs: null };
}

function extractSectionMessage(section, index) {
  const candidates = Array.from(section.querySelectorAll('[data-message-author-role]'));
  if (candidates.length === 0) {
    return null;
  }

  const byMessageId = new Map();
  for (const node of candidates) {
    const role = node.getAttribute('data-message-author-role') || 'unknown';
    const messageId = node.getAttribute('data-message-id') || `${role}:${index}:${byMessageId.size}`;
    const model = node.getAttribute('data-message-model-slug') || null;
    const text = (node.innerText || '').trim();
    if (!text) {
      continue;
    }
    const existing = byMessageId.get(messageId);
    if (!existing || text.length > existing.text.length) {
      byMessageId.set(messageId, {
        role,
        model,
        messageId,
        text,
        textLength: text.length,
      });
    }
  }

  const items = Array.from(byMessageId.values());
  if (items.length === 0) {
    return null;
  }
  items.sort((a, b) => b.textLength - a.textLength);
  const best = items[0];
  const thoughtButton = Array.from(section.querySelectorAll('button')).find((node) => /Thought for/i.test((node.textContent || '').trim()));
  const thoughtButtonText = thoughtButton ? (thoughtButton.textContent || '').trim().replace(/\s+/g, ' ') : null;

  return {
    index,
    sectionTestId: section.getAttribute('data-testid') || null,
    role: best.role,
    model: best.model,
    messageId: best.messageId,
    textLength: best.textLength,
    text: best.text,
    hasThought: !!thoughtButton,
    thoughtButtonText,
    thoughtDuration: thoughtButtonText ? durationFromButtonText(thoughtButtonText) : null,
    _button: thoughtButton,
  };
}

const href = location.href;
const title = document.title;
const sections = Array.from(document.querySelectorAll('section[data-testid^="conversation-turn-"]'));
const items = [];
for (let i = 0; i < sections.length; i += 1) {
  const item = extractSectionMessage(sections[i], items.length);
  if (item) {
    items.push(item);
  }
}

let activityExported = 0;
if (withActivity) {
  for (const item of items) {
    if (item.role !== 'assistant' || !item._button) {
      continue;
    }
    if (activityLimit > 0 && activityExported >= activityLimit) {
      break;
    }
    const activity = await normalizeAndOpen(item._button);
    item.activityFound = !!activity.text;
    item.activityText = activity.text || null;
    item.activityTextLength = activity.text ? activity.text.length : 0;
    item.activitySelector = activity.selector;
    item.activityAttempts = activity.attempt;
    item.activityWaitedMs = activity.waitedMs;
    activityExported += 1;
    await sleep(500);
  }
}

const transcript = items.map((item) => ({
  index: item.index,
  sectionTestId: item.sectionTestId,
  role: item.role,
  model: item.model,
  messageId: item.messageId,
  textLength: item.textLength,
  text: item.text,
  hasThought: item.hasThought,
  thoughtButtonText: item.thoughtButtonText,
  thoughtDuration: item.thoughtDuration,
  activityFound: item.activityFound || false,
  activityTextLength: item.activityTextLength || 0,
  activityText: item.activityText || null,
  activitySelector: item.activitySelector || null,
  activityAttempts: item.activityAttempts || null,
  activityWaitedMs: item.activityWaitedMs || null,
}));

return {
  href,
  title,
  turnCount: transcript.length,
  withActivity,
  activityLimit,
  activityExported,
  transcript,
};
