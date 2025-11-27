// Comparison operators
export const COMPARISON_OPERATORS = [
  { name: '$eq', description: 'Equal to', insertText: '$eq' },
  { name: '$ne', description: 'Not equal to', insertText: '$ne' },
  { name: '$gt', description: 'Greater than', insertText: '$gt' },
  { name: '$gte', description: 'Greater than or equal', insertText: '$gte' },
  { name: '$lt', description: 'Less than', insertText: '$lt' },
  { name: '$lte', description: 'Less than or equal', insertText: '$lte' },
];

// Logical operators
export const LOGICAL_OPERATORS = [
  { name: '$and', description: 'Logical AND', insertText: '$and' },
  { name: '$all_of', description: 'All conditions must be true', insertText: '$all_of' },
  { name: '$or', description: 'Logical OR', insertText: '$or' },
  { name: '$any_of', description: 'Any condition must be true', insertText: '$any_of' },
  { name: '$xor', description: 'Exactly one condition must be true', insertText: '$xor' },
  { name: '$one_of', description: 'Exactly one condition must be true', insertText: '$one_of' },
  { name: '$not', description: 'Logical NOT', insertText: '$not' },
  { name: '$none_of', description: 'No conditions must be true', insertText: '$none_of' },
  { name: '$in', description: 'Value is in list', insertText: '$in' },
  { name: '$nin', description: 'Value is not in list', insertText: '$nin' },
];

// String operators
export const STRING_OPERATORS = [
  { name: '$contains', description: 'String contains substring', insertText: '$contains' },
  { name: '$starts_with', description: 'String starts with substring', insertText: '$starts_with' },
  { name: '$ends_with', description: 'String ends with substring', insertText: '$ends_with' },
];

// Arithmetic operators
export const ARITHMETIC_OPERATORS = [
  { name: '$add', description: 'Addition', insertText: '$add' },
  { name: '$sub', description: 'Subtraction', insertText: '$sub' },
  { name: '$mult', description: 'Multiplication', insertText: '$mult' },
  { name: '$div', description: 'Division (float)', insertText: '$div' },
  { name: '$div_num', description: 'Division (integer)', insertText: '$div_num' },
  { name: '$rem', description: 'Remainder', insertText: '$rem' },
  { name: '$abs', description: 'Absolute value', insertText: '$abs' },
];

// Miscellaneous operators
export const MISC_OPERATORS = [
  { name: '$has', description: 'Record has labels', insertText: '$has' },
  { name: '$exists', description: 'Record has labels', insertText: '$exists' },
  { name: '$cast', description: 'Cast to type', insertText: '$cast' },
  { name: '$ref', description: 'Reference label', insertText: '$ref' },
  { name: '$timestamp', description: 'Record timestamp', insertText: '$timestamp' },
  { name: '$id', description: 'Record timestamp', insertText: '$id' },
];
