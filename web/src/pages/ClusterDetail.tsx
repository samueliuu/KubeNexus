import React, { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { ProCard, ProDescriptions } from '@ant-design/pro-components'
import { Button, Tag, Descriptions, message } from 'antd'
import { ArrowLeftOutlined } from '@ant-design/icons'
import { clusterApi, Cluster } from '../api'

const statusColors: Record<string, string> = {
  active: 'green', registered: 'blue', unavailable: 'red',
  provisioning: 'orange', degraded: 'warning', error: 'red',
}
const statusLabels: Record<string, string> = {
  active: '在线', registered: '已注册', unavailable: '离线',
  provisioning: '部署中', degraded: '降级', error: '错误',
}

const ClusterDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [cluster, setCluster] = useState<Cluster | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (id) {
      clusterApi.get(id).then((res) => {
        setCluster(res.data)
        setLoading(false)
      }).catch(() => setLoading(false))
    }
  }, [id])

  const handleRotateToken = async () => {
    if (!id) return
    try {
      await clusterApi.rotateToken(id)
      message.success('Token 已轮换')
      clusterApi.get(id).then((res) => setCluster(res.data))
    } catch (err: any) {
      message.error(err.response?.data?.error || '轮换失败')
    }
  }

  if (!cluster) return <ProCard loading={loading} />

  return (
    <div>
      <Button icon={<ArrowLeftOutlined />} style={{ marginBottom: 16 }} onClick={() => navigate('/clusters')}>返回集群列表</Button>
      <ProCard title="集群详情" headerBordered loading={loading}>
        <Descriptions column={2}>
          <Descriptions.Item label="名称">{cluster.name}</Descriptions.Item>
          <Descriptions.Item label="显示名称">{cluster.display_name || '-'}</Descriptions.Item>
          <Descriptions.Item label="状态"><Tag color={statusColors[cluster.status] || 'default'}>{statusLabels[cluster.status] || cluster.status}</Tag></Descriptions.Item>
          <Descriptions.Item label="版本">{cluster.version || '-'}</Descriptions.Item>
          <Descriptions.Item label="节点数">{cluster.node_count}</Descriptions.Item>
          <Descriptions.Item label="组织">{cluster.org_name || '-'}</Descriptions.Item>
          <Descriptions.Item label="地域">{cluster.region || '-'}</Descriptions.Item>
          <Descriptions.Item label="WebSocket"><Tag color={cluster.ws_connected ? 'green' : 'default'}>{cluster.ws_connected ? '已连接' : '未连接'}</Tag></Descriptions.Item>
          <Descriptions.Item label="CPU容量">{cluster.cpu_capacity || '-'}</Descriptions.Item>
          <Descriptions.Item label="内存容量">{cluster.mem_capacity || '-'}</Descriptions.Item>
          <Descriptions.Item label="最后心跳">{cluster.last_heartbeat ? new Date(cluster.last_heartbeat).toLocaleString() : '-'}</Descriptions.Item>
          <Descriptions.Item label="创建时间">{cluster.created_at ? new Date(cluster.created_at).toLocaleString() : '-'}</Descriptions.Item>
        </Descriptions>
      </ProCard>
      <ProCard title="操作" headerBordered style={{ marginTop: 24 }}>
        <Button type="primary" onClick={handleRotateToken}>轮换 Token</Button>
      </ProCard>
    </div>
  )
}

export default ClusterDetail
