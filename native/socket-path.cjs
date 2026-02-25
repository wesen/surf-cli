const DEFAULT_SOCKET_PATH = "/tmp/surf.sock";
const DEFAULT_PIPE_PATH = "//./pipe/surf";

function getSocketPath(platform = process.platform) {
  if (process.env.SURF_SOCKET_PATH) {
    return process.env.SURF_SOCKET_PATH;
  }
  return platform === "win32" ? DEFAULT_PIPE_PATH : DEFAULT_SOCKET_PATH;
}

module.exports = {
  DEFAULT_SOCKET_PATH,
  DEFAULT_PIPE_PATH,
  getSocketPath,
};
