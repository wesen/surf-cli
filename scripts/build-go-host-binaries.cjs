#!/usr/bin/env node
const fs = require("fs");
const path = require("path");
const { execSync } = require("child_process");

const repoRoot = path.resolve(__dirname, "..");
const goRoot = path.join(repoRoot, "go");
const goMain = path.join(goRoot, "cmd", "surf-host-go", "main.go");

if (!fs.existsSync(goMain)) {
  console.error("Error: go/cmd/surf-host-go/main.go not found");
  process.exit(1);
}

function getGoArch() {
  try {
    return execSync("go env GOARCH", { cwd: goRoot, encoding: "utf8" }).trim();
  } catch (e) {
    console.error(`Error: failed to detect GOARCH: ${e.message}`);
    process.exit(1);
  }
}

const goarch = process.env.GOARCH || getGoArch();
const targets = [
  { goos: "linux", ext: "" },
  { goos: "darwin", ext: "" },
  { goos: "windows", ext: ".exe" },
];

const built = [];
for (const t of targets) {
  const outDir = path.join(repoRoot, "dist", "go", `${t.goos}-${goarch}`);
  fs.mkdirSync(outDir, { recursive: true });
  const out = path.join(outDir, `surf-host-go${t.ext}`);

  const cmd = `go build -o "${out}" ./cmd/surf-host-go`;
  try {
    execSync(cmd, {
      cwd: goRoot,
      stdio: "inherit",
      env: {
        ...process.env,
        CGO_ENABLED: "0",
        GOOS: t.goos,
        GOARCH: goarch,
      },
    });
    if (t.ext === "") {
      try {
        fs.chmodSync(out, 0o755);
      } catch {}
    }
    built.push(out);
  } catch (e) {
    console.error(`Error building ${t.goos}/${goarch}: ${e.message}`);
    process.exit(1);
  }
}

console.log("Built Go host binaries:");
for (const p of built) {
  console.log(`  ${p}`);
}
