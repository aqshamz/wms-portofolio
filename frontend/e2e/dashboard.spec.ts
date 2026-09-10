import { expect, test } from "@playwright/test";

test("dashboard exposes the primary warehouse workflow", async ({ page }) => {
  await page.goto("/");

  await expect(
    page.getByRole("heading", { name: "Good morning, Andi" }),
  ).toBeVisible();
  await expect(page.getByText("Priority work queue")).toBeVisible();
  await expect(page.getByRole("button", { name: "Scan item" })).toBeVisible();
});
