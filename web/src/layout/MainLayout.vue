<template>
  <el-container class="app-wrapper">
    <!-- 侧边栏 -->
    <el-aside width="240px" class="sidebar">
      <div class="logo">
        <div class="logo-icon">U</div>
        <span>UCAS</span>
      </div>
      
      <el-menu
        :default-active="activeMenu"
        class="side-menu"
        router
      >
        <el-menu-item index="/">
          <el-icon><LayoutDashboard :size="20" /></el-icon>
          <span>仪表盘概览</span>
        </el-menu-item>
        <el-menu-item index="/accounts">
          <el-icon><User :size="20" /></el-icon>
          <span>账号管理</span>
        </el-menu-item>
        <el-menu-item index="/tasks">
          <el-icon><ListChecks :size="20" /></el-icon>
          <span>任务列表</span>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <el-container>
      <!-- 顶栏 -->
      <el-header height="64px" class="navbar">
        <div class="header-left">
          <el-breadcrumb separator="/">
            <el-breadcrumb-item>首页</el-breadcrumb-item>
            <el-breadcrumb-item>{{ currentPageTitle }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="header-right">
          <el-button 
            circle 
            :icon="isDark ? Sun : Moon" 
            @click="toggleDark()"
            class="theme-toggle"
          />
          <el-divider direction="vertical" />
          <el-avatar :size="32" src="https://github.com/identicons/user.png" />
        </div>
      </el-header>

      <!-- 内容区 -->
      <el-main class="main-content">
        <router-view v-slot="{ Component }">
          <transition name="fade-transform" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useDark, useToggle } from '@vueuse/core'
import { 
  LayoutDashboard, 
  User, 
  ListChecks, 
  Moon, 
  Sun 
} from 'lucide-vue-next'

const route = useRoute()
const isDark = useDark()
const toggleDark = useToggle(isDark)

const activeMenu = computed(() => route.path)
const currentPageTitle = computed(() => {
  const titles = {
    '/': '仪表盘',
    '/accounts': '账号管理',
    '/tasks': '任务管理'
  }
  return titles[route.path] || '概览'
})
</script>

<style scoped>
.app-wrapper {
  height: 100vh;
}

.sidebar {
  background: white;
  border-right: 1px solid rgba(0, 0, 0, 0.05);
  display: flex;
  flex-direction: column;
}

html.dark .sidebar {
  background: #111827;
  border-right: 1px solid rgba(255, 255, 255, 0.05);
}

.logo {
  height: 64px;
  display: flex;
  align-items: center;
  padding: 0 24px;
  gap: 12px;
  font-size: 18px;
  font-weight: 800;
  color: var(--el-color-primary);
}

.logo-icon {
  width: 32px;
  height: 32px;
  background: var(--el-color-primary);
  color: white;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.side-menu {
  border-right: none;
  padding: 0 12px;
}

.el-menu-item {
  height: 48px;
  line-height: 48px;
  margin: 4px 0;
  border-radius: 8px;
  color: #64748b;
}

.el-menu-item.is-active {
  background: #f1f5f9;
  color: var(--el-color-primary);
}

html.dark .el-menu-item.is-active {
  background: rgba(79, 70, 229, 0.1);
}

.navbar {
  background: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(8px);
  border-bottom: 1px solid rgba(0, 0, 0, 0.05);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  z-index: 10;
}

html.dark .navbar {
  background: rgba(17, 24, 39, 0.8);
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.fade-transform-enter-active,
.fade-transform-leave-active {
  transition: all 0.3s;
}

.fade-transform-enter-from {
  opacity: 0;
  transform: translateX(-20px);
}

.fade-transform-leave-to {
  opacity: 0;
  transform: translateX(20px);
}
</style>
