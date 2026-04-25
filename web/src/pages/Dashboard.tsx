import React, { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { ProCard, StatisticCard } from '@ant-design/pro-components'
import { Progress, message } from 'antd'
import {
  ClusterOutlined,
  CheckCircleOutlined,
  WarningOutlined,
  AppstoreOutlined,
  CloudUploadOutlined,
  TeamOutlined,
  AlertOutlined,
  SafetyCertificateOutlined,
} from '@ant-design/icons'
import { dashboardApi, licenseApi, DashboardStats, LicenseQuota } from '../api'

const Dashboard: React.FC = () => {
  const navigate = useNavigate()
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [quota, setQuota] = useState<LicenseQuota | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    // 并行请求统计数据和配额信息
    Promise.all([
      dashboardApi.getStats().then((res) => setStats(res.data)).catch(() => {}),
      licenseApi.getQuota().then((res) => setQuota(res.data)).catch(() => {}),
    ]).finally(() => setLoading(false))
  }, [])

  return (
    <div>
      <ProCard title="系统概览" headerBordered split="vertical">
        <ProCard split="horizontal">
          <ProCard split="vertical">
            <StatisticCard
              loading={loading}
              statistic={{
                title: '集群总数',
                value: stats?.total_clusters || 0,
                icon: <ClusterOutlined style={{ fontSize: 24, color: '#1890ff' }} />,
              }}
              onClick={() => navigate('/clusters')}
              style={{ cursor: 'pointer' }}
            />
            <StatisticCard
              loading={loading}
              statistic={{
                title: '在线集群',
                value: stats?.active_clusters || 0,
                icon: <CheckCircleOutlined style={{ fontSize: 24, color: '#52c41a' }} />,
                valueStyle: { color: '#52c41a' },
              }}
              onClick={() => navigate('/clusters')}
              style={{ cursor: 'pointer' }}
            />
            <StatisticCard
              loading={loading}
              statistic={{
                title: '离线集群',
                value: stats?.unavailable_clusters || 0,
                icon: <WarningOutlined style={{ fontSize: 24, color: '#ff4d4f' }} />,
                valueStyle: { color: '#ff4d4f' },
              }}
              onClick={() => navigate('/clusters')}
              style={{ cursor: 'pointer' }}
            />
          </ProCard>
          <ProCard split="vertical">
            <StatisticCard
              loading={loading}
              statistic={{
                title: '应用总数',
                value: stats?.total_applications || 0,
                icon: <AppstoreOutlined style={{ fontSize: 24, color: '#722ed1' }} />,
              }}
              onClick={() => navigate('/applications')}
              style={{ cursor: 'pointer' }}
            />
            <StatisticCard
              loading={loading}
              statistic={{
                title: '部署总数',
                value: stats?.total_deployments || 0,
                icon: <CloudUploadOutlined style={{ fontSize: 24, color: '#13c2c2' }} />,
              }}
              onClick={() => navigate('/deployments')}
              style={{ cursor: 'pointer' }}
            />
            <StatisticCard
              loading={loading}
              statistic={{
                title: '组织总数',
                value: stats?.total_organizations || 0,
                icon: <TeamOutlined style={{ fontSize: 24, color: '#fa8c16' }} />,
              }}
              onClick={() => navigate('/organizations')}
              style={{ cursor: 'pointer' }}
            />
          </ProCard>
          <ProCard>
            <StatisticCard
              loading={loading}
              statistic={{
                title: '近期告警',
                value: stats?.recent_alerts || 0,
                icon: <AlertOutlined style={{ fontSize: 24, color: '#eb2f96' }} />,
                valueStyle: { color: '#eb2f96' },
              }}
              onClick={() => navigate('/alerts')}
              style={{ cursor: 'pointer' }}
            />
          </ProCard>
        </ProCard>
      </ProCard>

      {/* License 配额展示 */}
      <ProCard
        title="License 配额"
        headerBordered
        style={{ marginTop: 24 }}
      >
        {quota ? (
          <ProCard split="vertical">
            <ProCard>
              <div style={{ textAlign: 'center' }}>
                <SafetyCertificateOutlined style={{ fontSize: 32, color: '#1890ff', marginBottom: 8 }} />
                <div style={{ fontSize: 14, color: 'rgba(0,0,0,0.45)', marginBottom: 4 }}>集群配额</div>
                <div style={{ fontSize: 20, fontWeight: 600, marginBottom: 12 }}>
                  {quota.clusters.current} / {quota.clusters.max}
                </div>
                <Progress
                  type="dashboard"
                  percent={quota.clusters.max > 0 ? Math.round((quota.clusters.current / quota.clusters.max) * 100) : 0}
                  strokeColor={quota.clusters.max > 0 && quota.clusters.current / quota.clusters.max > 0.8 ? '#ff4d4f' : '#1890ff'}
                  size={120}
                />
              </div>
            </ProCard>
            <ProCard>
              <div style={{ textAlign: 'center' }}>
                <CloudUploadOutlined style={{ fontSize: 32, color: '#13c2c2', marginBottom: 8 }} />
                <div style={{ fontSize: 14, color: 'rgba(0,0,0,0.45)', marginBottom: 4 }}>部署配额</div>
                <div style={{ fontSize: 20, fontWeight: 600, marginBottom: 12 }}>
                  {quota.deployments.current} / {quota.deployments.max}
                </div>
                <Progress
                  type="dashboard"
                  percent={quota.deployments.max > 0 ? Math.round((quota.deployments.current / quota.deployments.max) * 100) : 0}
                  strokeColor={quota.deployments.max > 0 && quota.deployments.current / quota.deployments.max > 0.8 ? '#ff4d4f' : '#13c2c2'}
                  size={120}
                />
              </div>
            </ProCard>
          </ProCard>
        ) : (
          <div style={{ textAlign: 'center', color: 'rgba(0,0,0,0.25)', padding: 24 }}>
            暂无 License 配额信息
          </div>
        )}
      </ProCard>
    </div>
  )
}

export default Dashboard
