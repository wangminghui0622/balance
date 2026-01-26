<template>
	<el-card class="stores-card">
		<template #header>
			<div class="card-header">
				<span class="header-title">
					<span class="title-bar"></span>
					<span>我的店铺 ({{ storeList.length }})</span>
				</span>
				<a href="#" class="more-link">
					更多
					<svg class="arrow-icon" viewBox="0 0 12 12" fill="none" xmlns="http://www.w3.org/2000/svg">
						<path d="M4.5 2.5L8 6L4.5 9.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
					</svg>
				</a>
			</div>
		</template>
		<div class="stores-list">
			<div v-for="(store, index) in storeList" :key="index" class="store-item">
				<el-avatar :size="40" shape="square" :src="store.avatar" />
				<div class="store-info">
					<div class="store-name">{{ store.name }}</div>
					<div class="store-status">
						<el-tag size="small" :type="store.isAuthorized ? 'success' : 'warning'">
							{{ store.isAuthorized ? '已授权' : '未授权' }}
						</el-tag>
					</div>
					<div class="store-id">4444店铺ID: {{ store.storeId }}</div>
				</div>
			</div>
		</div>
	</el-card>
	<el-card class="auth-card">
		<div class="auth-banner">
			<div class="auth-left">
				<div class="auth-title">
					<span class="title-bar"></span>
					<span>授权店铺获取更多收益！</span>
				</div>
				<div class="auth-desc">提升您的店铺收益潜力，立即开通返佣功能</div>
			</div>

			<div class="auth-right">
				<el-button class="auth-button" :loading="storeList[0]?.authLoading" @click="handleAuth(storeList[0])">
					{{ storeList[0]?.isAuthorized ? '重新授权' : '授权' }}
				</el-button>
			</div>
		</div>
	</el-card>
</template>

<script setup lang="ts">
	import { ref } from 'vue'
	import { ElMessage } from 'element-plus'
	import { shopeeApi } from '@share/api/shopee'

	interface Store {
		avatar : string
		name : string
		storeId : string
		shopId : number // Shopee Shop ID
		isAuthorized : boolean
		authLoading ?: boolean
	}

	const storeList = ref<Store[]>([
		{
			avatar: '',
			name: '店铺名435543称示例文字占位符',
			storeId: '1234567890',
			shopId: 226445936, // Shopee Shop ID
			isAuthorized: false,
			authLoading: false
		},
		{
			avatar: '',
			name: '店铺名称示例文字占位符',
			storeId: '1234567891',
			shopId: 226445937, // Shopee Shop ID
			isAuthorized: true,
			authLoading: false
		}
	])

	const handleAuth = async (store : Store) => {
		if (!store.shopId) {
			ElMessage.warning('店铺 Shop ID 未配置')
			return
		}

		store.authLoading = true
		try {
			const res = await shopeeApi.getAuthURL(store.shopId)

			if (res.code === 200 && res.auth_url) {
				// 在新窗口打开授权链接
				window.open(res.auth_url, '_blank')
				ElMessage.success('正在跳转到 Shopee 授权页面...')
			} else {
				ElMessage.error(res.message || '获取授权链接失败')
			}
		} catch (err : any) {
			ElMessage.error(err?.response?.data?.message || err?.message || '获取授权链接失败')
		} finally {
			store.authLoading = false
		}
	}
</script>

<style scoped lang="scss">
	.stores-card {
		border-radius: 8px;
		overflow: hidden;
	}

	.auth-card {
		border-radius: 8px;
		overflow: hidden;
		margin-top: 12px;
	}

	.card-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		font-weight: 500;
		font-size: 16px;
	}

	.header-title {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.card-header .title-bar {
		width: 3px;
		height: 16px;
		background-color: #ff6a3a;
		border-radius: 2px;
	}

	.more-link {
		display: flex;
		align-items: center;
		gap: 5px;
		font-size: 12px;
		font-weight: 400;
		line-height: 1;
		color: #909399;
		text-decoration: none;
		
		&:hover {
			color: #606266;
		}
		
		.arrow-icon {
			width: 12px;
			height: 12px;
		}
	}

	.stores-list {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	.store-item {
		display: flex;
		align-items: flex-start;
		gap: 12px;
		padding: 8px;
		border-radius: 4px;
		transition: background-color 0.3s;

		&:hover {
			background-color: #f5f7fa;
		}
	}

	.store-info {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.store-name {
		font-size: 14px;
		font-weight: 500;
		color: #303133;
		line-height: 1.4;
		max-width: 210px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.store-status {
		margin: 4px 0;
	}

	.store-id {
		font-size: 12px;
		color: #909399;
	}

	:deep(.auth-card .el-card__body) {
		padding: 0;
	}

	.auth-banner {
		background: linear-gradient(90deg, #fff5f0 0%, #ffffff 100%);
		border-radius: 8px;
		padding: 12px 2px;
		display: grid;
		grid-template-columns: minmax(0, 1fr) auto;
		align-items: center;
		column-gap: 0;
		width: 100%;
		box-sizing: border-box;
		overflow: hidden;
	}

	.auth-left {
		display: flex;
		flex-direction: column;
		gap: 6px;
		min-width: 0;
		overflow: hidden;
	}

	.auth-title {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 14px;
		font-weight: 600;
		color: #ff6a3a;
		line-height: 1.4;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;

		.title-bar {
			width: 3px;
			height: 16px;
			background-color: #ff6a3a;
			border-radius: 2px;
			flex-shrink: 0;
		}
	}

	.auth-desc {
		font-size: 11px;
		color: #909399;
		line-height: 1.5;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: clip;
	}
	.auth-right {
		justify-self: end;
	}

	.auth-button {
		background-color: #ff6a3a;
		border-color: #ff6a3a;
		color: #ffffff;
		padding: 1px 12px;
		border-radius: 12px;
		font-size: 12px;
		font-weight: 500;
		line-height: 1.2;
		flex-shrink: 0;

		&:hover {
			background-color: #ff8555;
			border-color: #ff8555;
		}

		&:active {
			background-color: #e65a30;
			border-color: #e65a30;
		}
	}
</style>