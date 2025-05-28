<template>
  <div class="my-benefits">
    <div class="page-header">
      <div class="left">
        <h2>我的福利</h2>
        <p>管理您创建的所有福利</p>
      </div>
      <div class="right">
        <el-button type="primary" @click="createBenefit">
          <el-icon><Plus /></el-icon>
          发布新福利
        </el-button>
      </div>
    </div>
    
    <el-tabs v-model="activeTab" class="benefit-tabs">
      <el-tab-pane label="活跃福利" name="active">
        <el-skeleton :rows="3" animated v-if="loading" />
        
        <template v-else>
          <div v-if="activeBenefits.length === 0" class="empty-state">
            <el-empty description="暂无活跃福利" />
          </div>
          
          <el-card v-for="benefit in activeBenefits" :key="benefit.uuid" class="benefit-card">
            <div class="benefit-header">
              <h3>{{ benefit.title }}</h3>
              <el-tag type="success" v-if="benefit.status === 'active'">活跃</el-tag>
              <el-tag type="warning" v-else-if="benefit.status === 'inactive'">未激活</el-tag>
            </div>
            
            <p class="benefit-description">{{ benefit.description }}</p>
            
            <div class="benefit-meta">
              <div class="meta-item">
                <el-icon><Calendar /></el-icon>
                <span>过期时间: {{ formatDate(benefit.expires_at) }}</span>
              </div>
              <div class="meta-item">
                <el-icon><TicketIcon /></el-icon>
                <span>兑换码: {{ benefit.total_count }} 个</span>
              </div>
              <div class="meta-item">
                <el-icon><User /></el-icon>
                <span>已领取: {{ benefit.claimed_count }} / {{ benefit.total_count }}</span>
              </div>
            </div>
            
            <div class="benefit-actions">
              <el-button-group>
                <el-button @click="viewBenefit(benefit)">详情</el-button>
                <el-button type="primary" @click="shareBenefit(benefit)">分享</el-button>
                <el-button 
                  type="danger" 
                  @click="stopBenefit(benefit)"
                  v-if="benefit.status === 'active'"
                >停用</el-button>
                <el-button 
                  type="success" 
                  @click="activateBenefit(benefit)"
                  v-else-if="benefit.status === 'inactive'"
                >激活</el-button>
              </el-button-group>
            </div>
          </el-card>
        </template>
      </el-tab-pane>
      
      <el-tab-pane label="已过期/停用" name="expired">
        <el-skeleton :rows="3" animated v-if="loading" />
        
        <template v-else>
          <div v-if="expiredBenefits.length === 0" class="empty-state">
            <el-empty description="暂无已过期或停用的福利" />
          </div>
          
          <el-card v-for="benefit in expiredBenefits" :key="benefit.uuid" class="benefit-card">
            <div class="benefit-header">
              <h3>{{ benefit.title }}</h3>
              <el-tag type="info" v-if="benefit.status === 'expired'">已过期</el-tag>
              <el-tag type="danger" v-else-if="benefit.status === 'deleted'">已停用</el-tag>
            </div>
            
            <p class="benefit-description">{{ benefit.description }}</p>
            
            <div class="benefit-meta">
              <div class="meta-item">
                <el-icon><Calendar /></el-icon>
                <span>过期时间: {{ formatDate(benefit.expires_at) }}</span>
              </div>
              <div class="meta-item">
                <el-icon><TicketIcon /></el-icon>
                <span>兑换码: {{ benefit.total_count }} 个</span>
              </div>
              <div class="meta-item">
                <el-icon><User /></el-icon>
                <span>已领取: {{ benefit.claimed_count }} / {{ benefit.total_count }}</span>
              </div>
            </div>
            
            <div class="benefit-actions">
              <el-button-group>
                <el-button @click="viewBenefit(benefit)">详情</el-button>
                <el-button 
                  type="success" 
                  @click="reactivateBenefit(benefit)"
                  v-if="benefit.status === 'deleted' && !isExpired(benefit.expires_at)"
                >重新激活</el-button>
              </el-button-group>
            </div>
          </el-card>
        </template>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import { useBenefitStore } from '../../stores/benefit';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Plus, Calendar, Ticket as TicketIcon, User } from '@element-plus/icons-vue';

const router = useRouter();
const benefitStore = useBenefitStore();
const loading = ref(true);
const activeTab = ref('active');

// 获取福利列表
onMounted(async () => {
  try {
    await benefitStore.fetchMyBenefits();
  } catch (err) {
    ElMessage.error('获取福利列表失败');
    console.error(err);
  } finally {
    loading.value = false;
  }
});

// 活跃福利
const activeBenefits = computed(() => {
  return benefitStore.activeBenefits;
});

// 已过期/已停用福利
const expiredBenefits = computed(() => {
  return benefitStore.expiredBenefits;
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

// 检查是否已过期
const isExpired = (dateStr) => {
  return new Date(dateStr) < new Date();
};

// 创建福利
const createBenefit = () => {
  router.push('/dashboard/benefits/create');
};

// 查看福利详情
const viewBenefit = (benefit) => {
  router.push(`/dashboard/benefits/${benefit.uuid}`);
};

// 分享福利链接
const shareBenefit = (benefit) => {
  const url = `${window.location.origin}/claim/${benefit.uuid}`;
  
  // 尝试使用剪贴板API
  if (navigator.clipboard) {
    navigator.clipboard.writeText(url)
      .then(() => {
        ElMessage.success('福利链接已复制到剪贴板');
      })
      .catch(() => {
        showShareDialog(url);
      });
  } else {
    showShareDialog(url);
  }
};

// 显示分享对话框
const showShareDialog = (url) => {
  ElMessageBox.alert(url, '分享福利链接', {
    confirmButtonText: '确定',
    center: true,
    callback: () => {
      ElMessage({
        type: 'info',
        message: '请复制以上链接分享给用户'
      });
    }
  });
};

// 停用福利
const stopBenefit = (benefit) => {
  ElMessageBox.confirm('确定要停用该福利吗？停用后用户将无法领取', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await benefitStore.updateBenefitStatus(benefit.uuid, 'deleted');
      ElMessage.success('福利已停用');
    } catch (err) {
      ElMessage.error('停用福利失败');
      console.error(err);
    }
  }).catch(() => {});
};

// 激活福利
const activateBenefit = async (benefit) => {
  try {
    await benefitStore.updateBenefitStatus(benefit.uuid, 'active');
    ElMessage.success('福利已激活');
  } catch (err) {
    ElMessage.error('激活福利失败');
    console.error(err);
  }
};

// 重新激活已停用的福利
const reactivateBenefit = async (benefit) => {
  try {
    await benefitStore.updateBenefitStatus(benefit.uuid, 'active');
    ElMessage.success('福利已重新激活');
  } catch (err) {
    ElMessage.error('重新激活福利失败');
    console.error(err);
  }
};
</script>

<style scoped>
.my-benefits {
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0 0 5px;
  font-size: 1.5rem;
  font-weight: 500;
}

.page-header p {
  margin: 0;
  color: #666;
  font-size: 0.9rem;
}

.benefit-tabs {
  margin-top: 20px;
}

.benefit-card {
  margin-bottom: 15px;
}

.benefit-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.benefit-header h3 {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 500;
}

.benefit-description {
  color: #666;
  margin-bottom: 15px;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.benefit-meta {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
  margin-bottom: 15px;
}

.meta-item {
  display: flex;
  align-items: center;
  color: #666;
  font-size: 0.9rem;
}

.meta-item .el-icon {
  margin-right: 5px;
  color: #409eff;
}

.benefit-actions {
  display: flex;
  justify-content: flex-end;
}

.empty-state {
  margin: 30px 0;
}

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
  }
  
  .page-header .right {
    margin-top: 10px;
    width: 100%;
  }
  
  .page-header .right .el-button {
    width: 100%;
  }
  
  .benefit-meta {
    grid-template-columns: 1fr;
  }
}
</style> 