const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

function normalizeText(value) {
  return (value || '').replace(/\s+/g, ' ').trim();
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

const RESULTS_SELECTOR = 'main div._0_SRI.search-result, main div.__srgi';
const QUICK_ANSWER_SELECTOR = 'main .qa-container-box .qa-content';

const ready = await waitForCondition(() => document.querySelector(RESULTS_SELECTOR), 15000);
if (!ready) {
  throw new Error('No Kagi search results found');
}

function extractQuickAnswer() {
  const node = document.querySelector(QUICK_ANSWER_SELECTOR);
  if (!node) {
    return null;
  }
  const title = normalizeText(node.querySelector('h3')?.textContent || '');
  const rawText = normalizeText(node.innerText || '');
  const text = normalizeText(
    rawText
      .replace(/\bQuick Answer\b/gi, '')
      .replace(/\bReferences\b/gi, '')
      .replace(/\bContinue in Assistant\b/gi, '')
  );
  if (!text) {
    return null;
  }
  return {
    title: title || null,
    text,
    textLength: text.length,
  };
}

const seen = new Set();
const results = [];
for (const block of document.querySelectorAll(RESULTS_SELECTOR)) {
  const titleLink = block.querySelector('h3 a[href^="http"]');
  if (!titleLink) {
    continue;
  }
  const url = titleLink.href || '';
  const title = normalizeText(titleLink.textContent || '');
  if (!url || !title || seen.has(url)) {
    continue;
  }
  seen.add(url);

  const displayUrl = normalizeText(
    block.querySelector('.__sri-url-box, .__sri-url, .__sri_url_path_box')?.textContent || ''
  );
  const snippet = normalizeText(
    block.querySelector('._0_DESC.__sri-desc, .__sri-desc')?.textContent || ''
  );

  results.push({
    index: results.length + 1,
    title,
    url,
    displayUrl: displayUrl || null,
    snippet: snippet || null,
    grouped: block.classList.contains('__srgi') || !!block.closest('.sr-group'),
  });

  if (results.length >= 5) {
    break;
  }
}

return {
  query: new URL(location.href).searchParams.get('q') || '',
  href: location.href,
  title: document.title,
  waitedMs: ready.waitedMs,
  quickAnswer: extractQuickAnswer(),
  resultCount: results.length,
  results,
};
