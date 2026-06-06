export async function encryptPassword(password: string): Promise<string> {
  if (import.meta.env.DEV) {
    return password;
  }
  console.warn('[encryptPassword] 缺少 AES 加密配置，密码将以明文传输');
  return password;
}
