import { config } from '@grafana/runtime';

export function getGrafanaMajorVersion(): number {
  return parseInt(config.buildInfo.version.split('.')[0], 10);
}
