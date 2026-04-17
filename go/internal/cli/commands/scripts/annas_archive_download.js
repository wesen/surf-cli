var result = {
  url: location.href,
  pageType: location.pathname.startsWith('/scidb/') ? 'scidb' : 'unknown',
  metadata: {},
  downloads: { fast: [] }
};

// Check if we're on a SciDB page
if (result.pageType !== 'scidb') {
  result.error = 'Not on a SciDB page. Navigate to /scidb/{doi} first.';
  return result;
}

// Extract DOI from link
var doiLink = document.querySelector('a[href*="doi.org"]');
if (doiLink) {
  result.metadata.doi = doiLink.getAttribute('href').replace('https://doi.org/', '');
}

// Extract DOI from URL path
var pathMatch = location.pathname.match(/\/scidb\/(.+)\/?$/);
if (pathMatch && !result.metadata.doi) {
  result.metadata.doi = decodeURIComponent(pathMatch[1]);
}

// Extract title from page
var bodyText = document.body.innerText;
var titleMatch = bodyText.match(/([A-Z][^.!?]{20,150})/);
if (titleMatch) {
  result.metadata.title = titleMatch[1].trim().split('\n')[0];
}

// Extract format and size
var metaMatch = bodyText.match(/\.pdf[,\s]*(\d+\.?\d*\s*(MB|GB|KB))/i);
if (metaMatch) {
  result.metadata.format = 'PDF';
  result.metadata.size = metaMatch[1].trim();
}

// Find download links
var allLinks = document.querySelectorAll('a[href]');
for (var i = 0; i < allLinks.length; i++) {
  var link = allLinks[i];
  var href = link.getAttribute('href');
  if (href && href.includes('.pdf') && href.startsWith('http') && href.length > 100) {
    result.downloads.fast.push({
      url: href,
      text: 'Download'
    });
  }
}

// Extract MD5 from any link
var md5Match = bodyText.match(/([a-f0-9]{32})/i);
if (md5Match) {
  result.md5 = md5Match[1];
}

return result;
