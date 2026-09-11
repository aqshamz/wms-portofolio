"use client";

import { createContext, useContext, useMemo, type ReactNode } from "react";

import {
  hasPermission,
  isAuthorized,
  type PermissionCode,
  type PermissionRequirement,
} from "@/lib/auth/permissions";
import type { AuthenticatedUser } from "@/lib/auth/schemas";

interface AuthContextValue {
  user: AuthenticatedUser;
  can: (permission: PermissionCode) => boolean;
  canAccess: (requirement?: PermissionRequirement) => boolean;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({
  user,
  children,
}: {
  user: AuthenticatedUser;
  children: ReactNode;
}) {
  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      can: (permission) => hasPermission(user.permissions, permission),
      canAccess: (requirement) => isAuthorized(user.permissions, requirement),
    }),
    [user],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within AuthProvider.");
  }
  return context;
}

export function PermissionGate({
  requirement,
  fallback = null,
  children,
}: {
  requirement: PermissionRequirement;
  fallback?: ReactNode;
  children: ReactNode;
}) {
  const { canAccess } = useAuth();
  return canAccess(requirement) ? children : fallback;
}
