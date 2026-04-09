const databases = typeof indexedDB.databases === 'function' ? await indexedDB.databases() : [];
const results = [];
for (const dbInfo of databases) {
  const name = dbInfo && dbInfo.name;
  if (!name) {
    continue;
  }
  try {
    const opened = await new Promise((resolve, reject) => {
      const req = indexedDB.open(name);
      req.onerror = () => reject(req.error);
      req.onsuccess = () => resolve(req.result);
    });
    const objectStores = Array.from(opened.objectStoreNames);
    results.push({ name, version: opened.version, objectStores });
    opened.close();
  } catch (error) {
    results.push({ name, openError: String(error) });
  }
}
return { href: location.href, databases: results };
