<template>
  <div class="tasks-container">
    <div class="page-header">
      <div class="title-section">
        <h2>任务管理</h2>
        <p>监控并自动转存 139 和 Quark 的分享资源</p>
      </div>
      <el-button type="primary" :icon="Plus" @click="openAddDialog">创建任务</el-button>
    </div>

    <el-card class="table-card">
      <el-table :data="taskList" v-loading="loading" style="width: 100%">
        <el-table-column label="任务名称" min-width="180">
          <template #default="{ row }">
            <div class="task-name-cell">
              <span class="name">{{ row.name }}</span>
              <div class="account-tag">
                <el-tag size="small" :type="row.account.platform === 'quark' ? 'success' : 'warning'">
                  {{ row.account.nickname || row.account.platform }}
                </el-tag>
              </div>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column prop="save_path" label="保存路径" min-width="150" show-overflow-tooltip />
        
        <el-table-column label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              <div class="status-inner">
                <el-icon v-if="row.status === 'running'" class="is-loading"><RefreshCw /></el-icon>
                {{ row.status.toUpperCase() }}
              </div>
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="最后运行" width="180">
          <template #default="{ row }">
            {{ formatTime(row.last_run) }}
          </template>
        </el-table-column>

        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button-group>
              <el-button link type="primary" :icon="Play" :disabled="row.status === 'running'" @click="handleRun(row)">运行</el-button>
              <el-button link type="primary" :icon="Edit" @click="handleEdit(row)">编辑</el-button>
              <el-button link type="danger" :icon="Trash2" @click="handleDelete(row)">删除</el-button>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 创建/编辑任务对话框 -->
    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑任务' : '创建新任务'" width="600px">
      <el-form :model="form" label-position="top" ref="formRef">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="任务名称" required>
              <el-input v-model="form.name" placeholder="给任务起个名字" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="执行账号" required>
              <el-select v-model="form.account_id" placeholder="选择账号" style="width: 100%">
                <el-option
                  v-for="acc in accounts"
                  :key="acc.id"
                  :label="`${acc.nickname} (${acc.platform})`"
                  :value="acc.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="分享链接" required>
          <el-input v-model="form.share_url" placeholder="请输入 139 或 Quark 分享链接" />
        </el-form-item>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="提取码">
              <el-input v-model="form.extract_code" placeholder="如果有提取码请填写" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="保存路径" required>
              <el-input v-model="form.save_path" placeholder="云盘内的保存目录，如 /电影/2024" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-divider>整理规则 (可选)</el-divider>
        
        <el-form-item label="重命名正则 (Pattern)">
          <el-input v-model="form.pattern" placeholder="匹配文件名的正则表达式" />
        </el-form-item>
        
        <el-form-item label="替换规则 (Replacement / Magic Variables)">
          <el-input v-model="form.replacement" placeholder="支持 {TASKNAME}, {YEAR}, {DATE} 等变量" />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitForm">确认并保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Plus, Play, Edit, Trash2, RefreshCw } from 'lucide-vue-next'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getTasks, createTask, deleteTask, runTask } from '../api/task'
import { getAccounts } from '../api/account'

const taskList = ref([])
const accounts = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const submitting = ref(false)

const form = ref({
  id: null,
  name: '',
  account_id: '',
  share_url: '',
  extract_code: '',
  save_path: '/',
  pattern: '',
  replacement: ''
})

const fetchList = async () => {
  loading.value = true
  try {
    const [taskData, accountData] = await Promise.all([getTasks(), getAccounts()])
    taskList.value = taskData
    accounts.value = accountData
  } catch (err) {
    console.error(err)
  } finally {
    loading.value = false
  }
}

const openAddDialog = () => {
  form.value = { id: null, name: '', account_id: '', share_url: '', extract_code: '', save_path: '/', pattern: '', replacement: '' }
  dialogVisible.value = true
}

const submitForm = async () => {
  if (!form.value.name || !form.value.account_id || !form.value.share_url) {
    return ElMessage.warning('请填写必要的信息')
  }
  
  submitting.value = true
  try {
    await createTask(form.value)
    ElMessage.success('任务保存成功')
    dialogVisible.value = false
    fetchList()
  } catch (err) {
    console.error(err)
  } finally {
    submitting.value = false
  }
}

const handleRun = async (row) => {
  try {
    await runTask(row.id)
    ElMessage.success('任务已提交执行队列')
    fetchList()
  } catch (err) {}
}

const handleDelete = (row) => {
  ElMessageBox.confirm('确定要删除此转存任务吗？', '确认', {
    type: 'warning'
  }).then(async () => {
    await deleteTask(row.id)
    ElMessage.success('任务已删除')
    fetchList()
  })
}

const getStatusType = (status) => {
  const map = {
    pending: 'info',
    running: 'primary',
    success: 'success',
    failed: 'danger'
  }
  return map[status] || 'info'
}

const formatTime = (timeStr) => {
  if (!timeStr || timeStr.startsWith('0001')) return '从不'
  return new Date(timeStr).toLocaleString()
}

onMounted(fetchList)
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
}

.title-section p {
  color: #64748b;
  margin: 4px 0 0 0;
}

.task-name-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.task-name-cell .name {
  font-weight: 600;
  color: #1e293b;
}

html.dark .task-name-cell .name {
  color: #f1f5f9;
}

.status-inner {
  display: flex;
  align-items: center;
  gap: 6px;
}

.is-loading {
  animation: rotate 2s linear infinite;
}

@keyframes rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
