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
                <el-icon v-if="row.status === 'running'" class="icon-spin"><RefreshCw /></el-icon>
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
          <el-col :span="24">
            <el-form-item label="提取码">
              <el-input v-model="form.extract_code" placeholder="如果有提取码请填写" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="起始转存点 (可选)">
          <div class="path-input-group">
            <el-input 
              v-model="selectedStartFileName" 
              placeholder="从该文件开始向前转存 (为空则转存全量)" 
              readonly
              class="save-path-input"
            >
              <template #append>
                <el-button @click="openStartFileDialog" :loading="parsingShare">选择文件</el-button>
              </template>
            </el-input>
          </div>
          <div class="start-id-tip" v-if="form.start_file_id">
            <el-icon><Info /></el-icon>
            <span>当前已锁定 ID: <code class="id-code" :title="form.start_file_id">{{ form.start_file_id }}</code></span>
          </div>
        </el-form-item>

        <el-form-item label="保存路径" required>
          <div class="path-input-group">
            <el-input 
              v-model="form.save_path" 
              placeholder="可手动输入或点击右侧选择" 
              class="save-path-input"
            >
              <template #prepend v-if="selectedAccountPlatform">
                [{{ selectedAccountPlatform }}]
              </template>
              <template #append>
                <el-button @click="openFolderDialog">选择目录</el-button>
              </template>
            </el-input>
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

    <!-- 目录选择独立弹窗 -->
    <el-dialog 
      v-model="folderDialogVisible" 
      title="选择保存目录" 
      width="600px" 
      class="folder-dialog"
      append-to-body
      destroy-on-close
    >
      <div class="folder-tree-container" v-loading="loadingFolders">
        <el-tree
          ref="folderTreeRef"
          lazy
          :load="loadFolders"
          :props="{ label: 'label', isLeaf: 'isLeaf' }"
          node-key="path"
          highlight-current
          @current-change="handleTreeCurrentChange"
          :expand-on-click-node="false"
        >
          <template #default="{ node, data }">
            <span class="custom-tree-node">
              <el-icon><Folder /></el-icon>
              <span>{{ node.label }}</span>
            </span>
          </template>
        </el-tree>
      </div>
      <template #footer>
        <div class="folder-dialog-footer">
          <div class="create-folder-action">
            <el-input v-model="newFolderName" placeholder="新文件夹名称" size="default">
              <template #append>
                <el-button @click="handleInlineCreateFolder" :loading="creatingFolder">新建</el-button>
              </template>
            </el-input>
          </div>
          <div class="dialog-actions">
            <el-button @click="folderDialogVisible = false">取消</el-button>
            <el-button type="primary" @click="confirmFolderSelection">确认路径</el-button>
          </div>
        </div>
      </template>
    </el-dialog>

    <!-- 选择起始文件独立弹窗 -->
    <el-dialog 
      v-model="startFileDialogVisible" 
      title="选择起始转存文件" 
      width="700px"
      append-to-body
      destroy-on-close
    >
      <div class="share-files-dialog-content">
        <el-alert title="逻辑说明" type="info" :closable="false" show-icon style="margin-bottom: 15px">
          系统将从选中的文件开始，按更新时间向前转存所有更新的文件（含所选文件本身）。
        </el-alert>
        
        <el-table 
          :data="shareFiles" 
          max-height="400" 
          size="default" 
          border 
          stripe 
          v-loading="parsingShare"
          highlight-current-row
          @current-change="handleStartFileTableChange"
        >
          <el-table-column width="55" align="center">
            <template #default="{ row }">
              <el-radio v-model="tempStartFileId" :label="row.id">&nbsp;</el-radio>
            </template>
          </el-table-column>
          <el-table-column label="文件名" show-overflow-tooltip min-width="200">
            <template #default="{ row }">
              <div style="display: flex; align-items: center; gap: 8px">
                <el-icon size="18">
                  <Folder v-if="row.is_folder" color="#eab308" />
                  <File v-else color="#64748b" />
                </el-icon>
                <span :style="{ fontWeight: row.is_folder ? '600' : 'normal' }">{{ row.name }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="类型" width="90" align="center">
            <template #default="{ row }">
              <el-tag size="small" :type="row.is_folder ? 'warning' : 'info'" effect="plain">
                {{ row.is_folder ? '文件夹' : '文件' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="updated_at" label="更新时间" width="170" sortable />
        </el-table>
      </div>
      <template #footer>
        <el-button @click="startFileDialogVisible = false">取消</el-button>
        <el-button @click="clearStartFile">清除选择</el-button>
        <el-button type="primary" @click="confirmStartFileSelection">确认选择</el-button>
      </template>
    </el-dialog>

    <!-- 预览结果对话框 -->
    <el-dialog v-model="previewVisible" title="重命名预览" width="800px">
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
        <p>* 过滤规则：如果设置了起始文件，则该文件之后（更旧）的文件将不执行转存。</p>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { Plus, Play, Edit, Trash2, RefreshCw, Folder, File, Info } from 'lucide-vue-next'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getTasks, createTask, updateTask, deleteTask, runTask, previewTask, parseShareLink } from '../api/task'
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

// 分享内容解析相关
const shareFiles = ref([])
const parsingShare = ref(false)
const startFileDialogVisible = ref(false)
const selectedStartFileName = ref('')
const tempStartFileId = ref('')

// 独立目录弹窗相关
const folderDialogVisible = ref(false)
const folderTreeRef = ref(null)
const selectedTreePath = ref('')
const newFolderName = ref('')
const creatingFolder = ref(false)

const selectedAccountPlatform = computed(() => {
  if (!form.value.account_id) return ''
  const account = accounts.value.find(acc => acc.id === form.value.account_id)
  if (account) {
    return account.platform === '139' ? '移动云盘' : 'Quark'
  }
  return ''
})

const form = ref({
  id: null,
  name: '',
  account_id: '',
  share_url: '',
  extract_code: '',
  save_path: '/',
  pattern: '',
  replacement: '',
  start_file_id: ''
})

// 手动打开解析起始文件的对话框
const openStartFileDialog = async () => {
  if (!form.value.share_url || !form.value.account_id) {
    return ElMessage.warning('请先填写执行账号和分享链接')
  }
  
  startFileDialogVisible.value = true
  parsingShare.value = true
  tempStartFileId.value = form.value.start_file_id
  shareFiles.value = []
  
  try {
    const data = await parseShareLink({
      account_id: form.value.account_id,
      share_url: form.value.share_url,
      extract_code: form.value.extract_code
    })
    shareFiles.value = data
    
    // 如果已经有选中的 ID，尝试匹配出文件名用于回显
    if (form.value.start_file_id) {
      const selected = data.find(f => f.id === form.value.start_file_id)
      if (selected) {
        selectedStartFileName.value = selected.name
      }
    }
  } catch (err) {
    console.error('解析链接失败:', err)
    ElMessage.error('解析分享链接失败，请检查链接或账号状态')
    startFileDialogVisible.value = false
  } finally {
    parsingShare.value = false
  }
}

const handleStartFileTableChange = (row) => {
  if (row) {
    tempStartFileId.value = row.id
  }
}

const confirmStartFileSelection = () => {
  if (tempStartFileId.value) {
    form.value.start_file_id = tempStartFileId.value
    const selected = shareFiles.value.find(f => f.id === tempStartFileId.value)
    if (selected) {
      selectedStartFileName.value = selected.name
    }
  }
  startFileDialogVisible.value = false
}

const clearStartFile = () => {
  form.value.start_file_id = ''
  selectedStartFileName.value = ''
  tempStartFileId.value = ''
  startFileDialogVisible.value = false
}

let pollTimer = null

const fetchList = async (silent = false) => {
  if (!silent) loading.value = true
  try {
    const [taskData, accountData] = await Promise.all([getTasks(), getAccounts()])
    taskList.value = taskData
    accounts.value = accountData
  } catch (err) {
    console.error(err)
  } finally {
    if (!silent) loading.value = false
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
    folders.forEach(f => {
      pathIdMap.value[f.path] = f.id
    })
    resolve(folders)
  } catch (err) {
    console.error('加载目录失败:', err)
    resolve([])
  }
}

// 内嵌新建文件夹处理
const handleInlineCreateFolder = async () => {
  if (!newFolderName.value.trim()) {
    return ElMessage.warning('请输入文件夹名称')
  }
  
  const currentPath = selectedTreePath.value || '/'
  const currentID = pathIdMap.value[currentPath]

  if (currentID === undefined) {
    return ElMessage.warning('无法确定当前选中目录的 ID，请重新展开该节点')
  }

  creatingFolder.value = true
  try {
    const res = await createFolder(form.value.account_id, currentID, currentPath, newFolderName.value.trim())
    ElMessage.success('文件夹创建成功')
    pathIdMap.value[res.path] = res.id
    if (folderTreeRef.value) {
      const currentNode = folderTreeRef.value.getNode(currentPath)
      if (currentNode) {
        folderTreeRef.value.append(res, currentNode)
        currentNode.expanded = true
      }
    }
    selectedTreePath.value = res.path
    if (folderTreeRef.value) {
       folderTreeRef.value.setCurrentKey(res.path)
    }
    newFolderName.value = ''
  } catch (err) {
    console.error('创建文件夹失败:', err)
  } finally {
    creatingFolder.value = false
  }
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
  form.value.start_file_id = ''
  selectedStartFileName.value = ''
  pathIdMap.value = { '/': '' }
}

const openFolderDialog = () => {
  if (!form.value.account_id) {
    return ElMessage.warning('请先选择执行账号')
  }
  selectedTreePath.value = form.value.save_path || '/'
  newFolderName.value = ''
  folderDialogVisible.value = true
}

const handleTreeCurrentChange = (data) => {
  selectedTreePath.value = data.path
}

const confirmFolderSelection = () => {
  if (selectedTreePath.value) {
    form.value.save_path = selectedTreePath.value
  }
  folderDialogVisible.value = false
}

const openAddDialog = () => {
  form.value = { id: null, name: '', account_id: '', share_url: '', extract_code: '', save_path: '/', pattern: '', replacement: '', start_file_id: '' }
  shareFiles.value = []
  selectedStartFileName.value = ''
  dialogVisible.value = true
}

const handleEdit = async (row) => {
  shareFiles.value = []
  // 初始回显 ID 信息
  selectedStartFileName.value = row.start_file_id ? `载入中 (ID: ${row.start_file_id})...` : ''
  
  form.value = { 
    id: row.id,
    name: row.name,
    account_id: row.account_id,
    share_url: row.share_url,
    extract_code: row.extract_code,
    save_path: row.save_path,
    pattern: row.pattern,
    replacement: row.replacement,
    start_file_id: row.start_file_id
  }
  dialogVisible.value = true

  // 关键优化：如果带有起始文件 ID，尝试在后台找回文件名
  if (row.start_file_id && row.share_url) {
    try {
      const data = await parseShareLink({
        account_id: row.account_id,
        share_url: row.share_url,
        extract_code: row.extract_code
      })
      const found = data.find(f => f.id === row.start_file_id)
      if (found) {
        selectedStartFileName.value = found.name
      } else {
        selectedStartFileName.value = `ID: ${row.start_file_id} (文件在分享中已失效)`
      }
    } catch (err) {
      selectedStartFileName.value = `ID: ${row.start_file_id}`
    }
  }
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
  const map = { pending: 'info', running: 'primary', success: 'success', failed: 'danger' }
  return map[status] || 'info'
}

const formatTime = (timeStr) => {
  if (!timeStr || timeStr.startsWith('0001')) return '从不'
  return new Date(timeStr).toLocaleString()
}

onMounted(() => {
  fetchList()
  pollTimer = setInterval(() => {
    fetchList(true)
  }, 5000)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
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

.icon-spin {
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

.start-id-tip {
  margin-top: 6px;
  font-size: 12px;
  color: #64748b;
  display: flex;
  align-items: center;
  gap: 4px;
}

.id-code {
  background-color: var(--el-fill-color-light);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: monospace;
  color: var(--el-color-primary);
  max-width: 300px;
  display: inline-block;
  vertical-align: bottom;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.share-files-dialog-content {
  padding: 10px 0;
}

.share-tips {
  font-size: 12px;
  color: #94a3b8;
  margin-top: 8px;
}

.save-path-input {
  width: 100%;
}

.save-path-input :deep(.el-input-group__prepend) {
  background-color: var(--el-fill-color-light);
  color: var(--el-text-color-regular);
  font-weight: bold;
}

.folder-tree-container {
  height: 400px;
  overflow-y: auto;
  border: 1px solid var(--el-border-color-light);
  border-radius: 4px;
  padding: 8px;
}

.folder-dialog-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.create-folder-action {
  flex: 1;
  margin-right: 20px;
  max-width: 300px;
}

.dialog-actions {
  flex-shrink: 0;
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

.custom-tree-node {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>
