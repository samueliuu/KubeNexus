import React, { useEffect, useState, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { ProCard } from '@ant-design/pro-components'
import {
  Descriptions,
  Tag,
  Button,
  Input,
  message,
  Modal,
  Table,
  Space,
} from 'antd'
import {
  ArrowLeftOutlined,
  PlusOutlined,
  CopyOutlined,
  CodeOutlined,
  FileTextOutlined,
  LineChartOutlined,
  TagsOutlined,
  ReloadOutlined,
} from '@ant-design/icons'
import { clusterApi, Cluster, Heartbeat } from '../api'

const statusColors: Record<string, string> = {
  active: 'green',
  registered: 'blue',
  unavailable: 'red',
  provisioning: 'orange',
  degraded: 'warning',
  error: 'red',
}
const statusLabels: Record<string, string> = {
  active: '在线',
  registered: '已注册',
  unavailable: '离线',
  provisioning: '部署中',
  degraded: '降级',
  error: '错误',
}

const ClusterDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()

  // 集群详情
  const [cluster, setCluster] = useState<Cluster | null>(null)
  const [loading, setLoading] = useState(true)

  // 标签管理
  const [labelModalVisible, setLabelModalVisible] = useState(false)
  const [newLabelKey, setNewLabelKey] = useState('')
  const [newLabelValue, setNewLabelValue] = useState('')
  const [labelSaving, setLabelSaving] = useState(false)

  // 指标数据
  const [metrics, setMetrics] = useState<Heartbeat[]>([])
  const [metricsLoading, setMetricsLoading] = useState(false)

  // 安装脚本
  const [installScript, setInstallScript] = useState('')
  const [scriptLoading, setScriptLoading] = useState(false)
  const [scriptModalVisible, setScriptModalVisible] = useState(false)

  // 注册YAML
  const [registrationYaml, setRegistrationYaml] = useState('')
  const [yamlLoading, setYamlLoading] = useState(false)
  const [yamlModalVisible, setYamlModalVisible] = useState(false)

  // 加载集群详情
  const fetchCluster = useCallback(() => {
    if (!id) return
    setLoading(true)
    clusterApi
      .get(id)
      .then((res) => setCluster(res.data))
      .catch(() => message.error('获取集群详情失败'))
      .finally(() => setLoading(false))
  }, [id])

  useEffect(() => {
    fetchCluster()
  }, [fetchCluster])

  // 加载指标数据
  const fetchMetrics = useCallback(() => {
    if (!id) return
    setMetricsLoading(true)
    clusterApi
      .getMetrics(id)
      .then((res) => setMetrics(res.data.items || []))
      .catch(() => message.error('获取指标数据失败'))
      .finally(() => setMetricsLoading(false))
  }, [id])

  useEffect(() => {
    fetchMetrics()
  }, [fetchMetrics])

  // 加载安装脚本
  const fetchInstallScript = useCallback(() => {
    if (!id) return
    setScriptLoading(true)
    clusterApi
      .getInstallScript(id)
      .then((res) => {
        // responseType 为 text 时，数据直接在 res.data 中
        setInstallScript(typeof res.data === 'string' ? res.data : String(res.data))
      })
      .catch(() => message.error('获取安装脚本失败'))
      .finally(() => setScriptLoading(false))
  }, [id])

  // 加载注册YAML
  const fetchRegistrationYaml = useCallback(() => {
    if (!id) return
    setYamlLoading(true)
    clusterApi
      .getRegistrationYaml(id)
      .then((res) => {
        setRegistrationYaml(typeof res.data === 'string' ? res.data : String(res.data))
      })
      .catch(() => message.error('获取注册YAML失败'))
      .finally(() => setYamlLoading(false))
  }, [id])

  // 轮换Token
  const handleRotateToken = async () => {
    if (!id) return
    try {
      await clusterApi.rotateToken(id)
      message.success('Token 已轮换')
      fetchCluster()
    } catch (err: any) {
      message.error(err.response?.data?.error || '轮换失败')
    }
  }

  // 添加标签
  const handleAddLabel = async () => {
    if (!id || !cluster) return
    const key = newLabelKey.trim()
    const value = newLabelValue.trim()
    if (!key) {
      message.warning('标签键不能为空')
      return
    }
    setLabelSaving(true)
    try {
      const newLabels = { ...cluster.labels, [key]: value }
      await clusterApi.updateLabels(id, newLabels)
      message.success('标签已添加')
      setLabelModalVisible(false)
      setNewLabelKey('')
      setNewLabelValue('')
      fetchCluster()
    } catch (err: any) {
      message.error(err.response?.data?.error || '添加标签失败')
    } finally {
      setLabelSaving(false)
    }
  }

  // 删除标签
  const handleDeleteLabel = async (key: string) => {
    if (!id || !cluster) return
    try {
      const newLabels = { ...cluster.labels }
      delete newLabels[key]
      await clusterApi.updateLabels(id, newLabels)
      message.success('标签已删除')
      fetchCluster()
    } catch (err: any) {
      message.error(err.response?.data?.error || '删除标签失败')
    }
  }

  // 复制到剪贴板
  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard
      .writeText(text)
      .then(() => message.success(`${label}已复制到剪贴板`))
      .catch(() => message.error('复制失败'))
  }

  // 指标表格列定义
  const metricsColumns = [
    {
      title: '上报时间',
      dataIndex: 'reported_at',
      key: 'reported_at',
      width: 180,
      render: (val: string) => (val ? new Date(val).toLocaleString() : '-'),
    },
    {
      title: 'CPU使用率',
      dataIndex: 'cpu_usage',
      key: 'cpu_usage',
      width: 120,
      render: (val: number) => (val != null ? `${val.toFixed(1)}%` : '-'),
    },
    {
      title: '内存使用率',
      dataIndex: 'mem_usage',
      key: 'mem_usage',
      width: 120,
      render: (val: number) => (val != null ? `${val.toFixed(1)}%` : '-'),
    },
    {
      title: '节点数',
      dataIndex: 'node_count',
      key: 'node_count',
      width: 80,
    },
    {
      title: 'Pod数',
      dataIndex: 'pod_count',
      key: 'pod_count',
      width: 80,
    },
    {
      title: '版本',
      dataIndex: 'version',
      key: 'version',
      width: 100,
    },
  ]

  if (!cluster) return <ProCard loading={loading} />

  // 标签列表渲染
  const labelEntries = Object.entries(cluster.labels || {})

  return (
    <div>
      <Button
        icon={<ArrowLeftOutlined />}
        style={{ marginBottom: 16 }}
        onClick={() => navigate('/clusters')}
      >
        返回集群列表
      </Button>

      {/* 集群基本信息 */}
      <ProCard title="集群详情" headerBordered loading={loading}>
        <Descriptions column={2}>
          <Descriptions.Item label="名称">{cluster.name}</Descriptions.Item>
          <Descriptions.Item label="显示名称">{cluster.display_name || '-'}</Descriptions.Item>
          <Descriptions.Item label="状态">
            <Tag color={statusColors[cluster.status] || 'default'}>
              {statusLabels[cluster.status] || cluster.status}
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label="版本">{cluster.version || '-'}</Descriptions.Item>
          <Descriptions.Item label="节点数">{cluster.node_count}</Descriptions.Item>
          <Descriptions.Item label="组织">{cluster.org_name || '-'}</Descriptions.Item>
          <Descriptions.Item label="地域">{cluster.region || '-'}</Descriptions.Item>
          <Descriptions.Item label="WebSocket">
            <Tag color={cluster.ws_connected ? 'green' : 'default'}>
              {cluster.ws_connected ? '已连接' : '未连接'}
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label="CPU容量">{cluster.cpu_capacity || '-'}</Descriptions.Item>
          <Descriptions.Item label="内存容量">{cluster.mem_capacity || '-'}</Descriptions.Item>
          <Descriptions.Item label="最后心跳">
            {cluster.last_heartbeat ? new Date(cluster.last_heartbeat).toLocaleString() : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="创建时间">
            {cluster.created_at ? new Date(cluster.created_at).toLocaleString() : '-'}
          </Descriptions.Item>
        </Descriptions>
      </ProCard>

      {/* 标签管理 */}
      <ProCard
        title={
          <Space>
            <TagsOutlined />
            <span>标签管理</span>
          </Space>
        }
        headerBordered
        style={{ marginTop: 24 }}
        extra={
          <Button
            type="primary"
            size="small"
            icon={<PlusOutlined />}
            onClick={() => setLabelModalVisible(true)}
          >
            添加标签
          </Button>
        }
      >
        {labelEntries.length > 0 ? (
          <Space wrap>
            {labelEntries.map(([key, value]) => (
              <Tag
                key={key}
                closable
                onClose={(e) => {
                  e.preventDefault()
                  handleDeleteLabel(key)
                }}
                style={{ padding: '4px 8px' }}
              >
                {key}: {value}
              </Tag>
            ))}
          </Space>
        ) : (
          <span style={{ color: '#999' }}>暂无标签</span>
        )}
      </ProCard>

      {/* 指标数据 */}
      <ProCard
        title={
          <Space>
            <LineChartOutlined />
            <span>指标趋势</span>
          </Space>
        }
        headerBordered
        style={{ marginTop: 24 }}
        extra={
          <Button
            size="small"
            icon={<ReloadOutlined />}
            onClick={fetchMetrics}
            loading={metricsLoading}
          >
            刷新
          </Button>
        }
      >
        <Table
          rowKey="id"
          columns={metricsColumns}
          dataSource={metrics}
          loading={metricsLoading}
          size="small"
          pagination={{ pageSize: 10, showSizeChanger: false }}
          scroll={{ x: 680 }}
        />
      </ProCard>

      {/* 操作区域 */}
      <ProCard title="操作" headerBordered style={{ marginTop: 24 }}>
        <Space wrap>
          <Button type="primary" onClick={handleRotateToken}>
            轮换 Token
          </Button>
          <Button
            icon={<CodeOutlined />}
            loading={scriptLoading}
            onClick={() => {
              fetchInstallScript()
              setScriptModalVisible(true)
            }}
          >
            查看安装脚本
          </Button>
          <Button
            icon={<FileTextOutlined />}
            loading={yamlLoading}
            onClick={() => {
              fetchRegistrationYaml()
              setYamlModalVisible(true)
            }}
          >
            查看注册YAML
          </Button>
        </Space>
      </ProCard>

      {/* 添加标签弹窗 */}
      <Modal
        title="添加标签"
        open={labelModalVisible}
        onOk={handleAddLabel}
        onCancel={() => {
          setLabelModalVisible(false)
          setNewLabelKey('')
          setNewLabelValue('')
        }}
        confirmLoading={labelSaving}
        destroyOnHidden
      >
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <div>
            <div style={{ marginBottom: 4 }}>键</div>
            <Input
              placeholder="请输入标签键"
              value={newLabelKey}
              onChange={(e) => setNewLabelKey(e.target.value)}
            />
          </div>
          <div>
            <div style={{ marginBottom: 4 }}>值</div>
            <Input
              placeholder="请输入标签值"
              value={newLabelValue}
              onChange={(e) => setNewLabelValue(e.target.value)}
            />
          </div>
        </div>
      </Modal>

      {/* 安装脚本弹窗 */}
      <Modal
        title="安装脚本"
        open={scriptModalVisible}
        onCancel={() => setScriptModalVisible(false)}
        footer={
          <Space>
            <Button
              icon={<CopyOutlined />}
              type="primary"
              onClick={() => copyToClipboard(installScript, '安装脚本')}
            >
              复制
            </Button>
            <Button onClick={() => setScriptModalVisible(false)}>关闭</Button>
          </Space>
        }
        width={720}
        destroyOnHidden
      >
        <pre
          style={{
            background: '#1e1e1e',
            color: '#d4d4d4',
            padding: 16,
            borderRadius: 8,
            overflow: 'auto',
            maxHeight: 480,
            fontSize: 13,
            lineHeight: 1.6,
            margin: 0,
          }}
        >
          {installScript || '加载中...'}
        </pre>
      </Modal>

      {/* 注册YAML弹窗 */}
      <Modal
        title="注册YAML"
        open={yamlModalVisible}
        onCancel={() => setYamlModalVisible(false)}
        footer={
          <Space>
            <Button
              icon={<CopyOutlined />}
              type="primary"
              onClick={() => copyToClipboard(registrationYaml, '注册YAML')}
            >
              复制
            </Button>
            <Button onClick={() => setYamlModalVisible(false)}>关闭</Button>
          </Space>
        }
        width={720}
        destroyOnHidden
      >
        <pre
          style={{
            background: '#1e1e1e',
            color: '#d4d4d4',
            padding: 16,
            borderRadius: 8,
            overflow: 'auto',
            maxHeight: 480,
            fontSize: 13,
            lineHeight: 1.6,
            margin: 0,
          }}
        >
          {registrationYaml || '加载中...'}
        </pre>
      </Modal>
    </div>
  )
}

export default ClusterDetail
