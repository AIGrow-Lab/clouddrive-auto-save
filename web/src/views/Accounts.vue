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
        <el-table-column label="平台" width="120">
          <template #default="{ row }">
            <el-tag :type="row.platform === 'quark' ? 'success' : 'warning'" effect="dark">
              {{ row.platform === 'quark' ? '夸克网盘' : '移动云盘' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="nickname" label="昵称" min-width="150" />
        <el-table-column prop="account_name" label="账号/手机号" width="150" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-badge :is-dot="true" :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? '正常' : '失效' }}
            </el-badge>
          </template>
        </el-table-column>
        <el-table-column prop="last_check" label="最后检查" width="180">
          <template #default="{ row }">
            {{ formatTime(row.last_check) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button-group>
              <el-button link type="primary" :icon="RefreshCcw" @click="handleCheck(row)">校验</el-button>
              <el-button link type="danger" :icon="Trash2" @click="handleDelete(row)">删除</el-button>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 添加账号对话框 -->
    <el-dialog v-model="dialogVisible" title="添加新账号" width="500px" destroy-on-close>
      <el-form :model="accountForm" label-position="top" ref="formRef">
        <el-form-item label="网盘平台" required>
          <el-radio-group v-model="accountForm.platform">
            <el-radio-button label="139">移动云盘</el-radio-button>
            <el-radio-button label="quark">夸克网盘</el-radio-button>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="备注/手机号" required>
          <el-input v-model="accountForm.account_name" placeholder="仅用于识别，如手机号或备注名" />
        </el-form-item>

        <!-- 139 特有字段 -->
        <template v-if="accountForm.platform === '139'">
          <el-form-item label="Authorization (推荐)" help="抓包获取的 Basic xxxx 字符串">
            <el-input v-model="accountForm.auth_token" type="textarea" :rows="3" placeholder="请输入 Authorization" />
          </el-form-item>
          <el-divider>或者使用 Cookie</el-divider>
        </template>

        <el-form-item label="Cookie">
          <el-input v-model="accountForm.cookie" type="textarea" :rows="4" placeholder="请输入 Cookie 字符串" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitForm">确认添加</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Plus, RefreshCcw, Trash2 } from 'lucide-vue-next'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getAccounts, createAccount, deleteAccount, checkAccount } from '../api/account'

const accountList = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const submitting = ref(false)

const accountForm = ref({
  platform: '139',
  account_name: '',
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
  accountForm.value = { platform: '139', account_name: '', cookie: '', auth_token: '' }
  dialogVisible.value = true
}

const submitForm = async () => {
  if (!accountForm.value.account_name) return ElMessage.warning('请输入账号备注')
  
  submitting.value = true
  try {
    await createAccount(accountForm.value)
    ElMessage.success('账号添加成功')
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
</style>
