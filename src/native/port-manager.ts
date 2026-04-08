import { debugLog } from "../utils/debug";
import { handleNativeApiResponse } from "./native-api-transport";

let nativePort: chrome.runtime.Port | null = null;
let messageHandler: ((msg: any) => Promise<any>) | null = null;
let reconnectTimeout: ReturnType<typeof setTimeout> | null = null;
let pendingNativeRequests = new Map<number, { resolve: (value: any) => void; reject: (err: Error) => void }>();
let nativeRequestId = 0;

function summarizeNativePayload(value: any): any {
  if (value == null) return value;
  if (Array.isArray(value)) {
    return { type: "array", length: value.length };
  }
  if (typeof value === "string") {
    return value.length > 160 ? `${value.slice(0, 160)}...` : value;
  }
  if (typeof value !== "object") {
    return value;
  }

  const summary: Record<string, any> = { keys: Object.keys(value) };
  if ("type" in value) summary.type = value.type;
  if ("id" in value) summary.id = value.id;
  if ("tabId" in value) summary.tabId = value.tabId;
  if ("success" in value) summary.success = value.success;
  if ("error" in value) summary.error = value.error;
  if ("result" in value && value.result && typeof value.result === "object") {
    summary.resultKeys = Object.keys(value.result);
    const innerResult = value.result as Record<string, any>;
    if ("value" in innerResult) {
      const innerValue = innerResult.value;
      if (typeof innerValue === "string") {
        summary.resultValuePreview = innerValue.slice(0, 120);
      } else if (Array.isArray(innerValue)) {
        summary.resultValueLength = innerValue.length;
      } else if (innerValue && typeof innerValue === "object") {
        summary.resultValueKeys = Object.keys(innerValue);
        if (typeof innerValue.text === "string") {
          summary.resultTextLength = innerValue.text.length;
          summary.resultTextPreview = innerValue.text.slice(0, 120);
        }
        if ("stopVisible" in innerValue) summary.resultStopVisible = innerValue.stopVisible;
        if ("finished" in innerValue) summary.resultFinished = innerValue.finished;
        if ("messageId" in innerValue) summary.resultMessageId = innerValue.messageId;
      } else {
        summary.resultValue = innerValue;
      }
    }
  }
  if ("cookies" in value && Array.isArray(value.cookies)) {
    summary.cookies = value.cookies.length;
  }
  if ("text" in value && typeof value.text === "string") {
    summary.textPreview = value.text.slice(0, 120);
  }
  if ("value" in value) {
    const inner = value.value;
    if (typeof inner === "string") {
      summary.valuePreview = inner.slice(0, 120);
    } else if (Array.isArray(inner)) {
      summary.valueLength = inner.length;
    } else if (inner && typeof inner === "object") {
      summary.valueKeys = Object.keys(inner);
    } else {
      summary.value = inner;
    }
  }
  return summary;
}

export function initNativeMessaging(
  handler: (msg: any) => Promise<any>
): void {
  messageHandler = handler;
  connect();
}

export function sendToNativeHost(msg: any): Promise<any> {
  return new Promise((resolve, reject) => {
    if (!nativePort) {
      reject(new Error("Native host not connected"));
      return;
    }
    
    if (msg.type === "API_REQUEST") {
      nativePort.postMessage(msg);
      resolve({ sent: true });
      return;
    }
    
    const id = ++nativeRequestId;
    pendingNativeRequests.set(id, { resolve, reject });
    nativePort.postMessage({ ...msg, id });
    
    setTimeout(() => {
      if (pendingNativeRequests.has(id)) {
        pendingNativeRequests.delete(id);
        reject(new Error("Native host request timeout"));
      }
    }, 10000);
  });
}

export function postToNativeHost(msg: any): void {
  if (nativePort) {
    nativePort.postMessage(msg);
  }
}

function connect(): void {
  if (reconnectTimeout) {
    clearTimeout(reconnectTimeout);
    reconnectTimeout = null;
  }

  try {
    nativePort = chrome.runtime.connectNative("surf.browser.host");
    debugLog("Connecting to native host...");

    nativePort.onMessage.addListener(async (msg) => {
      debugLog("Received from native host:", msg.type || msg.id);

      if (msg.type === "HOST_READY") {
        debugLog("Native host ready");
        return;
      }

      if (msg.type?.startsWith("API_RESPONSE_")) {
        handleNativeApiResponse(msg);
        chrome.runtime.sendMessage(msg).catch(() => {});
        return;
      }

      if (msg.id && pendingNativeRequests.has(msg.id)) {
        const { resolve } = pendingNativeRequests.get(msg.id)!;
        pendingNativeRequests.delete(msg.id);
        resolve(msg);
        return;
      }

      if (!messageHandler) return;

      const startedAt = Date.now();
      debugLog("Handling native host request:", {
        id: msg.id,
        type: msg.type,
        tabId: msg.tabId,
      });

      try {
        const result = await messageHandler(msg);
        debugLog("Native host handler resolved:", {
          id: msg.id,
          type: msg.type,
          tookMs: Date.now() - startedAt,
          result: summarizeNativePayload(result),
        });
        if (!nativePort) {
          debugLog("Cannot send response - native host disconnected:", msg.id);
          return;
        }
        const response = { id: msg.id, ...result };
        nativePort.postMessage(response);
        debugLog("Sent response to native host:", {
          id: msg.id,
          type: msg.type,
          response: summarizeNativePayload(response),
        });
      } catch (err) {
        debugLog("Native host handler failed:", {
          id: msg.id,
          type: msg.type,
          tookMs: Date.now() - startedAt,
          error: err instanceof Error ? err.message : String(err),
        });
        if (!nativePort) {
          debugLog("Cannot send error - native host disconnected:", msg.id);
          return;
        }
        const errorResponse = {
          id: msg.id,
          error: err instanceof Error ? err.message : "Unknown error",
        };
        nativePort.postMessage(errorResponse);
        debugLog("Sent error to native host:", errorResponse);
      }
    });

    nativePort.onDisconnect.addListener(() => {
      const error = chrome.runtime.lastError;
      debugLog(
        "Native host disconnected:",
        error?.message || "unknown reason"
      );
      nativePort = null;

      if (!error?.message?.includes("not found")) {
        reconnectTimeout = setTimeout(connect, 5000);
      }
    });
  } catch (err) {
    debugLog("Failed to connect to native host:", err);
    reconnectTimeout = setTimeout(connect, 10000);
  }
}
