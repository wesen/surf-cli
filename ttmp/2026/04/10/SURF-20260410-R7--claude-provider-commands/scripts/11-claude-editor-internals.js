const input = document.querySelector('[data-testid="chat-input"]');
if (!input) return { ok: false, error: 'chat input not found', href: location.href };
const ownKeys = Reflect.ownKeys(input).map(String).filter((k) => /react|pm|prose|view|editor|fiber/i.test(k));
const winKeys = Reflect.ownKeys(window).map(String).filter((k) => /prose|pm|editor|slate|lexical/i.test(k)).slice(0, 100);
const interesting = {};
for (const k of ownKeys) {
  try {
    const v = input[k];
    interesting[k] = {
      type: typeof v,
      ctor: v && v.constructor ? v.constructor.name : null,
      keys: v && typeof v === 'object' ? Object.keys(v).slice(0, 20) : null,
      str: v && typeof v !== 'object' ? String(v).slice(0, 200) : null,
    };
  } catch (e) {
    interesting[k] = { error: String(e) };
  }
}
return {
  ok: true,
  href: location.href,
  ownKeys,
  winKeys,
  textContent: input.textContent,
  innerHTML: input.innerHTML,
  interesting,
};
