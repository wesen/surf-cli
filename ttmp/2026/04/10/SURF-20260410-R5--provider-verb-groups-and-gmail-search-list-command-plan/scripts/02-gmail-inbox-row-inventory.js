const rows = Array.from(document.querySelectorAll('tr[role="row"], [role="main"] table tr'));
return {
  href: location.href,
  title: document.title,
  rowCount: rows.length,
  sample: rows.slice(0, 8).map((row, i) => ({
    i,
    className: row.className,
    dataThreadPermId: row.getAttribute('data-legacy-thread-id') || row.getAttribute('data-thread-id') || null,
    ariaLabel: row.getAttribute('aria-label'),
    text: (row.innerText || '').trim().slice(0, 400),
    linkHrefs: Array.from(row.querySelectorAll('a[href]')).slice(0, 4).map(a => a.getAttribute('href')),
  })),
};
