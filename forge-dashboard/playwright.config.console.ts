// Config temporaire : capture console du navigateur dans w1-console.log
import { defineConfig } from "@playwright/test";
export default defineConfig({
  testDir: "./tests",
  reporter: "list",
  use: {
    launchOptions: { args: ["--no-sandbox"] },
  },
});
