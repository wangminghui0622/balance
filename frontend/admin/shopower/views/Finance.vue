<template>
	<div class="finance-page">
		<!-- 账单总览 -->
		<el-card class="summary-card">
			<div class="summary-header">
				<span class="summary-title">账单总览</span>
			</div>
			<el-row :gutter="20" class="summary-cards">
				<el-col :xs="24" :sm="12" :md="6" :lg="6" :xl="6">
					<div class="stat-card">
						<div class="stat-label">预估未结算佣金</div>
						<div class="stat-value">NT${{ formatAmount(summaryData.unsettledCommission) }}</div>
					</div>
				</el-col>
				<el-col :xs="24" :sm="12" :md="6" :lg="6" :xl="6">
					<div class="stat-card">
						<div class="stat-label">已结算佣金</div>
						<div class="stat-value">NT${{ formatAmount(summaryData.settledCommission) }}</div>
					</div>
				</el-col>
				<el-col :xs="24" :sm="12" :md="6" :lg="6" :xl="6">
					<div class="stat-card">
						<div class="stat-label">账款调整</div>
						<div class="stat-value">NT${{ formatAmount(summaryData.adjustment) }}</div>
					</div>
				</el-col>
				<el-col :xs="24" :sm="12" :md="6" :lg="6" :xl="6">
					<div class="stat-card">
						<div class="stat-label" style="white-space: nowrap;">预估佣金总额（未结算+已结算+账款调整）</div>
						<div class="stat-value highlight">NT${{ formatAmount(summaryData.totalCommission) }}</div>
					</div>
				</el-col>
			</el-row>
		</el-card>

		<!-- 订单列表 -->
		<el-card class="orders-card">
			<div class="tabs-header">
				<el-tabs v-model="activeTab" class="finance-tabs">
					<el-tab-pane label="未结算" name="unsettled" />
					<el-tab-pane label="已结算" name="settled" />
					<el-tab-pane label="账款调整" name="adjustment" />
				</el-tabs>
				<!-- 搜索和导出 -->
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
					<el-button class="export-btn" @click="handleExport">
						<svg class="export-icon" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
							<rect x="1" y="1" width="14" height="14" rx="2" stroke="currentColor" stroke-width="1.5" fill="none"/>
						</svg>
						导出报表
					</el-button>
				</div>
			</div>

			<!-- 表格 -->
			<el-table :data="filteredOrders" style="width: 100%" class="finance-table">
				<el-table-column prop="date" label="日期" min-width="1" align="left" header-align="center" />
				<el-table-column prop="storeId" label="店铺编号" min-width="1" align="center" header-align="center" />
				<el-table-column prop="orderNo" label="订单编号" min-width="1" align="center" header-align="center" />
				<el-table-column prop="orderStatus" label="订单状态" min-width="1" align="center" header-align="center">
					<template #default="{ row }">
						<span :class="['status-tag', getStatusClass(row.orderStatus)]">{{ row.orderStatus }}</span>
					</template>
				</el-table-column>
				<el-table-column prop="countdown" label="订单计时" min-width="1" align="center" header-align="center" />
				<el-table-column prop="orderAmount" label="订单金额" min-width="1" align="center" header-align="center">
					<template #default="{ row }">
						NT${{ row.orderAmount }}
					</template>
				</el-table-column>
				<el-table-column prop="commission" label="未结算佣金" min-width="1" align="center" header-align="center">
					<template #default="{ row }">
						NT${{ row.commission }}
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

interface FinanceOrder {
	date: string
	storeId: string
	orderNo: string
	orderStatus: string
	countdown: string
	orderAmount: string
	commission: string
}

const activeTab = ref('unsettled')
const searchKeyword = ref('')

const summaryData = reactive({
	unsettledCommission: 245,
	settledCommission: 123.00,
	adjustment: 123.00,
	totalCommission: 123456.00
})

const pagination = reactive({
	page: 1,
	pageSize: 10,
	total: 92
})

// Mock数据
const orders = ref<FinanceOrder[]>([
	{
		date: '2026-12-12 23:59:59',
		storeId: '5123456789o',
		orderNo: 'X25090a4KQ2Po78R',
		orderStatus: '待发货',
		countdown: '1天 03:23:42',
		orderAmount: '6338.00',
		commission: '8333.00'
	},
	{
		date: '2026-12-12 23:59:59',
		storeId: '5123456789o',
		orderNo: 'X25090a4KQ2Po78R',
		orderStatus: '待发货',
		countdown: '1天 03:23:42',
		orderAmount: '68.00',
		commission: '8.00'
	},
	{
		date: '2026-12-12 23:59:59',
		storeId: '5123456789o',
		orderNo: 'X25090a4KQ2Po78R',
		orderStatus: '待发货',
		countdown: '1天 03:23:42',
		orderAmount: '68.00',
		commission: '8.00'
	},
	{
		date: '2026-12-12 23:59:59',
		storeId: '5123456789o',
		orderNo: 'X25090a4KQ2Po78R',
		orderStatus: '待发货',
		countdown: '1天 03:23:42',
		orderAmount: '68.00',
		commission: '8.00'
	},
	{
		date: '2026-12-12 23:59:59',
		storeId: '5123456789o',
		orderNo: 'X25090a4KQ2Po78R',
		orderStatus: '待发货',
		countdown: '1天 03:23:42',
		orderAmount: '68.00',
		commission: '8.00'
	},
	{
		date: '2026-12-12 23:59:59',
		storeId: '5123456789o',
		orderNo: 'X25090a4KQ2Po78R',
		orderStatus: '待发货',
		countdown: '1天 03:23:42',
		orderAmount: '68.00',
		commission: '8.00'
	},
	{
		date: '2026-12-12 23:59:59',
		storeId: '5123456789o',
		orderNo: 'X25090a4KQ2Po78R',
		orderStatus: '待发货',
		countdown: '1天 03:23:42',
		orderAmount: '68.00',
		commission: '8.00'
	},
	{
		date: '2026-12-12 23:59:59',
		storeId: '5123456789o',
		orderNo: 'X25090a4KQ2Po78R',
		orderStatus: '待发货',
		countdown: '1天 03:23:42',
		orderAmount: '68.00',
		commission: '8.00'
	},
	{
		date: '2026-12-12 23:59:59',
		storeId: '5123456789o',
		orderNo: 'X25090a4KQ2Po78R',
		orderStatus: '待发货',
		countdown: '1天 03:23:42',
		orderAmount: '68.00',
		commission: '8.00'
	},
	{
		date: '2026-12-12 23:59:59',
		storeId: '5123456789o',
		orderNo: 'X25090a4KQ2Po78R',
		orderStatus: '待发货',
		countdown: '1天 03:23:42',
		orderAmount: '68.00',
		commission: '8.00'
	}
])

const filteredOrders = computed(() => {
	let result = [...orders.value]
	if (searchKeyword.value) {
		const keyword = searchKeyword.value.toLowerCase()
		result = result.filter(order =>
			order.orderNo.toLowerCase().includes(keyword) ||
			order.storeId.toLowerCase().includes(keyword)
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

const getStatusClass = (status: string) => {
	switch (status) {
		case '待发货':
			return 'status-pending'
		case '已发货':
			return 'status-shipped'
		case '已完成':
			return 'status-completed'
		case '已取消':
			return 'status-cancelled'
		default:
			return ''
	}
}

const handleExport = () => {
	ElMessage.success('正在导出报表...')
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
.finance-page {
	padding: 20px;
}

.summary-card {
	margin-bottom: 20px;
	border-radius: 8px;

	.summary-header {
		margin-bottom: 20px;

		.summary-title {
			font-size: 16px;
			font-weight: 500;
			color: #303133;
		}
	}

	.summary-cards {
		.stat-card {
			background-color: #f5f7fa;
			border-radius: 8px;
			padding: 20px;
			text-align: center;

			.stat-label {
				font-size: 14px;
				color: #909399;
				margin-bottom: 12px;
			}

			.stat-value {
				font-size: 28px;
				font-weight: 600;
				color: #303133;

				&.highlight {
					color: #ff6a3a;
				}
			}
		}
	}
}

.orders-card {
	border-radius: 8px;

	.tabs-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		flex-wrap: wrap;
		margin-bottom: 20px;
		border-bottom: 1px solid #e4e7ed;
	}

	.finance-tabs {
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
			width: 200px;

			:deep(.el-input__wrapper) {
				border-radius: 30px;
			}
		}

		.export-btn {
			display: flex;
			align-items: center;
			gap: 4px;

			.export-icon {
				width: 16px;
				height: 16px;
			}
		}
	}

	.finance-table {
		:deep(.el-table__header-wrapper) {
			border-radius: 8px;
			overflow: hidden;
		}

		:deep(.el-table__header th) {
			background-color: #f5f7fa;
			color: #606266;
			font-weight: 500;
		}

		.status-tag {
			padding: 4px 8px;
			border-radius: 4px;
			font-size: 12px;

			&.status-pending {
				background-color: #fff7e6;
				color: #fa8c16;
			}

			&.status-shipped {
				background-color: #e6f7ff;
				color: #1890ff;
			}

			&.status-completed {
				background-color: #f6ffed;
				color: #52c41a;
			}

			&.status-cancelled {
				background-color: #fff1f0;
				color: #ff4d4f;
			}
		}
	}

	.pagination-wrapper {
		margin-top: 20px;
		display: flex;
		justify-content: flex-end;
	}
}
</style>
