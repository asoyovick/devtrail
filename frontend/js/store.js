const cache = new Map();
const inflight = new Map();

export function cacheGet(key) { return cache.get(key); }
export function cacheSet(key, value) { cache.set(key, value); return value; }
export function cacheDel(key) { cache.delete(key); }
export function cacheClear() { cache.clear(); inflight.clear(); }

export async function cached(key, fn, { ttl = 30000 } = {}) {
  const entry = cache.get(key);
  if (entry && entry.expires > Date.now()) return entry.value;
  if (inflight.has(key)) return inflight.get(key);
  const p = Promise.resolve(fn()).then(value => {
    cache.set(key, { value, expires: Date.now() + ttl });
    inflight.delete(key);
    return value;
  }).catch(err => { inflight.delete(key); throw err; });
  inflight.set(key, p);
  return p;
}
