<template>
  <div>
    <div class="toolbar">
      <el-select v-model="filters.status" clearable placeholder="状态" style="width: 120px" @change="load">
        <el-option label="online" value="online" />
        <el-option label="draft" value="draft" />
        <el-option label="offline" value="offline" />
      </el-select>
      <el-input v-model="filters.city" placeholder="城市" clearable style="width: 140px" @keyup.enter="load" />
      <el-button @click="load">查询</el-button>
      <el-button type="primary" @click="openCreate">新建搭子</el-button>
    </div>

    <el-table :data="list" v-loading="loading" border>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="nickname" label="昵称" min-width="110" />
      <el-table-column prop="city" label="城市" width="90" />
      <el-table-column prop="gender" label="性别" width="80" />
      <el-table-column prop="hourly_price_fen" label="时价(分)" width="100" />
      <el-table-column prop="min_hours" label="最短时长" width="90" />
      <el-table-column prop="union_id" label="工会ID" width="90" />
      <el-table-column prop="status" label="状态" width="90" />
      <el-table-column label="标签" min-width="140">
        <template #default="{ row }">
          {{ (row.tags || []).map((t) => t.tag || t).join(', ') }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="100" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
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

    <el-dialog v-model="visible" :title="form.id ? '编辑搭子' : '新建搭子'" width="640px">
      <el-form :model="form" label-width="120px">
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="昵称" required>
              <el-input v-model="form.nickname" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="城市" required>
              <el-input v-model="form.city" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="性别">
              <el-select v-model="form.gender" style="width: 100%">
                <el-option label="female" value="female" />
                <el-option label="male" value="male" />
                <el-option label="unknown" value="unknown" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="状态">
              <el-select v-model="form.status" style="width: 100%">
                <el-option label="online" value="online" />
                <el-option label="draft" value="draft" />
                <el-option label="offline" value="offline" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="时价(分)">
              <el-input-number v-model="form.hourly_price_fen" :min="0" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="最短时长">
              <el-input-number v-model="form.min_hours" :min="0.5" :step="0.5" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="周末加价万分比">
              <el-input-number v-model="form.weekend_surcharge_rate" :min="0" :max="10000" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="夜间加价万分比">
              <el-input-number v-model="form.night_surcharge_rate" :min="0" :max="10000" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="工会ID">
              <el-input-number v-model="form.union_id" :min="0" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="电话">
              <el-input v-model="form.phone" placeholder="编辑时重新填写" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="亮点">
              <el-input v-model="form.highlight" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="简介">
              <el-input v-model="form.bio" type="textarea" :rows="2" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="标签">
              <el-input v-model="form.tagsText" placeholder="逗号分隔，如 网球,咖啡,徒步" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="图集URL">
              <el-input v-model="form.galleryText" type="textarea" :rows="2" placeholder="逗号分隔的图片 URL" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="头像URL">
              <el-input v-model="form.avatar_url" />
            </el-form-item>
          </el-col>
        </el-row>
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
import { listPartners, createPartner, updatePartner } from '../api'

const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const loading = ref(false)
const visible = ref(false)
const saving = ref(false)
const filters = reactive({ city: '', status: '' })

const emptyForm = () => ({
  id: null,
  nickname: '',
  city: '',
  gender: 'unknown',
  bio: '',
  highlight: '',
  avatar_url: '',
  phone: '',
  status: 'draft',
  hourly_price_fen: 10000,
  min_hours: 1,
  weekend_surcharge_rate: 0,
  night_surcharge_rate: 0,
  union_id: null,
  tagsText: '',
  galleryText: '',
})

const form = reactive(emptyForm())

function splitCsv(text) {
  return (text || '')
    .split(/[,，\n]/)
    .map((s) => s.trim())
    .filter(Boolean)
}

async function load() {
  loading.value = true
  try {
    const data = await listPartners({
      page: page.value,
      page_size: pageSize,
      city: filters.city || undefined,
      status: filters.status || undefined,
    })
    list.value = data?.list || []
    total.value = data?.total || 0
  } finally {
    loading.value = false
  }
}

function openCreate() {
  Object.assign(form, emptyForm())
  visible.value = true
}

function openEdit(row) {
  Object.assign(form, emptyForm(), {
    id: row.id,
    nickname: row.nickname,
    city: row.city,
    gender: row.gender || 'unknown',
    bio: row.bio || '',
    highlight: row.highlight || '',
    avatar_url: row.avatar_url || '',
    phone: '',
    status: row.status || 'draft',
    hourly_price_fen: row.hourly_price_fen,
    min_hours: row.min_hours,
    weekend_surcharge_rate: row.weekend_surcharge_rate || 0,
    night_surcharge_rate: row.night_surcharge_rate || 0,
    union_id: row.union_id || null,
    tagsText: (row.tags || []).map((t) => t.tag || t).join(', '),
    galleryText: (row.gallery || []).join(', '),
  })
  visible.value = true
}

async function save() {
  if (!form.nickname.trim() || !form.city.trim()) {
    ElMessage.warning('请填写昵称和城市')
    return
  }
  saving.value = true
  try {
    const payload = {
      nickname: form.nickname,
      city: form.city,
      gender: form.gender,
      bio: form.bio,
      highlight: form.highlight,
      avatar_url: form.avatar_url,
      phone: form.phone,
      status: form.status,
      hourly_price_fen: form.hourly_price_fen,
      min_hours: form.min_hours,
      weekend_surcharge_rate: form.weekend_surcharge_rate,
      night_surcharge_rate: form.night_surcharge_rate,
      union_id: form.union_id || null,
      tags: splitCsv(form.tagsText),
      gallery: splitCsv(form.galleryText),
    }
    if (form.id) {
      await updatePartner(form.id, payload)
    } else {
      await createPartner(payload)
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
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}
.pager {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
