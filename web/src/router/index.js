import { createRouter, createWebHistory } from 'vue-router'
import MainLayout from '../layout/MainLayout.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      component: MainLayout,
      children: [
        {
          path: '',
          name: 'dashboard',
          component: () => import('../views/Dashboard.vue')
        },
        {
          path: 'accounts',
          name: 'accounts',
          component: () => import('../views/Accounts.vue')
        },
        {
          path: 'tasks',
          name: 'tasks',
          component: () => import('../views/Tasks.vue')
        }
      ]
    }
  ]
})

export default router
