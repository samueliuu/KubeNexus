import React, { useRef } from 'react'
import { ProTable, ModalForm, ProFormText, ProFormSelect, ProFormTextArea } from '@ant-design/pro-components'
import { Button, Tag, message } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { organizationApi, Organization } from '../api'

const typeLabels: Record<string, string> = { department: '部门', project: '项目', team: '团队' }

const Organizations: React.FC = () => {
  const actionRef = useRef<any>()

  const columns: any[] = [
    { title: '名称', dataIndex: 'name' },
    { title: '编码', dataIndex: 'code' },
    { title: '类型', dataIndex: 'type', valueType: 'select',
      valueEnum: { department: { text: '部门' }, project: { text: '项目' }, team: { text: '团队' } },
      render: (_: string) => <Tag>{typeLabels[_] || _}</Tag> },
    { title: '联系人', dataIndex: 'contact', hideInSearch: true },
    { title: '电话', dataIndex: 'phone', hideInSearch: true },
    { title: '邮箱', dataIndex: 'email', hideInSearch: true },
    { title: '创建时间', dataIndex: 'created_at', valueType: 'dateTime', hideInSearch: true },
    { title: '操作', valueType: 'option', render: (_: any, r: Organization) => [
      <a key="edit" onClick={() => { organizationApi.update(r.id, r).then(() => { message.success('更新成功'); actionRef.current?.reload() }) }}>编辑</a>,
      <a key="del" style={{ color: '#ff4d4f' }} onClick={() => { organizationApi.delete(r.id).then(() => { message.success('删除成功'); actionRef.current?.reload() }) }}>删除</a>,
    ]},
  ]

  return (
    <ProTable<Organization>
      headerTitle="组织管理"
      actionRef={actionRef}
      rowKey="id"
      columns={columns}
      request={async () => { const res = await organizationApi.list(); return { data: res.data.items || [], success: true } }}
      toolBarRender={() => [
        <ModalForm
          key="create"
          title="创建组织"
          trigger={<Button type="primary" icon={<PlusOutlined />}>创建组织</Button>}
          modalProps={{ destroyOnHidden: true }}
          onFinish={async (values: any) => {
            try { await organizationApi.create(values); message.success('创建成功'); actionRef.current?.reload(); return true }
            catch (err: any) { message.error(err.response?.data?.error || '创建失败'); return false }
          }}
        >
          <ProFormText name="name" label="组织名称" rules={[{ required: true }]} />
          <ProFormText name="code" label="组织编码" rules={[{ required: true }]} />
          <ProFormSelect name="type" label="组织类型" rules={[{ required: true }]} options={[{ label: '部门', value: 'department' }, { label: '项目', value: 'project' }, { label: '团队', value: 'team' }]} />
          <ProFormText name="contact" label="联系人" />
          <ProFormText name="phone" label="联系电话" />
          <ProFormText name="email" label="邮箱" />
          <ProFormTextArea name="description" label="描述" />
        </ModalForm>,
      ]}
    />
  )
}

export default Organizations
