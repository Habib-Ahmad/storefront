export const SESSION_IDLE_TIMEOUT_MS = 24 * 60 * 60 * 1000;
export const SESSION_TIMEOUT_ERROR_CODE = "session_timeout";

import { removeStorageKey, SESSION_LAST_ACTIVITY_AT_STORAGE_KEY } from "@/lib/storage";

type SessionTimeoutState = {
  timedOut: boolean;
  lastActivityAt: number;
};

function parseTimestamp(raw: string | null) {
  if (!raw) {
    return null;
  }

  const parsed = Number(raw);
  if (!Number.isFinite(parsed) || parsed <= 0) {
    return null;
  }

  return parsed;
}

function getStoredLastActivity(now: number) {
  return parseTimestamp(window.localStorage.getItem(SESSION_LAST_ACTIVITY_AT_STORAGE_KEY)) ?? now;
}

export function initializeSessionPolicy(now = Date.now()) {
  if (parseTimestamp(window.localStorage.getItem(SESSION_LAST_ACTIVITY_AT_STORAGE_KEY)) === null) {
    window.localStorage.setItem(SESSION_LAST_ACTIVITY_AT_STORAGE_KEY, String(now));
  }
}

export function noteSessionActivity(at = Date.now()) {
  window.localStorage.setItem(SESSION_LAST_ACTIVITY_AT_STORAGE_KEY, String(at));
}

export function clearSessionPolicy() {
  removeStorageKey(SESSION_LAST_ACTIVITY_AT_STORAGE_KEY);
}

export function getSessionTimeoutState(now = Date.now()): SessionTimeoutState {
  const lastActivityAt = getStoredLastActivity(now);

  return {
    timedOut: now - lastActivityAt >= SESSION_IDLE_TIMEOUT_MS,
    lastActivityAt,
  };
}
