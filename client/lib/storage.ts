export const PRODUCTS_KNOWN_STORAGE_KEY = "storefront.products.has-items";
export const SESSION_LAST_ACTIVITY_AT_STORAGE_KEY = "storefront:auth:last-activity-at";

export function getScopedStorageKey(baseKey: string, scope: string | null | undefined) {
  if (!scope) {
    return null;
  }

  return `${baseKey}:${scope}`;
}

export function removeStorageKey(key: string) {
  if (typeof window === "undefined") {
    return;
  }

  window.localStorage.removeItem(key);
}

export function removeStorageKeysByPrefix(prefix: string) {
  if (typeof window === "undefined") {
    return;
  }

  const keysToRemove: string[] = [];
  for (let index = 0; index < window.localStorage.length; index += 1) {
    const key = window.localStorage.key(index);
    if (key?.startsWith(prefix)) {
      keysToRemove.push(key);
    }
  }

  for (const key of keysToRemove) {
    window.localStorage.removeItem(key);
  }
}
