<template>
  <div class="accounts-container">
    <div class="page-header">
      <div class="title-section">
        <h2>账号管理</h2>
        <p>管理您的移动云盘和夸克网盘账号</p>
      </div>
      <div class="header-actions">
        <el-radio-group v-model="viewMode" size="default" class="view-toggle" @change="toggleViewMode">
          <el-radio-button label="table">
            <el-icon><List /></el-icon>
          </el-radio-button>
          <el-radio-button label="card">
            <el-icon><LayoutGrid /></el-icon>
          </el-radio-button>
        </el-radio-group>
        <el-button type="primary" :icon="Plus" @click="openAddDialog">添加账号</el-button>
      </div>
    </div>

    <!-- 表格视图 -->
    <el-card v-if="viewMode === 'table'" class="table-card">
      <el-table v-if="accountList.length > 0 || loading" :data="accountList" v-loading="loading" style="width: 100%">
        <el-table-column label="平台" width="140">
          <template #default="{ row }">
            <div class="platform-cell">
              <el-icon :class="row.platform" class="platform-icon">
                <HardDrive />
              </el-icon>
              <span class="platform-name">
                {{ row.platform === 'quark' ? '夸克网盘' : '移动云盘' }}
              </span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="nickname" label="昵称" min-width="120" />
        <el-table-column prop="vip_name" label="会员" width="100">
          <template #default="{ row }">
            <el-tag size="small" type="info" v-if="row.vip_name">{{ row.vip_name }}</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="存储空间" min-width="200">
          <template #default="{ row }">
            <div v-if="row.capacity_total > 0" class="capacity-container">
              <div class="capacity-header">
                <span class="capacity-used">{{ formatBytes(row.capacity_used) }} / {{ formatBytes(row.capacity_total) }}</span>
                <span v-if="row.capacity_total >= row.capacity_used" class="capacity-remaining">
                  剩 {{ formatBytes(row.capacity_total - row.capacity_used) }}
                </span>
                <span v-else class="capacity-remaining is-over">
                  已超额 {{ formatBytes(row.capacity_used - row.capacity_total) }}
                </span>
              </div>
              <el-progress 
                :percentage="Math.min(100, calcPercentage(row.capacity_used, row.capacity_total))" 
                :show-text="false"
                :stroke-width="6"
                :status="getCapacityStatus(row.capacity_used, row.capacity_total)"
                class="gradient-progress"
              />
            </div>
            <span v-else class="empty-text">未获取容量</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" effect="light" round class="status-tag">
              {{ row.status === 1 ? '正常' : '失效' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="last_check" label="最后检查" width="180">
          <template #default="{ row }">
            {{ formatTime(row.last_check) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button-group>
              <el-button link type="primary" :icon="RefreshCcw" @click="handleCheck(row)">校验</el-button>
              <el-button link type="primary" :icon="Edit" @click="handleEdit(row)">编辑</el-button>
              <el-button link type="danger" :icon="Trash2" @click="handleDelete(row)">删除</el-button>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-else description="您还没有绑定任何云盘账号">
        <el-button type="primary" :icon="Plus" @click="openAddDialog">立即绑定账号</el-button>
      </el-empty>
    </el-card>

    <!-- 卡片视图 -->
    <div v-else class="card-view-container" v-loading="loading">
      <template v-if="accountList.length > 0 || loading">
        <el-row :gutter="20">
          <el-col v-for="row in accountList" :key="row.id" :xs="24" :sm="12" :md="8" :lg="6">
            <el-card class="account-card" body-style="padding: 20px">
              <div class="card-header">
                <div class="card-title">
                  <el-icon :class="row.platform" class="platform-icon mini">
                    <HardDrive />
                  </el-icon>
                  <div class="account-info">
                    <div class="nickname">{{ row.nickname }}</div>
                    <div class="platform-tag">{{ row.platform === 'quark' ? '夸克网盘' : '移动云盘' }}</div>
                  </div>
                </div>
                <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small" effect="light" round>
                  {{ row.status === 1 ? '正常' : '失效' }}
                </el-tag>
              </div>

              <div class="card-content">
                <div v-if="row.capacity_total > 0" class="capacity-section">
                  <div class="capacity-header">
                    <span>{{ formatBytes(row.capacity_used) }} / {{ formatBytes(row.capacity_total) }}</span>
                    <span v-if="row.capacity_total >= row.capacity_used" class="remaining">
                      剩 {{ formatBytes(row.capacity_total - row.capacity_used) }}
                    </span>
                    <span v-else class="remaining is-over">
                      已超额 {{ formatBytes(row.capacity_used - row.capacity_total) }}
                    </span>
                  </div>
                  <el-progress 
                    :percentage="Math.min(100, calcPercentage(row.capacity_used, row.capacity_total))" 
                    :show-text="false"
                    :stroke-width="8"
                    :status="getCapacityStatus(row.capacity_used, row.capacity_total)"
                    class="gradient-progress"
                  />
                </div>
                <div v-else class="empty-capacity">
                  <el-icon><Info /></el-icon> 未同步容量信息
                </div>
                
                <div class="meta-info">
                  <div class="meta-item" v-if="row.vip_name">
                    <span class="label">会员状态</span>
                    <el-tag size="small" type="warning" effect="plain">{{ row.vip_name }}</el-tag>
                  </div>
                  <div class="meta-item">
                    <span class="label">最后校验</span>
                    <span class="value">{{ formatTime(row.last_check) }}</span>
                  </div>
                </div>
              </div>

              <div class="card-footer">
                <el-button type="primary" link :icon="RefreshCcw" @click="handleCheck(row)">校验</el-button>
                <el-button type="primary" link :icon="Edit" @click="handleEdit(row)">编辑</el-button>
                <el-button type="danger" link :icon="Trash2" @click="handleDelete(row)">删除</el-button>
              </div>
            </el-card>
          </el-col>
        </el-row>
      </template>
      <el-empty v-else description="您还没有绑定任何云盘账号">
        <el-button type="primary" :icon="Plus" @click="openAddDialog">立即绑定账号</el-button>
      </el-empty>
    </div>

    <!-- 添加账号对话框 -->
    <el-dialog v-model="dialogVisible" :title="accountForm.id ? '编辑账号' : '添加新账号'" width="520px" destroy-on-close>
      <el-form :model="accountForm" label-position="top" ref="formRef" class="account-form">
        <el-form-item label="网盘平台" required>
          <el-radio-group v-model="accountForm.platform" @change="handlePlatformChange">
            <el-radio-button label="139">移动云盘</el-radio-button>
            <el-radio-button label="quark">夸克网盘</el-radio-button>
          </el-radio-group>
        </el-form-item>

        <el-divider border-style="dashed" />

        <!-- 139 特有字段 -->
        <template v-if="accountForm.platform === '139'">
          <el-alert
            title="认证建议"
            type="success"
            :closable="false"
            show-icon
            style="margin-bottom: 18px"
          >
            建议优先使用 <b>Authorization</b> (Basic 格式)，它能提供更长久的有效期且支持更多高级功能。
            <el-link type="primary" :underline="false" href="https://doc.oplist.org/guide/drivers/139#authorization-1" target="_blank" style="margin-left: 8px; font-weight: bold; vertical-align: baseline;">
              查看教程 <el-icon><ExternalLink /></el-icon>
            </el-link>
          </el-alert>
          <el-form-item label="Authorization">
            <el-input v-model="accountForm.auth_token" type="textarea" :rows="3" placeholder="只需要填写 Basic 空格后面开始的内容，不要包含 Basic！" />
          </el-form-item>
          <div class="form-or-divider">或</div>
        </template>

        <!-- 夸克特有提示 -->
        <el-alert
          v-if="accountForm.platform === 'quark'"
          title="使用须知"
          type="info"
          :closable="false"
          show-icon
          style="margin-bottom: 18px"
        >
          夸克网盘仅支持 Cookie 认证。
          <el-link type="primary" :underline="false" href="https://doc.oplist.org/guide/drivers/quark#cookie-1" target="_blank" style="margin-left: 8px; font-weight: bold; vertical-align: baseline;">
            如何获取？ <el-icon><ExternalLink /></el-icon>
          </el-link>
        </el-alert>

        <el-form-item label="Cookie 全量字符串">
          <el-input v-model="accountForm.cookie" type="textarea" :rows="4" placeholder="通过浏览器 F12 网络选项卡获取" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitForm">{{ accountForm.id ? '确认更新' : '确认添加' }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Plus, RefreshCcw, Trash2, Edit, HardDrive, Info, LayoutGrid, List, ExternalLink } from 'lucide-vue-next'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getAccounts, createAccount, updateAccount, deleteAccount, checkAccount } from '../api/account'

const accountList = ref([])
const loading = ref(true)
const dialogVisible = ref(false)
const submitting = ref(false)
const viewMode = ref(localStorage.getItem('accountViewMode') || 'table')

const toggleViewMode = (mode) => {
  viewMode.value = mode
  localStorage.setItem('accountViewMode', mode)
}

const accountForm = ref({
  id: null,
  platform: '139',
  cookie: '',
  auth_token: ''
})

const fetchList = async () => {
  loading.value = true
  try {
    const res = await getAccounts()
    accountList.value = res
  } catch (err) {
    console.error(err)
  } finally {
    loading.value = false
  }
}

const openAddDialog = () => {
  accountForm.value = { id: null, platform: '139', cookie: '', auth_token: '' }
  dialogVisible.value = true
}

const handleEdit = (row) => {
  accountForm.value = {
    id: row.id,
    platform: row.platform,
    cookie: row.cookie,
    auth_token: row.auth_token
  }
  dialogVisible.value = true
}

const handlePlatformChange = () => {
  accountForm.value.cookie = ''
  accountForm.value.auth_token = ''
}

const submitForm = async () => {
  submitting.value = true
  try {
    if (accountForm.value.id) {
      const res = await updateAccount(accountForm.value.id, accountForm.value)
      if (res.status === 1) {
        ElMessage.success('账号更新并校验成功')
      } else {
        ElMessage.warning('账号已更新，但连通性校验失败，请检查认证信息')
      }
    } else {
      const res = await createAccount(accountForm.value)
      if (res.status === 1) {
        ElMessage.success('账号添加并校验成功')
      } else {
        ElMessage.warning('账号已添加，但连通性校验失败，请检查认证信息')
      }
    }
    dialogVisible.value = false
  } catch (err) {
    console.error(err)
  } finally {
    submitting.value = false
    fetchList()
  }
}

const handleCheck = async (row) => {
  try {
    const updatedAccount = await checkAccount(row.id)
    Object.assign(row, updatedAccount)
    ElMessage.success('账号状态正常')
  } catch (err) {
    // 错误已由拦截器展示，这里从错误响应中尝试提取后端更新后的账号状态 (包含最新的校验时间)
    if (err.response && err.response.data && err.response.data.account) {
      Object.assign(row, err.response.data.account)
    }
  }
}

const handleDelete = (row) => {
  ElMessageBox.confirm('确定要删除该账号吗？只有在没有关联任务的情况下才能成功删除。', '删除账号', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await deleteAccount(row.id)
      ElMessage.success('已删除')
      fetchList()
    } catch (err) {
      // API 请求失败（例如存在关联任务），拦截器已统一抛出提示，此处静默捕获即可。
    }
  }).catch(() => {})
}

const formatTime = (timeStr) => {
  if (!timeStr || timeStr.startsWith('0001')) return '从未检查'
  return new Date(timeStr).toLocaleString()
}

const formatBytes = (bytes) => {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const calcPercentage = (used, total) => {
  if (!total) return 0
  const p = (used / total) * 100
  // 返回真实百分比用于逻辑判断，但渲染组件时会限制在 100
  return Number(p.toFixed(1))
}

const getCapacityStatus = (used, total) => {
  const p = (used / total) * 100
  if (p >= 90) return 'exception'
  if (p >= 70) return 'warning'
  return 'success'
}

onMounted(() => {
  fetchList()
})
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.title-section h2 {
  margin: 0;
  font-size: 24px;
  color: #1e293b;
}

html.dark .title-section h2 {
  color: #f1f5f9;
}

.title-section p {
  color: #64748b;
  margin: 4px 0 0 0;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}

.view-toggle :deep(.el-radio-button__inner) {
  padding: 8px 12px;
  display: flex;
  align-items: center;
}

.account-card {
  border-radius: 16px;
  margin-bottom: 20px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border: 1px solid var(--el-border-color-lighter);
}

.account-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 24px -4px rgba(0, 0, 0, 0.1);
  border-color: var(--el-color-primary-light-5);
}

.account-card .card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
}

.card-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.platform-icon.mini {
  padding: 8px;
  font-size: 16px;
}

.account-info .nickname {
  font-weight: 600;
  font-size: 16px;
  color: var(--el-text-color-primary);
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.account-info .platform-tag {
  font-size: 12px;
  color: #94a3b8;
  margin-top: 2px;
}

.capacity-section {
  margin-bottom: 20px;
}

.capacity-section .capacity-header {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  margin-bottom: 8px;
  color: #64748b;
}

.capacity-section .remaining {
  color: #10b981;
  font-weight: 600;
}

.empty-capacity {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  background-color: var(--el-fill-color-lighter);
  border-radius: 8px;
  color: #94a3b8;
  font-size: 13px;
  margin-bottom: 20px;
}

.meta-info {
  background-color: var(--el-fill-color-blank);
  border-radius: 12px;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  border: 1px dashed var(--el-border-color-lighter);
}

.meta-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
}

.meta-item .label {
  color: #94a3b8;
}

.meta-item .value {
  color: var(--el-text-color-regular);
}

.card-footer {
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid var(--el-border-color-extra-light);
  display: flex;
  justify-content: space-around;
}

.table-card {
  border-radius: 12px;
}

.platform-cell {
  display: flex;
  align-items: center;
  gap: 10px;
}

.platform-icon {
  font-size: 18px;
  padding: 6px;
  border-radius: 8px;
}

.platform-icon.quark {
  background-color: rgba(103, 194, 58, 0.1);
  color: #67c23a;
}

.platform-icon.\31 39 {
  background-color: rgba(230, 162, 60, 0.1);
  color: #e6a23c;
}

.capacity-container {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.capacity-header {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
}

.capacity-used {
  color: #64748b;
}

.capacity-remaining {
  color: #10b981;
  font-weight: 500;
}

.status-tag {
  padding: 0 12px;
  font-weight: 500;
}

.gradient-progress :deep(.el-progress-bar__inner) {
  transition: all 0.3s;
}

/* 正常状态 (Green) */
.gradient-progress.is-success :deep(.el-progress-bar__inner) {
  background: linear-gradient(90deg, #10b981 0%, #34d399 100%);
}

/* 预警状态 (Yellow) */
.gradient-progress.is-warning :deep(.el-progress-bar__inner) {
  background: linear-gradient(90deg, #f59e0b 0%, #fbbf24 100%);
}

/* 危险/超额状态 (Red) */
.gradient-progress.is-exception :deep(.el-progress-bar__inner) {
  background: linear-gradient(90deg, #ef4444 0%, #f87171 100%);
}

.capacity-remaining.is-over, .remaining.is-over {
  color: #ef4444 !important;
}

.empty-text {
  color: #94a3b8;
  font-style: italic;
  font-size: 13px;
}

.account-form {
  padding: 0 10px;
}

.form-label-with-link {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.help-link {
  font-size: 12px;
  font-weight: normal;
}

.help-link .el-icon {
  margin-right: 4px;
}

.form-or-divider {
  text-align: center;
  margin: 15px 0;
  position: relative;
  color: #94a3b8;
  font-size: 12px;
}

.form-or-divider::before,
.form-or-divider::after {
  content: "";
  position: absolute;
  top: 50%;
  width: 40%;
  height: 1px;
  background-color: var(--el-border-color-lighter);
}

.form-or-divider::before {
  left: 0;
}

.form-or-divider::after {
  right: 0;
}
</style>
