<template>
  <div class="user-management-page">
    <!-- 用户概览 -->
    <div class="user-summary">
      <div class="summary-cards">
        <div class="summary-card">
          <div class="card-content">
            <div class="card-info">
              <div class="card-label">店主数量</div>
              <div class="card-value">
                <span class="value-number">{{ summaryData.ownerCount }}</span>
              </div>
            </div>
            <el-button type="primary" size="small" @click="handleNewOwner">新增店主</el-button>
          </div>
        </div>
        <div class="summary-card">
          <div class="card-content">
            <div class="card-info">
              <div class="card-label">运营数量</div>
              <div class="card-value">
                <span class="value-number">{{ summaryData.operatorCount }}</span>
              </div>
            </div>
            <el-button type="primary" size="small" @click="handleNewOperator">新增运营</el-button>
          </div>
        </div>
      </div>
    </div>

    <!-- 筛选区域 -->
    <div class="filter-section">
      <el-input
        v-model="filterForm.keyword"
        placeholder="输入ID/姓名"
        :prefix-icon="Search"
        clearable
        style="width: 180px"
      />
      <el-select v-model="filterForm.userStatus" placeholder="选择用户状态" clearable style="width: 150px">
        <el-option label="全部" value="" />
        <el-option label="正常" value="normal" />
        <el-option label="禁用" value="disabled" />
      </el-select>
      <div class="filter-actions">
        <el-button type="primary" @click="handleSearch">查询</el-button>
        <el-button @click="handleReset">重置</el-button>
      </div>
    </div>

    <!-- Tab和表格 -->
    <el-card class="user-card" shadow="never">
      <div class="user-header">
        <el-tabs v-model="activeTab" @tab-change="handleTabChange">
          <el-tab-pane label="店主列表" name="owner" />
          <el-tab-pane label="运营列表" name="operator" />
        </el-tabs>
        <div class="header-actions">
          <el-button :icon="Download">导出报表</el-button>
        </div>
      </div>

      <!-- 店主列表 -->
      <el-table v-if="activeTab === 'owner'" :data="userList" style="width: 100%" v-loading="loading">
        <el-table-column prop="name" label="姓名" min-width="140">
          <template #default="{ row }">
            <div class="user-info">
              <div class="user-avatar"></div>
              <span>{{ row.name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="userId" label="店主ID" min-width="120" />
        <el-table-column prop="phone" label="电话" min-width="120" />
        <el-table-column prop="email" label="邮箱" min-width="160" />
        <el-table-column prop="storeCount" label="店铺数量" min-width="100" />
        <el-table-column prop="status" label="用户状态" min-width="100">
          <template #default="{ row }">
            <span class="status-text">{{ row.status }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" min-width="80" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="handleManage(row)">管理</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 运营列表 -->
      <el-table v-if="activeTab === 'operator'" :data="userList" style="width: 100%" v-loading="loading">
        <el-table-column prop="name" label="姓名" min-width="140">
          <template #default="{ row }">
            <div class="user-info">
              <div class="user-avatar"></div>
              <span>{{ row.name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="userId" label="运营ID" min-width="120" />
        <el-table-column prop="phone" label="电话" min-width="120" />
        <el-table-column prop="email" label="邮箱" min-width="160" />
        <el-table-column prop="storeCount" label="店铺数量" min-width="100" />
        <el-table-column prop="status" label="用户状态" min-width="100">
          <template #default="{ row }">
            <span class="status-text">{{ row.status }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" min-width="80" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="handleManage(row)">管理</el-button>
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

    <!-- 新增店主弹框 -->
    <el-dialog
      v-model="ownerDialogVisible"
      title="新增店主"
      width="500px"
      :close-on-click-modal="false"
    >
      <el-form :model="ownerForm" label-width="80px">
        <el-form-item label="姓名">
          <el-input v-model="ownerForm.name" placeholder="请输入" />
        </el-form-item>
        <el-form-item label="电话">
          <el-input v-model="ownerForm.phone" placeholder="请输入" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="ownerForm.email" placeholder="请输入" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="ownerDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitOwner">确定</el-button>
      </template>
    </el-dialog>

    <!-- 新增运营弹框 -->
    <el-dialog
      v-model="operatorDialogVisible"
      title="新增运营"
      width="500px"
      :close-on-click-modal="false"
    >
      <el-form :model="operatorForm" label-width="80px">
        <el-form-item label="姓名">
          <el-input v-model="operatorForm.name" placeholder="请输入" />
        </el-form-item>
        <el-form-item label="电话">
          <el-input v-model="operatorForm.phone" placeholder="请输入" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="operatorForm.email" placeholder="请输入" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="operatorDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitOperator">确定</el-button>
      </template>
    </el-dialog>

    <!-- 管理弹框 -->
    <el-dialog
      v-model="manageDialogVisible"
      :title="currentUser.name"
      width="400px"
      :close-on-click-modal="false"
    >
      <div class="manage-dialog">
        <!-- 头部信息 -->
        <div class="manage-header">
          <div class="header-info">
            <div class="user-name">{{ currentUser.name }}</div>
            <div class="user-status-label">用户状态</div>
          </div>
          <div class="header-avatar"></div>
        </div>

        <!-- 用户状态 -->
        <div class="status-section">
          <span class="section-label">用户状态：</span>
          <el-radio-group v-model="currentUser.statusType">
            <el-radio value="normal">正常</el-radio>
            <el-radio value="frozen">冻结</el-radio>
            <el-radio value="banned">封禁</el-radio>
          </el-radio-group>
        </div>

        <!-- 用户信息 -->
        <div class="info-section">
          <div class="section-title">用户信息</div>
          <div class="info-row">
            <span class="info-label">姓名：</span>
            <span class="info-value">{{ currentUser.name }}</span>
            <div class="info-avatar">J</div>
          </div>
          <div class="info-row">
            <span class="info-label">店主ID：</span>
            <span class="info-value">{{ currentUser.userId }}</span>
          </div>
          <div class="info-row">
            <span class="info-label">电话：</span>
            <span class="info-value">{{ currentUser.phone }}</span>
          </div>
          <div class="info-row">
            <span class="info-label">邮箱：</span>
            <span class="info-value">{{ currentUser.email }}</span>
          </div>
        </div>

        <!-- 店铺列表 -->
        <div class="store-section">
          <div class="section-title">店铺数量：{{ currentUser.storeCount }}</div>
          <div class="store-list">
            <div class="store-item" v-for="(store, index) in storeList" :key="index">
              <div class="store-icon"></div>
              <div class="store-info">
                <div class="store-name">{{ store.name }}</div>
                <div class="store-status">
                  <span>店铺状态</span>
                  <span>授权状态</span>
                  <span>运营状态</span>
                </div>
              </div>
              <el-button type="primary" link size="small">查看</el-button>
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <el-button @click="manageDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveManage">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Search, Download } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

interface UserRecord {
  name: string
  userId: string
  phone: string
  email: string
  storeCount: number
  status: string
}

const activeTab = ref('owner')
const loading = ref(false)
const ownerDialogVisible = ref(false)
const operatorDialogVisible = ref(false)
const manageDialogVisible = ref(false)

const currentUser = reactive({
  name: '',
  userId: '',
  phone: '',
  email: '',
  storeCount: 0,
  statusType: 'normal'
})

const storeList = ref([
  { name: '店铺名称示例文字占位符' },
  { name: '店铺名称示例文字占位符' },
  { name: '店铺名称示例文字占位符' },
  { name: '店铺名称示例文字占位符' },
  { name: '店铺名称示例文字占位符' }
])

const summaryData = reactive({
  ownerCount: 52,
  operatorCount: 8
})

const filterForm = reactive({
  keyword: '',
  userStatus: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 123
})

const ownerForm = reactive({
  name: '',
  phone: '',
  email: ''
})

const operatorForm = reactive({
  name: '',
  phone: '',
  email: ''
})

const userList = ref<UserRecord[]>([])

const fetchUserList = async () => {
  loading.value = true
  try {
    userList.value = Array.from({ length: 7 }, () => ({
      name: '名字示例文字',
      userId: '1234567890',
      phone: '1234567890',
      email: '12345678@qq.com',
      storeCount: 5,
      status: '正常'
    }))
    pagination.total = 123
  } catch (err) {
    console.error('获取用户列表失败:', err)
  } finally {
    loading.value = false
  }
}

const handleTabChange = () => {
  pagination.page = 1
  fetchUserList()
}

const handleSearch = () => {
  pagination.page = 1
  fetchUserList()
}

const handleReset = () => {
  filterForm.keyword = ''
  filterForm.userStatus = ''
  pagination.page = 1
  fetchUserList()
}

const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  pagination.page = 1
  fetchUserList()
}

const handlePageChange = (page: number) => {
  pagination.page = page
  fetchUserList()
}

const handleNewOwner = () => {
  ownerForm.name = ''
  ownerForm.phone = ''
  ownerForm.email = ''
  ownerDialogVisible.value = true
}

const handleNewOperator = () => {
  operatorForm.name = ''
  operatorForm.phone = ''
  operatorForm.email = ''
  operatorDialogVisible.value = true
}

const handleSubmitOwner = () => {
  ownerDialogVisible.value = false
  ElMessage.success('店主创建成功')
  fetchUserList()
}

const handleSubmitOperator = () => {
  operatorDialogVisible.value = false
  ElMessage.success('运营创建成功')
  fetchUserList()
}

const handleManage = (row: UserRecord) => {
  currentUser.name = row.name
  currentUser.userId = row.userId
  currentUser.phone = row.phone
  currentUser.email = row.email
  currentUser.storeCount = row.storeCount
  currentUser.statusType = 'normal'
  manageDialogVisible.value = true
}

const handleSaveManage = () => {
  manageDialogVisible.value = false
  ElMessage.success('保存成功')
  fetchUserList()
}

onMounted(() => {
  fetchUserList()
})
</script>

<style scoped lang="scss">
.user-management-page {
  padding: 20px;
  background: #f5f7fa;
  min-height: 100%;
}

.user-summary {
  margin-bottom: 20px;

  .summary-cards {
    display: flex;
    gap: 16px;
  }

  .summary-card {
    flex: 1;
    background: #fff;
    border-radius: 8px;
    padding: 20px 24px;
    border: 1px solid #ebeef5;

    .card-content {
      display: flex;
      justify-content: space-between;
      align-items: center;
    }

    .card-info {
      .card-label {
        font-size: 14px;
        color: #909399;
        margin-bottom: 8px;
      }

      .card-value {
        .value-number {
          font-size: 32px;
          font-weight: 600;
          color: #303133;
        }
      }
    }
  }
}

.filter-section {
  display: flex;
  gap: 12px;
  align-items: center;
  margin-bottom: 20px;
  padding: 16px 20px;
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 8px;

  .filter-actions {
    margin-left: auto;
    display: flex;
    gap: 12px;
  }
}

.user-card {
  border-radius: 8px;

  :deep(.el-card__body) {
    padding: 0;
  }
}

.user-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 20px;
  border-bottom: 1px solid #ebeef5;

  :deep(.el-tabs) {
    .el-tabs__header {
      margin-bottom: 0;
    }

    .el-tabs__nav-wrap::after {
      display: none;
    }
  }

  .header-actions {
    display: flex;
    gap: 12px;
  }
}

.pagination-wrapper {
  padding: 20px;
  display: flex;
  justify-content: center;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;

  .user-avatar {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    background: #e0e0e0;
  }
}

.status-text {
  color: #67c23a;
}

.manage-dialog {
  .manage-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 20px;
    padding-bottom: 16px;
    border-bottom: 1px solid #ebeef5;

    .header-info {
      .user-name {
        font-size: 18px;
        font-weight: 600;
        color: #303133;
        margin-bottom: 4px;
      }

      .user-status-label {
        font-size: 12px;
        color: #909399;
      }
    }

    .header-avatar {
      width: 60px;
      height: 60px;
      border-radius: 50%;
      background: #e0e0e0;
    }
  }

  .status-section {
    display: flex;
    align-items: center;
    gap: 16px;
    margin-bottom: 20px;
    padding-bottom: 16px;
    border-bottom: 1px solid #ebeef5;

    .section-label {
      font-size: 14px;
      color: #606266;
    }
  }

  .info-section {
    margin-bottom: 20px;
    padding-bottom: 16px;
    border-bottom: 1px solid #ebeef5;

    .section-title {
      font-size: 14px;
      font-weight: 500;
      color: #303133;
      margin-bottom: 12px;
    }

    .info-row {
      display: flex;
      align-items: center;
      margin-bottom: 8px;

      .info-label {
        font-size: 14px;
        color: #909399;
        min-width: 60px;
      }

      .info-value {
        font-size: 14px;
        color: #303133;
        flex: 1;
      }

      .info-avatar {
        width: 24px;
        height: 24px;
        border-radius: 50%;
        background: #67c23a;
        color: #fff;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 12px;
        margin-left: 8px;
      }
    }
  }

  .store-section {
    .section-title {
      font-size: 14px;
      font-weight: 500;
      color: #303133;
      margin-bottom: 12px;
    }

    .store-list {
      max-height: 200px;
      overflow-y: auto;
    }

    .store-item {
      display: flex;
      align-items: center;
      gap: 12px;
      padding: 12px 0;
      border-bottom: 1px solid #f0f0f0;

      &:last-child {
        border-bottom: none;
      }

      .store-icon {
        width: 40px;
        height: 40px;
        border-radius: 4px;
        background: #e0e0e0;
      }

      .store-info {
        flex: 1;

        .store-name {
          font-size: 14px;
          color: #303133;
          margin-bottom: 4px;
        }

        .store-status {
          display: flex;
          gap: 12px;
          font-size: 12px;
          color: #909399;
        }
      }
    }
  }
}
</style>
