import { test, expect, Page } from "@playwright/test";
import { RegisterAndloginAsUser } from "../lib/auth-helper";
import path from "path";
import fs from "fs";

async function GetStatValue(page: Page, statLabel: 'Energy' | 'Experience'): Promise<number> {
  const row = page.locator('div.flex.items-center.gap-4').filter({ hasText: statLabel });
  const text = await row.locator('p.text-sm.font-black').textContent();
  if (!text) return 0;
  const currentPart = text.includes('/') ? text.split('/')[0] : text;
  return parseInt(currentPart.trim(), 10) || 0;
}

test.describe("Study Session — Critical User Journey", () => {
  const pdfPath = path.resolve(__dirname, "../fixtures/sample-study.pdf");

  test("should complete the full Study-to-Earn loop: upload, study, quiz, and reward", async ({ page }) => {
    await page.route('**/study/generate-quiz/**', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            id: 123,
            session_id: 5,
            questions: [
              {
                id: 1,
                quiz_id: 123,
                question_text: "What is the primary benefit of network mocking in E2E testing?",
                option_a: "It isolates tests from third-party API downtime",
                option_b: "It increases cloud costs",
                option_c: "It makes tests run slower",
                option_d: "It has no benefits",
                correct_answer: "A",
                explanation: "Network mocking prevents flaky tests caused by third-party service outages."
              }
            ],
            quiz: [
              {
                question: "What is the primary benefit of network mocking in E2E testing?",
                options: ["It isolates tests from third-party API downtime", "It increases cloud costs", "It makes tests run slower", "It has no benefits"],
                correctAnswer: 0
              }
            ]
          })
        });
      } else {
        await route.continue();
      }
    });

    // 1. Authentication: Register a fresh user and land on Home
    await RegisterAndloginAsUser(page);
    await expect(page).toHaveURL(/.*home/, { timeout: 10000 });

    // 2. Snapshot Initial State: Capture starting energy
    const energyBefore = await GetStatValue(page, 'Energy');

    // 3. Navigation: Move to the Study Session page
    const studyButton = page.getByRole('button', { name: /BREW COFFEE \(STUDY\)/i });
    await expect(studyButton).toBeEnabled({ timeout: 10000 });
    await studyButton.click();
    await expect(page).toHaveURL(/.*study/, { timeout: 10000 });

    // 4. Setup Phase: Upload the PDF material
    const fileInput = page.getByTestId('study-file-input');
    await fileInput.setInputFiles(pdfPath);
    
    await expect(page.getByTestId('study-file-name')).toBeVisible({ timeout: 10000 });

    // 5. Execution Phase: Start the study timer
    await page.getByTestId('study-start-button').click();
    await expect(page.getByTestId('study-timer')).toBeVisible({ timeout: 10000 });

    // 6. Transition: Skip the timer to trigger quiz generation immediately
    await page.getByTestId('study-skip-quiz').click();

    // 7. Assessment Phase: Wait for AI quiz generation and container to appear
    await expect(page.getByTestId('quiz-container')).toBeVisible({ timeout: 10000 });

    // 8. Interaction: Answer all generated questions
    const optionButtons = page.locator('button[data-testid^="quiz-option-"]');
    
    // Explicitly wait for the first option button element to build in DOM
    await optionButtons.first().waitFor({ state: 'visible', timeout: 10000 });
    
    // Select the answer option
    await optionButtons.first().click();
    
    // 9. Submission: Finalize the quiz
    await page.getByTestId('quiz-submit-button').click();

    // 10. Verification (Results Page): Ensure results are shown and reward is indicated
    await expect(page.getByText(/Session Complete!/i)).toBeVisible({ timeout: 10000 });
    await expect(page.getByText(/ENERGY EARNED/i)).toBeVisible({ timeout: 10000 });

    // 11. Final Validation: Return home and verify the energy stat has increased
    await page.getByRole('button', { name: /RETURN TO CAFETERIA/i }).click();
    await expect(page).toHaveURL(/.*home/, { timeout: 10000 });

    const energyAfter = await GetStatValue(page, 'Energy');
    expect(energyAfter).toBeGreaterThan(energyBefore);
  });
});