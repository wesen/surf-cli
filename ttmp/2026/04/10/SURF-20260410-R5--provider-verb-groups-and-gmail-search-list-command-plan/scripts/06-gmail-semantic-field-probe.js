const rows = Array.from(document.querySelectorAll('tr.zA')).slice(0, 3);
return rows.map((row, i) => ({
  i,
  className: row.className,
  participant: row.querySelector('.yP, .yW span[email], .yW')?.innerText || null,
  subject: row.querySelector('.bog, .y6 span[id]')?.innerText || null,
  snippet: row.querySelector('.y2')?.innerText || null,
  timestamp: row.querySelector('.xW span, .xW .xS')?.innerText || null,
  starTitle: row.querySelector('span[title*="star" i], span[data-tooltip*="star" i]')?.getAttribute('title') || row.querySelector('span[title*="star" i], span[data-tooltip*="star" i]')?.getAttribute('data-tooltip') || null,
  labels: Array.from(row.querySelectorAll('.ar, .at')).map(el => (el.innerText || '').trim()).filter(Boolean),
}));
