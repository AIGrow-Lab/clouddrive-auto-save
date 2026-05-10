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
                @change="() => saveGlobalSchedule(false)"
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
                <el-button type="primary" :loading="savingSchedule" @click="saveGlobalSchedule(true)">
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
                @change="() => saveBarkSettings(false)"
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

              <!-- 高级设置折叠 -->
              <el-collapse class="advanced-collapse">
                <el-collapse-item name="1">
                  <template #title>
                    <span class="collapse-title">高级推送设置</span>
                  </template>

                  <el-form-item label="自定义图标 URL">
                    <el-input v-model="settings.bark_icon" placeholder="https://example.com/icon.png" />
                  </el-form-item>

                  <el-form-item label="自动保存到历史记录">
                    <el-switch v-model="settings.bark_archive" active-value="true" inactive-value="false" />
                  </el-form-item>
                  
                  <el-divider content-position="left">成功通知配置</el-divider>
                  <el-row :gutter="12">
                    <el-col :span="12">
                      <el-form-item label="通知级别">
                        <el-select v-model="settings.bark_success_level" placeholder="选择级别" style="width: 100%">
                          <el-option v-for="l in barkLevels" :key="l.value" :label="l.label" :value="l.value" />
                        </el-select>
                      </el-form-item>
                    </el-col>
                    <el-col :span="12">
                      <el-form-item label="提醒铃声">
                        <el-select v-model="settings.bark_success_sound" placeholder="选择铃声" style="width: 100%">
                          <el-option v-for="s in barkSounds" :key="s.value" :label="s.label" :value="s.value" />
                        </el-select>
                      </el-form-item>
                    </el-col>
                  </el-row>

                  <el-divider content-position="left">失败通知配置</el-divider>
                  <el-row :gutter="12">
                    <el-col :span="12">
                      <el-form-item label="通知级别">
                        <el-select v-model="settings.bark_failure_level" placeholder="选择级别" style="width: 100%">
                          <el-option v-for="l in barkLevels" :key="l.value" :label="l.label" :value="l.value" />
                        </el-select>
                      </el-form-item>
                    </el-col>
                    <el-col :span="12">
                      <el-form-item label="提醒铃声">
                        <el-select v-model="settings.bark_failure_sound" placeholder="选择铃声" style="width: 100%">
                          <el-option v-for="s in barkSounds" :key="s.value" :label="s.label" :value="s.value" />
                        </el-select>
                      </el-form-item>
                    </el-col>
                  </el-row>
                </el-collapse-item>
              </el-collapse>

              <div class="form-actions">
                <el-button type="primary" plain @click="openTestDialog" :loading="testing" style="margin-right: 12px">
                  发送测试消息
                </el-button>
                <el-button type="primary" :loading="savingBark" @click="saveBarkSettings(true)">
                  保存配置
                </el-button>
              </div>
            </el-form>
          </div>
        </el-card>
      </el-col>

      <!-- OpenList 扫描配置 -->
      <el-col :xs="24" :lg="12">
        <el-card class="settings-card">
          <template #header>
            <div class="card-header">
              <div class="header-title">
                <el-icon><Scan /></el-icon>
                <span>OpenList 扫描</span>
              </div>
              <el-switch
                v-model="settings.openlist_enabled"
                active-value="true"
                inactive-value="false"
                @change="() => saveOpenListSettings(false)"
              />
            </div>
          </template>
          <div class="card-content">
            <el-form label-position="top">
              <el-form-item label="API 地址">
                <el-input
                  v-model="settings.openlist_api_url"
                  placeholder="http://127.0.0.1:23541"
                />
              </el-form-item>
              <el-form-item label="API Token">
                <el-input
                  v-model="settings.openlist_api_token"
                  placeholder="openlist-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
                  type="password"
                  show-password
                />
              </el-form-item>

              <div class="form-tip">
                配置 OpenList API 信息后，转存任务完成时将自动触发扫描。也可手动点击按钮触发。
              </div>

              <div class="form-actions">
                <el-button
                  type="primary"
                  plain
                  :loading="openlistScanning"
                  @click="handleOpenListScan"
                  style="margin-right: 12px"
                >
                  手动扫描
                </el-button>
                <el-button type="primary" :loading="savingOpenlist" @click="saveOpenListSettings(true)">
                  保存配置
                </el-button>
              </div>
            </el-form>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Bark 测试对话框 -->
    <el-dialog
      v-model="testDialogVisible"
      title="发送测试推送"
      width="400px"
      append-to-body
      class="custom-dialog"
    >
      <el-form :model="testForm" label-position="top">
        <el-form-item label="推送标题">
          <el-input v-model="testForm.title" placeholder="输入推送标题" />
        </el-form-item>
        <el-form-item label="推送内容">
          <el-input v-model="testForm.body" type="textarea" :rows="3" placeholder="输入推送内容" />
        </el-form-item>
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="通知级别">
              <el-select v-model="testForm.level" style="width: 100%">
                <el-option v-for="l in barkLevels" :key="l.value" :label="l.label" :value="l.value" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="提醒铃声">
              <el-select v-model="testForm.sound" style="width: 100%">
                <el-option v-for="s in barkSounds" :key="s.value" :label="s.label" :value="s.value" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="自定义图标 URL">
          <el-input v-model="testForm.icon" placeholder="https://example.com/icon.png" />
        </el-form-item>
        <el-form-item label="自动保存到历史记录">
          <el-switch v-model="testForm.isArchive" active-value="true" inactive-value="false" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="testDialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="testing" @click="handleTestBark">
            立即发送
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { Calendar, Bell, Info, Scan } from 'lucide-vue-next'
import { getGlobalSettings, updateGlobalSettings, testBark, triggerOpenListScan } from '../api/task'
import { ElMessage, ElMessageBox } from 'element-plus'

const settings = ref({
  global_schedule_enabled: 'false',
  global_schedule_cron: '0 0 0 * * *',
  global_schedule_ui_mode: 'daily',
  bark_enabled: 'false',
  bark_server: 'https://api.day.app',
  bark_device_key: '',
  bark_success_sound: 'birdsong.caf',
  bark_success_level: 'active',
  bark_failure_sound: 'alarm.caf',
  bark_failure_level: 'timeSensitive',
  bark_archive: 'true',
  bark_icon: '',
  openlist_enabled: 'false',
  openlist_api_url: '',
  openlist_api_token: ''
})

const cronMode = ref('daily')
const dailyTime = ref(new Date(new Date().setHours(0, 0, 0, 0)))
const testing = ref(false)
const savingSchedule = ref(false)
const savingBark = ref(false)
const isProcessing = ref(false)
const pageLoading = ref(true)

// OpenList 相关状态
const openlistScanning = ref(false)
const savingOpenlist = ref(false)

// Bark 可选项
const barkLevels = [
  { label: '活跃 (默认)', value: 'active' },
  { label: '时效性 (专注模式可见)', value: 'timeSensitive' },
  { label: '静默', value: 'passive' },
  { label: '告警 (忽略静音)', value: 'critical' }
]

const barkSounds = [
  { label: '清脆鸟鸣 (birdsong.caf)', value: 'birdsong.caf' },
  { label: '警示音 (alarm.caf)', value: 'alarm.caf' },
  { label: '小步舞曲 (minuet.caf)', value: 'minuet.caf' },
  { label: '经典电铃 (bell.caf)', value: 'bell.caf' },
  { label: '默认 (系统)', value: 'default' }
]

// 测试对话框状态
const testDialogVisible = ref(false)
const testForm = ref({
  title: 'UCAS 测试通知',
  body: '这是一条自定义参数的测试推送消息。',
  level: 'active',
  sound: 'birdsong.caf',
  icon: '',
  isArchive: 'true'
})

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
    
    // 铃声值容错处理：如果为空字符串（Bark 默认），映射为 UI 的 'default'
    if (settings.value.bark_success_sound === '') settings.value.bark_success_sound = 'default'
    if (settings.value.bark_failure_sound === '') settings.value.bark_failure_sound = 'default'
    // 确保 archive 有值
    if (settings.value.bark_archive === undefined) settings.value.bark_archive = 'true'

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

const saveGlobalSchedule = async (manual = false) => {
  if (isProcessing.value) return
  isProcessing.value = true
  if (manual) savingSchedule.value = true

  settings.value.global_schedule_ui_mode = cronMode.value
  
  try {
    await updateGlobalSettings(settings.value)
    if (manual) ElMessage.success('全局调度设置已保存')
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '保存失败')
  } finally {
    isProcessing.value = false
    savingSchedule.value = false
  }
}

const saveBarkSettings = async (manual = false) => {
  if (isProcessing.value) return
  isProcessing.value = true
  if (manual) savingBark.value = true
  
  try {
    await updateGlobalSettings(settings.value)
    if (manual) ElMessage.success('Bark 推送设置已保存')
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '保存失败')
  } finally {
    isProcessing.value = false
    savingBark.value = false
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
      bark_device_key: settings.value.bark_device_key,
      ...testForm.value
    })
    ElMessage.success('测试消息已发送，请检查手机')
    testDialogVisible.value = false
  } catch (error) {
    ElMessage.error('测试发送失败: ' + (error.response?.data?.error || error.message))
  } finally {
    testing.value = false
  }
}

const openTestDialog = () => {
  if (!settings.value.bark_device_key) {
    ElMessage.warning('请先填写 Device Key')
    return
  }
  // 默认使用当前配置的值作为测试初始值
  testForm.value.level = settings.value.bark_success_level || 'active'
  testForm.value.sound = settings.value.bark_success_sound || 'birdsong.caf'
  testForm.value.icon = settings.value.bark_icon || ''
  testForm.value.isArchive = settings.value.bark_archive || 'true'
  testDialogVisible.value = true
}

const showCronHelp = () => {
  ElMessageBox.alert(
    'Cron 表达式由 5 或 6 个字段组成：<br/>秒 分 时 日 月 周<br/>例如：<br/><b>0 0 0 * * *</b> (每天凌晨)<br/><b>0 30 15 * * *</b> (每天 15:30:00)',
    'Cron 帮助',
    { dangerouslyUseHTMLString: true }
  )
}

const handleOpenListScan = async () => {
  openlistScanning.value = true
  try {
    await triggerOpenListScan()
    ElMessage.success('OpenList 扫描已触发')
  } catch (error) {
    ElMessage.error('触发扫描失败: ' + (error.response?.data?.error || error.message))
  } finally {
    openlistScanning.value = false
  }
}

const saveOpenListSettings = async (manual = false) => {
  if (isProcessing.value) return
  isProcessing.value = true
  if (manual) savingOpenlist.value = true

  try {
    await updateGlobalSettings(settings.value)
    if (manual) ElMessage.success('OpenList 扫描设置已保存')
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '保存失败')
  } finally {
    isProcessing.value = false
    savingOpenlist.value = false
  }
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
  min-height: 32px;
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

.advanced-collapse {
  margin-top: 20px;
  border: none !important;
}

:deep(.advanced-collapse .el-collapse-item__header) {
  height: 40px;
  border-bottom: none;
  background-color: var(--el-fill-color-light);
  padding: 0 12px;
  border-radius: 8px;
}

:deep(.advanced-collapse .el-collapse-item__wrap) {
  border-bottom: none;
  padding: 12px;
}

.collapse-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--el-text-color-secondary);
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
