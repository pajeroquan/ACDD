<template>
  <div>
    <div class="toolbar">
      <el-button type="primary" @click="load">刷新报表</el-button>
    </div>
    <el-table :data="list" v-loading="loading" border>
      <el-table-column prop="union_id" label="工会ID" width="90" />
      <el-table-column prop="union_name" label="工会名称" min-width="160" />
      <el-table-column prop="order_count" label="订单数" width="100" />
      <el-table-column label="分成总额(分)" min-width="140">
        <template #default="{ row }">{{ row.amount_fen }}</template>
      </el-table-column>
      <el-table-column label="待结算(分)" min-width="140">
        <template #default="{ row }">{{ row.pending_fen }}</template>
      </el-table-column>
      <el-table-column label="约合(元)" width="120">
        <template #default="{ row }">{{ ((row.amount_fen || 0) / 100).toFixed(2) }}</template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { commissionReport } from '../api'

const list = ref([])
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    list.value = (await commissionReport()) || []
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.toolbar {
  margin-bottom: 12px;
}
</style>
