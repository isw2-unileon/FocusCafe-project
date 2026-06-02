import { test, expect } from "@playwright/test";
import { loginAsUser } from "../lib/auth-helper";

test.describe("Orders — Personal", () => {
  test("should complete a personal order and decrease energy", async ({ page }) => {
    // 1. Login and land on Home
    await loginAsUser(page);
    await expect(page).toHaveURL(/.*home/);

    // 2. Wait for orders to load
    await expect(page.getByText("Pending orders")).toBeVisible();

    // 3. Capture energy before
    const energyLocator = page.locator('[data-testid="stat-energy-value"]');
    await expect(energyLocator).toBeVisible();
    const energyBeforeText = await energyLocator.textContent();
    const energyBefore = parseInt(energyBeforeText!.trim().split(" ")[0]);
    expect(energyBefore).toBeGreaterThan(0);

    // 4. Capture the name of the first order before completing it
    const firstOrderName = await page.locator('span.font-semibold.text-stone-700').first().textContent();
    expect(firstOrderName).toBeTruthy();

    // 5. Click the first "Complete" button in Pending orders
    const completeButton = page.getByRole("button", { name: "Complete" }).first();
    await expect(completeButton).toBeVisible();
    await completeButton.click();

    // 6. Wait for the order to disappear from the list (robust: works with toast or level-up modal)
    await expect(page.getByText(firstOrderName!)).not.toBeVisible({ timeout: 10000 });

    // 7. Verify energy decreased
    const energyAfterText = await energyLocator.textContent();
    const energyAfter = parseInt(energyAfterText!.trim().split(" ")[0]);
    expect(energyAfter).toBeLessThan(energyBefore);

    // 8. Verify XP is visible and positive
    const xpLocator = page.locator('[data-testid="stat-experience-value"]');
    const xpAfterText = await xpLocator.textContent();
    const xpAfter = parseInt(xpAfterText!.trim().split(" ")[0]);
    expect(xpAfter).toBeGreaterThanOrEqual(0);
  });
});
