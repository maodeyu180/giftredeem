<template>
  <div class="home-container">
    <div class="hero-section">
      <div class="logo-container">
        <img src="../assets/logo.png" alt="GiftRedeem Logo" class="logo" />
        <h1>Gift Redeem</h1>
      </div>
      <h2>私密福利分发平台</h2>
      <p class="description">基于多渠道 OAuth 认证，福利仅通过私密链接分享，确保分发的精准性和私密性</p>
      
      <div class="action-buttons">
        <template v-if="!isAuthenticated">
          <el-button type="primary" size="large" @click="navigateToLogin">立即登录</el-button>
        </template>
        <template v-else>
          <el-button type="primary" size="large" @click="navigateToDashboard">控制台</el-button>
        </template>
        <el-button size="large" @click="navigateToClaim" v-if="hasClaimCode">兑换福利</el-button>
      </div>
    </div>
    
    <div class="features-section">
      <h3>核心功能</h3>
      
      <div class="features-grid">
        <div class="feature-card">
          <el-icon><Lock /></el-icon>
          <h4>私密链接</h4>
          <p>福利通过唯一私密链接分享，确保只有获得链接的用户才能访问和领取</p>
        </div>
        
        <div class="feature-card">
          <el-icon><SetUp /></el-icon>
          <h4>多渠道认证</h4>
          <p>支持多渠道 OAuth 登录，用户可以使用已有账号快速登录</p>
        </div>
        
        <div class="feature-card">
          <el-icon><Promotion /></el-icon>
          <h4>福利发布</h4>
          <p>用户可发布各类兑换码福利，支持智能解析和条件设置</p>
        </div>
        
        <div class="feature-card">
          <el-icon><User /></el-icon>
          <h4>账户合并</h4>
          <p>支持多个 OAuth 账户绑定到同一用户，统一管理福利</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { Lock, Setting as SetUp, Promotion, User } from '@element-plus/icons-vue';
import { useAuthStore } from '../stores/auth';

const router = useRouter();
const route = useRoute();
const authStore = useAuthStore();
const hasClaimCode = ref(false);

// 检查用户是否已登录
const isAuthenticated = computed(() => authStore.isAuthenticated);

// 检查 URL 中是否有福利代码
onMounted(() => {
  // 如果URL中有claim参数，设置为有兑换码
  hasClaimCode.value = !!route.query.claim;
});

// 导航到登录页
const navigateToLogin = () => {
  router.push('/login');
};

// 导航到兑换页面
const navigateToClaim = () => {
  if (route.query.claim) {
    router.push(`/claim/${route.query.claim}`);
  }
};

// 导航到控制台
const navigateToDashboard = () => {
  router.push('/dashboard/benefits');
};
</script>

<style scoped>
.home-container {
  width: 100%;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.hero-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem;
  background: linear-gradient(135deg, #6e8efb, #a777e3);
  color: white;
  text-align: center;
  min-height: 60vh;
}

.logo-container {
  display: flex;
  align-items: center;
  margin-bottom: 1rem;
}

.logo {
  width: 50px;
  height: 50px;
  margin-right: 1rem;
}

h1 {
  font-size: 2.5rem;
  margin: 0;
}

h2 {
  font-size: 2rem;
  margin: 1rem 0;
}

.description {
  max-width: 600px;
  margin: 0 auto 2rem;
  font-size: 1.2rem;
  line-height: 1.6;
}

.action-buttons {
  display: flex;
  gap: 1rem;
  margin-top: 1rem;
}

.features-section {
  padding: 4rem 2rem;
  background-color: #f9f9f9;
}

h3 {
  text-align: center;
  font-size: 1.8rem;
  margin-bottom: 3rem;
  color: #333;
}

.features-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

.feature-card {
  background: white;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  padding: 2rem;
  text-align: center;
  transition: transform 0.3s ease;
}

.feature-card:hover {
  transform: translateY(-5px);
}

.feature-card .el-icon {
  font-size: 2.5rem;
  color: #6e8efb;
  margin-bottom: 1rem;
}

h4 {
  font-size: 1.4rem;
  margin: 1rem 0;
  color: #333;
}

.feature-card p {
  color: #666;
  line-height: 1.5;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .hero-section {
    padding: 3rem 1rem;
  }
  
  h1 {
    font-size: 2rem;
  }
  
  h2 {
    font-size: 1.5rem;
  }
  
  .description {
    font-size: 1rem;
  }
  
  .action-buttons {
    flex-direction: column;
  }
  
  .features-grid {
    grid-template-columns: 1fr;
  }
}
</style> 