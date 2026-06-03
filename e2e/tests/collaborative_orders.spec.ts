import { test, expect, Page } from '@playwright/test';

async function registerUser(page: Page, email: string) {
  await page.goto('/register');
  await page.getByPlaceholder('First Name').fill('Test');
  await page.getByPlaceholder('Last Name').fill('User');
  await page.getByPlaceholder('Email').fill(email);
  await page.getByPlaceholder('Password', { exact: true }).fill('Password123!');
  await page.getByPlaceholder('Confirm Password').fill('Password123!');
  await page.getByRole('button', { name: 'Create my Café' }).click();
  
  await expect(page).toHaveURL(/.*login/);
  await page.getByPlaceholder('Email').fill(email);
  await page.getByPlaceholder('Password', { exact: true }).fill('Password123!');
  await page.getByRole('button', { name: 'Enter the Café' }).click();
  await expect(page).toHaveURL(/.*home/);
}

//Helper function to obtain the numeric value of an Stat from StatCard
async function getStatValue(page: Page, statLabel: 'Energy' | 'Experience'): Promise<number> {
  const row = page.locator('div.flex.items-center.gap-4').filter({ hasText: statLabel });
  const text = await row.locator('p.text-sm.font-black').textContent();
  
  if (!text) return 0;
  const currentPart = text.includes('/') ? text.split('/')[0] : text;
  return parseInt(currentPart.trim(), 10) || 0;
}

test.describe('Collaborative Orders', () => {
  test('two users in a group should see shared orders and handle simultaneous completion', async ({ browser }) => {
    const emailA = `userA.${Date.now()}@example.com`;
    const emailB = `userB.${Date.now()}@example.com`;

    const contextA = await browser.newContext();
    const pageA = await contextA.newPage();
    await registerUser(pageA, emailA);

    // Create group
    const groupName = `TeamA.${Date.now()}`;
    await pageA.getByPlaceholder('New Team').fill(groupName);
    await pageA.getByTitle('Create Team').click();
    
    // Wait for the modal and get the code
    await expect(pageA.getByText('Team Created!')).toBeVisible();
    const inviteCode = await pageA.locator('span.font-mono.font-black').innerText();
    await pageA.keyboard.press('Escape'); // Close modal

    const doneButton = pageA.getByRole('button', { name: 'Done!' });
    await doneButton.click();

    await expect(pageA.locator('.fixed.inset-0')).not.toBeVisible();
    await expect(pageA.getByText('Team Created!')).not.toBeVisible();

    const contextB = await browser.newContext();
    const pageB = await contextB.newPage();
    await registerUser(pageB, emailB);

    // Join group
    await pageB.getByPlaceholder('Invite Code').fill(inviteCode);
    await pageB.getByTitle('Join Team').click();
    await expect(pageB.getByText(groupName, { exact: true })).toBeVisible();

    const groupOrdersA = pageA.locator('div.w-full:has(h2:text("Group orders"))');
    const groupOrdersB = pageB.locator('div.w-full:has(h2:text("Group orders"))');

    // Wait for at least one order to appear in both
    const firstCardA = groupOrdersA.locator('[data-testid^="order-card-"]').first();
    const firstCardB = groupOrdersB.locator('[data-testid^="order-card-"]').first();
    await expect(firstCardA).toBeVisible({ timeout: 10000 });
    await expect(firstCardB).toBeVisible({ timeout: 10000 });

    const targetTestId = await firstCardA.getAttribute('data-testid');
    expect(targetTestId).toBeTruthy();

    //Capture energy and xp before completing
    const energyBeforeA = await getStatValue(pageA, 'Energy');
    const energyBeforeB = await getStatValue(pageB, 'Energy');
    const xpBeforeA = await getStatValue(pageA, 'Experience');
    const xpBeforeB = await getStatValue(pageB, 'Experience');

    const rawOrderName = await groupOrdersA.locator('span.font-semibold.text-stone-700').first().textContent();
    const orderName = rawOrderName ? rawOrderName.trim() : '';
    
    let loserPage: Page | null = null;

    const handleDialogA = async (dialog: any) => {
      console.log(`User A lost the race: "${dialog.message()}"`);
      loserPage = pageA;
      await dialog.accept();
    };

    const handleDialogB = async (dialog: any) => {
      console.log(`User B lost the race: "${dialog.message()}"`);
      loserPage = pageB;
      await dialog.accept();
    };

    pageA.on('dialog', handleDialogA);
    pageB.on('dialog', handleDialogB);

    // Simultaneous completion
    const btnA = firstCardA.getByRole('button', { name: 'Complete' });
    const btnB = groupOrdersB.locator(`[data-testid="${targetTestId}"]`).getByRole('button', { name: 'Complete' });

    await Promise.all([
      btnA.click({ force: true }),
      btnB.click({ force: true })
    ]);

    await expect.poll(() => loserPage !== null, {
      message: "The loser user should have received an error",
      timeout: 5000
    }).toBeTruthy();

    const winnerPage = loserPage === pageA ? pageB : pageA;

    await expect(pageA.getByTestId(targetTestId!)).not.toBeVisible({ timeout: 10000 });
    await expect(pageB.getByTestId(targetTestId!)).not.toBeVisible({ timeout: 10000 });

    await pageA.waitForTimeout(1000);

    const energyAfterA = await getStatValue(pageA, 'Energy');
    const energyAfterB = await getStatValue(pageB, 'Energy');
    const xpAfterA = await getStatValue(pageA, 'Experience');
    const xpAfterB = await getStatValue(pageB, 'Experience');

    //Check Energy
    if (winnerPage === pageA) {
      expect(energyAfterA).toBeLessThan(energyBeforeA);
      expect(energyAfterB).toEqual(energyBeforeB);
    } else {
      expect(energyAfterB).toBeLessThan(energyBeforeB);
      expect(energyAfterA).toEqual(energyBeforeA);
    }

    //Check XP
    expect(xpAfterA).toBeGreaterThan(xpBeforeA);
    expect(xpAfterB).toBeGreaterThan(xpBeforeB);

    // Clean up contexts
    await contextA.close();
    await contextB.close();
  });
});