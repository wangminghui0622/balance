<template>
	<div class="finance-account-page">
		<!-- 页面标题 -->
		<div class="page-header">
			<h1 class="page-title">收款账户</h1>
		</div>

		<!-- 电子钱包 -->
		<div class="section">
			<div class="section-title-dot">电子钱包</div>
			<el-row :gutter="20">
				<el-col :xs="24" :sm="12" :md="8" :lg="8" :xl="6" v-for="(wallet, index) in eWallets" :key="'wallet-' + index">
					<div class="account-card wallet-card">
						<div class="card-header">
							<div class="logo-wrapper">
								<img v-if="wallet.logo" :src="wallet.logo" :alt="wallet.name" class="logo-img" />
								<el-icon v-else class="logo-icon"><Wallet /></el-icon>
							</div>
							<div class="wallet-info">
								<span class="wallet-name">{{ wallet.name }}</span>
								<el-tag size="small" :type="wallet.isConnected ? 'primary' : 'info'" class="status-tag">
									{{ wallet.isConnected ? '活跃' : '不活跃' }}
								</el-tag>
							</div>
							<div class="default-badge" v-if="defaultAccountId === wallet.id">
								<el-icon><CircleCheck /></el-icon>
								<span>默认</span>
							</div>
							<div v-else class="default-radio" @click.stop="handleChangeDefault(wallet.id, wallet.account)">
								<el-radio :model-value="defaultAccountId" :label="wallet.id">默认</el-radio>
							</div>
						</div>
						<div class="card-body">
							<div class="info-row">
								<div class="info-item">
									<div class="info-label">账号</div>
									<div class="info-value">{{ wallet.account || '-' }}</div>
								</div>
								<el-button v-if="!wallet.account" size="small" class="register-btn" @click.stop="handleWalletRegister(wallet)">注册/登录</el-button>
							</div>
						</div>
					</div>
				</el-col>
			</el-row>
		</div>

		<!-- 收款账户 -->
		<div class="section">
			<div class="section-title-dot">收款账户</div>
			<el-row :gutter="20">
				<el-col :xs="24" :sm="12" :md="8" :lg="8" :xl="6" v-for="(account, index) in bankAccounts" :key="'account-' + index">
					<div class="account-card bank-card" @click="handleViewAccount(account)">
						<div class="card-header">
							<div class="logo-wrapper">
								<img v-if="account.logo" :src="account.logo" :alt="account.name" class="logo-img" />
								<el-icon v-else class="logo-icon"><CreditCard /></el-icon>
							</div>
							<div class="wallet-info">
								<span class="wallet-name">{{ account.name }}</span>
								<el-tag size="small" :type="account.isActive ? 'success' : 'danger'" class="status-tag">
									{{ account.isActive ? '已激活' : '未激活' }}
								</el-tag>
							</div>
							<div class="default-badge" v-if="defaultAccountId === account.id">
								<el-icon><CircleCheck /></el-icon>
								<span>默认</span>
							</div>
							<div v-else class="default-radio" @click.stop="handleChangeDefault(account.id, account.account)">
								<el-radio :model-value="defaultAccountId" :label="account.id">默认</el-radio>
							</div>
						</div>
						<div class="card-body">
							<div class="info-row">
								<div class="info-item">
									<div class="info-label">账号</div>
									<div class="info-value">{{ account.account }}</div>
								</div>
								<div class="info-item">
									<div class="info-label">币种</div>
									<div class="info-value">{{ account.currency }}</div>
								</div>
							</div>
						</div>
					</div>
				</el-col>
				<!-- 新增收款账户 -->
				<el-col :xs="24" :sm="12" :md="8" :lg="8" :xl="6">
					<div class="add-account-card" @click="handleAddAccount">
						<el-icon class="add-icon"><FolderAdd /></el-icon>
						<span class="add-text">新增收款账户</span>
					</div>
				</el-col>
			</el-row>
		</div>
	</div>

	<!-- 更换默认账户确认对话框 -->
	<el-dialog
		v-model="changeDefaultDialogVisible"
		title="温馨提醒"
		width="400px"
		class="change-default-dialog"
		:close-on-click-modal="false"
	>
		<div class="reminder-content">
			<p>启用账户{{ pendingDefaultAccount }}@qq.com为默认收款账户,请注意,您上一周期已结算的款项已经/仍将打入原收款账户中。</p>
		</div>
		<template #footer>
			<div class="dialog-footer">
				<el-button type="primary" @click="confirmChangeDefault">确定</el-button>
			</div>
		</template>
	</el-dialog>

	<!-- 新增收款账户对话框 -->
	<el-dialog
		v-model="addAccountDialogVisible"
		title="新增收款账户"
		width="600px"
		class="add-account-dialog"
		:close-on-click-modal="false"
	>
		<el-form :model="accountForm" label-position="top" class="account-form">
			<el-row :gutter="20">
				<el-col :span="12">
					<el-form-item label="账户号码">
						<el-input v-model="accountForm.accountNo" placeholder="请输入">
							<template #suffix>
								<el-icon><Edit /></el-icon>
							</template>
						</el-input>
					</el-form-item>
				</el-col>
				<el-col :span="12">
					<el-form-item label="账户持有人姓名">
						<el-input v-model="accountForm.holderName" placeholder="请输入">
							<template #suffix>
								<el-icon><Edit /></el-icon>
							</template>
						</el-input>
					</el-form-item>
				</el-col>
			</el-row>
			<el-row :gutter="20">
				<el-col :span="12">
					<el-form-item label="币种">
						<el-select v-model="accountForm.currency" placeholder="请选择" style="width: 100%">
							<el-option label="CNY" value="CNY" />
							<el-option label="USD" value="USD" />
							<el-option label="TWD" value="TWD" />
							<el-option label="HKD" value="HKD" />
						</el-select>
					</el-form-item>
				</el-col>
				<el-col :span="12">
					<el-form-item label="SWIFT/BIC">
						<el-input v-model="accountForm.swiftCode" placeholder="请输入">
							<template #suffix>
								<el-icon><Edit /></el-icon>
							</template>
						</el-input>
					</el-form-item>
				</el-col>
			</el-row>
			<el-row :gutter="20">
				<el-col :span="12">
					<el-form-item label="银行或机构名称">
						<el-input v-model="accountForm.bankName" placeholder="请输入">
							<template #suffix>
								<el-icon><Edit /></el-icon>
							</template>
						</el-input>
					</el-form-item>
				</el-col>
				<el-col :span="12">
					<el-form-item label="银行代码">
						<el-input v-model="accountForm.bankCode" placeholder="请输入">
							<template #suffix>
								<el-icon><Edit /></el-icon>
							</template>
						</el-input>
					</el-form-item>
				</el-col>
			</el-row>
			<el-row :gutter="20">
				<el-col :span="12">
					<el-form-item label="分行代码">
						<el-input v-model="accountForm.branchCode" placeholder="请输入">
							<template #suffix>
								<el-icon><Edit /></el-icon>
							</template>
						</el-input>
					</el-form-item>
				</el-col>
				<el-col :span="12">
					<el-form-item label="银行所在国家/地区">
						<el-input v-model="accountForm.bankCountry" placeholder="请输入">
							<template #suffix>
								<el-icon><Edit /></el-icon>
							</template>
						</el-input>
					</el-form-item>
				</el-col>
			</el-row>
			<div class="agreement">
				<el-checkbox v-model="agreeAccountTerms">我已阅读示例文字占位符替换成可</el-checkbox>
			</div>
		</el-form>

		<template #footer>
			<div class="dialog-footer">
				<el-button @click="addAccountDialogVisible = false">取消</el-button>
				<el-button type="primary" :disabled="!agreeAccountTerms" @click="confirmAddAccount">添加</el-button>
			</div>
		</template>
	</el-dialog>

	<!-- 账户信息查看/编辑对话框 -->
	<el-dialog
		v-model="viewAccountDialogVisible"
		title="账户信息"
		width="900px"
		class="view-account-dialog"
		:close-on-click-modal="false"
	>
		<div class="account-detail">
			<div class="detail-item">
				<div class="detail-label">账户号码：</div>
				<div class="detail-value">{{ viewAccountData.accountNo }}</div>
			</div>
			<div class="detail-item">
				<div class="detail-label">账户持有人姓名：</div>
				<div class="detail-value">{{ viewAccountData.holderName }}</div>
			</div>
			<div class="detail-item">
				<div class="detail-label">币种：</div>
				<div class="detail-value">{{ viewAccountData.currency }}</div>
			</div>
			<div class="detail-item">
				<div class="detail-label">SEIFT/BIC：</div>
				<div class="detail-value">{{ viewAccountData.swiftCode }}</div>
			</div>
			<div class="detail-item">
				<div class="detail-label">银行或机构名称：</div>
				<div class="detail-value">{{ viewAccountData.bankName }}</div>
			</div>
			<div class="detail-item">
				<div class="detail-label">银行代码：</div>
				<div class="detail-value">{{ viewAccountData.bankCode }}</div>
			</div>
			<div class="detail-item">
				<div class="detail-label">分行代码：</div>
				<div class="detail-value">{{ viewAccountData.branchCode }}</div>
			</div>
			<div class="detail-item">
				<div class="detail-label">银行所在国家/地区：</div>
				<div class="detail-value">{{ viewAccountData.bankCountry }}</div>
			</div>
			<div class="detail-item">
				<div class="detail-label">银行或机构地址：</div>
				<div class="detail-value">{{ viewAccountData.bankAddress }}</div>
			</div>
			<div class="detail-item full-width">
				<div class="detail-label">备注：</div>
				<div class="detail-value">{{ viewAccountData.remark }}</div>
			</div>
		</div>

		<template #footer>
			<div class="dialog-footer">
				<el-button type="danger" link @click="handleDeleteAccount">删除账户</el-button>
				<div class="right-btns">
					<el-button @click="handleEditAccount">编辑</el-button>
					<el-button type="primary" @click="viewAccountDialogVisible = false">确定</el-button>
				</div>
			</div>
		</template>
	</el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { CircleCheck, Wallet, CreditCard, FolderAdd, Edit } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

interface WalletAccount {
	id: string
	name: string
	logo?: string
	account: string
	receiver: string
	isConnected: boolean
	isDefault: boolean
}

interface BankAccount {
	id: string
	name: string
	logo?: string
	account: string
	currency: string
	isActive: boolean
	isDefault: boolean
}

const defaultAccountId = ref('wallet-1')
const pendingDefaultId = ref('')
const pendingDefaultAccount = ref('')
const changeDefaultDialogVisible = ref(false)

const eWallets = ref<WalletAccount[]>([
	{
		id: 'wallet-1',
		name: 'PayPal支付',
		account: '12345678910121314l5',
		receiver: '123456789101',
		isConnected: true,
		isDefault: true
	},
	{
		id: 'wallet-2',
		name: 'PayPal支付',
		account: '',
		receiver: '',
		isConnected: false,
		isDefault: false
	}
])

const bankAccounts = ref<BankAccount[]>([
	{
		id: 'bank-1',
		name: '汇丰银行>',
		account: '12345678910121314l5',
		currency: 'CNY',
		isActive: false,
		isDefault: false
	}
])

const addAccountDialogVisible = ref(false)
const agreeAccountTerms = ref(false)

const accountForm = ref({
	accountNo: '',
	holderName: '',
	currency: '',
	swiftCode: '',
	bankName: '',
	bankCode: '',
	branchCode: '',
	bankCountry: ''
})

const handleAddAccount = () => {
	accountForm.value = {
		accountNo: '',
		holderName: '',
		currency: '',
		swiftCode: '',
		bankName: '',
		bankCode: '',
		branchCode: '',
		bankCountry: ''
	}
	agreeAccountTerms.value = false
	addAccountDialogVisible.value = true
}

const confirmAddAccount = () => {
	if (!agreeAccountTerms.value) {
		ElMessage.warning('请先同意协议')
		return
	}
	ElMessage.success('收款账户添加成功')
	addAccountDialogVisible.value = false
}

const viewAccountDialogVisible = ref(false)
const viewAccountData = ref({
	accountNo: '',
	holderName: '',
	currency: '',
	swiftCode: '',
	bankName: '',
	bankCode: '',
	branchCode: '',
	bankCountry: '',
	bankAddress: '',
	remark: ''
})

const handleViewAccount = (account: BankAccount) => {
	viewAccountData.value = {
		accountNo: account.account || '1234567890',
		holderName: '示例文字',
		currency: account.currency || '示例文字',
		swiftCode: '示例文字',
		bankName: '例文字示例文字示例文字示例文字adbcadsf示例文字示例文字',
		bankCode: '1234567890',
		branchCode: '1234567890',
		bankCountry: '示例文字示例文字示例文字',
		bankAddress: '例文字示例文字示例文字示例文字adbcadsf示例文字示例文字',
		remark: '示例文字示例文字示例文字示例文字示例\n文字示例文字示例文字'
	}
	viewAccountDialogVisible.value = true
}

const handleEditAccount = () => {
	ElMessage.info('编辑功能开发中...')
}

const handleDeleteAccount = () => {
	ElMessage.warning('删除功能开发中...')
}

const handleChangeDefault = (id: string, account: string) => {
	pendingDefaultId.value = id
	pendingDefaultAccount.value = account
	changeDefaultDialogVisible.value = true
}

const confirmChangeDefault = () => {
	defaultAccountId.value = pendingDefaultId.value
	changeDefaultDialogVisible.value = false
	ElMessage.success('默认收款账户已更改')
}

const handleWalletRegister = (wallet: WalletAccount) => {
	ElMessage.info(`${wallet.name} 注册/登录功能开发中...`)
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
	background-color: #fff;
	border-radius: 8px;
	padding: 20px;
	border: 1px solid #e4e7ed;

	.section-title-dot {
		font-size: 14px;
		font-weight: 500;
		color: #303133;
		margin-bottom: 16px;
		padding-left: 10px;
		position: relative;

		&::before {
			content: '';
			position: absolute;
			left: 0;
			top: 50%;
			transform: translateY(-50%);
			width: 4px;
			height: 14px;
			background-color: #ff6a3a;
			border-radius: 2px;
		}
	}
}

.account-card {
	border-radius: 8px;
	overflow: hidden;
	margin-bottom: 20px;

	&.wallet-card {
		.card-header {
			background-color: #409eff;
		}
	}

	&.bank-card {
		.card-header {
			background-color: #f56c6c;
		}
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

			.logo-img {
				width: 32px;
				height: 32px;
				object-fit: contain;
			}

			.logo-icon {
				font-size: 24px;
				color: #409eff;
			}
		}

		.wallet-info {
			flex: 1;
			display: flex;
			flex-direction: column;
			gap: 4px;

			.wallet-name {
				font-size: 14px;
				font-weight: 500;
			}

			.status-tag {
				width: fit-content;
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
	}

	.card-body {
		padding: 16px;
		background-color: #fff;
		border: 1px solid #e4e7ed;
		border-top: none;
		border-radius: 0 0 8px 8px;

		.info-row {
			display: flex;
			justify-content: space-between;
			align-items: center;
		}

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

		.register-btn {
			background-color: #f5f7fa;
			border-color: #dcdfe6;
			color: #606266;

			&:hover {
				background-color: #e4e7ed;
				border-color: #c0c4cc;
			}
		}
	}
}

.bank-card {
	.logo-wrapper {
		.logo-icon {
			color: #f56c6c !important;
		}
	}
}

.add-account-card {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	height: 140px;
	border: 1px dashed #dcdfe6;
	border-radius: 8px;
	cursor: pointer;
	transition: all 0.3s;
	background-color: #fafafa;

	&:hover {
		border-color: #409eff;
		background-color: #ecf5ff;

		.add-icon,
		.add-text {
			color: #409eff;
		}
	}

	.add-icon {
		font-size: 32px;
		color: #909399;
		margin-bottom: 8px;
	}

	.add-text {
		font-size: 14px;
		color: #909399;
	}
}

.change-default-dialog {
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

	.reminder-content {
		p {
			font-size: 14px;
			color: #606266;
			line-height: 1.8;
			margin: 0;
		}
	}

	.dialog-footer {
		display: flex;
		justify-content: flex-end;

		.el-button {
			background-color: #ff6a3a;
			border-color: #ff6a3a;

			&:hover {
				background-color: #ff8c5a;
				border-color: #ff8c5a;
			}
		}
	}
}

.view-account-dialog {
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

	.account-detail {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 16px 40px;

		.detail-item {
			display: flex;
			white-space: nowrap;

			.detail-label {
				font-size: 14px;
				color: #909399;
				flex-shrink: 0;
			}

			.detail-value {
				font-size: 14px;
				color: #303133;
			}
		}

		.full-width {
			grid-column: 1 / -1;
		}
	}

	.dialog-footer {
		display: flex;
		justify-content: space-between;
		align-items: center;

		.right-btns {
			display: flex;
			gap: 12px;
		}
	}
}

.add-account-dialog {
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

	.account-form {
		:deep(.el-form-item__label) {
			font-size: 14px;
			color: #606266;
			padding-bottom: 8px;
		}

		:deep(.el-input__suffix) {
			.el-icon {
				color: #909399;
			}
		}

		.agreement {
			margin-top: 16px;

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
