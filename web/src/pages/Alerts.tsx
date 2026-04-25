import React, { useRef, useState } from 'react'
import { ProTable, ModalForm, ProFormText, ProFormSelect, ProFormTextArea, ProCard } from '@ant-design/pro-components'
import { Button, Tag, Switch, message } from 'antd'
import { PlusOutlined, EditOutlined, CheckOutlined } from '@ant-design/icons'
import { alertApi, AlertRule, AlertRecord } from '../api'

const severityColors: Record<string, string> = { critical: 'red', warning: 'orange', info: 'blue' }
const severityLabels: Record<string, string> = { critical: '严重', warning: '警告', info: '信息' }

const Alerts: React.FC = () => {
  const ruleRef = useRef<any>()
  const recordRef = useRef<any>()
  // 编辑规则模态框状态
  const [editVisible, setEditVisible] = useState(false)
  const [currentRule, setCurrentRule] = useState<AlertRule | null>(null)

  // 打开编辑规则模态框
  const handleEditRule = (record: AlertRule) => {
    setCurrentRule(record)
    setEditVisible(true)
  }

  const ruleColumns: any[] = [
    { title: '规则名称', dataIndex: 'name' },
    { title: '类型', dataIndex: 'type', valueType: 'select',
      valueEnum: {
        cluster_down: { text: '集群离线' }, cpu_high: { text: 'CPU过高' },
        mem_high: { text: '内存过高' }, drift_detected: { text: '配置漂移' },
        license_expiring: { text: 'License即将过期' },
      } },
    { title: '严重级别', dataIndex: 'severity', valueType: 'select',
      valueEnum: { critical: { text: '严重' }, warning: { text: '警告' }, info: { text: '信息' } },
      render: (_: string) => <Tag color={severityColors[_]}>{severityLabels[_] || _}</Tag> },
    { title: '通知渠道', dataIndex: 'notify_channels', hideInSearch: true, ellipsis: true },
    { title: '启用', dataIndex: 'enabled', hideInSearch: true,
      render: (_: boolean, r: AlertRule) => <Switch checked={_} size="small" onChange={(c) => { alertApi.updateRule(r.id, { enabled: c }).then(() => ruleRef.current?.reload()) }} /> },
    { title: '最后触发', dataIndex: 'last_triggered', valueType: 'dateTime', hideInSearch: true },
    { title: '操作', valueType: 'option', render: (_: any, r: AlertRule) => [
      <a key="edit" onClick={() => handleEditRule(r)}><EditOutlined /> 编辑</a>,
      <a key="del" style={{ color: '#ff4d4f' }} onClick={() => { alertApi.deleteRule(r.id).then(() => { message.success('删除成功'); ruleRef.current?.reload() }) }}>删除</a>,
    ]},
  ]

  const recordColumns: any[] = [
    { title: '规则', dataIndex: 'rule_name' },
    { title: '集群', dataIndex: 'cluster_id', hideInSearch: true },
    { title: '严重级别', dataIndex: 'severity', hideInSearch: true, render: (_: string) => <Tag color={severityColors[_]}>{severityLabels[_] || _}</Tag> },
    { title: '消息', dataIndex: 'message', ellipsis: true, hideInSearch: true },
    { title: '状态', dataIndex: 'status', valueType: 'select',
      valueEnum: { firing: { text: '触发中', status: 'Error' }, resolved: { text: '已恢复', status: 'Success' } },
      render: (_: string) => <Tag color={_ === 'firing' ? 'red' : 'green'}>{_ === 'firing' ? '触发中' : '已恢复'}</Tag> },
    { title: '触发时间', dataIndex: 'triggered_at', valueType: 'dateTime', hideInSearch: true },
    { title: '操作', valueType: 'option', render: (_: any, r: AlertRecord) => [
      // 仅 firing 状态显示确认按钮
      r.status === 'firing' ? <a key="ack" onClick={() => { alertApi.acknowledgeRecord(r.id).then(() => { message.success('已确认'); recordRef.current?.reload() }).catch((err: any) => { message.error(err.response?.data?.error || '确认失败') }) }}><CheckOutlined /> 确认</a> : null,
    ].filter(Boolean)},
  ]

  // 规则表单字段（创建和编辑共用）
  const ruleFormFields = (
    <>
      <ProFormText name="name" label="规则名称" rules={[{ required: true }]} />
      <ProFormSelect name="type" label="类型" rules={[{ required: true }]} options={[
        { label: '集群离线', value: 'cluster_down' }, { label: 'CPU过高', value: 'cpu_high' },
        { label: '内存过高', value: 'mem_high' }, { label: '配置漂移', value: 'drift_detected' },
        { label: 'License即将过期', value: 'license_expiring' },
      ]} />
      <ProFormTextArea name="condition" label="条件 (JSON)" rules={[{ required: true }]} />
      <ProFormSelect name="severity" label="严重级别" options={[{ label: '严重', value: 'critical' }, { label: '警告', value: 'warning' }, { label: '信息', value: 'info' }]} />
      <ProFormText name="notify_channels" label="通知渠道" placeholder="例如: email,slack,webhook" />
    </>
  )

  return (
    <div>
      <ProCard title="告警规则" headerBordered style={{ marginBottom: 24 }}>
        <ProTable<AlertRule>
          actionRef={ruleRef}
          rowKey="id"
          search={false}
          columns={ruleColumns}
          request={async () => { const res = await alertApi.listRules(); return { data: res.data.items || [], success: true } }}
          toolBarRender={() => [
            <ModalForm
              key="create"
              title="添加告警规则"
              trigger={<Button type="primary" icon={<PlusOutlined />}>添加规则</Button>}
              modalProps={{ destroyOnHidden: true }}
              onFinish={async (values: any) => {
                try { await alertApi.createRule(values); message.success('创建成功'); ruleRef.current?.reload(); return true }
                catch (err: any) { message.error(err.response?.data?.error || '创建失败'); return false }
              }}
            >
              {ruleFormFields}
            </ModalForm>,
          ]}
        />
      </ProCard>

      {/* 编辑告警规则模态框 */}
      <ModalForm
        title="编辑告警规则"
        open={editVisible}
        onOpenChange={setEditVisible}
        modalProps={{ destroyOnHidden: true }}
        initialValues={currentRule ? {
          name: currentRule.name,
          type: currentRule.type,
          condition: currentRule.condition,
          severity: currentRule.severity,
          notify_channels: currentRule.notify_channels,
        } : undefined}
        onFinish={async (values: any) => {
          if (!currentRule) return false
          try {
            await alertApi.updateRule(currentRule.id, values)
            message.success('更新成功')
            ruleRef.current?.reload()
            return true
          } catch (err: any) {
            message.error(err.response?.data?.error || '更新失败')
            return false
          }
        }}
      >
        {ruleFormFields}
      </ModalForm>

      <ProCard title="告警记录" headerBordered>
        <ProTable<AlertRecord>
          actionRef={recordRef}
          rowKey="id"
          columns={recordColumns}
          request={async (params: any) => {
            const res = await alertApi.listRecords({ status: params?.status })
            return { data: res.data.items || [], success: true }
          }}
        />
      </ProCard>
    </div>
  )
}

export default Alerts
