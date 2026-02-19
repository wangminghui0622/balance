<template>
  <div class="finance-account-page">
    <div class="page-header">
      <h1 class="page-title">收款账户</h1>
    </div>

    <!-- 电子钱包 -->
    <div class="section">
      <div class="section-title">
        <span class="title-bar"></span>
        <span>电子钱包</span>
      </div>
      <div class="cards-grid">
        <div v-for="(wallet, index) in walletList" :key="index" class="account-card wallet-card">
          <div class="card-header">
            <div class="card-icon alipay">
              <span class="icon-text">支</span>
            </div>
            <div class="card-info">
              <div class="card-name">{{ wallet.name }}</div>
              <el-tag size="small" type="primary">活跃</el-tag>
            </div>
            <div class="card-default" v-if="wallet.isDefault">
              <el-icon><CircleCheck /></el-icon>
              <span>默认</span>
            </div>
          </div>
          <div class="card-details">
            <div class="detail-item">
              <span class="label">账号</span>
              <span class="value">{{ wallet.accountNumber }}</span>
            </div>
            <div class="detail-item">
              <span class="label">持有人</span>
              <span class="value">{{ wallet.holderName }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 收款账户 -->
    <div class="section">
      <div class="section-title">
        <span class="title-bar"></span>
        <span>收款账户</span>
      </div>
      <div class="cards-grid">
        <div v-for="(account, index) in bankAccountList" :key="index" class="account-card bank-card">
          <div class="card-header">
            <div class="card-icon hsbc">
              <span class="icon-text">H</span>
            </div>
            <div class="card-info">
              <div class="card-name">{{ account.bankName }}</div>
              <el-tag size="small" type="danger">未激活</el-tag>
            </div>
            <div class="card-default-radio">
              <el-radio v-model="defaultBankId" :value="account.id" @change="handleSetDefault(account)">默认</el-radio>
            </div>
          </div>
          <div class="card-details">
            <div class="detail-item">
              <span class="label">账号</span>
              <span class="value">{{ account.accountNumber }}</span>
            </div>
            <div class="detail-item">
              <span class="label">币种</span>
              <span class="value">{{ account.currency }}</span>
            </div>
          </div>
        </div>
        
        <!-- 新增收款账户 -->
        <div class="account-card add-card" @click="handleAddAccount">
          <div class="add-content">
            <el-icon :size="20"><Plus /></el-icon>
            <span>新增收款账户</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 添加/编辑账户弹框 -->
    <el-dialog v-model="accountDialogVisible" :title="isEdit ? '编辑账户' : '添加账户'" width="500px">
      <el-form :model="accountForm" label-width="100px" :rules="accountRules" ref="accountFormRef">
        <el-form-item label="账户类型" prop="type">
          <el-select v-model="accountForm.type" style="width: 100%">
            <el-option label="银行卡" value="bank" />
            <el-option label="支付宝" value="alipay" />
            <el-option label="微信" value="wechat" />
          </el-select>
        </el-form-item>
        <el-form-item label="银行名称" prop="bankName" v-if="accountForm.type === 'bank'">
          <el-input v-model="accountForm.bankName" placeholder="请输入银行名称" />
        </el-form-item>
        <el-form-item label="账户号码" prop="accountNumber">
          <el-input v-model="accountForm.accountNumber" placeholder="请输入账户号码" />
        </el-form-item>
        <el-form-item label="持卡人姓名" prop="holderName">
          <el-input v-model="accountForm.holderName" placeholder="请输入持卡人姓名" />
        </el-form-item>
        <el-form-item label="开户行" prop="bankBranch" v-if="accountForm.type === 'bank'">
          <el-input v-model="accountForm.bankBranch" placeholder="请输入开户行" />
        </el-form-item>
        <el-form-item label="设为默认">
          <el-switch v-model="accountForm.isDefault" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="accountDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmAccount">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { CircleCheck, Plus } from '@element-plus/icons-vue'

interface Wallet {
  id: number
  name: string
  accountNumber: string
  holderName: string
  isDefault: boolean
}

interface BankAccount {
  id: number
  bankName: string
  accountNumber: string
  currency: string
  isDefault: boolean
}

const accountDialogVisible = ref(false)
const isEdit = ref(false)
const accountFormRef = ref()
const defaultBankId = ref(1)

const walletList = ref<Wallet[]>([
  { id: 1, name: '支付宝', accountNumber: '12345678910121314115', holderName: '12345678910', isDefault: true }
])

const bankAccountList = ref<BankAccount[]>([
  { id: 1, bankName: '汇丰银行>', accountNumber: '12345678910121314115', currency: 'CNY', isDefault: false }
])

const accountForm = reactive({
  id: 0,
  type: 'bank',
  bankName: '',
  accountNumber: '',
  holderName: '',
  bankBranch: '',
  isDefault: false
})

const accountRules = {
  type: [{ required: true, message: '请选择账户类型', trigger: 'change' }],
  bankName: [{ required: true, message: '请输入银行名称', trigger: 'blur' }],
  accountNumber: [{ required: true, message: '请输入账户号码', trigger: 'blur' }],
  holderName: [{ required: true, message: '请输入持卡人姓名', trigger: 'blur' }]
}

function handleAddAccount() {
  isEdit.value = false
  accountForm.id = 0
  accountForm.type = 'bank'
  accountForm.bankName = ''
  accountForm.accountNumber = ''
  accountForm.holderName = ''
  accountForm.bankBranch = ''
  accountForm.isDefault = false
  accountDialogVisible.value = true
}

function handleSetDefault(account: BankAccount) {
  defaultBankId.value = account.id
  ElMessage.success('已设为默认账户')
}

function confirmAccount() {
  ElMessage.success('添加成功')
  accountDialogVisible.value = false
}

onMounted(() => {
  // 加载数据
})
</script>

<style scoped lang="scss">
.finance-account-page {
  .page-header {
    margin-bottom: 20px;

    .page-title {
      font-size: 20px;
      font-weight: 500;
      color: #303133;
      margin: 0;
    }
  }

  .section {
    margin-bottom: 24px;
    padding: 20px;
    background: #fff;
    border: 1px solid #ebeef5;
    border-radius: 8px;

    .section-title {
      display: flex;
      align-items: center;
      gap: 8px;
      margin-bottom: 20px;
      font-size: 14px;
      font-weight: 500;
      color: #303133;

      .title-bar {
        width: 3px;
        height: 14px;
        background: #ff6a3a;
        border-radius: 2px;
      }
    }
  }

  .cards-grid {
    display: flex;
    gap: 20px;
    flex-wrap: wrap;
  }

  .account-card {
    width: 260px;
    border-radius: 12px;
    overflow: hidden;

    &.wallet-card {
      background: #1890ff;
      color: #fff;
      padding: 16px 20px;

      .card-header {
        display: flex;
        align-items: center;
        margin-bottom: 20px;

        .card-icon {
          width: 36px;
          height: 36px;
          background: rgba(255, 255, 255, 0.25);
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;
          margin-right: 10px;

          .icon-text {
            font-size: 18px;
            font-weight: bold;
          }
        }

        .card-info {
          flex: 1;

          .card-name {
            font-size: 15px;
            font-weight: 500;
            margin-bottom: 4px;
          }

          :deep(.el-tag) {
            background: rgba(255, 255, 255, 0.25);
            border: none;
            color: #fff;
            font-size: 11px;
            padding: 0 6px;
            height: 20px;
            line-height: 20px;
          }
        }

        .card-default {
          display: flex;
          align-items: center;
          gap: 4px;
          font-size: 12px;
          opacity: 0.9;
        }
      }

      .card-details {
        display: flex;
        justify-content: space-between;
        padding-top: 12px;
        border-top: 1px solid rgba(255, 255, 255, 0.15);

        .detail-item {
          .label {
            font-size: 11px;
            opacity: 0.7;
            display: block;
            margin-bottom: 6px;
          }

          .value {
            font-size: 13px;
            font-weight: 500;
          }
        }
      }
    }

    &.bank-card {
      background: #f5222d;
      color: #fff;
      padding: 16px 20px;

      .card-header {
        display: flex;
        align-items: center;
        margin-bottom: 20px;

        .card-icon {
          width: 36px;
          height: 36px;
          background: rgba(255, 255, 255, 0.25);
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;
          margin-right: 10px;

          .icon-text {
            font-size: 18px;
            font-weight: bold;
          }
        }

        .card-info {
          flex: 1;

          .card-name {
            font-size: 15px;
            font-weight: 500;
            margin-bottom: 4px;
          }

          :deep(.el-tag) {
            background: rgba(255, 255, 255, 0.25);
            border: none;
            color: #fff;
            font-size: 11px;
            padding: 0 6px;
            height: 20px;
            line-height: 20px;
          }
        }

        .card-default-radio {
          :deep(.el-radio) {
            color: rgba(255, 255, 255, 0.9);
            font-size: 12px;

            .el-radio__inner {
              background: transparent;
              border-color: rgba(255, 255, 255, 0.8);
              width: 14px;
              height: 14px;
            }

            .el-radio__label {
              color: rgba(255, 255, 255, 0.9);
              padding-left: 4px;
            }
          }
        }
      }

      .card-details {
        display: flex;
        justify-content: space-between;
        padding-top: 12px;
        border-top: 1px solid rgba(255, 255, 255, 0.15);

        .detail-item {
          .label {
            font-size: 11px;
            opacity: 0.7;
            display: block;
            margin-bottom: 6px;
          }

          .value {
            font-size: 13px;
            font-weight: 500;
          }
        }
      }
    }

    &.add-card {
      border: 1px dashed #dcdfe6;
      background: #fafafa;
      display: flex;
      align-items: center;
      justify-content: center;
      min-height: 130px;
      cursor: pointer;
      transition: all 0.3s;

      &:hover {
        border-color: #409eff;
        color: #409eff;
      }

      .add-content {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 8px;
        color: #909399;

        span {
          font-size: 14px;
        }
      }
    }
  }
}
</style>
