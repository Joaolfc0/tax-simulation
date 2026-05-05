import {
  createContext,
  useContext,
  useMemo,
  type ReactNode,
} from "react";

export interface AuthUser {
  id: string;
  name: string;
  email: string;
}

export interface AuthState {
  user: AuthUser | null;
  loading: boolean;
  signIn: (email: string, password: string) => Promise<void>;
  signOut: () => Promise<void>;
}

const AuthContext = createContext<AuthState | undefined>(undefined);

/**
 * Stub AuthProvider — keeps the surface ready for the day the backend exposes
 * `/auth/*` endpoints. Today, `user` is always `null` and `RequireAuth` just
 * renders its children.
 *
 * To wire real auth later:
 * 1. Implement `signIn`/`signOut` with API calls.
 * 2. Persist the session (cookie, localStorage, etc.) and expose `user`.
 * 3. In `RequireAuth`, redirect to `/login` when `!user`.
 * 4. Pass an Authorization header through `apiClient` (see lib/api/client.ts).
 */
export function AuthProvider({ children }: { children: ReactNode }) {
  const value = useMemo<AuthState>(
    () => ({
      user: null,
      loading: false,
      signIn: async () => {
        throw new Error("Auth not implemented yet");
      },
      signOut: async () => {
        // no-op
      },
    }),
    [],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthState {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return ctx;
}
