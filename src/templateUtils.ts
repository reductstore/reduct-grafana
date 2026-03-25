import { ScopedVars } from '@grafana/data';
import { getTemplateSrv } from '@grafana/runtime';

import { ReductWhenCondition } from './types';

export function replaceWhenTemplateVariables(
  when: ReductWhenCondition | undefined | null,
  scopedVars?: ScopedVars
): ReductWhenCondition | undefined {
  if (when === undefined || when === null) {
    return undefined;
  }

  const templateSrv = getTemplateSrv();

  const replaceValue = (value: unknown): unknown => {
    if (value === null || value === undefined) {
      return value;
    }

    if (typeof value === 'string') {
      return templateSrv.replace(value, scopedVars);
    }

    if (Array.isArray(value)) {
      return value.map((item) => replaceValue(item));
    }

    if (typeof value === 'object') {
      return Object.entries(value).reduce((acc, [key, val]) => {
        acc[key] = replaceValue(val);
        return acc;
      }, {} as Record<string, unknown>);
    }

    return value;
  };

  if (typeof when === 'string') {
    const sanitized = templateSrv.replace(when, scopedVars);
    try {
      const parsed = JSON.parse(sanitized);
      return replaceValue(parsed) as ReductWhenCondition;
    } catch {
      return sanitized as unknown as ReductWhenCondition;
    }
  }

  return replaceValue(when) as ReductWhenCondition;
}
