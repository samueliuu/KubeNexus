import React from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { ProLayout } from '@ant-design/pro-components'
import {
  DashboardOutlined,
  ClusterOutlined,
  AppstoreOutlined,
  CloudUploadOutlined,
  TeamOutlined,
  AlertOutlined,
  SettingOutlined,
  ToolOutlined,
  LogoutOutlined,
} from '@ant-design/icons'
import { useAuth } from '../contexts/AuthContext'

const menuRoute = {
  path: '/',
  routes: [
    { path: '/dashboard', name: '仪表盘', icon: <DashboardOutlined /> },
    { path: '/clusters', name: '集群管理', icon: <ClusterOutlined /> },
    { path: '/applications', name: '应用市场', icon: <AppstoreOutlined /> },
    { path: '/deployments', name: '部署管理', icon: <CloudUploadOutlined /> },
    { path: '/organizations', name: '组织管理', icon: <TeamOutlined /> },
    { path: '/alerts', name: '监控告警', icon: <AlertOutlined /> },
    { path: '/configs', name: '配置中心', icon: <ToolOutlined /> },
    { path: '/settings', name: '系统设置', icon: <SettingOutlined /> },
  ],
}

const AppLayout: React.FC = () => {
  const navigate = useNavigate()
  const location = useLocation()
  const { user, logout } = useAuth()

  return (
    <ProLayout
      title="KubeNexus"
      logo={false}
      layout="mix"
      fixSiderbar
      fixedHeader
      route={menuRoute}
      location={{ pathname: location.pathname }}
      menuItemRender={(item, dom) => (
        <div onClick={() => item.path && navigate(item.path)}>{dom}</div>
      )}
      avatarProps={{
        title: user?.username,
        size: 'small',
        render: (_, dom) => (
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, cursor: 'pointer' }}
            onClick={() => { logout(); navigate('/login') }}>
            {dom}
            <span style={{ fontSize: 13 }}>{user?.role === 'admin' ? '管理员' : user?.role}</span>
            <LogoutOutlined />
          </div>
        ),
      }}
      contentStyle={{ padding: 24 }}
      footerRender={() => (
        <div style={{ textAlign: 'center', padding: '8px 0', color: '#999', fontSize: 12 }}>
          衡牧KubeNexusK3s多集群管理系统 V1.0.0
        </div>
      )}
    >
      <Outlet />
    </ProLayout>
  )
}

export default AppLayout
