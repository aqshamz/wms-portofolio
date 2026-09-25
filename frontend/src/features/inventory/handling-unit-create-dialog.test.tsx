import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { afterEach, beforeEach, expect, it, vi } from "vitest";

import { listLocations } from "@/features/storage-layout/storage-layout-api";
import { listHandlingUnitTypes } from "@/features/units-packaging/units-packaging-api";
import { createHandlingUnit } from "./inventory-api";
import { HandlingUnitCreateDialog } from "./handling-unit-create-dialog";

vi.mock("sonner", () => ({ toast: { success: vi.fn() } }));
vi.mock(
  "@/features/storage-layout/storage-layout-api",
  async (importOriginal) => ({
    ...(await importOriginal<
      typeof import("@/features/storage-layout/storage-layout-api")
    >()),
    listLocations: vi.fn(),
  }),
);
vi.mock(
  "@/features/units-packaging/units-packaging-api",
  async (importOriginal) => ({
    ...(await importOriginal<
      typeof import("@/features/units-packaging/units-packaging-api")
    >()),
    listHandlingUnitTypes: vi.fn(),
  }),
);
vi.mock("./inventory-api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("./inventory-api")>()),
  createHandlingUnit: vi.fn(),
}));

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(listHandlingUnitTypes).mockResolvedValue({
    items: [
      {
        handling_unit_type_id: "type-1",
        code: "PALLET",
        name: "Pallet",
        is_active: true,
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
        location_id: "location-1",
        warehouse_id: "warehouse-1",
        zone_id: "zone-1",
        zone_code: "RECEIVING",
        location_type_id: "type-receiving",
        code: "RCV-01",
        pick_sequence: 1,
        is_pick_face: false,
        is_locked: false,
        is_active: true,
        created_at: "2026-09-25T00:00:00Z",
      },
    ],
    page: 1,
    page_size: 100,
    total_items: 1,
    total_pages: 1,
  });
  vi.mocked(createHandlingUnit).mockResolvedValue({
    handling_unit_id: "HU-1",
    warehouse_id: "warehouse-1",
    owner_id: "owner-1",
    handling_unit_type_id: "type-1",
    parent_handling_unit_id: null,
    current_location_id: "location-1",
    barcode: "PALLET-BDG-0001",
    is_closed: false,
    positive_balance_count: 0,
    child_count: 0,
    created_at: "2026-09-25T00:00:00Z",
    created_by: "account-1",
  });
});

afterEach(cleanup);

it("registers an empty root handling unit at a warehouse location", async () => {
  const user = userEvent.setup();
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  const onCreated = vi.fn();
  render(
    <QueryClientProvider client={client}>
      <HandlingUnitCreateDialog
        ownerId="owner-1"
        warehouseId="warehouse-1"
        onCreated={onCreated}
        onOpenChange={vi.fn()}
      />
    </QueryClientProvider>,
  );

  await user.type(
    screen.getByRole("textbox", { name: /Barcode/ }),
    "PALLET-BDG-0001",
  );
  const type = screen.getByRole("combobox", { name: "Handling-unit type" });
  await waitFor(() => expect(type).toBeEnabled());
  await user.click(type);
  await user.click(
    await screen.findByRole("option", { name: "PALLET · Pallet" }),
  );
  const location = screen.getByRole("combobox", {
    name: "Handling-unit current location",
  });
  await waitFor(() => expect(location).toBeEnabled());
  await user.click(location);
  await user.click(
    await screen.findByRole("option", { name: "RCV-01 · RECEIVING" }),
  );
  await user.click(
    screen.getByRole("button", { name: "Create handling unit" }),
  );

  await waitFor(() => expect(createHandlingUnit).toHaveBeenCalledTimes(1));
  expect(createHandlingUnit).toHaveBeenCalledWith({
    owner_id: "owner-1",
    warehouse_id: "warehouse-1",
    handling_unit_type_id: "type-1",
    current_location_id: "location-1",
    barcode: "PALLET-BDG-0001",
  });
  expect(onCreated).toHaveBeenCalledWith(
    expect.objectContaining({ handling_unit_id: "HU-1" }),
  );
});
