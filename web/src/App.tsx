import React from 'react'
import { Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider, useAuth } from './contexts/AuthContext'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import Clusters from './pages/Clusters'
import ClusterDetail from './pages/ClusterDetail'
import Applications from './pages/Applications'
import Deployments from './pages/Deployments'
import Organizations from './pages/Organizations'
import Alerts from './pages/Alerts'
import ConfigCenter from './pages/ConfigCenter'
import Settings from './pages/Settings'
import AppLayout from './components/AppLayout'

const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { token, loading } = useAuth()
  if (loading) return null
  if (!token) return <Navigate to="/login" />
  return <>{children}</>
}

const App: React.FC = () => {
  return (
    <AuthProvider>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/" element={<ProtectedRoute><AppLayout /></ProtectedRoute>}>
          <Route index element={<Navigate to="/dashboard" />} />
          <Route path="dashboard" element={<Dashboard />} />
          <Route path="clusters" element={<Clusters />} />
          <Route path="clusters/:id" element={<ClusterDetail />} />
          <Route path="applications" element={<Applications />} />
          <Route path="deployments" element={<Deployments />} />
          <Route path="organizations" element={<Organizations />} />
          <Route path="alerts" element={<Alerts />} />
          <Route path="configs" element={<ConfigCenter />} />
          <Route path="settings" element={<Settings />} />
        </Route>
      </Routes>
    </AuthProvider>
  )
}

export default App
