<template>
  <div class="cooperation-page">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="page-title">合作管理</h1>
    </div>

    <!-- 合作概览 -->
    <div class="cooperation-summary">
      <div class="summary-title">合作概览</div>
      <div class="summary-cards">
        <div class="summary-card">
          <div class="card-label">全部</div>
          <div class="card-value">
            <span class="value-number">{{ summaryData.total }}</span>
            <span class="value-compare">VS上周</span>
            <span class="value-percent positive">↑{{ summaryData.totalPercent }}%</span>
          </div>
        </div>
        <div class="summary-card">
          <div class="card-label">合作中</div>
          <div class="card-value">
            <span class="value-number">{{ summaryData.active }}</span>
          </div>
        </div>
        <div class="summary-card">
          <div class="card-label">已取消</div>
          <div class="card-value">
            <span class="value-number">{{ summaryData.cancelled }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 筛选区域 -->
    <div class="filter-section">
      <el-input
        v-model="filterForm.keyword"
        placeholder="店铺名称/店铺编号"
        :prefix-icon="Search"
        clearable
        style="width: 200px"
      />
      <el-select v-model="filterForm.status" placeholder="选择合作状态" clearable style="width: 150px">
        <el-option label="全部" value="" />
        <el-option label="合作中" value="active" />
        <el-option label="已取消" value="cancelled" />
      </el-select>
      <el-date-picker
        v-model="filterForm.startDate"
        type="date"
        placeholder="开始日期"
        value-format="YYYY-MM-DD"
        style="width: 200px"
      />
      <span style="color: #909399;">-</span>
      <el-date-picker
        v-model="filterForm.endDate"
        type="date"
        placeholder="结束日期"
        value-format="YYYY-MM-DD"
        style="width: 200px"
      />
      <div class="filter-actions">
        <el-button type="primary" @click="handleSearch">查询</el-button>
        <el-button @click="handleReset">重置</el-button>
      </div>
    </div>

    <!-- 合作列表 -->
    <el-card class="cooperation-card" shadow="never">
      <!-- Tab和操作 -->
      <div class="cooperation-header">
        <el-tabs v-model="activeTab" class="cooperation-tabs" @tab-change="handleTabChange">
          <el-tab-pane label="全部合作" name="all" />
          <el-tab-pane label="合作中" name="active" />
          <el-tab-pane label="已取消" name="cancelled" />
        </el-tabs>
        <div class="header-actions">
          <el-button @click="handleAddCooperation">新增合作</el-button>
          <el-button :icon="Download">导出报表</el-button>
        </div>
      </div>

      <!-- 合作表格 -->
      <el-table :data="cooperationList" style="width: 100%" v-loading="loading">
        <el-table-column prop="date" label="日期" min-width="160" />
        <el-table-column prop="status" label="合作状态" min-width="80">
          <template #default="{ row }">
            <span :class="['status-tag', row.status === '合作中' ? 'active' : 'cancelled']">
              {{ row.status }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="ownerName" label="店主" min-width="100" />
        <el-table-column prop="storeId" label="店铺编号" min-width="120" />
        <el-table-column prop="storeName" label="店铺名称" min-width="200" show-overflow-tooltip />
        <el-table-column prop="operator" label="运营" min-width="100" />
        <el-table-column prop="shareRatio" label="分配百分比（店主/平台/运营）" min-width="200" />
        <el-table-column label="操作" min-width="80" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="handleShowDetail(row)">更多</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>

    <!-- 更多信息弹框 -->
    <el-dialog
      v-model="detailDialogVisible"
      title="更多信息"
      width="640px"
      :close-on-click-modal="false"
    >
      <div class="detail-content">
        <div class="detail-row">
          <span class="detail-label">合作编号：</span>
          <span class="detail-value">{{ currentDetail.cooperationId }}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">创建时间：</span>
          <span class="detail-value">{{ currentDetail.createTime }}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">所属店主：</span>
          <span class="detail-value">{{ currentDetail.ownerName }}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">店铺编号：</span>
          <span class="detail-value">{{ currentDetail.storeId }}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">店铺名称：</span>
          <span class="detail-value">{{ currentDetail.storeName }}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">合作运营：</span>
          <span class="detail-value">{{ currentDetail.operator }}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">开始时间：</span>
          <span class="detail-value">{{ currentDetail.startTime }}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">结束时间：</span>
          <span class="detail-value">{{ currentDetail.endTime }}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">履约方式：</span>
          <span class="detail-value">{{ currentDetail.fulfillmentType }}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">分配比例：</span>
          <span class="detail-value">{{ currentDetail.shareRatioDetail }}</span>
        </div>
        <div class="detail-row">
          <span class="detail-label">电子合约：</span>
          <span class="detail-value">
            {{ currentDetail.contract }}
            <el-link type="primary" style="margin-left: 8px;">点击下载></el-link>
          </span>
        </div>
        <div class="detail-row">
          <span class="detail-label">合作状态：</span>
          <span class="detail-value">
            {{ currentDetail.status }}
            <el-link type="primary" style="margin-left: 16px;" @click="handleCancelCooperation">取消合作></el-link>
          </span>
        </div>
        <div class="detail-row">
          <span class="detail-label">备注信息：</span>
          <span class="detail-value">{{ currentDetail.remark }}</span>
        </div>
      </div>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 新增合作弹框 -->
    <el-dialog
      v-model="addDialogVisible"
      title="新增合作"
      width="700px"
      :close-on-click-modal="false"
    >
      <div class="add-form">
        <div class="form-row">
          <span class="form-label">店铺信息</span>
          <div class="form-content">
            <span class="form-text">店铺编号：SXXXXXX /店铺名称：XXXXXXX/所属店主：张三（输入店铺编号查询）</span>
          </div>
        </div>
        <div class="form-row">
          <span class="form-label">运营信息</span>
          <div class="form-content">
            <span class="form-text">运营编号：S22543252/运营姓名：王五（输入店铺编号查询）</span>
          </div>
        </div>
        <div class="form-row">
          <span class="form-label">开始时间</span>
          <div class="form-content">
            <el-date-picker
              v-model="addForm.startTime"
              type="date"
              placeholder="请选择"
              value-format="YYYY-MM-DD"
              style="width: 180px"
            />
            <span class="form-label" style="margin-left: 24px;">结束时间</span>
            <el-date-picker
              v-model="addForm.endTime"
              type="date"
              placeholder="请选择"
              value-format="YYYY-MM-DD"
              style="width: 180px; margin-left: 12px;"
            />
          </div>
        </div>
        <div class="form-row">
          <span class="form-label">履约方式</span>
          <div class="form-content">
            <el-select v-model="addForm.fulfillmentType" placeholder="请选择" style="width: 180px">
              <el-option label="预付费模式" value="prepaid" />
              <el-option label="后付费模式" value="postpaid" />
            </el-select>
            <span class="form-label" style="margin-left: 24px;">分配比例</span>
            <div class="ratio-inputs">
              <span>店主%</span>
              <span>+</span>
              <span>平台%</span>
              <span>+</span>
              <span>运营%</span>
              <span>=</span>
              <span>100%</span>
            </div>
          </div>
        </div>
        <div class="form-row">
          <span class="form-label">备注信息</span>
          <div class="form-content">
            <el-input
              v-model="addForm.remark"
              type="textarea"
              :rows="4"
              placeholder="请输入"
            />
          </div>
        </div>
        <div class="form-row">
          <span class="form-label">电子合约</span>
          <div class="form-content">
            <el-upload
              action="#"
              :auto-upload="false"
            >
              <template #trigger>
                <div class="upload-trigger">
                  <el-icon><Plus /></el-icon>
                  <span>点击上传</span>
                </div>
              </template>
            </el-upload>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button type="primary" @click="handleConfirmAdd">合作确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Search, Download, Plus } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import { platformCooperationApi } from '@share/api/platform'

interface Cooperation {
  date: string
  status: string
  ownerName: string
  storeId: string
  storeName: string
  operator: string
  shareRatio: string
}

const activeTab = ref('all')
const loading = ref(false)
const detailDialogVisible = ref(false)
const addDialogVisible = ref(false)

const addForm = reactive({
  startTime: '',
  endTime: '',
  fulfillmentType: '',
  remark: ''
})

const currentDetail = reactive({
  cooperationId: '1234567890',
  createTime: '2026-12-12 23:59:59',
  ownerName: '店主名称',
  storeId: 'S1234567890',
  storeName: '示例文字示例文字占位符替换即可',
  operator: '文字占位符',
  startTime: '2025-12-12 23:59:59',
  endTime: '2026-12-12 23:59:59',
  fulfillmentType: '预付费模式',
  shareRatioDetail: '店主佣金比例（1:2）  平台佣金比例（1:2）  运营回款比例（1:2）',
  contract: '示例文字示例文字占位符替换即可示例文字示例文字占位符替换即可',
  status: '合作中',
  remark: '示例文字示例文字占位符替换即可示例文字示例文字占位符替换即可'
})

const filterForm = reactive({
  keyword: '',
  status: '',
  startDate: '2025-09-01',
  endDate: '2025-09-10'
})

const summaryData = reactive({
  total: 12,
  totalPercent: 24,
  active: 40,
  cancelled: 5
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 123
})

const cooperationList = ref<Cooperation[]>([])

const fetchCooperations = async () => {
  loading.value = true
  try {
    const res = await platformCooperationApi.getCooperations({
      page: pagination.page,
      page_size: pagination.pageSize,
      keyword: filterForm.keyword || undefined,
      status: filterForm.status ? parseInt(filterForm.status) : undefined
    })
    if (res.code === 0 && res.data) {
      cooperationList.value = res.data.list.map((item: any) => ({
        date: item.created_at,
        status: item.status === 1 ? '合作中' : item.status === 2 ? '暂停' : '已取消',
        ownerName: item.shop_owner_name || '-',
        storeId: String(item.shop_id),
        storeName: item.shop_name || '-',
        operator: item.operator_name || '-',
        shareRatio: `${item.shop_owner_ratio || 0}:${item.platform_ratio || 0}:${item.operator_ratio || 0}`
      }))
      pagination.total = res.data.total
    }
  } catch (err) {
    console.error('获取合作列表失败:', err)
  } finally {
    loading.value = false
  }
}

const fetchCooperationStats = async () => {
  try {
    const res = await platformCooperationApi.getCooperationStats()
    if (res.code === 0 && res.data) {
      summaryData.total = res.data.total || 0
      summaryData.active = res.data.active || 0
      summaryData.cancelled = res.data.cancelled || 0
    }
  } catch (err) {
    console.error('获取合作统计失败:', err)
  }
}

const handleTabChange = (_tab: string) => {
  pagination.page = 1
  fetchCooperations()
}

const handleSearch = () => {
  pagination.page = 1
  fetchCooperations()
}

const handleReset = () => {
  filterForm.keyword = ''
  filterForm.status = ''
  filterForm.startDate = '2025-09-01'
  filterForm.endDate = '2025-09-10'
  pagination.page = 1
  fetchCooperations()
}

const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  pagination.page = 1
  fetchCooperations()
}

const handlePageChange = (page: number) => {
  pagination.page = page
  fetchCooperations()
}

const handleShowDetail = (row: Cooperation) => {
  currentDetail.cooperationId = '1234567890'
  currentDetail.createTime = row.date
  currentDetail.ownerName = row.ownerName
  currentDetail.storeId = row.storeId
  currentDetail.storeName = row.storeName
  currentDetail.operator = row.operator
  currentDetail.startTime = '2025-12-12 23:59:59'
  currentDetail.endTime = '2026-12-12 23:59:59'
  currentDetail.fulfillmentType = '预付费模式'
  currentDetail.shareRatioDetail = '店主佣金比例（1:2）  平台佣金比例（1:2）  运营回款比例（1:2）'
  currentDetail.contract = '示例文字示例文字占位符替换即可示例文字示例文字占位符替换即可'
  currentDetail.status = row.status
  currentDetail.remark = '示例文字示例文字占位符替换即可示例文字示例文字占位符替换即可'
  detailDialogVisible.value = true
}

const handleAddCooperation = () => {
  addForm.startTime = ''
  addForm.endTime = ''
  addForm.fulfillmentType = ''
  addForm.remark = ''
  addDialogVisible.value = true
}

const handleConfirmAdd = () => {
  // TODO: 提交新增合作
  addDialogVisible.value = false
}

const handleCancelCooperation = () => {
  ElMessageBox.confirm(
    '合作取消后将终止一切事项，确定要取消吗？',
    '合作取消',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(() => {
    // TODO: 执行取消合作操作
    detailDialogVisible.value = false
    ElMessage.success('合作已取消')
    fetchCooperations()
  }).catch(() => {
    // 用户点击取消
  })
}

onMounted(() => {
  fetchCooperationStats()
  fetchCooperations()
})
</script>

<style lang="scss" scoped>
.cooperation-page {
  padding: 20px;
  background-color: #f5f7fa;
  min-height: 100vh;
}

.page-header {
  margin-bottom: 20px;

  .page-title {
    font-size: 20px;
    font-weight: 600;
    color: #303133;
    margin: 0;
  }
}

.cooperation-summary {
  margin-bottom: 20px;

  .summary-title {
    font-size: 14px;
    font-weight: 500;
    color: #606266;
    margin-bottom: 12px;
  }

  .summary-cards {
    display: flex;
    gap: 16px;
  }

  .summary-card {
    flex: 1;
    background: #fff;
    border: 1px solid #ebeef5;
    border-radius: 8px;
    padding: 20px;

    .card-label {
      font-size: 12px;
      color: #909399;
      margin-bottom: 8px;
    }

    .card-value {
      display: flex;
      align-items: baseline;
      gap: 8px;

      .value-number {
        font-size: 28px;
        font-weight: 600;
        color: #303133;
      }

      .value-compare {
        font-size: 12px;
        color: #909399;
      }

      .value-percent {
        font-size: 12px;

        &.positive {
          color: #67c23a;
        }

        &.negative {
          color: #f56c6c;
        }
      }
    }
  }
}

.filter-section {
  display: flex;
  gap: 12px;
  align-items: center;
  margin-bottom: 20px;
  padding: 16px 20px;
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 8px;

  .filter-actions {
    margin-left: auto;
    display: flex;
    gap: 12px;
  }
}

.cooperation-card {
  border-radius: 8px;

  :deep(.el-card__body) {
    padding: 0;
  }
}

.cooperation-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 20px;
  border-bottom: 1px solid #ebeef5;

  .cooperation-tabs {
    :deep(.el-tabs__header) {
      margin: 0;
    }

    :deep(.el-tabs__nav-wrap::after) {
      display: none;
    }
  }

  .header-actions {
    display: flex;
    gap: 12px;
  }
}

.pagination-wrapper {
  padding: 20px;
  display: flex;
  justify-content: center;
}

.status-tag {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;

  &.active {
    background: #e1f3d8;
    color: #67c23a;
  }

  &.cancelled {
    background: #fde2e2;
    color: #f56c6c;
  }
}

@media (max-width: 1200px) {
  .cooperation-summary .summary-cards {
    flex-wrap: wrap;
  }

  .cooperation-summary .summary-card {
    flex: 0 0 calc(50% - 8px);
  }
}

@media (max-width: 768px) {
  .cooperation-page {
    padding: 12px;
  }

  .cooperation-summary .summary-card {
    flex: 0 0 100%;
  }

  .filter-section {
    flex-wrap: wrap;
  }

  .cooperation-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
    padding: 12px 16px;
  }
}

.detail-content {
  .detail-row {
    display: flex;
    padding: 12px 0;
    border-bottom: 1px solid #f5f7fa;

    &:last-child {
      border-bottom: none;
    }

    .detail-label {
      width: 80px;
      flex-shrink: 0;
      font-size: 14px;
      color: #909399;
    }

    .detail-value {
      flex: 1;
      font-size: 14px;
      color: #303133;
    }
  }
}

.add-form {
  .form-row {
    display: flex;
    align-items: flex-start;
    padding: 16px 0;
    border-bottom: 1px solid #f5f7fa;

    &:last-child {
      border-bottom: none;
    }

    .form-label {
      width: 70px;
      flex-shrink: 0;
      font-size: 14px;
      color: #606266;
      line-height: 32px;
    }

    .form-content {
      flex: 1;
      display: flex;
      align-items: center;
      flex-wrap: wrap;
      gap: 8px;

      .form-text {
        font-size: 14px;
        color: #909399;
        line-height: 32px;
      }
    }

    .ratio-inputs {
      display: flex;
      align-items: center;
      gap: 8px;
      font-size: 14px;
      color: #606266;
      margin-left: 12px;
    }
  }

  .upload-trigger {
    display: flex;
    align-items: center;
    gap: 8px;
    color: #909399;
    cursor: pointer;

    &:hover {
      color: #ff6a3a;
    }
  }
}
</style>
