<template>
  <div class="create-benefit">
    <el-card class="form-card">
      <template #header>
        <div class="card-header">
          <h2>发布新福利</h2>
          <p>创建一个新的福利，生成后可分享给用户领取</p>
        </div>
      </template>
      
      <el-form 
        ref="formRef" 
        :model="form" 
        :rules="rules" 
        label-width="100px"
        label-position="top"
        class="benefit-form"
      >
        <!-- 基本信息 -->
        <el-form-item label="福利标题" prop="title">
          <el-input v-model="form.title" placeholder="请输入福利标题" />
        </el-form-item>
        
        <el-form-item label="福利描述" prop="description">
          <el-input 
            v-model="form.description" 
            type="textarea" 
            :rows="3" 
            placeholder="请详细描述福利内容、使用方法等" 
          />
        </el-form-item>
        
        <!-- 兑换码 -->
        <el-form-item label="兑换码" prop="codes">
          <el-input 
            v-model="form.codes" 
            type="textarea" 
            :rows="5" 
            placeholder="请输入兑换码，每行一个。系统将为每个访问用户随机分配一个兑换码。" 
          />
          <div class="tip">已输入 {{ codeCount }} 个兑换码</div>
        </el-form-item>
        
        <!-- 有效期 -->
        <el-form-item label="有效期" prop="expireAt">
          <el-date-picker
            v-model="form.expireAt"
            type="datetime"
            placeholder="请选择过期时间"
            format="YYYY-MM-DD HH:mm"
            value-format="YYYY-MM-DD HH:mm:ss"
            :disabled-date="disabledDate"
          />
        </el-form-item>
        
        <!-- 高级选项 -->
        <el-collapse>
          <el-collapse-item title="高级选项" name="advanced">
            <el-form-item label="每人限领次数" prop="claimLimit">
              <el-input-number
                v-model="form.claimLimit"
                :min="1"
                :max="99"
                :step="1"
                controls-position="right"
              />
              <span class="option-hint">每个用户最多可领取的次数</span>
            </el-form-item>
            
            <el-form-item label="总领取次数限制" prop="totalLimit">
              <el-input-number
                v-model="form.totalLimit"
                :min="0"
                :max="9999"
                :step="1"
                controls-position="right"
              />
              <span class="option-hint">0 表示不限制总领取次数</span>
            </el-form-item>
            
            <el-form-item label="状态" prop="status">
              <el-radio-group v-model="form.status">
                <el-radio :value="'active'">立即生效</el-radio>
                <el-radio :value="'inactive'">创建后手动激活</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-collapse-item>
        </el-collapse>
        
        <div class="form-actions">
          <el-button @click="resetForm">重置</el-button>
          <el-button type="primary" @click="submitForm" :loading="loading">创建福利</el-button>
        </div>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, reactive } from 'vue';
import { useRouter } from 'vue-router';
import { useBenefitStore } from '../../stores/benefit';
import { ElMessage } from 'element-plus';

const router = useRouter();
const benefitStore = useBenefitStore();
const formRef = ref(null);
const loading = ref(false);

// 表单数据
const form = reactive({
  title: '',
  description: '',
  codes: '',
  expireAt: '',
  claimLimit: 1,
  totalLimit: 0,
  status: 'active'
});

// 表单验证规则
const rules = {
  title: [
    { required: true, message: '请输入福利标题', trigger: 'blur' },
    { min: 2, max: 50, message: '标题长度应在 2 到 50 个字符之间', trigger: 'blur' }
  ],
  description: [
    { required: true, message: '请输入福利描述', trigger: 'blur' },
    { min: 10, max: 500, message: '描述长度应在 10 到 500 个字符之间', trigger: 'blur' }
  ],
  codes: [
    { required: true, message: '请输入至少一个兑换码', trigger: 'blur' }
  ],
  expireAt: [
    { required: true, message: '请选择过期时间', trigger: 'change' }
  ]
};

// 计算兑换码数量
const codeCount = computed(() => {
  if (!form.codes) return 0;
  // 分割每行，过滤空行
  const codes = form.codes.split('\n').filter(code => code.trim() !== '');
  return codes.length;
});

// 禁用过去的日期
const disabledDate = (time) => {
  return time.getTime() < Date.now() - 8.64e7; // 禁用今天之前的日期
};

// 重置表单
const resetForm = () => {
  formRef.value.resetFields();
};

// 提交表单
const submitForm = async () => {
  await formRef.value.validate(async (valid) => {
    if (!valid) {
      ElMessage.error('请完善表单信息');
      return;
    }
    
    // 验证兑换码
    if (codeCount.value === 0) {
      ElMessage.error('请输入至少一个兑换码');
      return;
    }
    
    loading.value = true;
    
    try {
      // 处理兑换码
      const codes = form.codes.split('\n')
        .filter(code => code.trim() !== '')
        .map(code => code.trim());
      
      // 准备提交的数据
      const benefitData = {
        title: form.title,
        description: form.description,
        codes: codes,
        expire_at: form.expireAt,
        claim_limit: form.claimLimit,
        total_limit: form.totalLimit,
        status: form.status
      };
      
      // 提交创建请求
      const response = await benefitStore.createBenefit(benefitData);
      
      ElMessage.success('福利创建成功');
      
      // 跳转到福利详情页
      router.push(`/dashboard/benefits/${response.benefit.uuid}`);
    } catch (err) {
      ElMessage.error(err.message || '创建福利失败，请重试');
    } finally {
      loading.value = false;
    }
  });
};
</script>

<style scoped>
.create-benefit {
  max-width: 800px;
  margin: 0 auto;
}

.form-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  flex-direction: column;
}

.card-header h2 {
  margin: 0 0 8px;
  font-size: 1.5rem;
  font-weight: 500;
}

.card-header p {
  margin: 0;
  color: #666;
  font-size: 0.9rem;
}

.benefit-form {
  margin-top: 20px;
}

.tip {
  font-size: 0.85rem;
  color: #909399;
  margin-top: 5px;
}

.option-hint {
  margin-left: 10px;
  color: #909399;
  font-size: 0.85rem;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 30px;
  gap: 10px;
}

@media (max-width: 768px) {
  .create-benefit {
    padding: 0 10px;
  }
  
  .option-hint {
    display: block;
    margin-left: 0;
    margin-top: 5px;
  }
}
</style> 