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

	<!-- 保证金充值对话框 -->
	<el-dialog
		v-model="rechargeDialogVisible"
		title="保证金充值"
		width="600px"
		class="deposit-recharge-dialog"
		:close-on-click-modal="false"
	>
		<div class="recharge-content">
			<!-- 注意事项 -->
			<div class="notice-section">
				<div class="notice-title">注意事项：</div>
				<div class="notice-list">
					<div class="notice-item">1. 示例文字占位符，替换成可示例文字占位符，替换成可示例文字占位符，替换成可</div>
					<div class="notice-item">2. 示例文字占位符，替换成可示例文字占位符，替换成可示例文字占位符，替换成可</div>
					<div class="notice-item">3. 示例文字占位符，替换成可示例文字占位符，替换成可示例文字占位符，替换成可</div>
				</div>
			</div>

			<!-- 保证金充值区域 -->
			<div class="deposit-form-section">
				<div class="section-title">保证金充值</div>
				<div class="form-content">
					<!-- 金额输入 -->
					<div class="amount-input-wrapper">
						<span class="currency">NT$</span>
						<el-input-number
							v-model="depositRechargeAmount"
							:min="0"
							:precision="2"
							:controls="false"
							placeholder="请输入充值金额"
							class="amount-input"
						/>
					</div>

					<!-- 充值公司 -->
					<div class="form-row">
						<span class="form-label">充值公司：</span>
						<span class="form-value">公司名称示例文字占位符</span>
					</div>

					<!-- 保证金门槛 -->
					<div class="form-row">
						<span class="form-label">保证金门槛：</span>
						<span class="form-value">¥3,000.00或以上</span>
					</div>

					<!-- 协议复选框 -->
					<div class="agreement">
						<el-checkbox v-model="agreeDepositTerms">我已阅读示例文字占位符替换成可示例文字占位符替换成可</el-checkbox>
					</div>

					<!-- 去支付按钮 -->
					<div class="pay-btn-wrapper">
						<el-button type="primary" :disabled="!agreeDepositTerms || depositRechargeAmount <= 0" @click="confirmDepositRecharge">去支付</el-button>
					</div>
				</div>
			</div>
		</div>
	</el-dialog>

	<!-- 结账对话框 -->
	<el-dialog
		v-model="checkoutDialogVisible"
		title="结账"
		width="400px"
		class="checkout-dialog"
		:close-on-click-modal="false"
	>
		<div class="checkout-content">
			<!-- 店主保证金信息 -->
			<div class="deposit-info">
				<div class="info-left">
					<div class="deposit-icon">
						<img src="" alt="" class="icon-img" />
					</div>
					<span class="deposit-name">店主保证金</span>
				</div>
				<div class="info-right">
					<span class="deposit-amount">¥3,000.00</span>
				</div>
			</div>

			<!-- 支付方式 -->
			<div class="payment-section">
				<div class="section-title">支付方式：</div>
				<div class="payment-methods">
					<div
						v-for="method in checkoutPaymentMethods"
						:key="method.id"
						class="payment-method"
						:class="{ active: selectedCheckoutPayment === method.id }"
						@click="selectedCheckoutPayment = method.id"
					>
						<img :src="method.icon" :alt="method.name" class="payment-icon" />
					</div>
				</div>
			</div>

			<!-- 支付金额 -->
			<div class="pay-amount">
				<span class="label">支付金额：</span>
				<span class="value">¥3,000.00</span>
			</div>
		</div>

		<template #footer>
			<div class="dialog-footer">
				<el-button @click="checkoutDialogVisible = false">取消</el-button>
				<el-button type="primary" :disabled="!selectedCheckoutPayment" @click="confirmCheckout">确定</el-button>
			</div>
		</template>
	</el-dialog>

	<!-- 支付详情对话框 -->
	<el-dialog
		v-model="paymentDetailVisible"
		width="500px"
		class="payment-detail-dialog"
		:close-on-click-modal="false"
		:show-close="false"
	>
		<template #header>
			<div class="custom-header">
				<el-button link @click="goBackToCheckout">
					<el-icon><ArrowLeft /></el-icon>
				</el-button>
				<span class="header-title">结账</span>
				<el-button link @click="paymentDetailVisible = false">
					<el-icon><Close /></el-icon>
				</el-button>
			</div>
		</template>

		<div class="payment-detail-content">
			<div class="content-layout">
				<!-- 左侧信息区域 -->
				<div class="left-section">
					<!-- 付款信息 -->
					<div class="payment-info">
						<div class="info-row">
							<span class="info-label">付款总额</span>
							<span class="info-value">¥3,000.00</span>
						</div>
						<div class="info-row">
							<span class="info-label">付款方式</span>
							<span class="info-value">方式占位符</span>
						</div>
					</div>

					<!-- 操作说明 -->
					<div class="instructions">
						<div class="instructions-title">操作说明：</div>
						<div class="instructions-list">
							<div class="instruction-item">1. 示例文字占位符，替换成可示例文字占位符</div>
							<div class="instruction-item">2. 示例文字占位符，替换成可示例文字占位符</div>
							<div class="instruction-item">3. 示例文字占位符，替换成可示例文字占位符</div>
						</div>
					</div>

					<!-- 协议复选框 -->
					<div class="agreement">
						<el-checkbox v-model="agreePaymentTerms">我已阅读示例文字占位符替换成可</el-checkbox>
					</div>
				</div>

				<!-- 右侧二维码区域 -->
				<div class="right-section">
					<div class="qrcode-wrapper" v-if="!qrcodeLoadFailed">
						<img 
							src="https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=payment_demo" 
							alt="支付二维码" 
							class="qrcode-img"
							@error="qrcodeLoadFailed = true"
						/>
					</div>
					<div class="qrcode-failed" v-else>
						<div class="failed-icon">
							<el-icon :size="32" color="#c0c4cc"><Picture /></el-icon>
						</div>
						<div class="failed-text">加载失败</div>
						<div class="failed-url">{{ paymentUrl }}</div>
						<el-button size="small" @click="copyPaymentUrl">复制</el-button>
					</div>
					<div class="qrcode-tip" v-if="!qrcodeLoadFailed">扫码二维码</div>
				</div>
			</div>
		</div>

		<template #footer>
			<div class="dialog-footer">
				<el-button @click="paymentDetailVisible = false">取消</el-button>
				<el-button type="primary" :disabled="!agreePaymentTerms" @click="confirmPayment">确定</el-button>
			</div>
		</template>
	</el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { Search, ArrowLeft, Close, Picture } from '@element-plus/icons-vue'
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

const rechargeDialogVisible = ref(false)
const depositRechargeAmount = ref(0)
const agreeDepositTerms = ref(false)

const handleRecharge = () => {
	depositRechargeAmount.value = 0
	agreeDepositTerms.value = false
	rechargeDialogVisible.value = true
}

const checkoutDialogVisible = ref(false)
const selectedCheckoutPayment = ref('')

const checkoutPaymentMethods = [
	{ id: 'paypal', name: 'PayPal', icon: 'https://www.paypalobjects.com/webstatic/mktg/Logo/pp-logo-100px.png' },
	{ id: 'alipay', name: '支付宝', icon: 'https://gw.alipayobjects.com/mdn/rms_0c75a8/afts/img/A*V3ICRJ-4bDcAAAAAAAAAAAAAARQnAQ' },
	{ id: 'linepay', name: 'LINE Pay', icon: 'https://scdn.line-apps.com/linepay/portal/v-240930/portal/img/sp/sp_logo_linepay_white.png' },
	{ id: 'visa', name: 'VISA', icon: 'https://upload.wikimedia.org/wikipedia/commons/thumb/5/5e/Visa_Inc._logo.svg/200px-Visa_Inc._logo.svg.png' }
]

const confirmDepositRecharge = () => {
	if (depositRechargeAmount.value <= 0) {
		ElMessage.warning('请输入充值金额')
		return
	}
	if (!agreeDepositTerms.value) {
		ElMessage.warning('请先同意协议')
		return
	}
	rechargeDialogVisible.value = false
	selectedCheckoutPayment.value = ''
	checkoutDialogVisible.value = true
}

const paymentDetailVisible = ref(false)
const agreePaymentTerms = ref(false)
const qrcodeLoadFailed = ref(false)
const paymentUrl = ref('https://example.com/pay?id=xxxxx')

const copyPaymentUrl = () => {
	navigator.clipboard.writeText(paymentUrl.value)
	ElMessage.success('链接已复制')
}

const confirmCheckout = () => {
	if (!selectedCheckoutPayment.value) {
		ElMessage.warning('请选择支付方式')
		return
	}
	checkoutDialogVisible.value = false
	agreePaymentTerms.value = false
	paymentDetailVisible.value = true
}

const goBackToCheckout = () => {
	paymentDetailVisible.value = false
	checkoutDialogVisible.value = true
}

const confirmPayment = () => {
	if (!agreePaymentTerms.value) {
		ElMessage.warning('请先同意协议')
		return
	}
	ElMessage.success('支付成功')
	paymentDetailVisible.value = false
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

.deposit-recharge-dialog {
	:deep(.el-dialog__header) {
		padding: 16px 20px;
		border-bottom: 1px solid #e4e7ed;
		margin-right: 0;
	}

	:deep(.el-dialog__body) {
		padding: 20px;
	}

	.recharge-content {
		.notice-section {
			margin-bottom: 20px;
			padding: 16px;
			background-color: #fef0f0;
			border-radius: 8px;

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

		.deposit-form-section {
			border: 1px solid #e4e7ed;
			border-radius: 8px;
			overflow: hidden;

			.section-title {
				font-size: 14px;
				font-weight: 500;
				color: #303133;
				padding: 12px 16px;
				background-color: #f5f7fa;
				border-bottom: 1px solid #e4e7ed;
				position: relative;
				padding-left: 24px;

				&::before {
					content: '';
					position: absolute;
					left: 12px;
					top: 50%;
					transform: translateY(-50%);
					width: 4px;
					height: 14px;
					background-color: #ff6a3a;
					border-radius: 2px;
				}
			}

			.form-content {
				padding: 20px;

				.amount-input-wrapper {
					display: flex;
					align-items: center;
					gap: 4px;
					margin-bottom: 20px;

					.currency {
						font-size: 16px;
						color: #409eff;
					}

					.amount-input {
						width: 180px;

						:deep(.el-input__wrapper) {
							box-shadow: none;
							border-bottom: 1px solid #dcdfe6;
							border-radius: 0;
							padding: 0;
						}

						:deep(.el-input__inner) {
							font-size: 28px;
							font-weight: 600;
							color: #409eff;
							text-align: left;
						}
					}
				}

				.form-row {
					margin-bottom: 12px;
					font-size: 14px;

					.form-label {
						color: #909399;
					}

					.form-value {
						color: #606266;
					}
				}

				.agreement {
					margin-top: 20px;
					margin-bottom: 20px;

					:deep(.el-checkbox__label) {
						font-size: 12px;
						color: #909399;
					}
				}

				.pay-btn-wrapper {
					.el-button {
						background-color: #ff6a3a;
						border-color: #ff6a3a;

						&:hover {
							background-color: #ff8c5a;
							border-color: #ff8c5a;
						}

						&:disabled {
							background-color: #fab6a0;
							border-color: #fab6a0;
						}
					}
				}
			}
		}
	}
}

.payment-detail-dialog {
	:deep(.el-dialog__header) {
		padding: 0;
		margin-right: 0;
	}

	:deep(.el-dialog__body) {
		padding: 20px;
	}

	:deep(.el-dialog__footer) {
		padding: 12px 20px;
		border-top: 1px solid #e4e7ed;
	}

	.custom-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 12px 16px;
		border-bottom: 1px solid #e4e7ed;

		.header-title {
			font-size: 16px;
			font-weight: 500;
			color: #303133;
		}
	}

	.payment-detail-content {
		.content-layout {
			display: flex;
			gap: 24px;

			.left-section {
				flex: 1;

				.payment-info {
					margin-bottom: 20px;

					.info-row {
						display: flex;
						justify-content: space-between;
						align-items: center;
						padding: 12px 0;
						border-bottom: 1px solid #f5f5f5;

						&:last-child {
							border-bottom: none;
						}

						.info-label {
							font-size: 14px;
							color: #909399;
						}

						.info-value {
							font-size: 14px;
							color: #303133;
						}
					}
				}

				.instructions {
					margin-bottom: 16px;

					.instructions-title {
						font-size: 14px;
						color: #909399;
						margin-bottom: 8px;
					}

					.instructions-list {
						.instruction-item {
							font-size: 12px;
							color: #909399;
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

			.right-section {
				display: flex;
				flex-direction: column;
				align-items: center;

				.qrcode-wrapper {
					width: 160px;
					height: 160px;
					border: 1px solid #e4e7ed;
					border-radius: 8px;
					padding: 8px;
					margin-bottom: 8px;

					.qrcode-img {
						width: 100%;
						height: 100%;
						object-fit: contain;
					}
				}

				.qrcode-failed {
					display: flex;
					flex-direction: column;
					align-items: center;
					width: 160px;
					padding: 16px;
					border: 1px solid #e4e7ed;
					border-radius: 8px;
					background-color: #fafafa;

					.failed-icon {
						margin-bottom: 8px;
					}

					.failed-text {
						font-size: 12px;
						color: #909399;
						margin-bottom: 12px;
					}

					.failed-url {
						font-size: 10px;
						color: #909399;
						word-break: break-all;
						text-align: center;
						margin-bottom: 12px;
						line-height: 1.4;
					}
				}

				.qrcode-tip {
					font-size: 12px;
					color: #909399;
				}
			}
		}
	}

	.dialog-footer {
		display: flex;
		justify-content: flex-end;
		gap: 12px;
	}
}

.checkout-dialog {
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

	.checkout-content {
		.deposit-info {
			display: flex;
			justify-content: space-between;
			align-items: center;
			padding: 16px;
			background-color: #f5f7fa;
			border-radius: 8px;
			margin-bottom: 20px;

			.info-left {
				display: flex;
				align-items: center;
				gap: 12px;

				.deposit-icon {
					width: 40px;
					height: 40px;
					background-color: #ff6a3a;
					border-radius: 8px;
					display: flex;
					align-items: center;
					justify-content: center;

					.icon-img {
						width: 24px;
						height: 24px;
					}
				}

				.deposit-name {
					font-size: 14px;
					font-weight: 500;
					color: #303133;
				}
			}

			.info-right {
				.deposit-amount {
					font-size: 16px;
					font-weight: 600;
					color: #303133;
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
						max-width: 50px;
						max-height: 20px;
						object-fit: contain;
					}
				}
			}
		}

		.pay-amount {
			text-align: right;

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
	}

	.dialog-footer {
		display: flex;
		justify-content: flex-end;
		gap: 12px;
	}
}
</style>
