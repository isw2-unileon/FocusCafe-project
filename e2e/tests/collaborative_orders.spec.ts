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
    await expect(groupOrdersA.getByRole('button', { name: 'Complete' }).first()).toBeVisible({ timeout: 10000 });
    await expect(groupOrdersB.getByRole('button', { name: 'Complete' }).first()).toBeVisible({ timeout: 10000 });

    const orderName = await groupOrdersA.locator('span.font-semibold.text-stone-700').first().innerText();
    await expect(groupOrdersB.getByText(orderName)).toBeVisible();

    let errorAlertTriggered = false;

    const handleDialog = async (dialog: any) => {
      console.log(`Alert detected: "${dialog.message()}"`);
      expect(dialog.message()).toContain('Error completing order');
      errorAlertTriggered = true;
      await dialog.accept();
    };

    pageA.on('dialog', handleDialog);
    pageB.on('dialog', handleDialog);

    // Simultaneous completion
    const btnA = groupOrdersA.getByRole('button', { name: 'Complete' }).first();
    const btnB = groupOrdersB.getByRole('button', { name: 'Complete' }).first();

    await Promise.all([
      btnA.dispatchEvent('click'),
      btnB.dispatchEvent('click')
    ]);

    await expect.poll(() => errorAlertTriggered, {
      message: "The loser user should have received an error",
      timeout: 5000
    }).toBeTruthy();

    await expect(groupOrdersA.getByText(orderName)).not.toBeVisible({ timeout: 7000 });
    await expect(groupOrdersB.getByText(orderName)).not.toBeVisible({ timeout: 7000 });

    // Clean up contexts
    await contextA.close();
    await contextB.close();
  });
});