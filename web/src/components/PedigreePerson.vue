<template>
  <router-link :to="'/persons/' + person.id" class="ped-person">
    <div class="ped-person-main">
      <div class="ped-sym" :class="symClass" />
      <div class="ped-text">
        <div class="ped-name">{{ person.name }}</div>
        <div v-for="(line, i) in extraLines" :key="i" class="ped-line">{{ line }}</div>
      </div>
    </div>
  </router-link>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  person: { type: Object, required: true },
})

const birthYearText = computed(() => {
  const b = props.person.birth_date || props.person.birthday
  if (!b || b.length < 4) return ''
  const y = parseInt(b.slice(0, 4), 10)
  if (!y) return ''
  return `${y}年出生`
})

const extraLines = computed(() => {
  const lines = []
  if (birthYearText.value) lines.push(birthYearText.value)
  if (props.person.is_alive === false) lines.push('已故')
  if (props.person.notes) lines.push(props.person.notes)
  return lines
})

const symClass = computed(() => {
  const g = props.person.gender
  if (g === 'male') return 'sym-male'
  if (g === 'female') return 'sym-female'
  return 'sym-unknown'
})
</script>

<style scoped>
.ped-person {
  display: block;
  text-decoration: none;
  color: inherit;
}

.ped-person:hover {
  opacity: 0.92;
}

.ped-person-main {
  display: flex;
  flex-direction: row;
  align-items: flex-start;
  gap: 10px;
}

.ped-sym {
  width: 28px;
  height: 28px;
  flex-shrink: 0;
  box-sizing: border-box;
  margin-top: 2px;
}

.sym-male {
  border-radius: 2px;
  background: #cfe8ff;
  border: 2px solid #2b6cb0;
}

.sym-female {
  border-radius: 50%;
  background: #ffe0ec;
  border: 2px solid #c5306b;
}

.sym-unknown {
  border-radius: 4px;
  background: #e8e8e8;
  border: 2px solid #718096;
}

.ped-text {
  max-width: 200px;
}

.ped-name {
  font-size: 14px;
  font-weight: 600;
  color: #1a1a1a;
  line-height: 1.35;
}

.ped-line {
  font-size: 12px;
  color: #4a5568;
  line-height: 1.4;
  margin-top: 2px;
}

@media (max-width: 768px) {
  .ped-text {
    max-width: 140px;
  }

  .ped-name {
    font-size: 13px;
  }

  .ped-line {
    font-size: 11px;
  }
}
</style>
