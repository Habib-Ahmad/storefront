"use client";

import {
  createContext,
  useContext,
  useEffect,
  useState,
  useCallback,
  useRef,
  type ReactNode,
} from "react";
import type { Session } from "@supabase/supabase-js";
import { getSupabase } from "@/lib/supabase";
import { api } from "@/lib/api";
import {
  clearSessionPolicy,
  getSessionTimeoutState,
  initializeSessionPolicy,
  noteSessionActivity,
  SESSION_TIMEOUT_ERROR_CODE,
} from "@/lib/session-policy";

interface AuthContextValue {
  session: Session | null;
  loading: boolean;
  signOut: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue>({
  session: null,
  loading: true,
  signOut: async () => {},
});

export function useSession() {
  return useContext(AuthContext);
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [session, setSession] = useState<Session | null>(null);
  const [loading, setLoading] = useState(true);
  const sessionRef = useRef<Session | null>(null);
  const timingOutRef = useRef(false);

  useEffect(() => {
    const supabase = getSupabase();
    if (!supabase) {
      setLoading(false);
      return;
    }
    const supabaseClient = supabase;

    let cancelled = false;

    function syncSession(newSession: Session | null) {
      sessionRef.current = newSession;
      setSession(newSession);
      api.setToken(newSession?.access_token ?? null);
    }

    function clearSessionState() {
      clearSessionPolicy();
      syncSession(null);
    }

    async function handleSessionTimeout() {
      if (timingOutRef.current) {
        return;
      }

      timingOutRef.current = true;
      clearSessionState();
      api.setRefreshHandler(null);

      try {
        await supabaseClient.auth.signOut();
      } finally {
        if (!cancelled) {
          setLoading(false);
          window.location.replace(`/login?error=${SESSION_TIMEOUT_ERROR_CODE}`);
        }
      }
    }

    async function applySession(newSession: Session | null) {
      if (!newSession) {
        timingOutRef.current = false;
        clearSessionState();
        if (!cancelled) {
          setLoading(false);
        }
        return;
      }

      initializeSessionPolicy();
      const timeoutState = getSessionTimeoutState();
      if (timeoutState.timedOut) {
        await handleSessionTimeout();
        return;
      }

      timingOutRef.current = false;
      noteSessionActivity();
      syncSession(newSession);
      if (!cancelled) {
        setLoading(false);
      }
    }

    supabaseClient.auth.getSession().then(({ data }) => {
      void applySession(data.session);
    });

    const {
      data: { subscription },
    } = supabaseClient.auth.onAuthStateChange((_event, newSession) => {
      void applySession(newSession);
    });

    function markSessionActive() {
      if (!sessionRef.current || timingOutRef.current) {
        return;
      }

      const timeoutState = getSessionTimeoutState();
      if (timeoutState.timedOut) {
        void handleSessionTimeout();
        return;
      }

      noteSessionActivity();
    }

    function handleVisibilityChange() {
      if (document.visibilityState === "visible") {
        markSessionActive();
      }
    }

    window.addEventListener("pointerdown", markSessionActive);
    window.addEventListener("keydown", markSessionActive);
    window.addEventListener("focus", markSessionActive);
    document.addEventListener("visibilitychange", handleVisibilityChange);

    const intervalId = window.setInterval(() => {
      if (!sessionRef.current || timingOutRef.current) {
        return;
      }

      const timeoutState = getSessionTimeoutState();
      if (timeoutState.timedOut) {
        void handleSessionTimeout();
      }
    }, 60_000);

    api.setRefreshHandler(async () => {
      if (sessionRef.current) {
        const timeoutState = getSessionTimeoutState();
        if (timeoutState.timedOut) {
          await handleSessionTimeout();
          return null;
        }
      }

      const { data } = await supabaseClient.auth.refreshSession();
      if (data.session) {
        initializeSessionPolicy();
        noteSessionActivity();
        syncSession(data.session);
      }
      return data.session?.access_token ?? null;
    });

    return () => {
      cancelled = true;
      subscription.unsubscribe();
      api.setRefreshHandler(null);
      window.clearInterval(intervalId);
      window.removeEventListener("pointerdown", markSessionActive);
      window.removeEventListener("keydown", markSessionActive);
      window.removeEventListener("focus", markSessionActive);
      document.removeEventListener("visibilitychange", handleVisibilityChange);
    };
  }, []);

  const signOut = useCallback(async () => {
    const supabase = getSupabase();
    clearSessionPolicy();
    if (supabase) await supabase.auth.signOut();
    api.setToken(null);
    sessionRef.current = null;
    setSession(null);
  }, []);

  return (
    <AuthContext.Provider value={{ session, loading, signOut }}>{children}</AuthContext.Provider>
  );
}
