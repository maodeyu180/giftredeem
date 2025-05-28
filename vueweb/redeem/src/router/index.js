import { createRouter, createWebHistory } from 'vue-router';

// 懒加载路由组件
const Home = () => import('../views/Home.vue');
const Login = () => import('../views/auth/Login.vue');
const Callback = () => import('../views/auth/Callback.vue');
const Dashboard = () => import('../views/layout/Dashboard.vue');
const MyBenefits = () => import('../views/benefit/MyBenefits.vue');
const CreateBenefit = () => import('../views/benefit/CreateBenefit.vue');
const BenefitDetail = () => import('../views/benefit/BenefitDetail.vue');
const MyClaims = () => import('../views/claim/MyClaims.vue');
const ClaimBenefit = () => import('../views/claim/ClaimBenefit.vue');
const NotFound = () => import('../views/NotFound.vue');

// 创建路由
const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: Home,
      meta: { title: '首页' }
    },
    {
      path: '/login',
      name: 'login',
      component: Login,
      meta: { title: '登录' }
    },
    {
      path: '/auth/callback/:provider',
      name: 'callback',
      component: Callback,
      meta: { title: '登录回调' }
    },
    {
      path: '/dashboard',
      name: 'dashboard',
      component: Dashboard,
      meta: { requiresAuth: true, title: '控制台' },
      children: [
        {
          path: 'benefits',
          name: 'my-benefits',
          component: MyBenefits,
          meta: { requiresAuth: true, title: '我发布的福利' }
        },
        {
          path: 'benefits/create',
          name: 'create-benefit',
          component: CreateBenefit,
          meta: { requiresAuth: true, title: '发布新福利' }
        },
        {
          path: 'benefits/:uuid',
          name: 'benefit-detail',
          component: BenefitDetail,
          meta: { requiresAuth: true, title: '福利详情' }
        },
        {
          path: 'claims',
          name: 'my-claims',
          component: MyClaims,
          meta: { requiresAuth: true, title: '我领取的福利' }
        }
      ]
    },
    {
      path: '/claim/:uuid',
      name: 'claim-benefit',
      component: ClaimBenefit,
      meta: { title: '领取福利' }
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: NotFound,
      meta: { title: '页面未找到' }
    }
  ]
});

// 导航守卫
router.beforeEach((to, from, next) => {
  // 设置页面标题
  document.title = `${to.meta.title || '福利兑换平台'} - GiftRedeem`;
  
  // 检查是否需要认证
  const requiresAuth = to.matched.some(record => record.meta.requiresAuth);
  const token = localStorage.getItem('token');
  
  if (requiresAuth && !token) {
    next({ name: 'login', query: { redirect: to.fullPath } });
  } else {
    next();
  }
});

export default router; 