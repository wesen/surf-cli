// @ts-expect-error - CommonJS module without type definitions
import * as chatgptClient from "../../native/chatgpt-client.cjs";

type CdpEvalResult = { result: { value: unknown } };

const mockEvaluate = async (_tabId: number, expression: string): Promise<CdpEvalResult> => {
  switch (true) {
    case expression.includes("document.readyState"):
      return { result: { value: "complete" } };
    case expression.includes("document.title.toLowerCase"):
      return { result: { value: "chatgpt" } };
    case expression.includes("challenge-platform"):
      return { result: { value: false } };
    case expression.includes("backend-api/me"):
      return { result: { value: { status: 200, hasLoginCta: false } } };
    case expression.includes("!node.hasAttribute('disabled')"):
      return { result: { value: true } };
    case expression.includes("data-surf-file-input-id"):
      return { result: { value: `[data-surf-file-input-id="surf-upload-1"]` } };
    case expression.includes("text.trim().length > 0"):
      return { result: { value: true } };
    case expression.includes("return 'clicked'"):
      return { result: { value: "clicked" } };
    case expression.includes("lastAssistantTurn"):
      return {
        result: {
          value: {
            text: "hello from chatgpt",
            stopVisible: false,
            finished: true,
            messageId: "msg-1",
          },
        },
      };
    default:
      return { result: { value: true } };
  }
};

describe("chatgpt-client", () => {
  it("uploads file(s) before sending prompt", async () => {
    let uploaded: { tabId: number; selector: string; files: string[] } | null = null;
    let insertedPrompt = "";
    let closedTabId = -1;

    const result = await chatgptClient.query({
      prompt: "review attachment",
      file: "/tmp/a.txt, /tmp/b.txt",
      timeout: 30000,
      getCookies: async () => ({
        cookies: [{ name: "__Secure-next-auth.session-token", value: "abc" }],
      }),
      createTab: async () => ({ tabId: 77 }),
      closeTab: async (tabId: number) => {
        closedTabId = tabId;
      },
      cdpEvaluate: mockEvaluate,
      cdpCommand: async (_tabId: number, method: string, params: { text?: string }) => {
        if (method === "Input.insertText") {
          insertedPrompt = params.text || "";
        }
        return {};
      },
      uploadFile: async (tabId: number, selector: string, files: string[]) => {
        uploaded = { tabId, selector, files };
        return { success: true, filesSet: files.length };
      },
      log: () => {},
    });

    expect(uploaded).toEqual({
      tabId: 77,
      selector: `[data-surf-file-input-id="surf-upload-1"]`,
      files: ["/tmp/a.txt", "/tmp/b.txt"],
    });
    expect(insertedPrompt).toBe("review attachment");
    expect(closedTabId).toBe(77);
    expect(result.response).toBe("hello from chatgpt");
  });

  it("returns upload failure errors", async () => {
    await expect(
      chatgptClient.query({
        prompt: "review attachment",
        file: "/tmp/missing.txt",
        timeout: 30000,
        getCookies: async () => ({
          cookies: [{ name: "__Secure-next-auth.session-token", value: "abc" }],
        }),
        createTab: async () => ({ tabId: 91 }),
        closeTab: async () => {},
        cdpEvaluate: mockEvaluate,
        cdpCommand: async () => ({}),
        uploadFile: async () => ({ error: "Could not find element by selector" }),
        log: () => {},
      }),
    ).rejects.toThrow("Could not find element by selector");
  });

  it("validates required cookie and exports", () => {
    expect(chatgptClient.hasRequiredCookies([{ name: "__Secure-next-auth.session-token", value: "x" }])).toBe(
      true,
    );
    expect(chatgptClient.hasRequiredCookies([{ name: "other", value: "x" }])).toBe(false);
    expect(chatgptClient.listModels).toBeInstanceOf(Function);
    expect(chatgptClient.CHATGPT_URL).toBe("https://chatgpt.com/");
  });

  it("lists available models", async () => {
    const cdpEvaluate = async (_tabId: number, expression: string): Promise<CdpEvalResult> => {
      switch (true) {
        case expression.includes("document.readyState"):
          return { result: { value: "complete" } };
        case expression.includes("document.title.toLowerCase"):
          return { result: { value: "chatgpt" } };
        case expression.includes("challenge-platform"):
          return { result: { value: false } };
        case expression.includes("backend-api/me"):
          return { result: { value: { status: 200, hasLoginCta: false } } };
        case expression.includes("!node.hasAttribute('disabled')"):
          return { result: { value: true } };
        case expression.includes("model-switcher-dropdown-button"):
          return { result: { value: true } };
        case expression.includes("const models = []"):
          return {
            result: {
              value: {
                found: true,
                models: ["GPT-4o", "o1"],
                selected: "GPT-4o",
              },
            },
          };
        default:
          return { result: { value: true } };
      }
    };

    const result = await chatgptClient.listModels({
      timeout: 30000,
      getCookies: async () => ({
        cookies: [{ name: "__Secure-next-auth.session-token", value: "abc" }],
      }),
      createTab: async () => ({ tabId: 77 }),
      closeTab: async () => {},
      cdpEvaluate,
      log: () => {},
    });

    expect(result.models).toEqual(["GPT-4o", "o1"]);
    expect(result.selected).toBe("GPT-4o");
  });
});
