return {
  kagiSettingsAssistant: window.kagiSettings && window.kagiSettings.assistant,
  translationKeys: window.translations && window.translations.modelSelect ? Object.keys(window.translations.modelSelect) : null,
  translationSample: window.translations && window.translations.modelSelect ? window.translations.modelSelect : null,
  bodyScriptHits: Array.from(document.scripts).map((s, i) => ({ i, text: s.textContent || '' })).filter((x) => /assistant|modelList|custom assistant|Kagi Research|ChatGPT|Claude|Gemini/i.test(x.text)).slice(0,4).map((x) => ({ i: x.i, excerpt: x.text.slice(0, 4000) })),
};
