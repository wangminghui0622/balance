<template>
	<div class="finance-account-page">
		<!-- 页面标题 -->
		<div class="page-header">
			<h1 class="page-title">收款账户</h1>
		</div>

		<!-- 电子钱包 -->
		<div class="section">
			<div class="section-title">电子钱包</div>
			<el-row :gutter="20">
				<el-col :xs="24" :sm="12" :md="8" :lg="8" :xl="6" v-for="(wallet, index) in eWallets" :key="'wallet-' + index">
					<div class="account-card" :class="{ 'is-active': wallet.isActive }">
						<div class="card-header">
							<div class="logo-wrapper">
								<span class="logo-text">logo</span>
							</div>
							<div class="wallet-info">
								<span class="wallet-name">{{ wallet.name }}</span>
								<span class="status-text">{{ wallet.isActive ? '活跃' : '不活跃' }}</span>
							</div>
							<div class="default-badge" v-if="wallet.isDefault">
								<el-icon><CircleCheck /></el-icon>
								<span>默认</span>
							</div>
							<el-radio v-else v-model="defaultWallet" :label="wallet.id" class="default-radio">默认</el-radio>
						</div>
						<div class="card-body">
							<div class="info-item">
								<div class="info-label">账号</div>
								<div class="info-value">{{ wallet.account }}</div>
							</div>
							<div class="info-item">
								<div class="info-label">收款人</div>
								<div class="info-value">{{ wallet.receiver }}</div>
							</div>
						</div>
					</div>
				</el-col>
			</el-row>
		</div>

		<!-- 三方支付 -->
		<div class="section">
			<div class="section-title">三方支付</div>
			<el-row :gutter="20">
				<el-col :xs="24" :sm="12" :md="8" :lg="8" :xl="6" v-for="(payment, index) in thirdPartyPayments" :key="'payment-' + index">
					<div class="account-card" :class="{ 'is-bound': payment.isBound, 'is-active': payment.isActive }">
						<div class="card-header">
							<div class="logo-wrapper">
								<span class="logo-text">logo</span>
							</div>
							<div class="wallet-info">
								<span class="wallet-name">{{ payment.name }}</span>
								<span v-if="payment.isBound" class="status-text">{{ payment.isActive ? '活跃' : '不活跃' }}</span>
							</div>
							<template v-if="payment.isBound">
								<div class="default-badge" v-if="payment.isDefault">
									<el-icon><CircleCheck /></el-icon>
									<span>默认</span>
								</div>
								<el-radio v-else v-model="defaultPayment" :label="payment.id" class="default-radio">默认</el-radio>
							</template>
							<el-button v-else size="small" type="primary" class="bind-btn" @click="handleBind(payment)">注册/登录</el-button>
						</div>
						<div class="card-body" v-if="payment.isBound">
							<div class="info-item">
								<div class="info-label">账号</div>
								<div class="info-value">{{ payment.account }}</div>
							</div>
							<div class="info-item">
								<div class="info-label">收款人</div>
								<div class="info-value">{{ payment.receiver }}</div>
							</div>
						</div>
					</div>
				</el-col>
			</el-row>
		</div>
	</div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { CircleCheck } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

interface WalletAccount {
	id: string
	name: string
	account: string
	receiver: string
	isActive: boolean
	isDefault: boolean
}

interface ThirdPartyPayment {
	id: string
	name: string
	account?: string
	receiver?: string
	isActive?: boolean
	isDefault?: boolean
	isBound: boolean
}

const defaultWallet = ref('wallet-1')
const defaultPayment = ref('payment-1')

const eWallets = ref<WalletAccount[]>([
	{
		id: 'wallet-1',
		name: '钱包名称示例文字',
		account: '12345678910121314l5',
		receiver: '123456789101',
		isActive: true,
		isDefault: true
	}
])

const thirdPartyPayments = ref<ThirdPartyPayment[]>([
	{
		id: 'payment-1',
		name: '钱包名称示例文字',
		account: '12345678910121314l5',
		receiver: '123456789101',
		isActive: false,
		isDefault: false,
		isBound: true
	},
	{
		id: 'payment-2',
		name: '钱包名称示例文字',
		isBound: false
	},
	{
		id: 'payment-3',
		name: '钱包名称示例文字',
		isBound: false
	}
])

const handleBind = (payment: ThirdPartyPayment) => {
	ElMessage.info(`正在跳转到 ${payment.name} 注册/登录页面...`)
}
</script>

<style scoped lang="scss">
.finance-account-page {
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

.section {
	margin-bottom: 32px;

	.section-title {
		font-size: 16px;
		font-weight: 500;
		color: #303133;
		margin-bottom: 16px;
	}
}

.account-card {
	background-color: #ffb088;
	border-radius: 8px;
	overflow: hidden;
	margin-bottom: 20px;

	&.is-active {
		background-color: #ff6a3a;
	}

	.card-header {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 16px;
		color: #fff;

		.logo-wrapper {
			width: 48px;
			height: 48px;
			background-color: #fff;
			border-radius: 50%;
			display: flex;
			align-items: center;
			justify-content: center;
			flex-shrink: 0;

			.logo-text {
				font-size: 12px;
				color: #909399;
			}
		}

		.wallet-info {
			flex: 1;
			display: flex;
			align-items: center;
			gap: 8px;

			.wallet-name {
				font-size: 14px;
				font-weight: 500;
			}

			.status-text {
				font-size: 12px;
				color: rgba(255, 255, 255, 0.8);
			}
		}

		.default-badge {
			display: flex;
			align-items: center;
			gap: 4px;
			font-size: 12px;
			color: rgba(255, 255, 255, 0.9);

			.el-icon {
				font-size: 14px;
			}
		}

		.default-radio {
			:deep(.el-radio__label) {
				color: rgba(255, 255, 255, 0.7);
			}
		}

		.bind-btn {
			margin-left: auto;
		}
	}

	.card-body {
		display: flex;
		gap: 32px;
		padding: 16px;
		background-color: #fff;

		.info-item {
			.info-label {
				font-size: 12px;
				color: #909399;
				margin-bottom: 4px;
			}

			.info-value {
				font-size: 14px;
				color: #303133;
				font-weight: 500;
			}
		}
	}
}
</style>
