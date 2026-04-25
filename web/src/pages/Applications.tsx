import React, { useRef, useState } from 'react'
import { ProTable, ModalForm, ProFormText, ProFormTextArea, ProFormSelect, ProFormSwitch } from '@ant-design/pro-components'
import { Button, Tag, message, Modal } from 'antd'
import { PlusOutlined, EditOutlined } from '@ant-design/icons'
import { applicationApi, Application } from '../api'

const Applications: React.FC = () => {
  const actionRef = useRef<any>()
  // 编辑模态框状态
  const [editVisible, setEditVisible] = useState(false)
  const [currentRecord, setCurrentRecord] = useState<Application | null>(null)

  // 打开编辑模态框
  const handleEdit = (record: Application) => {
    setCurrentRecord(record)
    setEditVisible(true)
  }

  const columns: any[] = [
    { title: '名称', dataIndex: 'name' },
    { title: '显示名称', dataIndex: 'display_name', hideInSearch: true },
    { title: 'Chart', dataIndex: 'chart_name', hideInSearch: true },
    { title: '仓库', dataIndex: 'chart_repo', hideInSearch: true, ellipsis: true },
    { title: '版本', dataIndex: 'chart_version', hideInSearch: true },
    { title: '分类', dataIndex: 'category', hideInSearch: true },
    { title: 'SaaS', dataIndex: 'is_saas', hideInSearch: true, render: (v: boolean) => v ? <Tag color="blue">是</Tag> : <Tag>否</Tag> },
    { title: '操作', valueType: 'option', render: (_: any, r: Application) => [
      <a key="edit" onClick={() => handleEdit(r)}><EditOutlined /> 编辑</a>,
      <a key="del" style={{ color: '#ff4d4f' }} onClick={() => { Modal.confirm({ title: '确认删除', content: `确定要删除应用 "${r.name}" 吗？此操作不可恢复。`, okType: 'danger', onOk: () => applicationApi.delete(r.id).then(() => { message.success('删除成功'); actionRef.current?.reload() }).catch((err: any) => { message.error(err.response?.data?.error || '操作失败') }) }) }}>删除</a>,
    ]},
  ]

  return (
    <>
      <ProTable<Application>
        headerTitle="应用市场"
        actionRef={actionRef}
        rowKey="id"
        columns={columns}
        request={async () => { const res = await applicationApi.list(); return { data: res.data.items || [], success: true } }}
        toolBarRender={() => [
          <ModalForm
            key="create"
            title="创建应用"
            trigger={<Button type="primary" icon={<PlusOutlined />}>创建应用</Button>}
            modalProps={{ destroyOnHidden: true }}
            onFinish={async (values: any) => {
              try { await applicationApi.create(values); message.success('创建成功'); actionRef.current?.reload(); return true }
              catch (err: any) { message.error(err.response?.data?.error || '创建失败'); return false }
            }}
          >
            <ProFormText name="name" label="应用名称" rules={[{ required: true }]} />
            <ProFormText name="display_name" label="显示名称" />
            <ProFormText name="chart_name" label="Chart名称" rules={[{ required: true }]} />
            <ProFormText name="chart_repo" label="Chart仓库" rules={[{ required: true }]} />
            <ProFormText name="chart_version" label="Chart版本" rules={[{ required: true }]} />
            <ProFormSelect name="category" label="分类" options={[{ label: '数据库', value: 'database' }, { label: '中间件', value: 'middleware' }, { label: '监控', value: 'monitoring' }, { label: '其他', value: 'other' }]} />
            <ProFormSwitch name="is_saas" label="SaaS应用" />
            <ProFormTextArea name="description" label="描述" />
          </ModalForm>,
        ]}
      />

      {/* 编辑应用模态框 */}
      <ModalForm
        title="编辑应用"
        open={editVisible}
        onOpenChange={setEditVisible}
        modalProps={{ destroyOnHidden: true }}
        initialValues={currentRecord ? {
          display_name: currentRecord.display_name,
          description: currentRecord.description,
          chart_name: currentRecord.chart_name,
          chart_repo: currentRecord.chart_repo,
          chart_version: currentRecord.chart_version,
          category: currentRecord.category,
          default_values: currentRecord.default_values,
        } : undefined}
        onFinish={async (values: any) => {
          if (!currentRecord) return false
          try {
            await applicationApi.update(currentRecord.id, values)
            message.success('更新成功')
            actionRef.current?.reload()
            return true
          } catch (err: any) {
            message.error(err.response?.data?.error || '更新失败')
            return false
          }
        }}
      >
        <ProFormText name="display_name" label="显示名称" />
        <ProFormText name="chart_name" label="Chart名称" rules={[{ required: true }]} />
        <ProFormText name="chart_repo" label="Chart仓库" rules={[{ required: true }]} />
        <ProFormText name="chart_version" label="Chart版本" rules={[{ required: true }]} />
        <ProFormSelect name="category" label="分类" options={[{ label: '数据库', value: 'database' }, { label: '中间件', value: 'middleware' }, { label: '监控', value: 'monitoring' }, { label: '其他', value: 'other' }]} />
        <ProFormTextArea name="default_values" label="默认Values" />
        <ProFormTextArea name="description" label="描述" />
      </ModalForm>
    </>
  )
}

export default Applications
