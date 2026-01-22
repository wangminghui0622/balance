<template>
  <div class="stores-page">
    <!-- 授权店铺BANNER -->
    <el-card class="auth-banner">
      <div class="banner-content">
        <div class="banner-left">
          <h3>授权店铺BANNER</h3>
          <p>授权你的店铺,开启合作之旅。</p>
        </div>
        <el-button type="primary" size="large" @click="handleQuickAuth">
          马上授权
        </el-button>
      </div>
    </el-card>

    <!-- 筛选和搜索区域 -->
    <el-card class="filter-card">
      <el-form :model="filterForm" :inline="true" class="filter-form">
        <el-form-item label="店铺名称/店铺编号">
          <el-input
            v-model="filterForm.keyword"
            placeholder="请输入店铺名称或编号"
            clearable
            class="filter-input"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="选择店铺状态">
          <el-select v-model="filterForm.shopStatus" placeholder="请选择" clearable class="filter-select-small">
            <el-option label="正常" value="normal" />
            <el-option label="暂停" value="paused" />
            <el-option label="关闭" value="closed" />
          </el-select>
        </el-form-item>
        <el-form-item label="选择授权状态">
          <el-select v-model="filterForm.authStatus" placeholder="请选择" clearable class="filter-select-small">
            <el-option label="已授权" value="authorized" />
            <el-option label="未授权" value="unauthorized" />
            <el-option label="已过期" value="expired" />
          </el-select>
        </el-form-item>
        <el-form-item label="选择运营状态">
          <el-select v-model="filterForm.operationStatus" placeholder="请选择" clearable class="filter-select-small">
            <el-option label="运营中" value="operating" />
            <el-option label="暂停" value="paused" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleQuery">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 店铺列表 -->
    <el-card class="stores-list-card">
      <template #header>
        <div class="card-header">
          <span>店铺列表 ({{ storeList.length }})</span>
          <el-button type="default" size="small" @click="handleExportReport">
            导出报表
          </el-button>
        </div>
      </template>
      <div class="stores-list">
        <div
          v-for="(store, index) in storeList"
          :key="index"
          class="store-item"
        >
          <!-- 左侧A区域 -->
          <div class="store-left">
            <el-avatar :size="72" shape="square" :src="store.avatar" class="store-avatar" />
            <div class="store-info">
              <div class="store-name">{{ store.name }}</div>
              <div class="store-tags">
                <el-tag size="small" :type="getShopStatusType(store.shopStatus)">
                  {{ getShopStatusText(store.shopStatus) }}
                </el-tag>
                <el-tag size="small" :type="getAuthStatusType(store.authStatus)">
                  {{ getAuthStatusText(store.authStatus) }}
                </el-tag>
                <el-tag size="small" :type="getOperationStatusType(store.operationStatus)">
                  {{ getOperationStatusText(store.operationStatus) }}
                </el-tag>
              </div>
            </div>
          </div>

          <!-- 中间B区域 -->
          <div class="store-middle">
            <div class="store-details">
              <div class="detail-item">
                <div class="detail-label">店铺编号</div>
                <div class="detail-value">S{{ store.storeId }}</div>
              </div>
              <div class="detail-item">
                <div class="detail-label">店铺账户</div>
                <div class="detail-value">{{ store.account }}</div>
              </div>
              <div class="detail-item">
                <div class="detail-label">主体类型</div>
                <div class="detail-value">{{ store.entityType }}</div>
              </div>
              <div class="detail-item">
                <div class="detail-label">主体姓名</div>
                <div class="detail-value">{{ store.entityName }}</div>
              </div>
              <div class="detail-item">
                <div class="detail-label">店铺健康</div>
                <div class="detail-value">{{ store.health }}</div>
              </div>
              <div class="detail-item">
                <div class="detail-label">授权到期时间</div>
                <div class="detail-value">{{ store.authExpireTime }}</div>
              </div>
            </div>
          </div>

          <!-- 右侧C区域 -->
          <div class="store-right">
            <el-button size="small" @click="handleMore(store)" class="action-btn">更多</el-button>
            <el-button type="success" size="small" @click="handleManage(store)" class="action-btn">经营</el-button>
          </div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'

interface Store {
  avatar: string
  name: string
  storeId: string
  shopId: number // Shopee Shop ID
  account: string
  entityType: string // 个人/企业
  entityName: string
  health: string
  authExpireTime: string
  shopStatus: string // normal/paused/closed
  authStatus: string // authorized/unauthorized/expired
  operationStatus: string // operating/paused
  authLoading?: boolean
}

const filterForm = reactive({
  keyword: '',
  shopStatus: '',
  authStatus: '',
  operationStatus: ''
})

const exportReport = ref(false)

const storeList = ref<Store[]>([
  {
    avatar: '',
    name: '店铺名铺名铺名铺名铺名dasqwet',
    storeId: '1234567890',
    shopId: 226445936,
    account: '1234567890@gmail.com',
    entityType: '个人',
    entityName: '张小凡',
    health: '占位符',
    authExpireTime: '2026-12-12 23:59:59',
    shopStatus: 'normal',
    authStatus: 'unauthorized',
    operationStatus: 'operating',
    authLoading: false
  },
  {
    avatar: '',
    name: '店铺名称示例文字占位符文字占位符',
    storeId: '1234567891',
    shopId: 226445937,
    account: '1234567891@gmail.com',
    entityType: '企业',
    entityName: '示例文字占位符',
    health: '占位符',
    authExpireTime: '2026-12-12 23:59:59',
    shopStatus: 'normal',
    authStatus: 'authorized',
    operationStatus: 'operating',
    authLoading: false
  },
  {
    avatar: '',
    name: '店铺名称示例文字占位符文字占位符',
    storeId: '1234567892',
    shopId: 226445938,
    account: '1234567892@gmail.com',
    entityType: '个人',
    entityName: '张小凡',
    health: '占位符',
    authExpireTime: '2026-12-12 23:59:59',
    shopStatus: 'normal',
    authStatus: 'expired',
    operationStatus: 'paused',
    authLoading: false
  },
  {
    avatar: '',
    name: '店铺名称示例文字占位符文字占位符',
    storeId: '1234567893',
    shopId: 226445939,
    account: '1234567893@gmail.com',
    entityType: '企业',
    entityName: '示例文字占位符',
    health: '占位符',
    authExpireTime: '2026-12-12 23:59:59',
    shopStatus: 'paused',
    authStatus: 'authorized',
    operationStatus: 'operating',
    authLoading: false
  },
  {
    avatar: '',
    name: '店铺名称示例文字占位符文字占位符',
    storeId: '1234567894',
    shopId: 226445940,
    account: '1234567894@gmail.com',
    entityType: '个人',
    entityName: '张小凡',
    health: '占位符',
    authExpireTime: '2026-12-12 23:59:59',
    shopStatus: 'normal',
    authStatus: 'unauthorized',
    operationStatus: 'operating',
    authLoading: false
  },
  {
    avatar: '',
    name: '店铺名称示例文字占位符文字占位符',
    storeId: '1234567895',
    shopId: 226445941,
    account: '1234567895@gmail.com',
    entityType: '企业',
    entityName: '示例文字占位符',
    health: '占位符',
    authExpireTime: '2026-12-12 23:59:59',
    shopStatus: 'normal',
    authStatus: 'authorized',
    operationStatus: 'operating',
    authLoading: false
  }
])

const getShopStatusType = (status: string) => {
  const map: Record<string, string> = {
    normal: 'success',
    paused: 'warning',
    closed: 'danger'
  }
  return map[status] || 'info'
}

const getShopStatusText = (status: string) => {
  const map: Record<string, string> = {
    normal: '店铺状态',
    paused: '暂停',
    closed: '关闭'
  }
  return map[status] || status
}

const getAuthStatusType = (status: string) => {
  const map: Record<string, string> = {
    authorized: 'success',
    unauthorized: 'warning',
    expired: 'danger'
  }
  return map[status] || 'info'
}

const getAuthStatusText = (status: string) => {
  const map: Record<string, string> = {
    authorized: '授权状态',
    unauthorized: '未授权',
    expired: '已过期'
  }
  return map[status] || status
}

const getOperationStatusType = (status: string) => {
  const map: Record<string, string> = {
    operating: 'success',
    paused: 'warning'
  }
  return map[status] || 'info'
}

const getOperationStatusText = (status: string) => {
  const map: Record<string, string> = {
    operating: '运营状态',
    paused: '暂停'
  }
  return map[status] || status
}

const handleQuickAuth = () => {
  ElMessage.info('快速授权功能开发中...')
}

const handleQuery = () => {
  ElMessage.success('查询功能开发中...')
}

const handleReset = () => {
  filterForm.keyword = ''
  filterForm.shopStatus = ''
  filterForm.authStatus = ''
  filterForm.operationStatus = ''
  ElMessage.success('已重置筛选条件')
}

const handleMore = (store: Store) => {
  ElMessage.info(`更多操作: ${store.name}`)
}

const handleManage = (store: Store) => {
  ElMessage.info(`经营店铺: ${store.name}`)
}

const handleExportReport = () => {
  if (exportReport.value) {
    ElMessage.success('正在导出报表...')
  } else {
    ElMessage.info('请先勾选"导出报表"')
  }
}
</script>

<style scoped lang="scss">
.stores-page {
  padding: 20px;
}

.auth-banner {
  margin-bottom: 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;

  :deep(.el-card__body) {
    padding: 30px;
  }

  .banner-content {
    display: flex;
    justify-content: space-between;
    align-items: center;
    color: white;

    .banner-left {
      h3 {
        margin: 0 0 10px 0;
        font-size: 24px;
        font-weight: 600;
      }

      p {
        margin: 0;
        font-size: 16px;
        opacity: 0.9;
      }
    }
  }
}

.filter-card {
  margin-bottom: 20px;

  .filter-form {
    margin-bottom: 0;
    display: flex;
    flex-wrap: nowrap;
    align-items: center;
    justify-content: flex-start;
    gap: 16px;

    .el-form-item {
      flex-shrink: 0;
    }

    .el-form-item:last-child {
      margin-left: auto;
      margin-right: -3px;
      flex-shrink: 0;
    }
  }

  .filter-input {
    width: 25ch;
  }

  .filter-select-small {
    width: 13ch;
  }
}

.stores-list-card {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-weight: 500;
    font-size: 16px;
  }

  .stores-list {
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .store-item {
    display: flex;
    align-items: stretch;
    gap: 20px;
    padding: 8px 0;
    border: 1px solid #e4e7ed;
    border-radius: 8px;
    transition: all 0.3s;

    &:hover {
      border-color: #409eff;
      box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
    }

    // 左侧A区域
    .store-left {
      display: flex;
      align-items: center;
      gap: 12px;
      flex: 1;
      min-width: 0;
      margin-left: 10px;

      .store-avatar {
        flex-shrink: 0;
        background-color: #f5f7fa;
        border-radius: 8px;
      }

      .store-info {
        display: flex;
        flex-direction: column;
        justify-content: center;
        gap: 10px;
        flex: 1;
        min-width: 0;

        .store-name {
          font-size: 16px;
          font-weight: 500;
          color: #000000;
          line-height: 1.4;
        }

        .store-tags {
          display: flex;
          gap: 8px;
          flex-wrap: wrap;
        }
      }
    }

    // 中间B区域
    .store-middle {
      flex: 1;
      min-width: 0;
      display: flex;
      align-items: center;
      margin-right: -230px;

      .store-details {
        display: grid;
        grid-template-columns: repeat(3, 1fr);
        gap: 8px 32px;
        width: 100%;

        .detail-item {
          display: flex;
          flex-direction: column;
          align-items: center;
          font-size: 14px;
          line-height: 1.5;

          .detail-label {
            color: #606266;
            margin-bottom: 2px;
            white-space: nowrap;
            font-size: 13px;
            font-weight: 500;
          }

          .detail-value {
            color: #c0c4cc;
            text-align: center;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
            width: 100%;
          }
        }
      }
    }

    // 右侧C区域
    .store-right {
      display: flex;
      flex-direction: column;
      justify-content: center;
      align-items: flex-end;
      gap: 12px;
      flex: 1;
      min-width: 0;
      margin-right: 10px;

      .action-btn {
        min-width: 80px;
      }
    }
  }
}

@media (max-width: 1200px) {
  .store-item {
    .store-middle {
      .store-details {
        grid-template-columns: repeat(2, 1fr);
      }
    }
  }
}

@media (max-width: 768px) {
  .store-item {
    flex-direction: column;

    .store-right {
      width: 100%;
      flex-direction: row;
      justify-content: flex-end;
    }

    .store-middle {
      .store-details {
        grid-template-columns: 1fr;
      }
    }
  }
}
</style>
