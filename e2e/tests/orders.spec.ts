import { test, expect, Page } from "@playwright/test";
import { RegisterAndloginAsUser } from "../lib/auth-helper";

//Helper function to obtain the numeric value of an Stat from StatCard
async function getStatValue(page: Page, statLabel: 'Energy' | 'Experience'): Promise<number> {
  const row = page.locator('div.flex.items-center.gap-4').filter({ hasText: statLabel });
  const text = await row.locator('p.text-sm.font-black').textContent();
  
  if (!text) return 0;
  const currentPart = text.includes('/') ? text.split('/')[0] : text;
  return parseInt(currentPart.trim(), 10) || 0;
}

test.describe("Orders — Personal", () => {
  test("should complete a personal order and decrease energy", async ({ page }) => {
    // 1. Login and land on Home
    await RegisterAndloginAsUser(page);
    await expect(page).toHaveURL(/.*home/);

    // 2. Wait for orders to load
    const pendingOrdersSection = page.locator('div').filter({ has: page.getByRole('heading', { name: 'Pending orders', level: 2 }) }).first();
    await expect(pendingOrdersSection).toBeVisible();

    // 3. Capture energy before
    const energyBefore = await getStatValue(page, 'Energy');
    const xpBefore = await getStatValue(page, 'Experience');

    // 4. Capture the name of the first order before completing it
    const firstOrderCard = pendingOrdersSection.locator('[data-testid^="order-card-"]').first();
    await expect(firstOrderCard).toBeVisible();
    const testId = await firstOrderCard.getAttribute('data-testid');

    // 5. Click the first "Complete" button in Pending orders
    await firstOrderCard.getByRole("button", { name: "Complete" }).click();

    //Close Level Up modal if it appears
    const levelUpModal = page.getByRole('dialog', { name: 'LEVEL UP!' });
     if (await levelUpModal.isVisible({ timeout: 3000 }).catch(() => false)) {
       await page.getByRole('button', { name: 'AWESOME!' }).click();
     }

    // 6. Wait for the order to disappear from the list (robust: works with toast or level-up modal)
    await expect(page.getByTestId(testId!)).not.toBeVisible({ timeout: 10000 });

    // 7. Verify energy decreased
    const energyAfter = await getStatValue(page, 'Energy');
    expect(energyAfter).toBeLessThan(energyBefore);

    // 8. Verify XP is visible and positive
    const xpAfter = await getStatValue(page, 'Experience');
    expect(xpAfter).toBeGreaterThanOrEqual(xpBefore);
  });
});
