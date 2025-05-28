<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <img src="../../assets/logo.png" alt="GiftRedeem Logo" class="logo" />
        <h1>欢迎使用 GiftRedeem</h1>
        <p>请选择一种登录方式继续</p>
      </div>
      
      <div class="login-body">
        <el-skeleton :rows="3" animated v-if="loading" />
        
        <template v-else>
          <div v-if="error" class="error-message">
            <el-alert :title="error" type="error" show-icon />
          </div>
          
          <div class="providers-list">
            <el-button 
              v-for="provider in providers" 
              :key="provider.name"
              class="provider-button"
              @click="login(provider.name)"
              :loading="loadingProvider === provider.name"
              size="large"
            >
              <img 
                :src="getProviderIcon(provider.name)" 
                :alt="provider.display_name" 
                class="provider-icon" 
              />
              通过 {{ provider.display_name }} 登录
            </el-button>
          </div>
        </template>
      </div>
      
      <div class="login-footer">
        <p>没有账号？通过第三方登录时会自动创建</p>
        <el-button text @click="goHome">返回首页</el-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useAuthStore } from '../../stores/auth';
import { ElMessage } from 'element-plus';

const router = useRouter();
const route = useRoute();
const authStore = useAuthStore();

const providers = ref([]);
const loading = ref(true);
const error = ref(null);
const loadingProvider = ref(null);

// 获取OAuth提供商
onMounted(async () => {
  try {
    providers.value = await authStore.fetchProviders();
  } catch (err) {
    error.value = '获取登录提供商失败，请刷新重试';
    console.error(err);
  } finally {
    loading.value = false;
  }
});

// 获取提供商图标
const getProviderIcon = (providerName) => {
  const icons = {
    'linuxdo': '/src/assets/providers/linuxdo.png',
    'github': '/src/assets/providers/github.png',
    'google': '/src/assets/providers/google.png',
    'wechat': '/src/assets/providers/wechat.png',
  };
  
  return icons[providerName] || '/src/assets/providers/default.png';
};

// 处理登录
const login = async (providerName) => {
  loadingProvider.value = providerName;
  try {
    const authUrl = await authStore.getLoginUrl(providerName);
    // 将重定向URL添加到会话存储中，以便在回调时恢复
    if (route.query.redirect) {
      sessionStorage.setItem('redirect_after_login', route.query.redirect);
    }
    // 重定向到OAuth授权URL
    window.location.href = authUrl;
  } catch (err) {
    ElMessage.error('获取登录链接失败，请重试');
    console.error(err);
    loadingProvider.value = null;
  }
};

// 返回首页
const goHome = () => {
  router.push('/');
};
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #f5f7fa, #c3cfe2);
  padding: 1rem;
}

.login-card {
  width: 100%;
  max-width: 450px;
  background: white;
  border-radius: 10px;
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.login-header {
  padding: 2rem;
  text-align: center;
  background: #f9f9f9;
  border-bottom: 1px solid #eee;
}

.logo {
  width: 60px;
  height: 60px;
  margin-bottom: 1rem;
}

h1 {
  font-size: 1.8rem;
  margin: 0.5rem 0;
  color: #333;
}

.login-header p {
  margin: 0.5rem 0 0;
  color: #666;
}

.login-body {
  padding: 2rem;
}

.error-message {
  margin-bottom: 1.5rem;
}

.providers-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.provider-button {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  padding: 12px;
  border-radius: 6px;
  transition: all 0.3s ease;
}

.provider-icon {
  width: 24px;
  height: 24px;
  margin-right: 8px;
}

.login-footer {
  padding: 1.5rem 2rem;
  text-align: center;
  border-top: 1px solid #eee;
}

.login-footer p {
  margin: 0 0 0.5rem;
  color: #666;
  font-size: 0.9rem;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .login-card {
    max-width: 100%;
    border-radius: 0;
  }
  
  .login-header {
    padding: 1.5rem;
  }
  
  .login-body {
    padding: 1.5rem;
  }
  
  h1 {
    font-size: 1.5rem;
  }
}
</style> 