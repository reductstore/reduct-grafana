import React, { useEffect, useState } from 'react';
import { Button, Spinner, Stack, useStyles2, Tooltip, Icon } from '@grafana/ui';
import { getBackendSrv } from '@grafana/runtime';
import { css } from '@emotion/css';
import { GrafanaTheme2 } from '@grafana/data';
import { ReductQuery, ServerInfo } from '../types';
import { DataSource } from '../datasource';

interface QueryHeaderProps {
  query: ReductQuery;
  datasource: DataSource;
  onRunQuery: () => void;
}

export function QueryHeader({ query, datasource, onRunQuery }: QueryHeaderProps) {
  const styles = useStyles2(getStyles);
  const [serverInfo, setServerInfo] = useState<ServerInfo | null>(null);
  const [loading, setLoading] = useState(true);

  const isQueryRunnable = !!(query.bucket && query.entry);

  useEffect(() => {
    setLoading(true);
    getBackendSrv()
      .get(`/api/datasources/${datasource.id}/resources/serverInfo`, undefined, undefined, {
        showErrorAlert: false,
      })
      .then((info: ServerInfo) => {
        setServerInfo(info);
      })
      .catch((error) => {
        console.warn('Failed to fetch server info:', error);
        setServerInfo(null);
      })
      .finally(() => {
        setLoading(false);
      });
  }, [datasource.id]);

  const formatBytes = (bytes?: number): string => {
    if (!bytes) {
      return 'N/A';
    }
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(1000));
    return Math.round(bytes / Math.pow(1000, i)) + ' ' + sizes[i];
  };

  const formatDate = (timestamp?: number): string => {
    if (!timestamp || timestamp === 0) {
      return 'N/A';
    }
    const date = new Date(timestamp / 1000);

    if (isNaN(date.getTime())) {
      return 'Invalid';
    }

    return date.toLocaleString();
  };

  return (
    <div className={styles.header}>
      <Stack direction="row" justifyContent="space-between" alignItems="center">
        <div className={styles.info}>
          {loading ? (
            <Stack direction="row" gap={0.5} alignItems="center">
              <Spinner size="xs" />
              <span>Loading server info...</span>
            </Stack>
          ) : serverInfo ? (
            <Stack direction="row" gap={2} alignItems="center" wrap="wrap">
              <span className={styles.infoItem}>
                <strong>Buckets:</strong> {serverInfo.bucket_count}
              </span>
              <span className={styles.infoItem}>
                <strong>Usage:</strong> {formatBytes(serverInfo.usage)}
              </span>

              {/* Data Range Info with Tooltip */}
              {(serverInfo.oldest_record || serverInfo.latest_record) && (
                <Tooltip
                  content={
                    <div>
                      <div>
                        <strong>Oldest record:</strong> {formatDate(serverInfo.oldest_record)}
                      </div>
                      <div>
                        <strong>Latest record:</strong> {formatDate(serverInfo.latest_record)}
                      </div>
                    </div>
                  }
                >
                  <span className={styles.infoItemWithTooltip}>
                    <Icon name="calendar-alt" size="xs" />
                    <span>Data range</span>
                  </span>
                </Tooltip>
              )}

              {/* Version & License Info with Tooltip */}
              <Tooltip
                content={
                  <div>
                    <div>
                      <strong>Version:</strong> {serverInfo.version || 'Unknown'}
                    </div>
                    {serverInfo.license ? (
                      <>
                        <div>
                          <strong>Licensee:</strong> {serverInfo.license.licensee || 'N/A'}
                        </div>
                        <div>
                          <strong>Plan:</strong> {serverInfo.license.plan || 'N/A'}
                        </div>
                        <div>
                          <strong>Devices:</strong> {serverInfo.license.device_number || 'N/A'}
                        </div>
                        <div>
                          <strong>Expires:</strong> {serverInfo.license.expiry_date || 'N/A'}
                        </div>
                        <div>
                          <strong>Invoice:</strong> {serverInfo.license.invoice || 'N/A'}
                        </div>
                      </>
                    ) : (
                      <div>No license information available</div>
                    )}
                  </div>
                }
              >
                <span className={styles.infoItemWithTooltip}>
                  <Icon name={serverInfo.license ? 'shield' : 'shield-exclamation'} size="xs" />
                  <span>
                    {serverInfo.version || 'Unknown'} {serverInfo.license ? '(Licensed)' : '(Unlicensed)'}
                  </span>
                </span>
              </Tooltip>
            </Stack>
          ) : (
            <span className={styles.infoItem}>
              <Icon name="exclamation-triangle" size="xs" />
              <span>Failed to load server info (check datasource connection)</span>
            </span>
          )}
        </div>
        {isQueryRunnable ? (
          <Button icon="play" variant="primary" size="sm" onClick={onRunQuery}>
            Run query
          </Button>
        ) : (
          <Button icon="play" variant="secondary" size="sm" disabled>
            Run query
          </Button>
        )}
      </Stack>
    </div>
  );
}

function getStyles(theme: GrafanaTheme2) {
  return {
    header: css`
      padding: ${theme.spacing(1)};
      border-bottom: 1px solid ${theme.colors.border.weak};
    `,
    info: css`
      font-size: ${theme.typography.bodySmall.fontSize};
    `,
    infoItem: css`
      display: inline-flex;
      align-items: center;
      gap: ${theme.spacing(0.5)};
      white-space: nowrap;
      color: ${theme.colors.text.secondary};

      strong {
        color: ${theme.colors.text.primary};
      }

      svg {
        flex-shrink: 0;
      }
    `,
    infoItemWithTooltip: css`
      display: flex;
      align-items: center;
      gap: ${theme.spacing(0.5)};
      white-space: nowrap;
      color: ${theme.colors.text.secondary};
      cursor: help;
      padding: ${theme.spacing(0.5)};
      border-radius: ${theme.shape.radius.default};
      transition: all 0.2s ease;

      &:hover {
        background-color: ${theme.colors.action.hover};
        color: ${theme.colors.text.primary};
      }
    `,
  };
}
