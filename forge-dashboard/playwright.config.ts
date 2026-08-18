/**
 * ForgeLocal — E2E BOOTSTRAP-RO-01.
 * Les traces, vidéos et captures sont désactivées : elles pourraient contenir
 * une valeur temporaire saisie dans le formulaire de bootstrap.
 */
import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: "./tests",
  timeout: 700_000,
  forbidOnly: true,
  reporter: "line",
  retries: 0,
  use: {
    browserName: "chromium",
    headless: true,
    screenshot: "off",
    trace: "off",
    video: "off",
  },
});
