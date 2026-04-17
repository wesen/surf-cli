/**
 * Script 01: Page Shape Probe
 * 
 * Probes the Anna's Archive search page to understand:
 * - Page structure and ready state
 * - Search input location
 * - Result container structure
 * - Metadata format
 */

const result = {
  href: location.href,
  title: document.title,
  readyState: document.readyState,
  timestamp: new Date().toISOString(),
  
  // Search input check
  searchInput: null,
  
  // Tab structure
  tabs: [],
  
  // Result count and sample
  resultCount: 0,
  sampleResults: [],
};

// Find search input
const searchInput = document.querySelector('input[placeholder*="Title"], input[placeholder*="DOI"], input[placeholder*="author"]');
if (searchInput) {
  result.searchInput = {
    tagName: searchInput.tagName,
    placeholder: searchInput.placeholder,
    id: searchInput.id,
    name: searchInput.name,
    type: searchInput.type,
    visible: searchInput.offsetParent !== null,
  };
}

// Find tabs
const tabLinks = document.querySelectorAll('a[href*="/search"], a[href*="index="]');
tabLinks.forEach(tab => {
  const text = tab.textContent.trim();
  const href = tab.getAttribute('href');
  if (text && href && (href.includes('index=') || href.includes('/search'))) {
    result.tabs.push({ text, href });
  }
});

// Find result containers - look for main content area
const mainContent = document.querySelector('main');
if (mainContent) {
  // Look for generic divs that contain md5 links
  const allLinks = mainContent.querySelectorAll('a[href^="/md5/"]');
  result.resultCount = allLinks.length;
  
  // Sample first few results
  const samples = Array.from(allLinks).slice(0, 3);
  samples.forEach((link, i) => {
    const container = link.closest('div');
    const parent = container ? container.parentElement : null;
    
    result.sampleResults.push({
      index: i,
      href: link.getAttribute('href'),
      text: link.textContent.trim().substring(0, 100),
      parentStructure: {
        containerTag: container ? container.tagName : null,
        containerClasses: container ? container.className : null,
        parentTag: parent ? parent.tagName : null,
        parentClasses: parent ? parent.className : null,
      },
      // Get metadata from nearby elements
      metadata: extractMetadata(link),
    });
  });
}

function extractMetadata(link) {
  const metadata = {};
  
  // Try to find size info nearby
  const parent = link.closest('div');
  if (parent) {
    const text = parent.textContent;
    
    // Look for file size patterns like "3.1MB", "1.2MB"
    const sizeMatch = text.match(/(\d+\.?\d*)\s*(MB|GB|KB)/i);
    if (sizeMatch) {
      metadata.size = sizeMatch[0];
    }
    
    // Look for format like "PDF", "EPUB"
    const formatMatch = text.match(/\b(PDF|EPUB|MOBI|FB2)\b/i);
    if (formatMatch) {
      metadata.format = formatMatch[1].toUpperCase();
    }
    
    // Look for year
    const yearMatch = text.match(/\b(19|20)\d{2}\b/);
    if (yearMatch) {
      metadata.year = yearMatch[0];
    }
    
    // Look for language
    const langMatch = text.match(/\[([a-z]{2})\]\s*$/i);
    if (langMatch) {
      metadata.language = langMatch[1];
    }
  }
  
  return metadata;
}

console.log(JSON.stringify(result, null, 2));
result;
