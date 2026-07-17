type GrafanaBootData = {
  settings?: {
    buildInfo?: {
      version?: string;
    };
  };
};

export function getGrafanaMajorVersion(): number {
  const grafanaBootData = (window as typeof window & { grafanaBootData?: GrafanaBootData }).grafanaBootData;
  const majorVersion = parseInt(grafanaBootData?.settings?.buildInfo?.version?.split('.')[0] ?? '', 10);

  return Number.isNaN(majorVersion) ? 0 : majorVersion;
}
