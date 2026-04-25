import React, { useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import { ProTable, ModalForm, ProFormText, ProFormSelect } from '@ant-design/pro-components'
import { Button, Tag, message } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { clusterApi, organizationApi, Cluster, Organization } from '../api'

const statusColors: Record<string, string> = {
  active: 'green', registered: 'blue', unavailable: 'red',
  provisioning: 'orange', degraded: 'warning', error: 'red',
}
const statusLabels: Record<string, string> = {
  active: '在线', registered: '已注册', unavailable: '离线',
  provisioning: '部署中', degraded: '降级', error: '错误',
}

const Clusters: React.FC = () => {
  const actionRef = useRef<any>()
  const navigate = useNavigate()

  const columns: any[] = [
    { title: '名称', dataIndex: 'name', render: (_: string, r: Cluster) => <a onClick={() => navigate(`/clusters/${r.id}`)}>{_}</a> },
    { title: '显示名称', dataIndex: 'display_name', hideInSearch: true },
    { title: '状态', dataIndex: 'status', valueType: 'select', valueEnum: Object.fromEntries(Object.entries(statusLabels).map(([k, v]) => [k, { text: v }])),
      render: (_: string) => <Tag color={statusColors[_] || 'default'}>{statusLabels[_] || _}</Tag> },
    { title: '版本', dataIndex: 'version', hideInSearch: true },
    { title: '节点数', dataIndex: 'node_count', hideInSearch: true },
    { title: '地域', dataIndex: 'region', hideInSearch: true },
    { title: '组织', dataIndex: 'org_name', hideInSearch: true },
    { title: '最后心跳', dataIndex: 'last_heartbeat', valueType: 'dateTime', hideInSearch: true },
    { title: '操作', valueType: 'option', render: (_: any, r: Cluster) => [
      <a key="detail" onClick={() => navigate(`/clusters/${r.id}`)}>详情</a>,
      <a key="del" style={{ color: '#ff4d4f' }} onClick={() => { clusterApi.delete(r.id).then(() => { message.success('删除成功'); actionRef.current?.reload() }) }}>删除</a>,
    ]},
  ]

  return (
    <ProTable<Cluster>
      headerTitle="集群管理"
      actionRef={actionRef}
      rowKey="id"
      columns={columns}
      request={async () => {
        const res = await clusterApi.list()
        return { data: res.data.items || [], success: true }
      }}
      toolBarRender={() => [
        <ModalForm
          key="create"
          title="注册新集群"
          trigger={<Button type="primary" icon={<PlusOutlined />}>注册集群</Button>}
          modalProps={{ destroyOnHidden: true }}
          onFinish={async (values: any) => {
            try {
              await clusterApi.create(values)
              message.success('集群注册成功')
              actionRef.current?.reload()
              return true
            } catch (err: any) {
              message.error(err.response?.data?.error || '创建失败')
              return false
            }
          }}
        >
          <ProFormText name="name" label="集群名称" rules={[{ required: true }]} placeholder="例如: prod-cluster-01" />
          <ProFormText name="display_name" label="显示名称" placeholder="例如: 生产集群" />
          <ProFormSelect name="org_id" label="所属组织" request={async () => { const res = await organizationApi.list(); return (res.data.items || []).map((o: Organization) => ({ label: o.name, value: o.id })) }} placeholder="选择组织（可选）" />
          <ProFormSelect name="region" label="地域" options={[{ label: '华东', value: 'east-china' }, { label: '华南', value: 'south-china' }, { label: '华北', value: 'north-china' }, { label: '西南', value: 'southwest' }]} placeholder="选择地域" />
        </ModalForm>,
      ]}
    />
  )
}

export default Clusters
