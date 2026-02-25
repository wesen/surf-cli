#!/usr/bin/env node
const fs = require('fs');
const path = require('path');

function usage() {
  console.error('Usage: node scripts/diff-go-node-summary.cjs <before-dir> <after-dir> [output-md]');
  process.exit(1);
}

const beforeDir = process.argv[2];
const afterDir = process.argv[3];
const outputPath = process.argv[4] || null;
if (!beforeDir || !afterDir) usage();

function readSummaries(dir) {
  const out = new Map();
  const entries = fs.readdirSync(dir).filter((f) => f.endsWith('.summary.json')).sort();
  for (const f of entries) {
    const full = path.join(dir, f);
    const raw = fs.readFileSync(full, 'utf8');
    const parsed = JSON.parse(raw);
    out.set(parsed.name, parsed);
  }
  return out;
}

function pick(obj, pathKey, fallback = '') {
  const parts = pathKey.split('.');
  let cur = obj;
  for (const p of parts) {
    if (!cur || typeof cur !== 'object' || !(p in cur)) return fallback;
    cur = cur[p];
  }
  return cur;
}

const before = readSummaries(beforeDir);
const after = readSummaries(afterDir);
const names = Array.from(new Set([...before.keys(), ...after.keys()])).sort();

const lines = [];
lines.push('---');
lines.push('Title: Output Shape Diff');
lines.push('Ticket: SURF-20260225-R2');
lines.push('Status: active');
lines.push('Topics:');
lines.push('  - go');
lines.push('  - chromium');
lines.push('  - native-messaging');
lines.push('DocType: reference');
lines.push('Intent: working');
lines.push('Owners: []');
lines.push('RelatedFiles: []');
lines.push('ExternalSources: []');
lines.push('Summary: \"Generated shape-only diff for two comparison runs\"');
lines.push(`LastUpdated: ${new Date().toISOString()}`);
lines.push('WhatFor: \"Track schema changes between baseline and current outputs\"');
lines.push('WhenToUse: \"Use when validating output format improvements\"');
lines.push('---');
lines.push('');
lines.push('# Output Shape Diff');
lines.push('');
lines.push(`- Before: \`${beforeDir}\``);
lines.push(`- After: \`${afterDir}\``);
lines.push('');
lines.push('| Case | Node Shape (before) | Go Shape (before) | Node Shape (after) | Go Shape (after) |');
lines.push('|---|---|---|---|---|');

for (const name of names) {
  const b = before.get(name) || {};
  const a = after.get(name) || {};
  const bNode = pick(b, 'node.stdout_shape', 'n/a');
  const bGo = pick(b, 'go.stdout_shape', 'n/a');
  const aNode = pick(a, 'node.stdout_shape', 'n/a');
  const aGo = pick(a, 'go.stdout_shape', 'n/a');
  lines.push(`| ${name} | ${bNode} | ${bGo} | ${aNode} | ${aGo} |`);
}

lines.push('');
lines.push('## Notes');
lines.push('');
lines.push('- This compares output *shapes* only; inspect raw stdout files for semantic differences.');
lines.push('- If Node shapes changed between runs, runtime state likely changed (tab count, page title, etc.).');

const md = lines.join('\n');
if (outputPath) {
  fs.mkdirSync(path.dirname(outputPath), { recursive: true });
  fs.writeFileSync(outputPath, md);
  console.log(`Wrote: ${outputPath}`);
} else {
  console.log(md);
}
