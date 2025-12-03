import type { Page } from '@playwright/test';

export async function getGrafanaMajorVersion(page: Page): Promise<number> {
  const resp = await page.request.get('/api/health');
  if (!resp.ok()) {
    throw new Error(`Failed to get Grafana version: ${resp.status()} ${resp.statusText()}`);
  }

  const json = await resp.json();
  const version: string = json.version ?? '0.0.0';
  const major = parseInt(version.split('.')[0], 10);
  return Number.isNaN(major) ? 0 : major;
}
