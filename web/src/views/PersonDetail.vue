<template>
  <div>
    <div class="page-header">
      <router-link :to="`/families/${person?.family_id}`" class="back">← 返回家族</router-link>
    </div>

    <div v-if="loading" class="empty">加载中...</div>
    <div v-else-if="!person" class="empty">成员不存在</div>

    <template v-else>
      <!-- 基本信息 -->
      <div class="card info-card">
        <div class="info-header">
          <div class="avatar" :class="person.gender">{{ person.name[0] }}</div>
          <div class="info-main">
            <h2>{{ person.name }}</h2>
            <div class="info-tags">
              <span class="tag" :class="genderClass(person.gender)">{{ genderLabel(person.gender) }}</span>
              <span class="tag" :class="person.is_alive ? 'tag-alive' : 'tag-dead'">
                {{ person.is_alive ? '在世' : '已故' }}
              </span>
              <span class="tag" v-if="person.generation">第{{ person.generation }}代</span>
            </div>
          </div>
          <button class="btn" @click="showEditModal = true">编辑</button>
        </div>

        <div class="info-grid">
          <div class="info-item" v-if="person.birthday">
            <span class="label">🎂 生日（提醒）</span>
            <span>{{ person.birthday }}（{{ person.birthday_type === 'lunar' ? '农历' : '公历' }}）</span>
          </div>
          <div class="info-item" v-if="person.birth_date">
            <span class="label">🍼 出生日期（公历）</span>
            <span>{{ person.birth_date }} {{ ageText }}</span>
          </div>
          <div class="info-item" v-if="person.death_date">
            <span class="label">🕯️ 死亡日期（公历）</span>
            <span>{{ person.death_date }}</span>
          </div>
          <div class="info-item" v-if="person.phone">
            <span class="label">📱 电话</span>
            <span>{{ person.phone }}</span>
          </div>
          <div class="info-item" v-if="person.address">
            <span class="label">📍 地址</span>
            <span>{{ person.address }}</span>
          </div>
          <div class="info-item" v-if="person.notes">
            <span class="label">📝 备注</span>
            <span>{{ person.notes }}</span>
          </div>
        </div>
      </div>

      <!-- 关系 -->
      <div class="card" style="margin-top: 16px">
        <div class="section-header">
          <h3>家庭关系</h3>
          <button class="btn btn-primary btn-sm" @click="showRelationModal = true">+ 添加关系</button>
        </div>

        <div v-if="relations.length === 0" class="empty">暂无关系记录</div>
        <div v-else class="relation-list">
          <div class="relation-item" v-for="r in relations" :key="r.relation_id + '-' + r.type + '-' + r.person_id">
            <router-link :to="`/persons/${r.person_id}`" class="relation-name">{{ r.person_name }}</router-link>
            <span class="relation-type" :class="{ 'relation-derived': r.derived }">{{ r.type_label }}{{ r.derived ? ' (自动)' : '' }}</span>
            <template v-if="!r.derived">
              <button class="btn btn-sm" @click="editRelation(r)">编辑</button>
              <button class="btn btn-sm btn-danger" @click="removeRelation(r.relation_id)">解除</button>
            </template>
          </div>
        </div>
      </div>
    </template>

    <!-- 编辑弹窗 -->
    <div class="modal-overlay" v-if="showEditModal" @click.self="showEditModal = false">
      <div class="modal">
        <h3>编辑成员</h3>
        <div class="form-group">
          <label>姓名</label>
          <input v-model="editForm.name" />
        </div>
        <div class="form-row">
          <div class="form-group">
            <label>性别</label>
            <select v-model="editForm.gender">
              <option value="unknown">未知</option>
              <option value="male">男</option>
              <option value="female">女</option>
            </select>
          </div>
          <div class="form-group">
            <label>辈分</label>
            <input v-model.number="editForm.generation" type="number" />
          </div>
        </div>
        <div class="form-row">
          <div class="form-group">
            <label>生日类型</label>
            <select v-model="editForm.birthday_type">
              <option value="solar">公历</option>
              <option value="lunar">农历</option>
            </select>
          </div>
          <div class="form-group">
            <label>生日（提醒用）</label>
            <input
              v-if="editForm.birthday_type === 'solar'"
              v-model="editForm.birthday"
              type="date"
            />
            <input
              v-else
              v-model="editForm.birthday"
              placeholder="如 08-15（农历月-日）"
            />
          </div>
        </div>
        <div class="form-row">
          <div class="form-group">
            <label>出生日期（公历）</label>
            <input v-model="editForm.birth_date" type="date" />
          </div>
          <div class="form-group">
            <label>死亡日期（公历）</label>
            <input v-model="editForm.death_date" type="date" />
          </div>
        </div>
        <div class="form-group">
          <label>电话</label>
          <input v-model="editForm.phone" />
        </div>
        <div class="form-group">
          <label>地址</label>
          <input v-model="editForm.address" />
        </div>
        <div class="form-group">
          <label>备注</label>
          <textarea v-model="editForm.notes" rows="2"></textarea>
        </div>
        <div class="form-group">
          <label>
            <input type="checkbox" v-model="editForm.is_alive" /> 在世
          </label>
        </div>
        <div class="modal-actions">
          <button class="btn" @click="showEditModal = false">取消</button>
          <button class="btn btn-primary" @click="saveEdit">保存</button>
        </div>
      </div>
    </div>

    <!-- 添加关系弹窗 -->
    <div class="modal-overlay" v-if="showRelationModal" @click.self="showRelationModal = false">
      <div class="modal">
        <h3>添加关系</h3>
        <div class="form-group">
          <label>关系人</label>
          <select v-model="relationForm.related_id">
            <option value="">请选择...</option>
            <option v-for="p in otherPersons" :key="p.id" :value="p.id">{{ p.name }}</option>
          </select>
        </div>
        <div class="form-group">
          <label>关系类型</label>
          <select v-model="relationForm.type">
            <option v-for="t in storedRelationTypes" :key="t.value" :value="t.value">{{ t.label }}</option>
          </select>
        </div>
        <p class="form-hint">💡 兄弟姐妹、祖孙、姻亲等关系会根据父母/配偶关系自动推导</p>
        <div class="modal-actions">
          <button class="btn" @click="showRelationModal = false">取消</button>
          <button class="btn btn-primary" @click="addRelation">添加</button>
        </div>
      </div>
    </div>

    <!-- 编辑关系弹窗 -->
    <div class="modal-overlay" v-if="showEditRelationModal" @click.self="showEditRelationModal = false">
      <div class="modal">
        <h3>编辑关系</h3>
        <div class="form-group">
          <label>关系人</label>
          <select v-model="editRelationForm.related_id">
            <option value="">请选择...</option>
            <option v-for="p in otherPersons" :key="p.id" :value="p.id">{{ p.name }}</option>
          </select>
        </div>
        <div class="form-group">
          <label>关系类型</label>
          <select v-model="editRelationForm.type">
            <option v-for="t in storedRelationTypes" :key="t.value" :value="t.value">{{ t.label }}</option>
          </select>
        </div>
        <div class="modal-actions">
          <button class="btn" @click="showEditRelationModal = false">取消</button>
          <button class="btn btn-primary" @click="saveEditRelation">保存</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { personApi, relationApi } from '../api'

const route = useRoute()
const personId = computed(() => route.params.id)

const person = ref(null)
const relations = ref([])
const allPersons = ref([])
const relationTypes = ref([])
const loading = ref(true)
const showEditModal = ref(false)
const showRelationModal = ref(false)
const showEditRelationModal = ref(false)
const editingRelationId = ref(null)

const editForm = ref({})
const relationForm = ref({ related_id: '', type: 'parent' })
const editRelationForm = ref({ related_id: '', type: 'parent' })

const genderLabel = (g) => ({ male: '男', female: '女', unknown: '未知' }[g] || '未知')
const genderClass = (g) => ({ male: 'tag-male', female: 'tag-female' }[g] || '')

const ageText = computed(() => {
  const source = person.value?.birth_date || person.value?.birthday
  if (!source) return ''
  const year = parseInt(source.substring(0, 4))
  if (!year) return ''
  const age = new Date().getFullYear() - year
  return `（${age}岁）`
})

const storedRelationTypes = computed(() =>
  relationTypes.value.filter(t => t.stored !== false)
)

const otherPersons = computed(() =>
  allPersons.value.filter(p => p.id !== Number(personId.value))
)

const load = async () => {
  loading.value = true
  try {
    const [pRes, rRes, tRes] = await Promise.all([
      personApi.get(personId.value),
      relationApi.getByPerson(personId.value),
      relationApi.types(),
    ])
    person.value = pRes.data
    editForm.value = { ...pRes.data }
    relations.value = rRes.data
    relationTypes.value = tRes.data

    if (pRes.data.family_id) {
      const { data } = await personApi.list(pRes.data.family_id)
      allPersons.value = data
    }
  } finally {
    loading.value = false
  }
}

const saveEdit = async () => {
  const payload = { ...editForm.value }
  if (payload.is_alive) {
    payload.death_date = ''
  }
  await personApi.update(personId.value, payload)
  showEditModal.value = false
  load()
}

const addRelation = async () => {
  if (!relationForm.value.related_id) return alert('请选择关系人')
  await relationApi.create({
    person_id: Number(personId.value),
    related_id: Number(relationForm.value.related_id),
    type: relationForm.value.type,
  })
  showRelationModal.value = false
  relationForm.value = { related_id: '', type: 'parent' }
  load()
}

const removeRelation = async (id) => {
  if (!confirm('确定解除此关系？')) return
  await relationApi.remove(id)
  load()
}

const editRelation = (r) => {
  editingRelationId.value = r.relation_id
  editRelationForm.value = {
    related_id: r.person_id,
    type: r.type,
  }
  showEditRelationModal.value = true
}

const saveEditRelation = async () => {
  if (!editRelationForm.value.related_id) return alert('请选择关系人')
  await relationApi.update(editingRelationId.value, {
    person_id: Number(personId.value),
    related_id: Number(editRelationForm.value.related_id),
    type: editRelationForm.value.type,
  })
  showEditRelationModal.value = false
  editingRelationId.value = null
  load()
}

onMounted(load)

watch(personId, () => {
  showEditModal.value = false
  showRelationModal.value = false
  showEditRelationModal.value = false
  load()
})
</script>

<style scoped>
.page-header {
  margin-bottom: 16px;
}

.back {
  color: #667eea;
  text-decoration: none;
  font-size: 14px;
}

.info-card {
  padding: 28px;
}

.info-header {
  display: flex;
  align-items: center;
  gap: 20px;
  margin-bottom: 24px;
}

.avatar {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  font-weight: 600;
  color: white;
  background: #ccc;
}

.avatar.male {
  background: linear-gradient(135deg, #667eea, #764ba2);
}

.avatar.female {
  background: linear-gradient(135deg, #f093fb, #f5576c);
}

.info-main h2 {
  font-size: 22px;
  margin-bottom: 8px;
}

.info-tags {
  display: flex;
  gap: 8px;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 16px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.info-item .label {
  font-size: 12px;
  color: #999;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.relation-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.relation-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 0;
  border-bottom: 1px solid #f0f0f0;
}

.relation-name {
  color: #333;
  text-decoration: none;
  font-weight: 500;
  min-width: 80px;
}

.relation-name:hover {
  color: #667eea;
}

.relation-type {
  color: #667eea;
  font-size: 13px;
  background: #f0f2ff;
  padding: 2px 10px;
  border-radius: 12px;
}

.form-hint {
  font-size: 12px;
  color: #999;
  margin-top: -8px;
  margin-bottom: 16px;
}

.relation-derived {
  background: #f5f5f5;
  color: #999;
}

@media (max-width: 768px) {
  .info-header {
    flex-wrap: wrap;
  }

  .info-main {
    flex: 1;
    min-width: 0;
  }

  .info-header .btn {
    margin-left: auto;
  }

  .info-grid {
    grid-template-columns: 1fr;
  }

  .info-tags {
    flex-wrap: wrap;
  }

  .relation-item {
    flex-wrap: wrap;
    gap: 8px;
  }
}

</style>
