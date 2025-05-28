<template>
  <div class="callback-container">
    <div class="loading-state">
      <el-card class="loading-card">
        <template #header>
          <div class="card-header">
            <h2>处理登录中...</h2>
          </div>
        </template>
        
        <el-result v-if="error" icon="error" :title="error">
          <template #extra>
            <el-button type="primary" @click="goToLogin">重新登录</el-button>
          </template>
        </el-result>
        
        <div v-else class="loading-content">
          <el-progress type="circle" :percentage="loadingProgress" :status="loadingStatus" />
          <p class="loading-text">{{ loadingMessage }}</p>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useAuthStore } from '../../stores/auth';
import { ElMessage } from 'element-plus';

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();

const error = ref(null);
const loadingProgress = ref(0);
const loadingStep = ref(0);
const loadingMessage = ref('验证您的登录信息...');

// 计算属性
const loadingStatus = computed(() => {
  if (error.value) return 'exception';
  if (loadingProgress.value === 100) return 'success';
  return '';
});

// 执行OAuth回调处理
onMounted(async () => {
  const { provider } = route.params;
  const { code, state } = route.query;
  
  if (!code || !state) {
    error.value = '无效的回调参数';
    return;
  }
  
  // 更新加载状态
  loadingStep.value = 1;
  loadingProgress.value = 25;
  loadingMessage.value = '验证OAuth回调...';
  
  try {
    // 直接从URL参数中获取token和用户信息
    const token = route.query.token;
    const user_id = route.query.user_id;
    const username = route.query.username;
    const avatar_url = route.query.avatar_url;
    
    // 如果URL中已经有token（方式二：直接回调到前端）
    if (token) {
      // 更新加载状态
      loadingStep.value = 2;
      loadingProgress.value = 50;
      loadingMessage.value = '处理登录信息...';
      
      // 直接使用URL中的用户数据
      const userData = {
        id: Number(user_id),
        username: username,
        avatar_url: avatar_url || ''
      };
      
      // 设置认证信息
      authStore.setAuth(token, userData);
      
      // 更新加载状态
      loadingStep.value = 3;
      loadingProgress.value = 75;
      loadingMessage.value = '登录成功，正在跳转...';
    } else {
      // 方式一：通过后端API处理回调
      // 处理 OAuth 回调
      loadingMessage.value = '与服务器通信中...';
      await authStore.handleCallback(provider, code, state);
      
      // 更新加载状态
      loadingStep.value = 2;
      loadingProgress.value = 50;
      loadingMessage.value = '获取用户信息...';
      
      // 获取用户信息
      await authStore.fetchUserProfile();
      
      // 更新加载状态
      loadingStep.value = 3;
      loadingProgress.value = 75;
      loadingMessage.value = '登录成功，正在跳转...';
    }
    
    // 延迟一下再跳转，让用户看到登录成功的状态
    setTimeout(() => {
      loadingStep.value = 4;
      loadingProgress.value = 100;
      
      // 跳转到之前存储的重定向URL或默认到控制台
      let redirectUrl = sessionStorage.getItem('redirect_after_login') || '/dashboard/benefits';
      sessionStorage.removeItem('redirect_after_login');
      
      // 强制确保重定向URL不为空
      if (!redirectUrl || redirectUrl === 'undefined' || redirectUrl === 'null') {
        redirectUrl = '/dashboard/benefits';
        console.log('重定向URL无效，使用默认值:', redirectUrl);
      }
      
      ElMessage.success('登录成功');
      console.log('准备跳转到:', redirectUrl);
      
      try {
        // 使用 router.push 代替 router.replace，并添加错误处理
        router.push(redirectUrl)
          .then(() => {
            console.log('路由跳转成功');
          })
          .catch(err => {
            console.error('路由跳转失败', err);
            // 如果路由跳转失败，尝试使用 window.location
            window.location.href = redirectUrl.startsWith('/') 
              ? window.location.origin + redirectUrl
              : redirectUrl;
          });
      } catch (err) {
        console.error('执行路由跳转时出错', err);
        // 备用方案：使用 window.location
        window.location.href = redirectUrl.startsWith('/') 
          ? window.location.origin + redirectUrl
          : redirectUrl;
      }
    }, 1000);
    
  } catch (err) {
    console.error('OAuth回调处理失败', err);
    error.value = '登录验证失败，请重试';
    loadingProgress.value = 100;
    loadingStatus.value = 'exception';
  }
});

// 返回登录页
const goToLogin = () => {
  router.push('/login');
};
</script>

<style scoped>
.callback-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  background: linear-gradient(135deg, #f5f7fa, #c3cfe2);
}

.loading-card {
  width: 100%;
  max-width: 450px;
}

.card-header {
  text-align: center;
}

.card-header h2 {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 500;
}

.loading-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 2rem 1rem;
}

.loading-text {
  margin-top: 1.5rem;
  color: #606266;
  text-align: center;
}
</style> 