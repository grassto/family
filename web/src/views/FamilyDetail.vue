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

    <!-- 视图切换 -->
    <div class="view-tabs" v-if="persons.length > 0">
      <button class="tab" :class="{ active: viewMode === 'list' }" @click="viewMode = 'list'">
        📋 成员列表
      </button>
      <button class="tab" :class="{ active: viewMode === 'tree' }" @click="viewMode = 'tree'">
        🌳 族谱视图
      </button>
      <button class="tab" :class="{ active: viewMode === 'pedigree' }" @click="viewMode = 'pedigree'">
        系谱图
      </button>
    </div>

    <div v-if="loading" class="empty">加载中...</div>
    <div v-else-if="persons.length === 0" class="empty-state">
      <div class="empty-icon">👤</div>
      <p>这个家族还没有成员</p>
      <button class="btn btn-primary" @click="openAddPerson">+ 添加第一个成员</button>
    </div>

    <!-- 族谱视图 -->
    <div class="card tree-card" v-if="viewMode === 'tree' && persons.length > 0">
      <FamilyTree :persons="persons" :relations="allRelations" />
    </div>

    <!-- 系谱图 -->
    <div class="card tree-card" v-if="viewMode === 'pedigree' && persons.length > 0">
      <FamilyPedigree :persons="persons" :relations="allRelations" />
    </div>

    <!-- 列表视图 -->
    <template v-if="viewMode === 'list' && persons.length > 0">
      <!-- 桌面端表格 -->
      <div class="card desktop-table">
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
      <div class="mobile-cards">
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
    </template>

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
          <label>生日类型</label>
          <div class="birthday-type-toggle">
            <button type="button" class="type-btn" :class="{ active: personForm.birthday_type === 'solar' }" @click="personForm.birthday_type = 'solar'">☀️ 公历</button>
            <button type="button" class="type-btn" :class="{ active: personForm.birthday_type === 'lunar' }" @click="personForm.birthday_type = 'lunar'">🌙 农历</button>
          </div>
        </div>
        <div class="form-group" v-if="personForm.birthday_type === 'solar'">
          <label>公历生日</label>
          <input v-model="personForm.birthday" type="date" />
        </div>
        <div class="form-row" v-if="personForm.birthday_type === 'lunar'">
          <div class="form-group">
            <label>农历月</label>
            <select v-model="lunarMonth">
              <option value="">请选择</option>
              <option v-for="m in 12" :key="m" :value="m">{{ lunarMonthLabel(m) }}</option>
            </select>
          </div>
          <div class="form-group">
            <label>农历日</label>
            <select v-model="lunarDay">
              <option value="">请选择</option>
              <option v-for="d in 30" :key="d" :value="d">{{ lunarDayLabel(d) }}</option>
            </select>
          </div>
        </div>
        <div class="form-group" v-if="personForm.birthday_type === 'lunar' && personForm.birthday">
          <label>出生年份（公历）</label>
          <input v-model="birthYear" type="number" min="1900" max="2100" placeholder="如：1990" />
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
import { familyApi, personApi, relationApi } from '../api'
import FamilyTree from '../components/FamilyTree.vue'
import FamilyPedigree from '../components/FamilyPedigree.vue'

const route = useRoute()
const familyId = route.params.id

const family = ref(null)
const persons = ref([])
const allRelations = ref([])
const loading = ref(true)
const keyword = ref('')
const showPersonModal = ref(false)
const editingPersonId = ref(null)
const viewMode = ref('list')

const personForm = ref({
  name: '', gender: 'unknown', birthday: '', birthday_type: 'solar', generation: null,
  phone: '', address: '', notes: '', is_alive: true,
})
const lunarMonth = ref('')
const lunarDay = ref('')
const birthYear = ref('')

const lunarMonthNames = ['正', '二', '三', '四', '五', '六', '七', '八', '九', '十', '冬', '腊']
const lunarDayNames = [
  '初一', '初二', '初三', '初四', '初五', '初六', '初七', '初八', '初九', '初十',
  '十一', '十二', '十三', '十四', '十五', '十六', '十七', '十八', '十九', '二十',
  '廿一', '廿二', '廿三', '廿四', '廿五', '廿六', '廿七', '廿八', '廿九', '三十',
]
const lunarMonthLabel = (m) => lunarMonthNames[m - 1] + '月'
const lunarDayLabel = (d) => lunarDayNames[d - 1]

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

const loadRelations = async () => {
  const { data } = await relationApi.getByFamily(familyId)
  allRelations.value = data
}

const load = async () => {
  loading.value = true
  try {
    await Promise.all([loadFamily(), loadPersons(), loadRelations()])
  } finally {
    loading.value = false
  }
}

const openAddPerson = () => {
  editingPersonId.value = null
  personForm.value = {
    name: '', gender: 'unknown', birthday: '', birthday_type: 'solar', generation: null,
    phone: '', address: '', notes: '', is_alive: true,
  }
  lunarMonth.value = ''
  lunarDay.value = ''
  birthYear.value = ''
  showPersonModal.value = true
}

const editPerson = (p) => {
  editingPersonId.value = p.id
  personForm.value = {
    name: p.name, gender: p.gender, birthday: p.birthday,
    birthday_type: p.birthday_type || 'solar',
    generation: p.generation, phone: p.phone, address: p.address, notes: p.notes,
    is_alive: p.is_alive,
  }
  if (p.birthday_type === 'lunar' && p.birthday) {
    const parts = p.birthday.split('-')
    birthYear.value = parts[0]
    lunarMonth.value = parseInt(parts[1])
    lunarDay.value = parseInt(parts[2])
  } else {
    lunarMonth.value = ''
    lunarDay.value = ''
    birthYear.value = ''
  }
  showPersonModal.value = true
}

const closePersonModal = () => {
  showPersonModal.value = false
  editingPersonId.value = null
}

const submitPerson = async () => {
  if (!personForm.value.name.trim()) return alert('请输入姓名')
  const data = { ...personForm.value }
  if (data.birthday_type === 'lunar') {
    if (!lunarMonth.value || !lunarDay.value) return alert('请选择农历月和日')
    if (!birthYear.value) return alert('请输入出生年份')
    data.birthday = `${birthYear.value}-${String(lunarMonth.value).padStart(2, '0')}-${String(lunarDay.value).padStart(2, '0')}`
  }
  if (editingPersonId.value) {
    await personApi.update(editingPersonId.value, data)
  } else {
    await personApi.create({ ...data, family_id: Number(familyId) })
  }
  closePersonModal()
  loadPersons()
  loadRelations()
}

const removePerson = async (p) => {
  if (!confirm(`确定删除「${p.name}」？`)) return
  await personApi.remove(p.id)
  load()
}

onMounted(load)
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 16px;
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

/* 视图切换 */
.view-tabs {
  display: flex;
  gap: 4px;
  margin-bottom: 16px;
  background: white;
  border-radius: 10px;
  padding: 4px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.tab {
  flex: 1;
  padding: 10px 16px;
  border: none;
  background: transparent;
  border-radius: 8px;
  font-size: 14px;
  cursor: pointer;
  color: #666;
  transition: all 0.2s;
  font-weight: 500;
}

.tab:hover {
  background: #f5f7fa;
}

.tab.active {
  background: linear-gradient(135deg, #667eea, #764ba2);
  color: white;
}

/* 族谱卡片 */
.tree-card {
  overflow: visible;
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

.birthday-type-toggle {
  display: flex;
  gap: 8px;
}

.type-btn {
  flex: 1;
  padding: 8px 12px;
  border: 2px solid #dcdfe6;
  border-radius: 8px;
  background: white;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.type-btn.active {
  border-color: #667eea;
  background: #f0f2ff;
  color: #667eea;
  font-weight: 600;
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
