import type { ReactNode } from "react";
import { useAuth } from "./AuthProvider";

/**
 * Today this is a passthrough since auth isn't wired yet. When auth lands:
 *
 *   const { user, loading } = useAuth();
 *   if (loading) return <FullPageSpinner />;
 *   if (!user) return <Navigate to="/login" replace />;
 *   return <>{children}</>;
 */
export function RequireAuth({ children }: { children: ReactNode }) {
  // Touch the context so a missing AuthProvider trips loudly during dev.
  useAuth();
  return <>{children}</>;
}
