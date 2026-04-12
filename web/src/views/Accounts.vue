<template>
  <div class="accounts-container">
    <div class="page-header">
      <div class="title-section">
        <h2>账号管理</h2>
        <p>管理您的移动云盘和夸克网盘账号</p>
      </div>
      <el-button type="primary" :icon="Plus" @click="openAddDialog">添加账号</el-button>
    </div>

    <el-card class="table-card">
      <el-table :data="accountList" v-loading="loading" style="width: 100%">
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
                <span class="capacity-remaining">剩 {{ formatBytes(row.capacity_total - row.capacity_used) }}</span>
              </div>
              <el-progress 
                :percentage="calcPercentage(row.capacity_used, row.capacity_total)" 
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
    </el-card>

    <!-- 添加账号对话框 -->
    <el-dialog v-model="dialogVisible" :title="accountForm.id ? '编辑账号' : '添加新账号'" width="520px" destroy-on-close>
      <el-form :model="accountForm" label-position="top" ref="formRef" class="account-form">
        <el-form-item label="网盘平台" required>
          <el-radio-group v-model="accountForm.platform">
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
            建议优先使用 <b>Authorization</b> (Basic 格式)，它能提供更长久的登录有效期且支持更多高级功能。
          </el-alert>
          <el-form-item label="Authorization">
            <el-input v-model="accountForm.auth_token" type="textarea" :rows="3" placeholder="格式如：Basic pc:138...:xxxxx" />
          </el-form-item>
          <div class="form-or-divider">或</div>
        </template>

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
import { Plus, RefreshCcw, Trash2, Edit, HardDrive, Info } from 'lucide-vue-next'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getAccounts, createAccount, updateAccount, deleteAccount, checkAccount } from '../api/account'

const accountList = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const submitting = ref(false)

const accountForm = ref({
  id: null,
  platform: '139',
  cookie: '',
  auth_token: ''
})

const fetchList = async () => {
  loading.ref = true
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

const submitForm = async () => {
  submitting.value = true
  try {
    if (accountForm.value.id) {
      await updateAccount(accountForm.value.id, accountForm.value)
      ElMessage.success('账号更新成功')
    } else {
      await createAccount(accountForm.value)
      ElMessage.success('账号添加成功')
    }
    dialogVisible.value = false
    fetchList()
  } catch (err) {
    console.error(err)
  } finally {
    submitting.value = false
  }
}

const handleCheck = async (row) => {
  try {
    await checkAccount(row.id)
    ElMessage.success('账号状态正常')
    fetchList()
  } catch (err) {}
}

const handleDelete = (row) => {
  ElMessageBox.confirm('确定要删除该账号吗？关联的任务可能无法执行。', '警告', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    await deleteAccount(row.id)
    ElMessage.success('已删除')
    fetchList()
  })
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
  return p > 100 ? 100 : Number(p.toFixed(1))
}

const getCapacityStatus = (used, total) => {
  const p = calcPercentage(used, total)
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
  background: linear-gradient(90deg, #6366f1 0%, #a855f7 100%);
}

.empty-text {
  color: #94a3b8;
  font-style: italic;
  font-size: 13px;
}

.account-form {
  padding: 0 10px;
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
