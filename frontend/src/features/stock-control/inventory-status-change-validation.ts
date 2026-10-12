import Decimal from "decimal.js";
import type {
  Balance,
  SerialState,
} from "@/features/inventory/inventory-types";
import type { InventoryStatus } from "@/features/inventory-classifications/inventory-classification-types";
import type { StockControlReason } from "./stock-control-types";

export const manualStatusCodes = [
  "AVAILABLE",
  "HOLD",
  "DAMAGED",
  "EXPIRED",
] as const;

export function isManualStatus(code: string) {
  return manualStatusCodes.some((status) => status === code);
}

export function statusChangeValidation({
  source,
  target,
  reason,
  quantity,
  serial,
  date,
  reference,
  notes,
}: {
  source: Balance;
  target: InventoryStatus | undefined;
  reason: StockControlReason | undefined;
  quantity: string;
  serial: SerialState | undefined;
  date: string;
  reference: string;
  notes: string;
}): string | undefined {
  if (
    !isManualStatus(source.inventory_status_code) ||
    !["STORAGE", "PICK_FACE"].includes(source.location_type_code ?? "") ||
    source.location_is_locked
  ) {
    return "Only stock in unlocked STORAGE or PICK_FACE locations with a manual inventory status can be changed.";
  }
  if (!target?.is_active || !isManualStatus(target.code))
    return "Select an active target inventory status.";
  if (target.inventory_status_id === source.inventory_status_id)
    return "Target status must differ from the current status.";
  if (!/^(0|[1-9][0-9]{0,13})(\.[0-9]{1,6})?$/.test(quantity.trim()))
    return "Quantity must be a positive number with up to 6 decimals.";
  const requested = new Decimal(quantity.trim());
  if (requested.lte(0) || requested.gt(source.available_qty))
    return `Quantity cannot exceed ${source.available_qty} ${source.uom_code} unreserved stock.`;
  if (
    source.serial_controlled &&
    (!requested.eq(1) || serial?.balance.balance_id !== source.balance_id)
  )
    return "Select one serial from this balance; each command changes exactly 1 base unit.";
  if (!date) return "Business date is required.";
  if (!reference.trim()) return "Status-change reference is required.";
  if (!reason?.is_active) return "Select an active inventory reason.";
  if (reason.requires_note && !notes.trim())
    return `Notes are required for ${reason.name}.`;
}
