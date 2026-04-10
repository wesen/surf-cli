const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

const options = typeof SURF_OPTIONS === 'object' && SURF_OPTIONS !== null ? SURF_OPTIONS : {};
const targetModel = options.model || '';
const targetAssistant = options.assistant || '';
const targetSection = options.section || '';

function normalizeText(value) {
  return (value || '').replace(/\s+/g, ' ').trim();
}

async function waitFor(fn, timeoutMs = 5000, intervalMs = 100) {
  const started = Date.now();
  while (Date.now() - started < timeoutMs) {
    const value = fn();
    if (value) {
      return { value, waitedMs: Date.now() - started };
    }
    await sleep(intervalMs);
  }
  return null;
}

function getDialog() {
  return document.querySelector('dialog.promptOptionsSelector[open]');
}

function getProfileButton() {
  return document.querySelector('#profile-select');
}

function summarizeButton(el) {
  if (!el) return null;
  return {
    id: el.id || null,
    ariaLabel: el.getAttribute('aria-label'),
    text: normalizeText(el.textContent || ''),
    title: el.getAttribute('title'),
    ariaExpanded: el.getAttribute('aria-expanded'),
  };
}

function openEntries() {
  const dialog = getDialog();
  if (!dialog) return [];
  return Array.from(dialog.querySelectorAll('li.option[role="option"]')).map((el) => {
    const icon = el.querySelector('svg.model-icon[data-model]');
    const text = normalizeText(el.textContent || '');
    const heading = normalizeText(el.closest('[role="listbox"]')?.getAttribute('aria-label') || el.closest('[role="listbox"]')?.previousElementSibling?.textContent || '');
    return {
      text,
      modelId: icon?.getAttribute('data-model') || null,
      selected: el.classList.contains('selected') || el.getAttribute('aria-selected') === 'true',
      section: heading || null,
    };
  });
}

function radioSections() {
  const dialog = getDialog();
  if (!dialog) return [];
  return Array.from(dialog.querySelectorAll('input[type="radio"]')).map((input) => ({
    value: input.value || null,
    checked: !!input.checked,
    label: normalizeText(input.closest('label')?.textContent || ''),
  }));
}

async function ensureDialogOpen() {
  const button = getProfileButton();
  if (!button) throw new Error('Profile button not found');
  button.click();
  let ready = await waitFor(() => getDialog(), 5000, 100);
  if (ready) return ready;
  button.click();
  ready = await waitFor(() => getDialog(), 5000, 100);
  if (!ready) throw new Error('Profile dialog did not open');
  return ready;
}

function chooseSectionByLabel(label) {
  if (!label) return false;
  const dialog = getDialog();
  if (!dialog) return false;
  const want = label.toLowerCase();
  const radio = Array.from(dialog.querySelectorAll('label')).find((el) => normalizeText(el.textContent || '').toLowerCase() === want);
  if (!radio) return false;
  radio.click();
  return true;
}

function chooseProfile() {
  const dialog = getDialog();
  if (!dialog) return null;
  const wantModel = targetModel.toLowerCase();
  const wantAssistant = targetAssistant.toLowerCase();
  const options = Array.from(dialog.querySelectorAll('li.option[role="option"]'));
  const match = options.find((el) => {
    const text = normalizeText(el.textContent || '').toLowerCase();
    const modelId = (el.querySelector('svg.model-icon[data-model]')?.getAttribute('data-model') || '').toLowerCase();
    if (wantModel) {
      return modelId === wantModel || text === wantModel;
    }
    if (wantAssistant) {
      return text === wantAssistant;
    }
    return false;
  });
  if (!match) return null;
  const before = summarizeButton(getProfileButton());
  match.click();
  return {
    text: normalizeText(match.textContent || ''),
    modelId: match.querySelector('svg.model-icon[data-model]')?.getAttribute('data-model') || null,
    before,
  };
}

await ensureDialogOpen();
if (targetSection) {
  chooseSectionByLabel(targetSection);
  await sleep(150);
}
const selected = chooseProfile();
if (selected) {
  await sleep(800);
}
return {
  href: location.href,
  profileButton: summarizeButton(getProfileButton()),
  sections: radioSections(),
  options: openEntries(),
  selected,
};
