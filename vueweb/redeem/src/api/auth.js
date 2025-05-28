import api from './index';

// 认证相关的API
export const authApi = {
  // 获取可用的OAuth提供商
  getProviders: () => api.get('/auth/providers'),
  
  // 获取OAuth登录URL
  getLoginUrl: (provider) => api.get(`/auth/login/${provider}`),
  
  // 处理OAuth回调 - 两种方式
  // 1. 后端处理 - 当直接访问API时使用
  handleCallback: (provider, code, state) => api.get(`/auth/callback/${provider}`, { 
    params: { code, state, response_type: 'json' } 
  }),
  
  // 2. 验证代码 - 当前端直接接收回调时使用
  verifyCode: (provider, code) => api.post(`/auth/verify/${provider}`, { code }),
  
  // 获取当前用户信息
  getUserProfile: () => api.get('/auth/profile'),
}; 