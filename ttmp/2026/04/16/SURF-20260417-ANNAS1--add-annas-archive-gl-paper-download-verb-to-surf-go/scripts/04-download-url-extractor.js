/**
 * Script 04: Download URL Extractor from SciDB page
 * 
 * Extracts the actual download URL from Anna's Archive SciDB page.
 * Tested with DOI: 10.1038/nature12373
 */

const result = {
  url: location.href,
  title: document.title,
  timestamp: new Date().toISOString(),
  
  // Paper metadata
  paper: {
    doi: '',
    title: '',
    authors: '',
    format: '',
    size: '',
    sources: [],
  },
  
  // Download links
  downloads: {
    direct: null,
    scidbViewer: null,
    externalSources: [],
  },
};

try {
  // Extract DOI from page
  const doiEl = document.querySelector('a[href*="doi.org"]');
  if (doiEl) {
    result.paper.doi = doiEl.getAttribute('href').replace('https://doi.org/', '');
  }
  
  // Extract title
  const titleEl = document.querySelector('body');
  if (titleEl) {
    const text = titleEl.textContent;
    // Look for the title pattern
    const titleMatch = text.match(/([A-Z][^.]+thermometry[^.]+\.)/);
    if (titleMatch) {
      result.paper.title = titleMatch[1].trim();
    }
    // Alternative: look near "Download" link
    const downloadLink = document.querySelector('a[href*="Download"]');
    if (downloadLink) {
      const parent = downloadLink.closest('ul, div');
      if (parent) {
        const allText = parent.textContent;
        // Extract title from URL
        const urlMatch = downloadLink.getAttribute('href').match(/--\s*([^--]+)\s*--/);
        if (urlMatch) {
          result.paper.title = decodeURIComponent(urlMatch[1].replace(/%20/g, ' ').trim());
        }
      }
    }
  }
  
  // Extract metadata from main content
  const metaText = document.querySelector('body > div > div')?.textContent || '';
  const formatMatch = metaText.match(/\.pdf/i);
  if (formatMatch) result.paper.format = 'PDF';
  
  const sizeMatch = metaText.match(/(\d+\.?\d*)\s*(MB|GB|KB)/i);
  if (sizeMatch) result.paper.size = sizeMatch[0];
  
  // Find all links
  const allLinks = document.querySelectorAll('a[href]');
  allLinks.forEach(link => {
    const href = link.getAttribute('href');
    const text = link.textContent.trim();
    
    // Direct download link (external URL)
    if (href.includes('.pdf') && href.startsWith('http') && !href.includes('scihub') && !href.includes('doi.org')) {
      result.downloads.direct = {
        url: href,
        text: 'Download',
      };
    }
    
    // Record link (Anna's Archive)
    if (href.includes('/md5/')) {
      result.downloads.scidbViewer = {
        url: href,
        text: 'Record in Anna\'s Archive',
      };
    }
    
    // External sources
    if (href.includes('sci-hub') || href.includes('doi.org')) {
      result.downloads.externalSources.push({
        url: href,
        text: text,
      });
    }
  });
  
  // Also extract from list items
  const listItems = document.querySelectorAll('li');
  listItems.forEach(li => {
    const text = li.textContent.trim();
    const link = li.querySelector('a');
    if (!link) return;
    
    const href = link.getAttribute('href');
    
    // Download link in list
    if (text.includes('Download') && href.startsWith('http')) {
      result.downloads.direct = {
        url: href,
        text: 'Download',
      };
    }
  });
  
} catch (e) {
  result.error = e.message;
}

console.log(JSON.stringify(result, null, 2));
result;
