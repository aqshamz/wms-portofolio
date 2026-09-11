export interface PermissionRequirement {
  anyOf?: readonly string[];
  anyPrefix?: readonly string[];
}

export function hasPermission(
  permissions: readonly string[],
  permission: string,
) {
  return permissions.includes("*") || permissions.includes(permission);
}

export function isAuthorized(
  permissions: readonly string[],
  requirement?: PermissionRequirement,
) {
  if (!requirement) {
    return true;
  }

  if (permissions.includes("*")) {
    return true;
  }

  return (
    requirement.anyOf?.some((permission) =>
      permissions.includes(permission),
    ) === true ||
    requirement.anyPrefix?.some((prefix) =>
      permissions.some((permission) => permission.startsWith(prefix)),
    ) === true
  );
}
