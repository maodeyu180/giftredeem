import { defineStore } from 'pinia';
import { benefitApi } from '../api/benefit';
import { ref, computed } from 'vue';

export const useBenefitStore = defineStore('benefit', () => {
  // 状态
  const myBenefits = ref([]);
  const myClaims = ref([]);
  const currentBenefit = ref(null);
  const loading = ref(false);
  const error = ref(null);

  // 计算属性
  const activeBenefits = computed(() => 
    myBenefits.value.filter(b => b.status === 'active')
  );
  
  const expiredBenefits = computed(() => 
    myBenefits.value.filter(b => b.status === 'expired' || b.status === 'deleted')
  );

  // 动作
  // 创建福利
  async function createBenefit(benefitData) {
    loading.value = true;
    error.value = null;
    try {
      const response = await benefitApi.createBenefit(benefitData);
      myBenefits.value.unshift(response.benefit);
      return response;
    } catch (err) {
      error.value = err.message || '创建福利失败';
      throw err;
    } finally {
      loading.value = false;
    }
  }

  // 获取我创建的福利
  async function fetchMyBenefits() {
    loading.value = true;
    error.value = null;
    try {
      const response = await benefitApi.getUserBenefits();
      myBenefits.value = response.benefits;
      return myBenefits.value;
    } catch (err) {
      error.value = err.message || '获取福利列表失败';
      throw err;
    } finally {
      loading.value = false;
    }
  }

  // 更新福利状态
  async function updateBenefitStatus(uuid, status) {
    loading.value = true;
    error.value = null;
    try {
      await benefitApi.updateBenefitStatus(uuid, status);
      
      // 更新本地状态
      const index = myBenefits.value.findIndex(b => b.uuid === uuid);
      if (index !== -1) {
        myBenefits.value[index].status = status;
      }
      
      return true;
    } catch (err) {
      error.value = err.message || '更新福利状态失败';
      throw err;
    } finally {
      loading.value = false;
    }
  }

  // 获取福利领取记录
  async function fetchBenefitClaims(uuid) {
    loading.value = true;
    error.value = null;
    try {
      const response = await benefitApi.getBenefitClaims(uuid);
      return response.claims;
    } catch (err) {
      error.value = err.message || '获取领取记录失败';
      throw err;
    } finally {
      loading.value = false;
    }
  }

  // 通过UUID获取福利
  async function getBenefitByUuid(uuid) {
    loading.value = true;
    error.value = null;
    try {
      const response = await benefitApi.getBenefitByUuid(uuid);
      currentBenefit.value = response.benefit;
      return currentBenefit.value;
    } catch (err) {
      error.value = err.message || '获取福利详情失败';
      throw err;
    } finally {
      loading.value = false;
    }
  }

  // 领取福利
  async function claimBenefit(uuid) {
    loading.value = true;
    error.value = null;
    try {
      const response = await benefitApi.claimBenefit(uuid);
      return response;
    } catch (err) {
      error.value = err.message || '领取福利失败';
      throw err;
    } finally {
      loading.value = false;
    }
  }

  // 获取我领取的福利
  async function fetchMyClaims() {
    loading.value = true;
    error.value = null;
    try {
      const response = await benefitApi.getUserClaims();
      myClaims.value = response.claims;
      return myClaims.value;
    } catch (err) {
      error.value = err.message || '获取已领取福利失败';
      throw err;
    } finally {
      loading.value = false;
    }
  }

  return {
    // 状态
    myBenefits,
    myClaims,
    currentBenefit,
    loading,
    error,
    // 计算属性
    activeBenefits,
    expiredBenefits,
    // 方法
    createBenefit,
    fetchMyBenefits,
    updateBenefitStatus,
    fetchBenefitClaims,
    getBenefitByUuid,
    claimBenefit,
    fetchMyClaims
  };
}); 