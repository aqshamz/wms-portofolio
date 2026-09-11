import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { Select } from "@/components/ui/select";

const options = [
  { value: "all", label: "All statuses" },
  { value: "active", label: "Active" },
  { value: "inactive", label: "Inactive" },
] as const;

describe("Select", () => {
  it("opens its portalled options and reports the selected value", async () => {
    const user = userEvent.setup();
    const onValueChange = vi.fn();

    render(
      <Select
        ariaLabel="Filter by status"
        value="all"
        options={options}
        onValueChange={onValueChange}
      />,
    );

    const trigger = screen.getByRole("combobox", { name: "Filter by status" });
    trigger.focus();
    await user.keyboard(" ");

    expect(
      await screen.findByRole("option", { name: "Active" }),
    ).toBeInTheDocument();

    await user.keyboard("[ArrowDown][Enter]");

    expect(onValueChange).toHaveBeenCalledWith("active");
  });
});
