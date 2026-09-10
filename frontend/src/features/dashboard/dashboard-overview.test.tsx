import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { DashboardOverview } from "@/features/dashboard/dashboard-overview";

describe("DashboardOverview", () => {
  it("shows the operational summary and priority work", () => {
    render(<DashboardOverview />);

    expect(
      screen.getByRole("heading", { name: /good morning, andi/i }),
    ).toBeInTheDocument();
    expect(screen.getByText("Priority work queue")).toBeInTheDocument();
    expect(screen.getByText("IB-260910-0142")).toBeInTheDocument();
  });
});
