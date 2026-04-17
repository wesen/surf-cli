var result = {
  url: location.href,
  query: '',
  totalResults: 0,
  results: []
};

// Extract query from URL
var urlParams = new URL(location.href).searchParams;
result.query = urlParams.get('q') || '';

// Find main content area
var main = document.querySelector('main');
if (!main) {
  result.error = 'No main element found';
  return result;
}

// Find the results container - look for the generic containing "Results"
var resultsContainer = null;
var directChildren = main.querySelectorAll(':scope > *');
for (var i = 0; i < directChildren.length; i++) {
  var el = directChildren[i];
  var text = el.textContent || '';
  if (text.includes('Results 1-')) {
    resultsContainer = el;
    break;
  }
}

if (!resultsContainer) {
  result.error = 'Results container not found';
  return result;
}

// Find result count
var containerText = resultsContainer.textContent;
var countMatch = containerText.match(/Results\s+(\d+)-(\d+)\s*\((\d+)\s*total\)/);
if (countMatch) {
  result.totalResults = parseInt(countMatch[3], 10);
}

// Find all MD5 links
var md5Links = resultsContainer.querySelectorAll('a[href^="/md5/"]');
var seenMd5s = {};

for (var i = 0; i < md5Links.length; i++) {
  var link = md5Links[i];
  var href = link.getAttribute('href');
  var md5 = href.replace('/md5/', '');
  
  // Skip if we've already seen this MD5
  if (seenMd5s[md5]) continue;
  seenMd5s[md5] = true;
  
  // Get title from parent structure
  var title = '';
  var parent = link.closest('div');
  if (parent) {
    var links = parent.querySelectorAll('a');
    for (var j = 0; j < links.length; j++) {
      var t = links[j].textContent.trim();
      // Skip dates and short links
      if (t.length > 30 && !t.match(/^\d{4}-\d{2}/)) {
        title = t;
        break;
      }
    }
  }
  
  // Extract metadata from container text
  var format = '';
  var size = '';
  var year = '';
  
  // Get text around this link for metadata
  var linkText = link.textContent || '';
  
  // Format
  var formatMatch = containerText.match(/\b(PDF|EPUB|MOBI)\b/i);
  if (formatMatch) format = formatMatch[0].toUpperCase();
  
  // Size
  var sizeMatch = containerText.match(/(\d+\.?\d*)\s*(MB|GB|KB)/i);
  if (sizeMatch) size = sizeMatch[0];
  
  // Year - look for 4-digit years
  var yearMatch = containerText.match(/\b(19|20)\d{2}\b/);
  if (yearMatch) year = yearMatch[0];
  
  result.results.push({
    md5: md5,
    href: href,
    title: title,
    format: format,
    size: size,
    year: year
  });
  
  // Limit results
  if (result.results.length >= 20) break;
}

return result;
