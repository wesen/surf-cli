/**
 * annas_archive.js - Anna's Archive paper download extractor
 * 
 * Extracts paper metadata and download options from Anna's Archive.
 * Supports both SciDB paper page and search page extraction.
 * 
 * Modes:
 *   - paper: Extract from SciDB page (when URL is /scidb/{doi})
 *   - search: Extract from search results page
 */

const options = typeof SURF_OPTIONS === 'object' && SURF_OPTIONS !== null ? SURF_OPTIONS : {
  maxResults: 10,
  waitMs: 1000,
};

// Determine which page we're on
const isSearchPage = location.pathname === '/search' || location.search.includes('q=');
const isScidbPage = location.pathname.startsWith('/scidb/');

let result = {
  url: location.href,
  timestamp: new Date().toISOString(),
  pageType: isScidbPage ? 'scidb' : isSearchPage ? 'search' : 'unknown',
};

if (isScidbPage) {
  result = { ...result, ...extractScidbPage() };
} else if (isSearchPage) {
  result = { ...result, ...extractSearchResults() };
} else {
  result.error = 'Unsupported page type. Navigate to a SciDB paper page or search page.';
}

function extractSearchResults() {
  const data = {
    query: new URL(location.href).searchParams.get('q') || '',
    index: new URL(location.href).searchParams.get('index') || '',
    totalResults: 0,
    papers: [],
  };
  
  // Find result count
  const pageText = document.body.textContent;
  const countMatch = pageText.match(/Results\s+(\d+)-(\d+)\s*\((\d+)\s*total\)/);
  if (countMatch) {
    data.totalResults = parseInt(countMatch[3], 10);
  }
  
  // Find main content area
  const main = document.querySelector('main');
  if (!main) {
    return { error: 'No main element found' };
  }
  
  // Find result containers - dedupe by MD5
  const md5Links = main.querySelectorAll('a[href^="/md5/"]');
  const seenMd5s = new Set();
  let count = 0;
  
  md5Links.forEach(link => {
    if (count >= options.maxResults) return;
    
    const href = link.getAttribute('href');
    const md5 = href.replace('/md5/', '');
    
    if (seenMd5s.has(md5)) return;
    seenMd5s.add(md5);
    count++;
    
    // Get container with metadata
    const container = link.closest('div');
    const outer = container ? container.closest('div') : null;
    const text = outer ? outer.textContent : '';
    
    // Extract metadata
    const paper = extractPaperMetadata(text, md5);
    paper.href = href;
    paper.md5 = md5;
    
    data.papers.push(paper);
  });
  
  return data;
}

function extractScidbPage() {
  const data = {
    metadata: {},
    downloads: {
      scidb: [],
      fast: [],
      slow: [],
    },
  };
  
  // Extract DOI from page URL or link
  const doiLink = document.querySelector('a[href*="doi.org"]');
  if (doiLink) {
    const href = doiLink.getAttribute('href');
    data.metadata.doi = href.replace('https://doi.org/', '');
  }
  
  // Extract DOI from URL path if present
  const scidbPath = location.pathname;
  const doiMatch = scidbPath.match(/\/scidb\/(.+)/);
  if (doiMatch && !data.metadata.doi) {
    data.metadata.doi = decodeURIComponent(doiMatch[1]);
  }
  
  // Extract title - look for the main content
  const bodyText = document.body.innerText;
  const titleMatch = bodyText.match(/([A-Z][^.!?]{20,150})/);
  if (titleMatch) {
    data.metadata.title = titleMatch[1].trim().split('\n')[0];
  }
  
  // Extract format and size from metadata text
  const metaMatch = bodyText.match(/\.pdf[,\s]*(\d+\.?\d*\s*(MB|GB|KB))/i);
  if (metaMatch) {
    data.metadata.format = 'PDF';
    data.metadata.size = metaMatch[1].trim();
  }
  
  // Find all links
  const allLinks = document.querySelectorAll('a[href]');
  allLinks.forEach(link => {
    const href = link.getAttribute('href');
    const text = link.textContent.trim();
    
    // Direct download link (external URL with .pdf)
    if (href && href.includes('.pdf') && href.startsWith('http') && href.length > 100) {
      data.downloads.fast.push({
        url: href,
        text: 'Download',
        recommended: true,
      });
    }
    
    // Anna's Archive record link
    if (href && href.includes('/md5/')) {
      data.downloads.scidb.push({
        url: href,
        text: 'Record in Anna\'s Archive',
      });
    }
    
    // Sci-Hub link
    if (href && href.includes('sci-hub')) {
      data.downloads.slow.push({
        url: href,
        text: 'Sci-Hub',
      });
    }
  });
  
  return data;
}

function extractPaperDetails() {
  const data = {
    md5: location.pathname.replace('/md5/', ''),
    metadata: {},
    downloads: {
      scidb: [],
      fast: [],
      slow: [],
    },
  };
  
  // Extract title
  const main = document.querySelector('main');
  if (main) {
    const heading = main.querySelector('h1, h2');
    if (heading) {
      data.metadata.title = heading.textContent.trim();
    }
    
    const pageText = main.textContent;
    
    // Extract DOI
    const doiMatch = pageText.match(/doi[:\s]*([\d.]+\/[\w./%-]+)/i);
    if (doiMatch) {
      data.metadata.doi = doiMatch[1];
    }
    
    // Extract format
    const formatMatch = pageText.match(/\b(PDF|EPUB|MOBI|FB2|DJVU)\b/i);
    if (formatMatch) {
      data.metadata.format = formatMatch[1].toUpperCase();
    }
    
    // Extract size
    const sizeMatch = pageText.match(/(\d+\.?\d*)\s*(MB|GB|KB)/i);
    if (sizeMatch) {
      data.metadata.size = sizeMatch[0];
    }
    
    // Extract year
    const yearMatch = pageText.match(/\b(19|20)\d{2}\b/);
    if (yearMatch) {
      data.metadata.year = yearMatch[0];
    }
    
    // Extract authors - look for patterns like "Name, Name; Name"
    const authorMatch = pageText.match(/([A-Z][a-z]+,\s*[A-Z][.\s]+(?:;[A-Z][a-z]+,\s*[A-Z][.\s]+)*)/);
    if (authorMatch) {
      data.metadata.authors = authorMatch[1].split(';').map(a => a.trim());
    }
    
    // Extract journal/publisher info
    const journalMatch = pageText.match(/(?:Nature|Science|Cell|PNAS|arxiv|Elsevier|Springer)[\w\s,]*?(?=\s+\d{4}|\s+PDF|\s+·\s+\d)/i);
    if (journalMatch) {
      data.metadata.journal = journalMatch[0].trim();
    }
  }
  
  // Extract download links
  const allLinks = document.querySelectorAll('a[href*="download"], a[href*="scidb"]');
  
  allLinks.forEach(link => {
    const href = link.getAttribute('href');
    const text = link.textContent.trim();
    
    if (!href || !href.startsWith('/')) return;
    
    if (href.includes('/fast_download/')) {
      // Check if we already have this URL
      const exists = data.downloads.fast.some(d => d.url === href);
      if (!exists) {
        const parts = href.replace('/fast_download/', '').split('/');
        data.downloads.fast.push({
          url: href,
          serverIndex: parseInt(parts[2], 10),
          text: text.substring(0, 60),
          recommended: text.toLowerCase().includes('recommended'),
        });
      }
    } else if (href.includes('/slow_download/')) {
      const exists = data.downloads.slow.some(d => d.url === href);
      if (!exists) {
        const parts = href.replace('/slow_download/', '').split('/');
        data.downloads.slow.push({
          url: href,
          serverIndex: parseInt(parts[2], 10),
          text: text.substring(0, 60),
        });
      }
    } else if (href.includes('/scidb')) {
      const exists = data.downloads.scidb.some(d => d.url === href);
      if (!exists && !href.includes('?')) {
        // Only add base scidb URLs without query params as main entries
        data.downloads.scidb.push({
          url: href,
          text: text.substring(0, 60),
        });
      }
    }
  });
  
  // Add SciDB DOI URL if we have a DOI
  if (data.metadata.doi) {
    const scidbExists = data.downloads.scidb.some(d => d.url.includes('doi='));
    if (!scidbExists) {
      data.downloads.scidb.unshift({
        url: `/scidb?doi=${encodeURIComponent(data.metadata.doi)}`,
        text: '🧬 SciDB (recommended)',
      });
    }
  }
  
  return data;
}

function extractPaperMetadata(text, md5) {
  const meta = {
    md5: md5,
    title: '',
    authors: [],
    format: '',
    size: '',
    year: '',
    sources: [],
  };
  
  // Clean text
  text = text.replace(/\s+/g, ' ').trim();
  
  // Extract title - look for text before author pattern or before first comma following a title-like string
  const titlePatterns = [
    /^(.{10,150}?)(?=\s+[A-Z][a-z]+,\s*[A-Z])/,
    /^(.{10,150}?)(?=\s+[A-Z]{2,})/,
    /^([^.]+)/,
  ];
  
  for (const pattern of titlePatterns) {
    const match = text.match(pattern);
    if (match && match[1].length > 10) {
      let candidate = match[1].trim();
      // Clean up common prefixes
      candidate = candidate.replace(/^(scihub|nexusstc|lgli|zlib|upload|ia|hathi)\//i, '');
      candidate = candidate.replace(/^[\d./]+\/\s*/, '');
      candidate = candidate.replace(/\.pdf$/i, '');
      if (candidate.length > 10 && candidate.length < 200) {
        meta.title = candidate;
        break;
      }
    }
  }
  
  // Extract format
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
  
  // Extract sources (🧬, 🚀, zlib, lgli, scihub, etc.)
  const sourcePatterns = [
    /🧬\s*(\w+)/g,
    /🚀\s*(\w+)/g,
    /zlib|lgli|scihub|nexusstc|upload|hathi|ia|duxiu/gi,
  ];
  const sources = new Set();
  sourcePatterns.forEach(pattern => {
    let match;
    while ((match = pattern.exec(text)) !== null) {
      const src = (match[1] || match[0]).toLowerCase();
      if (src.length > 1 && !['mb', 'gb', 'kb'].includes(src)) {
        sources.add(src);
      }
    }
  });
  meta.sources = Array.from(sources);
  
  return meta;
}

console.log(JSON.stringify(result, null, 2));
result;
