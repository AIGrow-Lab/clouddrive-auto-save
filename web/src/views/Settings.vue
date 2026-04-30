<template>
  <div class="settings-container" v-loading="pageLoading">
    <div class="welcome-section">
      <h2>系统设置 ⚙️</h2>
      <p>管理全局调度任务与消息推送配置</p>
    </div>

    <el-row v-if="!pageLoading" :gutter="24">
      <!-- 全局调度设置 -->
      <el-col :xs="24" :lg="12">
        <el-card class="settings-card">
          <template #header>
            <div class="card-header">
              <div class="header-title">
                <el-icon><Calendar /></el-icon>
                <span>全局定时任务</span>
              </div>
              <el-switch
                v-model="settings.global_schedule_enabled"
                active-value="true"
                inactive-value="false"
                @change="() => saveSettings(false)"
              />
            </div>
          </template>
          <div class="card-content">
            <!-- 当前规则摘要 -->
            <div class="schedule-summary" :class="{ 'is-disabled': settings.global_schedule_enabled === 'false' }">
              <el-icon><Info /></el-icon>
              <span class="summary-text">
                当前设定：{{ settings.global_schedule_enabled === 'true' ? getCronDescription(settings.global_schedule_cron) : '未开启全局调度' }}
              </span>
            </div>

            <el-form label-position="top">
              <el-form-item label="配置模式">
                <el-radio-group v-model="cronMode" size="small">
                  <el-radio-button label="daily">简易定时</el-radio-button>
                  <el-radio-button label="advanced">高级 Cron</el-radio-button>
                </el-radio-group>
              </el-form-item>

              <el-form-item v-if="cronMode === 'daily'" label="每天运行时间">
                <div class="daily-picker-container">
                  <el-time-picker
                    v-model="dailyTime"
                    format="HH:mm"
                    placeholder="选择时间"
                    @change="handleTimeChange"
                    style="width: 100%"
                  />
                  <div class="presets-container">
                    <el-button-group size="small">
                      <el-button @click="setPreset('00:00')">凌晨</el-button>
                      <el-button @click="setPreset('08:00')">早晨</el-button>
                      <el-button @click="setPreset('12:00')">中午</el-button>
                    </el-button-group>
                  </div>
                </div>
              </el-form-item>

              <el-form-item v-else label="全局 Cron 表达式">
                <el-input
                  v-model="settings.global_schedule_cron"
                  placeholder="e.g. 0 0 0 * * *"
                >
                  <template #append>
                    <el-tooltip content="Cron 帮助" placement="top">
                      <el-button :icon="Info" @click="showCronHelp" />
                    </el-tooltip>
                  </template>
                </el-input>
              </el-form-item>
              
              <div class="form-tip">
                设置全局默认运行时间，个别任务可单独重写此设置。
              </div>

              <div class="form-actions">
                <el-button type="primary" :loading="saving" @click="saveSettings(true)">
                  保存配置
                </el-button>
              </div>
            </el-form>
          </div>
        </el-card>
      </el-col>

      <!-- Bark 推送设置 -->
      <el-col :xs="24" :lg="12">
        <el-card class="settings-card">
          <template #header>
            <div class="card-header">
              <div class="header-title">
                <el-icon><Bell /></el-icon>
                <span>Bark 消息推送</span>
                <el-link
                  type="primary"
                  href="https://bark.day.app/"
                  target="_blank"
                  :underline="false"
                  style="margin-left: 8px; font-size: 13px;"
                >
                  查看教程
                </el-link>
              </div>
              <el-switch
                v-model="settings.bark_enabled"
                active-value="true"
                inactive-value="false"
                @change="() => saveSettings(false)"
              />
            </div>
          </template>
          <div class="card-content">
            <el-form label-position="top">
              <el-form-item label="Bark 服务器地址">
                <el-input
                  v-model="settings.bark_server"
                  placeholder="https://api.day.app"
                />
              </el-form-item>
              <el-form-item label="Device Key">
                <el-input
                  v-model="settings.bark_device_key"
                  placeholder="您的 Bark 设备 Key"
                  type="password"
                  show-password
                />
              </el-form-item>
              <div class="form-actions">
                <el-button type="primary" plain @click="handleTestBark" :loading="testing" style="margin-right: 12px">
                  发送测试消息
                </el-button>
                <el-button type="primary" :loading="saving" @click="saveSettings(true)">
                  保存配置
                </el-button>
              </div>
            </el-form>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { Calendar, Bell, Info } from 'lucide-vue-next'
import { getGlobalSettings, updateGlobalSettings, testBark } from '../api/task'
import { ElMessage, ElMessageBox } from 'element-plus'

const settings = ref({
  global_schedule_enabled: 'false',
  global_schedule_cron: '0 0 0 * * *',
  global_schedule_ui_mode: 'daily',
  bark_enabled: 'false',
  bark_server: 'https://api.day.app',
  bark_device_key: ''
})

const cronMode = ref('daily')
const dailyTime = ref(new Date(new Date().setHours(0, 0, 0, 0)))
const testing = ref(false)
const saving = ref(false)
const isProcessing = ref(false)
const pageLoading = ref(true)

// 简单的 Cron 转中文描述
const getCronDescription = (cron) => {
  if (!cron) return '未设置'
  const parts = cron.split(' ')
  if (parts.length < 5) return '格式不完整'
  
  // 补齐到 6 位 (秒 分 时 日 月 周)
  const p = parts.length === 5 ? ['0', ...parts] : parts
  
  if (p[3] === '*' && p[4] === '*' && p[5] === '*') {
    return `每天 ${p[2].padStart(2, '0')}:${p[1].padStart(2, '0')}:${p[0].padStart(2, '0')}`
  }
  return cron // 复杂格式直接显示原始字符串
}

const fetchSettings = async () => {
  pageLoading.value = true
  try {
    const data = await getGlobalSettings()
    // 合并默认值
    settings.value = { ...settings.value, ...data }
    
    // 优先使用持久化的 UI 模式
    if (settings.value.global_schedule_ui_mode) {
      cronMode.value = settings.value.global_schedule_ui_mode
    } else {
      // 降级：通过 Cron 自动推断模式
      const cron = settings.value.global_schedule_cron || '0 0 0 * * *'
      const parts = cron.split(' ')
      if (parts.length >= 3 && parts[0] === '0' && parts[3] === '*' && parts[4] === '*' && parts[5] === '*') {
        cronMode.value = 'daily'
      } else {
        cronMode.value = 'advanced'
      }
    }

    // 初始化时间选择器
    const cron = settings.value.global_schedule_cron || '0 0 0 * * *'
    const parts = cron.split(' ')
    const p = parts.length === 5 ? ['0', ...parts] : parts
    if (p.length >= 3) {
      const d = new Date()
      d.setHours(parseInt(p[2]), parseInt(p[1]), parseInt(p[0]), 0)
      dailyTime.value = d
    }
  } catch (error) {
    ElMessage.error({ message: '加载设置失败', grouping: true })
  } finally {
    pageLoading.value = false
  }
}

const handleTimeChange = (val) => {
  if (!val) return
  const h = val.getHours()
  const m = val.getMinutes()
  const s = val.getSeconds()
  settings.value.global_schedule_cron = `${s} ${m} ${h} * * *`
}

watch(cronMode, (newMode) => {
  if (newMode === 'daily') {
    handleTimeChange(dailyTime.value)
  }
})

const setPreset = (timeStr) => {
  const [h, m] = timeStr.split(':').map(Number)
  const d = new Date()
  d.setHours(h, m, 0, 0)
  dailyTime.value = d
  handleTimeChange(d)
}

const saveSettings = async (manual = false) => {
  if (isProcessing.value) return
  isProcessing.value = true
  if (manual === true) saving.value = true
  
  // 同步当前的 UI 模式到设置对象中以进行持久化
  settings.value.global_schedule_ui_mode = cronMode.value

  try {
    await updateGlobalSettings(settings.value)
    if (manual === true) {
      ElMessage.success('设置已保存')
    }
  } catch (error) {
    const msg = error.response?.data?.error || '保存设置失败'
    ElMessage.error({ message: msg, grouping: true })
  } finally {
    isProcessing.value = false
    saving.value = false
  }
}

const handleTestBark = async () => {
  if (!settings.value.bark_device_key) {
    ElMessage.warning('请先填写 Device Key')
    return
  }
  testing.value = true
  try {
    await testBark({
      bark_server: settings.value.bark_server,
      bark_device_key: settings.value.bark_device_key
    })
    ElMessage.success('测试消息已发送，请检查手机')
  } catch (error) {
    ElMessage.error('测试发送失败: ' + (error.response?.data?.error || error.message))
  } finally {
    testing.value = false
  }
}

const showCronHelp = () => {
  ElMessageBox.alert(
    'Cron 表达式由 5 或 6 个字段组成：<br/>秒 分 时 日 月 周<br/>例如：<br/><b>0 0 0 * * *</b> (每天凌晨)<br/><b>0 30 15 * * *</b> (每天 15:30:00)',
    'Cron 帮助',
    { dangerouslyUseHTMLString: true }
  )
}

onMounted(() => {
  fetchSettings()
})
</script>

<style scoped>
.settings-container {
  padding: 24px;
}

.welcome-section {
  margin-bottom: 32px;
}

.welcome-section h2 {
  font-size: 24px;
  font-weight: 600;
  margin-bottom: 8px;
  color: var(--el-text-color-primary);
}

.welcome-section p {
  color: var(--el-text-color-secondary);
}

.settings-card {
  margin-bottom: 24px;
  border-radius: 12px;
  border: none;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
  transition: transform 0.3s ease;
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
  font-size: 16px;
}

.schedule-summary {
  display: flex;
  align-items: center;
  gap: 8px;
  background-color: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
  padding: 10px 14px;
  border-radius: 8px;
  margin-bottom: 20px;
  font-size: 13px;
  font-weight: 500;
  border: 1px solid var(--el-color-primary-light-8);
}

.schedule-summary.is-disabled {
  background-color: var(--el-fill-color-lighter);
  color: var(--el-text-color-secondary);
  border-color: var(--el-border-color-lighter);
}

.summary-text {
  flex: 1;
}

.form-tip {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 8px;
}

.daily-picker-container {
  width: 100%;
}

.presets-container {
  margin-top: 12px;
}

.form-actions {
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
}
</style>
