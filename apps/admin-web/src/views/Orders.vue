<template>
  <div>
    <div class="toolbar">
      <el-select v-model="status" clearable placeholder="订单状态" style="width: 160px" @change="load">
        <el-option label="pending_pay" value="pending_pay" />
        <el-option label="paid" value="paid" />
        <el-option label="notified" value="notified" />
        <el-option label="cancelled" value="cancelled" />
      </el-select>
      <el-button @click="load">刷新</el-button>
    </div>
    <el-table :data="list" v-loading="loading" border>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="order_no" label="订单号" min-width="180" />
      <el-table-column label="搭子" min-width="120">
        <template #default="{ row }">{{ row.partner?.nickname || row.partner_id }}</template>
      </el-table-column>
      <el-table-column label="预约日期" width="120">
        <template #default="{ row }">{{ formatDate(row.schedule_date) }}</template>
      </el-table-column>
      <el-table-column prop="start_time" label="开始" width="100" />
      <el-table-column prop="duration_hours" label="时长" width="80" />
      <el-table-column label="金额(分)" width="110">
        <template #default="{ row }">{{ row.total_amount_fen }}</template>
      </el-table-column>
      <el-table-column label="分成(分)" width="110">
        <template #default="{ row }">{{ row.commission_amount_fen }}</template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="110" />
      <el-table-column label="创建时间" width="170">
        <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="100" fixed="right">
        <template #default="{ row }">
          <el-button
            v-if="canRefund(row.status)"
            link
            type="danger"
            @click="onRefund(row)"
          >退款</el-button>
        </template>
      </el-table-column>
    </el-table>
    <div class="pager">
      <el-pagination
        background
        layout="total, prev, pager, next"
        :total="total"
        v-model:current-page="page"
        :page-size="pageSize"
        @current-change="load"
      />
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listOrders, refundOrder } from '../api'

const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const status = ref('')
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    const data = await listOrders({
      page: page.value,
      page_size: pageSize,
      status: status.value || undefined,
    })
    list.value = data?.list || []
    total.value = data?.total || 0
  } finally {
    loading.value = false
  }
}

function canRefund(s) {
  return ['pending_pay', 'paid', 'notified', 'in_service'].includes(s)
}

async function onRefund(row) {
  await ElMessageBox.confirm(`确认退款订单 ${row.order_no}？`, '退款确认', { type: 'warning' })
  await refundOrder(row.id)
  ElMessage.success('已退款')
  load()
}

function formatDate(v) {
  if (!v) return '-'
  return String(v).slice(0, 10)
}

function formatTime(v) {
  if (!v) return '-'
  return new Date(v).toLocaleString()
}

onMounted(load)
</script>

<style scoped>
.toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}
.pager {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
