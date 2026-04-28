import { test as base, expect } from '@grafana/plugin-e2e';

export const test = base.extend({
  page: async ({ page }, use) => {
    // Auto-dismiss "What's new in Grafana" dialog that appears in Grafana 13+
    page.addLocatorHandler(page.getByRole('dialog', { name: "What's new in Grafana" }), async (dialog) => {
      await dialog.getByRole('button', { name: 'Close' }).click();
    });
    await use(page);
  },
});

export { expect };
