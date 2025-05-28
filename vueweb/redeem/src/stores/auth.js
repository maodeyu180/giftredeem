import { defineStore } from 'pinia';
import { authApi } from '../api/auth';
import { ref, computed } from 'vue';

export const useAuthStore = defineStore('auth', () => {
  // 状态
  const token = ref(localStorage.getItem('token') || '');
  const user = ref(JSON.parse(localStorage.getItem('user') || 'null'));
  const providers = ref([]);
  const loading = ref(false);
  const error = ref(null);

  // 计算属性
  const isAuthenticated = computed(() => !!token.value);
  const userProfile = computed(() => user.value);

  // 动作
  // 获取OAuth提供商列表
  async function fetchProviders() {
    loading.value = true;
    error.value = null;
    try {
      const response = await authApi.getProviders();
      providers.value = response.providers;
      return providers.value;
    } catch (err) {
      error.value = err.message || '获取认证提供商失败';
      throw err;
    } finally {
      loading.value = false;
    }
  }

  // 获取登录URL
  async function getLoginUrl(provider) {
    loading.value = true;
    error.value = null;
    try {
      const response = await authApi.getLoginUrl(provider);
      return response.auth_url;
    } catch (err) {
      error.value = err.message || '获取登录链接失败';
      throw err;
    } finally {
      loading.value = false;
    }
  }

  // 处理回调
  async function handleCallback(provider, code, state) {
    loading.value = true;
    error.value = null;
    try {
      const response = await authApi.handleCallback(provider, code, state);
      setAuth(response.token, response.user);
      return response;
    } catch (err) {
      error.value = err.message || '登录验证失败';
      throw err;
    } finally {
      loading.value = false;
    }
  }

  // 获取用户信息
  async function fetchUserProfile() {
    if (!isAuthenticated.value) return null;

    loading.value = true;
    error.value = null;
    try {
      const response = await authApi.getUserProfile();
      user.value = response.user;
      localStorage.setItem('user', JSON.stringify(response.user));
      return response.user;
    } catch (err) {
      error.value = err.message || '获取用户信息失败';
      throw err;
    } finally {
      loading.value = false;
    }
  }

  // 设置认证信息
  function setAuth(newToken, newUser) {
    token.value = newToken;
    user.value = newUser;
    localStorage.setItem('token', newToken);
    localStorage.setItem('user', JSON.stringify(newUser));
  }

  // 退出登录
  function logout() {
    token.value = '';
    user.value = null;
    localStorage.removeItem('token');
    localStorage.removeItem('user');
  }

  return {
    // 状态
    token,
    user,
    providers,
    loading,
    error,
    // 计算属性
    isAuthenticated,
    userProfile,
    // 方法
    fetchProviders,
    getLoginUrl,
    handleCallback,
    fetchUserProfile,
    setAuth,
    logout
  };
}); 