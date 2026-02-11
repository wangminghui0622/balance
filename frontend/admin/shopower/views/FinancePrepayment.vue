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

	<!-- 充值对话框 -->
	<el-dialog
		v-model="rechargeDialogVisible"
		title="充值"
		width="480px"
		class="recharge-dialog"
		:close-on-click-modal="false"
	>
		<div class="recharge-content">
			<!-- 金额显示 -->
			<div class="amount-display">
				<span class="currency">NT$</span>
				<span class="amount" :class="{ 'is-zero': rechargeAmount <= 0 }">{{ formatAmount(rechargeAmount) }}</span>
			</div>

			<!-- 快捷金额选择 -->
			<div class="quick-amounts">
				<div
					v-for="amount in quickAmounts"
					:key="amount"
					class="quick-amount-btn"
					:class="{ active: rechargeAmount === amount }"
					@click="rechargeAmount = amount"
				>
					{{ amount }}
				</div>
			</div>

			<!-- 支付平台 -->
			<div class="payment-section">
				<div class="section-title">支付平台</div>
				<div class="payment-methods">
					<div
						v-for="method in paymentMethods"
						:key="method.id"
						class="payment-method"
						:class="{ active: selectedPayment === method.id }"
						@click="selectedPayment = method.id"
					>
						<img :src="method.icon" :alt="method.name" class="payment-icon" />
					</div>
				</div>
			</div>

			<!-- 支付金额 -->
			<div class="pay-amount">
				<span class="label">支付金额：</span>
				<span class="value">NT${{ formatAmount(rechargeAmount) }}</span>
			</div>

			<!-- 注意事项 -->
			<div class="notice-section">
				<div class="notice-title">注意事项：</div>
				<div class="notice-list">
					<div class="notice-item">1. 示例文字占位符，替换成示例文字占位符，替换成示例文字占位符，替换成示例文字占位符，替换成</div>
					<div class="notice-item">2. 示例文字占位符，替换成示例文字占位符，替换成示例文字占位符，替换成示例文字占位符，替换成</div>
					<div class="notice-item">3. 示例文字占位符，替换成示例文字占位符，替换成示例文字占位符，替换成示例文字占位符，替换成</div>
				</div>
			</div>

			<!-- 协议复选框 -->
			<div class="agreement">
				<el-checkbox v-model="agreeTerms">我已阅读示例文字并同意授权协议</el-checkbox>
			</div>
		</div>

		<template #footer>
			<div class="dialog-footer">
				<el-button @click="rechargeDialogVisible = false">取消</el-button>
				<el-button type="primary" :disabled="!agreeTerms || rechargeAmount <= 0 || !selectedPayment" @click="confirmRecharge">确定</el-button>
			</div>
		</template>
	</el-dialog>

	<!-- 提现对话框 -->
	<el-dialog
		v-model="withdrawDialogVisible"
		title="提现"
		width="480px"
		class="withdraw-dialog"
		:close-on-click-modal="false"
	>
		<div class="withdraw-content">
			<!-- 金额输入 -->
			<div class="amount-row">
				<div class="amount-input-wrapper">
					<span class="currency">NT$</span>
					<el-input-number
						v-model="withdrawAmount"
						:min="0"
						:max="summaryData.balance"
						:precision="2"
						:controls="false"
						placeholder="请输入提现金额"
						class="amount-input"
					/>
				</div>
				<div class="balance-info">
					<span class="balance-label">余额：</span>
					<span class="balance-value">NT${{ formatAmount(summaryData.balance) }}</span>
					<el-button type="primary" link size="small" @click="withdrawAmount = summaryData.balance">全部</el-button>
				</div>
			</div>

			<!-- 提现渠道 -->
			<div class="channel-section">
				<div class="section-title">提现渠道</div>
				<div class="channel-methods">
					<div
						v-for="method in withdrawMethods"
						:key="method.id"
						class="channel-method"
						:class="{ active: selectedWithdrawChannel === method.id }"
						@click="selectedWithdrawChannel = method.id"
					>
						<img :src="method.icon" :alt="method.name" class="channel-icon" />
					</div>
				</div>
			</div>

			<!-- 提现手续费 -->
			<div class="fee-info">
				<span class="fee-label">提现手续费：</span>
				<span class="fee-value">2%</span>
			</div>

			<!-- 注意事项 -->
			<div class="notice-section">
				<div class="notice-title">注意事项：</div>
				<div class="notice-list">
					<div class="notice-item">1. 示例文字占位符，替换成可示例文字占位符，替换成可示例文字占位符，替换成可示例文字占位符，替换成</div>
					<div class="notice-item">2. 示例文字占位符，替换成可示例文字占位符，替换成可示例文字占位符，替换成可示例文字占位符，替换成</div>
					<div class="notice-item">3. 示例文字占位符，替换成可示例文字占位符，替换成可示例文字占位符，替换成可示例文字占位符，替换成</div>
				</div>
			</div>

			<!-- 协议复选框 -->
			<div class="agreement">
				<el-checkbox v-model="agreeWithdrawTerms">参数示例文字占位符需要时可</el-checkbox>
			</div>
		</div>

		<template #footer>
			<div class="dialog-footer">
				<el-button @click="withdrawDialogVisible = false">取消</el-button>
				<el-button type="primary" :disabled="!agreeWithdrawTerms || withdrawAmount <= 0 || !selectedWithdrawChannel" @click="confirmWithdraw">确定</el-button>
			</div>
		</template>
	</el-dialog>
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

const rechargeDialogVisible = ref(false)
const rechargeAmount = ref(0)
const selectedPayment = ref('')
const agreeTerms = ref(false)

const quickAmounts = [500, 1000, 2000, 3000, 4000, 5000, 10000]

const paymentMethods = [
	{ id: 'paypal', name: 'PayPal', icon: 'https://www.paypalobjects.com/webstatic/mktg/Logo/pp-logo-100px.png' },
	{ id: 'alipay', name: '支付宝', icon: 'https://gw.alipayobjects.com/mdn/rms_0c75a8/afts/img/A*V3ICRJ-4bDcAAAAAAAAAAAAAARQnAQ' },
	{ id: 'linepay', name: 'LINE Pay', icon: 'https://scdn.line-apps.com/linepay/portal/v-240930/portal/img/sp/sp_logo_linepay_white.png' },
	{ id: 'visa', name: 'VISA', icon: 'https://upload.wikimedia.org/wikipedia/commons/thumb/5/5e/Visa_Inc._logo.svg/200px-Visa_Inc._logo.svg.png' },
	{ id: 'paypal2', name: 'PayPal', icon: 'https://www.paypalobjects.com/webstatic/mktg/Logo/pp-logo-100px.png' },
	{ id: 'alipay2', name: '支付宝', icon: 'https://gw.alipayobjects.com/mdn/rms_0c75a8/afts/img/A*V3ICRJ-4bDcAAAAAAAAAAAAAARQnAQ' },
	{ id: 'linepay2', name: 'LINE Pay', icon: 'https://scdn.line-apps.com/linepay/portal/v-240930/portal/img/sp/sp_logo_linepay_white.png' },
	{ id: 'visa2', name: 'VISA', icon: 'https://upload.wikimedia.org/wikipedia/commons/thumb/5/5e/Visa_Inc._logo.svg/200px-Visa_Inc._logo.svg.png' }
]

const handleRecharge = () => {
	rechargeAmount.value = 0
	selectedPayment.value = ''
	agreeTerms.value = false
	rechargeDialogVisible.value = true
}

const confirmRecharge = () => {
	if (!agreeTerms.value) {
		ElMessage.warning('请先同意授权协议')
		return
	}
	if (rechargeAmount.value <= 0) {
		ElMessage.warning('请选择充值金额')
		return
	}
	if (!selectedPayment.value) {
		ElMessage.warning('请选择支付平台')
		return
	}
	ElMessage.success(`充值 NT$${formatAmount(rechargeAmount.value)} 成功`)
	rechargeDialogVisible.value = false
}

const withdrawDialogVisible = ref(false)
const withdrawAmount = ref(0)
const selectedWithdrawChannel = ref('')
const agreeWithdrawTerms = ref(false)

const withdrawMethods = [
	{ id: 'paypal', name: 'PayPal', icon: 'https://www.paypalobjects.com/webstatic/mktg/Logo/pp-logo-100px.png' },
	{ id: 'alipay', name: '支付宝', icon: 'https://gw.alipayobjects.com/mdn/rms_0c75a8/afts/img/A*V3ICRJ-4bDcAAAAAAAAAAAAAARQnAQ' },
	{ id: 'linepay', name: 'LINE Pay', icon: 'https://scdn.line-apps.com/linepay/portal/v-240930/portal/img/sp/sp_logo_linepay_white.png' },
	{ id: 'visa', name: 'VISA', icon: 'https://upload.wikimedia.org/wikipedia/commons/thumb/5/5e/Visa_Inc._logo.svg/200px-Visa_Inc._logo.svg.png' }
]

const handleWithdraw = () => {
	withdrawAmount.value = 0
	selectedWithdrawChannel.value = ''
	agreeWithdrawTerms.value = false
	withdrawDialogVisible.value = true
}

const confirmWithdraw = () => {
	if (!agreeWithdrawTerms.value) {
		ElMessage.warning('请先同意协议')
		return
	}
	if (withdrawAmount.value <= 0) {
		ElMessage.warning('请输入提现金额')
		return
	}
	if (!selectedWithdrawChannel.value) {
		ElMessage.warning('请选择提现渠道')
		return
	}
	if (withdrawAmount.value > summaryData.balance) {
		ElMessage.warning('提现金额不能超过余额')
		return
	}
	ElMessage.success(`提现 NT$${formatAmount(withdrawAmount.value)} 申请已提交`)
	withdrawDialogVisible.value = false
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

.recharge-dialog {
	:deep(.el-dialog__header) {
		padding: 16px 20px;
		border-bottom: 1px solid #e4e7ed;
		margin-right: 0;
	}

	:deep(.el-dialog__body) {
		padding: 20px;
	}

	:deep(.el-dialog__footer) {
		padding: 12px 20px;
		border-top: 1px solid #e4e7ed;
	}

	.recharge-content {
		.amount-display {
			margin-bottom: 20px;

			.currency {
				font-size: 16px;
				color: #409eff;
				margin-right: 4px;
			}

			.amount {
				font-size: 24px;
				font-weight: 600;
				color: #409eff;

				&.is-zero {
					color: #c0c4cc;
				}
			}
		}

		.quick-amounts {
			display: flex;
			flex-wrap: wrap;
			gap: 12px;
			margin-bottom: 20px;

			.quick-amount-btn {
				width: calc((100% - 72px) / 7);
				min-width: 50px;
				padding: 8px 0;
				text-align: center;
				border: 1px solid #dcdfe6;
				border-radius: 4px;
				font-size: 14px;
				color: #606266;
				cursor: pointer;
				transition: all 0.3s;

				&:hover {
					border-color: #409eff;
					color: #409eff;
				}

				&.active {
					border-color: #409eff;
					background-color: #409eff;
					color: #fff;
				}
			}
		}

		.payment-section {
			margin-bottom: 20px;

			.section-title {
				font-size: 14px;
				color: #303133;
				margin-bottom: 12px;
				padding-left: 8px;
				border-left: 3px solid #ff6a3a;
			}

			.payment-methods {
				display: grid;
				grid-template-columns: repeat(4, 1fr);
				gap: 12px;

				.payment-method {
					display: flex;
					align-items: center;
					justify-content: center;
					height: 48px;
					border: 1px solid #dcdfe6;
					border-radius: 4px;
					cursor: pointer;
					transition: all 0.3s;
					background-color: #fff;

					&:hover {
						border-color: #409eff;
					}

					&.active {
						border-color: #409eff;
						border-width: 2px;
					}

					.payment-icon {
						max-width: 60px;
						max-height: 24px;
						object-fit: contain;
					}
				}
			}
		}

		.pay-amount {
			margin-bottom: 16px;
			padding: 12px;
			background-color: #f5f7fa;
			border-radius: 4px;

			.label {
				font-size: 14px;
				color: #606266;
			}

			.value {
				font-size: 16px;
				font-weight: 600;
				color: #ff6a3a;
			}
		}

		.notice-section {
			margin-bottom: 16px;
			padding: 12px;
			background-color: #fef0f0;
			border-radius: 4px;

			.notice-title {
				font-size: 14px;
				font-weight: 500;
				color: #f56c6c;
				margin-bottom: 8px;
			}

			.notice-list {
				.notice-item {
					font-size: 12px;
					color: #f56c6c;
					line-height: 1.8;
				}
			}
		}

		.agreement {
			:deep(.el-checkbox__label) {
				font-size: 12px;
				color: #909399;
			}
		}
	}

	.dialog-footer {
		display: flex;
		justify-content: flex-end;
		gap: 12px;
	}
}

.withdraw-dialog {
	:deep(.el-dialog__header) {
		padding: 16px 20px;
		border-bottom: 1px solid #e4e7ed;
		margin-right: 0;
	}

	:deep(.el-dialog__body) {
		padding: 20px;
	}

	:deep(.el-dialog__footer) {
		padding: 12px 20px;
		border-top: 1px solid #e4e7ed;
	}

	.withdraw-content {
		.amount-row {
			display: flex;
			justify-content: space-between;
			align-items: center;
			margin-bottom: 20px;

			.amount-input-wrapper {
				display: flex;
				align-items: center;
				gap: 4px;

				.currency {
					font-size: 16px;
					color: #409eff;
				}

				.amount-input {
					width: 150px;

					:deep(.el-input__wrapper) {
						box-shadow: none;
						border-bottom: 1px solid #dcdfe6;
						border-radius: 0;
						padding: 0;
					}

					:deep(.el-input__inner) {
						font-size: 24px;
						font-weight: 600;
						color: #409eff;
						text-align: left;
					}
				}
			}

			.balance-info {
				display: flex;
				align-items: center;
				gap: 4px;

				.balance-label {
					font-size: 12px;
					color: #909399;
				}

				.balance-value {
					font-size: 12px;
					color: #ff6a3a;
				}
			}
		}

		.channel-section {
			margin-bottom: 20px;

			.section-title {
				font-size: 14px;
				color: #303133;
				margin-bottom: 12px;
				padding-left: 8px;
				border-left: 3px solid #ff6a3a;
			}

			.channel-methods {
				display: grid;
				grid-template-columns: repeat(4, 1fr);
				gap: 12px;

				.channel-method {
					display: flex;
					align-items: center;
					justify-content: center;
					height: 48px;
					border: 1px solid #dcdfe6;
					border-radius: 4px;
					cursor: pointer;
					transition: all 0.3s;
					background-color: #fff;

					&:hover {
						border-color: #409eff;
					}

					&.active {
						border-color: #409eff;
						border-width: 2px;
					}

					.channel-icon {
						max-width: 60px;
						max-height: 24px;
						object-fit: contain;
					}
				}
			}
		}

		.fee-info {
			margin-bottom: 16px;
			padding-left: 8px;
			border-left: 3px solid #ff6a3a;

			.fee-label {
				font-size: 14px;
				color: #303133;
			}

			.fee-value {
				font-size: 14px;
				color: #ff6a3a;
			}
		}

		.notice-section {
			margin-bottom: 16px;
			padding: 12px;
			background-color: #fef0f0;
			border-radius: 4px;

			.notice-title {
				font-size: 14px;
				font-weight: 500;
				color: #f56c6c;
				margin-bottom: 8px;
			}

			.notice-list {
				.notice-item {
					font-size: 12px;
					color: #f56c6c;
					line-height: 1.8;
				}
			}
		}

		.agreement {
			:deep(.el-checkbox__label) {
				font-size: 12px;
				color: #909399;
			}
		}
	}

	.dialog-footer {
		display: flex;
		justify-content: flex-end;
		gap: 12px;
	}
}
</style>
