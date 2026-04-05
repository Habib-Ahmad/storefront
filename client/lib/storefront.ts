const DEFAULT_STOREFRONT_SLUG = "store";
const RESERVED_STOREFRONT_SUFFIX = "-store";
const STOREFRONT_SLUG_MAX_LENGTH = 50;
const RESERVED_STOREFRONT_SLUGS = new Set([
  "about",
  "admin",
  "api",
  "app",
  "contact",
  "login",
  "onboard",
  "signup",
  "track",
]);

export function normalizeStorefrontSlug(value: string) {
  return value
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/-{2,}/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, STOREFRONT_SLUG_MAX_LENGTH);
}

function isReservedStorefrontSlug(value: string) {
  return RESERVED_STOREFRONT_SLUGS.has(value);
}

function makeTemporaryStorefrontSlugSafe(value: string) {
  if (!isReservedStorefrontSlug(value)) {
    return value;
  }

  const safe = `${value}${RESERVED_STOREFRONT_SUFFIX}`.slice(0, STOREFRONT_SLUG_MAX_LENGTH);
  return safe.replace(/-+$/g, "") || DEFAULT_STOREFRONT_SLUG;
}

export function getTemporaryStorefrontSlugPreview(name: string) {
  const base = normalizeStorefrontSlug(name) || DEFAULT_STOREFRONT_SLUG;
  return makeTemporaryStorefrontSlugSafe(base);
}
