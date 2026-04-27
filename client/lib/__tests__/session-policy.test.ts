import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  clearSessionPolicy,
  getSessionTimeoutState,
  initializeSessionPolicy,
  noteSessionActivity,
  SESSION_IDLE_TIMEOUT_MS,
} from "@/lib/session-policy";

describe("session policy", () => {
  beforeEach(() => {
    window.localStorage.clear();
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-04-26T12:00:00Z"));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("uses the current time as the initial inactivity baseline", () => {
    const now = Date.parse("2026-04-26T12:00:00Z");
    initializeSessionPolicy(now);

    const state = getSessionTimeoutState(now);

    expect(state.lastActivityAt).toBe(now);
    expect(state.timedOut).toBe(false);
  });

  it("times out idle sessions after the configured inactivity window", () => {
    const now = Date.parse("2026-04-24T10:00:00Z");

    initializeSessionPolicy(now);
    const state = getSessionTimeoutState(now + SESSION_IDLE_TIMEOUT_MS + 1);

    expect(state.timedOut).toBe(true);
  });

  it("extends the session when activity is recorded inside the timeout window", () => {
    const startedAt = Date.parse("2026-04-26T08:00:00Z");
    initializeSessionPolicy(startedAt);

    noteSessionActivity(startedAt + 60_000);
    const state = getSessionTimeoutState(startedAt + SESSION_IDLE_TIMEOUT_MS);

    expect(state.timedOut).toBe(false);
    expect(state.lastActivityAt).toBe(startedAt + 60_000);
  });

  it("clears persisted timestamps on sign-out", () => {
    const now = Date.parse("2026-04-25T12:00:00Z");
    initializeSessionPolicy(now);
    clearSessionPolicy();

    const state = getSessionTimeoutState(now);

    expect(state.lastActivityAt).toBe(now);
    expect(state.timedOut).toBe(false);
  });
});
