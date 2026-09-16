export const WILDCARD_PERMISSION = "*";

export const PERMISSIONS = {
  AUTH: { READ: "AUTH.READ", WRITE: "AUTH.WRITE" },
  SECURITY: { READ: "SECURITY.READ", WRITE: "SECURITY.WRITE" },
  MASTER: { READ: "MASTER.READ", WRITE: "MASTER.WRITE" },
  INBOUND: {
    READ: "INBOUND.READ",
    WRITE: "INBOUND.WRITE",
    PLAN: "INBOUND.PLAN",
    APPROVE: "INBOUND.APPROVE",
    RECEIVE: "INBOUND.RECEIVE",
    QC: "INBOUND.QC",
    PUTAWAY: "INBOUND.PUTAWAY",
    ASSIGN: "INBOUND.ASSIGN",
    QUARANTINE_DISPOSE: "INBOUND.QUARANTINE_DISPOSE",
    REWORK: "INBOUND.REWORK",
    CANCEL: "INBOUND.CANCEL",
  },
  INVENTORY: { READ: "INVENTORY.READ", WRITE: "INVENTORY.WRITE" },
  OUTBOUND: { READ: "OUTBOUND.READ", WRITE: "OUTBOUND.WRITE" },
  BILLING: { READ: "BILLING.READ", WRITE: "BILLING.WRITE" },
  REPORTING: { READ: "REPORTING.READ", WRITE: "REPORTING.WRITE" },
} as const;

export type PermissionModule = keyof typeof PERMISSIONS;
type ValueOf<T> = T[keyof T];
type DistributedValueOf<T> = T extends unknown ? T[keyof T] : never;
export type PermissionCode = DistributedValueOf<ValueOf<typeof PERMISSIONS>>;

export interface PermissionRequirement {
  anyOf: readonly PermissionCode[];
}

function requires(permission: PermissionCode): PermissionRequirement {
  return { anyOf: [permission] };
}

export const MODULE_ACCESS = {
  AUTH: {
    READ: requires(PERMISSIONS.AUTH.READ),
    WRITE: requires(PERMISSIONS.AUTH.WRITE),
  },
  SECURITY: {
    READ: requires(PERMISSIONS.SECURITY.READ),
    WRITE: requires(PERMISSIONS.SECURITY.WRITE),
  },
  MASTER: {
    READ: requires(PERMISSIONS.MASTER.READ),
    WRITE: requires(PERMISSIONS.MASTER.WRITE),
  },
  INBOUND: {
    READ: requires(PERMISSIONS.INBOUND.READ),
    WRITE: requires(PERMISSIONS.INBOUND.WRITE),
    PLAN: requires(PERMISSIONS.INBOUND.PLAN),
    APPROVE: requires(PERMISSIONS.INBOUND.APPROVE),
    RECEIVE: requires(PERMISSIONS.INBOUND.RECEIVE),
    QC: requires(PERMISSIONS.INBOUND.QC),
    PUTAWAY: requires(PERMISSIONS.INBOUND.PUTAWAY),
    ASSIGN: requires(PERMISSIONS.INBOUND.ASSIGN),
    QUARANTINE_DISPOSE: requires(PERMISSIONS.INBOUND.QUARANTINE_DISPOSE),
    REWORK: requires(PERMISSIONS.INBOUND.REWORK),
    CANCEL: requires(PERMISSIONS.INBOUND.CANCEL),
  },
  INVENTORY: {
    READ: requires(PERMISSIONS.INVENTORY.READ),
    WRITE: requires(PERMISSIONS.INVENTORY.WRITE),
  },
  OUTBOUND: {
    READ: requires(PERMISSIONS.OUTBOUND.READ),
    WRITE: requires(PERMISSIONS.OUTBOUND.WRITE),
  },
  BILLING: {
    READ: requires(PERMISSIONS.BILLING.READ),
    WRITE: requires(PERMISSIONS.BILLING.WRITE),
  },
  REPORTING: {
    READ: requires(PERMISSIONS.REPORTING.READ),
    WRITE: requires(PERMISSIONS.REPORTING.WRITE),
  },
} as const;

export function hasPermission(
  permissions: readonly string[],
  permission: PermissionCode,
) {
  return (
    permissions.includes(WILDCARD_PERMISSION) ||
    permissions.includes(permission)
  );
}

export function isAuthorized(
  permissions: readonly string[],
  requirement?: PermissionRequirement,
) {
  if (!requirement) {
    return true;
  }

  if (permissions.includes(WILDCARD_PERMISSION)) {
    return true;
  }

  return requirement.anyOf.some((permission) =>
    permissions.includes(permission),
  );
}
