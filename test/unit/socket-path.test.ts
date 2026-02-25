// @ts-expect-error - CommonJS module without type definitions
import { getSocketPath } from "../../native/socket-path.cjs";

describe("getSocketPath", () => {
  const originalSocketPath = process.env.SURF_SOCKET_PATH;

  afterEach(() => {
    if (originalSocketPath === undefined) {
      delete process.env.SURF_SOCKET_PATH;
    } else {
      process.env.SURF_SOCKET_PATH = originalSocketPath;
    }
  });

  it("returns env override when SURF_SOCKET_PATH is set", () => {
    process.env.SURF_SOCKET_PATH = "/tmp/custom-surf.sock";
    expect(getSocketPath("linux")).toBe("/tmp/custom-surf.sock");
  });

  it("returns windows pipe path on win32", () => {
    delete process.env.SURF_SOCKET_PATH;
    expect(getSocketPath("win32")).toBe("//./pipe/surf");
  });

  it("returns default unix socket path on linux", () => {
    delete process.env.SURF_SOCKET_PATH;
    expect(getSocketPath("linux")).toBe("/tmp/surf.sock");
  });
});
