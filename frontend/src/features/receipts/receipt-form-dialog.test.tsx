import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { toast } from "sonner";

import {
  getInboundOrder,
  listInboundOrders,
} from "@/features/inbound-orders/inbound-order-api";
import type { InboundOrder } from "@/features/inbound-orders/inbound-order-types";
import { listItems } from "@/features/item-catalog/item-catalog-api";
import { createReceipt, updateReceipt } from "@/features/receipts/receipt-api";
import { ReceiptFormDialog } from "@/features/receipts/receipt-form-dialog";
import type { Receipt } from "@/features/receipts/receipt-types";
import {
  listLocations,
  listLocationTypes,
} from "@/features/storage-layout/storage-layout-api";

vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));
vi.mock(
  "@/features/inbound-orders/inbound-order-api",
  async (importOriginal) => ({
    ...(await importOriginal<
      typeof import("@/features/inbound-orders/inbound-order-api")
    >()),
    getInboundOrder: vi.fn(),
    listInboundOrders: vi.fn(),
  }),
);
vi.mock("@/features/item-catalog/item-catalog-api", async (importOriginal) => ({
  ...(await importOriginal<
    typeof import("@/features/item-catalog/item-catalog-api")
  >()),
  listItems: vi.fn(),
}));
vi.mock(
  "@/features/storage-layout/storage-layout-api",
  async (importOriginal) => ({
    ...(await importOriginal<
      typeof import("@/features/storage-layout/storage-layout-api")
    >()),
    listLocations: vi.fn(),
    listLocationTypes: vi.fn(),
  }),
);
vi.mock("@/features/receipts/receipt-api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/features/receipts/receipt-api")>()),
  createReceipt: vi.fn(),
  updateReceipt: vi.fn(),
}));

const order: InboundOrder = {
  inbound_id: "INB-100",
  purchase_order_id: "PO-100",
  owner_id: "owner-1",
  owner_code: "OWNER",
  warehouse_id: "warehouse-1",
  warehouse_code: "WH",
  vendor_id: "vendor-1",
  vendor_code: "VENDOR",
  vendor_name: "Supplier",
  business_date: "2026-09-16",
  status_code: "RELEASED",
  version_no: 1,
  created_at: "2026-09-16T09:00:00Z",
  lines: [
    {
      inbound_line_id: "INB-100-L0001",
      line_no: 1,
      item_id: "item-1",
      item_code: "ITEM",
      item_name: "Coffee",
      expected_qty: "10",
      completed_receipt_qty: "0",
      uom_id: "ea",
      uom_code: "EA",
    },
  ],
};
const receipt: Receipt = {
  receipt_id: "RCV-100",
  inbound_id: order.inbound_id,
  owner_id: "owner-1",
  owner_code: "OWNER",
  warehouse_id: "warehouse-1",
  warehouse_code: "WH",
  business_date: "2026-09-16",
  received_at: "2026-09-16T09:00:00Z",
  dock_location_id: "dock-1",
  status_code: "OPEN",
  version_no: 1,
  created_at: "2026-09-16T09:00:00Z",
  lines: [
    {
      receipt_line_id: "RCV-100-L0001",
      inbound_line_id: "INB-100-L0001",
      line_no: 1,
      item_id: "item-1",
      item_code: "ITEM",
      item_name: "Coffee",
      received_qty: "10",
      rejected_qty: "0",
      accepted_qty: "10",
      batched_qty: "10",
      uom_id: "ea",
      uom_code: "EA",
      batches: [
        {
          receipt_inventory_id: "batch-1",
          item_id: "item-1",
          source_qty: "10",
          source_uom_id: "ea",
          source_uom_code: "EA",
          base_qty: "10",
          base_uom_id: "ea",
          base_uom_code: "EA",
          received_location_id: "dock-1",
          received_location_code: "DOCK-01",
          initial_inventory_status_id: "qc",
          initial_inventory_status_code: "QC_PENDING",
        },
      ],
    },
  ],
};

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(getInboundOrder).mockResolvedValue(order);
  vi.mocked(listInboundOrders).mockResolvedValue({
    items: [order],
    page: 1,
    page_size: 100,
    total_items: 1,
    total_pages: 1,
  });
  vi.mocked(listItems).mockResolvedValue({
    items: [
      {
        item_id: "item-1",
        owner_id: "owner-1",
        code: "ITEM",
        name: "Coffee",
        base_uom_id: "ea",
        lot_controlled: false,
        serial_controlled: false,
        is_active: true,
        created_at: "2026-09-16",
        updated_at: "2026-09-16",
      },
    ],
    page: 1,
    page_size: 100,
    total_items: 1,
    total_pages: 1,
  });
  vi.mocked(listLocations).mockResolvedValue({
    items: [
      {
        location_id: "dock-1",
        warehouse_id: "warehouse-1",
        zone_id: "zone-1",
        location_type_id: "dock-type",
        code: "DOCK-01",
        is_active: true,
        is_locked: false,
        is_pick_face: false,
        pick_sequence: 0,
        created_at: "2026-09-16",
      },
    ],
    page: 1,
    page_size: 100,
    total_items: 1,
    total_pages: 1,
  });
  vi.mocked(listLocationTypes).mockResolvedValue([
    {
      location_type_id: "dock-type",
      code: "DOCK",
      name: "Dock",
      is_active: true,
      allows_receiving: true,
      allows_shipping: true,
      allows_storage: false,
      allows_picking: false,
    },
  ]);
  vi.mocked(createReceipt).mockResolvedValue(receipt);
  vi.mocked(updateReceipt).mockResolvedValue(receipt);
});
afterEach(cleanup);

describe("ReceiptFormDialog batch validation", () => {
  it.each([
    { mode: "create", batchQty: "9" },
    { mode: "edit", batchQty: "9" },
    { mode: "create", batchQty: "11" },
    { mode: "edit", batchQty: "11" },
  ] as const)(
    "shows the mismatch for batch total $batchQty on $mode, then saves after correction",
    async ({ mode, batchQty }) => {
      const user = userEvent.setup();
      const queryClient = new QueryClient({
        defaultOptions: { queries: { retry: false, gcTime: 0 } },
      });
      const onOpenChange = vi.fn();
      render(
        <QueryClientProvider client={queryClient}>
          <ReceiptFormDialog
            ownerId="owner-1"
            warehouseId="warehouse-1"
            scopeLabel="Owner · Warehouse"
            receipt={mode === "edit" ? receipt : undefined}
            onOpenChange={onOpenChange}
          />
        </QueryClientProvider>,
      );
      if (mode === "create") {
        const orderSelect = screen.getByRole("combobox", {
          name: "Source Inbound Order",
        });
        orderSelect.focus();
        await user.keyboard(" ");
        await screen.findByRole("option", { name: /INB-100/ });
        await user.keyboard("[Enter]");
        const dockSelect = screen.getByRole("combobox", {
          name: "Receiving dock",
        });
        await waitFor(() => expect(dockSelect).toBeEnabled());
        dockSelect.focus();
        await user.keyboard(" ");
        await screen.findByRole("option", { name: "DOCK-01" });
        await user.keyboard("[Enter]");
      }
      const quantity = await screen.findByLabelText(/Batch quantity/);
      await user.clear(quantity);
      await user.type(quantity, batchQty);
      expect(screen.getByText(/Quantities do not match/)).toBeInTheDocument();
      const saveButton = screen.getByRole("button", {
        name: mode === "create" ? "Open receipt" : "Save draft",
      });
      await waitFor(() => expect(saveButton).toBeEnabled());
      await user.click(saveButton);
      expect(
        await screen.findByText(
          /Batch quantities must equal the accepted quantity/,
        ),
      ).toBeInTheDocument();
      expect(toast.error).toHaveBeenCalledWith("Receipt could not be saved.", {
        description:
          "Line 1: Batch quantities must equal the accepted quantity.",
      });
      expect(createReceipt).not.toHaveBeenCalled();
      expect(updateReceipt).not.toHaveBeenCalled();
      await user.clear(quantity);
      await user.type(quantity, "10");
      expect(
        screen.getByText(/Batch total: 10\.000000.*Balanced/),
      ).toBeInTheDocument();
      await user.click(saveButton);
      await waitFor(() =>
        expect(
          mode === "create" ? createReceipt : updateReceipt,
        ).toHaveBeenCalledTimes(1),
      );
      await waitFor(() => expect(onOpenChange).toHaveBeenCalledWith(false));
      queryClient.clear();
    },
  );
});
