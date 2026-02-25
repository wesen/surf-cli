#!/usr/bin/env node
const fs = require("fs");
const path = require("path");
const os = require("os");
const { execSync } = require("child_process");

const HOST_NAME = "surf.browser.host";

const BROWSERS = {
  chrome: {
    name: "Google Chrome",
    darwin: "Library/Application Support/Google/Chrome/NativeMessagingHosts",
    linux: ".config/google-chrome/NativeMessagingHosts",
    win32: "Google\\Chrome",
  },
  chromium: {
    name: "Chromium",
    darwin: "Library/Application Support/Chromium/NativeMessagingHosts",
    linux: ".config/chromium/NativeMessagingHosts",
    win32: "Chromium",
  },
  brave: {
    name: "Brave",
    darwin:
      "Library/Application Support/BraveSoftware/Brave-Browser/NativeMessagingHosts",
    linux: ".config/BraveSoftware/Brave-Browser/NativeMessagingHosts",
    win32: "BraveSoftware\\Brave-Browser",
  },
  edge: {
    name: "Microsoft Edge",
    darwin: "Library/Application Support/Microsoft Edge/NativeMessagingHosts",
    linux: ".config/microsoft-edge/NativeMessagingHosts",
    win32: "Microsoft\\Edge",
  },
  arc: {
    name: "Arc",
    darwin: "Library/Application Support/Arc/User Data/NativeMessagingHosts",
    linux: null,
    win32: null,
  },
  helium: {
    name: "Helium",
    darwin: "Library/Application Support/net.imput.helium/NativeMessagingHosts",
    linux: null,
    win32: null,
  },
};

const NODE_PATHS = {
  darwin: ["/opt/homebrew/bin/node", "/usr/local/bin/node", "/usr/bin/node"],
  linux: ["/usr/bin/node", "/usr/local/bin/node"],
  win32: [
    "C:\\Program Files\\nodejs\\node.exe",
    "C:\\Program Files (x86)\\nodejs\\node.exe",
  ],
};

function findNode() {
  if (process.env.SURF_NODE_PATH && fs.existsSync(process.env.SURF_NODE_PATH)) {
    return process.env.SURF_NODE_PATH;
  }
  const platform = process.platform;
  const paths = NODE_PATHS[platform] || [];
  for (const p of paths) {
    if (fs.existsSync(p)) return p;
  }
  try {
    const which = platform === "win32" ? "where" : "which";
    const result = execSync(`${which} node`, { encoding: "utf8" }).trim();
    if (result) return result.split("\n")[0];
  } catch {}
  return null;
}

function findNpmGlobalRoot() {
  try {
    return execSync("npm root -g", { encoding: "utf8" }).trim();
  } catch {
    return null;
  }
}

function getWrapperDir() {
  const platform = process.platform;
  const home = os.homedir();
  switch (platform) {
    case "darwin":
      return path.join(home, "Library/Application Support/surf-cli");
    case "linux":
      return path.join(home, ".local/share/surf-cli");
    case "win32":
      return path.join(
        process.env.LOCALAPPDATA || path.join(home, "AppData/Local"),
        "surf-cli"
      );
    default:
      return null;
  }
}

function getHostPath() {
  if (process.env.SURF_HOST_PATH && fs.existsSync(process.env.SURF_HOST_PATH)) {
    return process.env.SURF_HOST_PATH;
  }
  const npmRoot = findNpmGlobalRoot();
  if (npmRoot) {
    const globalPath = path.join(npmRoot, "surf-cli/native/host.cjs");
    if (fs.existsSync(globalPath)) return globalPath;
  }
  const localPath = path.resolve(__dirname, "../native/host.cjs");
  if (fs.existsSync(localPath)) return localPath;
  return null;
}

function getChromiumSnapRoot() {
  if (process.platform !== "linux") return null;
  const root = path.join(os.homedir(), "snap/chromium/common");
  return fs.existsSync(root) ? root : null;
}

function detectHostPackageRoot(hostPath) {
  const nativeDir = path.dirname(hostPath);
  const packageRoot = path.dirname(nativeDir);
  if (
    fs.existsSync(path.join(packageRoot, "package.json")) &&
    fs.existsSync(path.join(packageRoot, "native", "host.cjs"))
  ) {
    return packageRoot;
  }
  return null;
}

function prepareSnapRuntime(wrapperDir, nodePath, hostPath) {
  const packageRoot = detectHostPackageRoot(hostPath);
  if (!packageRoot) {
    throw new Error(
      `Could not determine Surf package root from host path: ${hostPath}. Set SURF_HOST_PATH to a standard surf-cli native/host.cjs path.`
    );
  }

  const runtimeRoot = path.join(wrapperDir, "runtime");
  const runtimePackageRoot = path.join(runtimeRoot, "surf-cli");
  const runtimeNodePath = path.join(wrapperDir, "node");
  const runtimeHostPath = path.join(runtimePackageRoot, "native", "host.cjs");
  const runtimeSocketPath = path.join(wrapperDir, "surf.sock");

  fs.mkdirSync(wrapperDir, { recursive: true });
  fs.rmSync(runtimeRoot, { recursive: true, force: true });
  fs.mkdirSync(runtimeRoot, { recursive: true });

  fs.cpSync(packageRoot, runtimePackageRoot, { recursive: true });
  try {
    fs.copyFileSync(nodePath, runtimeNodePath);
    fs.chmodSync(runtimeNodePath, "755");
  } catch (e) {
    // If a previous snap wrapper is actively executing this node binary, keep it.
    if (!(e && e.code === "ETXTBSY" && fs.existsSync(runtimeNodePath))) {
      throw e;
    }
  }

  return {
    nodePath: runtimeNodePath,
    hostPath: runtimeHostPath,
    socketPath: runtimeSocketPath,
  };
}

function createWrapper(wrapperDir, nodePath, hostPath, options = {}) {
  const platform = process.platform;
  fs.mkdirSync(wrapperDir, { recursive: true });

  if (platform === "win32") {
    const batPath = path.join(wrapperDir, "host-wrapper.bat");
    const envLine = options.socketPath
      ? `set SURF_SOCKET_PATH=${options.socketPath}\r\n`
      : "";
    const content = `@echo off\r\n${envLine}"${nodePath}" "${hostPath}"\r\n`;
    fs.writeFileSync(batPath, content);
    return batPath;
  }

  const shPath = path.join(wrapperDir, "host-wrapper.sh");
  const hostDir = path.dirname(hostPath);
  const envLine = options.socketPath
    ? `export SURF_SOCKET_PATH=\"${options.socketPath}\"\n`
    : "";
  const content = `#!/bin/bash
${envLine}cd "${hostDir}"
exec "${nodePath}" "${hostPath}"
`;
  fs.writeFileSync(shPath, content);
  fs.chmodSync(shPath, "755");
  return shPath;
}

function writeManifest(manifestDir, extensionId, wrapperPath) {
  fs.mkdirSync(manifestDir, { recursive: true });

  const manifest = {
    name: HOST_NAME,
    description: "Surf CLI Native Host",
    path: wrapperPath,
    type: "stdio",
    allowed_origins: [`chrome-extension://${extensionId}/`],
  };

  const manifestPath = path.join(manifestDir, `${HOST_NAME}.json`);
  fs.writeFileSync(manifestPath, JSON.stringify(manifest, null, 2));
  return manifestPath;
}

function installWindowsRegistry(browser, extensionId, wrapperPath) {
  const browserConfig = BROWSERS[browser];
  const regPath = `HKCU\\Software\\${browserConfig.win32}\\NativeMessagingHosts\\${HOST_NAME}`;

  const manifestDir = getWrapperDir();
  const manifestPath = path.join(manifestDir, `${HOST_NAME}.json`);

  const manifest = {
    name: HOST_NAME,
    description: "Surf CLI Native Host",
    path: wrapperPath,
    type: "stdio",
    allowed_origins: [`chrome-extension://${extensionId}/`],
  };

  fs.writeFileSync(manifestPath, JSON.stringify(manifest, null, 2));

  try {
    execSync(`reg add "${regPath}" /ve /t REG_SZ /d "${manifestPath}" /f`, {
      stdio: "pipe",
    });
    return manifestPath;
  } catch (e) {
    console.error(`Failed to add registry entry: ${e.message}`);
    return null;
  }
}

function installForBrowser(browser, extensionId, nodePath, hostPath) {
  const platform = process.platform;
  const browserConfig = BROWSERS[browser];

  if (!browserConfig || !browserConfig[platform]) {
    return { installed: [], skipped: [BROWSERS[browser]?.name || browser], hints: [] };
  }

  if (platform === "win32") {
    const wrapperPath = createWrapper(getWrapperDir(), nodePath, hostPath);
    const manifestPath = installWindowsRegistry(browser, extensionId, wrapperPath);
    if (!manifestPath) return { installed: [], skipped: [browserConfig.name], hints: [] };
    return {
      installed: [{ browser: browserConfig.name, path: manifestPath }],
      skipped: [],
      hints: [],
    };
  }

  const installed = [];
  const skipped = [];
  const hints = [];

  const standardWrapperDir = getWrapperDir();
  const standardManifestDir = path.join(os.homedir(), browserConfig[platform]);

  if (browser === "chromium" && platform === "linux") {
    try {
      const wrapperPath = createWrapper(standardWrapperDir, nodePath, hostPath);
      const manifestPath = writeManifest(standardManifestDir, extensionId, wrapperPath);
      installed.push({ browser: browserConfig.name, path: manifestPath });
    } catch (e) {
      skipped.push(browserConfig.name);
      hints.push(`Failed standard Chromium install target: ${e.message}`);
    }

    const snapRoot = getChromiumSnapRoot();
    if (snapRoot) {
      const snapWrapperDir = path.join(snapRoot, "surf-cli");
      const snapManifestDir = path.join(snapRoot, "chromium", "NativeMessagingHosts");
      try {
        const snapRuntime = prepareSnapRuntime(snapWrapperDir, nodePath, hostPath);
        const snapWrapperPath = createWrapper(
          snapWrapperDir,
          snapRuntime.nodePath,
          snapRuntime.hostPath,
          { socketPath: snapRuntime.socketPath }
        );
        const snapManifestPath = writeManifest(snapManifestDir, extensionId, snapWrapperPath);
        installed.push({ browser: `${browserConfig.name} (snap)`, path: snapManifestPath });
        hints.push(
          `Snap target installed. Set SURF_SOCKET_PATH=${snapRuntime.socketPath} in your shell when using surf CLI outside snap.`
        );
      } catch (e) {
        skipped.push(`${browserConfig.name} (snap)`);
        hints.push(`Failed snap Chromium install target: ${e.message}`);
      }
    }

    return { installed, skipped, hints };
  }

  try {
    const wrapperPath = createWrapper(standardWrapperDir, nodePath, hostPath);
    const manifestPath = writeManifest(standardManifestDir, extensionId, wrapperPath);
    installed.push({ browser: browserConfig.name, path: manifestPath });
  } catch (e) {
    skipped.push(browserConfig.name);
    hints.push(`Failed ${browserConfig.name} install target: ${e.message}`);
  }

  return { installed, skipped, hints };
}

function parseArgs() {
  const args = process.argv.slice(2);
  const result = { extensionId: null, browsers: ["chrome"] };

  for (let i = 0; i < args.length; i++) {
    const arg = args[i];
    if (arg === "--browser" || arg === "-b") {
      const browserArg = args[++i];
      if (browserArg === "all") {
        result.browsers = Object.keys(BROWSERS);
      } else {
        result.browsers = browserArg.split(",").map((b) => b.trim().toLowerCase());
      }
    } else if (arg === "--help" || arg === "-h") {
      printHelp();
      process.exit(0);
    } else if (!arg.startsWith("-")) {
      result.extensionId = arg;
    }
  }
  return result;
}

function printHelp() {
  console.log(`
Surf CLI Native Host Installer

Usage: install-native-host.cjs <extension-id> [options]

Arguments:
  extension-id    Chrome extension ID (32 lowercase letters a-p)
                  Find at chrome://extensions with Developer Mode enabled

Options:
  -b, --browser   Browser(s) to install for (default: chrome)
                  Values: chrome, chromium, brave, edge, arc, helium, all
                  Multiple: --browser chrome,brave

Examples:
  node install-native-host.cjs abcdefghijklmnopabcdefghijklmnop
  node install-native-host.cjs abcdefghijklmnop --browser brave
  node install-native-host.cjs abcdefghijklmnop --browser all
`);
}

function main() {
  const { extensionId, browsers } = parseArgs();

  if (!extensionId) {
    console.error("Error: Extension ID required");
    console.error(
      "Usage: install-native-host.cjs <extension-id> [--browser chrome|chromium|brave|edge|arc|helium|all]"
    );
    console.error("\nFind your extension ID at chrome://extensions (enable Developer Mode)");
    process.exit(1);
  }

  if (!/^[a-p]{32}$/.test(extensionId)) {
    console.error("Error: Invalid extension ID format");
    console.error("Expected 32 lowercase letters (a-p)");
    process.exit(1);
  }

  const nodePath = findNode();
  if (!nodePath) {
    console.error("Error: Could not find Node.js");
    console.error("Make sure Node.js is installed and in your PATH");
    process.exit(1);
  }

  const hostPath = getHostPath();
  if (!hostPath) {
    console.error("Error: Could not find host.cjs");
    console.error("Make sure surf-cli is installed correctly");
    process.exit(1);
  }

  const wrapperDir = getWrapperDir();
  if (!wrapperDir) {
    console.error("Error: Unsupported platform");
    process.exit(1);
  }

  const snapRoot = getChromiumSnapRoot();

  console.log(`Platform: ${process.platform}`);
  console.log(`Node: ${nodePath}`);
  console.log(`Host: ${hostPath}`);
  console.log(`Wrapper dir: ${wrapperDir}`);
  if (snapRoot) {
    console.log(`Chromium snap root detected: ${snapRoot}`);
  }
  console.log("");

  const installed = [];
  const skipped = [];
  const hints = [];

  for (const browser of browsers) {
    if (!BROWSERS[browser]) {
      console.error(`Unknown browser: ${browser}`);
      continue;
    }

    const result = installForBrowser(browser, extensionId, nodePath, hostPath);
    installed.push(...result.installed);
    skipped.push(...result.skipped);
    hints.push(...result.hints);
  }

  if (installed.length > 0) {
    console.log("Installed for:");
    for (const { browser, path: p } of installed) {
      console.log(`  ${browser}: ${p}`);
    }
  }

  if (skipped.length > 0) {
    console.log(`\nSkipped (failed/unsupported on ${process.platform}): ${skipped.join(", ")}`);
  }

  if (hints.length > 0) {
    console.log("\nHints:");
    for (const hint of hints) {
      console.log(`  - ${hint}`);
    }
  }

  console.log("\nDone! Restart your browser for changes to take effect.");
}

main();
