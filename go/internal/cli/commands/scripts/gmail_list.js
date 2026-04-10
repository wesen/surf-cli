return (async () => {
  const options = typeof SURF_OPTIONS === 'object' && SURF_OPTIONS !== null ? SURF_OPTIONS : {};
  const maxResults = Number.isFinite(options.maxResults) ? Number(options.maxResults) : 25;

  const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));
  const normalizeText = (value) => (value || '').replace(/\u00a0/g, ' ').replace(/\s+/g, ' ').trim();
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
  const started = Date.now();
  while (Date.now() - started < 15000) {
    const rows = extractRows();
    if (rows.length > 0) {
      const trimmed = maxResults > 0 ? rows.slice(0, maxResults) : rows;
      return {
        href: location.href,
        title: document.title,
        mailbox: 'inbox',
        waitedMs: Date.now() - started,
        resultCount: trimmed.length,
        threads: trimmed,
      };
    }
    await sleep(250);
  }
  return {
    href: location.href,
    title: document.title,
    mailbox: 'inbox',
    waitedMs: Date.now() - started,
    resultCount: 0,
    threads: [],
  };
})()
