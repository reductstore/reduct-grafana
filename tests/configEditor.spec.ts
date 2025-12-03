import { test, expect } from '@grafana/plugin-e2e';
import { ReductSourceOptions, SecureJsonData } from '../src/types';

test.describe('ReductStore Config Editor', () => {
  test('smoke: should render config editor', async ({
    createDataSourceConfigPage,
    readProvisionedDataSource,
    page,
  }) => {
    const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
    await createDataSourceConfigPage({ type: ds.type });
    await expect(page.getByLabel('URL')).toBeVisible();
  });

  test('"Save & test" should be successful when configuration is valid', async ({
    createDataSourceConfigPage,
    readProvisionedDataSource,
    page,
  }) => {
    const ds = await readProvisionedDataSource<ReductSourceOptions, SecureJsonData>({ fileName: 'datasources.yml' });
    const configPage = await createDataSourceConfigPage({ type: ds.type });
    await page.getByRole('textbox', { name: 'Token' }).fill(ds.secureJsonData?.serverToken ?? '');
    await page.getByRole('textbox', { name: 'URL' }).fill(ds.jsonData.serverURL ?? '');
    await expect(configPage.saveAndTest()).toBeOK();
  });

  test('"Save & test" should fail when configuration is invalid', async ({
    createDataSourceConfigPage,
    readProvisionedDataSource,
    page,
  }) => {
    const ds = await readProvisionedDataSource<ReductSourceOptions, SecureJsonData>({ fileName: 'datasources.yml' });
    const configPage = await createDataSourceConfigPage({ type: ds.type });
    await page.getByRole('textbox', { name: 'Token' }).fill(ds.secureJsonData?.serverToken ?? '');
    await page.getByRole('textbox', { name: 'URL' }).fill('');
    await expect(configPage.saveAndTest()).not.toBeOK();
    await expect(configPage).toHaveAlert('error', { hasText: 'Server URL is missing' });
  });

  test('"Save & test" should fail when authentication fails', async ({
    createDataSourceConfigPage,
    readProvisionedDataSource,
    page,
  }) => {
    const ds = await readProvisionedDataSource<ReductSourceOptions, SecureJsonData>({
      fileName: 'datasources.yml',
    });

    const configPage = await createDataSourceConfigPage({ type: ds.type });

    await page.getByRole('textbox', { name: 'Token' }).fill('invalid-token-123');
    await page.getByRole('textbox', { name: 'URL' }).fill(ds.jsonData.serverURL ?? '');
    await expect(configPage.saveAndTest()).not.toBeOK();
    await expect(configPage).toHaveAlert('error', {
      hasText: 'Authentication failed or server error',
    });
  });
});
