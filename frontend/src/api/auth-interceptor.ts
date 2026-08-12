import axios from 'axios';
import type { InternalAxiosRequestConfig } from 'axios';

// ============================================
// 请求拦截器:自动附加 JWT(登录后 token 存于 localStorage)
// 所有 API 调用自动携带 Authorization: Bearer <token>
//
// 注意:本文件独立于 orval 生成的 client.ts,避免重新生成时被覆盖。
// 由 main.tsx 顶部 import 触发,全局生效。
// ============================================
axios.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export {};
