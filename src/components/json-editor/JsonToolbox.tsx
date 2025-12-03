import React, { useEffect } from 'react';
import { IconButton, Spinner, Stack, Tooltip, useStyles2 } from '@grafana/ui';
import { css } from '@emotion/css';
import { GrafanaTheme2 } from '@grafana/data';
import { getBackendSrv } from '@grafana/runtime';
import { ReductQuery } from '../../types';
import { useDebounce } from 'react-use';

enum ValidationStatus {
  Loading = 'loading',
  Valid = 'valid',
  Invalid = 'invalid',
}

interface JsonToolboxProps {
  query: ReductQuery;
  formatCode: () => void;
  onExpand: (expanded: boolean) => void;
  isExpanded: boolean;
  datasourceId: number;
}

export function JsonToolbox({ query, formatCode, onExpand, isExpanded, datasourceId }: JsonToolboxProps) {
  const styles = useStyles2(getStyles);

  const [when, setWhen] = React.useState(query.options?.when);
  const [status, setStatus] = React.useState<ValidationStatus>(ValidationStatus.Valid);
  const [error, setError] = React.useState<string | undefined>(undefined);
  const [missingBucket, setMissingBucket] = React.useState<boolean>(!query.bucket);
  const [missingEntry, setMissingEntry] = React.useState<boolean>(!query.entry);

  const toggleExpand = () => onExpand(!isExpanded);

  // Debounce validation and loading state
  useEffect(() => {
    setWhen(query.options?.when);
  }, [query.options?.when]);

  useDebounce(
    () => {
      const isBucketMissing = !query.bucket;
      const isEntryMissing = !query.entry;

      setMissingBucket(isBucketMissing);
      setMissingEntry(isEntryMissing);

      if (isBucketMissing || isEntryMissing) {
        return;
      }

      if (when === undefined || when === null) {
        setStatus(ValidationStatus.Valid);
        setError(undefined);
        return;
      }

      setStatus(ValidationStatus.Loading);
      setError(undefined);

      // If JSON is valid, proceed with backend validation
      getBackendSrv()
        .post(
          `/api/datasources/${datasourceId}/resources/validateCondition`,
          {
            bucket: query.bucket,
            entry: query.entry,
            condition: when,
          },
          {
            showErrorAlert: false,
          }
        )
        .then((response: { valid: boolean; error?: string }) => {
          if (response.valid) {
            setStatus(ValidationStatus.Valid);
            setError(undefined);
          } else {
            setStatus(ValidationStatus.Invalid);
            setError(response.error || 'Invalid condition');
          }
        })
        .catch((err: any) => {
          setStatus(ValidationStatus.Invalid);
          setError(err.data?.message || err.message || 'Validation failed');
        });
    },
    500,
    [when, query.bucket, query.entry, datasourceId]
  );

  return (
    <div className={styles.toolbar}>
      <div className={styles.validation}>
        {missingBucket && missingEntry && <span>Select bucket and entry to validate condition</span>}
        {missingBucket && !missingEntry && <span>Select bucket to validate condition</span>}
        {!missingBucket && missingEntry && <span>Select entry to validate condition</span>}
        {!missingBucket && !missingEntry && status === ValidationStatus.Loading && (
          <>
            <Spinner size="xs" />
            <span>Validating...</span>
          </>
        )}

        {!missingBucket && !missingEntry && status === ValidationStatus.Valid && (
          <>
            <span className={styles.checkmark}>✓</span>
            <span>Valid condition</span>
          </>
        )}

        {!missingBucket && !missingEntry && status === ValidationStatus.Invalid && (
          <>
            <span className={styles.error}>✗</span>
            <span>{error || 'Invalid condition'}</span>
          </>
        )}
      </div>

      <Stack gap={1}>
        <Tooltip content="Format JSON">
          <IconButton name="brackets-curly" size="xs" onClick={formatCode} aria-label="Format query" />
        </Tooltip>

        <Tooltip content={isExpanded ? 'Collapse editor' : 'Expand editor'}>
          <IconButton
            size="xs"
            name={isExpanded ? 'compress-arrows' : 'expand-arrows'}
            onClick={toggleExpand}
            aria-label={isExpanded ? 'Collapse editor' : 'Expand editor'}
          />
        </Tooltip>
      </Stack>
    </div>
  );
}

function getStyles(theme: GrafanaTheme2) {
  return {
    toolbar: css`
      padding: ${theme.spacing(1)};
      display: flex;
      align-items: center;
      justify-content: space-between;
      border-bottom: 1px solid ${theme.colors.border.weak};
      background: ${theme.colors.background.secondary};
      font-size: ${theme.typography.bodySmall.fontSize};
    `,
    validation: css`
      display: flex;
      align-items: center;
      gap: ${theme.spacing(0.5)};
    `,
    checkmark: css`
      font-weight: bold;
      color: ${theme.colors.success.text};
    `,
    error: css`
      font-weight: bold;
      color: ${theme.colors.error.text};
    `,
  };
}
