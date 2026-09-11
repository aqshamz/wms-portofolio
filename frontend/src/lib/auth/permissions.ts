export const WILDCARD_PERMISSION = "*";

export const PERMISSIONS = {
  AUTH: { READ: "AUTH.READ", WRITE: "AUTH.WRITE" },
  SECURITY: { READ: "SECURITY.READ", WRITE: "SECURITY.WRITE" },
  MASTER: { READ: "MASTER.READ", WRITE: "MASTER.WRITE" },
  INBOUND: { READ: "INBOUND.READ", WRITE: "INBOUND.WRITE" },
  INVENTORY: { READ: "INVENTORY.READ", WRITE: "INVENTORY.WRITE" },
  OUTBOUND: { READ: "OUTBOUND.READ", WRITE: "OUTBOUND.WRITE" },
  BILLING: { READ: "BILLING.READ", WRITE: "BILLING.WRITE" },
  REPORTING: { READ: "REPORTING.READ", WRITE: "REPORTING.WRITE" },
} as const;

export type PermissionModule = keyof typeof PERMISSIONS;
export type PermissionAction = keyof (typeof PERMISSIONS)[PermissionModule];
export type PermissionCode = {
  [
    TModule in PermissionModule
  ]: (typeof PERMISSIONS)[TModule][PermissionAction];
}[PermissionModule];

export interface PermissionRequirement {
  anyOf: readonly PermissionCode[];
}

type ModuleAccessRequirements = {
  [TModule in PermissionModule]: {
    [TAction in PermissionAction]: PermissionRequirement;
  };
};

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
} as const satisfies ModuleAccessRequirements;

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
