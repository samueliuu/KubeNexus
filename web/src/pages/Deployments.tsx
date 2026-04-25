import React, { useRef } from 'react'
import { ProTable, ModalForm, ProFormSelect, ProFormDigit, ProFormTextArea } from '@ant-design/pro-components'
import { Button, Tag, message } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { deploymentApi, applicationApi, clusterApi, Deployment, Application, Cluster } from '../api'

const statusColors: Record<string, string> = {
  pending: 'default', deployed: 'green', syncing: 'processing',
  synced: 'green', error: 'red', stopped: 'default', drifted: 'warning',
}

const Deployments: React.FC = () => {
  const actionRef = useRef<any>()

  const columns: any[] = [
    { title: '应用名称', dataIndex: 'name' },
    { title: '集群', dataIndex: 'cluster_id', hideInSearch: true },
    { title: '状态', dataIndex: 'status', valueType: 'select',
      valueEnum: { pending: { text: '待部署' }, deployed: { text: '已部署' }, syncing: { text: '同步中' }, synced: { text: '已同步' }, error: { text: '错误' }, drifted: { text: '漂移' } },
      render: (_: string) => <Tag color={statusColors[_] || 'default'}>{_}</Tag> },
    { title: '副本数', dataIndex: 'replicas', hideInSearch: true },
    { title: '版本', dataIndex: 'version', hideInSearch: true },
    { title: '创建时间', dataIndex: 'created_at', valueType: 'dateTime', hideInSearch: true },
    { title: '操作', valueType: 'option', render: (_: any, r: Deployment) => [
      <a key="del" style={{ color: '#ff4d4f' }} onClick={() => { deploymentApi.delete(r.id).then(() => { message.success('删除成功'); actionRef.current?.reload() }) }}>删除</a>,
    ]},
  ]

  return (
    <ProTable<Deployment>
      headerTitle="部署管理"
      actionRef={actionRef}
      rowKey="id"
      columns={columns}
      request={async () => { const res = await deploymentApi.list(); return { data: res.data.items || [], success: true } }}
      toolBarRender={() => [
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
              message.success('部署创建成功'); actionRef.current?.reload(); return true
            } catch (err: any) { message.error(err.response?.data?.error || '创建失败'); return false }
          }}
        >
          <ProFormSelect name="application_id" label="应用" rules={[{ required: true }]}
            request={async () => { const res = await applicationApi.list(); return (res.data.items || []).map((a: Application) => ({ label: a.display_name || a.name, value: a.id })) }} />
          <ProFormSelect name="cluster_id" label="目标集群" rules={[{ required: true }]}
            request={async () => { const res = await clusterApi.list(); return (res.data.items || []).map((c: Cluster) => ({ label: c.display_name || c.name, value: c.id })) }} />
          <ProFormDigit name="replicas" label="副本数" min={1} fieldProps={{ defaultValue: 1 }} />
          <ProFormTextArea name="values" label="自定义 Values (YAML)" />
        </ModalForm>,
      ]}
    />
  )
}

export default Deployments
