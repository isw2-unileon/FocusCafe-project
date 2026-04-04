import * as dotenv from "dotenv";
import { test, expect } from "@playwright/test";
import {resolve} from "path";
//absolute path to .env file
dotenv.config({path: resolve(__dirname, "../../.env") });


test("homepage loads", async ({ page }) => {
  await page.goto("/");
  await expect(page.locator("h1")).toHaveText("App");
});

test("health endpoint responds", async ({ request }) => {
  const port = process.env.PORT || 808;
  const response = await request.get(`http://localhost:${port}/health`);
  expect(response.ok()).toBeTruthy();
  expect(await response.json()).toEqual({ status: "ok" });
});
