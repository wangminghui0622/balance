<template>
	<div class="orders-page">
		<div class="page-header">
			<h1 class="page-title">我的订单</h1>
			<div class="header-actions">
				<el-icon class="header-icon">
					<Search />
				</el-icon>
				<el-badge :value="4" class="header-icon chat-badge">
					<el-icon>
						<ChatDotRound />
					</el-icon>
				</el-badge>
			</div>
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
							<el-input v-model="filterForm.shopKeyword" placeholder="请输入店铺名称或编号" clearable
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
						<el-form-item label="选择订单状态">
							<el-select v-model="filterForm.orderStatus" placeholder="请选择" clearable style="width: 100%">
								<el-option label="待发货" value="pending_shipment" />
								<el-option label="已发货" value="shipped" />
								<el-option label="已完成" value="completed" />
								<el-option label="已取消" value="cancelled" />
							</el-select>
						</el-form-item>
					</el-col>
				</el-row>
				<el-row :gutter="20" style="margin-top: 20px;">
					<el-col :xs="24" :sm="8">
						<el-form-item label="选择付款状态" label-class="spaced-label">
							<el-select v-model="filterForm.paymentStatus" placeholder="请选择" clearable
								style="width: 100%">
								<el-option label="未付款" value="unpaid" />
								<el-option label="已付款" value="paid" />
								<el-option label="已退款" value="refunded" />
							</el-select>
						</el-form-item>
					</el-col>
					<el-col :xs="24" :sm="8">
						<el-form-item label="日期范围">
							<el-date-picker v-model="filterForm.dateRange" type="daterange" range-separator="至"
								start-placeholder="开始日期" end-placeholder="结束日期" format="YYYY-MM-DD" value-format="YYYY-MM-DD"
								style="width: 100%" />
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
		<el-card class="orders-card">
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
					<el-button :icon="Document" @click="handleExport">导出报表</el-button>
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
						<div class="info-row">
							<span>下单时间: {{ order.orderTime }}</span>
							<span>店铺编号: {{ order.storeId }}</span>
						</div>
						<div class="info-row">
							<span>店铺名称: {{ order.storeName }}</span>
							<span v-if="order.shopeeOrderNo">虾皮订单号: {{ order.shopeeOrderNo }}</span>
						</div>
						<div class="info-row" v-if="order.shopeeStatus">
							<span>虾皮订单状态: {{ order.shopeeStatus }}</span>
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
									颜色: {{ product.color || 'xxx' }} 尺寸: {{ product.size || 'xxx' }}
								</div>
								<div class="product-price-info">
									<span>单价: NT${{ product.unitPrice }}</span>
									<span>数量: {{ product.quantity }}</span>
									<span>小计: {{ product.subtotal }}</span>
								</div>
								<div class="product-shopee-amount" v-if="product.shopeeAmount">
									虾皮订单金额: NT${{ product.shopeeAmount }}
								</div>
							</div>
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
	import { Search, View, Document, Picture, ChatDotRound } from '@element-plus/icons-vue'
	import { ElMessage } from 'element-plus'

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
		products ?: Product[]
	}

	const activeTab = ref('all')
	const showProductInfo = ref(true)
	const shopCount = ref(6)

	const allOrdersCount = ref(245)
	const allOrdersAmount = ref(38420)
	const unsettledOrdersCount = ref(12)
	const unsettledOrdersAmount = ref(456)
	const settledOrdersCount = ref(24)
	const settledOrdersAmount = ref(832)
	const adjustmentOrdersCount = ref(2)
	const adjustmentOrdersAmount = ref(60)

	const filterForm = reactive({
		shopKeyword: '',
		orderNo: '',
		orderStatus: '',
		paymentStatus: '',
		dateRange: null as string[] | null,
	})

	const pagination = reactive({
		page: 1,
		pageSize: 10,
		total: 0,
	})

	const orders = ref<Order[]>([
		{
			orderNo: 'X250904KQ2P078R',
			orderTime: '2025-12-10 23:59:59',
			storeId: 'S1234567890',
			storeName: '示例文字占位符示例文字占位符示例文字占位符',
			shopeeOrderNo: '250904KQ2P078R',
			shopeeStatus: '待发货',
			orderAmount: '36.00',
			paymentStatus: 'paid',
			unsettledCommission: '8.00',
			products: [
				{
					image: '',
					name: '商品名称示例文字占位符替换即可文字占位符替换即可',
					color: 'xxx',
					size: 'xxx',
					unitPrice: '46.00',
					quantity: 1,
					subtotal: '46.00',
					shopeeAmount: '46.00'
				}
			]
		},
		{
			orderNo: 'X250904KQ2P078S',
			orderTime: '2025-12-10 22:30:00',
			storeId: 'S1234567891',
			storeName: '示例文字占位符示例文字占位符',
			shopeeOrderNo: '250904KQ2P078S',
			shopeeStatus: '待发货',
			orderAmount: '88.00',
			paymentStatus: 'paid',
			unsettledCommission: '12.00',
			products: [
				{
					image: '',
					name: '商品名称示例文字占位符替换即可',
					color: '蓝色',
					size: 'L',
					unitPrice: '88.00',
					quantity: 1,
					subtotal: '88.00',
					shopeeAmount: '88.00'
				},
				{
					image: '',
					name: '商品名称示例文字占位符替换即可',
					color: '红色',
					size: 'M',
					unitPrice: '50.00',
					quantity: 1,
					subtotal: '50.00',
					shopeeAmount: '50.00'
				}
			]
		}
	])

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
			result = result.filter(order => order.unsettledCommission && parseFloat(order.unsettledCommission) > 0)
		} else if (activeTab.value === 'settled') {
			result = result.filter(order => !order.unsettledCommission || parseFloat(order.unsettledCommission) === 0)
		} else if (activeTab.value === 'adjustments') {
			// 账款调整的逻辑需要根据实际业务定义
			result = []
		}

		// 根据筛选条件过滤
		if (filterForm.shopKeyword) {
			const keyword = filterForm.shopKeyword.toLowerCase()
			result = result.filter(order =>
				order.storeName.toLowerCase().includes(keyword) ||
				order.storeId.toLowerCase().includes(keyword)
			)
		}

		if (filterForm.orderNo) {
			result = result.filter(order =>
				order.orderNo.toLowerCase().includes(filterForm.orderNo.toLowerCase())
			)
		}

		if (filterForm.orderStatus) {
			result = result.filter(order => order.shopeeStatus === filterForm.orderStatus)
		}

		if (filterForm.paymentStatus) {
			result = result.filter(order => order.paymentStatus === filterForm.paymentStatus)
		}

		if (filterForm.dateRange && filterForm.dateRange.length === 2) {
			const [startDate, endDate] = filterForm.dateRange
			result = result.filter(order => {
				const orderDate = order.orderTime.split(' ')[0]
				return orderDate >= startDate && orderDate <= endDate
			})
		}

		// 更新分页总数
		pagination.total = result.length

		// 分页
		const start = (pagination.page - 1) * pagination.pageSize
		const end = start + pagination.pageSize
		return result.slice(start, end)
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
		ElMessage.success('查询成功')
	}

	const handleReset = () => {
		filterForm.shopKeyword = ''
		filterForm.orderNo = ''
		filterForm.orderStatus = ''
		filterForm.paymentStatus = ''
		filterForm.dateRange = null
		pagination.page = 1
		ElMessage.info('已重置筛选条件')
	}

	const handleTabChange = () => {
		pagination.page = 1
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
	}

	const handlePageChange = (page : number) => {
		pagination.page = page
	}

	// 生命周期
	onMounted(() => {
		// 初始化数据，可以在这里调用API获取订单列表
		pagination.total = orders.value.length
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
					.card-title {
						font-size: 14px;
						color: #909399;
						margin-bottom: 12px;
					}

					.card-value {
						display: flex;
						align-items: center;
						gap: 20px;
						flex-wrap: nowrap;

						.count {
							font-size: 30px;
							font-weight: 600;
							color: #303133;
							position: relative;
							white-space: nowrap;

							.unit {
								font-size: 12px;
								font-weight: 400;
								color: #303133;
								position: absolute;
								bottom: 0.6em;
								right: -18px;
								line-height: 1;
							}
						}

						.equals {
							font-size: 18px;
							font-weight: 400;
							color: #606266;
							margin: 0 4px;
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

		:deep(.spaced-label) {
			letter-spacing: 1.5px;
		}
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

					span {
						white-space: nowrap;
					}
				}
			}

			.order-info {
				margin-bottom: 16px;

				.info-row {
					display: flex;
					gap: 24px;
					margin-bottom: 8px;
					font-size: 14px;
					color: #606266;

					span {
						white-space: nowrap;
					}
				}
			}

			.products-section {
				margin-top: 16px;
				padding: 16px;
				background-color: #fafafa;
				border-radius: 6px;

				.product-item {
					display: flex;
					gap: 16px;
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
						flex: 1;
						display: flex;
						flex-direction: column;
						gap: 8px;

						.product-name {
							font-size: 14px;
							color: #303133;
							font-weight: 500;
							line-height: 1.5;
							word-break: break-word;
						}

						.product-specs {
							font-size: 12px;
							color: #909399;
						}

						.product-price-info {
							display: flex;
							gap: 16px;
							font-size: 13px;
							color: #606266;
							flex-wrap: wrap;

							span {
								white-space: nowrap;
							}
						}

						.product-shopee-amount {
							font-size: 13px;
							color: #ff6a3a;
							font-weight: 500;
							margin-top: 4px;
						}
					}
				}
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
	}
</style>
