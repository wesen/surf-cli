/**
 * Script 03: Paper Detail Page Extractor
 * 
 * Extracts paper metadata and download options from Anna's Archive paper pages.
 * Validated on: https://annas-archive.gl/md5/d89c394b00116f093b5d9d6a6611f975
 */

const options = typeof SURF_OPTIONS === 'object' && SURF_OPTIONS !== null ? SURF_OPTIONS : {
  extractDownloads: true,
};

const result = {
  href: location.href,
  title: document.title,
  timestamp: new Date().toISOString(),
  
  // Paper metadata
  metadata: {
    title: '',
    authors: [],
    journal: '',
    doi: '',
    year: '',
    format: '',
    size: '',
    language: '',
    pages: '',
    publisher: '',
  },
  
  // MD5 from URL
  md5: location.pathname.replace('/md5/', ''),
  
  // Download options
  downloads: {
    scidb: [],
    fast: [],
    slow: [],
    external: [],
  },
  
  // Errors/warnings
  errors: [],
};

try {
  // Extract title - h1 or prominent heading
  const heading = document.querySelector('main h1') || document.querySelector('main h2') || document.querySelector('main h3');
  if (heading) {
    result.metadata.title = heading.textContent.trim();
  }
  
  // Extract DOI from page text or metadata
  const pageText = document.body.textContent;
  const doiMatch = pageText.match(/doi[:\s]*([\d.]+\/[\w./%-]+)/i);
  if (doiMatch) {
    result.metadata.doi = doiMatch[1];
  }
  
  // Extract metadata from "description" section
  const descSection = document.querySelector('[class*="description"], [class*="metadata"]');
  if (descSection) {
    const descText = descSection.textContent;
    
    // Format
    const formatMatch = descText.match(/\b(PDF|EPUB|MOBI|FB2)\b/i);
    if (formatMatch) result.metadata.format = formatMatch[1].toUpperCase();
    
    // Size
    const sizeMatch = descText.match(/(\d+\.?\d*)\s*(MB|GB|KB)/i);
    if (sizeMatch) result.metadata.size = sizeMatch[0];
    
    // Year
    const yearMatch = descText.match(/\b(19|20)\d{2}\b/);
    if (yearMatch) result.metadata.year = yearMatch[0];
  }
  
  // Extract downloads
  if (options.extractDownloads) {
    extractDownloads(result);
  }
  
} catch (e) {
  result.errors.push(e.message);
}

function extractDownloads(result) {
  // Look for download section
  const downloadSections = document.querySelectorAll('section, div, [role="tabpanel"], [aria-label]');
  
  downloadSections.forEach(section => {
    const sectionText = section.textContent.toLowerCase();
    
    if (sectionText.includes('download')) {
      // Find all download links
      const links = section.querySelectorAll('a[href*="/fast_download/"], a[href*="/slow_download/"], a[href*="/scidb"]');
      
      links.forEach(link => {
        const href = link.getAttribute('href');
        const text = link.textContent.trim();
        
        if (href.includes('/fast_download/')) {
          const parts = href.replace('/fast_download/', '').split('/');
          result.downloads.fast.push({
            url: href,
            serverIndex: parseInt(parts[2], 10),
            recommended: text.includes('recommended'),
            viewer: href.includes('viewer='),
          });
        } else if (href.includes('/slow_download/')) {
          const parts = href.replace('/slow_download/', '').split('/');
          result.downloads.slow.push({
            url: href,
            serverIndex: parseInt(parts[2], 10),
            recommended: text.includes('recommended'),
          });
        } else if (href.includes('/scidb')) {
          result.downloads.scidb.push({
            url: href,
            text: text,
          });
        }
      });
    }
  });
  
  // Alternative: find all download links anywhere on page
  const allLinks = document.querySelectorAll('a[href*="download"], a[href*="scidb"]');
  allLinks.forEach(link => {
    const href = link.getAttribute('href');
    const text = link.textContent.trim();
    
    if (href && href.startsWith('/') && !href.startsWith('//')) {
      if (href.includes('/fast_download/') && !result.downloads.fast.find(d => d.url === href)) {
        const parts = href.replace('/fast_download/', '').split('/');
        result.downloads.fast.push({
          url: href,
          serverIndex: parseInt(parts[2], 10),
          text: text,
        });
      } else if (href.includes('/slow_download/') && !result.downloads.slow.find(d => d.url === href)) {
        result.downloads.slow.push({
          url: href,
          text: text,
        });
      } else if (href.includes('/scidb') && !result.downloads.scidb.find(d => d.url === href)) {
        result.downloads.scidb.push({
          url: href,
          text: text,
        });
      }
    }
  });
}

console.log(JSON.stringify(result, null, 2));
result;
