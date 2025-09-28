import JSON5 from 'json5';

/**
 * Safely parse a JSON string into an object.
 * Throws if the string is invalid JSON.
 */
export function parseJson<T = any>(text: string): T {
  return JSON5.parse(text) as T;
}

/**
 * Pretty-print an object as a JSON string.
 * @param value The object/value to stringify
 * @param space Indentation spaces (default: 2)
 */
export function stringifyJson(value: any, space = 2): string {
  return JSON5.stringify(value, null, space);
}
