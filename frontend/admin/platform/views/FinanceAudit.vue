<template>
  <div class="finance-audit-page">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="page-title">财务审核</h1>
    </div>

    <!-- 顶部Tab -->
    <el-tabs v-model="activeMainTab" class="main-tabs" @tab-change="handleMainTabChange">
      <el-tab-pane label="提现审核" name="withdraw" />
      <el-tab-pane label="充值审核" name="recharge" />
      <el-tab-pane label="转存审核" name="transfer" />
    </el-tabs>

    <!-- 概览 -->
    <div class="audit-summary">
      <div class="summary-title">概览</div>
      <div class="summary-cards">
        <div class="summary-card">
          <div class="card-label">待审核笔数</div>
          <div class="card-value">{{ summaryData.pending }}</div>
        </div>
        <div class="summary-card">
          <div class="card-label">已审核笔数</div>
          <div class="card-value">{{ summaryData.approved }}</div>
        </div>
        <div class="summary-card">
          <div class="card-label">平均处理时效</div>
          <div class="card-value">
            <span class="time-value">{{ summaryData.avgTime }}</span>
            <span class="time-compare">VS昨天 12:00</span>
            <span class="time-percent positive">↑{{ summaryData.timePercent }}%</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 筛选区域 - 提现审核 -->
    <div class="filter-section" v-if="activeMainTab === 'withdraw'">
      <el-input
        v-model="filterForm.keyword"
        placeholder="店主/运营姓名"
        :prefix-icon="Search"
        clearable
        style="width: 180px"
      />
      <el-select v-model="filterForm.withdrawType" placeholder="选择提现类型" clearable style="width: 150px">
        <el-option label="全部" value="" />
        <el-option label="佣金" value="commission" />
        <el-option label="店主预付款" value="prepaid" />
        <el-option label="保证金" value="deposit" />
        <el-option label="运营回款" value="operator" />
      </el-select>
      <el-select v-model="filterForm.auditStatus" placeholder="选择审批状态" clearable style="width: 150px">
        <el-option label="全部" value="" />
        <el-option label="待审批" value="pending" />
        <el-option label="已审批" value="approved" />
      </el-select>
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
      <div class="filter-actions">
        <el-button type="primary" @click="handleSearch">查询</el-button>
        <el-button @click="handleReset">重置</el-button>
      </div>
    </div>

    <!-- 筛选区域 - 充值审核 -->
    <div class="filter-section" v-if="activeMainTab === 'recharge'">
      <el-input
        v-model="filterForm.keyword"
        placeholder="店主/运营姓名"
        :prefix-icon="Search"
        clearable
        style="width: 150px"
      />
      <el-input
        v-model="filterForm.transactionNo"
        placeholder="输入交易单号"
        :prefix-icon="Search"
        clearable
        style="width: 150px"
      />
      <el-select v-model="filterForm.rechargeType" placeholder="选择充值类型" clearable style="width: 150px">
        <el-option label="全部" value="" />
        <el-option label="店主预付款" value="prepaid" />
        <el-option label="店主保证金" value="deposit" />
        <el-option label="运营保证金" value="operator_deposit" />
      </el-select>
      <el-select v-model="filterForm.auditStatus" placeholder="选择审批状态" clearable style="width: 150px">
        <el-option label="全部" value="" />
        <el-option label="待审批" value="pending" />
        <el-option label="已审批" value="approved" />
      </el-select>
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
      <div class="filter-actions">
        <el-button type="primary" @click="handleSearch">查询</el-button>
        <el-button @click="handleReset">重置</el-button>
      </div>
    </div>

    <!-- 筛选区域 - 转存审核 -->
    <div class="filter-section" v-if="activeMainTab === 'transfer'">
      <el-input
        v-model="filterForm.keyword"
        placeholder="店主姓名"
        :prefix-icon="Search"
        clearable
        style="width: 150px"
      />
      <el-input
        v-model="filterForm.transactionNo"
        placeholder="输入交易单号"
        :prefix-icon="Search"
        clearable
        style="width: 150px"
      />
      <el-select v-model="filterForm.auditStatus" placeholder="选择审批状态" clearable style="width: 150px">
        <el-option label="全部" value="" />
        <el-option label="待审批" value="pending" />
        <el-option label="已审批" value="approved" />
      </el-select>
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
      <div class="filter-actions">
        <el-button type="primary" @click="handleSearch">查询</el-button>
        <el-button @click="handleReset">重置</el-button>
      </div>
    </div>

    <!-- 审核列表 -->
    <el-card class="audit-card" shadow="never">
      <!-- 标题和操作 -->
      <div class="audit-header">
        <span class="header-title">全部明细</span>
        <div class="header-actions">
          <el-button :icon="Download">导出报表</el-button>
        </div>
      </div>

      <!-- 审核表格 - 提现审核 -->
      <el-table v-if="activeMainTab === 'withdraw'" :data="auditList" style="width: 100%" v-loading="loading">
        <el-table-column prop="date" label="日期" min-width="160" />
        <el-table-column prop="roleType" label="角色类型" min-width="100" />
        <el-table-column prop="name" label="姓名" min-width="100" />
        <el-table-column prop="withdrawType" label="提现类型" min-width="120" />
        <el-table-column prop="account" label="收款账户" min-width="140" />
        <el-table-column prop="amount" label="金额" min-width="130">
          <template #default="{ row }">
            <span class="amount-text">NT${{ row.amount }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="审批状态" min-width="100">
          <template #default="{ row }">
            <span :class="['status-tag', row.status === '待审批' ? 'pending' : row.status === '已拒绝' ? 'rejected' : 'approved']">
              {{ row.status }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="操作" min-width="80" fixed="right">
          <template #default="{ row }">
            <el-button 
              v-if="row.status === '待审批'" 
              type="primary" 
              link 
              @click="handleAudit(row)"
            >
              审批
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 审核表格 - 充值审核 -->
      <el-table v-if="activeMainTab === 'recharge'" :data="auditList" style="width: 100%" v-loading="loading">
        <el-table-column prop="date" label="日期" min-width="160" />
        <el-table-column prop="roleType" label="角色类型" min-width="100" />
        <el-table-column prop="name" label="姓名" min-width="100" />
        <el-table-column prop="rechargeType" label="充值类型" min-width="120" />
        <el-table-column prop="transactionChannel" label="交易渠道" min-width="100" />
        <el-table-column prop="transactionNo" label="交易单号" min-width="140" />
        <el-table-column prop="amount" label="金额" min-width="130">
          <template #default="{ row }">
            <span class="amount-text">NT${{ row.amount }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="审批状态" min-width="100">
          <template #default="{ row }">
            <span :class="['status-tag', row.status === '待审批' ? 'pending' : row.status === '已拒绝' ? 'rejected' : 'approved']">
              {{ row.status }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="操作" min-width="80" fixed="right">
          <template #default="{ row }">
            <el-button 
              v-if="row.status === '待审批'" 
              type="primary" 
              link 
              @click="handleAudit(row)"
            >
              审批
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 审核表格 - 转存审核 -->
      <el-table v-if="activeMainTab === 'transfer'" :data="auditList" style="width: 100%" v-loading="loading">
        <el-table-column prop="date" label="日期" min-width="160" />
        <el-table-column prop="roleType" label="角色类型" min-width="100" />
        <el-table-column prop="name" label="姓名" min-width="100" />
        <el-table-column prop="transferType" label="交易类型" min-width="100" />
        <el-table-column prop="transferChannel" label="交易渠道" min-width="110" />
        <el-table-column prop="transactionNo" label="交易单号" min-width="140" />
        <el-table-column prop="amount" label="金额" min-width="130">
          <template #default="{ row }">
            <span class="amount-text">NT${{ row.amount }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="审批状态" min-width="100">
          <template #default="{ row }">
            <span :class="['status-tag', row.status === '待审批' ? 'pending' : row.status === '已拒绝' ? 'rejected' : 'approved']">
              {{ row.status }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="操作" min-width="80" fixed="right">
          <template #default="{ row }">
            <el-button 
              v-if="row.status === '待审批'" 
              type="primary" 
              link 
              @click="handleAudit(row)"
            >
              审批
            </el-button>
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

    <!-- 审批弹框 - 提现审核 -->
    <el-dialog
      v-model="auditDialogVisible"
      title="审批"
      width="900px"
      :close-on-click-modal="false"
      v-if="activeMainTab === 'withdraw'"
    >
      <!-- 审批信息表格 -->
      <el-table :data="[currentAuditRecord]" style="width: 100%; margin-bottom: 24px;" border>
        <el-table-column prop="date" label="日期" min-width="160" />
        <el-table-column prop="roleType" label="角色类型" min-width="100" />
        <el-table-column prop="name" label="姓名" min-width="100" />
        <el-table-column prop="withdrawType" label="提现类型" min-width="120" />
        <el-table-column prop="account" label="收款账户" min-width="140" />
        <el-table-column prop="amount" label="金额" min-width="130">
          <template #default="{ row }">
            <span>NT${{ row.amount }}</span>
          </template>
        </el-table-column>
      </el-table>

      <!-- 财务审核表单 -->
      <div class="audit-form">
        <div class="form-title">财务审核</div>
        <div class="form-row">
          <div class="form-item">
            <span class="form-label">付款渠道</span>
            <span class="form-value">{{ auditForm.payChannel }}</span>
          </div>
          <div class="form-item">
            <span class="form-label">付款单号</span>
            <span class="form-value">{{ auditForm.payNo }}</span>
          </div>
          <div class="form-item">
            <span class="form-label">经办人</span>
            <span class="form-value">{{ auditForm.handler }}</span>
          </div>
        </div>
        <div class="form-row">
          <div class="form-item full-width">
            <span class="form-label">付款凭证</span>
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
        <div class="form-row">
          <div class="form-item full-width">
            <span class="form-label">备注</span>
            <div class="form-value">
              <el-input
                v-model="auditForm.remark"
                type="textarea"
                :rows="3"
                placeholder="请输入"
              />
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <el-button @click="handleReject">拒绝</el-button>
        <el-button type="primary" @click="handleApprove">通过</el-button>
      </template>
    </el-dialog>

    <!-- 审批弹框 - 充值审核 -->
    <el-dialog
      v-model="auditDialogVisible"
      title="审批"
      width="900px"
      :close-on-click-modal="false"
      v-if="activeMainTab === 'recharge'"
    >
      <!-- 审批信息表格 -->
      <el-table :data="[currentAuditRecord]" style="width: 100%; margin-bottom: 24px;" border>
        <el-table-column prop="date" label="日期" min-width="140" />
        <el-table-column prop="roleType" label="角色类型" min-width="90" />
        <el-table-column prop="name" label="姓名" min-width="80" />
        <el-table-column prop="transactionType" label="交易类型" min-width="100" />
        <el-table-column prop="transactionChannel" label="交易渠道" min-width="90" />
        <el-table-column prop="transactionNo" label="交易单号" min-width="120" />
        <el-table-column prop="amount" label="金额" min-width="120">
          <template #default="{ row }">
            <span>NT${{ row.amount }}</span>
          </template>
        </el-table-column>
      </el-table>

      <!-- 财务审核表单 -->
      <div class="audit-form">
        <div class="form-title">财务审核</div>
        <div class="form-row">
          <div class="form-item">
            <span class="form-label">收款渠道</span>
            <span class="form-value">{{ auditForm.receiveChannel }}</span>
          </div>
          <div class="form-item">
            <span class="form-label">收款单号</span>
            <span class="form-value">{{ auditForm.receiveNo }}</span>
          </div>
          <div class="form-item">
            <span class="form-label">经办人</span>
            <span class="form-value">{{ auditForm.handler }}</span>
          </div>
        </div>
        <div class="form-row">
          <div class="form-item full-width">
            <span class="form-label">收款凭证</span>
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
        <div class="form-row">
          <div class="form-item full-width">
            <span class="form-label">备注</span>
            <div class="form-value">
              <el-input
                v-model="auditForm.remark"
                type="textarea"
                :rows="3"
                placeholder="请输入"
              />
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <el-button @click="handleReject">拒绝</el-button>
        <el-button type="primary" @click="handleApprove">通过</el-button>
      </template>
    </el-dialog>

    <!-- 审批弹框 - 转存审核 -->
    <el-dialog
      v-model="auditDialogVisible"
      title="审批"
      width="900px"
      :close-on-click-modal="false"
      v-if="activeMainTab === 'transfer'"
    >
      <!-- 审批信息表格 -->
      <el-table :data="[currentAuditRecord]" style="width: 100%; margin-bottom: 24px;" border>
        <el-table-column prop="date" label="日期" min-width="160" />
        <el-table-column prop="roleType" label="角色类型" min-width="100" />
        <el-table-column prop="name" label="姓名" min-width="80" />
        <el-table-column prop="transferType" label="交易类型" min-width="100" />
        <el-table-column prop="transactionNo" label="交易单号" min-width="140" />
        <el-table-column prop="amount" label="金额" min-width="130">
          <template #default="{ row }">
            <span>NT${{ row.amount }}</span>
          </template>
        </el-table-column>
      </el-table>

      <!-- 财务审核表单 -->
      <div class="audit-form">
        <div class="form-title">财务审核</div>
        <div class="form-row">
          <div class="form-item">
            <span class="form-label">付款渠道</span>
            <span class="form-value">{{ currentAuditRecord.transferChannel }}</span>
          </div>
          <div class="form-item">
            <span class="form-label">付款单号</span>
            <span class="form-value">{{ auditForm.payNo }}</span>
          </div>
          <div class="form-item">
            <span class="form-label">经办人</span>
            <span class="form-value">{{ auditForm.handler }}</span>
          </div>
        </div>
        <div class="form-row">
          <div class="form-item full-width">
            <span class="form-label">付款凭证</span>
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
        <div class="form-row">
          <div class="form-item full-width">
            <span class="form-label">备注</span>
            <div class="form-value">
              <el-input
                v-model="auditForm.remark"
                type="textarea"
                :rows="3"
                placeholder="请输入"
              />
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <el-button @click="handleReject">拒绝</el-button>
        <el-button type="primary" @click="handleApprove">通过</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Search, Download, Plus } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { platformFinanceAuditApi } from '@share/api/platform'

interface AuditRecord {
  date: string
  roleType: string
  name: string
  withdrawType: string
  account: string
  amount: string
  status: string
  transactionType: string
  transactionChannel: string
  transactionNo: string
  rechargeType: string
  transferType: string
  transferChannel: string
}

const activeMainTab = ref('withdraw')
const loading = ref(false)
const auditDialogVisible = ref(false)

const currentAuditRecord = reactive<AuditRecord>({
  date: '',
  roleType: '',
  name: '',
  withdrawType: '',
  account: '',
  amount: '',
  status: '',
  transactionType: '',
  transactionChannel: '',
  transactionNo: '',
  rechargeType: '',
  transferType: '',
  transferChannel: ''
})

const auditForm = reactive({
  payChannel: '文字占位符占位符',
  payNo: '文字占位符占位符',
  handler: '文字占位符默认样式不可编辑',
  remark: '',
  receiveChannel: '文字占位符占位符',
  receiveNo: '文字占位符占位符'
})

const filterForm = reactive({
  keyword: '',
  withdrawType: '',
  auditStatus: '',
  startDate: '2025-09-01',
  endDate: '2025-09-10',
  transactionNo: '',
  rechargeType: ''
})

const summaryData = reactive({
  pending: 12,
  approved: 40,
  avgTime: '5:00',
  timePercent: 36.8
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 123
})

const auditList = ref<AuditRecord[]>([])

const fetchAuditStats = async () => {
  try {
    const res = await platformFinanceAuditApi.getAuditStats(activeMainTab.value)
    if (res.code === 0 && res.data) {
      summaryData.pending = res.data.pending || 0
      summaryData.approved = res.data.approved || 0
      summaryData.avgTime = res.data.avg_time || '0:00'
      summaryData.timePercent = res.data.time_percent || 0
    }
  } catch (err) {
    console.error('获取审核统计失败:', err)
  }
}

const fetchAuditList = async () => {
  loading.value = true
  try {
    let res
    if (activeMainTab.value === 'withdraw') {
      res = await platformFinanceAuditApi.getWithdrawAuditList({
        status: filterForm.auditStatus || undefined,
        keyword: filterForm.keyword || undefined,
        withdraw_type: filterForm.withdrawType || undefined,
        page: pagination.page,
        page_size: pagination.pageSize
      })
    } else if (activeMainTab.value === 'recharge') {
      res = await platformFinanceAuditApi.getRechargeAuditList({
        status: filterForm.auditStatus || undefined,
        page: pagination.page,
        page_size: pagination.pageSize
      })
    } else {
      // 转存审核暂无
      auditList.value = []
      pagination.total = 0
      loading.value = false
      return
    }

    if (res && res.code === 0 && res.data) {
      auditList.value = res.data.list.map((item: any) => ({
        id: item.id,
        date: item.created_at,
        roleType: item.account_type === 'prepayment' ? '店主' : item.account_type === 'operator' ? '运营' : '平台',
        name: item.username || '-',
        withdrawType: item.account_type === 'prepayment' ? '店主预付款' : item.account_type === 'deposit' ? '保证金' : '运营回款',
        account: '-',
        amount: item.amount || '0',
        status: item.status_text || (item.status === 0 ? '待审批' : item.status === 1 ? '已审批' : '已拒绝'),
        transactionType: item.transaction_type || '-',
        transactionChannel: '-',
        transactionNo: item.transaction_no || '-',
        rechargeType: item.account_type === 'prepayment' ? '店主预付款' : item.account_type === 'deposit' ? '店主保证金' : '运营保证金',
        transferType: '-',
        transferChannel: '-'
      }))
      pagination.total = res.data.total
    }
  } catch (err) {
    console.error('获取审核列表失败:', err)
  } finally {
    loading.value = false
  }
}

const handleMainTabChange = (_tab: string) => {
  pagination.page = 1
  fetchAuditList()
}

const handleSearch = () => {
  pagination.page = 1
  fetchAuditList()
}

const handleReset = () => {
  filterForm.keyword = ''
  filterForm.withdrawType = ''
  filterForm.auditStatus = ''
  filterForm.startDate = '2025-09-01'
  filterForm.endDate = '2025-09-10'
  filterForm.transactionNo = ''
  filterForm.rechargeType = ''
  pagination.page = 1
  fetchAuditList()
}

const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  pagination.page = 1
  fetchAuditList()
}

const handlePageChange = (page: number) => {
  pagination.page = page
  fetchAuditList()
}

const handleAudit = (row: AuditRecord) => {
  currentAuditRecord.date = row.date
  currentAuditRecord.roleType = row.roleType
  currentAuditRecord.name = row.name
  currentAuditRecord.withdrawType = row.withdrawType
  currentAuditRecord.account = row.account
  currentAuditRecord.amount = row.amount
  currentAuditRecord.status = row.status
  currentAuditRecord.transactionType = row.transactionType
  currentAuditRecord.transactionChannel = row.transactionChannel
  currentAuditRecord.transactionNo = row.transactionNo
  currentAuditRecord.rechargeType = row.rechargeType
  currentAuditRecord.transferType = row.transferType
  currentAuditRecord.transferChannel = row.transferChannel
  auditForm.remark = ''
  auditDialogVisible.value = true
}

const handleReject = () => {
  // TODO: 执行拒绝操作
  auditDialogVisible.value = false
  ElMessage.warning('已拒绝')
  fetchAuditList()
}

const handleApprove = () => {
  // TODO: 执行通过操作
  auditDialogVisible.value = false
  ElMessage.success('审批通过')
  fetchAuditList()
}

onMounted(() => {
  fetchAuditStats()
  fetchAuditList()
})
</script>

<style lang="scss" scoped>
.finance-audit-page {
  padding: 20px;
  background-color: #f5f7fa;
  min-height: 100vh;
}

.page-header {
  margin-bottom: 16px;

  .page-title {
    font-size: 20px;
    font-weight: 600;
    color: #303133;
    margin: 0;
  }
}

.main-tabs {
  margin-bottom: 20px;

  :deep(.el-tabs__header) {
    margin: 0;
    background: transparent;
  }

  :deep(.el-tabs__nav-wrap::after) {
    display: none;
  }
}

.audit-summary {
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
      font-size: 28px;
      font-weight: 600;
      color: #303133;

      .time-value {
        font-size: 28px;
      }

      .time-compare {
        font-size: 12px;
        color: #909399;
        font-weight: normal;
      }

      .time-percent {
        font-size: 12px;
        font-weight: normal;

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

.audit-card {
  border-radius: 8px;

  :deep(.el-card__body) {
    padding: 0;
  }
}

.audit-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #ebeef5;

  .header-title {
    font-size: 14px;
    font-weight: 500;
    color: #303133;
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

.amount-text {
  color: #303133;
  font-weight: 500;
}

.status-tag {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;

  &.pending {
    background: #fef0f0;
    color: #f56c6c;
  }

  &.approved {
    background: #f0f9eb;
    color: #67c23a;
  }

  &.rejected {
    background: #909399;
    color: #fff;
  }
}

@media (max-width: 1200px) {
  .audit-summary .summary-cards {
    flex-wrap: wrap;
  }

  .audit-summary .summary-card {
    flex: 0 0 calc(50% - 8px);
  }
}

@media (max-width: 768px) {
  .finance-audit-page {
    padding: 12px;
  }

  .audit-summary .summary-card {
    flex: 0 0 100%;
  }

  .filter-section {
    flex-wrap: wrap;
  }
}

.audit-form {
  .form-title {
    font-size: 14px;
    font-weight: 500;
    color: #303133;
    margin-bottom: 16px;
  }

  .form-row {
    display: flex;
    gap: 24px;
    margin-bottom: 16px;

    &:last-child {
      margin-bottom: 0;
    }
  }

  .form-item {
    display: flex;
    align-items: flex-start;
    gap: 12px;

    &.full-width {
      flex: 1;
    }

    .form-label {
      width: 60px;
      flex-shrink: 0;
      font-size: 14px;
      color: #606266;
      line-height: 32px;
    }

    .form-value {
      flex: 1;
      font-size: 14px;
      color: #303133;
      line-height: 32px;
    }
  }

  .upload-trigger {
    width: 60px;
    height: 60px;
    border: 1px dashed #dcdfe6;
    border-radius: 4px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    color: #909399;

    &:hover {
      border-color: #ff6a3a;
      color: #ff6a3a;
    }
  }
}
</style>
