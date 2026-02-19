<template>
	<el-card class="stores-card">
		<template #header>
			<div class="card-header">
				<span class="header-title" @click="goToStores">
					<span class="title-bar"></span>
					<span>我的店铺 ({{ storeList.length }})</span>
				</span>
				<span class="more-link" @click="goToStores">
					更多
					<svg class="arrow-icon" viewBox="0 0 12 12" fill="none" xmlns="http://www.w3.org/2000/svg">
						<path d="M4.5 2.5L8 6L4.5 9.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
					</svg>
				</span>
			</div>
		</template>
		<div class="stores-list">
			<div v-if="!isShopOwner" class="store-item">
				<div class="store-info">
					<div class="store-name">此功能仅限店主类型用户使用</div>
				</div>
			</div>
			<div v-else-if="loading" class="store-item">
				<div class="store-info">
					<div class="store-name">加载中...</div>
				</div>
			</div>
			<div v-else-if="storeList.length === 0" class="store-item">
				<div class="store-info">
					<div class="store-name">暂无店铺，请先授权</div>
				</div>
			</div>
			<template v-else>
				<div v-for="(store, index) in storeList" :key="`store-${store.shopId}-${index}`" class="store-item">
					<el-avatar :size="40" shape="square" :src="store.avatar" />
					<div class="store-info">
						<div class="store-name">{{ store.name || '未知店铺' }}</div>
						<div class="store-status">
							<div class="store-status-row">
								<el-tag size="small" :type="store.isAuthorized ? 'success' : 'warning'">
									{{ store.isAuthorized ? '已授权' : '未授权' }}
								</el-tag>
								<span class="store-id-inline">店铺编号: {{ store.shopId || '--' }}</span>
							</div>
						</div>
					</div>
				</div>
			</template>
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

			<div class="auth-right" v-if="isShopOwner">
				<el-button 
					class="auth-button" 
					:loading="loading || (storeList[0]?.authLoading ?? false)" 
					@click="handleAuth(storeList[0] || { shopId: 0, isAuthorized: false, authLoading: false } as Store)"
				>
					授权
				</el-button>
			</div>
		</div>
	</el-card>
</template>

<script setup lang="ts">
	import { ref, onMounted, onActivated, computed } from 'vue'
	import { useRouter, useRoute } from 'vue-router'
	import { ElMessage } from 'element-plus'
	import { shopeeApi, type ShopeeShop } from '@share/api/shopee'
	import { STORAGE_KEYS, ROUTE_PATH, USER_TYPE_NUM, HTTP_STATUS } from '@share/constants'

	interface Store {
		avatar : string
		name : string
		storeId : string
		shopId : number // Shopee Shop ID
		isAuthorized : boolean
		authLoading ?: boolean
		shopData ?: ShopeeShop // 完整的店铺数据
	}

	const storeList = ref<Store[]>([])
	const loading = ref(false)
	const router = useRouter()
	const route = useRoute()

	// 检查是否为店主类型（userType=1）
	const isShopOwner = computed(() => {
		const userType = localStorage.getItem(STORAGE_KEYS.USER_TYPE)
		return userType === USER_TYPE_NUM.SHOPOWNER.toString()
	})

	// 获取店铺列表
	const fetchShopList = async (forceRefresh: boolean = false) => {
		// 只有店主类型才能调用此接口
		if (!isShopOwner.value) {
			ElMessage.warning('此功能仅限店主类型用户使用')
			return
		}

		const token = localStorage.getItem(STORAGE_KEYS.TOKEN)
		if (!token) {
			ElMessage.warning('请先登录')
			return
		}

		// 如果是强制刷新，先清空旧数据
		if (forceRefresh) {
			storeList.value = []
		}

		loading.value = true
		try {
			const res = await shopeeApi.getShopList()
			
			if (res.code === HTTP_STATUS.OK) {
				const shopList = res.data?.list || []
				if (shopList.length > 0) {
					storeList.value = shopList.map((shop: ShopeeShop) => ({
						avatar: shop.shopSlug || '',
						name: shop.shopName || `店铺 ${shop.shopId}`,
						storeId: shop.shopIdStr || shop.shopId?.toString() || '',
						shopId: shop.shopId,
						isAuthorized: shop.authStatus === 1,
						authLoading: false,
						shopData: shop
					}))
				} else {
					storeList.value = []
				}
			} else {
				ElMessage.error(res.message || '获取店铺列表失败')
			}
		} catch (err: any) {
			console.error('获取店铺列表失败:', err)
			ElMessage.error(err?.response?.data?.message || err?.message || '获取店铺列表失败')
		} finally {
			loading.value = false
		}
	}

	// 组件挂载时获取店铺列表（但如果有 refresh 参数，让父组件控制刷新）
	onMounted(() => {
		// 检查路由参数，如果有 refresh 参数，不在这里执行，让父组件控制
		if (route.query.refresh !== 'true') {
			fetchShopList()
		}
	})

	// 组件激活时刷新店铺列表（用于从授权回调页面返回时刷新）
	onActivated(() => {
		// 如果有 refresh 参数，让父组件控制刷新
		if (route.query.refresh !== 'true') {
			fetchShopList()
		}
	})


	// 点击"我的店铺"或"更多"时，跳转到店铺列表页面
	const goToStores = (event?: Event) => {
		if (event) {
			event.preventDefault()
			event.stopPropagation()
		}
		router.push('/shopowner/stores')
	}

	// 暴露方法给父组件调用
	defineExpose({
		fetchShopList
	})

	const handleAuth = async (store : Store) => {
		// 检测登录态
		const token = localStorage.getItem(STORAGE_KEYS.TOKEN)
		const userId = localStorage.getItem(STORAGE_KEYS.USER_ID)
		
		if (!token || !userId) {
			ElMessage.warning('请先登录后再进行授权操作')
			setTimeout(() => {
				router.push(ROUTE_PATH.LOGIN)
			}, 1500)
			return
		}

		// 如果没有店铺，也可以进行授权（首次授权）
		if (!store || !store.shopId) {
			// 首次授权，直接获取授权链接
			loading.value = true
			try {
				const res = await shopeeApi.getAuthURL()

				if (res.code === HTTP_STATUS.OK && res.auth_url) {
					// 在新窗口打开授权链接
					window.open(res.auth_url, '_blank')
					ElMessage.success('正在跳转到 Shopee 授权页面...')
				} else {
					ElMessage.error(res.message || '获取授权链接失败')
				}
			} catch (err : any) {
				ElMessage.error(err?.response?.data?.message || err?.message || '获取授权链接失败')
			} finally {
				loading.value = false
			}
			return
		}

		store.authLoading = true
		try {
			const res = await shopeeApi.getAuthURL()

			if (res.code === HTTP_STATUS.OK && res.auth_url) {
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
		cursor: pointer;
		user-select: none;
		color: #303133;
		
		&:hover {
			opacity: 0.8;
		}
		
		&:active {
			opacity: 0.6;
		}
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
		gap: 8px;
	}

	.store-item {
		display: flex;
		align-items: flex-start;
		gap: 12px;
		padding: 4px 8px;
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
		line-height: 1.2;
		max-width: 210px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.store-status {
		margin: 0;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.store-status-row {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.store-id-inline {
		font-size: 12px;
		color: #909399;
	}

	.store-expire-row {
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