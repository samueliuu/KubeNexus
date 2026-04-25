# 衡牧KubeNexusK3s多集群管理系统 - 开发指南

## 项目结构说明

```
KubeNexus/
├── server/                          # 后端服务
│   ├── cmd/
│   │   ├── server/main.go           # 服务入口，初始化各组件并启动 HTTP 服务
│   │   └── hashpw/main.go           # 密码哈希工具
│   └── internal/
│       ├── api/                     # API 层：HTTP 请求处理与路由注册
│       │   └── api.go
│       ├── middleware/              # 中间件：JWT 认证、权限校验、Agent Token 认证
│       │   └── auth.go
│       ├── model/                   # 数据模型：GORM 模型定义
│       │   └── model.go
│       ├── service/                 # 业务逻辑层：核心业务处理
│       │   ├── alert_engine.go      # 告警引擎
│       │   ├── application.go       # 应用管理
│       │   ├── auth.go              # 认证与密码工具
│       │   ├── cluster.go           # 集群管理
│       │   └── services.go          # 其他业务服务
│       ├── store/                   # 数据访问层：数据库操作封装
│       │   └── store.go
│       └── tunnel/                  # WebSocket 隧道：Agent 连接管理与请求代理
│           └── tunnel.go
├── web/                             # 前端应用
│   └── src/
│       ├── api/                     # API 调用封装：axios 实例与各模块接口定义
│       │   └── index.ts
│       ├── components/              # 公共组件
│       │   └── AppLayout.tsx        # 应用布局（侧边栏、顶栏）
│       ├── contexts/                # React Context
│       │   └── AuthContext.tsx      # 认证状态管理
│       ├── pages/                   # 页面组件
│       │   ├── Alerts.tsx           # 告警管理
│       │   ├── Applications.tsx     # 应用商店
│       │   ├── ClusterDetail.tsx    # 集群详情
│       │   ├── Clusters.tsx         # 集群列表
│       │   ├── ConfigCenter.tsx     # 配置中心
│       │   ├── Dashboard.tsx        # 仪表盘
│       │   ├── Deployments.tsx      # 部署管理
│       │   ├── Login.tsx            # 登录页
│       │   ├── Organizations.tsx    # 组织管理
│       │   └── Settings.tsx         # 系统设置
│       ├── App.tsx                  # 路由配置
│       ├── main.tsx                 # 应用入口
│       └── index.css                # 全局样式
├── agent/                           # Agent 端
│   ├── cmd/agent/main.go            # Agent 入口
│   └── internal/agent/agent.go      # Agent 核心逻辑：心跳上报、状态同步、Helm 操作
├── charts/                          # Helm Charts
│   ├── kubenexus-agent/             # Agent Helm Chart
│   └── saas-app/                    # SaaS 应用 Helm Chart
├── scripts/
│   └── install-k3s.sh               # K3s + Agent 一键安装脚本
└── docs/                            # 项目文档
```

## 后端开发规范

### 分层架构

后端严格遵循三层架构，请求处理流程为：

```
HTTP 请求 → API 层 → Service 层 → Store 层 → 数据库
```

各层职责：

| 层级 | 目录 | 职责 |
|------|------|------|
| API 层 | `internal/api/` | 参数绑定与校验、调用 Service、返回 HTTP 响应、记录审计日志 |
| Service 层 | `internal/service/` | 业务逻辑处理、数据组装、跨模块协调 |
| Store 层 | `internal/store/` | 数据库 CRUD 操作、GORM 查询封装 |

**禁止跨层调用**：API 层不得直接调用 Store 层，必须通过 Service 层中转。

### API 层规范

1. 使用 `c.ShouldBindJSON()` 绑定请求参数
2. 参数校验失败返回 `400 Bad Request`
3. 业务错误返回对应状态码和 `{"error": "错误描述"}`
4. 写操作必须记录审计日志：`h.auditLog(c, action, resourceType, resourceID, resourceName)`
5. 需要认证的路由使用 `middleware.AuthMiddleware()`
6. 需要管理员权限的路由追加 `middleware.AdminMiddleware()`
7. Agent 专用路由使用 `middleware.AgentAuthMiddleware(h.store)`

路由注册示例：

```go
clusters := api.Group("/clusters")
{
    clusters.POST("", middleware.AuthMiddleware(), h.RegisterCluster)
    clusters.GET("", middleware.AuthMiddleware(), h.ListClusters)
    clusters.DELETE("/:id", middleware.AuthMiddleware(), middleware.AdminMiddleware(), h.DeleteCluster)
}
```

### Service 层规范

1. 每个 Service 对应一个业务领域（如 `ClusterService`、`ApplicationService`）
2. Service 通过构造函数注入 Store 依赖
3. 业务错误通过 `error` 返回，由 API 层决定 HTTP 状态码
4. 复杂业务逻辑拆分为私有方法，公开方法保持简洁

### Store 层规范

1. 每个 Model 对应一组 CRUD 方法
2. 使用 GORM 链式调用构建查询
3. 软删除使用 `gorm.DeletedAt`
4. 列表查询支持过滤条件（如 `clusterID`、`status`）
5. 分页使用 `Limit` + `Offset`

### 错误处理

1. API 层统一使用 `gin.H{"error": "描述"}` 返回错误
2. Service 层返回 `error`，不直接操作 HTTP 响应
3. Store 层直接返回 GORM 错误，由上层判断处理
4. 关键错误使用 `log.Printf` 记录日志

### 审计日志

所有写操作（创建、更新、删除、部署、Token 轮换等）必须记录审计日志：

```go
h.auditLog(c, "create", "cluster", cluster.ID, cluster.Name)
```

审计日志字段：用户 ID、用户名、操作类型、资源类型、资源 ID、资源名称、客户端 IP。

## 前端开发规范

### ProComponents 使用

项目使用 Ant Design ProComponents 构建界面，常用组件：

| 组件 | 用途 |
|------|------|
| `ProTable` | 数据列表页，内置搜索、分页、工具栏 |
| `ProForm` | 表单，支持多种字段类型 |
| `ProCard` | 卡片容器 |
| `ProLayout` | 页面布局（通过 `AppLayout` 封装） |

### ModalForm 使用规范

使用 `ModalForm` 时必须设置 `destroyOnHidden` 属性，确保弹窗关闭后表单状态重置：

```tsx
<ModalForm
  title="新建资源"
  destroyOnHidden
  trigger={<Button type="primary">新建</Button>}
  onFinish={async (values) => {
    await someApi.create(values)
    return true
  }}
>
  {/* 表单字段 */}
</ModalForm>
```

### API 调用与错误处理

1. 所有 API 调用使用 `src/api/index.ts` 中封装的接口
2. axios 拦截器已处理 401 自动跳转登录页
3. 业务错误在组件中通过 `message.error()` 提示用户

标准调用模式：

```tsx
const fetchData = async () => {
  try {
    const { data } = await someApi.list()
    setDataSource(data.items || [])
  } catch (error: any) {
    message.error(error.response?.data?.error || '请求失败')
  }
}
```

### 认证与路由

1. 使用 `AuthContext` 获取认证状态和用户信息
2. 受保护路由通过 `ProtectedRoute` 组件包裹，未登录自动跳转 `/login`
3. Token 存储在 `localStorage`，通过 axios 拦截器自动附加到请求头

## 数据库迁移

系统使用 GORM `AutoMigrate` 自动管理数据库表结构。每次服务启动时，`store.New()` 会自动执行迁移：

```go
db.AutoMigrate(
    &model.Cluster{},
    &model.Application{},
    &model.Deployment{},
    &model.Organization{},
    &model.License{},
    &model.AlertRule{},
    &model.AlertRecord{},
    &model.ConfigTemplate{},
    &model.AuditLog{},
    &model.Heartbeat{},
    &model.User{},
)
```

### 添加新模型

1. 在 `internal/model/model.go` 中定义结构体，添加 GORM 标签
2. 在 `store.New()` 的 `AutoMigrate` 调用中添加新模型
3. 在 `internal/store/store.go` 中添加对应的 CRUD 方法
4. 重启服务，GORM 会自动创建表

注意事项：

- `AutoMigrate` 只会添加缺失的字段和索引，不会删除或修改已有列
- 新增字段必须设置合理的默认值或允许为空
- 主键使用 `size:64` 的字符串类型，由代码生成 UUID
- 需要软删除的模型包含 `gorm.DeletedAt` 字段

## 添加新 API 的步骤

以添加"通知渠道"功能为例：

### 1. 定义数据模型

在 `internal/model/model.go` 中添加：

```go
type NotifyChannel struct {
    ID        string         `gorm:"primaryKey;size:64" json:"id"`
    Name      string         `gorm:"size:128" json:"name"`
    Type      string         `gorm:"size:64" json:"type"`
    Config    string         `gorm:"type:text" json:"config"`
    Enabled   bool           `gorm:"default:1" json:"enabled"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
```

### 2. 添加 Store 方法

在 `internal/store/store.go` 中添加 CRUD 方法：

```go
func (s *Store) CreateNotifyChannel(n *model.NotifyChannel) error {
    return s.DB.Create(n).Error
}

func (s *Store) ListNotifyChannels() ([]model.NotifyChannel, error) {
    var channels []model.NotifyChannel
    if err := s.DB.Find(&channels).Error; err != nil {
        return nil, err
    }
    return channels, nil
}
```

### 3. 添加 Service

在 `internal/service/` 下创建或扩展 Service：

```go
type NotifyChannelService struct {
    store *store.Store
}

func NewNotifyChannelService(s *store.Store) *NotifyChannelService {
    return &NotifyChannelService{store: s}
}
```

### 4. 添加 API Handler

在 `internal/api/api.go` 中：

- 在 `Handler` 结构体中添加 Service 字段
- 在 `NewHandler` 中注入 Service
- 在 `RegisterRoutes` 中注册路由
- 实现处理方法

### 5. 注册 AutoMigrate

在 `store.New()` 中添加 `&model.NotifyChannel{}`。

### 6. 在 main.go 中初始化

在 `cmd/server/main.go` 中创建 Service 实例并传入 Handler。

### 7. 前端对接

在 `web/src/api/index.ts` 中添加接口定义和 API 调用方法。

## 添加新页面的步骤

以添加"通知渠道"页面为例：

### 1. 定义 API 接口

在 `web/src/api/index.ts` 中添加类型和 API 方法：

```typescript
export interface NotifyChannel {
  id: string
  name: string
  type: string
  config: string
  enabled: boolean
  created_at: string
}

export const notifyChannelApi = {
  list: () => api.get<{ items: NotifyChannel[] }>('/notify-channels'),
  create: (data: any) => api.post('/notify-channels', data),
  update: (id: string, data: any) => api.put(`/notify-channels/${id}`, data),
  delete: (id: string) => api.delete(`/notify-channels/${id}`),
}
```

### 2. 创建页面组件

在 `web/src/pages/` 下创建 `NotifyChannels.tsx`：

```tsx
import React, { useEffect, useState } from 'react'
import { ProTable } from '@ant-design/pro-components'
import { Button, message, ModalForm, ProFormText, ProFormSelect } from '@ant-design/pro-components'
import { notifyChannelApi, type NotifyChannel } from '@/api'

const NotifyChannels: React.FC = () => {
  const [data, setData] = useState<NotifyChannel[]>([])

  const fetchData = async () => {
    try {
      const { data } = await notifyChannelApi.list()
      setData(data.items || [])
    } catch (error: any) {
      message.error(error.response?.data?.error || '获取数据失败')
    }
  }

  useEffect(() => { fetchData() }, [])

  return (
    <ProTable<NotifyChannel>
      headerTitle="通知渠道"
      dataSource={data}
      rowKey="id"
      search={false}
      toolBarRender={() => [
        <ModalForm
          key="create"
          title="新建通知渠道"
          destroyOnHidden
          trigger={<Button type="primary">新建</Button>}
          onFinish={async (values) => {
            await notifyChannelApi.create(values)
            message.success('创建成功')
            fetchData()
            return true
          }}
        >
          <ProFormText name="name" label="名称" rules={[{ required: true }]} />
          <ProFormSelect name="type" label="类型" options={[
            { label: '钉钉', value: 'dingtalk' },
            { label: '企业微信', value: 'wecom' },
          ]} />
        </ModalForm>,
      ]}
      columns={[
        { title: '名称', dataIndex: 'name' },
        { title: '类型', dataIndex: 'type' },
        { title: '状态', dataIndex: 'enabled' },
      ]}
    />
  )
}

export default NotifyChannels
```

### 3. 注册路由

在 `web/src/App.tsx` 中添加路由：

```tsx
import NotifyChannels from './pages/NotifyChannels'

// 在路由配置中添加
<Route path="notify-channels" element={<NotifyChannels />} />
```

### 4. 添加导航菜单

在 `web/src/components/AppLayout.tsx` 中的菜单配置添加对应条目。

## 代码风格

### Go 代码

- 使用 `gofmt` 格式化代码
- 使用中文注释
- 导入分组顺序：标准库 → 第三方库 → 项目内部包，各组之间空行分隔
- 错误处理不使用 panic，使用 `log.Fatalf` 处理启动阶段致命错误，运行时错误通过 `error` 返回
- 变量命名使用驼峰式，导出使用大驼峰，内部使用小驼峰
- 常量使用驼峰式，不使用全大写

### TypeScript 代码

- 使用 ESLint 规范代码风格
- 使用中文注释
- 使用 `interface` 定义对象类型
- 优先使用 `const`，其次 `let`，禁止 `var`
- 使用可选链 `?.` 和空值合并 `??`
- 组件使用函数式写法 + Hooks
- 导入使用路径别名 `@/` 代替相对路径
- API 响应类型在 `api/index.ts` 中统一定义
