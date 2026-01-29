<template>
	<div class="orders-page">
		<div class="page-header">
			<h1 class="page-title">444444</h1>
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
					<span class="section-title">33333{{ shopCount }}</span>
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
								<div class="card-title">已结算订单</div>
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

		<!-- 鎼滅储绛涢€夋爮 -->
		<el-card class="filter-card">
			<el-form :model="filterForm" class="filter-form">
				<el-row :gutter="20">
					<el-col :xs="24" :sm="8">
						<el-form-item label="搴楅摵鍚嶇О/搴楅摵缂栧彿">
							<el-input v-model="filterForm.shopKeyword" placeholder="璇疯緭鍏ュ簵閾哄悕绉版垨缂栧彿" clearable
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
						<el-form-item label="璁㈠崟缂栧彿">
							<el-input v-model="filterForm.orderNo" placeholder="璇疯緭鍏ヨ鍗曠紪鍙?" clearable
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
						<el-form-item label="閫夋嫨浠樻鐘舵€?" label-class="spaced-label">
							<el-select v-model="filterForm.paymentStatus" placeholder="璇烽€夋嫨" clearable
								style="width: 100%">
								<el-option label="鏈粯娆" value="unpaid" />
								<el-option label="已付款" value="paid" />
								<el-option label="已退款" value="refunded" />
							</el-select>
						</el-form-item>
					</el-col>
					<el-col :xs="24" :sm="8">
						<el-form-item label="鏃ユ湡鑼冨洿">
							<el-date-picker v-model="filterForm.dateRange" type="daterange" range-separator="鑷?
                start-placeholder=" 寮€濮嬫棩鏈? end-placeholder="缁撴潫鏃ユ湡" format="YYYY-MM-DD" value-format="YYYY-MM-DD"
								style="width: 100%" />
						</el-form-item>
					</el-col>
					<el-col :xs="24" :sm="8">
						<div style="display: flex; justify-content: flex-end; gap: 10px;">
							<el-button type="primary" @click="handleSearch" style="margin-bottom: 0;">
								<el-icon>
									<Search />
								</el-icon>
								鏌ヨ
							</el-button>
							<el-button @click="handleReset" style="margin-bottom: 0;">閲嶇疆</el-button>
						</div>
					</el-col>
				</el-row>
			</el-form>
		</el-card>

		<!-- 璁㈠崟鍒楄〃鍖哄煙 -->
		<el-card class="orders-card">
			<div class="orders-header">
				<el-tabs v-model="activeTab" @tab-change="handleTabChange" class="order-tabs">
					<el-tab-pane label="1111" name="all" />
					<el-tab-pane label="2222" name="unsettled" />
					<el-tab-pane label="33333" name="settled" />
					<el-tab-pane label="1444444" name="adjustments" />
				</el-tabs>
				<div class="action-buttons">
					<el-button :icon="View" @click="toggleProductInfo">
						{{ showProductInfo ? '闅愯棌鍟嗗搧淇℃伅' : '鏄剧ず鍟嗗搧淇℃伅' }}
					</el-button>
					<el-button :icon="Document" @click="handleExport">瀵煎嚭鎶ヨ〃</el-button>
				</div>
			</div>

			<!-- 璁㈠崟鍒楄〃 -->
			<div class="orders-list">
				<div v-for="(order, index) in filteredOrders" :key="index" class="order-item">
					<!-- 璁㈠崟澶撮儴 -->
					<div class="order-header">
						<div class="order-number">
							璁㈠崟缂栧彿: {{ order.orderNo }}
							<el-tag v-if="order.paymentStatus" :type="getPaymentStatusType(order.paymentStatus)"
								size="small" style="margin-left: 8px">
								{{ getPaymentStatusText(order.paymentStatus) }}
							</el-tag>
						</div>
						<div class="order-amount-info">
							<span v-if="order.unsettledCommission">
								鏈粨绠椾剑閲? NT${{ order.unsettledCommission }}
							</span>
							<span>璁㈠崟閲戦: NT${{ order.orderAmount }}</span>
						</div>
					</div>

					<!-- 璁㈠崟淇℃伅 -->
					<div class="order-info">
						<div class="info-row">
							<span>涓嬪崟鏃堕棿: {{ order.orderTime }}</span>
							<span>搴楅摵缂栧彿: {{ order.storeId }}</span>
						</div>
						<div class="info-row">
							<span>搴楅摵鍚嶇О: {{ order.storeName }}</span>
							<span v-if="order.shopeeOrderNo">铏剧毊璁㈠崟鍙? {{ order.shopeeOrderNo }}</span>
						</div>
						<div class="info-row" v-if="order.shopeeStatus">
							<span>铏剧毊璁㈠崟鐘舵€? {{ order.shopeeStatus }}</span>
						</div>
					</div>

					<!-- 鍟嗗搧淇℃伅 -->
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
									棰滆壊: {{ product.color || 'xxx' }} 灏哄: {{ product.size || 'xxx' }}
								</div>
								<div class="product-price-info">
									<span>鍗曚环: NT${{ product.unitPrice }}</span>
									<span>鏁伴噺: {{ product.quantity }}</span>
									<span>灏忚: {{ product.subtotal }}</span>
								</div>
								<div class="product-shopee-amount" v-if="product.shopeeAmount">
									铏剧毊璁㈠崟閲戦: NT${{ product.shopeeAmount }}
								</div>
							</div>
						</div>
					</div>
				</div>

				<!-- 绌虹姸鎬?-->
				<el-empty v-if="filteredOrders.length === 0" description="鏆傛棤璁㈠崟鏁版嵁" />
			</div>

			<!-- 鍒嗛〉 -->
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

	const allOrdersAmount = ref(123) 
	const unsettledOrdersCount = ref(3)  
	const unsettledOrdersAmount = ref(567)
	const settledOrdersCount = ref(5)
	const settledOrdersAmount = ref(666)
	const adjustmentOrdersCount = ref(0)
	const adjustmentOrdersAmount = ref(0)

	const filterForm = reactive({
		shopKeyword: '',
		orderNo: '',
		orderStatus: '',
		paymentStatus: '',
		dateRange: null,
	})

	const pagination = reactive({
		page: 1,
		pageSize: 10,
		total: 0,
	})

	const orders = ref<Order[]>([{
			orderNo: 'X250904KQ2P078R',
			orderTime: '2025-12-10 23:59:59',
			storeId: 'S1234567890',
			storeName: '44444444',
			shopeeOrderNo: '250904KQ2P078R',
			shopeeStatus: '555555',
			orderAmount: '36.00',
			paymentStatus: 'paid',
			unsettledCommission: '8.00',
			products: [
				{
					image: '',
					name: '6666666',
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
			storeName: '绀轰緥鏂囧瓧鍗犱綅绗︾ず渚嬫枃瀛楀崰浣嶇',
			shopeeOrderNo: '250904KQ2P078S',
shopeeStatus: '寰呭彂璐',
    orderAmount: '88.00',
			paymentStatus: 'paid',
			unsettledCommission: '12.00',
			products: [
				{
					image: '',
					name: '鍟嗗搧鍚嶇О绀轰緥鏂囧瓧鍗犱綅绗︽浛鎹㈠嵆鍙'
        color: '钃濊壊',
					size: 'L',
					unitPrice: '88.00',
					quantity: 1,
					subtotal: '88.00',
					shopeeAmount: '88.00'
				},
				{
					image: '',
					name: '鍟嗗搧鍚嶇О绀轰緥鏂囧瓧鍗犱綅绗︽浛鎹㈠嵆鍙?',
					color: '绾㈣壊',
					size: 'M',
					unitPrice: '50.00',
					quantity: 1,
					subtotal: '50.00',
					shopeeAmount: '50.00'
				}
			]
		},
		{
			orderNo: 'X250904KQ2P078T',
			orderTime: '2025-12-10 20:15:30',
			storeId: 'S1234567892',
storeName: '搴楅摵鍚嶇О绀轰緥鏂囧瓧鍗犱綅绗︽浛鎹㈠嵆鍙',
    shopeeOrderNo: '250904KQ2P078T',
shopeeStatus: '宸插彂璐',
    orderAmount: '156.00',
			paymentStatus: 'paid',
			unsettledCommission: '18.00',
			products: [
				{
					image: '',
					name: '鍟嗗搧鍚嶇О绀轰緥鏂囧瓧鍗犱綅绗︽浛鎹㈠嵆鍙晢鍝佸悕绉扮ず渚'
        color: '缁胯壊',
					size: 'XL',
					unitPrice: '156.00',
					quantity: 1,
					subtotal: '156.00',
					shopeeAmount: '156.00'
				}
			]
		},
		{
			orderNo: 'X250904KQ2P078U',
			orderTime: '2025-12-10 18:45:20',
			storeId: 'S1234567893',
storeName: '绀轰緥鏂囧瓧鍗犱綅绗︾ず渚嬫枃瀛',
    shopeeOrderNo: '250904KQ2P078U',
shopeeStatus: '宸插畬鎴',
    orderAmount: '268.00',
			paymentStatus: 'paid',
			unsettledCommission: '0.00',
			products: [
				{
					image: '',
					name: '鍟嗗搧鍚嶇О绀轰緥鏂囧瓧鍗犱綅绗︽浛鎹㈠嵆鍙枃瀛楀崰浣嶇',
					color: '榛戣壊',
					size: 'M',
					unitPrice: '128.00',
					quantity: 1,
					subtotal: '128.00',
					shopeeAmount: '128.00'
				},
				{
					image: '',
					name: '鍟嗗搧鍚嶇О绀轰緥鏂囧瓧鍗犱綅绗︽浛鎹㈠嵆鍙?',
					color: '鐧借壊',
					size: 'S',
					unitPrice: '140.00',
					quantity: 1,
					subtotal: '140.00',
					shopeeAmount: '140.00'
				}
			]
		},
		{
			orderNo: 'X250904KQ2P078V',
			orderTime: '2025-12-10 16:30:15',
			storeId: 'S1234567894',
storeName: '搴楅摵鍚嶇О绀轰緥鏂囧瓧鍗犱綅绗',
    shopeeOrderNo: '250904KQ2P078V',
shopeeStatus: '寰呭彂璐',
    orderAmount: '92.00',
			paymentStatus: 'unpaid',
			unsettledCommission: '15.00',
			products: [
				{
					image: '',
					name: '鍟嗗搧鍚嶇О绀轰緥鏂囧瓧鍗犱綅绗︽浛鎹㈠嵆鍙晢鍝佸悕绉'
        color: '榛勮壊',
					size: 'L',
					unitPrice: '92.00',
					quantity: 1,
					subtotal: '92.00',
					shopeeAmount: '92.00'
				}
			]
		},
		{
			orderNo: 'X250904KQ2P078W',
			orderTime: '2025-12-10 14:20:10',
			storeId: 'S1234567895',
			storeName: '绀轰緥鏂囧瓧鍗犱綅绗︾ず渚嬫枃瀛楀崰浣嶇绀轰緥',
			shopeeOrderNo: '250904KQ2P078W',
shopeeStatus: '寰呭彂璐',
    orderAmount: '74.00',
			paymentStatus: 'paid',
			unsettledCommission: '10.00',
			products: [
				{
					image: '',
					name: '鍟嗗搧鍚嶇О绀轰緥鏂囧瓧鍗犱綅绗︽浛鎹㈠嵆鍙'
        color: '绱壊',
					size: 'M',
					unitPrice: '74.00',
					quantity: 1,
					subtotal: '74.00',
					shopeeAmount: '74.00'
				}
			]
		},
		{
			orderNo: 'X250904KQ2P078X',
			orderTime: '2025-12-10 12:10:05',
			storeId: 'S1234567896',
storeName: '搴楅摵鍚嶇О绀轰緥鏂囧瓧鍗犱綅绗︽浛鎹㈠嵆鍙',
    shopeeOrderNo: '250904KQ2P078X',
shopeeStatus: '宸插彂璐',
    orderAmount: '320.00',
			paymentStatus: 'paid',
			unsettledCommission: '25.00',
			products: [
				{
					image: '',
					name: '鍟嗗搧鍚嶇О绀轰緥鏂囧瓧鍗犱綅绗︽浛鎹㈠嵆鍙枃瀛楀崰浣嶇鏇挎崲鍗冲彲',
					color: '鐏拌壊',
					size: 'XXL',
					unitPrice: '320.00',
					quantity: 1,
					subtotal: '320.00',
					shopeeAmount: '320.00'
				}
			]
		},
		{
			orderNo: 'X250904KQ2P078Y',
			orderTime: '2025-12-10 10:05:00',
			storeId: 'S1234567890',
storeName: '绀轰緥鏂囧瓧鍗犱綅绗︾ず渚嬫枃瀛楀崰浣嶇绀轰緥鏂囧瓧鍗犱綅绗',
    shopeeOrderNo: '250904KQ2P078Y',
shopeeStatus: '宸插畬鎴',
    orderAmount: '198.00',
			paymentStatus: 'paid',
			unsettledCommission: '0.00',
			products: [
				{
					image: '',
					name: '鍟嗗搧鍚嶇О绀轰緥鏂囧瓧鍗犱綅绗︽浛鎹㈠嵆鍙'
        color: '妫曡壊',
					size: 'L',
					unitPrice: '98.00',
					quantity: 1,
					subtotal: '98.00',
					shopeeAmount: '98.00'
				},
				{
					image: '',
					name: '鍟嗗搧鍚嶇О绀轰緥鏂囧瓧鍗犱綅绗︽浛鎹㈠嵆鍙?',
					color: '绮夎壊',
					size: 'S',
					unitPrice: '100.00',
					quantity: 1,
					subtotal: '100.00',
					shopeeAmount: '100.00'
				}
			]
		}
	])

	// 璁＄畻缁熻鏁版嵁
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

	// 璁＄畻灞炴€?const filteredOrders = computed(() => {
	let result = [...orders.value]

	// 鏍规嵁鏍囩椤佃繃婊?  if (activeTab.value === 'unsettled') {
	result = result.filter(order => order.unsettledCommission && parseFloat(order.unsettledCommission) > 0)
  } else if (activeTab.value === 'settled') {
		result = result.filter(order => !order.unsettledCommission || parseFloat(order.unsettledCommission) === 0)
	} else if (activeTab.value === 'adjustments') {
		// 璐︽璋冩暣鐨勯€昏緫闇€瑕佹牴鎹疄闄呬笟鍔″畾涔?    result = []
	}

	// 鏍规嵁绛涢€夋潯浠惰繃婊?  if (filterForm.shopKeyword) {
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

	// 鏇存柊鍒嗛〉鎬绘暟
	pagination.total = result.length

	// 鍒嗛〉
	const start = (pagination.page - 1) * pagination.pageSize
	const end = start + pagination.pageSize
	return result.slice(start, end)
})

	// 鏂规硶
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
paid: '宸蹭粯娆',
unpaid: '鏈粯娆',
			refunded: '已退款'
		}
		return statusMap[status] || '鏈煡'
	}

	const handleSearch = () => {
		pagination.page = 1
		ElMessage.success('鏌ヨ鎴愬姛')
	}

	const handleReset = () => {
		filterForm.shopKeyword = ''
		filterForm.orderNo = ''
		filterForm.orderStatus = ''
		filterForm.paymentStatus = ''
		filterForm.dateRange = null
		pagination.page = 1
		ElMessage.info('宸查噸缃瓫閫夋潯浠?)
}

	const handleTabChange = (tabName : string) => {
		pagination.page = 1
	}

	const toggleProductInfo = () => {
		showProductInfo.value = !showProductInfo.value
	}

	const handleExport = () => {
		ElMessage.info('瀵煎嚭鍔熻兘寮€鍙戜腑...')
	}

	const handleSizeChange = (size : number) => {
		pagination.pageSize = size
		pagination.page = 1
	}

	const handlePageChange = (page : number) => {
		pagination.page = page
	}

	// 鐢熷懡鍛ㄦ湡
	onMounted(() => {
		// 鍒濆鍖栨暟鎹紝鍙互鍦ㄨ繖閲岃皟鐢ˋPI鑾峰彇璁㈠崟鍒楄〃
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
						/* 鍑忓皯鏁翠綋闂磋窛 */
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
							/* 鍑忓皯 "=" 涓よ竟鐨勮竟璺?*/
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
			/* 澧炲姞瀛楃闂磋窛 */
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

	.orders-card {
		border-radius: 8px;

		:deep(.el-card__body) {
			padding: 0;
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