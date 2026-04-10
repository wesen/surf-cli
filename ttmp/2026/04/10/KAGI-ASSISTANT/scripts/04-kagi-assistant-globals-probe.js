const keys = Object.keys(window).filter((k) => /assistant|model|kagi|llm|provider|plan|thread|chat/i.test(k)).sort();
const summary = {};
for (const key of keys) {
  let value;
  try {
    value = window[key];
  } catch (err) {
    value = `[error:${err && err.message ? err.message : String(err)}]`;
  }
  const type = Array.isArray(value) ? 'array' : typeof value;
  let sample = null;
  if (Array.isArray(value)) {
    sample = value.slice(0, 20);
  } else if (value && type === 'object') {
    sample = Object.keys(value).slice(0, 40);
  } else if (type !== 'function') {
    sample = value;
  }
  summary[key] = { type, sample };
}
return summary;
