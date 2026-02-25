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
    darwin: "Library/Application Support/BraveSoftware/Brave-Browser/NativeMessagingHosts",
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

function getWrapperDir() {
  const platform = process.platform;
  const home = os.homedir();
  switch (platform) {
    case "darwin":
      return path.join(home, "Library/Application Support/surf-cli");
    case "linux":
      return path.join(home, ".local/share/surf-cli");
    case "win32":
      return path.join(process.env.LOCALAPPDATA || path.join(home, "AppData/Local"), "surf-cli");
    default:
      return null;
  }
}

function getChromiumSnapRoot() {
  if (process.platform !== "linux") return null;
  const root = path.join(os.homedir(), "snap/chromium/common");
  return fs.existsSync(root) ? root : null;
}

function getManifestPaths(browser) {
  const platform = process.platform;
  const browserConfig = BROWSERS[browser];

  if (!browserConfig || !browserConfig[platform]) {
    return [];
  }

  if (platform === "win32") {
    return [
      `HKCU\\Software\\${browserConfig.win32}\\NativeMessagingHosts\\${HOST_NAME}`,
    ];
  }

  const manifestPaths = [
    path.join(os.homedir(), browserConfig[platform], `${HOST_NAME}.json`),
  ];

  if (platform === "linux" && browser === "chromium") {
    const snapRoot = getChromiumSnapRoot();
    if (snapRoot) {
      manifestPaths.push(
        path.join(snapRoot, "chromium", "NativeMessagingHosts", `${HOST_NAME}.json`)
      );
    }
  }

  return manifestPaths;
}

function removeWindowsRegistry(regPath) {
  try {
    execSync(`reg delete "${regPath}" /f`, { stdio: "pipe" });
    return regPath;
  } catch {
    return null;
  }
}

function removeManifest(browser) {
  const platform = process.platform;
  const manifestPaths = getManifestPaths(browser);
  const removed = [];

  for (const manifestPath of manifestPaths) {
    if (platform === "win32") {
      const result = removeWindowsRegistry(manifestPath);
      if (result) removed.push(result);
      continue;
    }

    try {
      fs.unlinkSync(manifestPath);
      removed.push(manifestPath);
    } catch {}
  }

  return removed;
}

function removeWrapperDirs(allBrowsers) {
  const removed = [];
  const defaultWrapperDir = getWrapperDir();

  if (defaultWrapperDir) {
    try {
      fs.rmSync(defaultWrapperDir, { recursive: true, force: true });
      removed.push(defaultWrapperDir);
    } catch {}
  }

  if (process.platform === "linux" && allBrowsers.includes("chromium")) {
    const snapRoot = getChromiumSnapRoot();
    if (snapRoot) {
      const snapWrapperDir = path.join(snapRoot, "surf-cli");
      try {
        fs.rmSync(snapWrapperDir, { recursive: true, force: true });
        removed.push(snapWrapperDir);
      } catch {}
    }
  }

  return removed;
}

function parseArgs() {
  const args = process.argv.slice(2);
  const result = { browsers: ["chrome"], all: false };

  for (let i = 0; i < args.length; i++) {
    const arg = args[i];
    if (arg === "--browser" || arg === "-b") {
      const browserArg = args[++i];
      if (browserArg === "all") {
        result.browsers = Object.keys(BROWSERS);
        result.all = true;
      } else {
        result.browsers = browserArg.split(",").map((b) => b.trim().toLowerCase());
      }
    } else if (arg === "--all" || arg === "-a") {
      result.browsers = Object.keys(BROWSERS);
      result.all = true;
    } else if (arg === "--help" || arg === "-h") {
      printHelp();
      process.exit(0);
    }
  }
  return result;
}

function printHelp() {
  console.log(`
Surf CLI Native Host Uninstaller

Usage: uninstall-native-host.cjs [options]

Options:
  -b, --browser   Browser(s) to uninstall from (default: chrome)
                  Values: chrome, chromium, brave, edge, arc, helium, all
  -a, --all       Uninstall from all browsers and remove wrapper

Examples:
  node uninstall-native-host.cjs
  node uninstall-native-host.cjs --browser brave
  node uninstall-native-host.cjs --all
`);
}

function main() {
  const { browsers, all } = parseArgs();

  console.log(`Platform: ${process.platform}`);
  console.log("");

  const removed = [];
  const notFound = [];

  for (const browser of browsers) {
    if (!BROWSERS[browser]) {
      console.error(`Unknown browser: ${browser}`);
      continue;
    }

    const removedPaths = removeManifest(browser);
    if (removedPaths.length > 0) {
      for (const p of removedPaths) {
        removed.push({ browser: BROWSERS[browser].name, path: p });
      }
    } else {
      notFound.push(BROWSERS[browser].name);
    }
  }

  if (removed.length > 0) {
    console.log("Removed manifests:");
    for (const { browser, path: p } of removed) {
      console.log(`  ${browser}: ${p}`);
    }
  }

  if (notFound.length > 0) {
    console.log(`\nNot found: ${notFound.join(", ")}`);
  }

  if (all) {
    const removedWrapperDirs = removeWrapperDirs(browsers);
    if (removedWrapperDirs.length > 0) {
      console.log("\nRemoved wrapper directories:");
      for (const dir of removedWrapperDirs) {
        console.log(`  ${dir}`);
      }
    }
  }

  console.log("\nDone!");
}

main();
