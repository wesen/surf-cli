const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

const options = typeof SURF_OPTIONS === 'object' && SURF_OPTIONS !== null ? SURF_OPTIONS : {};
const maxResults = Number.isFinite(options.maxResults) ? Number(options.maxResults) : 10;

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

function resultBlocks() {
  return Array.from(document.querySelectorAll(RESULTS_SELECTOR));
}

function extractResults(limit) {
  const seen = new Set();
  const results = [];
  const blocks = resultBlocks();

  for (const block of blocks) {
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

    let hostname = '';
    try {
      hostname = new URL(url).hostname;
    } catch (_) {
      hostname = '';
    }

    results.push({
      index: results.length + 1,
      title,
      url,
      displayUrl: displayUrl || null,
      snippet: snippet || null,
      snippetLength: snippet.length,
      hostname: hostname || null,
      grouped: block.classList.contains('__srgi') || !!block.closest('.sr-group'),
    });

    if (limit > 0 && results.length >= limit) {
      break;
    }
  }

  return results;
}

const query = new URL(location.href).searchParams.get('q') || '';
const quickAnswer = extractQuickAnswer();
const results = extractResults(maxResults);

return {
  query,
  href: location.href,
  title: document.title,
  waitedMs: ready.waitedMs,
  maxResults,
  quickAnswer,
  resultCount: results.length,
  results,
};
