<template>
  <div class="pedigree-root">
    <div v-if="persons.length === 0" class="empty">暂无成员数据</div>
    <template v-else>
      <div class="pedigree-scroll" ref="scrollRef" @scroll.passive="scheduleMeasure">
        <div class="pedigree-canvas" ref="canvasRef">
          <svg
            class="pedigree-svg"
            :width="svgSize.w"
            :height="svgSize.h"
            aria-hidden="true"
          >
            <path
              v-for="(seg, i) in lineSegments"
              :key="i"
              :d="seg.d"
              fill="none"
              stroke="#1a1a1a"
              stroke-width="1"
            />
          </svg>
          <div class="pedigree-rows">
            <div
              v-for="row in layoutRows"
              :key="row.gen"
              class="ped-row"
              :data-gen="row.gen"
            >
              <div
                v-for="(unit, ui) in row.units"
                :key="unit.key"
                class="ped-unit"
                :data-unit-key="unit.key"
              >
                <div class="ped-unit-inner">
                  <template v-if="unit.kind === 'couple'">
                    <div class="ped-marriage">
                      <div
                        class="ped-anchor ped-anchor-person"
                        :ref="(el) => setPersonEl(unit.members[0].id, el)"
                      >
                        <PedigreePerson :person="unit.members[0]" />
                      </div>
                      <div class="ped-marriage-line" aria-hidden="true" />
                      <div
                        class="ped-anchor ped-anchor-person"
                        :ref="(el) => setPersonEl(unit.members[1].id, el)"
                      >
                        <PedigreePerson :person="unit.members[1]" />
                      </div>
                    </div>
                  </template>
                  <template v-else-if="unit.kind === 'siblings'">
                    <div class="ped-sibling-row">
                      <div
                        v-for="p in unit.members"
                        :key="p.id"
                        class="ped-sibling-slot"
                      >
                        <div
                          class="ped-anchor ped-anchor-person"
                          :ref="(el) => setPersonEl(p.id, el)"
                        >
                          <PedigreePerson :person="p" />
                        </div>
                      </div>
                    </div>
                  </template>
                  <template v-else>
                    <div
                      class="ped-anchor ped-anchor-person"
                      :ref="(el) => setPersonEl(unit.members[0].id, el)"
                    >
                      <PedigreePerson :person="unit.members[0]" />
                    </div>
                  </template>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div class="pedigree-footer">
        <div class="pedigree-legend">
          <div class="legend-title">图例</div>
          <div class="legend-row">
            <span class="leg-sym male" />
            <span>男性</span>
          </div>
          <div class="legend-row">
            <span class="leg-sym female" />
            <span>女性</span>
          </div>
          <div class="legend-row">
            <span class="leg-sym unknown" />
            <span>未知</span>
          </div>
        </div>
        <p class="pedigree-hint">← 左右滑动查看 →</p>
      </div>
    </template>
  </div>
</template>

<script setup>
import {
  computed,
  ref,
  watch,
  onMounted,
  onBeforeUnmount,
  nextTick,
} from 'vue'
import PedigreePerson from './PedigreePerson.vue'

const props = defineProps({
  persons: { type: Array, default: () => [] },
  relations: { type: Array, default: () => [] },
})

/** @param {import('vue').Ref} el */
function setPersonEl(personId, el) {
  if (!personId) return
  if (el) personEls.set(personId, el)
  else personEls.delete(personId)
}

const scrollRef = ref(null)
const canvasRef = ref(null)
const personEls = new Map()
const svgSize = ref({ w: 0, h: 0 })
const lineSegments = ref([])

const personMap = computed(() => {
  const m = {}
  props.persons.forEach((p) => {
    m[p.id] = p
  })
  return m
})

function genOf(p) {
  return p.generation ?? 0
}

/** 所有人未填辈分时，用 parent + spouse 关系推断层级 */
function computeInferredGenerations(personIds, p2c, spouses) {
  const level = Object.fromEntries(personIds.map((id) => [id, 0]))
  const limit = personIds.length + 5
  for (let pass = 0; pass < limit; pass++) {
    let changed = false
    personIds.forEach((pid) => {
      const kids = p2c[pid] || []
      kids.forEach((cid) => {
        if (level[cid] === undefined) return
        const nv = level[pid] + 1
        if (level[cid] < nv) {
          level[cid] = nv
          changed = true
        }
      })
    })
    spouses.forEach(({ a, b }) => {
      if (level[a] === undefined || level[b] === undefined) return
      const L = Math.max(level[a], level[b])
      if (level[a] !== L || level[b] !== L) {
        level[a] = L
        level[b] = L
        changed = true
      }
    })
    if (!changed) break
  }
  return level
}

/** parent_id -> child_id[] */
const parentToChildren = computed(() => {
  const m = {}
  props.relations.forEach((r) => {
    if (r.type === 'parent') {
      if (!m[r.person_id]) m[r.person_id] = []
      m[r.person_id].push(r.related_id)
    } else if (r.type === 'child') {
      if (!m[r.related_id]) m[r.related_id] = []
      m[r.related_id].push(r.person_id)
    }
  })
  Object.keys(m).forEach((k) => {
    m[k] = [...new Set(m[k])]
  })
  return m
})

/** child_id -> parent_id[] */
const childToParents = computed(() => {
  const m = {}
  props.relations.forEach((r) => {
    if (r.type === 'parent') {
      const c = r.related_id
      if (!m[c]) m[c] = []
      m[c].push(r.person_id)
    } else if (r.type === 'child') {
      const c = r.person_id
      if (!m[c]) m[c] = []
      m[c].push(r.related_id)
    }
  })
  Object.keys(m).forEach((k) => {
    m[k] = [...new Set(m[k])]
  })
  return m
})

const spousePairs = computed(() => {
  const pairs = []
  const seen = new Set()
  props.relations.forEach((r) => {
    if (r.type !== 'spouse') return
    const a = r.person_id
    const b = r.related_id
    const key = a < b ? `${a}-${b}` : `${b}-${a}`
    if (seen.has(key)) return
    seen.add(key)
    pairs.push({ a, b, key })
  })
  return pairs
})

/** 在 computed 内调用，避免 watch / 声明顺序导致的引用问题 */
function buildEffectiveGenById() {
  const list = props.persons
  if (list.length === 0) return {}
  const allMissing = list.every(
    (p) => p.generation === null || p.generation === undefined,
  )
  if (!allMissing) {
    const o = {}
    list.forEach((p) => {
      o[p.id] = genOf(p)
    })
    return o
  }
  const ids = list.map((p) => p.id)
  return computeInferredGenerations(
    ids,
    parentToChildren.value,
    spousePairs.value,
  )
}

function childrenOfCouple(a, b) {
  const ca = parentToChildren.value[a] || []
  const cb = parentToChildren.value[b] || []
  const inter = ca.filter((id) => cb.includes(id))
  if (inter.length) return inter
  return [...new Set([...ca, ...cb])]
}

function unionFindMake(ids) {
  const p = Object.fromEntries(ids.map((id) => [id, id]))
  function find(x) {
    if (p[x] !== x) p[x] = find(p[x])
    return p[x]
  }
  function union(x, y) {
    const rx = find(x)
    const ry = find(y)
    if (rx !== ry) p[rx] = ry
  }
  return { find, union }
}

function buildUnitsForGen(gen, idsInGen) {
  const pmap = personMap.value
  const idSet = new Set(idsInGen)
  const used = new Set()
  const units = []

  const pairsHere = spousePairs.value.filter(
    (sp) => idSet.has(sp.a) && idSet.has(sp.b) && pmap[sp.a] && pmap[sp.b],
  )
  pairsHere.sort((x, y) => Math.min(x.a, x.b) - Math.min(y.a, y.b))
  pairsHere.forEach((sp) => {
    if (used.has(sp.a) || used.has(sp.b)) return
    used.add(sp.a)
    used.add(sp.b)
    units.push({
      key: `c-${gen}-${sp.key}`,
      kind: 'couple',
      members: [pmap[sp.a], pmap[sp.b]].sort((u, v) => u.id - v.id),
    })
  })

  const remaining = idsInGen.filter((id) => !used.has(id))
  if (remaining.length === 0) return units

  const uf = unionFindMake(remaining)
  props.relations.forEach((r) => {
    if (r.type !== 'sibling') return
    if (idSet.has(r.person_id) && idSet.has(r.related_id)) {
      uf.union(r.person_id, r.related_id)
    }
  })
  for (let i = 0; i < remaining.length; i++) {
    for (let j = i + 1; j < remaining.length; j++) {
      const a = remaining[i]
      const b = remaining[j]
      const pa = childToParents.value[a] || []
      const pb = childToParents.value[b] || []
      if (pa.some((x) => pb.includes(x))) uf.union(a, b)
    }
  }

  const clusters = new Map()
  remaining.forEach((id) => {
    const r = uf.find(id)
    if (!clusters.has(r)) clusters.set(r, [])
    clusters.get(r).push(id)
  })

  const clusterLists = [...clusters.values()].sort(
    (u, v) => Math.min(...u) - Math.min(...v),
  )
  clusterLists.forEach((ids) => {
    ids.sort((x, y) => x - y)
    const members = ids.map((id) => pmap[id]).filter(Boolean)
    if (members.length === 0) return
    if (members.length === 1) {
      units.push({
        key: `s-${gen}-${members[0].id}`,
        kind: 'single',
        members,
      })
    } else {
      units.push({
        key: `g-${gen}-${ids.join('-')}`,
        kind: 'siblings',
        members,
      })
    }
  })

  units.sort((A, B) => {
    const ma = Math.min(...A.members.map((m) => m.id))
    const mb = Math.min(...B.members.map((m) => m.id))
    return ma - mb
  })
  return units
}

/** gen -> { gen, units } */
const layoutRows = computed(() => {
  const pmap = personMap.value
  const effGen = buildEffectiveGenById()
  const byGen = {}
  props.persons.forEach((p) => {
    const g = effGen[p.id] ?? 0
    if (!byGen[g]) byGen[g] = []
    byGen[g].push(p.id)
  })
  const gens = Object.keys(byGen)
    .map(Number)
    .sort((a, b) => a - b)
  if (gens.length === 0) return []

  const rows = gens.map((g) => ({
    gen: g,
    units: buildUnitsForGen(g, byGen[g]),
  }))

  const unitIndexByGen = {}
  rows.forEach((row) => {
    const map = {}
    row.units.forEach((u, idx) => {
      u.members.forEach((m) => {
        map[m.id] = idx
      })
    })
    unitIndexByGen[row.gen] = map
  })

  function parentAnchorIndex(childId, parentGen) {
    const parents = childToParents.value[childId] || []
    const umap = unitIndexByGen[parentGen]
    if (!umap) return 9999
    const idxs = parents
      .map((pid) => umap[pid])
      .filter((x) => x !== undefined)
    if (idxs.length === 0) return 9999
    return Math.min(...idxs)
  }

  rows.forEach((row, ri) => {
    if (ri === 0) return
    const prevGen = rows[ri - 1].gen
    row.units.sort((A, B) => {
      const ka = Math.min(
        ...A.members.map((m) => parentAnchorIndex(m.id, prevGen)),
      )
      const kb = Math.min(
        ...B.members.map((m) => parentAnchorIndex(m.id, prevGen)),
      )
      if (ka !== kb) return ka - kb
      return (
        Math.min(...A.members.map((m) => m.id)) -
        Math.min(...B.members.map((m) => m.id))
      )
    })
  })

  return rows
})

/** @typedef {{ d: string }} Seg */

function measureLines() {
  const canvas = canvasRef.value
  if (!canvas) {
    lineSegments.value = []
    svgSize.value = { w: 0, h: 0 }
    return
  }
  const crect = canvas.getBoundingClientRect()
  svgSize.value = { w: Math.ceil(crect.width), h: Math.ceil(crect.height) }
  const ox = crect.left
  const oy = crect.top
  /** @type {Seg[]} */
  const segs = []

  function elCenterBottom(el) {
    const r = el.getBoundingClientRect()
    return { x: r.left - ox + r.width / 2, y: r.bottom - oy }
  }
  function elCenterTop(el) {
    const r = el.getBoundingClientRect()
    return { x: r.left - ox + r.width / 2, y: r.top - oy }
  }

  const rows = layoutRows.value
  const pmap = personMap.value
  const p2c = parentToChildren.value

  for (let ri = 0; ri < rows.length - 1; ri++) {
    const row = rows[ri]
    const nextRowIds = new Set()
    rows[ri + 1].units.forEach((u) => {
      u.members.forEach((m) => nextRowIds.add(m.id))
    })
    row.units.forEach((unit) => {
      /** @type {{ x: number, y: number, childIds: number[] }[]} */
      const sources = []
      if (unit.kind === 'couple') {
        const [a, b] = unit.members.map((m) => m.id)
        const childIds = childrenOfCouple(a, b).filter(
          (id) => pmap[id] && nextRowIds.has(id),
        )
        if (childIds.length === 0) return
        const elA = personEls.get(a)
        const elB = personEls.get(b)
        if (!elA || !elB) return
        const ba = elCenterBottom(elA)
        const bb = elCenterBottom(elB)
        sources.push({
          x: (ba.x + bb.x) / 2,
          y: Math.max(ba.y, bb.y),
          childIds,
        })
      } else if (unit.kind === 'single') {
        const pid = unit.members[0].id
        const childIds = (p2c[pid] || []).filter(
          (id) => pmap[id] && nextRowIds.has(id),
        )
        if (childIds.length === 0) return
        const el = personEls.get(pid)
        if (!el) return
        const p = elCenterBottom(el)
        sources.push({ x: p.x, y: p.y, childIds })
      } else if (unit.kind === 'siblings') {
        unit.members.forEach((m) => {
          const pid = m.id
          const childIds = (p2c[pid] || []).filter(
            (id) => pmap[id] && nextRowIds.has(id),
          )
          if (childIds.length === 0) return
          const el = personEls.get(pid)
          if (!el) return
          const p = elCenterBottom(el)
          sources.push({ x: p.x, y: p.y, childIds })
        })
      }

      sources.forEach((src) => {
        const targets = []
        src.childIds.forEach((cid) => {
          const el = personEls.get(cid)
          if (el) targets.push(elCenterTop(el))
        })
        if (targets.length === 0) return
        const yMid = (src.y + Math.min(...targets.map((t) => t.y))) / 2
        if (targets.length === 1) {
          const t = targets[0]
          segs.push({ d: `M ${src.x} ${src.y} L ${src.x} ${yMid} L ${t.x} ${yMid} L ${t.x} ${t.y}` })
        } else {
          const xs = targets.map((t) => t.x).sort((a, b) => a - b)
          const xL = xs[0]
          const xR = xs[xs.length - 1]
          segs.push({ d: `M ${src.x} ${src.y} L ${src.x} ${yMid} L ${xL} ${yMid} L ${xR} ${yMid}` })
          targets.forEach((t) => {
            segs.push({ d: `M ${t.x} ${yMid} L ${t.x} ${t.y}` })
          })
        }
      })
    })
  }

  lineSegments.value = segs
}

let ro
function scheduleMeasure() {
  nextTick(() => {
    requestAnimationFrame(() => {
      measureLines()
      requestAnimationFrame(() => measureLines())
    })
  })
}

onMounted(() => {
  scheduleMeasure()
  ro = typeof ResizeObserver !== 'undefined' ? new ResizeObserver(scheduleMeasure) : null
  nextTick(() => {
    if (ro && canvasRef.value) ro.observe(canvasRef.value)
  })
  window.addEventListener('resize', scheduleMeasure)
})

onBeforeUnmount(() => {
  if (ro) ro.disconnect()
  window.removeEventListener('resize', scheduleMeasure)
})

watch(
  [() => props.persons, () => props.relations, layoutRows],
  () => scheduleMeasure(),
  { deep: true },
)
</script>

<style scoped>
.pedigree-root {
  position: relative;
  width: 100%;
}

.empty {
  text-align: center;
  color: #999;
  padding: 24px;
}

.pedigree-scroll {
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
  padding: 16px 0 8px;
}

.pedigree-canvas {
  position: relative;
  display: inline-block;
  min-width: 100%;
  vertical-align: top;
}

.pedigree-svg {
  position: absolute;
  left: 0;
  top: 0;
  pointer-events: none;
  z-index: 0;
  overflow: visible;
}

.pedigree-rows {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  gap: 48px;
  padding: 8px 12px 32px;
}

.ped-row {
  display: flex;
  flex-direction: row;
  flex-wrap: nowrap;
  align-items: flex-start;
  justify-content: flex-start;
  gap: 36px;
}

.ped-unit {
  flex-shrink: 0;
}

.ped-unit-inner {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.ped-marriage {
  display: flex;
  flex-direction: row;
  align-items: flex-start;
  gap: 0;
}

.ped-marriage-line {
  align-self: center;
  width: 24px;
  height: 1px;
  background: #1a1a1a;
  margin-top: 20px;
  flex-shrink: 0;
}

.ped-sibling-row {
  display: flex;
  flex-direction: row;
  align-items: flex-start;
  gap: 20px;
}

.ped-sibling-slot {
  flex-shrink: 0;
}

.ped-anchor-person {
  display: block;
}

.pedigree-footer {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 6px;
  margin-top: 8px;
  padding-right: 4px;
}

.pedigree-legend {
  border: 1px dashed #cbd5e0;
  background: rgba(255, 255, 255, 0.95);
  padding: 8px 10px;
  font-size: 11px;
  color: #4a5568;
  border-radius: 4px;
}

.legend-title {
  font-weight: 600;
  margin-bottom: 6px;
  color: #2d3748;
}

.legend-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 4px;
}

.leg-sym {
  width: 14px;
  height: 14px;
  flex-shrink: 0;
  box-sizing: border-box;
}

.leg-sym.male {
  border-radius: 1px;
  background: #cfe8ff;
  border: 1.5px solid #2b6cb0;
}

.leg-sym.female {
  border-radius: 50%;
  background: #ffe0ec;
  border: 1.5px solid #c5306b;
}

.leg-sym.unknown {
  border-radius: 2px;
  background: #e8e8e8;
  border: 1.5px solid #718096;
}

.pedigree-hint {
  text-align: center;
  font-size: 12px;
  color: #ccc;
  margin: 0;
  width: 100%;
}

@media (max-width: 768px) {
  .ped-row {
    gap: 24px;
  }

  .pedigree-footer {
    align-items: stretch;
  }

  .pedigree-legend {
    width: fit-content;
    align-self: flex-end;
  }
}
</style>
