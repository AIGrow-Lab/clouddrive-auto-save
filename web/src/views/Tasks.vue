<template>
  <div class="tasks-container">
    <div class="page-header">
      <div class="title-section">
        <h2>任务管理</h2>
        <p>监控并自动转存移动云盘和夸克网盘的分享资源</p>
      </div>
      <div class="header-actions">
        <el-popconfirm
          title="确定要一键启动所有可运行的任务吗？"
          confirm-button-text="确认"
          cancel-button-text="取消"
          width="240"
          @confirm="handleRunAll"
        >
          <template #reference>
            <el-button :icon="Play" :loading="runningAll">全部运行</el-button>
          </template>
        </el-popconfirm>
        <el-button type="primary" :icon="Plus" @click="openAddDialog">创建任务</el-button>
      </div>
    </div>

    <el-card class="table-card">
      <el-table v-if="taskList.length > 0 || loading" :data="taskList" v-loading="loading" style="width: 100%">
        <el-table-column label="任务名称" min-width="180">
          <template #default="{ row }">
            <div class="task-name-cell">
              <span class="name">{{ row.name }}</span>
              <div class="account-tag" v-if="row.account">
                <el-tag size="small" :type="row.account.platform === 'quark' ? 'success' : 'warning'">
                  {{ row.account.nickname || row.account.platform }}
                </el-tag>
              </div>
              <div class="account-tag" v-else>
                <el-tag size="small" type="danger">账号已移除</el-tag>
              </div>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column prop="save_path" label="保存路径" min-width="150" show-overflow-tooltip />
        
        <el-table-column prop="schedule_mode" label="调度规则" width="140">
          <template #default="{ row }">
            <div v-if="row.schedule_mode === 'global'" class="schedule-tag">
              <el-tag size="small" type="primary" effect="plain">跟随全局</el-tag>
              <div class="schedule-sub" v-if="globalSchedule.enabled">{{ globalSchedule.cron }}</div>
              <div class="schedule-sub disabled" v-else>全局已关闭</div>
            </div>
            <div v-else-if="row.schedule_mode === 'custom'" class="schedule-tag">
              <el-tag size="small" type="warning" effect="plain">自定义</el-tag>
              <div class="schedule-sub">{{ row.cron }}</div>
            </div>
            <el-tag v-else size="small" type="info" effect="plain">手动执行</el-tag>
          </template>
        </el-table-column>

        <el-table-column label="状态" width="120">
          <template #default="{ row }">
            <div class="status-wrapper">
              <el-tooltip v-if="row.message && row.message.includes('[Fatal]')" :content="row.message" placement="top" effect="dark">
                <el-tag type="danger" style="cursor:help"><div class="status-inner"><el-icon><AlertTriangle /></el-icon>LINK ERROR</div></el-tag>
              </el-tooltip>
              <el-tag v-else :type="getStatusType(row.status)">
                <div class="status-inner"><el-icon v-if="row.status === 'running'" class="icon-spin"><RefreshCw /></el-icon>{{ row.status.toUpperCase() }}</div>
              </el-tag>
            </div>
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
              <el-button 
                link 
                type="success" 
                :icon="Play" 
                :disabled="row.status === 'running' || !!(row.message && row.message.includes('[Fatal]'))" 
                @click="handleRun(row)"
              >
                运行
              </el-button>
              <el-button link type="primary" :icon="Edit" @click="handleEdit(row)">编辑</el-button>
              <el-button link type="danger" :icon="Trash2" @click="handleDelete(row)">删除</el-button>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-else description="当前没有任何转存任务">
        <el-button type="primary" :icon="Plus" @click="openAddDialog">创建新任务</el-button>
      </el-empty>
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
                <el-option-group
                  v-for="group in groupedAccounts"
                  :key="group.label"
                  :label="group.label"
                >
                  <el-option
                    v-for="acc in group.options"
                    :key="acc.id"
                    :value="acc.id"
                    :disabled="acc.status === 0"
                    :label="acc.nickname"
                  >
                    <div class="account-option-item">
                      <div class="acc-info">
                        <el-icon class="acc-icon" :color="acc.platform === 'quark' ? '#10b981' : '#f59e0b'">
                          <Cloud />
                        </el-icon>
                        <span class="acc-name">{{ acc.nickname }}</span>
                      </div>
                      <div class="acc-meta">
                        <span class="acc-cap" v-if="acc.capacity_total > 0">
                          剩余 {{ formatSize(acc.capacity_total - acc.capacity_used) }}
                        </span>
                        <el-tag v-if="acc.status === 0" size="small" type="danger" effect="dark">已失效</el-tag>
                      </div>
                    </div>
                  </el-option>
                </el-option-group>
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="分享链接" required>
          <el-input v-model="form.share_url" placeholder="请输入 139 或 Quark 分享链接" @change="handleUrlChange">
            <template #append>
              <el-button
                :icon="FolderOpen"
                title="浏览分享内容并选择目录"
                :disabled="!form.share_url || !form.account_id"
                @click="openBrowseShareDialog"
              />
              <div class="append-divider"></div>
              <el-button
                :icon="ExternalLink"
                title="在新标签页中打开链接"
                :disabled="!form.share_url"
                @click="openExternalLink(form.share_url, form.extract_code)"
              />
            </template>
          </el-input>
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

        <el-divider>调度规则与预览</el-divider>
        
        <el-row :gutter="20">
          <el-col :span="24">
            <el-form-item label="调度模式" required>
              <div class="schedule-mode-selector">
                <el-radio-group v-model="form.schedule_mode" @change="handleScheduleModeChange">
                  <el-radio label="global">跟随全局</el-radio>
                  <el-radio label="custom">自定义频率</el-radio>
                  <el-radio label="off">手动执行</el-radio>
                </el-radio-group>
              </div>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20" v-if="form.schedule_mode === 'custom'">
          <el-col :span="24">
            <el-form-item label="自定义频率 (Cron)">
              <div style="display: flex; align-items: center; gap: 15px; width: 100%;">
                <el-select v-model="cronPreset" placeholder="选择预设频率" style="width: 180px" @change="handleCronPreset">
                  <el-option label="每小时" value="0 0 * * * *" />
                  <el-option label="每 6 小时" value="0 0 */6 * * *" />
                  <el-option label="每天凌晨 2 点" value="0 0 2 * * *" />
                  <el-option label="每周一凌晨 2 点" value="0 0 2 * * 1" />
                  <el-option label="使用自定义表达式" value="custom" />
                </el-select>
                <el-input v-if="cronPreset === 'custom'" v-model="form.cron" placeholder="秒 分 时 日 月 周" style="flex: 1">
                  <template #suffix>
                    <el-tooltip placement="top">
                      <template #content>
                        格式：秒 分 时 日 月 周 (6位)<br/>
                        例如: */5 * * * * * (每5秒执行一次)
                      </template>
                      <el-icon style="cursor: help"><Info /></el-icon>
                    </el-tooltip>
                  </template>
                </el-input>
              </div>
            </el-form-item>
          </el-col>
        </el-row>

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
          :default-expanded-keys="['/']"
          highlight-current
          @current-change="handleTreeCurrentChange"
          :expand-on-click-node="false"
          :empty-text="loadingFolders ? '加载中...' : '暂无目录'"
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

    <!-- 选择起始文件 / 浏览分享内容弹窗 -->
    <el-dialog
      v-model="startFileDialogVisible"
      :title="browseMode === 'selectShareUrl' ? '浏览分享内容' : '选择起始转存文件'"
      width="900px"
      append-to-body
      destroy-on-close
    >
      <div class="share-files-dialog-content">
        <!-- 面包屑导航 -->
        <div class="breadcrumb-nav" style="margin-bottom: 12px;">
          <el-breadcrumb separator="/">
            <el-breadcrumb-item>
              <a href="#" @click.prevent="navigateToBreadcrumb(-1)" class="breadcrumb-link">根目录</a>
            </el-breadcrumb-item>
            <el-breadcrumb-item v-for="(crumb, index) in breadcrumbs" :key="crumb.id">
              <a href="#" @click.prevent="navigateToBreadcrumb(index)" class="breadcrumb-link">{{ crumb.name }}</a>
            </el-breadcrumb-item>
          </el-breadcrumb>
        </div>

        <el-alert v-if="browseMode === 'selectShareUrl'" title="选择目录" type="info" :closable="false" show-icon style="margin-bottom: 15px">
          点击文件夹进入子目录，选择目标文件夹后点击"选择为分享链接"，将更新分享链接为该目录的地址。
        </el-alert>
        <el-alert v-else title="逻辑说明" type="info" :closable="false" show-icon style="margin-bottom: 15px">
          系统将从选中的文件开始，按更新时间向前转存所有更新的文件（含所选文件本身）。<b>此处已应用您的重命名规则并执行同名预检。</b> 点击文件夹可进入子目录。
        </el-alert>

        <el-table
          :data="shareFiles"
          max-height="500"
          size="default"
          border
          stripe
          v-loading="parsingShare"
          highlight-current-row
          :row-class-name="tableRowClassName"
          @current-change="handleStartFileTableChange"
          @row-dblclick="handleRowDblClick"
        >
          <!-- startFile 模式且在根目录时显示 radio 列 -->
          <el-table-column v-if="browseMode === 'startFile' && breadcrumbs.length === 0" width="40" align="center">
            <template #default="{ row }">
              <el-radio v-if="!row.is_folder" v-model="tempStartFileId" :label="row.id" class="naked-radio"><span></span></el-radio>
            </template>
          </el-table-column>

          <el-table-column label="原始文件名" show-overflow-tooltip min-width="180">
            <template #default="{ row }">
              <div class="name-main" :class="{ 'folder-clickable': row.is_folder }" @click="row.is_folder && enterFolder(row)" @dblclick="!row.is_folder && handleRowDblClick(row)">
                <el-icon size="16">
                  <Folder v-if="row.is_folder" color="#eab308" />
                  <File v-else color="#64748b" />
                </el-icon>
                <span>{{ row.name }}</span>
              </div>
            </template>
          </el-table-column>

          <el-table-column label="预览文件名 (入库名)" show-overflow-tooltip min-width="220">
            <template #default="{ row }">
              <span :style="{
                fontWeight: row.is_folder ? '600' : 'normal',
                color: (row.new_name && row.new_name !== row.name) ? 'var(--el-color-primary)' : 'inherit'
              }">
                {{ row.new_name || row.name }}
              </span>
            </template>
          </el-table-column>

          <el-table-column label="类型" width="80" align="center">
            <template #default="{ row }">
              <el-tag size="small" :type="row.is_folder ? 'warning' : 'info'" effect="plain">
                {{ row.is_folder ? '目录' : '文件' }}
              </el-tag>
            </template>
          </el-table-column>

          <el-table-column label="状态" width="100" align="center">
            <template #default="{ row }">
              <el-tag v-if="row.is_existed" size="small" type="success" effect="light">已在网盘</el-tag>
              <el-tag v-else size="small" type="info" effect="plain">待转存</el-tag>
            </template>
          </el-table-column>

          <el-table-column prop="updated_at" label="分享更新时间" width="160" sortable />

          <!-- selectShareUrl 模式显示进入按钮 -->
          <el-table-column v-if="browseMode === 'selectShareUrl'" label="操作" width="80" align="center">
            <template #default="{ row }">
              <el-button v-if="row.is_folder" type="primary" link size="small" @click="enterFolder(row)">
                进入
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
      <template #footer>
        <el-button @click="startFileDialogVisible = false">取消</el-button>
        <template v-if="browseMode === 'selectShareUrl'">
          <el-button type="primary" @click="confirmSelectShareUrl">
            选择当前目录（{{ currentDirName }}）
          </el-button>
        </template>
        <template v-else>
          <el-button @click="clearStartFile">清除选择</el-button>
          <el-button type="primary" @click="confirmStartFileSelection" :disabled="!tempStartFileId">确认选择</el-button>
        </template>
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
import { Plus, Play, Edit, Trash2, RefreshCw, Folder, File, Info, Cloud, ExternalLink, AlertTriangle, Clock, FolderOpen } from 'lucide-vue-next'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getTasks, createTask, updateTask, deleteTask, runTask, runAllTasks, previewTask, parseShareLink, getScheduleSettings } from '../api/task'
import { getAccounts, getFolders, createFolder } from '../api/account'

const taskList = ref([])
const accounts = ref([])
const loading = ref(false)
const runningAll = ref(false)
const dialogVisible = ref(false)
const submitting = ref(false)

// 全局调度设置
const globalSchedule = ref({
  enabled: false,
  cron: ''
})

const fetchGlobalSettings = async () => {
  try {
    const data = await getScheduleSettings()
    globalSchedule.value = data
  } catch (err) {
    console.error('获取全局设置失败:', err)
  }
}

const saveGlobalSettings = async () => {
  try {
    await updateScheduleSettings(globalSchedule.value)
    ElMessage.success('全局调度设置已更新')
  } catch (err) {
    console.error('更新全局设置失败:', err)
  }
}

// 定时执行相关
const cronPreset = ref('')

const handleScheduleModeChange = (val) => {
  if (val === 'custom' && !form.value.cron) {
    cronPreset.value = '0 0 * * * *'
    form.value.cron = '0 0 * * * *'
  }
}

const handleCronPreset = (val) => {
  if (val !== 'custom') {
    form.value.cron = val
  }
}

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

// 子目录浏览相关
const breadcrumbs = ref([]) // [{ id, name }]
const currentParentId = ref('')
const browseMode = ref('startFile') // 'startFile' | 'selectShareUrl'

// 计算当前目录名称（用于 selectShareUrl 模式的按钮显示）
const currentDirName = computed(() => {
  if (breadcrumbs.value.length === 0) {
    return '根目录'
  }
  return breadcrumbs.value[breadcrumbs.value.length - 1].name
})

// 处理表格行样式
const tableRowClassName = ({ row }) => {
  if (row.is_existed) {
    return 'existed-row'
  }
  return ''
}

// 独立目录弹窗相关
const folderDialogVisible = ref(false)
const loadingFolders = ref(false)
const folderTreeRef = ref(null)
const selectedTreePath = ref('')
const selectedTreeId = ref('')
const newFolderName = ref('')
const creatingFolder = ref(false)

const formatSize = (bytes) => {
  const b = Number(bytes)
  if (isNaN(b) || b <= 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(b) / Math.log(1024))
  return `${(b / Math.pow(1024, i)).toFixed(2)} ${units[i]}`
}

const groupedAccounts = computed(() => {
  if (!accounts.value) return []
  const groups = [
    { label: '移动云盘', platform: '139', options: [] },
    { label: '夸克网盘', platform: 'quark', options: [] }
  ]
  
  accounts.value.forEach(acc => {
    if (!acc) return
    const group = groups.find(g => g.platform === acc.platform)
    if (group) {
      group.options.push(acc)
    }
  })
  
  return groups.filter(g => g.options.length > 0)
})

const selectedAccountPlatform = computed(() => {
  if (!form.value.account_id || !accounts.value) return ''
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
  start_file_id: '',
  start_file_name: '',
  cron: '',
  schedule_mode: 'global'
})

// 链接变更处理
const handleUrlChange = () => {
  form.value.start_file_id = ''
  form.value.start_file_name = ''
  selectedStartFileName.value = ''
}

const openExternalLink = async (url, extractCode) => {
  if (url) {
    let finalUrl = url
    if (extractCode) {
      try {
        const urlObj = new URL(url)
        urlObj.searchParams.set('pwd', extractCode)
        finalUrl = urlObj.toString()
      } catch (e) {
        // 针对非标准 URL 的 Fallback 处理
        finalUrl = url.includes('?') 
          ? `${url}&pwd=${extractCode}` 
          : `${url}?pwd=${extractCode}`
      }
      
      try {
        await navigator.clipboard.writeText(extractCode)
        ElMessage.success(`提取码 ${extractCode} 已复制，请在新页面粘贴`)
      } catch (err) {
        console.warn('复制提取码失败:', err)
      }
    }
    window.open(finalUrl, '_blank')
  }
}

// 手动打开解析起始文件的对话框
const openStartFileDialog = async () => {
  if (!form.value.share_url || !form.value.account_id) {
    return ElMessage.warning('请先填写执行账号和分享链接')
  }

  browseMode.value = 'startFile'
  startFileDialogVisible.value = true
  parsingShare.value = true
  tempStartFileId.value = form.value.start_file_id
  shareFiles.value = []
  breadcrumbs.value = []

  // 使用 share_parent_id 作为初始目录（139 平台）
  const initialParentId = form.value.share_parent_id || ''
  currentParentId.value = initialParentId

  // 如果有 share_parent_id，设置面包屑导航
  if (initialParentId) {
    // 从 share_parent_id 中提取目录名（如果可能）
    // 这里我们只知道 ID，不知道名称，所以显示为 "子目录"
    breadcrumbs.value = [{ id: initialParentId, name: '已选择的目录' }]
  }

  await loadShareFiles(initialParentId)
}

// 打开浏览分享内容对话框（用于选择目录作为新的分享链接）
const openBrowseShareDialog = async () => {
  if (!form.value.share_url || !form.value.account_id) {
    return ElMessage.warning('请先填写执行账号和分享链接')
  }

  browseMode.value = 'selectShareUrl'
  startFileDialogVisible.value = true
  parsingShare.value = true
  tempStartFileId.value = ''
  shareFiles.value = []
  breadcrumbs.value = []
  currentParentId.value = ''

  await loadShareFiles('')
}

// 加载分享文件列表（支持子目录浏览）
const loadShareFiles = async (parentId) => {
  parsingShare.value = true
  shareFiles.value = []
  currentParentId.value = parentId

  try {
    const data = await parseShareLink({
      account_id: form.value.account_id,
      share_url: form.value.share_url,
      extract_code: form.value.extract_code,
      parent_id: parentId,
      save_path: form.value.save_path,
      pattern: form.value.pattern,
      replacement: form.value.replacement,
      name: form.value.name
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
    startFileDialogVisible.value = false
  } finally {
    parsingShare.value = false
  }
}

// 进入子目录
const enterFolder = async (folder) => {
  breadcrumbs.value.push({ id: folder.id, name: folder.name })
  await loadShareFiles(folder.id)
}

// 点击面包屑导航
const navigateToBreadcrumb = async (index) => {
  if (index === -1) {
    // 返回根目录
    breadcrumbs.value = []
    await loadShareFiles('')
  } else {
    breadcrumbs.value = breadcrumbs.value.slice(0, index + 1)
    await loadShareFiles(breadcrumbs.value[index].id)
  }
}

const handleStartFileTableChange = (row) => {
  if (row) {
    tempStartFileId.value = row.id
  }
}

// 双击行处理：文件夹进入，文件选中（仅在 startFile 模式的根目录下）
const handleRowDblClick = (row) => {
  if (browseMode.value === 'startFile' && !row.is_folder && breadcrumbs.value.length === 0) {
    tempStartFileId.value = row.id
    confirmStartFileSelection()
  }
}

const confirmStartFileSelection = () => {
  if (tempStartFileId.value) {
    form.value.start_file_id = tempStartFileId.value
    const selected = shareFiles.value.find(f => f.id === tempStartFileId.value)
    if (selected) {
      form.value.start_file_name = selected.name
      selectedStartFileName.value = selected.name
    }
  }
  startFileDialogVisible.value = false
}

// 确认选择目录作为新的分享链接
const confirmSelectShareUrl = () => {
  // 根据平台生成新的分享链接
  const account = accounts.value.find(acc => acc.id === form.value.account_id)
  if (!account) return

  const originalUrl = form.value.share_url
  let newUrl = originalUrl

  // 获取当前目录 ID（根目录时使用默认值）
  const currentDirId = currentParentId.value || ''

  if (account.platform === 'quark') {
    // 夸克网盘：替换 URL 中的 pdirFID
    // 格式：https://pan.quark.cn/s/{pwdID}#/list/share/{pdirFID}
    const match = originalUrl.match(/\/s\/(\w+)/)
    if (match) {
      const pwdID = match[1]
      // 根目录时使用 "0"，子目录时使用 currentDirId
      const pdirFID = currentDirId || '0'
      newUrl = `https://pan.quark.cn/s/${pwdID}#/list/share/${pdirFID}`
    }
    // 夸克网盘不需要 share_parent_id，清空它
    form.value.share_parent_id = ''
  } else if (account.platform === '139') {
    // 移动云盘：URL 不变，但存储 share_parent_id
    // 139 通过 pCaID 区分目录，URL 格式不变
    form.value.share_parent_id = currentDirId || ''
    newUrl = originalUrl
  }

  form.value.share_url = newUrl
  // 清空起始文件选择
  form.value.start_file_id = ''
  form.value.start_file_name = ''
  selectedStartFileName.value = ''

  ElMessage.success(`已选择目录：${currentDirName.value}`)
  startFileDialogVisible.value = false
}

const clearStartFile = () => {
  form.value.start_file_id = ''
  form.value.start_file_name = ''
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
  
  if (node.level === 0) {
    pathIdMap.value['/'] = '' 
    return resolve([{ label: '根目录', path: '/', id: '', isLeaf: false }])
  }

  const parentID = node.data.id
  const parentPath = node.data.path
  
  // 仅在根目录加载时显示全局遮罩，子目录展开将利用 Tree 自带的 node-loading
  if (node.level === 1) {
    loadingFolders.value = true
  }
  
  try {
    const folders = await getFolders(form.value.account_id, parentID, parentPath)
    
    // 异步更新映射表，不阻塞渲染
    setTimeout(() => {
      const newMappings = {}
      folders.forEach(f => {
        newMappings[f.path] = f.id
      })
      Object.assign(pathIdMap.value, newMappings)
    }, 0)
    
    resolve(folders)
  } catch (err) {
    console.error('加载目录失败:', err)
    resolve([])
  } finally {
    if (node.level === 1) {
      loadingFolders.value = false
    }
  }
}

// 内嵌新建文件夹处理
const handleInlineCreateFolder = async () => {
  if (!newFolderName.value.trim()) {
    return ElMessage.warning('请输入文件夹名称')
  }
  
  const currentPath = selectedTreePath.value || '/'
  const currentID = selectedTreeId.value || '' 

  if (folderTreeRef.value) {
    const currentNode = folderTreeRef.value.getNode(currentPath)
    if (currentNode && currentNode.childNodes) {
      const isDuplicate = currentNode.childNodes.some(
        child => child.data && child.data.label === newFolderName.value.trim()
      )
      if (isDuplicate) {
        return ElMessage.warning('该目录下已存在同名文件夹')
      }
    }
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
      selectedTreePath.value = res.path
      selectedTreeId.value = res.id
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
  form.value.start_file_name = ''
  selectedStartFileName.value = ''
  pathIdMap.value = { '/': '' }
}

const openFolderDialog = () => {
  if (!form.value.account_id) {
    return ElMessage.warning('请先选择执行账号')
  }
  loadingFolders.value = true
  const initialPath = form.value.save_path || '/'
  selectedTreePath.value = initialPath
  // 尝试从映射表中恢复 ID，若是根目录则默认为空字符串
  selectedTreeId.value = pathIdMap.value[initialPath] || (initialPath === '/' ? '' : '')
  newFolderName.value = ''
  folderDialogVisible.value = true
}

const handleTreeCurrentChange = (data) => {
  if (data) {
    selectedTreePath.value = data.path
    selectedTreeId.value = data.id
  }
}

const confirmFolderSelection = () => {
  if (selectedTreePath.value) {
    form.value.save_path = selectedTreePath.value
  }
  folderDialogVisible.value = false
}

const openAddDialog = () => {
  form.value = {
    id: null,
    name: '',
    account_id: '',
    share_url: '',
    extract_code: '',
    save_path: '/',
    pattern: '',
    replacement: '',
    start_file_id: '',
    start_file_name: '',
    share_parent_id: '',
    cron: '',
    schedule_mode: 'global'
  }
  shareFiles.value = []
  selectedStartFileName.value = ''
  cronPreset.value = ''
  dialogVisible.value = true
}

const handleEdit = async (row) => {
  shareFiles.value = []
  
  form.value = {
    id: row.id,
    name: row.name,
    account_id: row.account_id,
    share_url: row.share_url,
    extract_code: row.extract_code,
    save_path: row.save_path,
    pattern: row.pattern,
    replacement: row.replacement,
    start_file_id: row.start_file_id,
    start_file_name: row.start_file_name,
    share_parent_id: row.share_parent_id || '',
    cron: row.cron,
    schedule_mode: row.schedule_mode || 'global'
  }

  // 初始化定时配置状态
  if (form.value.schedule_mode === 'custom' && row.cron) {
    const presets = ['0 0 * * * *', '0 0 */6 * * *', '0 0 2 * * *', '0 0 2 * * 1']
    cronPreset.value = presets.includes(row.cron) ? row.cron : 'custom'
  } else {
    cronPreset.value = ''
  }

  if (row.start_file_id) {
    selectedStartFileName.value = row.start_file_name || `ID: ${row.start_file_id} (文件名未记录)`
  } else {
    selectedStartFileName.value = ''
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
      const updatedTask = await updateTask(form.value.id, form.value)
      // 原地更新列表中的任务数据
      const idx = taskList.value.findIndex(t => t.id === form.value.id)
      if (idx > -1) {
        // 合并数据，保留可能需要的关联对象（如 account）
        Object.assign(taskList.value[idx], updatedTask)
      }
      ElMessage.success('任务更新成功')
    } else {
      await createTask(form.value)
      ElMessage.success('任务保存成功')
      fetchList() // 新增任务还是刷新一下列表比较稳妥
    }
    dialogVisible.value = false
  } catch (err) {
    console.error(err)
  } finally {
    submitting.value = false
  }
}

const handleRun = async (row) => {
  try {
    await runTask(row.id)
    // 立即在前端反馈状态，避免等待轮询
    // 增加检查，防止覆盖已经由 SSE 更新的成功状态
    if (row.status !== 'success') {
      row.status = 'running'
    }
    ElMessage.success('任务已提交执行队列')
  } catch (err) {
    // 错误已由拦截器展示
  }
}

const handleRunAll = async () => {
  runningAll.value = true
  try {
    const res = await runAllTasks()
    ElMessage.success(res.message || `批量执行已开启，已成功触发 ${res.count} 个任务`)
    
    // 立即在前端反馈所有满足条件的任务状态
    taskList.value.forEach(task => {
      const isFatal = task.message && task.message.includes('[Fatal]')
      if (task.status !== 'running' && task.status !== 'success' && !isFatal) {
        task.status = 'running'
      }
    })
  } catch (err) {
    // 错误已由拦截器展示
  } finally {
    runningAll.value = false
  }
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

let eventSource = null

const initSSE = () => {
  eventSource = new EventSource('/api/dashboard/logs')
  eventSource.onmessage = (event) => {
    const msg = event.data
    if (msg.includes('[EVENT:')) {
      const match = msg.match(/\[EVENT:(.+)\]/)
      if (match) {
        try {
          const ev = JSON.parse(match[1])
          if (ev.type === 'task_update' && ev.task) {
            const idx = taskList.value.findIndex(t => t.id === ev.task.id)
            if (idx > -1) {
              const row = taskList.value[idx]
              // 手动更新进度相关字段，确保触发响应式，同时防止覆盖 account 信息
              row.status = ev.task.status
              row.message = ev.task.message
              row.last_run = ev.task.last_run
              row.percent = ev.task.percent
              row.stage = ev.task.stage

              // 如果任务完成或失败，额外执行一次列表同步以确保最终一致性
              if (ev.task.status === 'success' || ev.task.status === 'failed') {
                fetchList(true)
              }
            } else {
              // 如果是新任务且列表里没有（可能是刚创建的），刷新一次列表
              fetchList()
            }
          } else if (ev.type === 'task_delete') {
            const idx = taskList.value.findIndex(t => t.id === ev.taskId)
            if (idx > -1) {
              taskList.value.splice(idx, 1)
            }
          }
        } catch (e) {
          console.error('解析实时事件失败:', e)
        }
      }
    }
  }
}

onMounted(() => {
  fetchList()
  fetchGlobalSettings()
  initSSE()
})

onUnmounted(() => {
  if (eventSource) eventSource.close()
})
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
}

.header-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.title-section h2 {

  margin: 0;
  font-size: 24px;
}

.title-section p {
  color: #64748b;
  margin: 4px 0 0 0;
}

.global-settings-card {
  border-radius: 12px;
  background: linear-gradient(135deg, var(--el-color-primary-light-9) 0%, var(--el-color-white) 100%);
  border: 1px solid var(--el-color-primary-light-7);
}

.global-settings-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 20px;
}

.setting-info {
  display: flex;
  align-items: center;
  gap: 16px;
}

.setting-icon {
  font-size: 24px;
  color: var(--el-color-primary);
  background: var(--el-color-white);
  padding: 10px;
  border-radius: 10px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.05);
}

.setting-text {
  display: flex;
  flex-direction: column;
}

.setting-title {
  font-weight: 600;
  font-size: 16px;
  color: var(--el-text-color-primary);
}

.setting-desc {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  margin-top: 2px;
}

.setting-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.action-item {
  display: flex;
  align-items: center;
  gap: 10px;
}

.action-label {
  font-size: 13px;
  color: var(--el-text-color-regular);
  font-weight: 500;
}

.schedule-tag {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.schedule-sub {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  font-family: monospace;
}

.schedule-sub.disabled {
  color: var(--el-color-danger);
}

.schedule-mode-selector {
  margin-bottom: 10px;
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

.existed-row {
  background-color: var(--el-fill-color-lighter) !important;
  color: #94a3b8;
}

.existed-row span {
  opacity: 0.8;
}

.existed-tag {
  margin-left: 8px;
  flex-shrink: 0;
}

/* 彻底关闭目录树展开动画，消除“一行行刷新”的视觉差 */
:deep(.el-tree-node__children) {
  transition: none !important;
}
:deep(.el-collapse-transition-leave-active),
:deep(.el-collapse-transition-enter-active) {
  transition: none !important;
  display: none !important;
}

.naked-radio :deep(.el-radio__label) {
  display: none !important;
}

.custom-tree-node {
  display: flex;
  align-items: center;
  gap: 8px;
}

.name-main {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
}

.name-sub-column {
  font-size: 13px;
  color: #64748b;
}

.account-option-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  gap: 12px;
}

.acc-info {
  display: flex;
  align-items: center;
  gap: 8px;
  overflow: hidden;
}

.acc-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.acc-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.acc-cap {
  font-size: 12px;
  color: #94a3b8;
}

.breadcrumb-nav {
  padding: 8px 0;
}

.breadcrumb-link {
  color: var(--el-color-primary);
  cursor: pointer;
  text-decoration: none;
}

.breadcrumb-link:hover {
  text-decoration: underline;
}

.folder-clickable {
  cursor: pointer;
}

.folder-clickable:hover {
  color: var(--el-color-primary);
}

.append-divider {
  width: 1px;
  height: 20px;
  background: var(--el-border-color);
  margin: 0 4px;
}
</style>
