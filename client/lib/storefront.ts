const DEFAULT_STOREFRONT_SLUG = "store";
const STOREFRONT_SLUG_MAX_LENGTH = 50;

export function normalizeStorefrontSlug(value: string) {
  return value
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/-{2,}/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, STOREFRONT_SLUG_MAX_LENGTH);
}

export function getTemporaryStorefrontSlugPreview(name: string) {
  return normalizeStorefrontSlug(name) || DEFAULT_STOREFRONT_SLUG;
}
