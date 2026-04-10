return (async () => {
  const options = typeof SURF_OPTIONS === 'object' && SURF_OPTIONS !== null ? SURF_OPTIONS : {};
  const query = (options.query || '').trim();
  const maxResults = Number.isFinite(options.maxResults) ? Number(options.maxResults) : 25;
  if (!query) {
    throw new Error('query required');
  }

  const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));
  const normalizeText = (value) => (value || '').replace(/\u00a0/g, ' ').replace(/\s+/g, ' ').trim();
  const expectedHash = '#search/' + encodeURIComponent(query).replace(/%20/g, '+');
  const exactTitle = (row, value) =>
    Array.from(row.querySelectorAll('[title], [data-tooltip]')).some((el) => {
      const candidate = normalizeText(el.getAttribute('title') || el.getAttribute('data-tooltip') || '');
      return candidate.toLowerCase() === value.toLowerCase();
    });
  const extractThreadIds = (row) => {
    const node = row.querySelector('[data-thread-id][data-legacy-thread-id]') || row.querySelector('[data-thread-id]') || row.querySelector('[data-legacy-thread-id]');
    return {
      threadId: node?.getAttribute('data-thread-id') || null,
      legacyThreadId: node?.getAttribute('data-legacy-thread-id') || null,
    };
  };
  const extractRows = () => {
    return Array.from(document.querySelectorAll('tr.zA')).map((row, i) => {
      const ids = extractThreadIds(row);
      const participant = normalizeText(row.querySelector('.yP, .yW span[email], .yW')?.innerText);
      const subject = normalizeText(row.querySelector('.bog, .y6 span[id]')?.innerText);
      const snippet = normalizeText(row.querySelector('.y2')?.innerText).replace(/^-+\s*/, '');
      const timestamp = normalizeText(row.querySelector('.xW span, .xW .xS')?.innerText);
      const rawText = normalizeText(row.innerText);
      return {
        index: i + 1,
        threadId: ids.threadId,
        legacyThreadId: ids.legacyThreadId,
        unread: row.classList.contains('zE'),
        starred: exactTitle(row, 'Starred'),
        hasAttachment: rawText.includes('Attachment:') || !!row.querySelector('.aQw, img[alt*="Attachment"], span[aria-label*="Attachment"]'),
        participant,
        subject,
        snippet,
        timestamp,
        rawText,
      };
    }).filter((row) => row.subject || row.participant || row.snippet);
  };

  const input = document.querySelector('input[name="q"], input[aria-label*="Search mail"]');
  if (!input) {
    throw new Error('gmail search input not found');
  }
  const initialSnapshot = extractRows().slice(0, 5).map((row) => row.legacyThreadId || row.threadId || row.subject).join('|');
  input.focus();
  input.value = query;
  input.dispatchEvent(new InputEvent('input', { bubbles: true, data: query, inputType: 'insertText' }));
  input.dispatchEvent(new Event('change', { bubbles: true }));
  const form = input.closest('form');
  form?.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }));
  input.dispatchEvent(new KeyboardEvent('keydown', { bubbles: true, cancelable: true, key: 'Enter', code: 'Enter' }));
  input.dispatchEvent(new KeyboardEvent('keypress', { bubbles: true, cancelable: true, key: 'Enter', code: 'Enter' }));
  input.dispatchEvent(new KeyboardEvent('keyup', { bubbles: true, cancelable: true, key: 'Enter', code: 'Enter' }));
  document.querySelector('button[aria-label="Search mail"]')?.click();

  const started = Date.now();
  while (Date.now() - started < 20000) {
    const inSearchView = location.href.includes(expectedHash) || /Search results/i.test(document.title);
    const rows = extractRows();
    const currentSnapshot = rows.slice(0, 5).map((row) => row.legacyThreadId || row.threadId || row.subject).join('|');
    const noResults = /No messages matched your search/i.test(document.body?.innerText || '');
    if ((inSearchView && rows.length > 0 && currentSnapshot !== initialSnapshot) || noResults) {
      const trimmed = maxResults > 0 ? rows.slice(0, maxResults) : rows;
      return {
        href: location.href,
        title: document.title,
        query,
        waitedMs: Date.now() - started,
        noResults,
        resultCount: trimmed.length,
        threads: trimmed,
      };
    }
    await sleep(250);
  }
  return {
    href: location.href,
    title: document.title,
    query,
    waitedMs: Date.now() - started,
    noResults: false,
    resultCount: 0,
    threads: [],
  };
})()
