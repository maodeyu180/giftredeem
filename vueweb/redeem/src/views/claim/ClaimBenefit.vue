<template>
  <div class="claim-benefit">
    <div class="container">
      <!-- 加载状态 -->
      <div v-if="loading" class="loading-state">
        <el-skeleton style="width: 100%" animated :count="3" />
      </div>
      
      <!-- 错误状态 -->
      <div v-else-if="error" class="error-state">
        <el-result
          icon="error"
          :title="error"
          sub-title="无法获取福利信息，可能已过期或已被删除"
        >
          <template #extra>
            <el-button type="primary" @click="goHome">返回首页</el-button>
          </template>
        </el-result>
      </div>
      
      <!-- 兑换成功状态 -->
      <div v-else-if="claimSuccess" class="success-state">
        <el-result
          icon="success"
          title="福利领取成功！"
          sub-title="请查看以下兑换码，并妥善保存"
        >
          <div class="code-container">
            <div class="code-card">
              <h3>您的兑换码</h3>
              <div class="code-value">{{ claimedCode }}</div>
              <el-button size="small" type="primary" @click="copyCode">
                <el-icon><CopyDocument /></el-icon> 复制兑换码
              </el-button>
            </div>
          </div>
          
          <template #extra>
            <div class="action-buttons">
              <el-button @click="goHome">返回首页</el-button>
              <el-button type="primary" @click="viewMyClaims" v-if="isAuthenticated">查看我的领取</el-button>
            </div>
          </template>
        </el-result>
      </div>
      
      <!-- 福利详情 -->
      <template v-else-if="benefit">
        <el-card class="benefit-card">
          <div class="benefit-header">
            <h1>{{ benefit.title }}</h1>
            <div class="tag-container">
              <el-tag type="success" v-if="!isExpired && isBenefitActive && !isFullyClaimed">可领取</el-tag>
              <el-tag type="info" v-else-if="isFullyClaimed">已领完</el-tag>
              <el-tag type="danger" v-else>已失效</el-tag>
            </div>
          </div>
          
          <div class="benefit-info">
            <h3>福利详情</h3>
            <p class="description">{{ benefit.description }}</p>
            
            <div class="info-items">
              <div class="info-item">
                <el-icon><Calendar /></el-icon>
                <span>有效期至: {{ formatDate(benefit.expires_at) }}</span>
              </div>
              
              <div class="info-item">
                <el-icon><User /></el-icon>
                <span>
                  已领取: {{ benefit.claimed_count }} 
                  {{ benefit.total_count ? '/ ' + benefit.total_count : '' }}
                </span>
              </div>
              
              <div class="info-item">
                <el-icon><InfoFilled /></el-icon>
                <span>每人限领: {{ benefit.claim_limit || 1 }} 次</span>
              </div>
            </div>
          </div>
          
          <div class="claim-action">
            <!-- 未登录状态 -->
            <template v-if="!isAuthenticated">
              <p class="notice">请先登录后再领取福利</p>
              <el-button type="primary" @click="goToLogin">立即登录</el-button>
            </template>
            
            <!-- 已领取达到上限 -->
            <template v-else-if="userClaimCount >= benefit.claim_limit">
              <el-alert
                title="您已达到领取上限"
                type="warning"
                :closable="false"
                show-icon
              >
                <p>您已领取过该福利，无法再次领取</p>
              </el-alert>
              <el-button @click="viewMyClaims">查看我的领取</el-button>
            </template>
            
            <!-- 福利总量已领完 -->
            <template v-else-if="isFullyClaimed">
              <el-alert
                title="福利已被领完"
                type="info"
                :closable="false"
                show-icon
              >
                <p>该福利已被领完，请下次再来</p>
              </el-alert>
            </template>
            
            <!-- 福利已过期或停用 -->
            <template v-else-if="isExpired || !isBenefitActive">
              <el-alert
                title="福利已失效"
                type="error"
                :closable="false"
                show-icon
              >
                <p>该福利已过期或被停用，无法领取</p>
              </el-alert>
            </template>
            
            <!-- 可以领取 -->
            <template v-else>
              <p class="notice">确认领取该福利？领取后将获得兑换码</p>
              <el-button 
                type="primary" 
                @click="claimBenefit" 
                :loading="claimLoading"
                :disabled="claimLoading"
              >立即领取</el-button>
            </template>
          </div>
        </el-card>
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useAuthStore } from '../../stores/auth';
import { useBenefitStore } from '../../stores/benefit';
import { benefitApi } from '../../api/benefit';
import { ElMessage } from 'element-plus';
import { Calendar, User, InfoFilled, CopyDocument } from '@element-plus/icons-vue';

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();
const benefitStore = useBenefitStore();

const loading = ref(true);
const error = ref(null);
const benefit = ref(null);
const claimLoading = ref(false);
const claimSuccess = ref(false);
const claimedCode = ref('');
const userClaimCount = ref(0);
const claimStatus = ref('');

// 计算属性
const isAuthenticated = computed(() => authStore.isAuthenticated);
const isExpired = computed(() => {
  if (!benefit.value) return true;
  
  if (!benefit.value.expires_at) return false;
  
  try {
    const date = new Date(benefit.value.expires_at);
    if (isNaN(date.getTime())) {
      return false;
    }
    return date < new Date();
  } catch (err) {
    console.error('Error checking expiration:', err);
    return false;
  }
});

const isBenefitActive = computed(() => {
  // 优先使用API返回的claim_status
  if (claimStatus.value === 'available') {
    return true;
  }
  
  // 如果没有claim_status，则使用status字段
  if (benefit.value && benefit.value.status) {
    return benefit.value.status === 'active';
  }
  
  // 没有任何状态字段，默认为可用（仅在未过期时）
  return !isExpired.value;
});

// 判断福利是否已被领取完
const isFullyClaimed = computed(() => {
  if (!benefit.value) return false;
  
  // 如果claimed_count和total_count都存在，且claimed_count >= total_count，则福利已被领取完
  return (
    benefit.value.claimed_count !== undefined && 
    benefit.value.total_count !== undefined && 
    benefit.value.total_count > 0 && 
    benefit.value.claimed_count >= benefit.value.total_count
  );
});

// 格式化日期
const formatDate = (dateStr) => {
  if (!dateStr) return '未设置';
  
  try {
    const date = new Date(dateStr);
    if (isNaN(date.getTime())) {
      return '未设置';
    }
    return date.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    });
  } catch (err) {
    console.error('Date formatting error:', err);
    return '未设置';
  }
};

// 获取福利信息
onMounted(async () => {
  const uuid = route.params.uuid;
  if (!uuid) {
    error.value = '无效的福利ID';
    loading.value = false;
    return;
  }
  
  try {
    // 获取福利详情
    const response = await benefitApi.getBenefitByUuid(uuid);
    benefit.value = response.benefit;
    claimStatus.value = response.claim_status;
    
    console.log('Benefit details:', benefit.value);
    console.log('Claim status:', claimStatus.value);
    
    // 如果已登录，检查用户是否已领取
    if (isAuthenticated.value) {
      const claims = await benefitStore.fetchMyClaims();
      userClaimCount.value = claims.filter(claim => claim.benefit_uuid === uuid).length;
    }
  } catch (err) {
    error.value = '获取福利信息失败';
    console.error(err);
  } finally {
    loading.value = false;
  }
});

// 跳转到登录页
const goToLogin = () => {
  // 设置回调地址，登录后返回
  router.push({
    path: '/login',
    query: { redirect: route.fullPath }
  });
};

// 返回首页
const goHome = () => {
  router.push('/');
};

// 查看我的领取记录
const viewMyClaims = () => {
  router.push('/dashboard/claims');
};

// 领取福利
const claimBenefit = async () => {
  if (!isAuthenticated.value) {
    goToLogin();
    return;
  }
  
  claimLoading.value = true;
  
  try {
    const response = await benefitStore.claimBenefit(benefit.value.uuid);
    claimedCode.value = response.code;
    claimSuccess.value = true;
    
    // 更新领取计数
    userClaimCount.value++;
  } catch (err) {
    ElMessage.error(err.message || '领取福利失败');
    console.error(err);
  } finally {
    claimLoading.value = false;
  }
};

// 复制兑换码
const copyCode = () => {
  if (navigator.clipboard) {
    navigator.clipboard.writeText(claimedCode.value)
      .then(() => {
        ElMessage.success('兑换码已复制到剪贴板');
      })
      .catch(() => {
        ElMessage.warning('复制失败，请手动复制');
      });
  } else {
    ElMessage.warning('您的浏览器不支持自动复制，请手动复制');
  }
};
</script>

<style scoped>
.claim-benefit {
  min-height: 100vh;
  background: linear-gradient(135deg, #f5f7fa, #c3cfe2);
  padding: 2rem 1rem;
  display: flex;
  align-items: center;
  justify-content: center;
}

.container {
  width: 100%;
  max-width: 700px;
  margin: 0 auto;
}

.loading-state, .error-state, .success-state {
  background: white;
  border-radius: 8px;
  padding: 2rem;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.benefit-card {
  background: white;
  border-radius: 8px;
  overflow: hidden;
}

.benefit-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid #eee;
}

.tag-container {
  display: flex;
  gap: 8px;
}

.benefit-header h1 {
  margin: 0;
  font-size: 1.8rem;
  font-weight: 500;
}

.benefit-info {
  margin-bottom: 2rem;
}

.benefit-info h3 {
  font-size: 1.2rem;
  margin: 0 0 1rem;
  color: #333;
}

.description {
  margin: 0 0 1.5rem;
  color: #666;
  line-height: 1.6;
}

.info-items {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
}

.info-item {
  display: flex;
  align-items: center;
  color: #666;
}

.info-item .el-icon {
  margin-right: 8px;
  color: #409eff;
}

.claim-action {
  background: #f9f9f9;
  border-top: 1px solid #eee;
  padding: 1.5rem;
  text-align: center;
}

.notice {
  margin: 0 0 1rem;
  color: #666;
}

.code-container {
  margin: 2rem 0;
}

.code-card {
  background: #f9f9f9;
  border: 1px dashed #ddd;
  border-radius: 6px;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  max-width: 400px;
  margin: 0 auto;
}

.code-card h3 {
  margin: 0 0 1rem;
  color: #333;
}

.code-value {
  font-family: monospace;
  font-size: 1.4rem;
  background: #fff;
  border: 1px solid #eee;
  padding: 0.8rem 1.5rem;
  border-radius: 4px;
  margin-bottom: 1rem;
  word-break: break-all;
  width: 100%;
  text-align: center;
}

.action-buttons {
  display: flex;
  justify-content: center;
  gap: 1rem;
}

@media (max-width: 768px) {
  .claim-benefit {
    padding: 1rem;
  }
  
  .benefit-header {
    flex-direction: column;
    align-items: flex-start;
  }
  
  .benefit-header h1 {
    margin-bottom: 0.5rem;
    font-size: 1.5rem;
  }
  
  .info-items {
    grid-template-columns: 1fr;
  }
  
  .action-buttons {
    flex-direction: column;
  }
}
</style>