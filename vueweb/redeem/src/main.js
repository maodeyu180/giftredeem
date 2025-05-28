import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import App from './App.vue'
import router from './router'

// 创建 Vue 应用
const app = createApp(App)

// 添加 Pinia 状态管理
app.use(createPinia())

// 添加 Element Plus UI 库
app.use(ElementPlus)

// 添加路由
app.use(router)

// 挂载应用
app.mount('#app')
