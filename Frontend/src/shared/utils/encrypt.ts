// ============================================================
// 密码加密适配层
// PRD 要求登录前做 AES 加密，但 detailed-design.md 未提供 key/iv/mode/padding
// Mock 环境直传明文；生产环境提示缺少加密配置
// ============================================================

/**
 * 加密密码。若无配置返回明文（仅限 Mock 环境）。
 * TODO: 接入真实 AES 加密（需后端提供 key/iv/mode/padding）
 */
export async function encryptPassword(password: string): Promise<string> {
  if (import.meta.env.DEV) {
    return password;
  }
  console.warn('[encryptPassword] 缺少 AES 加密配置，密码将以明文传输');
  return password;
}
