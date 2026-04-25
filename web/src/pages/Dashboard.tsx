import React, { useEffect, useState } from 'react'
import { ProCard } from '@ant-design/pro-components'
import { StatisticCard } from '@ant-design/pro-components'
import {
  ClusterOutlined,
  CheckCircleOutlined,
  WarningOutlined,
  AppstoreOutlined,
  CloudUploadOutlined,
  TeamOutlined,
  AlertOutlined,
} from '@ant-design/icons'
import { dashboardApi, DashboardStats } from '../api'

const { Statistic } = StatisticCard

const Dashboard: React.FC = () => {
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    dashboardApi.getStats().then((res) => {
      setStats(res.data)
      setLoading(false)
    }).catch(() => setLoading(false))
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
            />
            <StatisticCard
              loading={loading}
              statistic={{
                title: '在线集群',
                value: stats?.active_clusters || 0,
                icon: <CheckCircleOutlined style={{ fontSize: 24, color: '#52c41a' }} />,
                valueStyle: { color: '#52c41a' },
              }}
            />
            <StatisticCard
              loading={loading}
              statistic={{
                title: '离线集群',
                value: stats?.unavailable_clusters || 0,
                icon: <WarningOutlined style={{ fontSize: 24, color: '#ff4d4f' }} />,
                valueStyle: { color: '#ff4d4f' },
              }}
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
            />
            <StatisticCard
              loading={loading}
              statistic={{
                title: '部署总数',
                value: stats?.total_deployments || 0,
                icon: <CloudUploadOutlined style={{ fontSize: 24, color: '#13c2c2' }} />,
              }}
            />
            <StatisticCard
              loading={loading}
              statistic={{
                title: '组织总数',
                value: stats?.total_organizations || 0,
                icon: <TeamOutlined style={{ fontSize: 24, color: '#fa8c16' }} />,
              }}
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
            />
          </ProCard>
        </ProCard>
      </ProCard>
    </div>
  )
}

export default Dashboard
