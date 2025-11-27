export const EXAMPLES = [
  {
    name: 'Simple label comparison',
    insertText: '{\n  "&sensor_id": { "$eq": "sensor_001" }\n}',
    description: 'Basic label equality check',
  },
  {
    name: 'Numeric range filter',
    insertText: '{\n  "&temperature": { "$gte": 20 },\n  "&temperature": { "$lte": 30 }\n}',
    description: 'Filter values within range',
  },
  {
    name: 'String contains',
    insertText: '{\n  "&message": { "$contains": "error" }\n}',
    description: 'Filter strings containing text',
  },
  {
    name: 'Multiple conditions (AND)',
    insertText: '{\n  "$and": [\n    { "&status": { "$eq": "active" } },\n    { "&count": { "$gt": 10 } }\n  ]\n}',
    description: 'Multiple conditions with AND logic',
  },
  {
    name: 'Any condition (OR)',
    insertText: '{\n  "$or": [\n    { "&priority": { "$eq": "high" } },\n    { "&urgent": { "$eq": true } }\n  ]\n}',
    description: 'Multiple conditions with OR logic',
  },
  {
    name: 'Label exists check',
    insertText: '{\n  "$has": ["sensor_id", "timestamp"]\n}',
    description: 'Check if record has specific labels',
  },
  {
    name: 'String pattern matching',
    insertText: '{\n  "&filename": { "$starts_with": "log_" }\n}',
    description: 'String starts with pattern',
  },
  {
    name: 'Value in list',
    insertText: '{\n  "$in": ["&status", "active", "pending", "running"]\n}',
    description: 'Check if value is in list',
  },
  {
    name: 'Complex nested condition',
    insertText:
      '{\n  "$and": [\n    {\n      "$or": [\n        { "&level": { "$eq": "error" } },\n        { "&level": { "$eq": "critical" } }\n      ]\n    },\n    { "&timestamp": { "$gte": 1640995200000 } }\n  ]\n}',
    description: 'Nested logical operators example',
  },
];
