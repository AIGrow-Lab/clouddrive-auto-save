<template>
  <div class="settings-container">
    <div class="welcome-section">
      <h2>系统设置 ⚙️</h2>
      <p>管理全局调度任务与消息推送配置</p>
    </div>

    <el-row :gutter="24">
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
                @change="saveSettings"
              />
            </div>
          </template>
          <div class="card-content">
            <el-form label-position="top">
              <el-form-item label="全局 Cron 表达式">
                <el-input
                  v-model="settings.global_schedule_cron"
                  placeholder="e.g. 0 0 * * *"
                  @blur="saveSettings"
                  @keyup.enter="saveSettings"
                >
                  <template #append>
                    <el-tooltip content="Cron 帮助" placement="top">
                      <el-button :icon="Info" @click="showCronHelp" />
                    </el-tooltip>
                  </template>
                </el-input>
                <div class="form-tip">
                  设置全局默认运行时间，个别任务可单独重写此设置。
                </div>
              </el-form-item>
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
              </div>
              <el-switch
                v-model="settings.bark_enabled"
                active-value="true"
                inactive-value="false"
                @change="saveSettings"
              />
            </div>
          </template>
          <div class="card-content">
            <el-form label-position="top">
              <el-form-item label="Bark 服务器地址">
                <el-input
                  v-model="settings.bark_server"
                  placeholder="https://api.day.app"
                  @blur="saveSettings"
                />
              </el-form-item>
              <el-form-item label="Device Key">
                <el-input
                  v-model="settings.bark_device_key"
                  placeholder="您的 Bark 设备 Key"
                  @blur="saveSettings"
                  type="password"
                  show-password
                />
              </el-form-item>
              <div class="form-actions">
                <el-button type="primary" plain @click="handleTestBark" :loading="testing">
                  发送测试消息
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
import { ref, onMounted } from 'vue'
import { Calendar, Bell, Info } from 'lucide-vue-next'
import { getGlobalSettings, updateGlobalSettings, testBark } from '../api/task'
import { ElMessage, ElMessageBox } from 'element-plus'

const settings = ref({
  global_schedule_enabled: 'false',
  global_schedule_cron: '0 0 * * *',
  bark_enabled: 'false',
  bark_server: 'https://api.day.app',
  bark_device_key: ''
})

const testing = ref(false)

const fetchSettings = async () => {
  try {
    const data = await getGlobalSettings()
    // 合并默认值
    settings.value = { ...settings.value, ...data }
  } catch (error) {
    ElMessage.error('加载设置失败')
  }
}

const saveSettings = async () => {
  try {
    await updateGlobalSettings(settings.value)
    ElMessage.success('设置已保存')
  } catch (error) {
    ElMessage.error('保存设置失败')
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
    'Cron 表达式由 5 或 6 个字段组成：<br/>秒 分 时 日 月 周<br/>例如：<br/><b>0 0 * * *</b> (每天凌晨)<br/><b>0 30 15 * * *</b> (每天 15:30:00)',
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

.form-tip {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 8px;
}

.form-actions {
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
}
</style>
