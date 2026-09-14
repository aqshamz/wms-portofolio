"use client";

import { KeyRound, ShieldCheck, UsersRound } from "lucide-react";
import { parseAsStringLiteral, useQueryState } from "nuqs";

import { AccountPermissionAssignments } from "@/features/permissions/account-permission-assignments";
import { RolesPanel } from "@/features/permissions/roles-panel";
import type { PermissionView } from "@/features/permissions/permission-types";
import { cn } from "@/lib/utils";

const views = ["roles", "accounts"] as const;

export function PermissionsScreen({
  canWrite,
  currentAccountId,
}: {
  canWrite: boolean;
  currentAccountId: string;
}) {
  const [view, setView] = useQueryState(
    "view",
    parseAsStringLiteral(views).withDefault("roles"),
  );
  const currentView: PermissionView = view;

  return (
    <div className="space-y-6">
      <header>
        <div className="flex items-center gap-3">
          <div className="grid size-11 place-items-center rounded-xl bg-slate-950 text-cyan-300">
            <ShieldCheck className="size-5" />
          </div>
          <div>
            <h1 className="text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">
              Permissions
            </h1>
            <p className="mt-1 text-sm text-slate-600">
              Define reusable roles and control how accounts receive access.
            </p>
          </div>
        </div>
      </header>

      <div
        className="flex gap-2 overflow-x-auto pb-1"
        role="tablist"
        aria-label="Permission management sections"
      >
        <button
          type="button"
          role="tab"
          aria-selected={currentView === "roles"}
          onClick={() => void setView("roles")}
          className={cn(
            "inline-flex min-h-11 shrink-0 items-center gap-2 rounded-xl px-4 text-sm font-semibold",
            currentView === "roles"
              ? "bg-slate-950 text-white"
              : "border border-slate-200 bg-white text-slate-600 hover:bg-slate-50",
          )}
        >
          <KeyRound className="size-4" />
          Roles and permission sets
        </button>
        <button
          type="button"
          role="tab"
          aria-selected={currentView === "accounts"}
          onClick={() => void setView("accounts")}
          className={cn(
            "inline-flex min-h-11 shrink-0 items-center gap-2 rounded-xl px-4 text-sm font-semibold",
            currentView === "accounts"
              ? "bg-slate-950 text-white"
              : "border border-slate-200 bg-white text-slate-600 hover:bg-slate-50",
          )}
        >
          <UsersRound className="size-4" />
          Account assignments
        </button>
      </div>

      {currentView === "roles" ? (
        <RolesPanel canWrite={canWrite} />
      ) : (
        <AccountPermissionAssignments
          canWrite={canWrite}
          currentAccountId={currentAccountId}
        />
      )}
    </div>
  );
}
