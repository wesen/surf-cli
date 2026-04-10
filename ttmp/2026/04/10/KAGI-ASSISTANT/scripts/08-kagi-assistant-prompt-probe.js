const textarea = document.querySelector('textarea[placeholder="Ask Assistant"]');
const form = textarea ? textarea.closest('form') : null;
const promptOptions = document.querySelector('.prompt-options');
const submit = form ? form.querySelector('button[type="submit"], button[aria-label*="Send"], button[title*="Send"]') : null;
return {
  textarea: textarea ? {
    placeholder: textarea.getAttribute('placeholder'),
    className: (textarea.className || '').toString(),
    html: textarea.outerHTML,
  } : null,
  form: form ? {
    className: (form.className || '').toString(),
    html: form.outerHTML.slice(0, 5000),
  } : null,
  promptOptions: promptOptions ? {
    text: (promptOptions.textContent || '').trim().replace(/\s+/g, ' '),
    html: promptOptions.outerHTML.slice(0, 5000),
  } : null,
  submit: submit ? {
    text: (submit.textContent || '').trim().replace(/\s+/g, ' '),
    aria: submit.getAttribute('aria-label'),
    title: submit.getAttribute('title'),
    className: (submit.className || '').toString(),
    html: submit.outerHTML,
  } : null,
};
