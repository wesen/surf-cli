const input = document.querySelector('[data-testid="chat-input"]');
if (!input || !input.editor) return { ok: false, error: 'editor not found', href: location.href };
const editor = input.editor;
const commandKeys = editor.commandManager ? Object.keys(editor.commandManager.commands || {}).slice(0, 200) : [];
const directMethods = Object.getOwnPropertyNames(Object.getPrototypeOf(editor)).filter((k) => typeof editor[k] === 'function').slice(0, 200);
const chainMethods = editor.chain ? Object.keys(editor.chain()).slice(0, 100) : [];
return {
  ok: true,
  href: location.href,
  textContent: input.textContent || '',
  isInitialized: editor.isInitialized,
  isFocused: editor.isFocused,
  commandKeys,
  directMethods,
  chainMethods,
};
