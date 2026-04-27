export const SESSION_IDLE_TIMEOUT_MS = 24 * 60 * 60 * 1000;
export const SESSION_TIMEOUT_ERROR_CODE = "session_timeout";

const lastActivityAtKey = "storefront:auth:last-activity-at";

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
  return parseTimestamp(window.localStorage.getItem(lastActivityAtKey)) ?? now;
}

export function initializeSessionPolicy(now = Date.now()) {
  if (parseTimestamp(window.localStorage.getItem(lastActivityAtKey)) === null) {
    window.localStorage.setItem(lastActivityAtKey, String(now));
  }
}

export function noteSessionActivity(at = Date.now()) {
  window.localStorage.setItem(lastActivityAtKey, String(at));
}

export function clearSessionPolicy() {
  window.localStorage.removeItem(lastActivityAtKey);
}

export function getSessionTimeoutState(now = Date.now()): SessionTimeoutState {
  const lastActivityAt = getStoredLastActivity(now);

  return {
    timedOut: now - lastActivityAt >= SESSION_IDLE_TIMEOUT_MS,
    lastActivityAt,
  };
}
