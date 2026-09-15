"use client";

import { useState, type ReactNode } from "react";
import * as Dialog from "@radix-ui/react-dialog";
import type { LucideIcon } from "lucide-react";
import {
  Archive,
  BarChart3,
  Boxes,
  ChevronDown,
  ClipboardCheck,
  LayoutDashboard,
  LoaderCircle,
  LogOut,
  Menu,
  PackageCheck,
  PanelLeftClose,
  PanelLeftOpen,
  ShieldCheck,
  Truck,
  UsersRound,
  Warehouse,
  X,
} from "lucide-react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";

import { useAuth } from "@/components/auth/auth-provider";
import { ConnectionStatus } from "@/components/layout/connection-status";
import { Button } from "@/components/ui/button";
import {
  MODULE_ACCESS,
  type PermissionRequirement,
} from "@/lib/auth/permissions";
import { cn } from "@/lib/utils";

interface NavigationItem {
  label: string;
  icon: LucideIcon;
  href?: string;
  badge?: string;
  access?: PermissionRequirement;
}

interface NavigationGroup {
  label: string;
  items: NavigationItem[];
}

interface ReportNavigationItem {
  label: string;
  href: string;
}

interface MasterDataNavigationItem {
  label: string;
  href: string;
}

const navigation: NavigationGroup[] = [
  {
    label: "Workspace",
    items: [{ label: "Overview", icon: LayoutDashboard, href: "/" }],
  },
  {
    label: "Operations",
    items: [
      {
        label: "Inbound",
        icon: Archive,
        badge: "12",
        access: MODULE_ACCESS.INBOUND.READ,
      },
      {
        label: "Inventory",
        icon: Boxes,
        access: MODULE_ACCESS.INVENTORY.READ,
      },
      {
        label: "Stock control",
        icon: ClipboardCheck,
        badge: "3",
        access: MODULE_ACCESS.INVENTORY.READ,
      },
      {
        label: "Outbound",
        icon: PackageCheck,
        badge: "8",
        access: MODULE_ACCESS.OUTBOUND.READ,
      },
      {
        label: "Transport",
        icon: Truck,
        access: MODULE_ACCESS.OUTBOUND.READ,
      },
    ],
  },
  {
    label: "Administration",
    items: [
      {
        label: "Accounts",
        icon: UsersRound,
        href: "/accounts",
        access: MODULE_ACCESS.SECURITY.READ,
      },
      {
        label: "Permissions",
        icon: ShieldCheck,
        href: "/permissions",
        access: MODULE_ACCESS.SECURITY.READ,
      },
    ],
  },
];

const masterDataNavigation: MasterDataNavigationItem[] = [
  { label: "All master data", href: "/master-data" },
  { label: "Core masters", href: "/master-data/core" },
  { label: "Catalog", href: "/master-data/catalog" },
  { label: "Operational setup", href: "/master-data/operational" },
];

const reportNavigation: ReportNavigationItem[] = [
  { label: "All reports", href: "/reports" },
  {
    label: "Master data",
    href: "/reports#master-data",
  },
  {
    label: "Inbound",
    href: "/reports#inbound",
  },
  {
    label: "Stock control",
    href: "/reports#stock-control",
  },
  {
    label: "Outbound",
    href: "/reports#outbound",
  },
  {
    label: "Billing",
    href: "/reports#billing",
  },
];

function initials(name: string) {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  return (
    parts.length > 1
      ? `${parts[0][0]}${parts.at(-1)?.[0] ?? ""}`
      : parts[0]?.slice(0, 2) || "WU"
  ).toUpperCase();
}

function isSectionActive(pathname: string, href: string) {
  return (
    pathname === href ||
    (href !== "/master-data" && pathname.startsWith(`${href}/`))
  );
}

function Brand({
  collapsed = false,
  onToggleCollapsed,
}: {
  collapsed?: boolean;
  onToggleCollapsed?: () => void;
}) {
  return (
    <div
      className={cn(
        "relative flex h-16 shrink-0 items-center gap-3",
        collapsed ? "justify-center px-2" : "px-5",
      )}
    >
      <div className="grid size-9 place-items-center rounded-xl bg-cyan-400 text-sm font-black text-slate-950 shadow-[0_0_0_4px_rgba(34,211,238,0.12)]">
        W
      </div>
      {!collapsed ? (
        <div>
          <p className="text-[15px] font-bold tracking-tight text-white">
            NEXA WMS
          </p>
          <p className="text-xs text-slate-400">Operations control</p>
        </div>
      ) : null}
      {onToggleCollapsed ? (
        <button
          type="button"
          onClick={onToggleCollapsed}
          aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
          title={collapsed ? "Expand sidebar" : "Collapse sidebar"}
          className="absolute top-1/2 -right-3 z-10 grid size-7 -translate-y-1/2 place-items-center rounded-full border border-slate-700 bg-slate-900 text-slate-300 shadow-md transition-colors hover:border-slate-500 hover:bg-slate-800 hover:text-white"
        >
          {collapsed ? (
            <PanelLeftOpen className="size-4" />
          ) : (
            <PanelLeftClose className="size-4" />
          )}
        </button>
      ) : null}
    </div>
  );
}

function NavigationEntry({
  item,
  active,
  collapsed = false,
  onNavigate,
}: {
  item: NavigationItem;
  active: boolean;
  collapsed?: boolean;
  onNavigate?: () => void;
}) {
  const Icon = item.icon;
  const className = cn(
    "flex min-h-11 w-full items-center gap-3 rounded-xl text-left text-sm font-medium transition-colors",
    collapsed ? "justify-center px-2" : "px-3",
    active
      ? "bg-cyan-400 text-slate-950"
      : "text-slate-300 hover:bg-white/[0.07] hover:text-white",
  );
  const content = (
    <>
      <Icon className="size-[18px] shrink-0" />
      <span className={collapsed ? "sr-only" : "flex-1"}>{item.label}</span>
      {item.badge && !collapsed ? (
        <span
          className={cn(
            "rounded-full px-2 py-0.5 text-xs font-bold",
            active
              ? "bg-slate-950/10 text-slate-950"
              : "bg-white/10 text-slate-300",
          )}
        >
          {item.badge}
        </span>
      ) : null}
    </>
  );

  if (item.href) {
    return (
      <Link
        href={item.href}
        onClick={onNavigate}
        className={className}
        title={collapsed ? item.label : undefined}
        aria-current={active ? "page" : undefined}
      >
        {content}
      </Link>
    );
  }

  return (
    <button
      type="button"
      onClick={onNavigate}
      className={className}
      title={collapsed ? item.label : undefined}
      aria-label={collapsed ? item.label : undefined}
    >
      {content}
    </button>
  );
}

function SidebarNavigation({
  collapsed = false,
  onToggleCollapsed,
  onNavigate,
}: {
  collapsed?: boolean;
  onToggleCollapsed?: () => void;
  onNavigate?: () => void;
}) {
  const pathname = usePathname();
  const router = useRouter();
  const { user, canAccess } = useAuth();
  const [loggingOut, setLoggingOut] = useState(false);
  const [reportsOpen, setReportsOpen] = useState(
    pathname.startsWith("/reports"),
  );
  const [masterDataOpen, setMasterDataOpen] = useState(
    pathname.startsWith("/master-data"),
  );
  const reportsActive = pathname.startsWith("/reports");
  const masterDataActive = pathname.startsWith("/master-data");
  const canViewMasterData = canAccess(MODULE_ACCESS.MASTER.READ);
  const canViewReports = canAccess(MODULE_ACCESS.REPORTING.READ);
  const accountInitials = initials(user.display_name || user.username);

  async function logout() {
    setLoggingOut(true);
    try {
      await fetch("/api/auth/logout", { method: "POST" });
    } finally {
      router.replace("/login");
      router.refresh();
    }
  }

  return (
    <>
      <Brand collapsed={collapsed} onToggleCollapsed={onToggleCollapsed} />
      <div
        className={cn(
          "mt-2 rounded-xl border border-white/10 bg-white/[0.06]",
          collapsed ? "mx-2 grid h-11 place-items-center p-2" : "mx-4 p-3",
        )}
        title={collapsed ? "Active warehouse: Jakarta Distribution" : undefined}
      >
        {collapsed ? (
          <Warehouse className="size-[18px] text-cyan-300" />
        ) : (
          <div className="flex items-center justify-between gap-3">
            <div className="min-w-0">
              <p className="text-xs font-medium text-slate-400">
                Active warehouse
              </p>
              <p className="mt-0.5 truncate text-sm font-semibold text-white">
                Jakarta Distribution
              </p>
            </div>
            <ChevronDown className="size-4 shrink-0 text-slate-400" />
          </div>
        )}
      </div>

      <nav
        aria-label="Primary navigation"
        className={cn(
          "flex-1 overflow-y-auto py-5",
          collapsed ? "px-2" : "px-3",
        )}
      >
        {navigation.slice(0, 2).map((group) => {
          const allowedItems = group.items.filter((item) =>
            canAccess(item.access),
          );
          if (allowedItems.length === 0) return null;

          return (
            <div
              key={group.label}
              className={cn(
                "mb-6",
                collapsed &&
                  "border-t border-white/10 pt-3 first:border-t-0 first:pt-0",
              )}
            >
              <p
                className={cn(
                  "mb-2 px-3 text-[11px] font-bold tracking-[0.16em] text-slate-500 uppercase",
                  collapsed && "sr-only",
                )}
              >
                {group.label}
              </p>
              <ul className="space-y-1">
                {allowedItems.map((item) => (
                  <li key={item.label}>
                    <NavigationEntry
                      item={item}
                      active={item.href === "/" && pathname === "/"}
                      collapsed={collapsed}
                      onNavigate={onNavigate}
                    />
                  </li>
                ))}
              </ul>
            </div>
          );
        })}

        {canViewReports ? (
          <div
            className={cn("mb-6", collapsed && "border-t border-white/10 pt-3")}
          >
            <p
              className={cn(
                "mb-2 px-3 text-[11px] font-bold tracking-[0.16em] text-slate-500 uppercase",
                collapsed && "sr-only",
              )}
            >
              Reporting
            </p>
            <button
              type="button"
              aria-expanded={!collapsed && reportsOpen}
              aria-controls="reports-navigation"
              aria-label={collapsed ? "Reports" : undefined}
              title={collapsed ? "Reports" : undefined}
              onClick={() => {
                if (collapsed) {
                  onToggleCollapsed?.();
                  setReportsOpen(true);
                } else {
                  setReportsOpen((open) => !open);
                }
              }}
              className={cn(
                "flex min-h-11 w-full items-center gap-3 rounded-xl text-left text-sm font-medium transition-colors",
                collapsed ? "justify-center px-2" : "px-3",
                reportsActive
                  ? "bg-cyan-400 text-slate-950"
                  : "text-slate-300 hover:bg-white/[0.07] hover:text-white",
              )}
            >
              <BarChart3 className="size-[18px] shrink-0" />
              <span className={collapsed ? "sr-only" : "flex-1"}>Reports</span>
              {!collapsed ? (
                <ChevronDown
                  className={cn(
                    "size-4 transition-transform",
                    reportsOpen && "rotate-180",
                  )}
                />
              ) : null}
            </button>
            {reportsOpen && !collapsed ? (
              <ul
                id="reports-navigation"
                className="mt-1 ml-5 space-y-0.5 border-l border-white/10 pl-4"
              >
                {reportNavigation.map((item) => (
                  <li key={item.label}>
                    <Link
                      href={item.href}
                      onClick={onNavigate}
                      className="flex min-h-9 items-center rounded-lg px-3 text-sm text-slate-400 transition-colors hover:bg-white/[0.06] hover:text-white"
                    >
                      {item.label}
                    </Link>
                  </li>
                ))}
              </ul>
            ) : null}
          </div>
        ) : null}

        {navigation.slice(2).map((group) => {
          const allowedItems = group.items.filter((item) =>
            canAccess(item.access),
          );
          if (allowedItems.length === 0 && !canViewMasterData) return null;

          return (
            <div
              key={group.label}
              className={cn(
                "mb-6",
                collapsed && "border-t border-white/10 pt-3",
              )}
            >
              <p
                className={cn(
                  "mb-2 px-3 text-[11px] font-bold tracking-[0.16em] text-slate-500 uppercase",
                  collapsed && "sr-only",
                )}
              >
                {group.label}
              </p>
              {canViewMasterData ? (
                <>
                  <button
                    type="button"
                    aria-expanded={!collapsed && masterDataOpen}
                    aria-controls="master-data-navigation"
                    aria-label={collapsed ? "Master data" : undefined}
                    title={collapsed ? "Master data" : undefined}
                    onClick={() => {
                      if (collapsed) {
                        onToggleCollapsed?.();
                        setMasterDataOpen(true);
                      } else {
                        setMasterDataOpen((open) => !open);
                      }
                    }}
                    className={cn(
                      "flex min-h-11 w-full items-center gap-3 rounded-xl text-left text-sm font-medium transition-colors",
                      collapsed ? "justify-center px-2" : "px-3",
                      masterDataActive
                        ? "bg-cyan-400 text-slate-950"
                        : "text-slate-300 hover:bg-white/[0.07] hover:text-white",
                    )}
                  >
                    <Warehouse className="size-[18px] shrink-0" />
                    <span className={collapsed ? "sr-only" : "flex-1"}>
                      Master data
                    </span>
                    {!collapsed ? (
                      <ChevronDown
                        className={cn(
                          "size-4 transition-transform",
                          masterDataOpen && "rotate-180",
                        )}
                      />
                    ) : null}
                  </button>
                  {masterDataOpen && !collapsed ? (
                    <ul
                      id="master-data-navigation"
                      className="mt-1 ml-5 space-y-0.5 border-l border-white/10 pl-4"
                    >
                      {masterDataNavigation.map((item) => {
                        const active = isSectionActive(pathname, item.href);
                        return (
                          <li key={item.href}>
                            <Link
                              href={item.href}
                              onClick={onNavigate}
                              aria-current={active ? "page" : undefined}
                              className={cn(
                                "flex min-h-9 items-center rounded-lg px-3 text-sm transition-colors",
                                active
                                  ? "bg-white/10 font-semibold text-white"
                                  : "text-slate-400 hover:bg-white/[0.06] hover:text-white",
                              )}
                            >
                              {item.label}
                            </Link>
                          </li>
                        );
                      })}
                    </ul>
                  ) : null}
                </>
              ) : null}
              <ul
                className={cn(
                  "space-y-1",
                  canViewMasterData && !collapsed && "mt-2",
                )}
              >
                {allowedItems.map((item) => (
                  <li key={item.label}>
                    <NavigationEntry
                      item={item}
                      active={Boolean(
                        item.href && isSectionActive(pathname, item.href),
                      )}
                      collapsed={collapsed}
                      onNavigate={onNavigate}
                    />
                  </li>
                ))}
              </ul>
            </div>
          );
        })}
      </nav>

      <div
        className={cn("border-t border-white/10", collapsed ? "p-2" : "p-4")}
      >
        <div
          className={cn(
            "flex rounded-xl p-2",
            collapsed ? "flex-col items-center gap-2" : "items-center gap-3",
          )}
        >
          <div
            className="grid size-9 shrink-0 place-items-center rounded-full bg-slate-700 text-sm font-bold text-white"
            title={collapsed ? user.display_name || user.username : undefined}
          >
            {accountInitials}
          </div>
          {!collapsed ? (
            <div className="min-w-0 flex-1">
              <p className="truncate text-sm font-semibold text-white">
                {user.display_name || user.username}
              </p>
              <p className="truncate text-xs text-slate-400">
                @{user.username}
              </p>
            </div>
          ) : null}
          <button
            type="button"
            onClick={logout}
            disabled={loggingOut}
            className="grid size-9 shrink-0 place-items-center rounded-lg text-slate-400 transition hover:bg-white/10 hover:text-white disabled:opacity-60"
            aria-label="Sign out"
            title="Sign out"
          >
            {loggingOut ? (
              <LoaderCircle className="size-4 animate-spin" />
            ) : (
              <LogOut className="size-4" />
            )}
          </button>
        </div>
      </div>
    </>
  );
}

function MobileSidebar() {
  const [open, setOpen] = useState(false);

  return (
    <Dialog.Root open={open} onOpenChange={setOpen}>
      <Dialog.Trigger asChild>
        <Button
          variant="ghost"
          size="icon"
          className="lg:hidden"
          aria-label="Open navigation"
        >
          <Menu className="size-5" />
        </Button>
      </Dialog.Trigger>
      <Dialog.Portal>
        <Dialog.Overlay className="data-[state=closed]:animate-out data-[state=open]:animate-in fixed inset-0 z-40 bg-slate-950/60 backdrop-blur-sm" />
        <Dialog.Content className="fixed inset-y-0 left-0 z-50 flex w-[min(86vw,19rem)] flex-col bg-slate-950 shadow-2xl focus:outline-none">
          <Dialog.Title className="sr-only">Main navigation</Dialog.Title>
          <Dialog.Close asChild>
            <button
              type="button"
              aria-label="Close navigation"
              className="absolute top-3 right-3 z-20 grid size-10 place-items-center rounded-lg text-slate-300 hover:bg-white/10 hover:text-white"
            >
              <X className="size-5" />
            </button>
          </Dialog.Close>
          <SidebarNavigation onNavigate={() => setOpen(false)} />
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}

export function AppShell({ children }: { children: ReactNode }) {
  const { user } = useAuth();
  const accountInitials = initials(user.display_name || user.username);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);

  return (
    <div className="min-h-screen bg-slate-50 text-slate-950">
      <aside
        className={cn(
          "fixed inset-y-0 left-0 z-30 hidden flex-col bg-slate-950 transition-[width] duration-200 lg:flex",
          sidebarCollapsed ? "w-20" : "w-68",
        )}
      >
        <SidebarNavigation
          collapsed={sidebarCollapsed}
          onToggleCollapsed={() => setSidebarCollapsed((value) => !value)}
        />
      </aside>

      <div
        className={cn(
          "transition-[padding] duration-200",
          sidebarCollapsed ? "lg:pl-20" : "lg:pl-68",
        )}
      >
        <header className="sticky top-0 z-20 flex h-16 items-center border-b border-slate-200/80 bg-white/90 px-4 backdrop-blur-xl sm:px-6 lg:px-8">
          <MobileSidebar />
          <div className="ml-2 min-w-0 lg:ml-0">
            <p className="truncate text-sm font-semibold text-slate-950">
              Jakarta Distribution Center
            </p>
            <p className="hidden text-xs text-slate-500 sm:block">
              Operational day · 10 September 2026
            </p>
          </div>
          <div className="ml-auto flex items-center gap-2">
            <ConnectionStatus />
            <button
              type="button"
              className="grid size-10 place-items-center rounded-full bg-slate-950 text-sm font-bold text-white lg:hidden"
              aria-label="Open account menu"
            >
              {accountInitials}
            </button>
          </div>
        </header>

        <main className="surface-grid min-h-[calc(100vh-4rem)] px-4 py-6 sm:px-6 lg:px-8 lg:py-8">
          <div className="mx-auto max-w-[96rem]">{children}</div>
        </main>
      </div>
    </div>
  );
}
