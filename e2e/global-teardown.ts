import { rmSync } from "fs";
import { resolve } from "path";

export default function globalTeardown() {
  const authDir = resolve(process.cwd(), "playwright", ".auth");
  try {
    rmSync(authDir, { recursive: true, force: true });
    console.log("[teardown] Auth state cleaned up.");
  } catch {
    // Ignore if directory doesn't exist
  }
}
