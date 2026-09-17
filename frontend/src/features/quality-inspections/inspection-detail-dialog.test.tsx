import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { toast } from "sonner";
import { ApiError } from "@/lib/api/client";
import {
  listLocations,
  listLocationTypes,
} from "@/features/storage-layout/storage-layout-api";
import {
  cancelInspection,
  completeInspection,
  getInspection,
} from "./quality-inspection-api";
import { InspectionDetailDialog } from "./inspection-detail-dialog";
import type { QualityInspection } from "./quality-inspection-types";

vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));
vi.mock("./quality-inspection-api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("./quality-inspection-api")>()),
  getInspection: vi.fn(),
  completeInspection: vi.fn(),
  cancelInspection: vi.fn(),
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

const inspection: QualityInspection = {
  inspection_id: "QC-1",
  receipt_inventory_id: "batch-1",
  receipt_id: "RCV-1",
  source_balance_id: "balance-1",
  source_balance_version_no: 7,
  base_uom_code: "EA",
  owner_id: "owner-1",
  warehouse_id: "wh-1",
  item_id: "item-1",
  item_code: "COFFEE",
  location_code: "RCV-01",
  quality_status_code: "PENDING",
  inspected_qty: "10",
  passed_qty: "0",
  failed_qty: "0",
  version_no: 3,
  created_at: "2026-09-17T09:00:00Z",
};
const completed: QualityInspection = {
  ...inspection,
  inspected_at: "2026-09-17T10:00:00Z",
  quality_status_code: "FAILED",
  inspection_result_code: "REJECTED",
  passed_qty: "0",
  failed_qty: "10",
  version_no: 4,
  quarantine_case: {
    quarantine_case_id: "Q-1",
    quarantine_qty: "10",
    status_code: "OPEN",
  },
};

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(getInspection).mockResolvedValue(inspection);
  vi.mocked(completeInspection).mockImplementation(async () => {
    vi.mocked(getInspection).mockResolvedValue(completed);
    return completed;
  });
  vi.mocked(listLocations).mockResolvedValue({
    items: [
      {
        location_id: "target-1",
        warehouse_id: "wh-1",
        zone_id: "zone-1",
        location_type_id: "storage",
        code: "BULK-01",
        is_active: true,
        is_locked: false,
        is_pick_face: false,
        pick_sequence: 0,
        created_at: "2026-09-17",
      },
      {
        location_id: "dock-1",
        warehouse_id: "wh-1",
        zone_id: "zone-1",
        location_type_id: "dock",
        code: "DOCK-01",
        is_active: true,
        is_locked: false,
        is_pick_face: false,
        pick_sequence: 0,
        created_at: "2026-09-17",
      },
    ],
    page: 1,
    page_size: 100,
    total_items: 2,
    total_pages: 1,
  });
  vi.mocked(listLocationTypes).mockResolvedValue([
    {
      location_type_id: "storage",
      code: "BULK",
      name: "Bulk",
      is_active: true,
      allows_storage: true,
      allows_receiving: false,
      allows_shipping: false,
      allows_picking: false,
    },
    {
      location_type_id: "dock",
      code: "DOCK",
      name: "Dock",
      is_active: true,
      allows_storage: false,
      allows_receiving: true,
      allows_shipping: true,
      allows_picking: false,
    },
  ]);
});
afterEach(cleanup);

function setup(canQC = true, canCancel = true) {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: 0 } },
  });
  const onSelect = vi.fn();
  render(
    <QueryClientProvider client={client}>
      <InspectionDetailDialog
        inspectionId="QC-1"
        canQC={canQC}
        canCancel={canCancel}
        onSelect={onSelect}
        onOpenChange={vi.fn()}
      />
    </QueryClientProvider>,
  );
  return { user: userEvent.setup(), onSelect, client };
}

describe("QC result workflow", () => {
  it("blocks split results for an indivisible batch", async () => {
    vi.mocked(getInspection).mockResolvedValue({
      ...inspection,
      is_indivisible: true,
    });
    const { user } = setup();
    const passed = await screen.findByLabelText(/Passed quantity \(EA\)/);
    await user.clear(passed);
    await user.type(passed, "8");
    const failed = screen.getByLabelText(/Failed quantity \(EA\)/);
    await user.clear(failed);
    await user.type(failed, "2");
    await user.click(
      screen.getByRole("button", { name: "Complete inspection" }),
    );
    expect(await screen.findByText(/they cannot be split/)).toBeInTheDocument();
    expect(completeInspection).not.toHaveBeenCalled();
  });
  it("cannot submit without the stock version", async () => {
    vi.mocked(getInspection).mockResolvedValue({
      ...inspection,
      source_balance_version_no: undefined,
    });
    setup();
    expect(
      await screen.findByText(/Stock version is unavailable/),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Complete inspection" }),
    ).toBeDisabled();
    expect(completeInspection).not.toHaveBeenCalled();
  });
  it("reports an invalid total inline and with a toast before completing a rejection", async () => {
    const { user } = setup();
    await user.click(await screen.findByRole("button", { name: "Fail all" }));
    const failed = screen.getByLabelText(/Failed quantity \(EA\)/);
    await user.clear(failed);
    await user.type(failed, "9");
    await user.click(
      screen.getByRole("button", { name: "Complete inspection" }),
    );
    expect(
      await screen.findByText(/Passed \+ failed must equal 10/),
    ).toBeInTheDocument();
    expect(toast.error).toHaveBeenCalled();
    expect(completeInspection).not.toHaveBeenCalled();
    await user.clear(failed);
    await user.type(failed, "10");
    await user.click(
      screen.getByRole("button", { name: "Complete inspection" }),
    );
    await user.click(
      await screen.findByRole("button", { name: "Confirm completion" }),
    );
    await waitFor(() =>
      expect(completeInspection).toHaveBeenCalledWith(
        "QC-1",
        expect.objectContaining({
          expected_version: 3,
          expected_balance_version: 7,
          passed_qty: "0",
          failed_qty: "10",
          putaway_target_location_id: undefined,
        }),
      ),
    );
    expect(
      await screen.findByText("Quarantine case created"),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Complete inspection" }),
    ).not.toBeInTheDocument();
  });
  it("requires a storage target and excludes the dock", async () => {
    const { user } = setup();
    await user.click(
      await screen.findByRole("button", { name: "Complete inspection" }),
    );
    expect(
      await screen.findByText("Select a storage location for passed stock."),
    ).toBeInTheDocument();
    const select = screen.getByRole("combobox", { name: "Putaway target" });
    await waitFor(() => expect(select).toBeEnabled());
    select.focus();
    await user.keyboard(" ");
    expect(
      await screen.findByRole("option", { name: "BULK-01" }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("option", { name: "DOCK-01" }),
    ).not.toBeInTheDocument();
    await user.keyboard("[Enter]");
    await user.click(
      screen.getByRole("button", { name: "Complete inspection" }),
    );
    await user.click(
      await screen.findByRole("button", { name: "Confirm completion" }),
    );
    await waitFor(() =>
      expect(completeInspection).toHaveBeenCalledWith(
        "QC-1",
        expect.objectContaining({
          passed_qty: "10",
          failed_qty: "0",
          putaway_target_location_id: "target-1",
        }),
      ),
    );
  });
  it("keeps a concurrency error visible without displaying request IDs", async () => {
    vi.mocked(completeInspection).mockRejectedValue(
      new ApiError(
        "Stock changed; reload inspection",
        409,
        undefined,
        "internal-request-id",
      ),
    );
    const { user } = setup();
    await user.click(await screen.findByRole("button", { name: "Fail all" }));
    await user.click(
      screen.getByRole("button", { name: "Complete inspection" }),
    );
    await user.click(
      await screen.findByRole("button", { name: "Confirm completion" }),
    );
    expect(await screen.findByRole("alert")).toHaveTextContent(
      "Stock changed; reload inspection",
    );
    expect(screen.queryByText(/internal-request-id/)).not.toBeInTheDocument();
  });
  it("hides QC and cancellation controls for a read-only account", async () => {
    setup(false, false);
    expect(
      await screen.findByText(/INBOUND.QC is required/),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Complete inspection" }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Cancel inspection" }),
    ).not.toBeInTheDocument();
    expect(listLocations).not.toHaveBeenCalled();
  });
  it("validates the cancel reason and links to the automatically created replacement", async () => {
    const cancelled = {
      ...inspection,
      cancelled_at: "2026-09-17T10:00:00Z",
      quality_status_code: "WAIVED",
      replacement_inspection_id: "QC-2",
      cancellation_reason: "Wrong sampling plan",
      version_no: 4,
    };
    vi.mocked(cancelInspection).mockImplementation(async () => {
      vi.mocked(getInspection).mockResolvedValue(cancelled);
      return cancelled;
    });
    const { user, onSelect } = setup(false, true);
    await user.click(
      await screen.findByRole("button", { name: "Cancel inspection" }),
    );
    await user.click(
      screen.getByRole("button", { name: "Confirm cancellation" }),
    );
    expect(
      await screen.findByText("A cancellation reason is required."),
    ).toBeInTheDocument();
    expect(cancelInspection).not.toHaveBeenCalled();
    await user.type(
      screen.getByLabelText(/Cancellation reason/),
      "Wrong sampling plan",
    );
    await user.click(
      screen.getByRole("button", { name: "Confirm cancellation" }),
    );
    await waitFor(() =>
      expect(cancelInspection).toHaveBeenCalledWith(
        "QC-1",
        3,
        "Wrong sampling plan",
      ),
    );
    expect(
      await screen.findByText(/QC has not been skipped/),
    ).toBeInTheDocument();
    await user.click(
      screen.getByRole("button", { name: "Open replacement inspection" }),
    );
    expect(onSelect).toHaveBeenCalledWith("QC-2");
  });
});
