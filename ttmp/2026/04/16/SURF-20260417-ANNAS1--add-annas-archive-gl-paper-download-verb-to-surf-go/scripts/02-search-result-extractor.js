/**
 * Script 02: Search Result Extractor
 * 
 * Extracts structured paper results from Anna's Archive search page.
 * Validated with DOI search: 10.1038/nature12373
 */

const options = typeof SURF_OPTIONS === 'object' && SURF_OPTIONS !== null ? SURF_OPTIONS : {
  maxResults: 10,
  waitForSelector: 'main a[href^="/md5/"]',
  waitMs: 3000,
};

async function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

// Wait for results if needed
if (options.waitMs > 0) {
  await sleep(options.waitMs);
}

const result = {
  href: location.href,
  title: document.title,
  timestamp: new Date().toISOString(),
  query: new URL(location.href).searchParams.get('q') || '',
  index: new URL(location.href).searchParams.get('index') || '',
  
  // Summary
  totalResults: 0,
  exactMatches: 0,
  partialMatchesButton: null,
  
  // Papers
  papers: [],
};

// Find main content
const main = document.querySelector('main');
if (!main) {
  result.error = 'No main element found';
  return result;
}

// Find result count text
const countText = main.textContent.match(/Results\s+(\d+)-(\d+)\s*\((\d+)\s*total\)/);
if (countText) {
  result.totalResults = parseInt(countText[3], 10);
}

// Check for partial matches button
const partialBtn = main.querySelector('button');
if (partialBtn && partialBtn.textContent.includes('partial matches')) {
  result.partialMatchesButton = {
    text: partialBtn.textContent.trim(),
    visible: partialBtn.offsetParent !== null,
  };
}

// Find all result containers - look for divs with paper metadata
const resultContainers = [];
const md5Links = main.querySelectorAll('a[href^="/md5/"]');

// Deduplicate by href since same MD5 might appear multiple times
const seenMd5s = new Set();
md5Links.forEach(link => {
  const href = link.getAttribute('href');
  const md5 = href.replace('/md5/', '');
  
  if (!seenMd5s.has(md5)) {
    seenMd5s.add(md5);
    
    // Get the full result container
    const container = link.closest('div');
    const outerContainer = container ? container.closest('div') : null;
    
    // Get text content for metadata extraction
    const fullText = outerContainer ? outerContainer.textContent : link.textContent;
    
    // Extract metadata from text
    const metadata = extractMetadata(fullText, md5);
    
    // Get title and author links
    const allLinks = outerContainer ? outerContainer.querySelectorAll('a[href^="/md5/"], a[href*="/search?q="]') : [link];
    const titleLinks = [];
    allLinks.forEach(l => {
      const href = l.getAttribute('href');
      if (href && !href.startsWith('/md5/') && l.textContent.trim()) {
        titleLinks.push({
          text: l.textContent.trim().substring(0, 200),
          href: href,
        });
      }
    });
    
    resultContainers.push({
      md5: md5,
      href: href,
      title: metadata.title,
      authors: metadata.authors,
      metadata: metadata,
      titleLinks: titleLinks,
      // Store raw link for title
      titleHref: link.getAttribute('href'),
    });
  }
});

result.exactMatches = resultContainers.length;
result.papers = resultContainers.slice(0, options.maxResults);

function extractMetadata(text, md5) {
  const meta = {
    md5: md5,
    title: '',
    authors: [],
    format: '',
    size: '',
    year: '',
    language: '',
    source: [],
    doi: '',
    publisher: '',
    journal: '',
    pages: '',
  };
  
  // Clean text
  text = text.replace(/\s+/g, ' ').trim();
  
  // Extract title - usually before author names or first comma after a title-like string
  // Look for patterns like "Title" before author names
  const titleMatch = text.match(/([A-Z][^.]*?[A-Z][^,]*?)\s+(Kucsko|Y.*?;|Author|et al)/i);
  if (titleMatch) {
    meta.title = titleMatch[1].trim();
  }
  
  // Extract format (PDF, EPUB, etc)
  const formatMatch = text.match(/\b(PDF|EPUB|MOBI|FB2|DJVU)\b/i);
  if (formatMatch) {
    meta.format = formatMatch[1].toUpperCase();
  }
  
  // Extract file size
  const sizeMatch = text.match(/(\d+\.?\d*)\s*(MB|GB|KB)/i);
  if (sizeMatch) {
    meta.size = sizeMatch[0];
  }
  
  // Extract year
  const yearMatch = text.match(/\b(19|20)\d{2}\b/);
  if (yearMatch) {
    meta.year = yearMatch[0];
  }
  
  // Extract language code
  const langMatch = text.match(/\[([a-z]{2})\]\s*[·•]|$/i);
  if (langMatch && langMatch[1]) {
    meta.language = langMatch[1];
  }
  
  // Extract sources
  const sourcePatterns = [
    /🧬\s*(\w+)/g,
    /🚀\s*(\w+)/g,
    /zlib|lgli|scihub|nexusstc|upload|hathi|ia|duxiu/gi
  ];
  const sources = [];
  sourcePatterns.forEach(pattern => {
    let match;
    while ((match = pattern.exec(text)) !== null) {
      const src = match[1] || match[0];
      if (!sources.includes(src.toLowerCase())) {
        sources.push(src.toLowerCase());
      }
    }
  });
  meta.source = sources;
  
  // Extract DOI
  const doiMatch = text.match(/doi[:\s]*([\d.]+\/[\w./-]+)/i);
  if (doiMatch) {
    meta.doi = doiMatch[1];
  }
  
  return meta;
}

console.log(JSON.stringify(result, null, 2));
result;
