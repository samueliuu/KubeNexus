import React, { useRef, useState, useEffect } from 'react'
import { ProTable, ModalForm, ProFormText, ProFormTextArea, ProFormSelect } from '@ant-design/pro-components'
import { Button, message, Modal } from 'antd'
import { PlusOutlined, EditOutlined, EyeOutlined } from '@ant-design/icons'
import { configApi, applicationApi, organizationApi, ConfigTemplate, Application, Organization } from '../api'

const ConfigCenter: React.FC = () => {
  const actionRef = useRef<any>()
  // 应用和组织名称映射缓存
  const [appMap, setAppMap] = useState<Record<string, string>>({})
  const [orgMap, setOrgMap] = useState<Record<string, string>>({})
  // 查看详情弹窗
  const [detailVisible, setDetailVisible] = useState(false)
  const [detailRecord, setDetailRecord] = useState<ConfigTemplate | null>(null)
  // 编辑弹窗
  const [editVisible, setEditVisible] = useState(false)
  const [editRecord, setEditRecord] = useState<ConfigTemplate | null>(null)

  // 加载应用和组织映射
  useEffect(() => {
    applicationApi.list().then(res => {
      const map: Record<string, string> = {}
      ;(res.data.items || []).forEach((a: Application) => { map[a.id] = a.display_name || a.name })
      setAppMap(map)
    }).catch(() => {})
    organizationApi.list().then(res => {
      const map: Record<string, string> = {}
      ;(res.data.items || []).forEach((o: Organization) => { map[o.id] = o.name })
      setOrgMap(map)
    }).catch(() => {})
  }, [])

  const columns: any[] = [
    { title: '配置名称', dataIndex: 'name' },
    {
      title: '关联应用',
      dataIndex: 'application_id',
      hideInSearch: true,
      render: (_: string) => appMap[_] || _,
    },
    {
      title: '关联组织',
      dataIndex: 'org_id',
      hideInSearch: true,
      render: (_: string) => orgMap[_] || _,
    },
    { title: '描述', dataIndex: 'description', ellipsis: true, hideInSearch: true },
    { title: '创建时间', dataIndex: 'created_at', valueType: 'dateTime', hideInSearch: true },
    {
      title: '操作',
      valueType: 'option',
      render: (_: any, r: ConfigTemplate) => [
        <Button
          key="view"
          type="link"
          size="small"
          icon={<EyeOutlined />}
          onClick={() => { setDetailRecord(r); setDetailVisible(true) }}
        >
          查看
        </Button>,
        <Button
          key="edit"
          type="link"
          size="small"
          icon={<EditOutlined />}
          onClick={() => { setEditRecord(r); setEditVisible(true) }}
        >
          编辑
        </Button>,
        <a
          key="del"
          style={{ color: '#ff4d4f' }}
          onClick={() => {
            configApi.delete(r.id).then(() => { message.success('删除成功'); actionRef.current?.reload() })
          }}
        >
          删除
        </a>,
      ],
    },
  ]

  return (
    <>
      <ProTable<ConfigTemplate>
        headerTitle="配置中心"
        actionRef={actionRef}
        rowKey="id"
        columns={columns}
        request={async () => {
          const res = await configApi.list()
          return { data: res.data.items || [], success: true }
        }}
        toolBarRender={() => [
          <ModalForm
            key="create"
            title="创建配置模板"
            trigger={<Button type="primary" icon={<PlusOutlined />}>创建配置</Button>}
            modalProps={{ destroyOnHidden: true }}
            onFinish={async (values: any) => {
              try {
                await configApi.create(values)
                message.success('创建成功')
                actionRef.current?.reload()
                return true
              } catch (err: any) {
                message.error(err.response?.data?.error || '创建失败')
                return false
              }
            }}
          >
            <ProFormText name="name" label="配置名称" rules={[{ required: true }]} />
            <ProFormSelect name="application_id" label="关联应用" rules={[{ required: true }]}
              request={async () => {
                const res = await applicationApi.list()
                return (res.data.items || []).map((a: Application) => ({ label: a.display_name || a.name, value: a.id }))
              }} />
            <ProFormSelect name="org_id" label="关联组织"
              request={async () => {
                const res = await organizationApi.list()
                return (res.data.items || []).map((o: Organization) => ({ label: o.name, value: o.id }))
              }} />
            <ProFormTextArea name="values" label="配置内容 (YAML)" rules={[{ required: true }]} />
            <ProFormTextArea name="description" label="描述" />
          </ModalForm>,
        ]}
      />

      {/* 查看详情弹窗 */}
      <Modal
        title={`配置详情 - ${detailRecord?.name || ''}`}
        open={detailVisible}
        onCancel={() => { setDetailVisible(false); setDetailRecord(null) }}
        footer={null}
        width={640}
      >
        {detailRecord && (
          <div>
            <p><strong>配置名称：</strong>{detailRecord.name}</p>
            <p><strong>关联应用：</strong>{appMap[detailRecord.application_id] || detailRecord.application_id}</p>
            <p><strong>关联组织：</strong>{orgMap[detailRecord.org_id] || detailRecord.org_id}</p>
            <p><strong>描述：</strong>{detailRecord.description || '无'}</p>
            <p><strong>配置内容 (YAML)：</strong></p>
            <pre style={{
              background: '#f5f5f5',
              padding: 12,
              borderRadius: 6,
              maxHeight: 400,
              overflow: 'auto',
              fontSize: 13,
            }}>
              {detailRecord.values}
            </pre>
          </div>
        )}
      </Modal>

      {/* 编辑弹窗 */}
      <ModalForm
        title="编辑配置模板"
        open={editVisible}
        modalProps={{ destroyOnHidden: true }}
        initialValues={editRecord ? {
          name: editRecord.name,
          values: editRecord.values,
          description: editRecord.description,
        } : {}}
        onOpenChange={(visible) => {
          if (!visible) { setEditVisible(false); setEditRecord(null) }
        }}
        onFinish={async (values: any) => {
          if (!editRecord) return false
          try {
            await configApi.update(editRecord.id, values)
            message.success('更新成功')
            actionRef.current?.reload()
            return true
          } catch (err: any) {
            message.error(err.response?.data?.error || '更新失败')
            return false
          }
        }}
      >
        <ProFormText name="name" label="配置名称" rules={[{ required: true }]} />
        <ProFormTextArea name="values" label="配置内容 (YAML)" rules={[{ required: true }]} />
        <ProFormTextArea name="description" label="描述" />
      </ModalForm>
    </>
  )
}

export default ConfigCenter
