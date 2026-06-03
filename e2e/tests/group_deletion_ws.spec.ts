import { test, expect, Page } from '@playwright/test';
import { RegisterAndloginAsUser } from '../lib/auth-helper';

test.describe('Group Deletion WebSocket', () => {
  test('notifies group members in real-time when the leader deletes the group', async ({ browser }) => {
    // --- Setup: Create two independent browser contexts ---
    const contextA = await browser.newContext();
    const pageA = await contextA.newPage();

    const contextB = await browser.newContext();
    const pageB = await contextB.newPage();

    // --- User A (Leader): Register, login and create a group ---
    await RegisterAndloginAsUser(pageA);

    const groupName = `TeamWS.${Date.now()}`;
    await pageA.getByPlaceholder('New Team').fill(groupName);
    await pageA.getByTitle('Create Team').click();

    await expect(pageA.getByText('Team Created!')).toBeVisible({ timeout: 10000 });
    const inviteCode = await pageA.locator('span.font-mono.font-black').innerText();

    // Close the invite modal
    await pageA.getByRole('button', { name: 'Done!' }).click();
    await expect(pageA.locator('.fixed.inset-0')).not.toBeVisible({ timeout: 10000 });

    // Verify the leader sees the group in the header
    await expect(pageA.getByTestId('group-name')).toHaveText(groupName, { timeout: 5000 });

    // --- User B (Member): Register, login and join the group ---
    await RegisterAndloginAsUser(pageB);

    await pageB.getByPlaceholder('Invite Code').fill(inviteCode);
    await pageB.getByTitle('Join Team').click();
    await expect(pageB.getByText(groupName, { exact: true })).toBeVisible({ timeout: 10000 });

    // Verify the member sees the group in the header
    await expect(pageB.getByTestId('group-name')).toHaveText(groupName, { timeout: 5000 });

    // --- Leader deletes the group ---
    pageA.on('dialog', async (dialog) => {
      await dialog.accept();
    });

    await pageA.getByTitle('Delete Team').click();

    // Leader sees success toast and group disappears from UI
    await expect(pageA.getByText('Team deleted')).toBeVisible({ timeout: 10000 });
    await expect(pageA.getByTestId('group-display')).not.toBeVisible({ timeout: 10000 });

    // --- Member receives real-time WebSocket notification ---
    await expect(pageB.getByText('Your team has been deleted.')).toBeVisible({ timeout: 10000 });
    await expect(pageB.getByTestId('group-display')).not.toBeVisible({ timeout: 10000 });

    // Verify the member now sees the join/create inputs again
    await expect(pageB.getByPlaceholder('Invite Code')).toBeVisible({ timeout: 5000 });
    await expect(pageB.getByPlaceholder('New Team')).toBeVisible({ timeout: 5000 });

    // --- Cleanup ---
    await contextA.close();
    await contextB.close();
  });
});
