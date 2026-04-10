return {
  href: location.href,
  title: document.title,
  bodyClasses: document.body ? document.body.className : null,
  loginMarkers: {
    gmailApp: !!document.querySelector('div[role="main"], [gh="tl"], form[role="search"]'),
    accountChooser: !!document.querySelector('a[href*="accounts.google.com"], div[data-view-id], [data-profileindex]'),
    loginForm: !!document.querySelector('input[type="email"], input[type="password"]'),
  },
  markers: {
    threadList: !!document.querySelector('[role="main"] table, [gh="tl"], tr[role="row"]'),
    searchBox: !!document.querySelector('input[placeholder*="Search"], input[aria-label*="Search mail"], form[role="search"] input'),
    compose: !!document.querySelector('div[gh="cm"], [role="button"][gh="cm"]'),
  },
};
