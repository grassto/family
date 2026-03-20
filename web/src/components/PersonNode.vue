<template>
  <router-link :to="`/persons/${person.id}`" class="person-node" :class="[person.gender, { dead: !person.is_alive }]">
    <div class="node-avatar" :class="person.gender">{{ person.name?.[0] || '?' }}</div>
    <div class="node-info">
      <span class="node-name">{{ person.name }}</span>
      <div class="node-meta">
        <span class="node-gen" v-if="person.generation">第{{ person.generation }}代</span>
        <span class="node-age" v-if="age">{{ age }}岁</span>
      </div>
    </div>
  </router-link>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  person: { type: Object, required: true },
})

const age = computed(() => {
  if (!props.person.birthday) return null
  const year = parseInt(props.person.birthday.substring(0, 4))
  if (!year) return null
  return new Date().getFullYear() - year
})
</script>

<style scoped>
.person-node {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 12px 16px;
  background: white;
  border: 2px solid #e8eaed;
  border-radius: 12px;
  text-decoration: none;
  transition: all 0.25s ease;
  min-width: 90px;
  cursor: pointer;
  position: relative;
}

.person-node:hover {
  transform: translateY(-3px);
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.1);
  border-color: #667eea;
}

.person-node.male {
  border-color: #d0e4ff;
  background: linear-gradient(135deg, #f8faff 0%, #ffffff 100%);
}

.person-node.male:hover {
  border-color: #1890ff;
  box-shadow: 0 6px 20px rgba(24, 144, 255, 0.15);
}

.person-node.female {
  border-color: #ffe0ec;
  background: linear-gradient(135deg, #fff8fb 0%, #ffffff 100%);
}

.person-node.female:hover {
  border-color: #eb2f96;
  box-shadow: 0 6px 20px rgba(235, 47, 150, 0.15);
}

.person-node.dead {
  opacity: 0.6;
}

.person-node.dead:hover {
  opacity: 0.85;
}

.node-avatar {
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
  flex-shrink: 0;
}

.node-avatar.male {
  background: linear-gradient(135deg, #667eea, #764ba2);
}

.node-avatar.female {
  background: linear-gradient(135deg, #f093fb, #f5576c);
}

.node-info {
  text-align: center;
}

.node-name {
  font-size: 14px;
  font-weight: 600;
  color: #333;
  display: block;
  line-height: 1.3;
}

.node-meta {
  display: flex;
  gap: 6px;
  justify-content: center;
  margin-top: 2px;
}

.node-gen, .node-age {
  font-size: 11px;
  color: #999;
}

@media (max-width: 768px) {
  .person-node {
    padding: 8px 10px;
    min-width: 72px;
    border-radius: 10px;
  }

  .node-avatar {
    width: 36px;
    height: 36px;
    font-size: 15px;
  }

  .node-name {
    font-size: 12px;
  }

  .node-gen, .node-age {
    font-size: 10px;
  }
}
</style>
