const rows = Array.from(document.querySelectorAll('tr.zA')).slice(0, 5);
return rows.map((row, i) => ({
  i,
  className: row.className,
  rowAttrs: Object.fromEntries(Array.from(row.attributes).map(a => [a.name, a.value])),
  subjectNode: row.querySelector('.y6 span, .bog span, .bog'),
  participantNode: row.querySelector('.yP, .yW span, .zF'),
  starredNode: !!row.querySelector('[aria-label*="star" i], [title*="star" i], .T-KT'),
  attachmentNode: !!row.querySelector('[aria-label*="Attachment" i], img[alt*="Attachment" i], .aQw'),
  descendantDataAttrs: Array.from(row.querySelectorAll('*')).flatMap(el => Array.from(el.attributes).filter(a => a.name.startsWith('data-')).map(a => ({tag: el.tagName, name: a.name, value: a.value}))).slice(0, 20),
  text: (row.innerText || '').trim().slice(0, 500),
}));
