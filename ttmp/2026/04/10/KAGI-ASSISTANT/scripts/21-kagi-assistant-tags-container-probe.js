function normalizeText(value) {
  return (value || '').replace(/\s+/g, ' ').trim();
}

const tags = document.querySelector('#tags');
const tagsAdd = document.querySelector('#tags-add');
const threadMenu = document.querySelector('.thread-more-menu');

return {
  href: location.href,
  title: document.title,
  tags: tags ? {
    text: normalizeText(tags.textContent || ''),
    html: tags.outerHTML.slice(0, 3000),
  } : null,
  tagsAdd: tagsAdd ? {
    text: normalizeText(tagsAdd.textContent || ''),
    title: tagsAdd.getAttribute('title'),
    cls: (tagsAdd.className || '').toString().slice(0, 200),
    html: tagsAdd.outerHTML,
  } : null,
  threadMenu: threadMenu ? {
    text: normalizeText(threadMenu.textContent || ''),
    html: threadMenu.outerHTML.slice(0, 3000),
  } : null,
};
