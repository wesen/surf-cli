#!/usr/bin/env node
const assert = require("assert");
const fs = require("fs");
const os = require("os");
const path = require("path");
const { execFileSync } = require("child_process");

const repoRoot = path.resolve(__dirname, "..", "..");
const installScript = path.join(repoRoot, "scripts", "install-native-host.cjs");
const uninstallScript = path.join(repoRoot, "scripts", "uninstall-native-host.cjs");
const extensionId = "abcdefghijklmnopabcdefghijklmnop";

const tmpRoot = fs.mkdtempSync(path.join(os.tmpdir(), "surf-installer-smoke-"));
const fakeHome = path.join(tmpRoot, "home");
fs.mkdirSync(fakeHome, { recursive: true });

// Trigger snap-target branch for chromium installs.
fs.mkdirSync(path.join(fakeHome, "snap", "chromium", "common"), { recursive: true });

const env = {
  ...process.env,
  HOME: fakeHome,
  SURF_NODE_PATH: process.execPath,
  SURF_HOST_PATH: path.join(repoRoot, "native", "host.cjs"),
};

const runNode = (script, args) => {
  return execFileSync(process.execPath, [script, ...args], {
    cwd: repoRoot,
    env,
    encoding: "utf8",
  });
};

runNode(installScript, [extensionId, "--browser", "chromium", "--profile", "core-go"]);

const standardManifest = path.join(
  fakeHome,
  ".config",
  "chromium",
  "NativeMessagingHosts",
  "surf.browser.host.json"
);
const snapManifest = path.join(
  fakeHome,
  "snap",
  "chromium",
  "common",
  "chromium",
  "NativeMessagingHosts",
  "surf.browser.host.json"
);

assert.ok(fs.existsSync(standardManifest), "standard chromium manifest missing");
assert.ok(fs.existsSync(snapManifest), "snap chromium manifest missing");

const standardWrapper = path.join(fakeHome, ".local", "share", "surf-cli", "host-wrapper.sh");
const snapWrapper = path.join(fakeHome, "snap", "chromium", "common", "surf-cli", "host-wrapper.sh");
assert.ok(fs.existsSync(standardWrapper), "standard wrapper missing");
assert.ok(fs.existsSync(snapWrapper), "snap wrapper missing");

const standardContent = fs.readFileSync(standardWrapper, "utf8");
assert.ok(standardContent.includes("SURF_HOST_PROFILE"), "wrapper missing host profile logic");
assert.ok(standardContent.includes("SURF_HOST_PROFILE:-core-go"), "wrapper should default to core-go when requested");

runNode(uninstallScript, ["--browser", "chromium", "--all"]);

assert.ok(!fs.existsSync(standardManifest), "standard manifest should be removed");
assert.ok(!fs.existsSync(snapManifest), "snap manifest should be removed");

console.log("native-host installer smoke test passed");
