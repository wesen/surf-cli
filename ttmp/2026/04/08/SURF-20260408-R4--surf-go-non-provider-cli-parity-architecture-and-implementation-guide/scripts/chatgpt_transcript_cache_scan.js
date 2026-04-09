const conversationId = location.pathname.split('/').filter(Boolean).pop();
const hits = [];
const scan = (storageName, storage) => {
  for (let i = 0; i < storage.length; i += 1) {
    const key = storage.key(i);
    const value = storage.getItem(key) || '';
    if (key.indexOf(conversationId) !== -1 || value.indexOf(conversationId) !== -1 || /conversation/i.test(key)) {
      hits.push({
        storage: storageName,
        key,
        containsConversationId: value.indexOf(conversationId) !== -1,
        preview: value.slice(0, 500),
      });
    }
  }
};
scan('localStorage', localStorage);
scan('sessionStorage', sessionStorage);
return { conversationId, hitCount: hits.length, hits };
