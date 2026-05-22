import { test, expect } from "@playwright/test";

test.describe("Protected Routes — Unauthenticated", () => {
  test("should redirect unauthenticated users to login", async ({ page }) => {
    const protectedRoutes = ["/home", "/dashboard", "/edit-profile", "/adminDashboard"];

    for (const route of protectedRoutes) {
      await page.goto(route);
      await expect(page).toHaveURL(/.*login/);
    }
  });
});
