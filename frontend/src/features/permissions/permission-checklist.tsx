"use client";

import type { Permission } from "@/features/permissions/permission-types";

export function PermissionChecklist({
  permissions,
  selected,
  onChange,
  disabled,
}: {
  permissions: readonly Permission[];
  selected: readonly string[];
  onChange: (permissionIds: string[]) => void;
  disabled?: boolean;
}) {
  const selectedIds = new Set(selected);
  const groups = permissions.reduce((result, item) => {
    const group = result.get(item.module_code) ?? [];
    group.push(item);
    result.set(item.module_code, group);
    return result;
  }, new Map<string, Permission[]>());

  function toggle(permissionId: string) {
    const next = new Set(selectedIds);
    if (next.has(permissionId)) next.delete(permissionId);
    else next.add(permissionId);
    onChange([...next]);
  }

  function toggleModule(items: readonly Permission[]) {
    const next = new Set(selectedIds);
    const allSelected = items.every((item) => next.has(item.permission_id));
    for (const item of items) {
      if (allSelected) next.delete(item.permission_id);
      else next.add(item.permission_id);
    }
    onChange([...next]);
  }

  return (
    <div className="max-h-[min(25rem,50vh)] space-y-3 overflow-y-auto rounded-xl border border-slate-200 bg-slate-50 p-3">
      {[...groups.entries()].map(([moduleCode, items]) => {
        const allSelected = items.every((item) =>
          selectedIds.has(item.permission_id),
        );
        return (
          <section
            key={moduleCode}
            className="overflow-hidden rounded-xl border border-slate-200 bg-white"
          >
            <div className="flex items-center justify-between gap-3 border-b border-slate-100 bg-slate-50 px-3 py-2">
              <h3 className="text-xs font-bold tracking-wide text-slate-700 uppercase">
                {moduleCode}
              </h3>
              <button
                type="button"
                disabled={disabled}
                className="min-h-8 rounded-lg px-2 text-xs font-semibold text-cyan-800 hover:bg-cyan-50 disabled:opacity-50"
                onClick={() => toggleModule(items)}
              >
                {allSelected ? "Clear module" : "Select module"}
              </button>
            </div>
            <div className="grid gap-1 p-2 sm:grid-cols-2">
              {items.map((permission) => (
                <label
                  key={permission.permission_id}
                  className="flex min-h-12 cursor-pointer items-start gap-3 rounded-lg px-2 py-2 hover:bg-slate-50"
                >
                  <input
                    type="checkbox"
                    checked={selectedIds.has(permission.permission_id)}
                    disabled={disabled}
                    onChange={() => toggle(permission.permission_id)}
                    className="mt-0.5 size-4 shrink-0 accent-cyan-600"
                  />
                  <span className="min-w-0">
                    <span className="block text-sm font-semibold text-slate-800">
                      {permission.name}
                    </span>
                    <span className="block font-mono text-[11px] text-slate-500">
                      {permission.code}
                    </span>
                  </span>
                </label>
              ))}
            </div>
          </section>
        );
      })}
    </div>
  );
}
