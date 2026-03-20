<template>
  <div>
    <div class="page-header">
      <div>
        <router-link to="/families" class="back">← 返回</router-link>
        <h2>{{ family?.name || '...' }}</h2>
        <p class="desc" v-if="family?.description">{{ family.description }}</p>
      </div>
      <div class="actions">
        <input class="search" v-model="keyword" placeholder="搜索成员..." @input="loadPersons" />
        <button class="btn btn-primary" @click="openAddPerson">+ 添加成员</button>
      </div>
    </div>

    <div v-if="loading" class="empty">加载中...</div>
    <div v-else-if="persons.length === 0" class="empty-state">
      <div class="empty-icon">👤</div>
      <p>这个家族还没有成员</p>
      <button class="btn btn-primary" @click="openAddPerson">+ 添加第一个成员</button>
    </div>

    <!-- 桌面端表格 -->
    <div class="card desktop-table" v-if="persons.length > 0">
      <table class="table">
        <thead>
          <tr>
            <th>姓名</th>
            <th>性别</th>
            <th>生日</th>
            <th>辈分</th>
            <th>状态</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in persons" :key="p.id">
            <td>
              <router-link :to="`/persons/${p.id}`" class="name-link">{{ p.name }}</router-link>
            </td>
            <td>
              <span class="tag" :class="genderClass(p.gender)">{{ genderLabel(p.gender) }}</span>
            </td>
            <td>{{ p.birthday || '-' }}</td>
            <td>{{ p.generation ? `第${p.generation}代` : '-' }}</td>
            <td>
              <span class="tag" :class="p.is_alive ? 'tag-alive' : 'tag-dead'">
                {{ p.is_alive ? '在世' : '已故' }}
              </span>
            </td>
            <td>
              <button class="btn btn-sm" @click="editPerson(p)">编辑</button>
              <button class="btn btn-sm btn-danger" @click="removePerson(p)">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 移动端卡片列表 -->
    <div class="mobile-cards" v-if="persons.length > 0">
      <div class="card person-card" v-for="p in persons" :key="'m'+p.id">
        <div class="person-card-main">
          <router-link :to="`/persons/${p.id}`" class="person-card-name">{{ p.name }}</router-link>
          <div class="person-card-tags">
            <span class="tag" :class="genderClass(p.gender)">{{ genderLabel(p.gender) }}</span>
            <span class="tag" :class="p.is_alive ? 'tag-alive' : 'tag-dead'">
              {{ p.is_alive ? '在世' : '已故' }}
            </span>
            <span class="tag" v-if="p.generation">第{{ p.generation }}代</span>
          </div>
          <div class="person-card-meta" v-if="p.birthday">🎂 {{ p.birthday }}</div>
        </div>
        <div class="person-card-actions">
          <button class="btn btn-sm" @click="editPerson(p)">编辑</button>
          <button class="btn btn-sm btn-danger" @click="removePerson(p)">删除</button>
        </div>
      </div>
    </div>

    <!-- 添加/编辑成员弹窗 -->
    <div class="modal-overlay" v-if="showPersonModal" @click.self="closePersonModal">
      <div class="modal">
        <h3>{{ editingPersonId ? '编辑成员' : '添加成员' }}</h3>
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
        <div class="form-group">
          <label>生日</label>
          <input v-model="personForm.birthday" type="date" />
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
          <button class="btn btn-primary" @click="submitPerson">{{ editingPersonId ? '保存' : '添加' }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { familyApi, personApi } from '../api'

const route = useRoute()
const familyId = route.params.id

const family = ref(null)
const persons = ref([])
const loading = ref(true)
const keyword = ref('')
const showPersonModal = ref(false)
const editingPersonId = ref(null)

const personForm = ref({
  name: '', gender: 'unknown', birthday: '', generation: null,
  phone: '', address: '', notes: '', is_alive: true,
})

const genderLabel = (g) => ({ male: '男', female: '女', unknown: '未知' }[g] || '未知')
const genderClass = (g) => ({ male: 'tag-male', female: 'tag-female' }[g] || '')

const loadFamily = async () => {
  const { data } = await familyApi.get(familyId)
  family.value = data
}

const loadPersons = async () => {
  const { data } = await personApi.list(familyId, keyword.value)
  persons.value = data
}

const load = async () => {
  loading.value = true
  try {
    await Promise.all([loadFamily(), loadPersons()])
  } finally {
    loading.value = false
  }
}

const openAddPerson = () => {
  editingPersonId.value = null
  personForm.value = {
    name: '', gender: 'unknown', birthday: '', generation: null,
    phone: '', address: '', notes: '', is_alive: true,
  }
  showPersonModal.value = true
}

const editPerson = (p) => {
  editingPersonId.value = p.id
  personForm.value = {
    name: p.name, gender: p.gender, birthday: p.birthday,
    generation: p.generation, phone: p.phone, address: p.address, notes: p.notes,
    is_alive: p.is_alive,
  }
  showPersonModal.value = true
}

const closePersonModal = () => {
  showPersonModal.value = false
  editingPersonId.value = null
}

const submitPerson = async () => {
  if (!personForm.value.name.trim()) return alert('请输入姓名')
  if (editingPersonId.value) {
    await personApi.update(editingPersonId.value, personForm.value)
  } else {
    await personApi.create({ ...personForm.value, family_id: Number(familyId) })
  }
  closePersonModal()
  loadPersons()
}

const removePerson = async (p) => {
  if (!confirm(`确定删除「${p.name}」？`)) return
  await personApi.remove(p.id)
  loadPersons()
}

onMounted(load)
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
}

.back {
  color: #667eea;
  text-decoration: none;
  font-size: 14px;
}

.desc {
  color: #888;
  font-size: 13px;
  margin-top: 4px;
}

.actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.search {
  padding: 8px 14px;
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  font-size: 14px;
  width: 200px;
}

.search:focus {
  outline: none;
  border-color: #667eea;
}

.table {
  width: 100%;
  border-collapse: collapse;
}

.table th,
.table td {
  padding: 12px 16px;
  text-align: left;
  border-bottom: 1px solid #f0f0f0;
}

.table th {
  color: #999;
  font-weight: 500;
  font-size: 13px;
}

.table tr:hover {
  background: #fafafa;
}

.name-link {
  color: #333;
  text-decoration: none;
  font-weight: 500;
}

.name-link:hover {
  color: #667eea;
}

/* 空状态 */
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
  margin-bottom: 20px;
}

/* 移动端卡片 - 默认隐藏 */
.mobile-cards {
  display: none;
  flex-direction: column;
  gap: 12px;
}

.person-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 16px;
}

.person-card-main {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.person-card-name {
  font-weight: 600;
  font-size: 16px;
  color: #333;
  text-decoration: none;
}

.person-card-name:hover {
  color: #667eea;
}

.person-card-tags {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.person-card-meta {
  font-size: 13px;
  color: #888;
}

.person-card-actions {
  display: flex;
  gap: 6px;
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

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }

  .actions {
    flex-wrap: wrap;
  }

  .search {
    flex: 1;
    width: auto;
    min-width: 0;
  }

  .desktop-table {
    display: none;
  }

  .mobile-cards {
    display: flex;
  }
}
</style>
