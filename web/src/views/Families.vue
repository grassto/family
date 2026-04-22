<template>
  <div>
    <div class="page-header">
      <h2>家族管理</h2>
      <div class="header-actions">
        <button class="btn" @click="triggerImport">导入家族</button>
        <button class="btn btn-primary" @click="showFamilyModal = true">+ 新建家族</button>
        <input ref="importInput" type="file" accept="application/json,.json" class="hidden-input" @change="handleImportFile" />
      </div>
    </div>

    <div v-if="loading" class="empty">加载中...</div>
    <div v-else-if="families.length === 0" class="empty-state">
      <div class="empty-icon">👨‍👩‍👧‍👦</div>
      <p>还没有家族，点击上方按钮创建一个</p>
    </div>

    <div class="grid" v-else>
      <div class="card family-card" v-for="f in families" :key="f.id">
        <div class="family-card-header">
          <h3 @click="$router.push(`/families/${f.id}`)">{{ f.name }}</h3>
          <div class="family-card-actions">
            <button class="btn btn-sm btn-export" @click="exportFamily(f)">导出</button>
            <button class="btn btn-sm" @click="editFamily(f)">编辑</button>
            <button class="btn btn-sm btn-danger" @click="removeFamily(f)">删除</button>
          </div>
        </div>
        <p class="desc" v-if="f.description">{{ f.description }}</p>
        <div class="family-card-footer">
          <div class="meta">
            <span class="member-count" @click="$router.push(`/families/${f.id}`)">
              👥 {{ f.member_count }} 位成员
            </span>
            <span class="tag" :class="f.webhook_key ? 'tag-alive' : ''">
              {{ f.webhook_key ? '✅ 已配置提醒' : '⚠️ 未配置提醒' }}
            </span>
          </div>
          <div class="card-footer-actions">
            <button class="btn btn-primary btn-sm" @click="openAddPerson(f)">+ 添加成员</button>
          </div>
        </div>
      </div>
    </div>

    <!-- 新建/编辑家族弹窗 -->
    <div class="modal-overlay" v-if="showFamilyModal" @click.self="closeFamilyModal">
      <div class="modal">
        <h3>{{ editingFamilyId ? '编辑家族' : '新建家族' }}</h3>
        <div class="form-group">
          <label>家族名称 *</label>
          <input v-model="familyForm.name" placeholder="如：张氏家族" />
        </div>
        <div class="form-group">
          <label>简介</label>
          <textarea v-model="familyForm.description" rows="3" placeholder="家族简介/堂号/起源地"></textarea>
        </div>
        <div class="form-group">
          <label>企业微信 Webhook Key</label>
          <input v-model="familyForm.webhook_key" placeholder="机器人 webhook 地址中的 key 参数" />
        </div>
        <div class="modal-actions">
          <button class="btn" @click="closeFamilyModal">取消</button>
          <button class="btn btn-primary" @click="submitFamily">{{ editingFamilyId ? '保存' : '创建' }}</button>
        </div>
      </div>
    </div>

    <!-- 快速添加成员弹窗 -->
    <div class="modal-overlay" v-if="showPersonModal" @click.self="closePersonModal">
      <div class="modal">
        <h3>向「{{ targetFamily?.name }}」添加成员</h3>
        <div class="form-group">
          <label>姓名 *</label>
          <input v-model="personForm.name" placeholder="成员姓名" />
        </div>
        <div class="form-row">
          <div class="form-group">
            <label>性别</label>
            <select v-model="personForm.gender">
              <option value="unknown">未知</option>
              <option value="male">男</option>
              <option value="female">女</option>
            </select>
          </div>
          <div class="form-group">
            <label>辈分</label>
            <input v-model.number="personForm.generation" type="number" min="1" placeholder="第几代" />
          </div>
        </div>
        <div class="form-row">
          <div class="form-group">
            <label>生日类型</label>
            <select v-model="personForm.birthday_type">
              <option value="solar">公历</option>
              <option value="lunar">农历</option>
            </select>
          </div>
          <div class="form-group">
            <label>生日（提醒用）</label>
            <input
              v-if="personForm.birthday_type === 'solar'"
              v-model="personForm.birthday"
              type="date"
            />
            <input
              v-else
              v-model="personForm.birthday"
              placeholder="如 08-15（农历月-日）"
            />
          </div>
        </div>
        <div class="form-row">
          <div class="form-group">
            <label>出生日期（公历）</label>
            <input v-model="personForm.birth_date" type="date" />
          </div>
          <div class="form-group">
            <label>死亡日期（公历）</label>
            <input v-model="personForm.death_date" type="date" />
          </div>
        </div>
        <div class="form-row">
          <div class="form-group">
            <label>电话</label>
            <input v-model="personForm.phone" placeholder="手机号码" />
          </div>
          <div class="form-group">
            <label>地址</label>
            <input v-model="personForm.address" placeholder="居住地址" />
          </div>
        </div>
        <div class="form-group">
          <label>备注</label>
          <textarea v-model="personForm.notes" rows="2" placeholder="其他备注信息"></textarea>
        </div>
        <div class="form-group">
          <label class="checkbox-label">
            <input type="checkbox" v-model="personForm.is_alive" />
            <span>在世</span>
          </label>
        </div>
        <div class="modal-actions">
          <button class="btn" @click="closePersonModal">取消</button>
          <button class="btn btn-primary" @click="submitPerson">添加</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { familyApi, personApi } from '../api'

const families = ref([])
const loading = ref(true)
const importInput = ref(null)

// 家族相关
const showFamilyModal = ref(false)
const editingFamilyId = ref(null)
const familyForm = ref({ name: '', description: '', webhook_key: '' })

// 成员相关
const showPersonModal = ref(false)
const targetFamily = ref(null)
const personForm = ref({
  name: '', gender: 'unknown', birthday: '', birthday_type: 'solar', birth_date: '', death_date: '', generation: null,
  phone: '', address: '', notes: '', is_alive: true,
})

const load = async () => {
  loading.value = true
  try {
    const { data } = await familyApi.list()
    families.value = data
  } finally {
    loading.value = false
  }
}

// --- 家族操作 ---
const editFamily = (f) => {
  editingFamilyId.value = f.id
  familyForm.value = { name: f.name, description: f.description, webhook_key: f.webhook_key }
  showFamilyModal.value = true
}

const closeFamilyModal = () => {
  showFamilyModal.value = false
  editingFamilyId.value = null
  familyForm.value = { name: '', description: '', webhook_key: '' }
}

const submitFamily = async () => {
  if (!familyForm.value.name.trim()) return alert('请输入家族名称')
  if (editingFamilyId.value) {
    await familyApi.update(editingFamilyId.value, familyForm.value)
  } else {
    await familyApi.create(familyForm.value)
  }
  closeFamilyModal()
  load()
}

const removeFamily = async (f) => {
  if (!confirm(`确定删除「${f.name}」？所有成员和关系也会被删除。`)) return
  await familyApi.remove(f.id)
  load()
}

const exportFamily = async (family) => {
  const { data } = await familyApi.exportOne(family.id)
  const json = JSON.stringify(data, null, 2)
  const blob = new Blob([json], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  const safeName = (family.name || `family-${family.id}`).replace(/[\\/:*?"<>|]/g, '_')
  link.href = url
  link.download = `${safeName}.json`
  document.body.appendChild(link)
  link.click()
  link.remove()
  URL.revokeObjectURL(url)
}

const triggerImport = () => {
  importInput.value?.click()
}

const handleImportFile = async (event) => {
  const file = event.target.files?.[0]
  event.target.value = ''
  if (!file) return
  try {
    const text = await file.text()
    const payload = JSON.parse(text)
    await familyApi.import(payload)
    alert('导入成功，已创建新家族')
    load()
  } catch (error) {
    alert(error?.response?.data?.error || '导入失败，请检查 JSON 文件')
  }
}

// --- 成员操作 ---
const openAddPerson = (f) => {
  targetFamily.value = f
  personForm.value = {
    name: '', gender: 'unknown', birthday: '', birthday_type: 'solar', birth_date: '', death_date: '', generation: null,
    phone: '', address: '', notes: '', is_alive: true,
  }
  showPersonModal.value = true
}

const closePersonModal = () => {
  showPersonModal.value = false
  targetFamily.value = null
}

const submitPerson = async () => {
  if (!personForm.value.name.trim()) return alert('请输入姓名')
  const payload = {
    ...personForm.value,
    family_id: targetFamily.value.id,
  }
  if (payload.is_alive) {
    payload.death_date = ''
  }
  await personApi.create(payload)
  closePersonModal()
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

.header-actions {
  display: flex;
  gap: 8px;
  align-items: center;
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

.btn-export {
  background: #eef3ff;
  color: #4c63d2;
  border: 1px solid #cdd8ff;
}

.btn-export:hover {
  background: #dfe8ff;
  border-color: #aebfff;
}

.desc {
  color: #888;
  font-size: 13px;
  margin-bottom: 12px;
  line-height: 1.5;
}

.family-card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-footer-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.meta {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}

.member-count {
  font-size: 13px;
  color: #667eea;
  cursor: pointer;
  font-weight: 500;
}

.member-count:hover {
  text-decoration: underline;
}

.empty-state {
  text-align: center;
  padding: 80px 20px;
  color: #999;
}

.empty-icon {
  font-size: 64px;
  margin-bottom: 16px;
}

.empty-state p {
  font-size: 15px;
}

.checkbox-label {
  display: flex !important;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.checkbox-label input[type="checkbox"] {
  width: auto;
  margin: 0;
}

.hidden-input {
  display: none;
}

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }

  .header-actions {
    width: 100%;
  }

  .grid {
    grid-template-columns: 1fr;
  }

  .family-card-footer {
    flex-direction: column;
    align-items: stretch;
    gap: 10px;
  }

  .family-card-footer .btn {
    width: 100%;
    justify-content: center;
  }
}
</style>
