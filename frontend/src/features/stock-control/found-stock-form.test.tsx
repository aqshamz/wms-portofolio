import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { FoundStockForm } from "./found-stock-form";
import { addFoundStock, getFoundStockOptions } from "./stock-control-api";

vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));
vi.mock("./stock-control-api", () => ({
  addFoundStock: vi.fn(),
  getFoundStockOptions: vi.fn(),
}));
vi.mock("@/components/ui/select", () => ({
  Select: ({
    value,
    options,
    onValueChange,
    ariaLabel,
    id,
  }: {
    value: string;
    options: { value: string; label: string }[];
    onValueChange: (value: string) => void;
    ariaLabel: string;
    id: string;
  }) => (
    <select
      id={id}
      aria-label={ariaLabel}
      value={value}
      onChange={(event) => onValueChange(event.target.value)}
    >
      <option value="">Select</option>
      {options.map((option) => (
        <option key={option.value} value={option.value}>
          {option.label}
        </option>
      ))}
    </select>
  ),
}));

beforeEach(() => {
  vi.resetAllMocks();
  vi.mocked(getFoundStockOptions).mockResolvedValue({
    items: [
      {
        item_id: "item-1",
        code: "SERIAL-LOT",
        name: "Serialized lot item",
        uom_code: "EA",
        lot_controlled: true,
        serial_controlled: true,
      },
    ],
    locations: [{ id: "location-1", label: "STORAGE-1" }],
    statuses: [{ id: "available", label: "AVAILABLE" }],
    lots: [{ id: "lot-1", label: "LOT-1" }],
    handling_units: [{ id: "hu-1", label: "PALLET-1" }],
  });
});
afterEach(cleanup);

function mount() {
  const onAdded = vi.fn();
  render(
    <QueryClientProvider
      client={
        new QueryClient({ defaultOptions: { queries: { retry: false } } })
      }
    >
      <FoundStockForm
        count={{ cycle_count_id: "GSO-1", version_no: 2 }}
        onAdded={onAdded}
        onCancel={vi.fn()}
      />
    </QueryClientProvider>,
  );
  return { user: userEvent.setup(), onAdded };
}

async function fill() {
  const result = mount();
  await screen.findByRole("option", {
    name: "SERIAL-LOT · Serialized lot item",
  });
  await result.user.selectOptions(
    screen.getByRole("combobox", { name: "Found stock item" }),
    "item-1",
  );
  await result.user.selectOptions(
    screen.getByRole("combobox", { name: "Found stock location" }),
    "location-1",
  );
  await result.user.type(screen.getByLabelText(/Physical quantity/), "2");
  await result.user.type(
    screen.getByLabelText(/Discovery notes/),
    "Found behind rack",
  );
  return result;
}

it("requires the lot and whole serial units before submitting found stock for review", async () => {
  const { user, onAdded } = await fill();
  const submit = screen.getByRole("button", { name: "Record for review" });
  expect(submit).toBeDisabled();
  await user.selectOptions(
    screen.getByRole("combobox", { name: "Found stock lot" }),
    "lot-1",
  );
  await user.selectOptions(
    screen.getByRole("combobox", { name: "Found stock handling unit" }),
    "hu-1",
  );
  const quantity = screen.getByLabelText(/Physical quantity/);
  await user.clear(quantity);
  await user.type(quantity, "1.5");
  expect(submit).toBeDisabled();
  await user.clear(quantity);
  await user.type(quantity, "2");
  expect(submit).toBeEnabled();
  await user.click(submit);
  expect(addFoundStock).toHaveBeenCalledWith("GSO-1", {
    expected_version: 2,
    item_id: "item-1",
    location_id: "location-1",
    inventory_status_id: "available",
    lot_id: "lot-1",
    handling_unit_id: "hu-1",
    counted_quantity: "2",
    notes: "Found behind rack",
  });
  await waitFor(() => expect(onAdded).toHaveBeenCalledTimes(1));
});

it("keeps the form open and shows a stale-stock error for correction", async () => {
  vi.mocked(addFoundStock).mockRejectedValue(
    new Error("Stock changed; record a new count."),
  );
  const { user, onAdded } = await fill();
  await user.selectOptions(
    screen.getByRole("combobox", { name: "Found stock lot" }),
    "lot-1",
  );
  await user.click(screen.getByRole("button", { name: "Record for review" }));
  expect(await screen.findByRole("alert")).toHaveTextContent(
    "Stock changed; record a new count.",
  );
  expect(onAdded).not.toHaveBeenCalled();
  expect(screen.getByLabelText(/Discovery notes/)).toHaveValue(
    "Found behind rack",
  );
});
