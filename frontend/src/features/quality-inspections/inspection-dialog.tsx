"use client";

import {
  OperationDialog,
  type OperationDialogProps,
} from "@/components/ui/operation-dialog";

export function InspectionDialog(
  props: Omit<OperationDialogProps, "closeLabel">,
) {
  return <OperationDialog {...props} closeLabel="Close quality inspection" />;
}
