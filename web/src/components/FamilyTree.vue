<template>
  <div class="tree-container">
    <div v-if="persons.length === 0" class="empty">暂无成员数据</div>
    <div v-else class="tree-scroll">
      <div class="tree">
        <div class="generation-row" v-for="gen in generations" :key="gen.level">
          <div class="gen-label">第{{ gen.level }}代</div>
          <div class="gen-families">
            <div class="family-unit" v-for="(unit, idx) in gen.units" :key="idx">
              <!-- 夫妻对 / 单人 -->
              <div class="couple" v-if="unit.spousePair">
                <PersonNode :person="unit.spouseA" />
                <div class="spouse-connector">
                  <span class="spouse-line"></span>
                  <span class="spouse-label">配偶</span>
                  <span class="spouse-line"></span>
                </div>
                <PersonNode :person="unit.spouseB" />
              </div>
              <div class="single" v-else>
                <PersonNode :person="unit.person" />
              </div>
              <!-- 向下连接线到子代 -->
              <div class="down-connector" v-if="gen.level < maxGen && hasChildren(unit)">
                <div class="down-line"></div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    <p class="tree-hint" v-if="persons.length > 0">← 左右滑动查看完整族谱 →</p>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import PersonNode from './PersonNode.vue'

const props = defineProps({
  persons: { type: Array, default: () => [] },
  relations: { type: Array, default: () => [] },
})

// Build a map of person_id -> person
const personMap = computed(() => {
  const map = {}
  props.persons.forEach(p => { map[p.id] = p })
  return map
})

// Find spouse pairs
const spousePairs = computed(() => {
  const pairs = new Map()
  const visited = new Set()
  props.relations.forEach(r => {
    if (r.type === 'spouse') {
      const key = [Math.min(r.person_id, r.related_id), Math.max(r.person_id, r.related_id)].join('-')
      if (!visited.has(key)) {
        visited.add(key)
        pairs.set(key, {
          spouseA: personMap.value[r.person_id],
          spouseB: personMap.value[r.related_id],
        })
      }
    }
  })
  return pairs
})

// Check if a person is in any spouse pair
const personToSpousePair = computed(() => {
  const map = {}
  spousePairs.value.forEach((pair, key) => {
    if (pair.spouseA) map[pair.spouseA.id] = key
    if (pair.spouseB) map[pair.spouseB.id] = key
  })
  return map
})

// Parent -> children map
const parentChildren = computed(() => {
  const map = {}
  props.relations.forEach(r => {
    if (r.type === 'parent') {
      if (!map[r.person_id]) map[r.person_id] = new Set()
      map[r.person_id].add(r.related_id)
    }
    if (r.type === 'child') {
      if (!map[r.related_id]) map[r.related_id] = new Set()
      map[r.related_id].add(r.person_id)
    }
  })
  // Convert sets to arrays
  const result = {}
  Object.keys(map).forEach(k => { result[k] = [...map[k]] })
  return result
})

// Group by generation and build family units
const generations = computed(() => {
  const genMap = {}
  props.persons.forEach(p => {
    const gen = p.generation || 0
    if (!genMap[gen]) genMap[gen] = []
    genMap[gen].push(p)
  })

  const result = []
  const usedInUnit = new Set()

  Object.keys(genMap).sort((a, b) => Number(a) - Number(b)).forEach(genLevel => {
    const people = genMap[genLevel]
    const units = []

    people.forEach(p => {
      if (usedInUnit.has(p.id)) return

      const pairKey = personToSpousePair.value[p.id]
      if (pairKey && spousePairs.value[pairKey]) {
        const pair = spousePairs.value[pairKey]
        usedInUnit.add(pair.spouseA?.id)
        usedInUnit.add(pair.spouseB?.id)
        units.push({ spousePair: true, spouseA: pair.spouseA, spouseB: pair.spouseB })
      } else {
        usedInUnit.add(p.id)
        units.push({ spousePair: false, person: p })
      }
    })

    if (units.length > 0) {
      result.push({ level: Number(genLevel), units })
    }
  })

  return result
})

const maxGen = computed(() => {
  return Math.max(...props.persons.map(p => p.generation || 0), 0)
})

function hasChildren(unit) {
  if (unit.spousePair) {
    return parentChildren.value[unit.spouseA?.id]?.length > 0 ||
           parentChildren.value[unit.spouseB?.id]?.length > 0
  }
  return parentChildren.value[unit.person?.id]?.length > 0
}
</script>

<style scoped>
.tree-container {
  width: 100%;
}

.tree-scroll {
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
  padding: 12px 0 20px;
}

.tree {
  display: flex;
  flex-direction: column;
  gap: 0;
  min-width: fit-content;
}

.generation-row {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  padding: 12px 0;
  position: relative;
}

.gen-label {
  writing-mode: vertical-lr;
  text-orientation: mixed;
  font-size: 12px;
  color: #999;
  font-weight: 500;
  padding: 4px 6px;
  background: #f5f7fa;
  border-radius: 6px;
  flex-shrink: 0;
  letter-spacing: 2px;
}

.gen-families {
  display: flex;
  gap: 20px;
  flex-wrap: nowrap;
  align-items: flex-start;
}

.family-unit {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex-shrink: 0;
}

.couple {
  display: flex;
  align-items: center;
  gap: 0;
}

.spouse-connector {
  display: flex;
  align-items: center;
  gap: 0;
}

.spouse-line {
  display: block;
  width: 20px;
  height: 2px;
  background: #f5576c;
}

.spouse-label {
  font-size: 10px;
  color: #f5576c;
  padding: 0 4px;
  white-space: nowrap;
}

.single {
  display: flex;
}

.down-connector {
  display: flex;
  justify-content: center;
  width: 100%;
  padding: 4px 0;
}

.down-line {
  width: 2px;
  height: 24px;
  background: linear-gradient(to bottom, #dcdfe6, #667eea);
  border-radius: 1px;
}

.tree-hint {
  text-align: center;
  font-size: 12px;
  color: #ccc;
  margin-top: 4px;
}

@media (max-width: 768px) {
  .gen-families {
    gap: 14px;
  }

  .gen-label {
    font-size: 11px;
    padding: 3px 4px;
  }

  .tree-hint {
    display: block;
  }
}
</style>
