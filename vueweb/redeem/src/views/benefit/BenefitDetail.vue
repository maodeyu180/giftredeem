<template>
  <div class="benefit-detail">
    <div class="page-header">
      <el-page-header @back="goBack">
        <template #content>
          <div v-if="!loading && benefit">{{ benefit.title }} - 详情</div>
          <div v-else>福利详情</div>
        </template>
      </el-page-header>
    </div>
    
    <div class="page-content">
      <el-skeleton :rows="10" animated v-if="loading" />
      
      <template v-else-if="benefit">
        <el-tabs v-model="activeTab" @tab-change="handleTabChange">
          <el-tab-pane label="基本信息" name="info">
            <el-card>
              <template #header>
                <div class="card-header">
                  <div class="left">
                    <h3>{{ benefit.title }}</h3>
                    <el-tag type="success" v-if="benefit.status === 'active'">活跃</el-tag>
                    <el-tag type="warning" v-else-if="benefit.status === 'inactive'">未激活</el-tag>
                    <el-tag type="info" v-else-if="benefit.status === 'expired'">已过期</el-tag>
                    <el-tag type="danger" v-else-if="benefit.status === 'deleted'">已停用</el-tag>
                  </div>
                  <div class="right">
                    <el-button-group>
                      <el-button @click="shareBenefit">分享链接</el-button>
                      <el-button 
                        type="danger" 
                        @click="stopBenefit"
                        v-if="benefit.status === 'active'"
                      >停用</el-button>
                      <el-button 
                        type="success" 
                        @click="activateBenefit"
                        v-else-if="benefit.status === 'inactive' || benefit.status === 'deleted'"
                      >激活</el-button>
                    </el-button-group>
                  </div>
                </div>
              </template>
              
              <div class="benefit-info">
                <div class="info-section">
                  <h4>福利描述</h4>
                  <p class="description">{{ benefit.description }}</p>
                </div>
                
                <div class="info-section">
                  <h4>基本信息</h4>
                  <div class="info-grid">
                    <div class="info-item">
                      <span class="label">总兑换码数量:</span>
                      <span class="value">{{ benefit.total_count }}</span>
                    </div>
                    <div class="info-item">
                      <span class="label">已领取数量:</span>
                      <span class="value">{{ benefit.claimed_count }}</span>
                    </div>
                    <div class="info-item">
                      <span class="label">创建时间:</span>
                      <span class="value">{{ formatDate(benefit.created_at) }}</span>
                    </div>
                    <div class="info-item">
                      <span class="label">过期时间:</span>
                      <span class="value">{{ formatDate(benefit.expires_at) }}</span>
                    </div>
                    <div class="info-item">
                      <span class="label">福利状态:</span>
                      <span class="value">{{ getStatusText(benefit.status) }}</span>
                    </div>
                    <div class="info-item">
                      <span class="label">分享链接:</span>
                      <span class="value link">
                        {{ getClaimUrl() }}
                        <el-button 
                          size="small" 
                          type="primary" 
                          @click="copyClaimUrl" 
                          circle
                        >
                          <el-icon><CopyDocument /></el-icon>
                        </el-button>
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            </el-card>
          </el-tab-pane>
          
          <el-tab-pane label="领取记录" name="claims">
            <el-card>
              <template #header>
                <div class="card-header">
                  <h3>领取记录</h3>
                </div>
              </template>
              
              <el-skeleton :rows="5" animated v-if="claimsLoading" />
              
              <template v-else>
                <div v-if="claims.length === 0" class="empty-state">
                  <el-empty description="暂无领取记录" />
                </div>
                
                <el-table v-else :data="claims" stripe style="width: 100%">
                  <el-table-column label="用户信息" width="200">
                    <template #default="scope">
                      <div class="user-info">
                        <el-avatar :size="32" :src="scope.row.user.avatar_url || ''" />
                        <span class="username">{{ scope.row.user.username || '未知用户' }}</span>
                      </div>
                    </template>
                  </el-table-column>
                  <el-table-column prop="oauth_provider" label="认证提供商" width="120" />
                  <el-table-column label="领取时间" width="180">
                    <template #default="scope">
                      {{ formatDate(scope.row.claimed_at) }}
                    </template>
                  </el-table-column>
                  <el-table-column label="兑换码" min-width="150">
                    <template #default="scope">
                      <el-tag>{{ scope.row.code }}</el-tag>
                    </template>
                  </el-table-column>
                </el-table>
              </template>
            </el-card>
          </el-tab-pane>
        </el-tabs>
      </template>
      
      <div v-else class="error-state">
        <el-result
          icon="error"
          title="获取福利详情失败"
          sub-title="无法加载福利详情，请稍后重试"
        >
          <template #extra>
            <el-button @click="goBack">返回</el-button>
            <el-button type="primary" @click="fetchBenefitDetails">重试</el-button>
          </template>
        </el-result>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useBenefitStore } from '../../stores/benefit';
import { ElMessage, ElMessageBox } from 'element-plus';
import { CopyDocument } from '@element-plus/icons-vue';

const route = useRoute();
const router = useRouter();
const benefitStore = useBenefitStore();

const loading = ref(true);
const claimsLoading = ref(false);
const benefit = ref(null);
const claims = ref([]);
const activeTab = ref('info');

// 获取福利详情
const fetchBenefitDetails = async () => {
  const uuid = route.params.uuid;
  if (!uuid) {
    goBack();
    return;
  }
  
  loading.value = true;
  
  try {
    // 使用 benefitStore 的 getBenefitByUuid 方法，但是需要适配一下
    // 因为这个方法是为了领取页面设计的，不是为了管理员查看详情
    // 在实际开发中，可能需要添加一个专门的管理员获取福利详情的接口
    benefit.value = await benefitStore.getBenefitByUuid(uuid);
    
    // 如果是切换到 claims 标签页，自动加载领取记录
    if (activeTab.value === 'claims') {
      fetchClaims();
    }
  } catch (err) {
    ElMessage.error('获取福利详情失败');
    console.error(err);
    benefit.value = null;
  } finally {
    loading.value = false;
  }
};

// 获取领取记录
const fetchClaims = async () => {
  if (!benefit.value || !benefit.value.uuid) return;
  
  claimsLoading.value = true;
  
  try {
    claims.value = await benefitStore.fetchBenefitClaims(benefit.value.uuid);
  } catch (err) {
    ElMessage.error('获取领取记录失败');
    console.error(err);
    claims.value = [];
  } finally {
    claimsLoading.value = false;
  }
};

// 格式化日期
const formatDate = (dateStr) => {
  if (!dateStr) return '未设置';
  
  try {
    const date = new Date(dateStr);
    if (isNaN(date.getTime())) {
      return 'Invalid Date';
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
    return 'Invalid Date';
  }
};

// 获取状态文本
const getStatusText = (status) => {
  const statusMap = {
    'active': '活跃',
    'inactive': '未激活',
    'expired': '已过期',
    'deleted': '已停用'
  };
  return statusMap[status] || status;
};

// 获取领取链接
const getClaimUrl = () => {
  if (!benefit.value || !benefit.value.uuid) return '';
  return `${window.location.origin}/claim/${benefit.value.uuid}`;
};

// 复制领取链接
const copyClaimUrl = () => {
  const url = getClaimUrl();
  
  if (navigator.clipboard) {
    navigator.clipboard.writeText(url)
      .then(() => {
        ElMessage.success('领取链接已复制到剪贴板');
      })
      .catch(() => {
        ElMessage.warning('复制失败，请手动复制');
      });
  } else {
    ElMessage.warning('您的浏览器不支持自动复制，请手动复制');
  }
};

// 分享福利链接
const shareBenefit = () => {
  const url = getClaimUrl();
  
  ElMessageBox.alert(url, '分享福利链接', {
    confirmButtonText: '确定',
    center: true,
    callback: () => {}
  });
};

// 停用福利
const stopBenefit = () => {
  if (!benefit.value) return;
  
  ElMessageBox.confirm('确定要停用该福利吗？停用后用户将无法领取', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await benefitStore.updateBenefitStatus(benefit.value.uuid, 'deleted');
      ElMessage.success('福利已停用');
      benefit.value.status = 'deleted';
    } catch (err) {
      ElMessage.error('停用福利失败');
      console.error(err);
    }
  }).catch(() => {});
};

// 激活福利
const activateBenefit = async () => {
  if (!benefit.value) return;
  
  try {
    await benefitStore.updateBenefitStatus(benefit.value.uuid, 'active');
    ElMessage.success('福利已激活');
    benefit.value.status = 'active';
  } catch (err) {
    ElMessage.error('激活福利失败');
    console.error(err);
  }
};

// 返回
const goBack = () => {
  router.push('/dashboard/benefits');
};

// Tab change event handler
const handleTabChange = (tab) => {
  if (tab === 'claims' && benefit.value) {
    fetchClaims();
  }
};

// Watch for tab changes
watch(activeTab, (newTab) => {
  handleTabChange(newTab);
});

onMounted(() => {
  fetchBenefitDetails();
});
</script>

<style scoped>
.benefit-detail {
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header .left {
  display: flex;
  align-items: center;
}

.card-header h3 {
  margin: 0 10px 0 0;
  font-size: 1.2rem;
  font-weight: 500;
}

.benefit-info {
  margin-top: 20px;
}

.info-section {
  margin-bottom: 30px;
}

.info-section h4 {
  margin: 0 0 15px;
  font-size: 1.1rem;
  font-weight: 500;
  color: #333;
  border-bottom: 1px solid #eee;
  padding-bottom: 10px;
}

.description {
  color: #666;
  line-height: 1.6;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 15px;
}

.info-item {
  display: flex;
  flex-direction: column;
}

.label {
  color: #909399;
  font-size: 0.9rem;
  margin-bottom: 5px;
}

.value {
  font-weight: 500;
}

.value.link {
  display: flex;
  align-items: center;
  word-break: break-all;
}

.value.link .el-button {
  margin-left: 10px;
}

.user-info {
  display: flex;
  align-items: center;
}

.user-info .el-avatar {
  margin-right: 10px;
}

.empty-state {
  padding: 30px 0;
}

@media (max-width: 768px) {
  .card-header {
    flex-direction: column;
    align-items: flex-start;
  }
  
  .card-header .right {
    margin-top: 10px;
  }
  
  .info-grid {
    grid-template-columns: 1fr;
  }
}
</style> 