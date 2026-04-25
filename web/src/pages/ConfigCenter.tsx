import React, { useRef } from 'react'
import { ProTable, ModalForm, ProFormText, ProFormSelect, ProFormTextArea } from '@ant-design/pro-components'
import { Button, message } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { configApi, applicationApi, organizationApi, ConfigTemplate, Application, Organization } from '../api'

const ConfigCenter: React.FC = () => {
  const actionRef = useRef<any>()

  const columns: any[] = [
    { title: '配置名称', dataIndex: 'name' },
    { title: '关联应用', dataIndex: 'application_id', hideInSearch: true },
    { title: '关联组织', dataIndex: 'org_id', hideInSearch: true },
    { title: '描述', dataIndex: 'description', ellipsis: true, hideInSearch: true },
    { title: '创建时间', dataIndex: 'created_at', valueType: 'dateTime', hideInSearch: true },
    { title: '操作', valueType: 'option', render: (_: any, r: ConfigTemplate) => [
      <a key="del" style={{ color: '#ff4d4f' }} onClick={() => { configApi.delete(r.id).then(() => { message.success('删除成功'); actionRef.current?.reload() }) }}>删除</a>,
    ]},
  ]

  return (
    <ProTable<ConfigTemplate>
      headerTitle="配置中心"
      actionRef={actionRef}
      rowKey="id"
      columns={columns}
      request={async () => { const res = await configApi.list(); return { data: res.data.items || [], success: true } }}
      toolBarRender={() => [
        <ModalForm
          key="create"
          title="创建配置模板"
          trigger={<Button type="primary" icon={<PlusOutlined />}>创建配置</Button>}
          modalProps={{ destroyOnHidden: true }}
          onFinish={async (values: any) => {
            try { await configApi.create(values); message.success('创建成功'); actionRef.current?.reload(); return true }
            catch (err: any) { message.error(err.response?.data?.error || '创建失败'); return false }
          }}
        >
          <ProFormText name="name" label="配置名称" rules={[{ required: true }]} />
          <ProFormSelect name="application_id" label="关联应用" rules={[{ required: true }]}
            request={async () => { const res = await applicationApi.list(); return (res.data.items || []).map((a: Application) => ({ label: a.display_name || a.name, value: a.id })) }} />
          <ProFormSelect name="org_id" label="关联组织"
            request={async () => { const res = await organizationApi.list(); return (res.data.items || []).map((o: Organization) => ({ label: o.name, value: o.id })) }} />
          <ProFormTextArea name="values" label="配置内容 (YAML)" rules={[{ required: true }]} />
          <ProFormTextArea name="description" label="描述" />
        </ModalForm>,
      ]}
    />
  )
}

export default ConfigCenter
