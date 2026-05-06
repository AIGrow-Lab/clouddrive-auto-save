<template>
  <div class="sidebar-footer">
    <a
      class="github-link"
      href="https://github.com/zhaocongqi/clouddrive-auto-save"
      target="_blank"
      rel="noopener noreferrer"
    >
      <Github :size="16" />
      <span>GitHub 仓库</span>
      <ExternalLink :size="12" />
    </a>
    <div
      class="version-info"
      :class="{ 'has-update': hasUpdate }"
      @click="hasUpdate && openReleases()"
    >
      <span class="version-text">v{{ currentVersion }}</span>
      <span v-if="hasUpdate" class="update-badge">NEW</span>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Github, ExternalLink } from 'lucide-vue-next'
import { getVersion } from '../api/version'

const currentVersion = ref('...')
const hasUpdate = ref(false)

const GITHUB_RELEASES_URL = 'https://github.com/zhaocongqi/clouddrive-auto-save/releases'
const GITHUB_API_URL = 'https://api.github.com/repos/zhaocongqi/clouddrive-auto-save/releases/latest'

function parseVersion(v) {
  const cleaned = v.replace(/^v/, '')
  const parts = cleaned.split('.').map(Number)
  return parts.length === 3 && parts.every(n => !isNaN(n)) ? parts : null
}

function compareVersions(current, latest) {
  const c = parseVersion(current)
  const l = parseVersion(latest)
  if (!c || !l) return false
  for (let i = 0; i < 3; i++) {
    if (l[i] > c[i]) return true
    if (l[i] < c[i]) return false
  }
  return false
}

function openReleases() {
  window.open(GITHUB_RELEASES_URL, '_blank', 'noopener,noreferrer')
}

onMounted(async () => {
  // 获取当前版本
  try {
    const res = await getVersion()
    currentVersion.value = res.version || 'dev'
  } catch {
    currentVersion.value = 'dev'
  }

  // 如果是 dev 版本，跳过更新检查
  if (currentVersion.value === 'dev') return

  // 检查 GitHub 最新 release
  try {
    const resp = await fetch(GITHUB_API_URL)
    if (!resp.ok) return
    const data = await resp.json()
    const latestTag = data.tag_name || ''
    hasUpdate.value = compareVersions(currentVersion.value, latestTag)
  } catch {
    // 静默失败
  }
})
</script>

<style scoped>
.sidebar-footer {
  margin-top: auto;
  padding: 12px 16px;
  border-top: 1px solid rgba(0, 0, 0, 0.05);
}

html.dark .sidebar-footer {
  border-top: 1px solid rgba(255, 255, 255, 0.05);
}

.github-link {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 8px;
  color: #64748b;
  text-decoration: none;
  font-size: 13px;
  transition: background 0.2s, color 0.2s;
}

.github-link:hover {
  background: #f1f5f9;
  color: #334155;
}

html.dark .github-link {
  color: #94a3b8;
}

html.dark .github-link:hover {
  background: rgba(255, 255, 255, 0.05);
  color: #e2e8f0;
}

.version-info {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 12px;
  font-size: 12px;
  color: #94a3b8;
}

.version-info.has-update {
  cursor: pointer;
  color: #64748b;
  border-radius: 8px;
  padding: 6px 12px;
  margin: 2px 0;
  transition: background 0.2s;
}

.version-info.has-update:hover {
  background: #f1f5f9;
}

html.dark .version-info.has-update:hover {
  background: rgba(255, 255, 255, 0.05);
}

.update-badge {
  display: inline-flex;
  align-items: center;
  padding: 1px 6px;
  border-radius: 10px;
  background: #ef4444;
  color: white;
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.5px;
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.7; }
}
</style>
