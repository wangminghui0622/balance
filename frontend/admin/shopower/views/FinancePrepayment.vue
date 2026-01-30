<template>
	<div class="finance-prepayment-page">
		<!-- 页面标题 -->
		<div class="page-header">
			<h1 class="page-title">我的预付款</h1>
		</div>

		<!-- 预付款总览和计算说明 -->
		<el-row :gutter="20" class="summary-row">
			<el-col :xs="24" :sm="24" :md="14" :lg="14" :xl="14">
				<el-card class="summary-card">
					<div class="summary-header">
						<span class="summary-title">预付款总览</span>
						<a href="#" class="stats-link">预付款统计</a>
					</div>
					<div class="balance-card">
						<div class="balance-info">
							<div class="balance-label">预付款余额</div>
							<div class="balance-value">NT${{ formatAmount(summaryData.balance) }}</div>
							<div class="balance-subtitle">累计充值金额: NT${{ formatAmount(summaryData.totalRecharge) }}</div>
						</div>
						<div class="balance-actions">
							<el-button type="primary" @click="handleRecharge">充值</el-button>
							<el-button @click="handleWithdraw">提现</el-button>
						</div>
					</div>
				</el-card>
			</el-col>
			<el-col :xs="24" :sm="24" :md="10" :lg="10" :xl="10">
				<el-card class="calc-card">
					<div class="calc-header">
						<span class="calc-title">计算说明</span>
					</div>
					<div class="calc-content">
						<div class="calc-formula">预付款余额: NT${{ formatAmount(summaryData.balance) }}</div>
						<div class="calc-items">
							<div class="calc-item">
								<div class="calc-label">充值</div>
								<div class="calc-value">{{ summaryData.rechargeCount }}</div>
							</div>
							<div class="calc-operator">+</div>
							<div class="calc-item">
								<div class="calc-label">转存</div>
								<div class="calc-value">{{ summaryData.transferCount }}</div>
							</div>
							<div class="calc-operator">-</div>
							<div class="calc-item">
								<div class="calc-label">提现</div>
								<div class="calc-value">{{ summaryData.withdrawCount }}</div>
							</div>
							<div class="calc-operator">-</div>
							<div class="calc-item">
								<div class="calc-label">订单付款</div>
								<div class="calc-value">{{ summaryData.orderPayCount }}</div>
							</div>
							<div class="calc-operator">-</div>
							<div class="calc-item">
								<div class="calc-label">账款调整</div>
								<div class="calc-value">{{ summaryData.adjustmentCount }}</div>
							</div>
						</div>
					</div>
				</el-card>
			</el-col>
		</el-row>

		<!-- 预付款列表 -->
		<el-card class="list-card">
			<div class="tabs-header">
				<el-tabs v-model="activeTab" class="prepayment-tabs">
					<el-tab-pane label="全部" name="all" />
					<el-tab-pane label="充值" name="recharge" />
					<el-tab-pane label="转存" name="transfer" />
					<el-tab-pane label="提现" name="withdraw" />
					<el-tab-pane label="订单付款" name="orderPay" />
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
			<el-table :data="filteredList" style="width: 100%" class="prepayment-table">
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

interface PrepaymentRecord {
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
	balance: 223560.50,
	totalRecharge: 1445860.50,
	rechargeCount: 12344,
	transferCount: 12344,
	withdrawCount: 123,
	orderPayCount: 12344,
	adjustmentCount: 12344
})

const pagination = reactive({
	page: 1,
	pageSize: 10,
	total: 113
})

// Mock数据
const prepaymentList = ref<PrepaymentRecord[]>([
	{ date: '2026-12-12 23:59:59', type: '充值', channel: '文字占位符文字占位符', orderNo: 'X250904KQ2P078R', amount: '1,000.00', balance: '223,560.50', status: '已完成' },
	{ date: '2026-12-12 23:59:59', type: '转存', channel: '文字占位符文字占位符', orderNo: 'X250904KQ2P078R', amount: '1,000.00', balance: '223,560.50', status: '已完成' },
	{ date: '2026-12-12 23:59:59', type: '提现', channel: '文字占位符文字占位符', orderNo: 'X250904KQ2P078R', amount: '1,000.00', balance: '223,560.50', status: '已完成' },
	{ date: '2026-12-12 23:59:59', type: '订单付款', channel: '文字占位符文字占位符', orderNo: 'X250904KQ2P078R', amount: '1,000.00', balance: '223,560.50', status: '已完成' },
	{ date: '2026-12-12 23:59:59', type: '账款调整', channel: '文字占位符文字占位符', orderNo: 'X250904KQ2P078R', amount: '1,000.00', balance: '223,560.50', status: '已完成' },
	{ date: '2026-12-12 23:59:59', type: '充值', channel: '文字占位符文字占位符', orderNo: 'X250904KQ2P078R', amount: '1,000.00', balance: '223,560.50', status: '已完成' },
	{ date: '2026-12-12 23:59:59', type: '充值', channel: '文字占位符文字占位符', orderNo: 'X250904KQ2P078R', amount: '1,000.00', balance: '223,560.50', status: '已完成' },
	{ date: '2026-12-12 23:59:59', type: '充值', channel: '文字占位符文字占位符', orderNo: 'X250904KQ2P078R', amount: '1,000.00', balance: '223,560.50', status: '已完成' },
	{ date: '2026-12-12 23:59:59', type: '充值', channel: '文字占位符文字占位符', orderNo: 'X250904KQ2P078R', amount: '1,000.00', balance: '223,560.50', status: '已完成' },
	{ date: '2026-12-12 23:59:59', type: '充值', channel: '文字占位符文字占位符', orderNo: 'X250904KQ2P078R', amount: '1,000.00', balance: '223,560.50', status: '已完成' }
])

const filteredList = computed(() => {
	let result = [...prepaymentList.value]
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
.finance-prepayment-page {
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

.summary-row {
	margin-bottom: 20px;
}

.summary-card {
	border-radius: 8px;
	height: 100%;

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

.calc-card {
	border-radius: 8px;
	height: 100%;

	.calc-header {
		margin-bottom: 16px;

		.calc-title {
			font-size: 16px;
			font-weight: 500;
			color: #303133;
		}
	}

	.calc-content {
		background-color: #f5f7fa;
		border-radius: 8px;
		padding: 20px;

		.calc-formula {
			font-size: 14px;
			color: #606266;
			margin-bottom: 16px;
			padding-bottom: 12px;
			border-bottom: 1px solid #e4e7ed;
		}

		.calc-items {
			display: flex;
			align-items: center;
			justify-content: space-between;
			gap: 8px;

			.calc-item {
				text-align: center;

				.calc-label {
					font-size: 12px;
					color: #909399;
					margin-bottom: 4px;
				}

				.calc-value {
					font-size: 16px;
					font-weight: 500;
					color: #303133;
				}
			}

			.calc-operator {
				font-size: 16px;
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

	.prepayment-tabs {
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

	.prepayment-table {
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
