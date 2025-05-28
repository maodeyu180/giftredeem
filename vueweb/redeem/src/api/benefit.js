import api from './index';

// 福利相关的API
export const benefitApi = {
  // 创建新福利
  createBenefit: (data) => api.post('/benefits', data),
  
  // 获取当前用户创建的福利
  getUserBenefits: () => api.get('/benefits/my'),
  
  // 更新福利状态
  updateBenefitStatus: (uuid, status) => api.put(`/benefits/${uuid}/status`, { status }),
  
  // 获取福利的领取记录
  getBenefitClaims: (uuid) => api.get(`/benefits/${uuid}/claims`),
  
  // 获取福利详情（通过UUID）
  getBenefitByUuid: (uuid) => api.get(`/claim/${uuid}`),
  
  // 领取福利
  claimBenefit: (uuid) => api.post(`/claim/${uuid}`),
  
  // 获取当前用户领取的福利
  getUserClaims: () => api.get('/claims/my'),
}; 