<template>
	<div class="finance-deposit-page">
		<!-- 页面标题 -->
		<div class="page-header">
			<h1 class="page-title">店主保证金</h1>
		</div>

		<!-- 保证金概览 -->
		<el-card class="summary-card">
			<div class="summary-header">
				<span class="summary-title">保证金概览</span>
			</div>
			<div class="balance-card">
				<div class="balance-info">
					<div class="balance-label">保证金余额</div>
					<div class="balance-value">¥{{ formatAmount(summaryData.balance) }}</div>
					<div class="balance-subtitle">保证金门槛: NT$5,000.00或以上</div>
				</div>
				<div class="balance-actions">
					<el-button type="primary" @click="handleRecharge">充值</el-button>
					<el-button @click="handleWithdraw">提现</el-button>
				</div>
			</div>
		</el-card>

		<!-- 保证金列表 -->
		<el-card class="list-card">
			<div class="tabs-header">
				<el-tabs v-model="activeTab" class="deposit-tabs">
					<el-tab-pane label="全部" name="all" />
					<el-tab-pane label="充值" name="recharge" />
					<el-tab-pane label="提现" name="withdraw" />
					<el-tab-pane label="扣除" name="deduct" />
					<el-tab-pane label="补贴" name="subsidy" />
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
			<el-table :data="filteredList" style="width: 100%" class="deposit-table">
				<el-table-column prop="date" label="日期" min-width="1" align="left" header-align="center" />
				<el-table-column prop="type" label="交易类型" min-width="1" align="center" header-align="center" />
				<el-table-column prop="channel" label="交易渠道" min-width="1" align="center" header-align="center" />
				<el-table-column prop="orderNo" label="交易单号" min-width="1" align="center" header-align="center" />
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

interface DepositRecord {
	date: string
	type: string
	channel: string
	orderNo: string
	amount: string
	balance: string
	status: string
}

const activeTab = ref('all')
const searchKeyword = ref('')
const dateRange = ref<string[]>(['2025-09-01', '2025-09-10'])

const summaryData = reactive({
	balance: 5000.00
})

const pagination = reactive({
	page: 1,
	pageSize: 10,
	total: 123
})

// Mock数据
const depositList = ref<DepositRecord[]>([
	{ date: '2026-12-12 23:59:59', type: '充值', channel: '文字占位符文字占位符', orderNo: 'CZ1234567890', amount: '1,000.00', balance: '223,560.50', status: '已完成' },
	{ date: '2026-12-12 23:59:59', type: '充值', channel: '文字占位符文字占位符', orderNo: 'CZ1234567890', amount: '1,000.00', balance: '223,560.50', status: '已完成' },
	{ date: '2026-12-12 23:59:59', type: '充值', channel: '文字占位符文字占位符', orderNo: 'CZ1234567890', amount: '1,000.00', balance: '223,560.50', status: '已完成' },
	{ date: '2026-12-12 23:59:59', type: '充值', channel: '文字占位符文字占位符', orderNo: 'CZ1234567890', amount: '1,000.00', balance: '223,560.50', status: '已完成' },
	{ date: '2026-12-12 23:59:59', type: '充值', channel: '文字占位符文字占位符', orderNo: 'CZ1234567890', amount: '1,000.00', balance: '223,560.50', status: '已完成' },
	{ date: '2026-12-12 23:59:59', type: '充值', channel: '文字占位符文字占位符', orderNo: 'CZ1234567890', amount: '1,000.00', balance: '223,560.50', status: '已完成' },
	{ date: '2026-12-12 23:59:59', type: '充值', channel: '文字占位符文字占位符', orderNo: 'CZ1234567890', amount: '1,000.00', balance: '223,560.50', status: '已完成' },
	{ date: '2026-12-12 23:59:59', type: '充值', channel: '文字占位符文字占位符', orderNo: 'CZ1234567890', amount: '1,000.00', balance: '223,560.50', status: '已完成' },
	{ date: '2026-12-12 23:59:59', type: '充值', channel: '文字占位符文字占位符', orderNo: 'CZ1234567890', amount: '1,000.00', balance: '223,560.50', status: '已完成' },
	{ date: '2026-12-12 23:59:59', type: '充值', channel: '文字占位符文字占位符', orderNo: 'CZ1234567890', amount: '1,000.00', balance: '223,560.50', status: '已完成' }
])

const filteredList = computed(() => {
	let result = [...depositList.value]
	if (searchKeyword.value) {
		const keyword = searchKeyword.value.toLowerCase()
		result = result.filter(item =>
			item.orderNo.toLowerCase().includes(keyword) ||
			item.type.toLowerCase().includes(keyword)
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

const handleRecharge = () => {
	ElMessage.info('充值功能开发中...')
}

const handleWithdraw = () => {
	ElMessage.info('提现功能开发中...')
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
.finance-deposit-page {
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
	border: 2px solid #ff6a3a;

	.summary-header {
		margin-bottom: 16px;

		.summary-title {
			font-size: 16px;
			font-weight: 500;
			color: #303133;
		}
	}

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

	.deposit-tabs {
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

	.deposit-table {
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
