<template>
  <div class="violation-page">
    <!-- 违规概览 -->
    <div class="violation-summary">
      <div class="summary-cards">
        <div class="summary-card">
          <div class="card-content">
            <div class="card-info">
              <div class="card-label">罚款笔数</div>
              <div class="card-value">
                <span class="value-number">{{ summaryData.fineCount }}</span>
                <span class="value-compare">vs上周 {{ summaryData.fineLastWeek }}</span>
              </div>
            </div>
            <el-button type="primary" size="small" @click="handleNewFine">新建罚款</el-button>
          </div>
        </div>
        <div class="summary-card">
          <div class="card-content">
            <div class="card-info">
              <div class="card-label">补贴笔数</div>
              <div class="card-value">
                <span class="value-number">{{ summaryData.subsidyCount }}</span>
                <span class="value-compare">vs上周 {{ summaryData.subsidyLastWeek }}</span>
              </div>
            </div>
            <el-button type="primary" size="small" @click="handleNewSubsidy">新建补贴</el-button>
          </div>
        </div>
      </div>
    </div>

    <!-- Tab切换和筛选 -->
    <div class="filter-section">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="处罚列表" name="fine" />
        <el-tab-pane label="补贴列表" name="subsidy" />
      </el-tabs>
      <div class="filter-right">
        <el-input
          v-model="filterForm.keyword"
          placeholder="快速搜索"
          :prefix-icon="Search"
          clearable
          style="width: 200px"
        />
        <el-date-picker
          v-model="filterForm.startDate"
          type="date"
          placeholder="开始日期"
          value-format="YYYY-MM-DD"
          style="width: 140px"
        />
        <span style="color: #909399;">-</span>
        <el-date-picker
          v-model="filterForm.endDate"
          type="date"
          placeholder="结束日期"
          value-format="YYYY-MM-DD"
          style="width: 140px"
        />
        <el-button :icon="Download">导出报表</el-button>
      </div>
    </div>

    <!-- 数据表格 -->
    <el-card class="violation-card" shadow="never">
      <el-table :data="violationList" style="width: 100%" v-loading="loading">
        <el-table-column prop="date" label="日期" min-width="160" />
        <el-table-column prop="role" label="角色" min-width="80" />
        <el-table-column prop="name" label="姓名" min-width="80" />
        <el-table-column prop="transactionType" label="交易类型" min-width="100" />
        <el-table-column prop="transactionChannel" label="交易渠道" min-width="100" />
        <el-table-column prop="transactionNo" label="交易单号" min-width="140" />
        <el-table-column prop="amount" label="交易金额" min-width="130">
          <template #default="{ row }">
            <span class="amount-text">NT${{ row.amount }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" min-width="80">
          <template #default="{ row }">
            <span class="status-text">{{ row.status }}</span>
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

    <!-- 新建罚款弹框 -->
    <el-dialog
      v-model="fineDialogVisible"
      title="新建罚款"
      width="600px"
      :close-on-click-modal="false"
    >
      <div class="fine-form">
        <div class="form-section">
          <div class="section-title">扣款方/账户余额</div>
          <el-select v-model="fineForm.deductAccount" placeholder="选择或输入扣款方角色/姓名/账户余额" style="width: 100%" filterable>
            <el-option label="店主A / 保证金余额 NT$10,000" value="店主A-保证金" />
            <el-option label="店主B / 保证金余额 NT$8,000" value="店主B-保证金" />
            <el-option label="运营A / 保证金余额 NT$5,000" value="运营A-保证金" />
          </el-select>
        </div>
        <div class="form-row">
          <div class="form-item">
            <span class="form-label">罚款金额</span>
            <el-input v-model="fineForm.amount" placeholder="请输入" />
          </div>
        </div>
        <div class="form-row">
          <div class="form-item full-width">
            <span class="form-label">罚款描述</span>
            <el-input v-model="fineForm.description" type="textarea" :rows="2" placeholder="请输入" />
          </div>
        </div>
        <div class="form-row">
          <div class="form-item full-width">
            <span class="form-label">规则依据</span>
            <el-input v-model="fineForm.ruleReference" placeholder="请输入" />
          </div>
        </div>
        <div class="form-row">
          <div class="form-item full-width">
            <span class="form-label">证据材料</span>
            <div class="form-value">
              <el-upload
                action="#"
                :auto-upload="false"
              >
                <template #trigger>
                  <div class="upload-trigger">
                    <el-icon><Plus /></el-icon>
                  </div>
                </template>
              </el-upload>
            </div>
          </div>
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-checkbox v-model="fineForm.notifyUser">通知用户</el-checkbox>
          <el-button type="primary" @click="handleSubmitFine">执行</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 新建补贴弹框 -->
    <el-dialog
      v-model="subsidyDialogVisible"
      title="新建补贴"
      width="600px"
      :close-on-click-modal="false"
    >
      <div class="fine-form">
        <div class="form-section">
          <div class="section-title">收款方/账户余额</div>
          <el-select v-model="subsidyForm.receiveAccount" placeholder="选择或输入收款方角色/姓名/账户余额" style="width: 100%" filterable>
            <el-option label="店主A / 保证金余额 NT$10,000" value="店主A-保证金" />
            <el-option label="店主B / 保证金余额 NT$8,000" value="店主B-保证金" />
            <el-option label="运营A / 保证金余额 NT$5,000" value="运营A-保证金" />
          </el-select>
        </div>
        <div class="form-row">
          <div class="form-item">
            <span class="form-label">补贴金额</span>
            <el-input v-model="subsidyForm.amount" placeholder="请输入" />
          </div>
        </div>
        <div class="form-row">
          <div class="form-item full-width">
            <span class="form-label">补贴描述</span>
            <el-input v-model="subsidyForm.description" type="textarea" :rows="2" placeholder="请输入" />
          </div>
        </div>
        <div class="form-row">
          <div class="form-item full-width">
            <span class="form-label">规则依据</span>
            <el-input v-model="subsidyForm.ruleReference" placeholder="请输入" />
          </div>
        </div>
        <div class="form-row">
          <div class="form-item full-width">
            <span class="form-label">证据材料</span>
            <div class="form-value">
              <el-upload
                action="#"
                :auto-upload="false"
              >
                <template #trigger>
                  <div class="upload-trigger">
                    <el-icon><Plus /></el-icon>
                  </div>
                </template>
              </el-upload>
            </div>
          </div>
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-checkbox v-model="subsidyForm.notifyUser">通知用户</el-checkbox>
          <el-button type="primary" @click="handleSubmitSubsidy">执行</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Search, Download, Plus } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

interface ViolationRecord {
  date: string
  role: string
  name: string
  transactionType: string
  transactionChannel: string
  transactionNo: string
  amount: string
  status: string
}

const activeTab = ref('fine')
const loading = ref(false)
const fineDialogVisible = ref(false)
const subsidyDialogVisible = ref(false)

const summaryData = reactive({
  fineCount: 12,
  fineLastWeek: 8,
  subsidyCount: 8,
  subsidyLastWeek: 7
})

const filterForm = reactive({
  keyword: '',
  startDate: '2025-09-01',
  endDate: '2025-09-10'
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 123
})

const fineForm = reactive({
  deductAccount: '',
  amount: '',
  description: '',
  ruleReference: '',
  notifyUser: true
})

const subsidyForm = reactive({
  receiveAccount: '',
  amount: '',
  description: '',
  ruleReference: '',
  notifyUser: true
})

const violationList = ref<ViolationRecord[]>([])

const fetchViolationList = async () => {
  loading.value = true
  try {
    const roles = ['店主', '运营']
    const types = activeTab.value === 'fine' ? ['罚款', '罚款'] : ['补贴', '补贴']
    const channels = ['店主保证金', '运营保证金']
    violationList.value = Array.from({ length: 10 }, (_, i) => ({
      date: '2026-12-12 23:59:59',
      role: roles[i % 2],
      name: '占位符',
      transactionType: types[i % 2],
      transactionChannel: channels[i % 2],
      transactionNo: '3456789456789',
      amount: '223,560.50',
      status: '已完成'
    }))
    pagination.total = 123
  } catch (err) {
    console.error('获取违规列表失败:', err)
  } finally {
    loading.value = false
  }
}

const handleTabChange = () => {
  pagination.page = 1
  fetchViolationList()
}

const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  pagination.page = 1
  fetchViolationList()
}

const handlePageChange = (page: number) => {
  pagination.page = page
  fetchViolationList()
}

const handleNewFine = () => {
  fineForm.deductAccount = ''
  fineForm.amount = ''
  fineForm.description = ''
  fineForm.ruleReference = ''
  fineForm.notifyUser = true
  fineDialogVisible.value = true
}

const handleNewSubsidy = () => {
  subsidyForm.receiveAccount = ''
  subsidyForm.amount = ''
  subsidyForm.description = ''
  subsidyForm.ruleReference = ''
  subsidyForm.notifyUser = true
  subsidyDialogVisible.value = true
}

const handleSubmitFine = () => {
  fineDialogVisible.value = false
  ElMessage.success('罚款创建成功')
  fetchViolationList()
}

const handleSubmitSubsidy = () => {
  subsidyDialogVisible.value = false
  ElMessage.success('补贴创建成功')
  fetchViolationList()
}

onMounted(() => {
  fetchViolationList()
})
</script>

<style scoped lang="scss">
.violation-page {
  padding: 20px;
  background: #f5f7fa;
  min-height: 100%;
}

.violation-summary {
  margin-bottom: 20px;

  .summary-cards {
    display: flex;
    gap: 16px;
  }

  .summary-card {
    flex: 1;
    background: #fff;
    border-radius: 8px;
    padding: 20px 24px;
    border: 1px solid #ebeef5;

    .card-content {
      display: flex;
      justify-content: space-between;
      align-items: center;
    }

    .card-info {
      .card-label {
        font-size: 14px;
        color: #909399;
        margin-bottom: 8px;
      }

      .card-value {
        display: flex;
        align-items: baseline;
        gap: 12px;

        .value-number {
          font-size: 32px;
          font-weight: 600;
          color: #303133;
        }

        .value-compare {
          font-size: 14px;
          color: #909399;
        }
      }
    }
  }
}

.filter-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 0 20px;
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 8px;

  :deep(.el-tabs) {
    .el-tabs__header {
      margin-bottom: 0;
    }

    .el-tabs__nav-wrap::after {
      display: none;
    }
  }

  .filter-right {
    display: flex;
    align-items: center;
    gap: 12px;
  }
}

.violation-card {
  border-radius: 8px;

  :deep(.el-card__body) {
    padding: 0;
  }
}

.pagination-wrapper {
  padding: 20px;
  display: flex;
  justify-content: center;
}

.amount-text {
  color: #303133;
  font-weight: 500;
}

.status-text {
  color: #67c23a;
}

.fine-form {
  .form-section {
    margin-bottom: 20px;

    .section-title {
      font-size: 14px;
      color: #606266;
      margin-bottom: 8px;
    }
  }

  .form-row {
    display: flex;
    gap: 20px;
    margin-bottom: 16px;

    .form-item {
      display: flex;
      align-items: flex-start;
      gap: 12px;

      &.full-width {
        flex: 1;
      }

      .form-label {
        font-size: 14px;
        color: #606266;
        white-space: nowrap;
        min-width: 70px;
        line-height: 32px;
      }

      .form-value {
        flex: 1;
      }
    }
  }
}

.upload-trigger {
  width: 60px;
  height: 60px;
  border: 1px dashed #d9d9d9;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: #909399;

  &:hover {
    border-color: #409eff;
    color: #409eff;
  }
}

.dialog-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}
</style>
