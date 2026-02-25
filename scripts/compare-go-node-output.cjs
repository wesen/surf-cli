#!/usr/bin/env node
const fs = require('fs');
const path = require('path');
const { spawnSync } = require('child_process');

const repoRoot = path.resolve(__dirname, '..');
const goRoot = path.join(repoRoot, 'go');
const socketPath = process.env.SURF_SOCKET_PATH || path.join(process.env.HOME || '', 'snap/chromium/common/surf-cli/surf.sock');
const stamp = new Date().toISOString().replace(/[:.]/g, '-');
const outputRoot = process.argv[2] || path.join(repoRoot, 'ttmp/2026/02/25/SURF-20260225-R2--go-native-host-lite-core-browser-only-implementation-plan-with-glazed-commands/sources/output-compare', stamp);

fs.mkdirSync(outputRoot, { recursive: true });

const baseEnv = { ...process.env, SURF_SOCKET_PATH: socketPath };

function runCmd(command, args, cwd) {
  const result = spawnSync(command, args, {
    cwd,
    env: baseEnv,
    encoding: 'utf8',
    timeout: 30000,
  });

  return {
    ok: result.status === 0,
    status: result.status,
    stdout: result.stdout || '',
    stderr: result.stderr || '',
    signal: result.signal || null,
    timedOut: !!result.error && result.error.code === 'ETIMEDOUT',
    error: result.error ? String(result.error.message || result.error) : null,
  };
}

function tryParseJSON(s) {
  const trimmed = (s || '').trim();
  if (!trimmed) return null;
  try {
    return JSON.parse(trimmed);
  } catch {
    return null;
  }
}

function shape(v) {
  if (Array.isArray(v)) {
    if (v.length === 0) return 'array(0)';
    return `array(${v.length})<${shape(v[0])}>`;
  }
  if (v && typeof v === 'object') {
    return `object{${Object.keys(v).slice(0, 8).join(',')}}`;
  }
  return typeof v;
}

function savePair(name, nodeResult, goResult) {
  fs.writeFileSync(path.join(outputRoot, `${name}.node.stdout.txt`), nodeResult.stdout);
  fs.writeFileSync(path.join(outputRoot, `${name}.node.stderr.txt`), nodeResult.stderr);
  fs.writeFileSync(path.join(outputRoot, `${name}.go.stdout.txt`), goResult.stdout);
  fs.writeFileSync(path.join(outputRoot, `${name}.go.stderr.txt`), goResult.stderr);

  const parsed = {
    node: tryParseJSON(nodeResult.stdout),
    go: tryParseJSON(goResult.stdout),
  };

  const summary = {
    name,
    node: {
      ok: nodeResult.ok,
      status: nodeResult.status,
      stdout_shape: shape(parsed.node),
      stderr_nonempty: !!nodeResult.stderr.trim(),
    },
    go: {
      ok: goResult.ok,
      status: goResult.status,
      stdout_shape: shape(parsed.go),
      stderr_nonempty: !!goResult.stderr.trim(),
    },
  };
  fs.writeFileSync(path.join(outputRoot, `${name}.summary.json`), JSON.stringify(summary, null, 2));

  return summary;
}

// Ensure an active tab exists for page/network/console commands.
const setup = runCmd('node', ['native/cli.cjs', 'tab.new', 'https://example.com'], repoRoot);
fs.writeFileSync(path.join(outputRoot, '00-setup-tab-new.stdout.txt'), setup.stdout);
fs.writeFileSync(path.join(outputRoot, '00-setup-tab-new.stderr.txt'), setup.stderr);

const cases = [
  {
    name: 'tab-list',
    node: ['node', ['native/cli.cjs', 'tab.list', '--json'], repoRoot],
    go: ['go', ['run', './cmd/surf-go', 'tab', 'list', '--output', 'json'], goRoot],
  },
  {
    name: 'page-read',
    node: ['node', ['native/cli.cjs', 'page.read', '--json'], repoRoot],
    go: ['go', ['run', './cmd/surf-go', 'page', 'read', '--output', 'json'], goRoot],
  },
  {
    name: 'page-text',
    node: ['node', ['native/cli.cjs', 'page.text', '--json'], repoRoot],
    go: ['go', ['run', './cmd/surf-go', 'page', 'text', '--output', 'json'], goRoot],
  },
  {
    name: 'page-state',
    node: ['node', ['native/cli.cjs', 'page.state', '--json'], repoRoot],
    go: ['go', ['run', './cmd/surf-go', 'page', 'state', '--output', 'json'], goRoot],
  },
  {
    name: 'network-list',
    node: ['node', ['native/cli.cjs', 'network', '--limit', '3', '--json'], repoRoot],
    go: ['go', ['run', './cmd/surf-go', 'network', 'list', '--args-json', '{"limit":3}', '--output', 'json'], goRoot],
  },
  {
    name: 'console-read',
    node: ['node', ['native/cli.cjs', 'console', '--limit', '3', '--json'], repoRoot],
    go: ['go', ['run', './cmd/surf-go', 'console', 'read', '--args-json', '{"limit":3}', '--output', 'json'], goRoot],
  },
  {
    name: 'navigate',
    node: ['node', ['native/cli.cjs', 'navigate', 'https://example.org', '--json'], repoRoot],
    go: ['go', ['run', './cmd/surf-go', 'navigate', '--url', 'https://example.org', '--output', 'json'], goRoot],
  },
];

const summaries = [];
for (const c of cases) {
  const nodeResult = runCmd(c.node[0], c.node[1], c.node[2]);
  const goResult = runCmd(c.go[0], c.go[1], c.go[2]);
  summaries.push(savePair(c.name, nodeResult, goResult));
}

const reportPath = path.join(outputRoot, 'SUMMARY.md');
let md = `# Node vs Go Output Comparison\n\n`;
md += `- Socket: \`${socketPath}\`\n`;
md += `- Generated: ${new Date().toISOString()}\n\n`;
md += `## Setup\n\n`;
md += `- tab.new stdout: \`${setup.stdout.trim() || '(empty)'}\`\n`;
if (setup.stderr.trim()) {
  md += `- tab.new stderr: \`${setup.stderr.trim().replace(/`/g, '\\`')}\`\n`;
}
md += `\n## Cases\n\n`;
for (const s of summaries) {
  md += `### ${s.name}\n`;
  md += `- Node: ok=${s.node.ok} status=${s.node.status} shape=${s.node.stdout_shape}\n`;
  md += `- Go: ok=${s.go.ok} status=${s.go.status} shape=${s.go.stdout_shape}\n`;
}
fs.writeFileSync(reportPath, md);

console.log(`Wrote comparison artifacts to: ${outputRoot}`);
console.log(`Summary: ${reportPath}`);
