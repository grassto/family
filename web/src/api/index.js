import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
})

// Family
export const familyApi = {
  list: () => api.get('/families'),
  get: (id) => api.get(`/families/${id}`),
  create: (data) => api.post('/families', data),
  update: (id, data) => api.put(`/families/${id}`, data),
  remove: (id) => api.delete(`/families/${id}`),
  exportOne: (id) => api.get(`/families/${id}/export`),
  import: (data) => api.post('/families/import', data),
}

// Person
export const personApi = {
  list: (familyId, keyword) => api.get('/persons', { params: { family_id: familyId, keyword } }),
  get: (id) => api.get(`/persons/${id}`),
  create: (data) => api.post('/persons', data),
  update: (id, data) => api.put(`/persons/${id}`, data),
  remove: (id) => api.delete(`/persons/${id}`),
}

// Relation
export const relationApi = {
  getByPerson: (personId) => api.get(`/persons/${personId}/relations`),
  getByFamily: (familyId) => api.get(`/families/${familyId}/relations`),
  create: (data) => api.post('/relations', data),
  update: (id, data) => api.put(`/relations/${id}`, data),
  remove: (id) => api.delete(`/relations/${id}`),
  types: () => api.get('/relation-types'),
}

// Birthday
export const birthdayApi = {
  today: () => api.get('/birthdays/today'),
  upcoming: (days = 30) => api.get('/birthdays/upcoming', { params: { days } }),
}

export default api
