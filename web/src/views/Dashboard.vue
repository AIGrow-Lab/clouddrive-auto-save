<template>
  <div class="dashboard-container">
    <div class="welcome-section">
      <h2>欢迎回来，管理员 👋</h2>
      <p>这是您今日的云端转存概览</p>
    </div>

    <el-row :gutter="24" class="stat-cards">
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card" body-style="padding: 20px">
          <div class="stat-icon purple">
            <Activity :size="24" />
          </div>
          <div class="stat-info">
            <div class="stat-label">运行中任务</div>
            <div class="stat-value">12</div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card" body-style="padding: 20px">
          <div class="stat-icon blue">
            <HardDrive :size="24" />
          </div>
          <div class="stat-info">
            <div class="stat-label">已保存容量</div>
            <div class="stat-value">1.2 TB</div>
          </div>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card" body-style="padding: 20px">
          <div class="stat-icon green">
            <RefreshCw :size="24" />
          </div>
          <div class="stat-info">
            <div class="stat-label">今日完成</div>
            <div class="stat-value">48</div>
          </div>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card" body-style="padding: 20px">
          <div class="stat-icon orange">
            <User :size="24" />
          </div>
          <div class="stat-info">
            <div class="stat-label">活跃账号</div>
            <div class="stat-value">5</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="24" class="content-row">
      <el-col :span="16">
        <el-card class="chart-card">
          <template #header>
            <div class="card-header">
              <span>近期转存趋势</span>
              <el-button link type="primary">查看详情</el-button>
            </div>
          </template>
          <div class="placeholder-chart">
            <!-- 这里后续可以集成 ECharts -->
            <div class="chart-bar" style="height: 60%"></div>
            <div class="chart-bar" style="height: 80%"></div>
            <div class="chart-bar" style="height: 40%"></div>
            <div class="chart-bar" style="height: 90%"></div>
            <div class="chart-bar" style="height: 55%"></div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card class="activity-card">
          <template #header>
            <span>实时动态</span>
          </template>
          <el-timeline>
            <el-timeline-item timestamp="刚刚" type="primary">
              夸克任务 [黑镜] 转存成功
            </el-timeline-item>
            <el-timeline-item timestamp="10分钟前">
              移动云盘账号 Token 已自动刷新
            </el-timeline-item>
            <el-timeline-item timestamp="1小时前" type="success">
              系统完成全量任务巡检
            </el-timeline-item>
          </el-timeline>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { 
  Activity, 
  HardDrive, 
  RefreshCw, 
  User 
} from 'lucide-vue-next'
</script>

<style scoped>
.welcome-section {
  margin-bottom: 32px;
}

.welcome-section h2 {
  margin: 0;
  font-size: 24px;
  color: #1e293b;
}

html.dark .welcome-section h2 {
  color: #f1f5f9;
}

.welcome-section p {
  color: #64748b;
  margin: 8px 0 0 0;
}

.stat-card {
  display: flex;
  align-items: center;
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 16px;
}

.stat-icon.purple { background: rgba(139, 92, 246, 0.1); color: #8b5cf6; }
.stat-icon.blue { background: rgba(59, 130, 246, 0.1); color: #3b82f6; }
.stat-icon.green { background: rgba(16, 185, 129, 0.1); color: #10b981; }
.stat-icon.orange { background: rgba(245, 158, 11, 0.1); color: #f59e0b; }

.stat-label {
  font-size: 14px;
  color: #64748b;
}

.stat-value {
  font-size: 20px;
  font-weight: 700;
  color: #1e293b;
}

html.dark .stat-value { color: #f1f5f9; }

.content-row {
  margin-top: 24px;
}

.placeholder-chart {
  height: 240px;
  display: flex;
  align-items: flex-end;
  justify-content: space-around;
  padding: 20px;
}

.chart-bar {
  width: 40px;
  background: var(--el-color-primary);
  border-radius: 8px 8px 0 0;
  opacity: 0.6;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
