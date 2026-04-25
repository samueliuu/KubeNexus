import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  },
)

export default api

export interface Cluster {
  id: string
  name: string
  display_name: string
  status: string
  endpoint: string
  version: string
  node_count: number
  cpu_capacity: string
  mem_capacity: string
  labels: Record<string, string>
  region: string
  org_id: string
  org_name: string
  last_heartbeat: string
  ws_connected: boolean
  created_at: string
}

export interface Application {
  id: string
  name: string
  display_name: string
  description: string
  icon: string
  chart_name: string
  chart_repo: string
  chart_version: string
  category: string
  is_saas: boolean
  default_values: string
  created_at: string
}

export interface Deployment {
  id: string
  cluster_id: string
  application_id: string
  name: string
  namespace: string
  values: string
  status: string
  actual_status: string
  replicas: number
  version: string
  actual_version: string
  drift_detail: string
  message: string
  last_synced: string
  created_at: string
}

export interface Organization {
  id: string
  name: string
  code: string
  contact: string
  phone: string
  email: string
  type: string
  description: string
  created_at: string
}

export interface License {
  id: string
  product: string
  customer_name: string
  issued_at: string
  expires_at: string
  max_clusters: number
  max_deployments: number
  features: string
  is_valid: boolean
}

export interface AlertRule {
  id: string
  name: string
  type: string
  condition: string
  severity: string
  enabled: boolean
  notify_channels: string
  last_triggered: string
}

export interface AlertRecord {
  id: string
  rule_id: string
  rule_name: string
  cluster_id: string
  severity: string
  message: string
  status: string
  triggered_at: string
  resolved_at: string
}

export interface ConfigTemplate {
  id: string
  name: string
  org_id: string
  application_id: string
  values: string
  description: string
}

export interface AuditLog {
  id: string
  user_id: string
  username: string
  action: string
  resource_type: string
  resource_id: string
  resource_name: string
  detail: string
  ip: string
  created_at: string
}

export interface User {
  id: string
  username: string
  role: string
  created_at: string
}

export interface DashboardStats {
  total_clusters: number
  active_clusters: number
  unavailable_clusters: number
  total_applications: number
  total_deployments: number
  total_organizations: number
  recent_alerts: number
}

export interface Heartbeat {
  id: string
  cluster_id: string
  node_count: number
  cpu_usage: number
  mem_usage: number
  pod_count: number
  version: string
  reported_at: string
}

export interface LicenseQuota {
  clusters: { current: number; max: number }
  deployments: { current: number; max: number }
}

export const authApi = {
  login: (username: string, password: string) =>
    api.post('/auth/login', { username, password }),
  getMe: () => api.get('/auth/me'),
}

export const clusterApi = {
  list: () => api.get<{ items: Cluster[] }>('/clusters'),
  get: (id: string) => api.get<Cluster>(`/clusters/${id}`),
  create: (data: Partial<Cluster>) => api.post('/clusters', data),
  delete: (id: string) => api.delete(`/clusters/${id}`),
  updateLabels: (id: string, labels: Record<string, string>) =>
    api.put(`/clusters/${id}/labels`, { labels }),
  rotateToken: (id: string) => api.post(`/clusters/${id}/token/rotate`),
  getInstallScript: (id: string) =>
    api.get(`/clusters/${id}/install-script`, { responseType: 'text' }),
  getRegistrationYaml: (id: string) =>
    api.get(`/clusters/${id}/registration.yaml`, { responseType: 'text' }),
  getMetrics: (id: string) =>
    api.get<{ items: Heartbeat[] }>(`/clusters/${id}/metrics`),
  getNodes: (id: string) =>
    api.get(`/clusters/${id}/nodes`),
}

export const applicationApi = {
  list: () => api.get<{ items: Application[] }>('/applications'),
  get: (id: string) => api.get<Application>(`/applications/${id}`),
  create: (data: Partial<Application>) => api.post('/applications', data),
  update: (id: string, data: Partial<Application>) => api.put(`/applications/${id}`, data),
  delete: (id: string) => api.delete(`/applications/${id}`),
}

export const deploymentApi = {
  list: (clusterId?: string) =>
    api.get<{ items: Deployment[] }>('/deployments', { params: { cluster_id: clusterId } }),
  get: (id: string) => api.get<Deployment>(`/deployments/${id}`),
  create: (data: any) => api.post('/deployments', data),
  batchCreate: (data: any) => api.post('/deployments/batch', data),
  update: (id: string, data: any) => api.put(`/deployments/${id}`, data),
  delete: (id: string) => api.delete(`/deployments/${id}`),
}

export const organizationApi = {
  list: () => api.get<{ items: Organization[] }>('/organizations'),
  get: (id: string) => api.get<Organization>(`/organizations/${id}`),
  create: (data: Partial<Organization>) => api.post('/organizations', data),
  update: (id: string, data: Partial<Organization>) => api.put(`/organizations/${id}`, data),
  delete: (id: string) => api.delete(`/organizations/${id}`),
}

export const licenseApi = {
  get: () => api.get<License>('/license'),
  activate: (data: any) => api.post('/license/activate', data),
  getQuota: () => api.get<LicenseQuota>('/license/quota'),
}

export const alertApi = {
  listRules: () => api.get<{ items: AlertRule[] }>('/alerts/rules'),
  createRule: (data: any) => api.post('/alerts/rules', data),
  updateRule: (id: string, data: any) => api.put(`/alerts/rules/${id}`, data),
  deleteRule: (id: string) => api.delete(`/alerts/rules/${id}`),
  listRecords: (params?: { cluster_id?: string; status?: string }) =>
    api.get<{ items: AlertRecord[] }>('/alerts/records', { params }),
  acknowledgeRecord: (id: string) => api.put(`/alerts/records/${id}/acknowledge`),
}

export const configApi = {
  list: (orgId?: string) => api.get<{ items: ConfigTemplate[] }>('/configs', { params: { org_id: orgId } }),
  get: (id: string) => api.get<ConfigTemplate>(`/configs/${id}`),
  create: (data: any) => api.post('/configs', data),
  update: (id: string, data: any) => api.put(`/configs/${id}`, data),
  delete: (id: string) => api.delete(`/configs/${id}`),
}

export const auditApi = {
  list: (params?: { resource_type?: string; username?: string; action?: string }) =>
    api.get<{ items: AuditLog[] }>('/audit-logs', { params }),
}

export const userApi = {
  list: () => api.get<{ items: User[] }>('/users'),
  create: (data: any) => api.post('/users', data),
  update: (id: string, data: any) => api.put(`/users/${id}`, data),
  delete: (id: string) => api.delete(`/users/${id}`),
}

export const dashboardApi = {
  getStats: () => api.get<DashboardStats>('/dashboard/stats'),
}
