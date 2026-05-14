const { chromium } = require('@playwright/test');

(async () => {
  const browser = await chromium.launch({ headless: true, channel: 'chrome' });
  const page = await browser.newPage();
  page.on('console', msg => console.log('[console]', msg.type(), msg.text()));
  page.on('pageerror', err => console.log('[pageerror]', err.message));
  page.on('requestfailed', req => console.log('[reqfailed]', req.url(), req.failure()?.errorText));
  page.on('response', res => {
    const u = res.url();
    if (u.includes('/api/') || u.includes('/oauth2/auth') || u.includes('/auth/login')) {
      console.log('[resp]', res.status(), u);
    }
  });

  await page.goto('http://100.90.113.29:13000/login', { waitUntil: 'domcontentloaded' });
  console.log('url@load', page.url());
  await page.waitForTimeout(4000);
  console.log('url@4s', page.url());

  const btn = page.getByRole('button', { name: /continue to sign in/i });
  console.log('btn visible', await btn.isVisible().catch(() => false));
  await btn.click({ timeout: 5000 }).catch(e => console.log('click error', e.message));
  await page.waitForTimeout(4000);
  console.log('url@afterClick', page.url());
  await browser.close();
})();
