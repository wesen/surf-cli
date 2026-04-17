var result = {url: location.href, pageType: location.pathname.startsWith('/scidb/') ? 'scidb' : 'unknown', metadata: {}, downloads: { fast: [] }};
var doiLink = document.querySelector('a[href*="doi.org"]');
if (doiLink) {
  result.metadata.doi = doiLink.getAttribute('href').replace('https://doi.org/', '');
}
var allLinks = document.querySelectorAll('a[href]');
for (var i = 0; i < allLinks.length; i++) {
  var link = allLinks[i];
  var href = link.getAttribute('href');
  if (href && href.includes('.pdf') && href.startsWith('http') && href.length > 100) {
    result.downloads.fast.push({url: href, text: 'Download'});
  }
}
return result;
