const conversationId = location.pathname.split('/').filter(Boolean).pop();
if (!conversationId) {
  return { ok: false, error: 'No conversation id in URL', href: location.href };
}

const candidates = [
  '/backend-api/conversation/' + conversationId + '/textdocs',
  '/backend-api/conversation/' + conversationId + '/textdocs?limit=100',
];

const results = [];
for (const url of candidates) {
  try {
    const res = await fetch(url, {
      credentials: 'include',
      headers: {
        accept: 'application/json, text/plain, */*',
      },
    });
    const text = await res.text();
    let parsed = null;
    try {
      parsed = JSON.parse(text);
    } catch (_) {}
    results.push({
      url,
      ok: res.ok,
      status: res.status,
      contentType: res.headers.get('content-type'),
      bodyPreview: text.slice(0, 500),
      bodyKeys: parsed && typeof parsed === 'object' ? Object.keys(parsed).slice(0, 30) : null,
      itemCount: Array.isArray(parsed && parsed.items) ? parsed.items.length : null,
    });
  } catch (error) {
    results.push({ url, ok: false, fetchError: String(error) });
  }
}

return { href: location.href, conversationId, results };
