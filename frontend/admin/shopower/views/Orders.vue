<template>
	<div class="orders-page">
		<div class="page-header">
			<h1 class="page-title">我的订单</h1>
		</div>
		<el-card class="summary-section-card">
			<div class="summary-section">
				<div class="section-title-wrapper">
					<span class="title-bar"></span>
					<span class="section-title">我的店铺({{ shopCount }})</span>
				</div>
				<el-row :gutter="20" class="summary-cards">
					<el-col :xs="24" :sm="12" :md="6" :lg="6" :xl="6">
						<div class="summary-card">
							<div class="card-content">
								<div class="card-title">全部订单(NT$)</div>
								<div class="card-value">
									<span class="count">
										{{ summaryData.allOrders.count }}<span class="unit">单</span>
									</span>
									<span class="equals">=</span>
									<span class="amount">{{ formatAmount(summaryData.allOrders.amount) }}</span>
								</div>
							</div>
						</div>
					</el-col>
					<el-col :xs="24" :sm="12" :md="6" :lg="6" :xl="6">
						<div class="summary-card">
							<div class="card-content">
								<div class="card-title">未结算订单(NT$)</div>
								<div class="card-value">
									<span class="count">
										{{ summaryData.unsettledOrders.count }}<span class="unit">单</span>
									</span>
									<span class="equals">=</span>
									<span class="amount">{{ formatAmount(summaryData.unsettledOrders.amount) }}</span>
								</div>
							</div>
						</div>
					</el-col>
					<el-col :xs="24" :sm="12" :md="6" :lg="6" :xl="6">
						<div class="summary-card">
							<div class="card-content">
								<div class="card-title">已结算订单(NT$)</div>
								<div class="card-value">
									<span class="count">
										{{ summaryData.settledOrders.count }}<span class="unit">单</span>
									</span>
									<span class="equals">=</span>
									<span class="amount">{{ formatAmount(summaryData.settledOrders.amount) }}</span>
								</div>
							</div>
						</div>
					</el-col>
					<el-col :xs="24" :sm="12" :md="6" :lg="6" :xl="6">
						<div class="summary-card">
							<div class="card-content">
								<div class="card-title">账款调整(NT$)</div>
								<div class="card-value">
									<span class="count">
										{{ summaryData.adjustments.count }}<span class="unit">单</span>
									</span>
									<span class="equals">=</span>
									<span class="amount">{{ formatAmount(summaryData.adjustments.amount) }}</span>
								</div>
							</div>
						</div>
					</el-col>
				</el-row>
			</div>
		</el-card>

		<!-- 搜索筛选栏 -->
		<el-card class="filter-card">
			<el-form :model="filterForm" class="filter-form">
				<el-row :gutter="20">
					<el-col :xs="24" :sm="8">
						<el-form-item label="店铺名称/店铺编号">
							<el-select
								v-model="filterForm.shopKeyword"
								placeholder="全部"
								clearable
								filterable
								style="width: 100%; border-radius: 30px;"
								value-key="id"
							>
								<el-option label="全部" value="" />
								<el-option
                                  v-for="shop in shopOptions"
                                  :key="shop.id"
                                  :label="shop.name"
                                  :value="shop.name"
                                >
                                  <div style="display: flex; justify-content: center; align-items: center; width: 100%;">
                                    <span style="display: inline-block; width: 156px; text-align: left; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; text-align-last: justify; text-justify: distribute;">{{ shop.name }}</span>
                                    <span style="display: inline-block; width: 30px; text-align: center;">&nbsp;|&nbsp;</span>
                                    <span style="display: inline-block; width: 91px; text-align: left; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; letter-spacing: 1px; font-feature-settings: 'tnum';">{{ shop.id }}</span>
                                  </div>
								</el-option>
							</el-select>
						</el-form-item>
					</el-col>
					<el-col :xs="24" :sm="8">
						<el-form-item label="订单编号">
							<el-input v-model="filterForm.orderNo" placeholder="请输入订单编号" clearable
								style="width: 100%">
								<template #prefix>
									<el-icon>
										<Search />
									</el-icon>
								</template>
							</el-input>
						</el-form-item>
					</el-col>
					<el-col :xs="24" :sm="8">
						<el-form-item label="日期范围">
							<el-date-picker v-model="filterForm.dateRange" type="daterange" range-separator="至"
								start-placeholder="开始日期" end-placeholder="结束日期" format="YYYY-MM-DD" value-format="YYYY-MM-DD"
								style="width: 100%" />
						</el-form-item>
					</el-col>
				</el-row>
				<el-row :gutter="20" style="margin-top: 20px;">
					<el-col :xs="24" :sm="8">
						<el-form-item label="付款状态" label-class="spaced-label">
							<el-select v-model="filterForm.paymentStatus" placeholder="请选择"
								style="width: 35%">
								<el-option label="全部" value="all" />
								<el-option label="未付款" value="unpaid" />
								<el-option label="已付款" value="paid" />
								<el-option label="已退款" value="refunded" />
							</el-select>
						</el-form-item>
					</el-col>
					<el-col :xs="24" :sm="8">
						<el-form-item label="订单状态">
							<el-select v-model="filterForm.orderStatus" placeholder="请选择" style="width: 35%">
								<el-option label="全部" value="all" />
								<el-option label="待发货" value="待发货" />
								<el-option label="已发货" value="已发货" />
								<el-option label="已完成" value="已完成" />
								<el-option label="已取消" value="已取消" />
							</el-select>
						</el-form-item>
					</el-col>
					<el-col :xs="24" :sm="8">
						<div style="display: flex; justify-content: flex-end; gap: 10px;">
							<el-button type="primary" @click="handleSearch" style="margin-bottom: 0;">
								<el-icon>
									<Search />
								</el-icon>
								查询
							</el-button>
							<el-button @click="handleReset" style="margin-bottom: 0;">重置</el-button>
						</div>
					</el-col>
				</el-row>
			</el-form>
		</el-card>

		<!-- 订单列表区域 -->
		<el-card class="orders-card" v-loading="loading">
			<div class="orders-header">
				<el-tabs v-model="activeTab" @tab-change="handleTabChange" class="order-tabs">
					<el-tab-pane label="全部订单" name="all" />
					<el-tab-pane label="未结算" name="unsettled" />
					<el-tab-pane label="已结算" name="settled" />
					<el-tab-pane label="账款调整" name="adjustments" />
				</el-tabs>
				<div class="action-buttons">
					<el-button :icon="View" @click="toggleProductInfo">
						{{ showProductInfo ? '隐藏商品信息' : '显示商品信息' }}
					</el-button>
					<el-button @click="handleExport">
						<svg class="export-icon" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
							<rect x="1" y="1" width="14" height="14" rx="2" stroke="currentColor" stroke-width="1.5" fill="none"/>
						</svg>
						导出报表
					</el-button>
				</div>
			</div>

			<!-- 订单列表 -->
			<div class="orders-list">
				<div v-for="(order, index) in filteredOrders" :key="index" class="order-item">
					<!-- 订单头部 -->
					<div class="order-header">
						<div class="order-number">
							订单编号: {{ order.orderNo }}
							<el-tag v-if="order.paymentStatus" :type="getPaymentStatusType(order.paymentStatus)"
								size="small" style="margin-left: 8px">
								{{ getPaymentStatusText(order.paymentStatus) }}
							</el-tag>
						</div>
						<div class="order-amount-info">
							<span v-if="order.unsettledCommission">
								未结算佣金: NT${{ order.unsettledCommission }}
							</span>
							<span>订单金额: NT${{ order.orderAmount }}</span>
						</div>
					</div>

					<!-- 订单信息 -->
					<div class="order-info">
						<div class="order-info-line">
							<span class="info-item">下单时间: {{ order.orderTime }}</span>
							<span class="info-item">店铺编号: {{ order.storeId }}</span>
							<span class="info-item">店铺名称: {{ order.storeName }}</span>
							<span v-if="order.shopeeOrderNo" class="info-item">虾皮订单号: {{ order.shopeeOrderNo }}</span>
							<span v-if="order.shopeeStatus" class="info-item info-item-right">虾皮订单状态: {{ order.shopeeStatus }}</span>
						</div>
					</div>

					<!-- 商品信息 -->
					<div v-if="showProductInfo && order.products && order.products.length > 0" class="products-section">
						<div v-for="(product, pIndex) in order.products" :key="pIndex" class="product-item">
							<el-image :src="product.image || '/placeholder.png'" :alt="product.name"
								class="product-image" fit="cover">
								<template #error>
									<div class="image-slot">
										<el-icon>
											<Picture />
										</el-icon>
									</div>
								</template>
							</el-image>
							<div class="product-details">
								<div class="product-name">{{ product.name }}</div>
								<div class="product-specs">
									颜色: {{ product.color || 'xxx' }}&nbsp;&nbsp;&nbsp;&nbsp;尺寸: {{ product.size || 'xxx' }}
								</div>
								<div class="product-price-info">
									<span>单价: NT${{ product.unitPrice }}</span>
									<span>数量: {{ product.quantity }}</span>
									<span>小计: {{ product.subtotal }}</span>
								</div>
							</div>
						</div>
						<!-- 虾皮订单金额显示在订单级别，多个子商品共用 -->
						<div class="order-shopee-amount" v-if="order.shopeeAmount">
							虾皮订单金额: NT${{ order.shopeeAmount }}
						</div>
						<!-- 已结算佣金和订单金额 -->
						<div class="order-settlement-row">
							<span>已结算佣金: NT${{ order.settledCommission || '0.00' }}</span>
							<span>订单金额: NT${{ order.orderAmount }}</span>
						</div>
					</div>
				</div>

				<!-- 空状态 -->
				<el-empty v-if="filteredOrders.length === 0" description="暂无订单数据" />
			</div>

			<!-- 分页 -->
			<div class="pagination-wrapper" v-if="filteredOrders.length > 0">
				<el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize"
					:total="pagination.total" :page-sizes="[10, 20, 50, 100]"
					layout="total, sizes, prev, pager, next, jumper" @size-change="handleSizeChange"
					@current-change="handlePageChange" />
			</div>
		</el-card>
	</div>
</template>

<script setup lang="ts">
	import { ref, reactive, computed, onMounted } from 'vue'
	import { Search, View, Picture } from '@element-plus/icons-vue'
	import { ElMessage } from 'element-plus'
	import * as orderApi from '@share/api/order'
	import type { Order as ApiOrder } from '@share/api/order'
	import { shopeeApi } from '@share/api/shopee'
	import { HTTP_STATUS } from '@share/constants'

	interface Product {
		image : string
		name : string
		color ?: string
		size ?: string
		unitPrice : string
		quantity : number
		subtotal : string
		shopeeAmount ?: string
	}

	interface Order {
		orderNo : string
		orderTime : string
		storeId : string
		storeName : string
		shopeeOrderNo ?: string
		shopeeStatus ?: string
		orderAmount : string
		paymentStatus ?: string
		unsettledCommission ?: string
		settledCommission ?: string
		shopeeAmount ?: string
		products ?: Product[]
	}

	// 响应式数据
	const shopCount = ref(0)
	const activeTab = ref('all')
	const showProductInfo = ref(true) // 默认显示商品信息
	const loading = ref(false)
	
	// 店铺ID到名称的映射
	const shopNameMap = ref<Record<string, string>>({})

	// 店铺选项数据
	const shopOptions = ref<{ name: string; id: string }[]>([])

	// 订单统计相关数据 - 8个独立变量
	const allOrdersCount = ref(0)      // 全部订单数量
	const allOrdersAmount = ref(0) // 全部订单金额
	const unsettledOrdersCount = ref(0)  // 未结算订单数量
	const unsettledOrdersAmount = ref(0) // 未结算订单金额
	const settledOrdersCount = ref(0)    // 已结算订单数量
	const settledOrdersAmount = ref(0) // 已结算订单金额
	const adjustmentOrdersCount = ref(0) // 账款调整订单数量
	const adjustmentOrdersAmount = ref(0) // 账款调整订单金额

	// 获取默认日期范围（今天和10天前）
	const getDefaultDateRange = (): string[] => {
		const today = new Date()
		const tenDaysAgo = new Date()
		tenDaysAgo.setDate(today.getDate() - 10)
		
		const formatDate = (date: Date): string => {
			const year = date.getFullYear()
			const month = String(date.getMonth() + 1).padStart(2, '0')
			const day = String(date.getDate()).padStart(2, '0')
			return `${year}-${month}-${day}`
		}
		
		return [formatDate(tenDaysAgo), formatDate(today)]
	}

	const filterForm = reactive({
		shopKeyword: '',
		orderNo: '',
		orderStatus: 'all',
		paymentStatus: 'all',
		dateRange: getDefaultDateRange() as string[] | null,
	})

	const pagination = reactive({
		page: 1,
		pageSize: 10,
		total: 0,
	})

	// 原始API订单数据
	const apiOrders = ref<ApiOrder[]>([])
	
	// 格式化时间：将 ISO 8601 格式转换为 YYYY-MM-DD HH:mm:ss
	const formatDateTime = (isoString: string | null): string => {
		if (!isoString) return '-'
		const date = new Date(isoString)
		if (isNaN(date.getTime())) return isoString
		const year = date.getFullYear()
		const month = String(date.getMonth() + 1).padStart(2, '0')
		const day = String(date.getDate()).padStart(2, '0')
		const hours = String(date.getHours()).padStart(2, '0')
		const minutes = String(date.getMinutes()).padStart(2, '0')
		const seconds = String(date.getSeconds()).padStart(2, '0')
		return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
	}

	// 订单状态映射
	const statusMap: Record<string, string> = {
		'UNPAID': '未付款',
		'READY_TO_SHIP': '待发货',
		'PROCESSED': '已处理',
		'SHIPPED': '已发货',
		'COMPLETED': '已完成',
		'IN_CANCEL': '取消中',
		'CANCELLED': '已取消',
		'INVOICE_PENDING': '待开票'
	}

	// 将API订单数据转换为显示格式
	const transformOrder = (apiOrder: ApiOrder): Order => {
		const products: Product[] = (apiOrder.items || []).map(item => ({
			image: '',
			name: item.item_name,
			color: item.model_name?.split(',')[0] || '-',
			size: item.model_name?.split(',')[1] || '-',
			unitPrice: item.item_price,
			quantity: item.quantity,
			subtotal: (parseFloat(item.item_price) * item.quantity).toFixed(2)
		}))

		// 根据订单状态判断付款状态
		let paymentStatus = 'paid'
		if (apiOrder.order_status === 'UNPAID') {
			paymentStatus = 'unpaid'
		} else if (apiOrder.order_status === 'CANCELLED' || apiOrder.order_status === 'IN_CANCEL') {
			paymentStatus = 'refunded'
		}

		return {
			orderNo: apiOrder.order_sn,
			orderTime: formatDateTime(apiOrder.create_time || apiOrder.created_at),
			storeId: apiOrder.shop_id.toString(),
			storeName: shopNameMap.value[apiOrder.shop_id.toString()] || `店铺 ${apiOrder.shop_id}`,
			shopeeOrderNo: apiOrder.order_sn,
			shopeeStatus: statusMap[apiOrder.order_status] || apiOrder.order_status,
			orderAmount: apiOrder.total_amount,
			paymentStatus,
			unsettledCommission: ['READY_TO_SHIP', 'PROCESSED', 'SHIPPED'].includes(apiOrder.order_status) ? '0.00' : '0.00',
			settledCommission: apiOrder.order_status === 'COMPLETED' ? '0.00' : undefined,
			shopeeAmount: apiOrder.total_amount,
			products: products.length > 0 ? products : undefined
		}
	}

	// 将API订单转换为显示格式
	const orders = computed(() => {
		return apiOrders.value.map(transformOrder)
	})

	// 计算统计数据
	const summaryData = computed(() => {
		return {
			allOrders: {
				count: allOrdersCount.value,
				amount: allOrdersAmount.value
			},
			unsettledOrders: {
				count: unsettledOrdersCount.value,
				amount: unsettledOrdersAmount.value
			},
			settledOrders: {
				count: settledOrdersCount.value,
				amount: settledOrdersAmount.value
			},
			adjustments: {
				count: adjustmentOrdersCount.value,
				amount: adjustmentOrdersAmount.value
			}
		}
	})

	// 计算属性
	const filteredOrders = computed(() => {
		let result = [...orders.value]

		// 根据标签页过滤
		if (activeTab.value === 'unsettled') {
			result = result.filter(order => {
				const apiOrder = apiOrders.value.find(o => o.order_sn === order.orderNo)
				return apiOrder && ['READY_TO_SHIP', 'PROCESSED', 'SHIPPED'].includes(apiOrder.order_status)
			})
		} else if (activeTab.value === 'settled') {
			result = result.filter(order => {
				const apiOrder = apiOrders.value.find(o => o.order_sn === order.orderNo)
				return apiOrder && apiOrder.order_status === 'COMPLETED'
			})
		} else if (activeTab.value === 'adjustments') {
			// 账款调整的逻辑需要根据实际业务定义
			result = []
		}

		// 根据筛选条件过滤（前端二次过滤，主要筛选已在API层处理）
		if (filterForm.paymentStatus && filterForm.paymentStatus !== 'all') {
			result = result.filter(order => order.paymentStatus === filterForm.paymentStatus)
		}

		return result
	})

	// 方法
	const formatAmount = (amount : number) : string => {
		return amount.toLocaleString('zh-CN', {
			minimumFractionDigits: 2,
			maximumFractionDigits: 2
		})
	}

	const getPaymentStatusType = (status : string) : string => {
		const statusMap : Record<string, string> = {
			paid: 'success',
			unpaid: 'warning',
			refunded: 'info'
		}
		return statusMap[status] || 'info'
	}

	const getPaymentStatusText = (status : string) : string => {
		const statusMap : Record<string, string> = {
			paid: '已付款',
			unpaid: '未付款',
			refunded: '已退款'
		}
		return statusMap[status] || '未知'
	}

	const handleSearch = () => {
		pagination.page = 1
		fetchOrders()
	}

	const handleReset = () => {
		filterForm.shopKeyword = ''
		filterForm.orderNo = ''
		filterForm.orderStatus = 'all'
		filterForm.paymentStatus = 'all'
		filterForm.dateRange = getDefaultDateRange()
		pagination.page = 1
		fetchOrders()
		ElMessage.info('已重置筛选条件')
	}

	const handleTabChange = () => {
		pagination.page = 1
		fetchOrders()
	}

	const toggleProductInfo = () => {
		showProductInfo.value = !showProductInfo.value
	}

	const handleExport = () => {
		ElMessage.info('导出功能开发中...')
	}

	const handleSizeChange = (size : number) => {
		pagination.pageSize = size
		pagination.page = 1
		fetchOrders()
	}

	const handlePageChange = (page : number) => {
		pagination.page = page
		fetchOrders()
	}

	// 获取店铺列表
	const fetchShops = async () => {
		try {
			const res = await shopeeApi.getShopList()
			if (res.code === HTTP_STATUS.OK && res.data) {
				const shops = res.data.list || []
				shopCount.value = shops.length
				shopOptions.value = shops.map((shop: any) => ({
					name: shop.shopName || `店铺 ${shop.shopId}`,
					id: shop.shopId.toString()
				}))
				// 构建店铺ID到名称的映射
				shops.forEach((shop: any) => {
					shopNameMap.value[shop.shopId.toString()] = shop.shopName || `店铺 ${shop.shopId}`
				})
			}
		} catch (err) {
			console.error('获取店铺列表失败:', err)
		}
	}

	// 获取订单列表
	const fetchOrders = async () => {
		loading.value = true
		try {
			// 构建查询参数
			const params: orderApi.OrderListParams = {
				page: pagination.page,
				page_size: pagination.pageSize
			}

			// 添加筛选条件
			if (filterForm.orderNo) {
				params.order_sn = filterForm.orderNo
			}
			
			// 根据选中的店铺名称找到店铺ID
			if (filterForm.shopKeyword) {
				const selectedShop = shopOptions.value.find(s => s.name === filterForm.shopKeyword)
				if (selectedShop) {
					params.shop_id = parseInt(selectedShop.id)
				}
			}

			// 根据订单状态筛选
			if (filterForm.orderStatus && filterForm.orderStatus !== 'all') {
				// 需要将中文状态转换为API状态
				const statusReverseMap: Record<string, string> = {
					'待发货': 'READY_TO_SHIP',
					'已发货': 'SHIPPED',
					'已完成': 'COMPLETED',
					'已取消': 'CANCELLED'
				}
				params.status = statusReverseMap[filterForm.orderStatus] || filterForm.orderStatus
			}

			// 日期范围
			if (filterForm.dateRange && filterForm.dateRange.length === 2) {
				params.start_time = filterForm.dateRange[0]
				params.end_time = filterForm.dateRange[1]
			}

			const res = await orderApi.getOrderList(params)
			if (res.code === HTTP_STATUS.OK && res.data) {
				apiOrders.value = res.data.list || []
				pagination.total = res.data.total || 0

				// 更新统计数据
				updateStatistics()
			}
		} catch (err: any) {
			console.error('获取订单列表失败:', err)
			ElMessage.error(err?.message || '获取订单列表失败')
		} finally {
			loading.value = false
		}
	}

	// 更新统计数据
	const updateStatistics = () => {
		const all = apiOrders.value
		allOrdersCount.value = all.length
		allOrdersAmount.value = all.reduce((sum, o) => sum + parseFloat(o.total_amount || '0'), 0)

		const unsettled = all.filter(o => ['READY_TO_SHIP', 'PROCESSED', 'SHIPPED'].includes(o.order_status))
		unsettledOrdersCount.value = unsettled.length
		unsettledOrdersAmount.value = unsettled.reduce((sum, o) => sum + parseFloat(o.total_amount || '0'), 0)

		const settled = all.filter(o => o.order_status === 'COMPLETED')
		settledOrdersCount.value = settled.length
		settledOrdersAmount.value = settled.reduce((sum, o) => sum + parseFloat(o.total_amount || '0'), 0)
	}

	// 生命周期
	onMounted(async () => {
		await fetchShops()
		await fetchOrders()
	})
</script>

<style scoped lang="scss">
	.orders-page {
		padding: 20px;
		background-color: #f5f7fa;
		min-height: calc(100vh - 60px);
	}

	.page-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 16px;
		padding-bottom: 0;

		.page-title {
			font-size: 20px;
			font-weight: 500;
			color: #303133;
			margin: 0;
		}

		.header-actions {
			display: flex;
			align-items: center;
			gap: 16px;

			.header-icon {
				font-size: 20px;
				color: #606266;
				cursor: pointer;
				transition: color 0.3s;

				&:hover {
					color: #ff6a3a;
				}
			}

			.chat-badge {
				:deep(.el-badge__content) {
					background-color: #ff6a3a;
					border-color: #ff6a3a;
				}
			}
		}
	}

	.summary-section-card {
		margin-bottom: 20px;
		border-radius: 8px;

		:deep(.el-card__body) {
			padding: 20px;
		}
	}

	.summary-section {
		.section-title-wrapper {
			display: flex;
			align-items: center;
			gap: 8px;
			margin-bottom: 16px;

			.title-bar {
				width: 3px;
				height: 16px;
				background-color: #ff6a3a;
				border-radius: 2px;
				flex-shrink: 0;
			}

			.section-title {
				font-size: 16px;
				font-weight: 600;
				color: #303133;
			}
		}

		.summary-cards {
			.summary-card {
				padding: 16px;
				background-color: #f5f7fa;
				border-radius: 8px;
				border: 1px solid #ebeef5;
				transition: box-shadow 0.3s;

				&:hover {
					box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
				}

				.card-content {
					text-align: center;

					.card-title {
						font-size: 14px;
						color: #909399;
						margin-bottom: 12px;
						text-align: center;
					}

					.card-value {
						display: flex;
						align-items: center;
						justify-content: center;
						gap: 16px;
						flex-wrap: nowrap;

						.count {
							font-size: 30px;
							font-weight: 600;
							color: #303133;
							white-space: nowrap;

							.unit {
								font-size: 12px;
								font-weight: 400;
								color: #303133;
								margin-left: 2px;
							}
						}

						.equals {
							font-size: 18px;
							font-weight: 400;
							color: #606266;
						}

						.amount {
							font-size: 30px;
							font-weight: 600;
							color: #303133;
							white-space: nowrap;
						}
					}
				}
			}
		}
	}

	.filter-card {
		margin-bottom: 20px;
		border-radius: 8px;

		.filter-form {
			:deep(.el-form-item) {
				width: 100%;
				margin-right: 0;
			}
		}

		:deep(.el-form-item__label.spaced-label) {
			letter-spacing: 8px !important;
		}
		
		/* 强制选中项显示为与日期相同的颜色 */
		:deep(.el-popper__wrapper .el-select__popper .el-select-dropdown__item.selected),
		:deep(.el-popper__wrapper .el-select__popper .el-select-dropdown__item[aria-selected="true"]),
		:deep(.el-popper__wrapper .el-select__popper .el-select-dropdown__item),
		:deep(.el-select .el-input__inner),
		:deep(.el-select .el-input__wrapper) {
			color: #303133 !important;
			font-weight: normal !important;
		}
		
		/* 确保默认显示文本为与日期相同的颜色 */
		:deep(.el-select .el-input__inner) {
			color: #303133 !important;
		}
		
		/* 所有选项使用相同颜色 */
		:deep(.el-select span) {
			color: #303133 !important;
			font-weight: normal !important;
		}
		
		/* 强制下拉框和输入框圆角 */
		:deep(.el-input__wrapper),
		:deep(.el-input__inner),
		:deep(.el-select .el-select__wrapper) {
			border-radius: 30px !important;
			overflow: hidden;
		}
	}

	/* 下拉弹出框圆角（全局样式，需要不带scoped） */
	:global(.el-select__popper.el-popper) {
		border-radius: 8px !important;
		overflow: hidden;
	}
	
	:global(.el-select-dropdown) {
		border-radius: 8px !important;
		overflow: hidden;
	}
	
	:global(.el-select-dropdown__list) {
		border-radius: 8px !important;
	}
	
	/* 下拉选项内容居中 */
	:global(.el-select-dropdown__item) {
		text-align: center;
		justify-content: center;
	}

	.orders-card {
		border-radius: 8px;

		:deep(.el-card__body) {
			padding: 0;
		}

		.orders-header {
			display: flex;
			justify-content: space-between;
			align-items: center;
			padding: 20px;
			border-bottom: 1px solid #ebeef5;

			.order-tabs {
				flex: 1;
			}

			.action-buttons {
				display: flex;
				gap: 12px;

				.export-icon {
					width: 16px;
					height: 16px;
					margin-right: 4px;
				}
			}
		}

		.order-item {
			padding: 20px;
			margin-bottom: 16px;
			background-color: #ffffff;
			border: 1px solid #ebeef5;
			border-radius: 8px;
			transition: box-shadow 0.3s;

			&:hover {
				box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
			}

			.order-header {
				display: flex;
				justify-content: space-between;
				align-items: center;
				margin-bottom: 16px;
				padding-bottom: 12px;
				border-bottom: 1px solid #f5f7fa;

				.order-number {
					font-size: 16px;
					font-weight: 600;
					color: #303133;
				}

				.order-amount-info {
					display: flex;
					gap: 16px;
					font-size: 14px;
					color: #606266;
					padding-right: 16px;

					span {
						white-space: nowrap;
					}
				}
			}

			.order-info {
				margin-bottom: 16px;
				background-color: #ebeef5;
				border-radius: 4px;
				padding: 10px 14px;

				.order-info-line {
					display: flex;
					justify-content: flex-start;
					font-size: 12px;
					color: #909399;
					flex-wrap: nowrap;
					gap: 32px;
				}

				.info-item {
					white-space: nowrap;
					min-width: 0;
				}

				.info-item-right {
					margin-left: auto;
				}
			}

			.products-section {
				margin-top: 16px;
				padding: 16px 16px 0 16px;
				background-color: #fafafa;
				border-radius: 6px;

				.product-item {
					display: grid;
					grid-template-columns: 80px minmax(0, 1fr);
					column-gap: 16px;
					padding: 12px 0;
					border-bottom: 1px solid #ebeef5;

					&:last-child {
						border-bottom: none;
						padding-bottom: 0;
					}

					&:first-child {
						padding-top: 0;
					}

					.product-image {
						width: 80px;
						height: 80px;
						border-radius: 6px;
						flex-shrink: 0;
						background-color: #f5f7fa;
						border: 1px solid #ebeef5;

						.image-slot {
							display: flex;
							justify-content: center;
							align-items: center;
							width: 100%;
							height: 100%;
							background: #f5f7fa;
							color: #909399;
							font-size: 24px;
						}
					}

					.product-details {
						min-width: 0;
						display: grid;
						grid-template-columns: minmax(0, 1fr) 240px;
						grid-template-rows: auto auto auto auto;
						column-gap: 16px;
						row-gap: 6px;

						.product-name {
							font-size: 14px;
							color: #303133;
							font-weight: 500;
							line-height: 1.5;
							grid-column: 1;
							overflow: hidden;
							text-overflow: ellipsis;
							white-space: nowrap;
						}

						.product-specs {
							font-size: 12px;
							color: #909399;
							grid-column: 1;
						}

						.product-price-info {
							display: flex;
							gap: 24px;
							font-size: 13px;
							color: #606266;
							grid-column: 2;
							justify-content: flex-end;
							align-items: center;
							white-space: nowrap;
							padding-right: 16px;
						}

						.product-shopee-amount {
							font-size: 13px;
							color: #ff6a3a;
							font-weight: 500;
							grid-column: 2;
							justify-self: end;
						}
					}
				}
			}

			.order-shopee-amount {
				font-size: 12px;
				color: #909399;
				font-weight: 400;
				text-align: right;
				margin-top: 0;
				padding-top: 4px;
				padding-right: 16px;
				border-top: 1px solid #ebeef5;
			}

			.order-settlement-row {
				display: flex;
				justify-content: flex-end;
				gap: 48px;
				font-size: 13px;
				color: #606266;
				margin-top: 12px;
				margin-bottom: -12px;
				padding-right: 16px;
			}
		}
	}

	.pagination-wrapper {
		margin-top: 20px;
		padding: 20px;
		display: flex;
		justify-content: flex-end;
	}

	@media (max-width: 768px) {
		.orders-page {
			padding: 12px;
		}

		.summary-cards {
			:deep(.el-col) {
				margin-bottom: 12px;
			}
		}

		.filter-form {
			:deep(.el-form-item) {
				width: 100%;
				margin-right: 0;
			}
		}

		.orders-header {
			flex-direction: column;
			align-items: flex-start !important;
			gap: 16px;

			.action-buttons {
				width: 100%;

				.el-button {
					flex: 1;
				}
			}
		}

		.order-header {
			flex-direction: column;
			align-items: flex-start !important;
			gap: 12px;
		}

		.order-info {
			.order-info-line {
				flex-wrap: wrap;
				gap: 8px 16px;
			}
		}

		.products-section {
			padding: 12px;

			.product-item {
				grid-template-columns: 60px minmax(0, 1fr);

				.product-image {
					width: 60px;
					height: 60px;
				}

				.product-details {
					grid-template-columns: 1fr;

					.product-price-info {
						grid-column: 1;
						justify-items: start;
					}
				}
			}

			.order-settlement-row {
				flex-wrap: wrap;
				gap: 8px 24px;
			}
		}
	}
</style>
