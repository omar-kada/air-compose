// Records the UI scene with Playwright -> out/ui.webm
//
// ALL selectors live in SEL below. They are PLACEHOLDERS (data-testid):
// add matching data-testid attributes in frontend/ or edit this table.
import { mkdirSync, rmSync } from 'node:fs';
import { chromium } from 'playwright';

const {
  UI_URL,
  CONFIG_REPO_URL,
  DEMO_USER,
  DEMO_PASSWORD,
  DEMO_HEADED,
} = process.env;

const W = 1280, H = 720;

mkdirSync('out', { recursive: true });
rmSync('out/raw', { recursive: true, force: true });

const browser = await chromium.launch({
  headless: !DEMO_HEADED,
  slowMo: DEMO_HEADED ? 100 : 0,
});
const context = await browser.newContext({
  viewport: { width: W, height: H },
  deviceScaleFactor: 1,
  recordVideo: { dir: 'out/raw', size: { width: W, height: H } },
});

// Playwright doesn't draw a cursor in recordings -> inject a visible one
await context.addInitScript(() => {
  addEventListener('DOMContentLoaded', () => {
    const c = document.createElement('div');
    c.style.cssText =
      'position:fixed;z-index:2147483647;width:20px;height:20px;left:-50px;top:-50px;' +
      'border-radius:50%;background:rgba(255,70,70,.65);border:2px solid #fff;' +
      'pointer-events:none;transform:translate(-50%,-50%)';
    document.body.appendChild(c);
    addEventListener('mousemove', (e) => {
      c.style.left = e.clientX + 'px';
      c.style.top = e.clientY + 'px';
    }, true);
  });
});

const page = await context.newPage();
const byid = (id) => page.locator(`[id=${JSON.stringify(id)}]`);
const byname = (name) => page.locator(`[name=${JSON.stringify(name)}]`);
const bytype = (type) => page.locator(`[type=${JSON.stringify(type)}]`);
const byrole = (role) => page.locator(`[role=${JSON.stringify(role)}]`);
const button = (name) => page.getByRole("button", {name : name})

// glide the cursor to the element, pause briefly, then click
async function click(locator) {
  await locator.waitFor();
  const box = await locator.boundingBox();
  await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2, { steps: 30 });
  await page.waitForTimeout(250);
  await locator.click();
}
// human-ish typing
async function type(locator, text) {
  await click(locator);
  await locator.pressSequentially(text, { delay: 70 });
}

// ---------- scene ----------
await page.goto(UI_URL);
await page.waitForTimeout(1200);

// 1. register (TODO: add a confirm-password step if the form has one)
await type(byname('username'), DEMO_USER);
await type(byname('password'), DEMO_PASSWORD);
await click(bytype('submit'));

// await page.waitForURL('**/login')

// await type(tname('username'), DEMO_USER);
// await type(tname('password'), DEMO_PASSWORD);
// await click(ttype('submit'));

// 2. initial configuration
await page.waitForURL('**/init')

await type(byname('repo'), CONFIG_REPO_URL);
await type(byname('branch'), "main");
await click(byrole('switch'));
await click(page.getByText("Next"));
await page.waitForTimeout(1200);

await click(page.getByText("Daily"))

await click(page.getByText("Next"));
await page.waitForTimeout(1200);

// 3. first deployment

await click(byid('DEPLOYMENT_STARTED'));

await click(page.getByText("Save"));
await page.waitForURL('**/deployments**')
await page.waitForTimeout(1200);
await page.goto(`${UI_URL}/config`)

await click(button('Add service'));
await type(byname("services.0.name"), "homepage")
await type(byname("services.0.envVars.0.key"), "SERVICES_PATH")
await type(byname("services.0.envVars.0.value"), "..")
await type(byname("services.0.envVars.1.key"), "DATA_PATH")
await type(byname("services.0.envVars.1.value"), "./data")
await click(button('Save'));

// await page.goto(`${UI_URL}/deployments`)

const deploymentsLink = page.getByRole('link', { name: /deployments/i });
await click(deploymentsLink.first());
await page.waitForURL('**/deployments');
await page.getByText("Success").first().waitFor({ timeout: 120_000 });
await page.waitForTimeout(3000);

const logsLink = page.getByRole('link', { name: /logs/i });
await click(logsLink.first());
await page.waitForURL('**/logs');
await page.getByText("healthy").first().waitFor({ timeout: 120_000 });
await page.waitForTimeout(1400);

const statusLink = page.getByRole('link', { name: /status/i });
await click(statusLink.first());
await page.waitForURL('**/status');

await page.waitForTimeout(3000);

// ---------- finish ----------
await context.close();                       // finalizes the video
await page.video().saveAs('out/ui.webm');
await browser.close();
rmSync('out/raw', { recursive: true, force: true });
console.log('saved out/ui.webm');
