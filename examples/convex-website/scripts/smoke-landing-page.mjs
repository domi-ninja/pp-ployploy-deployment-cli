#!/usr/bin/env node

import { readFileSync } from "node:fs";
import { chromium } from "playwright";

let envFile = ".env.local";
let url = "";

for (let i = 2; i < process.argv.length; i += 1) {
  const arg = process.argv[i];
  if (arg === "--prod") {
    envFile = ".env.prod";
  } else if (arg === "--env-file") {
    envFile = process.argv[++i];
  } else if (arg === "--url") {
    url = process.argv[++i];
  } else {
    throw new Error(`Unknown argument: ${arg}`);
  }
}

function readEnv(path) {
  const values = {};
  for (const rawLine of readFileSync(path, "utf8").split(/\r?\n/)) {
    const line = rawLine.trim();
    if (!line || line.startsWith("#")) continue;
    const index = line.indexOf("=");
    if (index === -1) continue;
    const key = line.slice(0, index).trim();
    let value = line.slice(index + 1).trim();
    if (
      (value.startsWith('"') && value.endsWith('"')) ||
      (value.startsWith("'") && value.endsWith("'"))
    ) {
      value = value.slice(1, -1);
    }
    values[key] = value;
  }
  return values;
}

const env = readEnv(envFile);
const target = url || env.VITE_SITE_URL || env.SITE_URL;
if (!target) {
  throw new Error(`VITE_SITE_URL or SITE_URL is required in ${envFile}`);
}

const browser = await chromium.launch({ headless: true });
const page = await browser.newPage();
const failures = [];

page.on("pageerror", (error) => {
  failures.push(`pageerror: ${error.message}`);
});

page.on("console", (message) => {
  if (message.type() === "error") {
    failures.push(`console error: ${message.text()}`);
  }
});

const response = await page.goto(target, { waitUntil: "domcontentloaded", timeout: 30_000 });
if (!response) {
  failures.push("no response from landing page");
} else if (!response.ok()) {
  failures.push(`landing page returned HTTP ${response.status()}`);
}

await page.waitForTimeout(1_000);
await browser.close();

if (failures.length > 0) {
  console.error(failures.join("\n"));
  process.exit(1);
}

console.log(`Landing page smoke check passed: ${target}`);
