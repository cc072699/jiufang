// ============================================================
// 空值展示规则 - 详见 detailed-design.md 与前端开发方案
// ============================================================

import type { ColumnDef } from '../api/types';

/**
 * 格式化空值展示：
 * - number 类型空值 → 显示 "0"
 * - string/date/enum 空值 → 置空
 * - 金额/数量字段空值 → tooltip 提示（由组件层处理）
 */
export function formatNullValue(
  value: unknown,
  columnType: ColumnDef['type'],
): string | number {
  if (value === null || value === undefined) {
    if (columnType === 'number') return 0;
    return '';
  }
  return value as string | number;
}

/** 判断是否为金额/数量类字段（按列名启发式匹配） */
export function isMonetaryField(columnName: string): boolean {
  const monetaryKeywords = ['金额', '价格', '成本', '利润', '数量', 'amount', 'price', 'cost', 'profit', 'qty', 'quantity'];
  const lower = columnName.toLowerCase();
  return monetaryKeywords.some((kw) => lower.includes(kw));
}

/** 判断值是否为空（null/undefined） */
export function isNullValue(value: unknown): boolean {
  return value === null || value === undefined;
}
