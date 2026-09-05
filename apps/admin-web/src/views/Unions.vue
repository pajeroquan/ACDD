<template>
  <div>
    <div class="toolbar">
      <el-button type="primary" @click="openCreate">新建工会</el-button>
    </div>
    <el-table :data="list" v-loading="loading" border>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="name" label="名称" min-width="140" />
      <el-table-column label="分成(万分比)" width="120">
        <template #default="{ row }">{{ row.commission_rate }}</template>
      </el-table-column>
      <el-table-column prop="contact_name" label="联系人" width="120" />
      <el-table-column prop="settle_account" label="结算账户" min-width="140" />
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">
            {{ row.status === 1 ? '启用' : '停用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="100" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="visible" :title="form.id ? '编辑工会' : '新建工会'" width="480px">
      <el-form :model="form" label-width="110px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="分成万分比">
          <el-input-number v-model="form.commission_rate" :min="0" :max="10000" style="width: 100%" />
        </el-form-item>
        <el-form-item label="联系人">
          <el-input v-model="form.contact_name" />
        </el-form-item>
        <el-form-item label="联系电话">
          <el-input v-model="form.contact_phone" placeholder="编辑时重新填写" />
        </el-form-item>
        <el-form-item label="结算账户">
          <el-input v-model="form.settle_account" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="0">停用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="visible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { listUnions, createUnion, updateUnion } from '../api'

const list = ref([])
const loading = ref(false)
const visible = ref(false)
const saving = ref(false)
const form = reactive({
  id: null,
  name: '',
  commission_rate: 1000,
  contact_name: '',
  contact_phone: '',
  settle_account: '',
  status: 1,
})

async function load() {
  loading.value = true
  try {
    list.value = (await listUnions()) || []
  } finally {
    loading.value = false
  }
}

function resetForm() {
  Object.assign(form, {
    id: null,
    name: '',
    commission_rate: 1000,
    contact_name: '',
    contact_phone: '',
    settle_account: '',
    status: 1,
  })
}

function openCreate() {
  resetForm()
  visible.value = true
}

function openEdit(row) {
  Object.assign(form, {
    id: row.id,
    name: row.name,
    commission_rate: row.commission_rate,
    contact_name: row.contact_name || '',
    contact_phone: '',
    settle_account: row.settle_account || '',
    status: row.status,
  })
  visible.value = true
}

async function save() {
  if (!form.name.trim()) {
    ElMessage.warning('请填写名称')
    return
  }
  saving.value = true
  try {
    const payload = {
      name: form.name,
      commission_rate: form.commission_rate,
      contact_name: form.contact_name,
      contact_phone: form.contact_phone,
      settle_account: form.settle_account,
      status: form.status,
    }
    if (form.id) {
      await updateUnion(form.id, payload)
    } else {
      await createUnion(payload)
    }
    ElMessage.success('已保存')
    visible.value = false
    await load()
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.toolbar {
  margin-bottom: 12px;
}
</style>
