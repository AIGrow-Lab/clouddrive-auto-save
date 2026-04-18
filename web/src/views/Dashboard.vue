<template>
  <div class="dashboard-container">
    <div class="welcome-section">
      <h2>欢迎回来，管理员 👋</h2>
      <p>这是您今日的云端转存概览</p>
    </div>

    <el-row :gutter="24" class="stat-cards">
        <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card" body-style="padding: 20px">
          <div class="stat-icon purple">
            <Activity :size="24" />
          </div>
          <div class="stat-info">
            <div class="stat-label">已规划任务</div>
            <div class="stat-value">{{ stats.scheduled_tasks }}</div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card" body-style="padding: 20px">
          <div class="stat-icon blue">
            <HardDrive :size="24" />
          </div>
          <div class="stat-info">
            <div class="stat-label">已保存容量</div>
            <div class="stat-value">{{ formatSize(stats.capacity_used) }}</div>
          </div>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card" body-style="padding: 20px">
          <div class="stat-icon green">
            <RefreshCw :size="24" />
          </div>
          <div class="stat-info">
            <div class="stat-label">今日完成</div>
            <div class="stat-value">{{ stats.today_completed }}</div>
          </div>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card" body-style="padding: 20px">
          <div class="stat-icon orange">
            <User :size="24" />
          </div>
          <div class="stat-info">
            <div class="stat-label">活跃账号</div>
            <div class="stat-value">{{ stats.active_accounts }}</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="24" class="content-row">
      <!-- 左侧：实时日志终端 -->
      <el-col :xs="24" :lg="16">
        <el-card class="terminal-card" body-style="padding: 0">
          <template #header>
            <div class="card-header">
              <div class="header-title">
                <el-icon><Terminal /></el-icon>
                <span>实时日志流</span>
              </div>
              <div class="header-actions">
                <el-tooltip content="清空日志" placement="top">
                  <el-button link type="danger" :icon="Trash2" @click="clearLogs" />
                </el-tooltip>
              </div>
            </div>
          </template>
          <div class="terminal-window" ref="terminalRef">
            <div v-for="(log, index) in logs" :key="index" class="log-line" :class="getLogClass(log)">
              <span class="log-timestamp">{{ new Date().toLocaleTimeString() }}</span>
              <span class="log-content">{{ log }}</span>
            </div>
            <div v-if="logs.length === 0" class="terminal-empty">
              等待系统日志推送...
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 右侧：当前任务监控 -->
      <el-col :xs="24" :lg="8">
        <el-card class="monitor-card">
          <template #header>
            <div class="card-header">
              <span>任务执行监控</span>
              <el-tag size="small" type="info">{{ runningTasks.length }} 运行中</el-tag>
            </div>
          </template>
          
          <div class="running-tasks-list">
            <div v-for="task in runningTasks" :key="task.id" class="task-progress-card">
              <div class="task-info">
                <span class="task-name">{{ task.name }}</span>
                <div class="task-actions">
                  <el-icon v-if="task.percent < 100" class="is-loading"><Loader2 /></el-icon>
                  <el-icon v-else-if="task.stage === 'Success'" color="#67c23a"><CheckCircle2 /></el-icon>
                  <el-icon v-else-if="task.stage === 'Failed'" color="#f56c6c"><AlertCircle /></el-icon>
                  <el-button v-if="task.percent === 100" type="info" link @click="dismissTask(task.id)" style="margin-left: 8px; padding: 0">
                    <el-icon><X /></el-icon>
                  </el-button>
                </div>
              </div>
              <div class="task-stage">
                <el-tag size="small" :type="getStageTagType(task.stage)" effect="dark">{{ task.stage }}</el-tag>
                <span class="stage-msg">{{ task.message }}</span>
              </div>
              <el-progress 
                :percentage="task.percent" 
                :status="task.stage === 'Failed' ? 'exception' : (task.percent === 100 ? 'success' : '')"
                :stroke-width="10"
                striped
                :striped-flow="task.percent < 100"
              />
            </div>

            <div v-if="runningTasks.length === 0" class="monitor-empty">
              <el-empty :image-size="60" description="当前无运行中的任务" />
            </div>
          </div>

          <el-divider>近期动态</el-divider>
          <el-timeline class="compact-timeline">
            <el-timeline-item 
              v-for="activity in stats.recent_activities" 
              :key="activity.id"
              :timestamp="formatRelativeTime(activity.last_run)" 
              :type="getStatusType(activity.status)"
            >
              <div class="timeline-content">
                <span>{{ activity.name }}</span>
                <el-button v-if="activity.status === 'failed'" size="small" link type="primary" @click="handleRetry(activity.id)">重试</el-button>
              </div>
            </el-timeline-item>
          </el-timeline>
        </el-card>
      </el-col>
    </el-row>

    <!-- 浮动快捷操作 -->
    <div class="fab-container">
      <el-dropdown trigger="click" placement="top">
        <el-button type="primary" size="large" circle class="fab-main">
          <Plus :size="28" />
        </el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item @click="$router.push('/accounts')">添加账号</el-dropdown-item>
            <el-dropdown-item @click="$router.push('/tasks')">创建任务</el-dropdown-item>
            <el-dropdown-item divided @click="clearLogs">清空日志</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </div>
</template>

<script setup>
import { onMounted, onUnmounted, reactive, ref, nextTick } from 'vue'
import { 
  Activity, 
  HardDrive, 
  RefreshCw, 
  User,
  Terminal,
  Trash2,
  Play,
  CheckCircle2,
  AlertCircle,
  Loader2,
  X
} from 'lucide-vue-next'
import { getStats, clearLogsAPI } from '../api/dashboard'
import { runTask, dismissTask as runDismissTask } from '../api/task'
import { ElMessage } from 'element-plus'

const stats = reactive({
  scheduled_tasks: 0,
  capacity_used: 0,
  today_completed: 0,
  active_accounts: 0,
  recent_activities: [],
  running_tasks_list: []
})

// 日志与任务监控
const logs = ref([])
const terminalRef = ref(null)
const runningTasks = ref([])
let eventSource = null

const fetchStats = async (isPoll = false) => {
  try {
    const data = await getStats()
    Object.assign(stats, data)
    
    // 同步运行中及最近完成的任务列表
    const apiTasks = data.running_tasks_list || []
      
    // 1. 更新现有任务或添加新任务
    apiTasks.forEach(task => {
      const existing = runningTasks.value.find(t => String(t.id) === String(task.id))
      if (existing) {
        // 仅当 API 返回的进度比前端显示的更“先进”时才覆盖，防止回滚 SSE 的实时跳动
        if (task.percent >= existing.percent) {
          existing.name = task.name
          existing.percent = task.percent
          existing.stage = task.stage
          existing.message = task.message
        }
      } else {
        runningTasks.value.push({
          id: task.id,
          name: task.name,
          percent: task.percent,
          stage: task.stage,
          message: task.message
        })
      }
    })
    
    // 2. 移除 API 不再返回的任务（代表已过期或已隐藏）
    runningTasks.value = runningTasks.value.filter(t => 
      apiTasks.some(at => String(at.id) === String(t.id))
    )
  } catch (error) {
    console.error('获取统计数据失败:', error)
  }
}

let pollTimer = null

onMounted(() => {
  fetchStats()
  initSSE()
  fetchRecentLogs()

  // 恢复 5 秒轮询，负责处理隐式状态变化（如 8s 过期自动消失）
  pollTimer = setInterval(() => {
    fetchStats(true)
  }, 5000)
})

onUnmounted(() => {
  if (eventSource) eventSource.close()
  if (pollTimer) clearInterval(pollTimer)
})

const fetchRecentLogs = async () => {
  try {
    const response = await fetch('/api/dashboard/logs/recent')
    const data = await response.json()
    logs.value = data
    // 不再需要回放进度日志，因为 fetchStats 已经从 DB 拿到了最新状态
    scrollToBottom()
  } catch (error) {
    console.error('获取历史日志失败:', error)
  }
}

const initSSE = () => {
  // 注意：在开发环境下可能需要处理代理路径，这里使用相对路径
  eventSource = new EventSource('/api/dashboard/logs')
  eventSource.onmessage = (event) => {
    const msg = event.data
    if (msg.includes('[PROGRESS:')) {
      handleProgressMessage(msg)
    } else if (msg.includes('[EVENT:')) {
      handleSystemEvent(msg)
    } else {
      logs.value.push(msg)
      if (logs.value.length > 200) logs.value.shift()
      scrollToBottom()
    }
  }
  eventSource.onerror = () => {
    console.error('SSE 连接异常')
  }
}

const handleSystemEvent = (msg) => {
  const match = msg.match(/\[EVENT:(.+)\]/)
  if (!match) return
  try {
    const ev = JSON.parse(match[1])
    if (ev.type === 'stats_update') {
      // 当统计数据发生变化时（如任务完成、账号更新），刷新仪表盘
      fetchStats(true)
    }
  } catch (e) {
    console.error('解析系统事件失败:', e)
  }
}

const handleProgressMessage = (msg) => {
  // 协议格式: [PROGRESS:TaskID:Percent:Stage:Message]
  const match = msg.match(/\[PROGRESS:(.+)\]/)
  if (!match) return
  
  const parts = match[1].split(':')
  if (parts.length < 4) return
  
  const taskId = parts[0]
  const percent = parseInt(parts[1])
  const stage = parts[2]
  const info = parts.slice(3).join(':')
  
  const taskIdx = runningTasks.value.findIndex(t => String(t.id) === String(taskId))
  
  if (taskIdx > -1) {
    // 实时更新任务状态
    const task = runningTasks.value[taskIdx]
    task.percent = percent
    task.stage = stage
    task.message = info
  } else if (stage !== 'Success' && stage !== 'Failed' && stage !== 'Finished') {
    // 如果是新任务且尚未完成，加入列表
    runningTasks.value.push({ 
      id: taskId, 
      name: `任务 #${taskId}`, 
      percent, 
      stage, 
      message: info 
    })
    fetchStats()
  }
}

const dismissTask = async (taskId) => {
  try {
    await runDismissTask(taskId)
    const idx = runningTasks.value.findIndex(t => String(t.id) === String(taskId))
    if (idx > -1) {
      runningTasks.value.splice(idx, 1)
    }
  } catch (err) {
    console.error('忽略任务失败:', err)
  }
}

const getStageTagType = (stage) => {
  const map = {
    'Started': 'info',
    'Parsing': '',
    'Checking': 'warning',
    'Saving': 'primary',
    'Renaming': 'success',
    'Success': 'success',
    'Failed': 'danger'
  }
  return map[stage] || ''
}

const scrollToBottom = () => {
  nextTick(() => {
    if (terminalRef.value) {
      terminalRef.value.scrollTop = terminalRef.value.scrollHeight
    }
  })
}

const clearLogs = async () => {
  try {
    await clearLogsAPI()
    logs.value = []
    ElMessage.success('日志已彻底清空')
  } catch (err) {
    console.error('清空日志失败:', err)
    ElMessage.error('清空日志失败')
  }
}

const handleRetry = async (taskId) => {
  try {
    await runTask(taskId)
    ElMessage.success('已发起重试')
  } catch (err) {
    console.error(err)
  }
}

const getLogClass = (log) => {
  if (log.includes('ERROR')) return 'log-error'
  if (log.includes('WARN')) return 'log-warn'
  if (log.includes('SUCCESS')) return 'log-success'
  return ''
}

const formatSize = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i]
}

const formatRelativeTime = (timeStr) => {
  if (!timeStr || timeStr.startsWith('0001')) return '从未执行'
  const date = new Date(timeStr)
  const now = new Date()
  const diff = Math.floor((now - date) / 1000)

  if (diff < 60) return '刚刚'
  if (diff < 3600) return `${Math.floor(diff / 60)}分钟前`
  if (diff < 86400) return `${Math.floor(diff / 3600)}小时前`
  return `${Math.floor(diff / 86400)}天前`
}

const getStatusType = (status) => {
  const types = {
    'success': 'success',
    'failed': 'danger',
    'running': 'primary'
  }
  return types[status] || 'info'
}

const getStatusText = (status) => {
  const texts = {
    'success': '转存成功',
    'failed': '转存失败',
    'running': '正在执行'
  }
  return texts[status] || '已准备'
}
</script>

<style scoped>
.welcome-section {
  margin-bottom: 32px;
}

.welcome-section h2 {
  margin: 0;
  font-size: 24px;
  color: #1e293b;
}

html.dark .welcome-section h2 {
  color: #f1f5f9;
}

.welcome-section p {
  color: #64748b;
  margin: 8px 0 0 0;
}

.stat-card {
  display: flex;
  align-items: center;
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 16px;
}

.stat-icon.purple { background: rgba(139, 92, 246, 0.1); color: #8b5cf6; }
.stat-icon.blue { background: rgba(59, 130, 246, 0.1); color: #3b82f6; }
.stat-icon.green { background: rgba(16, 185, 129, 0.1); color: #10b981; }
.stat-icon.orange { background: rgba(245, 158, 11, 0.1); color: #f59e0b; }

.stat-label {
  font-size: 14px;
  color: #64748b;
}

.stat-value {
  font-size: 20px;
  font-weight: 700;
  color: #1e293b;
}

html.dark .stat-value { color: #f1f5f9; }

.content-row {
  margin-top: 24px;
}

.placeholder-chart {
  height: 240px;
  display: flex;
  align-items: flex-end;
  justify-content: space-around;
  padding: 20px;
}

.chart-bar {
  width: 40px;
  background: var(--el-color-primary);
  border-radius: 8px 8px 0 0;
  opacity: 0.6;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
}

/* 日志终端样式 */
.terminal-window {
  height: 450px;
  background-color: #0f172a;
  color: #e2e8f0;
  padding: 16px;
  font-family: 'Fira Code', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
  overflow-y: auto;
  border-radius: 0 0 8px 8px;
}

.log-line {
  margin-bottom: 4px;
  display: flex;
  gap: 12px;
}

.log-timestamp {
  color: #64748b;
  flex-shrink: 0;
}

.log-success { color: #10b981; }
.log-error { color: #ef4444; }
.log-warn { color: #f59e0b; }

.terminal-empty {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #475569;
  font-style: italic;
}

/* 任务监控样式 */
.task-progress-card {
  background-color: var(--el-fill-color-lighter);
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 12px;
  border: 1px solid var(--el-border-color-lighter);
}

.task-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.task-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.task-name {
  font-weight: 600;
  font-size: 14px;
}

.task-stage {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.stage-msg {
  font-size: 12px;
  color: #64748b;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.monitor-empty {
  padding: 20px 0;
}

.compact-timeline {
  margin-top: 16px;
  padding-left: 4px;
}

.timeline-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

/* 悬浮操作栏 */
.fab-container {
  position: fixed;
  right: 40px;
  bottom: 40px;
  z-index: 100;
}

.fab-main {
  width: 56px;
  height: 56px;
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.2);
  transition: transform 0.3s;
}

.fab-main:hover {
  transform: scale(1.1);
}

.is-loading {
  animation: rotate 2s linear infinite;
}

@keyframes rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
