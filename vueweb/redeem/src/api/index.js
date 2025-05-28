import axios from 'axios';
import { ElMessage } from 'element-plus';

// 创建 axios 实例
const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api',
  timeout: 20000,
  headers: {
    'Content-Type': 'application/json',
  }
});

// 请求拦截器 - 添加 token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器 - 处理标准化响应和错误
api.interceptors.response.use(
  (response) => {
    const res = response.data;
    
    // 检查返回的code，0表示成功
    if (res.code === 0) {
      return res.data; // 直接返回数据部分
    } else {
      // 非0表示有错误，显示错误信息
      ElMessage.error(res.msg || '未知错误');
      return Promise.reject(new Error(res.msg || '未知错误'));
    }
  },
  (error) => {
    // 处理网络错误等
    const message = error.response?.data?.msg || error.message || '网络错误';
    ElMessage.error(message);
    
    // 处理401错误 - 未授权
    if (error.response && error.response.status === 401) {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      window.location.href = '/login';
    }
    
    return Promise.reject(error);
  }
);

export default api; 