import { createClientUUID } from "@/lib/utils";

const STORAGE_KEY = "storefront.public-checkout-recovery.v1";
const RECOVERY_WINDOW_MS = 45 * 60 * 1000;

type CheckoutRecoveryRecord = {
  contextKey: string;
  storefrontSlug: string;
  checkoutId: string;
  trackingSlug: string | null;
  orderPath: string | null;
  updatedAt: number;
};

function isBrowser() {
  return typeof window !== "undefined" && typeof window.localStorage !== "undefined";
}

function parseStoredRecords(raw: string | null): CheckoutRecoveryRecord[] {
  if (!raw) {
    return [];
  }

  try {
    const parsed = JSON.parse(raw);
    if (!Array.isArray(parsed)) {
      return [];
    }

    return parsed.filter((value): value is CheckoutRecoveryRecord => {
      return (
        value &&
        typeof value === "object" &&
        typeof value.contextKey === "string" &&
        typeof value.storefrontSlug === "string" &&
        typeof value.checkoutId === "string" &&
        (typeof value.trackingSlug === "string" || value.trackingSlug === null) &&
        (typeof value.orderPath === "string" || value.orderPath === null) &&
        typeof value.updatedAt === "number"
      );
    });
  } catch {
    return [];
  }
}

function isFresh(record: CheckoutRecoveryRecord) {
  return Date.now() - record.updatedAt <= RECOVERY_WINDOW_MS;
}

function readRecords() {
  if (!isBrowser()) {
    return [];
  }

  return parseStoredRecords(window.localStorage.getItem(STORAGE_KEY)).filter(isFresh);
}

function writeRecords(records: CheckoutRecoveryRecord[]) {
  if (!isBrowser()) {
    return;
  }

  window.localStorage.setItem(STORAGE_KEY, JSON.stringify(records.filter(isFresh)));
}

function upsertRecord(record: CheckoutRecoveryRecord) {
  const existing = readRecords().filter((item) => item.contextKey !== record.contextKey);
  existing.push(record);
  writeRecords(existing);
}

export function basketRecoveryKey(storefrontSlug: string) {
  return `basket:${storefrontSlug}`;
}

export function productRecoveryKey(storefrontSlug: string, productId: string) {
  return `product:${storefrontSlug}:${productId}`;
}

export function getOrCreateCheckoutId(contextKey: string, storefrontSlug: string) {
  const existing = readRecords().find((record) => record.contextKey === contextKey);
  if (existing) {
    return existing.checkoutId;
  }

  const checkoutId = createClientUUID();
  upsertRecord({
    contextKey,
    storefrontSlug,
    checkoutId,
    trackingSlug: null,
    orderPath: null,
    updatedAt: Date.now(),
  });
  return checkoutId;
}

export function rememberPendingOrder(
  contextKey: string,
  storefrontSlug: string,
  trackingSlug: string,
) {
  const existing = readRecords().find((record) => record.contextKey === contextKey);

  upsertRecord({
    contextKey,
    storefrontSlug,
    checkoutId: existing?.checkoutId ?? createClientUUID(),
    trackingSlug,
    orderPath: `/order/${trackingSlug}`,
    updatedAt: Date.now(),
  });
}

export function getLatestPendingOrderForStorefront(storefrontSlug: string) {
  return (
    readRecords()
      .filter(
        (record) =>
          record.storefrontSlug === storefrontSlug && record.orderPath && record.trackingSlug,
      )
      .sort((left, right) => right.updatedAt - left.updatedAt)[0] ?? null
  );
}

export function clearPendingOrderByTrackingSlug(trackingSlug: string) {
  writeRecords(readRecords().filter((record) => record.trackingSlug !== trackingSlug));
}

export function keepPendingOrderByTrackingSlug(trackingSlug: string) {
  const records = readRecords();
  const target = records.find((record) => record.trackingSlug === trackingSlug);
  if (!target) {
    return;
  }

  upsertRecord({
    ...target,
    updatedAt: Date.now(),
  });
}
