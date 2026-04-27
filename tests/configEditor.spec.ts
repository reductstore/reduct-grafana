import { test, expect } from '@grafana/plugin-e2e';
import { ReductSourceOptions, SecureJsonData } from '../src/types';

const grafanaVersion = process.env.GRAFANA_VERSION ?? '';
const grafanaMajor = Number.parseInt(grafanaVersion.split('.')[0] ?? '', 10);
const skipOnGrafana13PlusOrNightly =
  grafanaVersion === 'nightly' || (Number.isFinite(grafanaMajor) && grafanaMajor >= 13);

test.describe('ReductStore Config Editor', () => {
  test.skip(
    skipOnGrafana13PlusOrNightly,
    'Config page Save & test selectors changed for Grafana 13+/nightly; covered by legacy versions until plugin-e2e update.',
  );
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
    await page.getByLabel(/Token/i).fill(ds.secureJsonData?.serverToken ?? '');
    await page.getByLabel(/URL/i).fill(ds.jsonData.serverURL ?? '');
    await expect(configPage.saveAndTest()).toBeOK();
  });

  test('"Save & test" should fail when configuration is invalid', async ({
    createDataSourceConfigPage,
    readProvisionedDataSource,
    page,
  }) => {
    const ds = await readProvisionedDataSource<ReductSourceOptions, SecureJsonData>({ fileName: 'datasources.yml' });
    const configPage = await createDataSourceConfigPage({ type: ds.type });
    await page.getByLabel(/Token/i).fill(ds.secureJsonData?.serverToken ?? '');
    await page.getByLabel(/URL/i).fill('');
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

    await page.getByLabel(/Token/i).fill('invalid-token-123');
    await page.getByLabel(/URL/i).fill(ds.jsonData.serverURL ?? '');
    await expect(configPage.saveAndTest()).not.toBeOK();
    await expect(configPage).toHaveAlert('error', {
      hasText: 'Authentication failed or server error',
    });
  });

  test('should persist the CA certificate path', async ({
    createDataSourceConfigPage,
    readProvisionedDataSource,
    page,
  }) => {
    const ds = await readProvisionedDataSource<ReductSourceOptions, SecureJsonData>({ fileName: 'datasources.yml' });
    const configPage = await createDataSourceConfigPage({ type: ds.type });
    const caCertPath = '/etc/ssl/certs/ca-certificates.crt';

    await page.getByLabel(/Token/i).fill(ds.secureJsonData?.serverToken ?? '');
    await page.getByLabel(/URL/i).fill(ds.jsonData.serverURL ?? '');
    await page.getByLabel(/CA Certificate Path/i).fill(caCertPath);

    await expect(configPage.saveAndTest()).toBeOK();

    await page.reload();

    await expect(page.getByLabel(/CA Certificate Path/i)).toHaveValue(caCertPath);
  });
});
