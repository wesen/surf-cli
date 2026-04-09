const ls = {};
for (let i = 0; i < localStorage.length; i += 1) {
  const key = localStorage.key(i);
  ls[key] = (localStorage.getItem(key) || '').slice(0, 200);
}
const ss = {};
for (let i = 0; i < sessionStorage.length; i += 1) {
  const key = sessionStorage.key(i);
  ss[key] = (sessionStorage.getItem(key) || '').slice(0, 200);
}
const cookiePreview = document.cookie.slice(0, 500);
const globalKeys = Object.keys(window).filter((key) => /auth|token|access|session|next|apollo/i.test(key)).slice(0, 100);
return {
  href: location.href,
  localStorage: ls,
  sessionStorage: ss,
  cookiePreview,
  globalKeys,
};
