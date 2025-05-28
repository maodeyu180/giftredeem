<template>
  <div class="dashboard-container">
    <el-container>
      <!-- 侧边栏 -->
      <el-aside width="auto">
        <div class="sidebar">
          <div class="logo-container">
            <img src="../../assets/logo.png" alt="Logo" class="logo" />
            <h1 v-if="!isCollapse">GiftRedeem</h1>
          </div>
          
          <el-menu
            :default-active="activeMenu"
            class="menu"
            :collapse="isCollapse"
            background-color="#001529"
            text-color="#fff"
            active-text-color="#409eff"
            router
          >
            <el-menu-item index="/dashboard/benefits">
              <el-icon><Gift /></el-icon>
              <template #title>我的福利</template>
            </el-menu-item>
            
            <el-menu-item index="/dashboard/benefits/create">
              <el-icon><Plus /></el-icon>
              <template #title>发布福利</template>
            </el-menu-item>
            
            <el-menu-item index="/dashboard/claims">
              <el-icon><Collection /></el-icon>
              <template #title>我的领取</template>
            </el-menu-item>
          </el-menu>
          
          <div class="collapse-button" @click="toggleCollapse">
            <el-icon v-if="isCollapse"><DArrowRight /></el-icon>
            <el-icon v-else><DArrowLeft /></el-icon>
          </div>
        </div>
      </el-aside>
      
      <el-container>
        <!-- 顶部导航 -->
        <el-header>
          <div class="header-container">
            <div class="left">
              <el-breadcrumb separator="/">
                <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
                <el-breadcrumb-item v-for="(item, index) in breadcrumbs" :key="index">
                  {{ item.title }}
                </el-breadcrumb-item>
              </el-breadcrumb>
            </div>
            
            <div class="right">
              <el-dropdown trigger="click" @command="handleCommand">
                <div class="user-info">
                  <el-avatar :size="32" :src="userAvatar" />
                  <span class="username">{{ userName }}</span>
                  <el-icon><CaretBottom /></el-icon>
                </div>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="profile">个人资料</el-dropdown-item>
                    <el-dropdown-item command="logout">退出登录</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </div>
        </el-header>
        
        <!-- 主内容区域 -->
        <el-main>
          <router-view v-slot="{ Component }">
            <transition name="fade" mode="out-in">
              <component :is="Component" />
            </transition>
          </router-view>
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useAuthStore } from '../../stores/auth';
import { useBenefitStore } from '../../stores/benefit';
import { 
  Present as Gift, 
  Plus, 
  Collection, 
  ArrowLeft as DArrowLeft, 
  ArrowRight as DArrowRight, 
  CaretBottom 
} from '@element-plus/icons-vue';
import { ElMessageBox, ElMessage } from 'element-plus';

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();

// 侧边栏折叠状态
const isCollapse = ref(window.innerWidth < 768);

// 监听窗口大小变化
window.addEventListener('resize', () => {
  isCollapse.value = window.innerWidth < 768;
});

// 切换侧边栏折叠状态
const toggleCollapse = () => {
  isCollapse.value = !isCollapse.value;
};

// 当前活动菜单
const activeMenu = computed(() => {
  return route.path;
});

// 用户信息
const userName = computed(() => {
  return authStore.userProfile?.username || '用户';
});

const userAvatar = computed(() => {
  return authStore.userProfile?.avatar_url || '';
});

// 面包屑导航
const breadcrumbs = computed(() => {
  const currentRoute = route.matched.filter(item => item.meta && item.meta.title);
  return currentRoute.map(item => ({
    path: item.path,
    title: item.meta.title
  }));
});

// 处理下拉菜单命令
const handleCommand = (command) => {
  if (command === 'logout') {
    ElMessageBox.confirm('确定要退出登录吗?', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }).then(() => {
      authStore.logout();
      router.push('/login');
    }).catch(() => {});
  } else if (command === 'profile') {
    // TODO: 跳转到个人资料页面
  }
};

// 获取用户信息和初始化数据
onMounted(async () => {
  try {
    if (authStore.isAuthenticated) {
      // 先获取用户信息
      if (!authStore.userProfile) {
        await authStore.fetchUserProfile();
      }
      
      // 控制台首次加载，确保数据初始化
      if (route.path.includes('/dashboard/benefits')) {
        console.log('正在初始化我的福利数据...');
        // 初始化福利数据，防止控制台空白
        const benefitStore = useBenefitStore();
        await benefitStore.fetchMyBenefits();
      }
    }
  } catch (err) {
    console.error('初始化控制台数据失败', err);
  }
});
</script>

<style scoped>
.dashboard-container {
  height: 100%;
  width: 100%;
  overflow: hidden;
}

.el-container {
  height: 100%;
  width: 100%;
}

.sidebar {
  height: 100%;
  background-color: #001529;
  position: relative;
  overflow: hidden;
  transition: width 0.3s;
}

.logo-container {
  height: 60px;
  display: flex;
  align-items: center;
  padding: 0 20px;
  color: white;
  background-color: #002140;
  overflow: hidden;
}

.logo {
  width: 30px;
  height: 30px;
  margin-right: 10px;
}

.logo-container h1 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  white-space: nowrap;
}

.menu {
  border-right: none;
  height: calc(100% - 120px);
}

.collapse-button {
  position: absolute;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background-color: #002140;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: white;
  transition: all 0.3s;
}

.collapse-button:hover {
  background-color: #1890ff;
}

.el-header {
  background-color: white;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
  padding: 0 20px;
}

.header-container {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.user-info {
  display: flex;
  align-items: center;
  cursor: pointer;
}

.username {
  margin: 0 8px;
  font-size: 14px;
}

.el-main {
  padding: 20px;
  overflow-y: auto;
  background-color: #f0f2f5;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .el-aside {
    position: fixed;
    z-index: 1000;
    height: 100%;
  }
  
  .el-main {
    margin-left: 64px;
  }
  
  .username {
    display: none;
  }
}

/* 页面过渡动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style> 