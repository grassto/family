<template>
  <div>
    <div class="page-header">
      <h2>🎂 生日提醒</h2>
      <div class="actions">
        <select v-model="days" @change="load" class="days-select">
          <option :value="7">未来 7 天</option>
          <option :value="30">未来 30 天</option>
          <option :value="90">未来 90 天</option>
        </select>
      </div>
    </div>

    <div v-if="loading" class="empty">加载中...</div>
    <div v-else-if="persons.length === 0" class="empty">
      未来{{ days }}天内没有人过生日 🎉
    </div>

    <div class="birthday-list" v-else>
      <div class="card birthday-card" v-for="p in persons" :key="p.id">
        <div class="birthday-info">
          <div class="avatar" :class="p.gender">{{ p.name[0] }}</div>
          <div>
            <router-link :to="`/persons/${p.id}`" class="name">{{ p.name }}</router-link>
            <div class="meta">
              <span v-if="p.age">{{ p.age }}岁</span>
              <span v-if="p.birthday_type === 'lunar'" class="lunar-tag">🌙 农历{{ p.lunar_label }}</span>
              <span v-if="p.birthday_type !== 'lunar'" class="solar-tag">☀️ 公历</span>
            </div>
          </div>
        </div>
        <div class="birthday-date">
          <div class="date-badge" :class="{ today: p.is_today }">
            {{ p.is_today ? '🎉 今天' : formatDate(p.next_birthday) }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { birthdayApi } from '../api'

const persons = ref([])
const loading = ref(true)
const days = ref(30)

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const parts = dateStr.split('-')
  return `${parseInt(parts[1])}月${parseInt(parts[2])}日`
}

const load = async () => {
  loading.value = true
  try {
    const { data } = await birthdayApi.upcoming(days.value)
    persons.value = data
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.days-select {
  padding: 8px 14px;
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  font-size: 14px;
  background: white;
}

.birthday-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.birthday-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
}

.birthday-info {
  display: flex;
  align-items: center;
  gap: 14px;
}

.avatar {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
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

.name {
  font-weight: 500;
  color: #333;
  text-decoration: none;
}

.name:hover {
  color: #667eea;
}

.meta {
  font-size: 13px;
  color: #999;
  margin-top: 2px;
  display: flex;
  gap: 8px;
  align-items: center;
}

.lunar-tag {
  color: #722ed1;
  font-size: 12px;
}

.solar-tag {
  color: #faad14;
  font-size: 12px;
}

.date-badge {
  padding: 6px 14px;
  border-radius: 20px;
  font-size: 14px;
  font-weight: 500;
  background: #f0f2ff;
  color: #667eea;
}

.date-badge.today {
  background: linear-gradient(135deg, #667eea, #764ba2);
  color: white;
}

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }

  .birthday-card {
    padding: 14px 16px;
  }

  .avatar {
    width: 38px;
    height: 38px;
    font-size: 16px;
  }

  .birthday-info {
    gap: 10px;
  }
}
</style>
