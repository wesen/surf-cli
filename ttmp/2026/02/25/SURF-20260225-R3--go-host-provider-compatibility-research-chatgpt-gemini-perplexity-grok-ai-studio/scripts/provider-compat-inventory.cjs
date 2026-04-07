#!/usr/bin/env node
const fs = require('fs');
const path = require('path');

function resolveRepoRoot(startDir) {
  let cur = startDir;
  for (let i = 0; i < 12; i++) {
    const candidate = path.join(cur, 'native', 'host-helpers.cjs');
    if (fs.existsSync(candidate)) return cur;
    const parent = path.dirname(cur);
    if (parent === cur) break;
    cur = parent;
  }
  throw new Error(`Could not locate repo root from ${startDir}`);
}

const repoRoot = resolveRepoRoot(__dirname);

const hostHelpersPath = path.join(repoRoot, 'native', 'host-helpers.cjs');
const goToolmapPath = path.join(repoRoot, 'go', 'internal', 'host', 'router', 'toolmap.go');
const serviceWorkerPath = path.join(repoRoot, 'src', 'service-worker', 'index.ts');

const outPath = path.join(
  __dirname,
  '..',
  'sources',
  '01-provider-compat-inventory.json'
);

const providerTools = new Set([
  'chatgpt',
  'gemini',
  'perplexity',
  'grok',
  'aistudio',
  'aistudio.build',
]);

function read(p) {
  return fs.readFileSync(p, 'utf8');
}

function extractProviderMappings(hostHelpers) {
  const lines = hostHelpers.split(/\r?\n/);
  const mappings = [];

  for (let i = 0; i < lines.length; i++) {
    const m = lines[i].match(/^\s*case\s+"([^"]+)":\s*(?:\{)?\s*$/);
    if (!m) continue;
    const tool = m[1];
    if (!providerTools.has(tool)) continue;

    const block = [];
    let j = i + 1;
    for (; j < lines.length; j++) {
      if (/^\s*case\s+"[^"]+":\s*(?:\{)?\s*$/.test(lines[j])) break;
      block.push(lines[j]);
      if (/^\s*return\s+\{/.test(lines[j])) {
        // continue scanning object
      }
    }

    const joined = block.join('\n');
    const types = Array.from(joined.matchAll(/\btype\s*:\s*"([A-Z0-9_]+)"/g)).map((m) => m[1]);

    const argKeys = [];
    for (const line of block) {
      const km = line.match(/^\s*([A-Za-z][A-Za-z0-9_-]*)\s*:\s*/);
      if (!km) continue;
      const k = km[1];
      if (k === 'type') continue;
      if (k === 'saveModels') {
        argKeys.push(k);
        continue;
      }
      if (k === 'timeout' || k === 'query' || k === 'model' || k === 'withPage' || k === 'file' || k === 'mode' || k === 'generateImage' || k === 'editImage' || k === 'output' || k === 'youtube' || k === 'aspectRatio' || k === 'deepSearch' || k === 'keepOpen') {
        argKeys.push(k);
      }
    }

    mappings.push({
      tool,
      native_message_type: types[0] || null,
      native_message_types: Array.from(new Set(types)),
      argument_keys: Array.from(new Set(argKeys)),
    });
  }

  return mappings;
}

function extractGoProviderPrefixes(goToolmap) {
  const block = goToolmap.match(/var\s+providerPrefixes\s*=\s*\[]string\s*\{([\s\S]*?)\n\}/m);
  if (!block) return [];
  return Array.from(block[1].matchAll(/"([^"]+)"/g)).map((m) => m[1]);
}

function extractServiceWorkerProviderCases(serviceWorker) {
  const caseMatches = Array.from(serviceWorker.matchAll(/case\s+"([A-Z0-9_]+)"\s*:\s*\{/g)).map((m) => m[1]);
  return caseMatches.filter((c) =>
    c.includes('CHATGPT') ||
    c.includes('PERPLEXITY') ||
    c.includes('GROK') ||
    c.includes('AISTUDIO') ||
    c === 'GET_GOOGLE_COOKIES' ||
    c === 'GET_TWITTER_COOKIES' ||
    c === 'GET_CHATGPT_COOKIES' ||
    c === 'DOWNLOADS_SEARCH'
  );
}

const hostHelpers = read(hostHelpersPath);
const goToolmap = read(goToolmapPath);
const serviceWorker = read(serviceWorkerPath);

const inventory = {
  generated_at: new Date().toISOString(),
  sources: {
    host_helpers: hostHelpersPath,
    go_toolmap: goToolmapPath,
    service_worker: serviceWorkerPath,
  },
  node_provider_tool_mappings: extractProviderMappings(hostHelpers),
  go_provider_prefixes_blocked_in_core: extractGoProviderPrefixes(goToolmap),
  service_worker_provider_message_handlers: extractServiceWorkerProviderCases(serviceWorker),
};

fs.mkdirSync(path.dirname(outPath), { recursive: true });
fs.writeFileSync(outPath, JSON.stringify(inventory, null, 2));
console.log(`Wrote ${outPath}`);
