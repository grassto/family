<template>
  <div>
    <div class="page-header">
      <h2>家族管理</h2>
      <button class="btn btn-primary" @click="showModal = true">+ 新建家族</button>
    </div>

    <div v-if="loading" class="empty">加载中...</div>
    <div v-else-if="families.length === 0" class="empty">还没有家族，点击上方按钮创建一个</div>

    <div class="grid" v-else>
      <div class="card family-card" v-for="f in families" :key="f.id">
        <div class="family-card-header">
          <h3 @click="$router.push(`/families/${f.id}`)">{{ f.name }}</h3>
          <div class="family-card-actions">
            <button class="btn btn-sm" @click="edit(f)">编辑</button>
            <button class="btn btn-sm btn-danger" @click="remove(f)">删除</button>
          </div>
        </div>
        <p class="desc" v-if="f.description">{{ f.description }}</p>
        <div class="meta">
          <span class="tag" :class="f.webhook_key ? 'tag-alive' : ''">
            {{ f.webhook_key ? '✅ 已配置提醒' : '⚠️ 未配置提醒' }}
          </span>
        </div>
      </div>
    </div>

    <!-- 新建/编辑弹窗 -->
    <div class="modal-overlay" v-if="showModal" @click.self="closeModal">
      <div class="modal">
        <h3>{{ editingId ? '编辑家族' : '新建家族' }}</h3>
        <div class="form-group">
          <label>家族名称 *</label>
          <input v-model="form.name" placeholder="如：张氏家族" />
        </div>
        <div class="form-group">
          <label>简介</label>
          <textarea v-model="form.description" rows="3" placeholder="家族简介/堂号/起源地"></textarea>
        </div>
        <div class="form-group">
          <label>企业微信 Webhook Key</label>
          <input v-model="form.webhook_key" placeholder="机器人 webhook 地址中的 key 参数" />
        </div>
        <div class="modal-actions">
          <button class="btn" @click="closeModal">取消</button>
          <button class="btn btn-primary" @click="submit">{{ editingId ? '保存' : '创建' }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { familyApi } from '../api'

const families = ref([])
const loading = ref(true)
const showModal = ref(false)
const editingId = ref(null)
const form = ref({ name: '', description: '', webhook_key: '' })

const load = async () => {
  loading.value = true
  try {
    const { data } = await familyApi.list()
    families.value = data
  } finally {
    loading.value = false
  }
}

const edit = (f) => {
  editingId.value = f.id
  form.value = { name: f.name, description: f.description, webhook_key: f.webhook_key }
  showModal.value = true
}

const closeModal = () => {
  showModal.value = false
  editingId.value = null
  form.value = { name: '', description: '', webhook_key: '' }
}

const submit = async () => {
  if (!form.value.name.trim()) return alert('请输入家族名称')
  if (editingId.value) {
    await familyApi.update(editingId.value, form.value)
  } else {
    await familyApi.create(form.value)
  }
  closeModal()
  load()
}

const remove = async (f) => {
  if (!confirm(`确定删除「${f.name}」？所有成员和关系也会被删除。`)) return
  await familyApi.remove(f.id)
  load()
}

onMounted(load)
</script>

<style scoped>
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}

.page-header h2 {
  font-size: 22px;
}

.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: 16px;
}

.family-card {
  transition: transform 0.2s, box-shadow 0.2s;
}

.family-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
}

.family-card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 8px;
}

.family-card-header h3 {
  cursor: pointer;
  color: #333;
  font-size: 18px;
}

.family-card-header h3:hover {
  color: #667eea;
}

.family-card-actions {
  display: flex;
  gap: 6px;
}

.desc {
  color: #888;
  font-size: 13px;
  margin-bottom: 12px;
  line-height: 1.5;
}

.meta {
  display: flex;
  gap: 8px;
}

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }

  .grid {
    grid-template-columns: 1fr;
  }
}
</style>
