import axios from 'axios'

export const api = axios.create({ baseURL: '/api/proxy' })

api.interceptors.response.use(
  (resp) => resp,
  (error) => {
    // Optionally map backend error shapes
    return Promise.reject(error)
  }
)
