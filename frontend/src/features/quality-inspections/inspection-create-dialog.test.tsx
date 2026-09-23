import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { getReceipt, listReceipts } from "@/features/receipts/receipt-api";
import type { Receipt } from "@/features/receipts/receipt-types";
import {
  createInspection,
  listReceiptInspections,
} from "./quality-inspection-api";
import { InspectionCreateDialog } from "./inspection-create-dialog";
import type { QualityInspection } from "./quality-inspection-types";

vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));
vi.mock("./quality-inspection-api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("./quality-inspection-api")>()),
  createInspection: vi.fn(),
  listReceiptInspections: vi.fn(),
}));
vi.mock("@/features/receipts/receipt-api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/features/receipts/receipt-api")>()),
  getReceipt: vi.fn(),
  listReceipts: vi.fn(),
}));

const batch = {
  receipt_inventory_id: "batch-1",
  item_id: "item-1",
  source_qty: "1",
  source_uom_id: "case",
  source_uom_code: "CASE",
  base_qty: "12",
  base_uom_id: "ea",
  base_uom_code: "EA",
  received_location_id: "rcv",
  received_location_code: "RCV-01",
  initial_inventory_status_id: "qc",
  initial_inventory_status_code: "QC_PENDING",
  initial_balance_id: "balance-1",
};
const receipt: Receipt = {
  receipt_id: "RCV-1",
  owner_id: "owner-1",
  owner_code: "OWNER",
  warehouse_id: "wh-1",
  warehouse_code: "WH",
  business_date: "2026-09-17",
  received_at: "2026-09-17T09:00:00Z",
  status_code: "COMPLETED",
  version_no: 2,
  created_at: "2026-09-17T09:00:00Z",
  lines: [
    {
      receipt_line_id: "line-1",
      line_no: 1,
      item_id: "item-1",
      item_code: "COFFEE",
      item_name: "Coffee",
      received_qty: "2",
      rejected_qty: "0",
      accepted_qty: "2",
      batched_qty: "2",
      uom_id: "case",
      uom_code: "CASE",
      batches: [
        batch,
        {
          ...batch,
          receipt_inventory_id: "batch-2",
          initial_balance_id: "balance-2",
          lot_number: "LOT-2",
        },
      ],
    },
  ],
};

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(listReceipts).mockResolvedValue({
    items: [receipt],
    page: 1,
    page_size: 10,
    total_items: 1,
    total_pages: 1,
  });
  vi.mocked(getReceipt).mockResolvedValue(receipt);
  vi.mocked(listReceiptInspections).mockResolvedValue([
    { receipt_inventory_id: "batch-1" } as QualityInspection,
  ]);
  vi.mocked(createInspection).mockResolvedValue({
    inspection_id: "QC-1",
  } as QualityInspection);
});
afterEach(cleanup);

function renderDialog() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: 0 } },
  });
  const onCreated = vi.fn();
  render(
    <QueryClientProvider client={client}>
      <InspectionCreateDialog
        ownerId="owner-1"
        warehouseId="wh-1"
        scopeLabel="Owner · Warehouse"
        onOpenChange={vi.fn()}
        onCreated={onCreated}
      />
    </QueryClientProvider>,
  );
  return { client, onCreated };
}

async function setup() {
  const { client, onCreated } = renderDialog();
  const user = userEvent.setup();
  const select = screen.getByRole("combobox", { name: "Completed receipt" });
  await waitFor(() => expect(select).toBeEnabled());
  select.focus();
  await user.keyboard(" ");
  await screen.findByRole("option", { name: /RCV-1/ });
  await user.keyboard("[Enter]");
  await waitFor(() =>
    expect(listReceiptInspections).toHaveBeenCalledWith(
      "owner-1",
      "wh-1",
      "RCV-1",
    ),
  );
  return { user, onCreated, client };
}

describe("starting quality inspection", () => {
  it("excludes inspected batches and uses base quantities when starting QC", async () => {
    const { user, onCreated, client } = await setup();
    const invalidate = vi.spyOn(client, "invalidateQueries");
    expect(listReceipts).toHaveBeenCalledWith(
      expect.objectContaining({
        ownerId: "owner-1",
        warehouseId: "wh-1",
        status: "COMPLETED",
        inspectionEligible: true,
      }),
    );
    const select = screen.getByRole("combobox", { name: "Receipt batch" });
    await waitFor(() => expect(select).toBeEnabled());
    select.focus();
    await user.keyboard(" ");
    expect(
      await screen.findByRole("option", { name: /12 EA.*LOT-2/ }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("option", { name: /batch-1/ }),
    ).not.toBeInTheDocument();
    await user.keyboard("[Enter]");
    await user.click(screen.getByRole("button", { name: "Start inspection" }));
    await waitFor(() =>
      expect(createInspection).toHaveBeenCalledWith({
        receipt_inventory_id: "batch-2",
        notes: undefined,
      }),
    );
    await waitFor(() => expect(onCreated).toHaveBeenCalledWith("QC-1"));
    expect(invalidate).toHaveBeenCalledWith({ queryKey: ["receipts", "list"] });
  });
  it("explains an empty eligible receipt list without allowing inspection", async () => {
    vi.mocked(listReceipts).mockResolvedValue({
      items: [],
      page: 1,
      page_size: 10,
      total_items: 0,
      total_pages: 0,
    });
    renderDialog();
    expect(
      await screen.findByText(/No receipts with uninspected batches match/),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Start inspection" }),
    ).toBeDisabled();
    expect(getReceipt).not.toHaveBeenCalled();
  });
  it("keeps an eligible selection when browsing another receipt page", async () => {
    vi.mocked(listReceipts).mockImplementation(async (filters) => ({
      items:
        filters.page === 1 ? [receipt] : [{ ...receipt, receipt_id: "RCV-2" }],
      page: filters.page,
      page_size: 10,
      total_items: 11,
      total_pages: 2,
    }));
    const { user } = await setup();
    await user.click(screen.getByRole("button", { name: "Next" }));
    await waitFor(() =>
      expect(listReceipts).toHaveBeenCalledWith(
        expect.objectContaining({ page: 2, inspectionEligible: true }),
      ),
    );
    const select = screen.getByRole("combobox", { name: "Completed receipt" });
    select.focus();
    await user.keyboard(" ");
    expect(
      await screen.findByRole("option", { name: /RCV-2/ }),
    ).toBeInTheDocument();
    expect(screen.getByRole("option", { name: /RCV-1/ })).toBeInTheDocument();
    await user.keyboard("[Escape]");
  });
  it("does not reinsert a fully inspected selection across receipt searches", async () => {
    vi.mocked(listReceiptInspections).mockResolvedValue([
      { receipt_inventory_id: "batch-1" } as QualityInspection,
      { receipt_inventory_id: "batch-2" } as QualityInspection,
    ]);
    const { user } = await setup();
    expect(
      await screen.findByText(/No uninspected batches remain/),
    ).toBeInTheDocument();
    vi.mocked(listReceipts).mockResolvedValue({
      items: [],
      page: 1,
      page_size: 10,
      total_items: 0,
      total_pages: 0,
    });
    await user.type(
      screen.getByLabelText("Search completed receipts"),
      "other",
    );
    await screen.findByText(/No receipts with uninspected batches match/);
    const select = screen.getByRole("combobox", { name: "Completed receipt" });
    select.focus();
    await user.keyboard(" ");
    expect(
      screen.queryByRole("option", { name: /RCV-1/ }),
    ).not.toBeInTheDocument();
    await user.keyboard("[Escape]");
    expect(
      screen.getByRole("button", { name: "Start inspection" }),
    ).toBeDisabled();
  });
  it("does not allow starting QC for a receipt outside the selected scope", async () => {
    vi.mocked(getReceipt).mockResolvedValue({
      ...receipt,
      owner_id: "other-owner",
    });
    await setup();
    expect(
      await screen.findByText(/No uninspected batches remain/),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Start inspection" }),
    ).toBeDisabled();
    expect(createInspection).not.toHaveBeenCalled();
  });
});
