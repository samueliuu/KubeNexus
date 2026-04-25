import React, { useRef, useState, useEffect } from 'react'
import { ProTable, ModalForm, ProFormSelect, ProFormDigit, ProFormTextArea } from '@ant-design/pro-components'
import { Button, Tag, message, Modal } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { deploymentApi, applicationApi, clusterApi, Deployment, Application, Cluster } from '../api'

const statusColors: Record<string, string> = {
  pending: 'default', deployed: 'green', syncing: 'processing',
  synced: 'green', error: 'red', stopped: 'default', drifted: 'warning',
}

const Deployments: React.FC = () => {
  const actionRef = useRef<any>()
  const [clusterMap, setClusterMap] = useState<Record<string, string>>({})
  const [appMap, setAppMap] = useState<Record<string, string>>({})
  const [editModalVisible, setEditModalVisible] = useState(false)
  const [editingRecord, setEditingRecord] = useState<Deployment | null>(null)

  // 加载集群和应用映射表，用于 ID -> 名称转换
  useEffect(() => {
    clusterApi.list().then((res) => {
      const map: Record<string, string> = {}
      ;(res.data.items || []).forEach((c: Cluster) => { map[c.id] = c.display_name || c.name })
      setClusterMap(map)
    }).catch(() => {})
    applicationApi.list().then((res) => {
      const map: Record<string, string> = {}
      ;(res.data.items || []).forEach((a: Application) => { map[a.id] = a.display_name || a.name })
      setAppMap(map)
    }).catch(() => {})
  }, [])

  const columns: any[] = [
    { title: '应用名称', dataIndex: 'name' },
    {
      title: '所属应用', dataIndex: 'application_id', hideInSearch: true,
      render: (_: string) => appMap[_] || _,
    },
    {
      title: '集群', dataIndex: 'cluster_id', hideInSearch: true,
      render: (_: string) => clusterMap[_] || _,
    },
    {
      title: '状态', dataIndex: 'status', valueType: 'select',
      valueEnum: {
        pending: { text: '待部署' }, deployed: { text: '已部署' },
        syncing: { text: '同步中' }, synced: { text: '已同步' },
        error: { text: '错误' }, drifted: { text: '漂移' },
      },
      render: (_: string) => <Tag color={statusColors[_] || 'default'}>{_}</Tag>,
    },
    { title: '副本数', dataIndex: 'replicas', hideInSearch: true },
    { title: '版本', dataIndex: 'version', hideInSearch: true },
    { title: '创建时间', dataIndex: 'created_at', valueType: 'dateTime', hideInSearch: true },
    {
      title: '操作', valueType: 'option',
      render: (_: any, r: Deployment) => [
        <a
          key="edit"
          onClick={() => { setEditingRecord(r); setEditModalVisible(true) }}
        >编辑</a>,
        <a
          key="del"
          style={{ color: '#ff4d4f' }}
          onClick={() => {
            Modal.confirm({
              title: '确认删除',
              content: `确定要删除部署 "${r.name}" 吗？此操作不可恢复。`,
              okType: 'danger',
              onOk: () => deploymentApi.delete(r.id).then(() => {
                message.success('删除成功')
                actionRef.current?.reload()
              }).catch((err: any) => { message.error(err.response?.data?.error || '操作失败') }),
            })
          }}
        >删除</a>,
      ],
    },
  ]

  return (
    <>
      <ProTable<Deployment>
        headerTitle="部署管理"
        actionRef={actionRef}
        rowKey="id"
        columns={columns}
        request={async () => {
          const res = await deploymentApi.list()
          return { data: res.data.items || [], success: true }
        }}
        toolBarRender={() => [
          /* 批量部署 */
          <ModalForm
            key="batch"
            title="批量部署"
            trigger={<Button icon={<PlusOutlined />}>批量部署</Button>}
            modalProps={{ destroyOnHidden: true }}
            onFinish={async (values: any) => {
              try {
                const appRes = await applicationApi.get(values.application_id)
                const app = appRes.data
                await deploymentApi.batchCreate({
                  application_id: values.application_id,
                  name: app.name,
                  namespace: values.namespace || 'default',
                  cluster_ids: values.cluster_ids,
                  replicas: values.replicas,
                  values_overrides: values.values_overrides || '',
                })
                message.success('批量部署创建成功')
                actionRef.current?.reload()
                return true
              } catch (err: any) {
                message.error(err.response?.data?.error || '批量部署失败')
                return false
              }
            }}
          >
            <ProFormSelect
              name="application_id" label="应用" rules={[{ required: true }]}
              request={async () => {
                const res = await applicationApi.list()
                return (res.data.items || []).map((a: Application) => ({
                  label: a.display_name || a.name, value: a.id,
                }))
              }}
            />
            <ProFormSelect
              name="cluster_ids" label="目标集群" rules={[{ required: true }]}
              mode="multiple"
              request={async () => {
                const res = await clusterApi.list()
                return (res.data.items || []).map((c: Cluster) => ({
                  label: c.display_name || c.name, value: c.id,
                }))
              }}
            />
            <ProFormDigit
              name="replicas" label="副本数" min={1}
              fieldProps={{ defaultValue: 1 }}
            />
            <ProFormTextArea
              name="values_overrides" label="自定义 Values (YAML)"
            />
          </ModalForm>,

          /* 创建单个部署 */
          <ModalForm
            key="create"
            title="创建部署"
            trigger={<Button type="primary" icon={<PlusOutlined />}>创建部署</Button>}
            modalProps={{ destroyOnHidden: true }}
            onFinish={async (values: any) => {
              try {
                const appRes = await applicationApi.get(values.application_id)
                const app = appRes.data
                await deploymentApi.create({ ...values, name: app.name, version: app.chart_version })
                message.success('部署创建成功')
                actionRef.current?.reload()
                return true
              } catch (err: any) {
                message.error(err.response?.data?.error || '创建失败')
                return false
              }
            }}
          >
            <ProFormSelect
              name="application_id" label="应用" rules={[{ required: true }]}
              request={async () => {
                const res = await applicationApi.list()
                return (res.data.items || []).map((a: Application) => ({
                  label: a.display_name || a.name, value: a.id,
                }))
              }}
            />
            <ProFormSelect
              name="cluster_id" label="目标集群" rules={[{ required: true }]}
              request={async () => {
                const res = await clusterApi.list()
                return (res.data.items || []).map((c: Cluster) => ({
                  label: c.display_name || c.name, value: c.id,
                }))
              }}
            />
            <ProFormDigit
              name="replicas" label="副本数" min={1}
              fieldProps={{ defaultValue: 1 }}
            />
            <ProFormTextArea name="values" label="自定义 Values (YAML)" />
          </ModalForm>,
        ]}
      />

      {/* 编辑部署弹窗 */}
      <ModalForm
        title="编辑部署"
        open={editModalVisible}
        modalProps={{ destroyOnHidden: true }}
        initialValues={editingRecord ? {
          replicas: editingRecord.replicas,
          values: editingRecord.values,
        } : {}}
        onOpenChange={(visible: boolean) => {
          setEditModalVisible(visible)
          if (!visible) setEditingRecord(null)
        }}
        onFinish={async (values: any) => {
          if (!editingRecord) return false
          try {
            await deploymentApi.update(editingRecord.id, {
              replicas: values.replicas,
              values: values.values,
            })
            message.success('编辑成功')
            actionRef.current?.reload()
            return true
          } catch (err: any) {
            message.error(err.response?.data?.error || '编辑失败')
            return false
          }
        }}
      >
        <ProFormDigit
          name="replicas" label="副本数" min={1}
          rules={[{ required: true }]}
        />
        <ProFormTextArea name="values" label="自定义 Values (YAML)" />
      </ModalForm>
    </>
  )
}

export default Deployments
