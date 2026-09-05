<template>
  <div class="match-page">
    <el-card shadow="never">
      <template #header>
        <span>自然语言匹配</span>
      </template>
      <el-input
        v-model="text"
        type="textarea"
        :rows="4"
        placeholder="例如：帮我找上海周末下午可以一起打网球的女生搭子，预算 150 以内"
      />
      <div class="actions">
        <el-button type="primary" :loading="loading" @click="onMatch">开始匹配</el-button>
        <el-button @click="reset">清空会话</el-button>
      </div>
      <p v-if="assistantMsg" class="assistant">{{ assistantMsg }}</p>
    </el-card>

    <el-card v-if="candidates.length" class="mt" shadow="never">
      <template #header>
        <div class="card-head">
          <span>Top {{ candidates.length }} 候选搭子</span>
          <el-button type="success" :loading="confirming" @click="onConfirm">确认并生成小程序链接</el-button>
        </div>
      </template>
      <el-table :data="candidates" border>
        <el-table-column type="index" label="#" width="50" />
        <el-table-column label="昵称" min-width="120">
          <template #default="{ row }">{{ row.partner?.nickname || row.partner_id }}</template>
        </el-table-column>
        <el-table-column label="城市" width="100">
          <template #default="{ row }">{{ row.partner?.city }}</template>
        </el-table-column>
        <el-table-column label="时价(分)" width="100">
          <template #default="{ row }">{{ row.partner?.hourly_price_fen }}</template>
        </el-table-column>
        <el-table-column prop="score" label="分数" width="80" />
        <el-table-column prop="reason" label="推荐理由" min-width="220" />
      </el-table>
    </el-card>

    <el-card v-if="confirmResult" class="mt" shadow="never">
      <template #header>小程序链接</template>
      <el-input v-model="confirmResult.mini_program_url" readonly>
        <template #append>
          <el-button @click="copyUrl">复制</el-button>
        </template>
      </el-input>
      <p class="meta">SID: {{ confirmResult.sid }} · 过期: {{ formatTime(confirmResult.expires_at) }}</p>
    </el-card>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { matchChat, matchConfirm } from '../api'

const text = ref('')
const loading = ref(false)
const confirming = ref(false)
const matchRequestId = ref(null)
const candidates = ref([])
const assistantMsg = ref('')
const confirmResult = ref(null)

async function onMatch() {
  if (!text.value.trim()) {
    ElMessage.warning('请输入匹配需求')
    return
  }
  loading.value = true
  confirmResult.value = null
  try {
    const data = await matchChat(text.value.trim(), matchRequestId.value)
    matchRequestId.value = data.match_request_id
    candidates.value = (data.candidates || []).slice(0, 5)
    assistantMsg.value = data.assistant_message || ''
    text.value = ''
  } finally {
    loading.value = false
  }
}

async function onConfirm() {
  if (!matchRequestId.value) return
  confirming.value = true
  try {
    const ids = candidates.value.map((c) => c.partner_id)
    confirmResult.value = await matchConfirm(matchRequestId.value, ids)
    ElMessage.success('已生成浏览链接')
  } finally {
    confirming.value = false
  }
}

function reset() {
  matchRequestId.value = null
  candidates.value = []
  assistantMsg.value = ''
  confirmResult.value = null
  text.value = ''
}

async function copyUrl() {
  const url = confirmResult.value?.mini_program_url
  if (!url) return
  try {
    await navigator.clipboard.writeText(url)
    ElMessage.success('已复制')
  } catch {
    ElMessage.info(url)
  }
}

function formatTime(v) {
  if (!v) return '-'
  return new Date(v).toLocaleString()
}
</script>

<style scoped>
.actions {
  margin-top: 12px;
  display: flex;
  gap: 8px;
}
.assistant {
  margin: 16px 0 0;
  color: #606266;
  line-height: 1.6;
}
.mt {
  margin-top: 16px;
}
.card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.meta {
  margin: 12px 0 0;
  color: #909399;
  font-size: 13px;
}
</style>
