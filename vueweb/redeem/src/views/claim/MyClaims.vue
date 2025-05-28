<template>
  <div class="my-claims">
    <div class="page-header">
      <div class="left">
        <h2>我的领取</h2>
        <p>查看您已领取的所有福利</p>
      </div>
    </div>
    
    <div class="claims-content">
      <el-skeleton :rows="3" animated v-if="loading" />
      
      <template v-else>
        <div v-if="claims.length === 0" class="empty-state">
          <el-empty description="您还没有领取过任何福利">
            <template #extra>
              <el-button type="primary" @click="goHome">去首页看看</el-button>
            </template>
          </el-empty>
        </div>
        
        <el-card v-for="claim in claims" :key="claim.id" class="claim-card">
          <div class="claim-header">
            <h3>{{ claim.benefit.title }}</h3>
            <span class="claimed-date">领取于: {{ formatDate(claim.claimed_at) }}</span>
          </div>
          
          <p class="description">{{ claim.benefit.description }}</p>
          
          <div class="code-section">
            <div class="code-card">
              <div class="code-header">
                <h4>兑换码</h4>
                <el-button size="small" type="primary" @click="copyCode(claim.code)">
                  <el-icon><CopyDocument /></el-icon> 复制兑换码
                </el-button>
              </div>
              <div class="code-value">{{ claim.code }}</div>
            </div>
          </div>
          
          <div class="claim-footer">
            <span class="provider">
              <el-icon><Connection /></el-icon>
              通过 {{ claim.oauth_provider }} 账号领取
            </span>
          </div>
        </el-card>
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useBenefitStore } from '../../stores/benefit';
import { ElMessage } from 'element-plus';
import { CopyDocument, Connection } from '@element-plus/icons-vue';

const router = useRouter();
const benefitStore = useBenefitStore();
const loading = ref(true);
const claims = ref([]);

// 获取我的领取记录
onMounted(async () => {
  try {
    claims.value = await benefitStore.fetchMyClaims();
    // 确保每个claim有benefit和code属性
    claims.value = claims.value.map(claim => {
      return {
        ...claim,
        benefit: claim.benefit || {},
        code: claim.code || 'N/A'
      };
    });
  } catch (err) {
    ElMessage.error('获取领取记录失败');
    console.error(err);
  } finally {
    loading.value = false;
  }
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

// 复制兑换码
const copyCode = (code) => {
  if (navigator.clipboard) {
    navigator.clipboard.writeText(code)
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

// 返回首页
const goHome = () => {
  router.push('/');
};
</script>

<style scoped>
.my-claims {
  max-width: 900px;
  margin: 0 auto;
}

.page-header {
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

.claims-content {
  margin-top: 20px;
}

.empty-state {
  margin: 40px 0;
  text-align: center;
}

.claim-card {
  margin-bottom: 20px;
}

.claim-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.claim-header h3 {
  margin: 0;
  font-size: 1.2rem;
  font-weight: 500;
}

.claimed-date {
  color: #909399;
  font-size: 0.9rem;
}

.description {
  color: #666;
  margin-bottom: 20px;
  line-height: 1.5;
}

.code-section {
  margin-bottom: 15px;
}

.code-card {
  background: #f9f9f9;
  border: 1px dashed #ddd;
  border-radius: 6px;
  padding: 15px;
}

.code-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.code-header h4 {
  margin: 0;
  font-size: 1rem;
  font-weight: 500;
  color: #333;
}

.code-value {
  font-family: monospace;
  font-size: 1.2rem;
  background: #fff;
  border: 1px solid #eee;
  padding: 10px 15px;
  border-radius: 4px;
  word-break: break-all;
}

.claim-footer {
  display: flex;
  justify-content: flex-end;
  margin-top: 10px;
}

.provider {
  display: flex;
  align-items: center;
  color: #909399;
  font-size: 0.9rem;
}

.provider .el-icon {
  margin-right: 5px;
  color: #409eff;
}

@media (max-width: 768px) {
  .claim-header {
    flex-direction: column;
    align-items: flex-start;
  }
  
  .claimed-date {
    margin-top: 5px;
  }
  
  .code-header {
    flex-direction: column;
    align-items: flex-start;
  }
  
  .code-header .el-button {
    margin-top: 10px;
  }
}
</style> 