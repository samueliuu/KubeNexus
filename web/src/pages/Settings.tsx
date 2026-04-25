import React, { useRef, useState, useEffect } from 'react'
import { ProCard, ProTable, ModalForm, ProFormText, ProFormDigit, ProFormSelect } from '@ant-design/pro-components'
import { Button, Descriptions, Tag, message, Progress } from 'antd'
import { PlusOutlined, EditOutlined } from '@ant-design/icons'
import { licenseApi, userApi, auditApi, License, User, AuditLog, LicenseQuota } from '../api'

const Settings: React.FC = () => {
  const userRef = useRef<any>()
  const auditRef = useRef<any>()
  const [license, setLicense] = useState<License | null>(null)
  const [quota, setQuota] = useState<LicenseQuota | null>(null)
  // 编辑用户弹窗
  const [editUserVisible, setEditUserVisible] = useState(false)
  const [editUserRecord, setEditUserRecord] = useState<User | null>(null)

  useEffect(() => {
    licenseApi.get().then(res => setLicense(res.data)).catch(() => {})
    licenseApi.getQuota().then(res => setQuota(res.data)).catch(() => {})
  }, [])

  // 计算配额使用百分比
  const getPercent = (current: number, max: number) => {
    if (max <= 0) return 0
    return Math.min(Math.round((current / max) * 100), 100)
  }

  // 配额使用状态颜色
  const getStatus = (percent: number): 'success' | 'normal' | 'exception' => {
    if (percent >= 90) return 'exception'
    if (percent >= 70) return 'normal'
    return 'success'
  }

  const userColumns: any[] = [
    { title: '用户名', dataIndex: 'username' },
    {
      title: '角色',
      dataIndex: 'role',
      valueType: 'select',
      valueEnum: { admin: { text: '管理员' }, operator: { text: '操作员' }, viewer: { text: '只读' } },
      render: (_: string) => <Tag color={_ === 'admin' ? 'red' : _ === 'operator' ? 'blue' : 'default'}>{_}</Tag>,
    },
    { title: '创建时间', dataIndex: 'created_at', valueType: 'dateTime', hideInSearch: true },
    {
      title: '操作',
      valueType: 'option',
      render: (_: any, r: User) => r.username !== 'admin' ? [
        <Button
          key="edit"
          type="link"
          size="small"
          icon={<EditOutlined />}
          onClick={() => { setEditUserRecord(r); setEditUserVisible(true) }}
        >
          编辑
        </Button>,
        <a
          key="del"
          style={{ color: '#ff4d4f' }}
          onClick={() => {
            userApi.delete(r.id).then(() => { message.success('删除成功'); userRef.current?.reload() })
          }}
        >
          删除
        </a>,
      ] : [],
    },
  ]

  const auditColumns: any[] = [
    { title: '时间', dataIndex: 'created_at', valueType: 'dateTime', width: 180, hideInSearch: true },
    { title: '用户', dataIndex: 'username', width: 100 },
    { title: '操作', dataIndex: 'action', width: 120,
      valueEnum: {
        create: { text: '创建' },
        update: { text: '更新' },
        delete: { text: '删除' },
        login: { text: '登录' },
        logout: { text: '登出' },
      },
    },
    { title: '资源类型', dataIndex: 'resource_type', width: 100, hideInSearch: true },
    { title: '资源名称', dataIndex: 'resource_name', width: 150, hideInSearch: true },
    { title: 'IP', dataIndex: 'ip', width: 130, hideInSearch: true },
  ]

  return (
    <div>
      <ProCard title="License 管理" headerBordered style={{ marginBottom: 24 }}
        extra={<ModalForm
          title="激活 License"
          trigger={<Button type="primary">激活 License</Button>}
          modalProps={{ destroyOnHidden: true }}
          onFinish={async (values: any) => {
            try {
              await licenseApi.activate(values)
              message.success('激活成功')
              licenseApi.get().then(res => setLicense(res.data))
              licenseApi.getQuota().then(res => setQuota(res.data))
              return true
            } catch (err: any) {
              message.error(err.response?.data?.error || '激活失败')
              return false
            }
          }}
        >
          <ProFormText name="key" label="License Key" rules={[{ required: true }]} />
          <ProFormText name="customer_name" label="客户名称" />
          <ProFormDigit name="max_clusters" label="最大集群数" min={1} />
          <ProFormDigit name="max_deployments" label="最大部署数" min={1} />
        </ModalForm>}
      >
        {license && (
          <Descriptions column={2}>
            <Descriptions.Item label="状态">{license.is_valid ? <Tag color="green">有效</Tag> : <Tag color="red">无效</Tag>}</Descriptions.Item>
            <Descriptions.Item label="客户名称">{license.customer_name}</Descriptions.Item>
            <Descriptions.Item label="最大集群数">{license.max_clusters}</Descriptions.Item>
            <Descriptions.Item label="最大部署数">{license.max_deployments}</Descriptions.Item>
            <Descriptions.Item label="签发时间">{new Date(license.issued_at).toLocaleDateString()}</Descriptions.Item>
            <Descriptions.Item label="过期时间">{new Date(license.expires_at).toLocaleDateString()}</Descriptions.Item>
          </Descriptions>
        )}

        {/* 配额使用量展示 */}
        {quota && (
          <div style={{ marginTop: 16, padding: '16px 0 0', borderTop: '1px solid #f0f0f0' }}>
            <div style={{ marginBottom: 16 }}>
              <div style={{ marginBottom: 4, display: 'flex', justifyContent: 'space-between' }}>
                <span>集群配额</span>
                <span>{quota.clusters.current} / {quota.clusters.max}</span>
              </div>
              <Progress
                percent={getPercent(quota.clusters.current, quota.clusters.max)}
                status={getStatus(getPercent(quota.clusters.current, quota.clusters.max))}
              />
            </div>
            <div>
              <div style={{ marginBottom: 4, display: 'flex', justifyContent: 'space-between' }}>
                <span>部署配额</span>
                <span>{quota.deployments.current} / {quota.deployments.max}</span>
              </div>
              <Progress
                percent={getPercent(quota.deployments.current, quota.deployments.max)}
                status={getStatus(getPercent(quota.deployments.current, quota.deployments.max))}
              />
            </div>
          </div>
        )}
      </ProCard>

      <ProCard title="用户管理" headerBordered style={{ marginBottom: 24 }}
        extra={<ModalForm
          title="创建用户"
          trigger={<Button type="primary" icon={<PlusOutlined />}>创建用户</Button>}
          modalProps={{ destroyOnHidden: true }}
          onFinish={async (values: any) => {
            try {
              await userApi.create(values)
              message.success('创建成功')
              userRef.current?.reload()
              return true
            } catch (err: any) {
              message.error(err.response?.data?.error || '创建失败')
              return false
            }
          }}
        >
          <ProFormText name="username" label="用户名" rules={[{ required: true }]} />
          <ProFormText.Password name="password" label="密码" rules={[{ required: true }]} />
          <ProFormSelect name="role" label="角色" options={[
            { label: '管理员', value: 'admin' },
            { label: '操作员', value: 'operator' },
            { label: '只读', value: 'viewer' },
          ]} />
        </ModalForm>}
      >
        <ProTable<User>
          actionRef={userRef}
          rowKey="id"
          search={false}
          columns={userColumns}
          request={async () => {
            const res = await userApi.list()
            return { data: res.data.items || [], success: true }
          }}
        />
      </ProCard>

      {/* 编辑用户弹窗 */}
      <ModalForm
        title="编辑用户"
        open={editUserVisible}
        modalProps={{ destroyOnHidden: true }}
        initialValues={editUserRecord ? { role: editUserRecord.role } : {}}
        onOpenChange={(visible) => {
          if (!visible) { setEditUserVisible(false); setEditUserRecord(null) }
        }}
        onFinish={async (values: any) => {
          if (!editUserRecord) return false
          try {
            await userApi.update(editUserRecord.id, values)
            message.success('更新成功')
            userRef.current?.reload()
            return true
          } catch (err: any) {
            message.error(err.response?.data?.error || '更新失败')
            return false
          }
        }}
      >
        <ProFormSelect name="role" label="角色" rules={[{ required: true }]} options={[
          { label: '管理员', value: 'admin' },
          { label: '操作员', value: 'operator' },
          { label: '只读', value: 'viewer' },
        ]} />
        <ProFormText.Password name="password" label="新密码" placeholder="不修改请留空" />
      </ModalForm>

      <ProCard title="审计日志" headerBordered>
        <ProTable<AuditLog>
          actionRef={auditRef}
          rowKey="id"
          columns={auditColumns}
          request={async (params: any) => {
            const res = await auditApi.list({
              username: params.username || undefined,
              action: params.action || undefined,
            })
            return { data: res.data.items || [], success: true }
          }}
        />
      </ProCard>
    </div>
  )
}

export default Settings
