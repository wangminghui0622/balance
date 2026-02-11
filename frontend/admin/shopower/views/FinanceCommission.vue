<template>
	<div class="finance-commission-page">
		<!-- 页面标题 -->
		<div class="page-header">
			<h1 class="page-title">我的佣金</h1>
		</div>

		<!-- 佣金总览 -->
		<el-card class="summary-card">
			<div class="summary-header">
				<span class="summary-title">佣金总览</span>
				<a href="#" class="stats-link">佣金统计</a>
			</div>
			<el-row :gutter="20" class="summary-content">
				<el-col :xs="24" :sm="12" :md="12" :lg="12" :xl="12">
					<div class="balance-card">
						<div class="balance-info">
							<div class="balance-label">可提领金额</div>
							<div class="balance-value">NT${{ formatAmount(summaryData.availableAmount) }}</div>
							<div class="balance-subtitle">累计佣金: NT${{ formatAmount(summaryData.totalCommission) }}</div>
						</div>
						<div class="balance-actions">
							<el-button type="primary" @click="handleWithdraw">提现</el-button>
							<el-button @click="handleTransfer">转存</el-button>
						</div>
					</div>
				</el-col>
				<el-col :xs="24" :sm="12" :md="12" :lg="12" :xl="12">
					<div class="pending-card">
						<div class="pending-label">即将结算佣金</div>
						<div class="pending-value">NT${{ formatAmount(summaryData.pendingAmount) }}</div>
						<div class="pending-subtitle">根据未结算订单预估即将结算金额</div>
					</div>
				</el-col>
			</el-row>
		</el-card>

		<!-- 佣金列表 -->
		<el-card class="list-card">
			<div class="tabs-header">
				<el-tabs v-model="activeTab" class="commission-tabs">
					<el-tab-pane label="全部" name="all" />
					<el-tab-pane label="佣金" name="commission" />
					<el-tab-pane label="提现" name="withdraw" />
					<el-tab-pane label="转存" name="transfer" />
					<el-tab-pane label="账款调整" name="adjustment" />
				</el-tabs>
				<div class="filter-row">
					<el-input
						v-model="searchKeyword"
						placeholder="快速搜索"
						clearable
						class="search-input"
					>
						<template #prefix>
							<el-icon><Search /></el-icon>
						</template>
					</el-input>
					<el-date-picker
						v-model="dateRange"
						type="daterange"
						range-separator="-"
						start-placeholder="开始日期"
						end-placeholder="结束日期"
						format="YYYY-MM-DD"
						value-format="YYYY-MM-DD"
						class="date-picker"
					/>
				</div>
			</div>

			<!-- 表格 -->
			<el-table :data="filteredList" style="width: 100%" class="commission-table">
				<el-table-column prop="date" label="日期" min-width="1" align="left" header-align="center" />
				<el-table-column prop="type" label="交易类型" min-width="1" align="center" header-align="center" />
				<el-table-column prop="storeId" label="店铺编号" min-width="1" align="center" header-align="center" />
				<el-table-column prop="orderNo" label="订单编号" min-width="1" align="center" header-align="center" />
				<el-table-column prop="amount" label="交易金额" min-width="1" align="center" header-align="center">
					<template #default="{ row }">
						NT${{ row.amount }}
					</template>
				</el-table-column>
				<el-table-column prop="balance" label="余额" min-width="1" align="center" header-align="center">
					<template #default="{ row }">
						NT${{ row.balance }}
					</template>
				</el-table-column>
				<el-table-column prop="status" label="状态" min-width="1" align="center" header-align="center">
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
	</div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

interface CommissionRecord {
	date: string
	type: string
	storeId: string
	orderNo: string
	amount: string
	balance: string
	status: string
}

const activeTab = ref('all')
const searchKeyword = ref('')
const dateRange = ref<string[]>(['2025-09-01', '2025-09-10'])

const summaryData = reactive({
	availableAmount: 5450.00,
	totalCommission: 24543.00,
	pendingAmount: 123.00
})

const pagination = reactive({
	page: 1,
	pageSize: 10,
	total: 123
})

// Mock数据
const commissionList = ref<CommissionRecord[]>([
	{ date: '2026-12-12 23:59:59', type: '佣金收入', storeId: 'S1234567890', orderNo: 'X250904KQ2P078R', amount: '1,000.00', balance: '223,560.50', status: '已结算' },
	{ date: '2026-12-12 23:59:59', type: '佣金收入', storeId: 'S1234567890', orderNo: 'X250904KQ2P078R', amount: '1,000.00', balance: '223,560.50', status: '已结算' },
	{ date: '2026-12-12 23:59:59', type: '佣金收入', storeId: 'S1234567890', orderNo: 'X250904KQ2P078R', amount: '1,000.00', balance: '223,560.50', status: '已结算' },
	{ date: '2026-12-12 23:59:59', type: '佣金收入', storeId: 'S1234567890', orderNo: 'X250904KQ2P078R', amount: '1,000.00', balance: '223,560.50', status: '已结算' },
	{ date: '2026-12-12 23:59:59', type: '佣金收入', storeId: 'S1234567890', orderNo: 'X250904KQ2P078R', amount: '1,000.00', balance: '223,560.50', status: '已结算' },
	{ date: '2026-12-12 23:59:59', type: '佣金收入', storeId: 'S1234567890', orderNo: 'X250904KQ2P078R', amount: '1,000.00', balance: '223,560.50', status: '已结算' },
	{ date: '2026-12-12 23:59:59', type: '佣金收入', storeId: 'S1234567890', orderNo: 'X250904KQ2P078R', amount: '1,000.00', balance: '223,560.50', status: '已结算' },
	{ date: '2026-12-12 23:59:59', type: '佣金收入', storeId: 'S1234567890', orderNo: 'X250904KQ2P078R', amount: '1,000.00', balance: '223,560.50', status: '已结算' },
	{ date: '2026-12-12 23:59:59', type: '佣金收入', storeId: 'S1234567890', orderNo: 'X250904KQ2P078R', amount: '1,000.00', balance: '223,560.50', status: '已结算' },
	{ date: '2026-12-12 23:59:59', type: '佣金收入', storeId: 'S1234567890', orderNo: 'X250904KQ2P078R', amount: '1,000.00', balance: '223,560.50', status: '已结算' }
])

const filteredList = computed(() => {
	let result = [...commissionList.value]
	if (searchKeyword.value) {
		const keyword = searchKeyword.value.toLowerCase()
		result = result.filter(item =>
			item.orderNo.toLowerCase().includes(keyword) ||
			item.storeId.toLowerCase().includes(keyword)
		)
	}
	return result
})

const formatAmount = (value: number) => {
	return value.toLocaleString('zh-CN', {
		minimumFractionDigits: 2,
		maximumFractionDigits: 2
	})
}

const handleWithdraw = () => {
	ElMessage.info('提现功能开发中...')
}

const handleTransfer = () => {
	ElMessage.info('转存功能开发中...')
}

const handleSizeChange = (size: number) => {
	pagination.pageSize = size
	pagination.page = 1
}

const handlePageChange = (page: number) => {
	pagination.page = page
}
</script>

<style scoped lang="scss">
.finance-commission-page {
	padding: 20px;
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

.summary-card {
	margin-bottom: 20px;
	border-radius: 8px;

	.summary-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 20px;

		.summary-title {
			font-size: 16px;
			font-weight: 500;
			color: #303133;
		}

		.stats-link {
			font-size: 14px;
			color: #909399;
			text-decoration: none;

			&:hover {
				color: #ff6a3a;
			}
		}
	}

	.summary-content {
		.balance-card {
			background-color: #f5f7fa;
			border-radius: 8px;
			padding: 24px;
			display: flex;
			justify-content: space-between;
			align-items: flex-end;

			.balance-info {
				.balance-label {
					font-size: 14px;
					color: #909399;
					margin-bottom: 8px;
				}

				.balance-value {
					font-size: 32px;
					font-weight: 600;
					color: #303133;
					margin-bottom: 8px;
				}

				.balance-subtitle {
					font-size: 12px;
					color: #909399;
				}
			}

			.balance-actions {
				display: flex;
				gap: 12px;
			}
		}

		.pending-card {
			background-color: #f5f7fa;
			border-radius: 8px;
			padding: 24px;
			height: 100%;

			.pending-label {
				font-size: 14px;
				color: #909399;
				margin-bottom: 8px;
			}

			.pending-value {
				font-size: 32px;
				font-weight: 600;
				color: #303133;
				margin-bottom: 8px;
			}

			.pending-subtitle {
				font-size: 12px;
				color: #909399;
			}
		}
	}
}

.list-card {
	border-radius: 8px;

	.tabs-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		flex-wrap: wrap;
		margin-bottom: 20px;
		border-bottom: 1px solid #e4e7ed;
	}

	.commission-tabs {
		flex: 1;

		:deep(.el-tabs__header) {
			margin-bottom: 0;
		}

		:deep(.el-tabs__nav-wrap::after) {
			display: none;
		}

		:deep(.el-tabs__item.is-active) {
			color: #303133;
		}

		:deep(.el-tabs__item:not(.is-active):hover) {
			color: #ff6a3a;
		}

		:deep(.el-tabs__active-bar) {
			background-color: #ff6a3a;
		}
	}

	.filter-row {
		display: flex;
		align-items: center;
		gap: 16px;

		.search-input {
			width: 160px;

			:deep(.el-input__wrapper) {
				border-radius: 30px;
			}
		}

		.date-picker {
			width: 240px;
		}
	}

	.commission-table {
		:deep(.el-table__header-wrapper) {
			border-radius: 8px;
			overflow: hidden;
		}

		:deep(.el-table__header th) {
			background-color: #f5f7fa;
			color: #606266;
			font-weight: 500;
		}

		.status-text {
			color: #ff6a3a;
		}
	}

	.pagination-wrapper {
		margin-top: 20px;
		display: flex;
		justify-content: flex-end;
	}
}
</style>
