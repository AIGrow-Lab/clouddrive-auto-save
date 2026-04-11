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
              <el-select v-model="form.account_id" placeholder="选择账号" style="width: 100%" @change="handleAccountChange">
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
            <el-form-item label="起始日期过滤" tooltip="仅转存此时间之后更新的文件">
              <el-date-picker
                v-model="form.start_date"
                type="datetime"
                placeholder="从哪个日期开始转存"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="保存路径" required>
          <div class="path-input-group">
            <el-tree-select
              :key="form.account_id"
              v-model="form.save_path"
              lazy
              :load="loadFolders"
              :props="{ label: 'label', value: 'path', isLeaf: 'isLeaf' }"
              placeholder="请选择网盘内的保存目录"
              check-strictly
              style="flex: 1"
            />
            <el-button type="primary" link :icon="Plus" @click="handleCreateFolder">新建</el-button>
          </div>
        </el-form-item>

        <el-divider>整理规则与预览</el-divider>
        
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="重命名正则 (Pattern)">
              <el-input v-model="form.pattern" placeholder="匹配文件名的正则表达式" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="替换规则 (Replacement)">
              <el-input v-model="form.replacement" placeholder="支持 {TASKNAME}, {YEAR} 等变量" />
            </el-form-item>
          </el-col>
        </el-row>

        <div class="preview-action">
          <el-button type="success" :icon="RefreshCw" @click="handlePreview" :loading="previewLoading">全量重命名预览</el-button>
        </div>
      </el-form>
      
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitForm">确认并保存</el-button>
      </template>
    </el-dialog>

    <!-- 预览结果对话框 -->
    <el-dialog v-model="previewVisible" title="重命名效果预览" width="800px">
      <el-table :data="previewData" height="400px" border stripe>
        <el-table-column prop="original_name" label="原始文件名" min-width="200" show-overflow-tooltip />
        <el-table-column label="重命名后效果" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <span v-if="row.is_filtered" class="filtered-text">(已过滤)</span>
            <span v-else-if="row.matched" class="matched-text">{{ row.new_name }}</span>
            <span v-else class="unmatched-text">{{ row.new_name }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag size="small" :type="row.is_filtered ? 'info' : (row.matched ? 'success' : 'warning')">
              {{ row.is_filtered ? '跳过' : (row.matched ? '匹配' : '未匹配') }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
      <div class="preview-tips">
        <p>* 列表基于当前分享链接的真实文件解析。</p>
        <p>* 过滤规则：早于 <b>{{ formatTime(form.start_date) }}</b> 的文件将不执行转存。</p>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Plus, Play, Edit, Trash2, RefreshCw } from 'lucide-vue-next'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getTasks, createTask, updateTask, deleteTask, runTask, previewTask } from '../api/task'
import { getAccounts, getFolders, createFolder } from '../api/account'

const taskList = ref([])
const accounts = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const submitting = ref(false)

// 预览相关
const previewVisible = ref(false)
const previewLoading = ref(false)
const previewData = ref([])

// 路径到 ID 的映射，用于辅助新建文件夹
const pathIdMap = ref({ '/': '' })

const form = ref({
  id: null,
  name: '',
  account_id: '',
  share_url: '',
  extract_code: '',
  save_path: '/',
  pattern: '',
  replacement: '',
  start_date: null
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

// 加载网盘目录树
const loadFolders = async (node, resolve) => {
  if (!form.value.account_id) {
    return resolve([])
  }
  
  const parentID = node.level === 0 ? '' : node.data.id
  const parentPath = node.level === 0 ? '/' : node.data.path
  try {
    const folders = await getFolders(form.value.account_id, parentID, parentPath)
    // 更新映射表
    folders.forEach(f => {
      pathIdMap.value[f.path] = f.id
    })
    resolve(folders)
  } catch (err) {
    console.error('加载目录失败:', err)
    resolve([])
  }
}

// 新建文件夹处理
const handleCreateFolder = () => {
  if (!form.value.account_id) return ElMessage.warning('请先选择执行账号')
  
  ElMessageBox.prompt('请输入文件夹名称', '新建文件夹', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    inputPlaceholder: '如：我的资源'
  }).then(async ({ value }) => {
    if (!value) return
    try {
      const parentPath = form.value.save_path || '/'
      const parentID = pathIdMap.value[parentPath]
      
      // 如果我们知道当前路径对应的 ID，则立即调用后端创建
      if (parentID !== undefined) {
        await createFolder(form.value.account_id, parentID, parentPath, value)
        ElMessage.success('文件夹创建成功，请重新展开目录树以查看')
      } else {
        // 如果不知道 ID（例如手动输入的路径），则只在前端追加路径
        ElMessage.success('保存路径已更新，目标文件夹将在执行转存时自动创建')
      }
      
      // 统一在前端更新路径显示
      const newPath = parentPath === '/' ? '/' + value : parentPath + '/' + value
      form.value.save_path = newPath
    } catch (err) {
      console.error(err)
    }
  })
}

// 重命名预览处理
const handlePreview = async () => {
  if (!form.value.account_id || !form.value.share_url) {
    return ElMessage.warning('请先填写执行账号和分享链接')
  }
  
  previewLoading.value = true
  try {
    const data = await previewTask(form.value)
    previewData.value = data
    previewVisible.value = true
  } catch (err) {
    console.error(err)
  } finally {
    previewLoading.value = false
  }
}

// 切换账号处理
const handleAccountChange = () => {
  form.value.save_path = '/'
  pathIdMap.value = { '/': '' }
}

const openAddDialog = () => {
  form.value = { id: null, name: '', account_id: '', share_url: '', extract_code: '', save_path: '/', pattern: '', replacement: '', start_date: null }
  dialogVisible.value = true
}

const handleEdit = (row) => {
  form.value = { 
    id: row.id,
    name: row.name,
    account_id: row.account_id,
    share_url: row.share_url,
    extract_code: row.extract_code,
    save_path: row.save_path,
    pattern: row.pattern,
    replacement: row.replacement,
    start_date: row.start_date
  }
  dialogVisible.value = true
}

const submitForm = async () => {
  if (!form.value.name || !form.value.account_id || !form.value.share_url) {
    return ElMessage.warning('请填写必要的信息')
  }
  
  submitting.value = true
  try {
    if (form.value.id) {
      await updateTask(form.value.id, form.value)
      ElMessage.success('任务更新成功')
    } else {
      await createTask(form.value)
      ElMessage.success('任务保存成功')
    }
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

.path-input-group {
  display: flex;
  gap: 12px;
  align-items: center;
}

.preview-action {
  display: flex;
  justify-content: center;
  margin-top: 12px;
}

.preview-tips {
  margin-top: 16px;
  color: #64748b;
  font-size: 13px;
}

.preview-tips p {
  margin: 4px 0;
}

.filtered-text {
  color: #94a3b8;
  font-style: italic;
}

.matched-text {
  color: #10b981;
  font-weight: 500;
}

.unmatched-text {
  color: #f59e0b;
}
</style>
