return {
  href: location.href,
  title: document.title,
  ready: document.readyState,
  headings: Array.from(document.querySelectorAll('h1,h2,h3')).slice(0,20).map((el) => ({tag: el.tagName.toLowerCase(), text: (el.textContent || '').trim()})),
  buttons: Array.from(document.querySelectorAll('button')).slice(0,40).map((el) => ({text: (el.textContent || '').trim().replace(/\s+/g,' '), aria: el.getAttribute('aria-label'), title: el.getAttribute('title')})),
  inputs: Array.from(document.querySelectorAll('textarea,input,[contenteditable="true"]')).slice(0,20).map((el) => ({tag: el.tagName.toLowerCase(), type: el.getAttribute('type'), name: el.getAttribute('name'), placeholder: el.getAttribute('placeholder'), aria: el.getAttribute('aria-label'), role: el.getAttribute('role')})),
};
