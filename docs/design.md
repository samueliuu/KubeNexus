# 衡牧KubeNexusK3s多集群管理系统 - 设计文档

> 架构参考：Rancher 多集群管理 + Fleet GitOps 应用交付
> 定位：面向中小企业的轻量 Rancher，专攻 K3s 场景
> 部署模式：**整体部署到客户本地**，单客户独占，非云端 SaaS

## 1. 项目背景与定位

### 1.1 背景

作为 SaaS 服务提供商，部分中小企业客户因数据安全、合规要求或网络隔离等原因，
需要将 SaaS 应用本地化部署到自己的服务器上。同时，这些客户自身也有业务容器化的需求。

### 1.2 部署模式

**KubeNexus 整体部署到客户本地环境**，而非供应商云端托管：

```
┌─────────────────────────────────────────────────────────┐
│              客户本地环境                                │
│                                                         │
│  ┌───────────────────────────────────────────────────┐  │
│  │          KubeNexus 管控面 (本地部署)               │  │
│  │                                                   │  │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌────────┐  │  │
│  │  │集群管理 │ │应用市场 │ │部署编排 │ │监控告警│  │  │
│  │  └─────────┘ └─────────┘ └─────────┘ └────────┘  │  │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌────────┐  │  │
│  │  │组织项目 │ │配置中心 │ │审计日志 │ │License │  │  │
│  │  └─────────┘ └─────────┘ └─────────┘ └────────┘  │  │
│  │                                                   │  │
│  │  ┌────────────┐ ┌──────────────┐                  │  │
│  │  │  SQLite/   │ │  Helm Chart  │                  │  │
│  │  │ PostgreSQL │ │  Registry    │                  │  │
│  │  └────────────┘ └──────────────┘                  │  │
│  └───────────────────────┬───────────────────────────┘  │
│                          │ HTTPS + WebSocket             │
│              ┌───────────┼───────────┐                   │
│              │           │           │                   │
│         ┌────┴────┐ ┌────┴────┐ ┌────┴────┐             │
│         │ 生产集群 │ │ 测试集群 │ │ 边缘集群 │             │
│         │ (K3s)   │ │ (K3s)   │ │ (K3s)   │             │
│         │ Agent   │ │ Agent   │ │ Agent   │             │
│         └─────────┘ └─────────┘ └─────────┘             │
└─────────────────────────────────────────────────────────┘
```

**关键区别**：
- 不需要跨企业多租户隔离（单客户独占整个平台）
- "客户管理"调整为"组织/项目管理"（管理客户内部的部门、项目、团队）
- License 控制本实例的配额（最大集群数、功能开关等）
- 数据不出客户网络，天然满足数据安全要求

### 1.3 与 Rancher 的对比

| Rancher 能力 | KubeNexus 借鉴 | 简化点 |
|-------------|----------------|--------|
| cattle-cluster-agent + WebSocket 隧道 | 同架构 Agent + WebSocket | 去掉 node-agent，合并为一个 Agent |
| Fleet GitOps 应用交付 | 声明式部署 + Drift 检测 | MVP 不走 Git，直接 API 触发 |
| 三级 RBAC (Global/Cluster/Project) | 两级权限 (管理员/操作员) | 不做 Project 细分 |
| 多种集群类型 (RKE2/EKS/AKS/GKE) | 仅支持 K3s | 大幅简化集群抽象 |
| etcd + CRD 存储 | SQLite/PostgreSQL + REST API | 不嵌入 etcd，运维更简单 |
| 多认证后端 (AD/LDAP/SAML/OIDC) | JWT + 可选 LDAP | MVP 仅 JWT，预留扩展 |
| 云端托管多租户 | 本地部署单客户独占 | 去掉跨企业隔离 |

---

## 2. 系统架构

### 2.1 核心设计原则

| 原则 | Rancher 做法 | KubeNexus 做法 |
|------|-------------|----------------|
| **Agent 主动出站** | cattle-cluster-agent 通过 WebSocket 回连 | 同理，Agent 主动建立 WebSocket 连接 |
| **声明式编排** | CRD + Controller 调谐 | Deployment 作为期望状态，Agent 持续调谐 |
| **隧道代理** | WebSocket 隧道代理 K8s API 请求 | 同理，支持远程 kubectl exec/logs |
| **本地自治** | Agent 断网后集群正常运行 | 同理，Agent 本地缓存期望状态 |
| **Drift 检测** | Fleet 检测实际 vs 期望偏差 | Agent 定期巡检，发现偏差自动纠正/告警 |

### 2.2 通信架构

```
┌─────────────┐                         ┌─────────────┐
│ API Server  │                         │   Agent     │
│            │◄─── WebSocket 连接 ──────┤ (主动建立)   │
│            │                         │            │
│            │◄─── 心跳 + 状态上报 ─────┤ 定时推送    │
│            │                         │            │
│            │──── 任务指令下发 ────────►│ 实时接收    │
│            │                         │            │
│            │──── K8s API 代理请求 ───►│ 转发到本地  │
│            │◄─── K8s API 代理响应 ────┤ K8s API    │
└─────────────┘                         └─────────────┘
```

MVP 同时支持 HTTP 心跳 + 任务轮询（兜底），WebSocket 作为增强通道。

---

## 3. 核心模块设计

### 3.1 模块总览

| 模块 | 职责 | 优先级 | 状态 |
|------|------|--------|------|
| **集群管理** | 注册、状态监控、K3s 安装引导 | P0 | ✅ 已实现 |
| **Agent** | WebSocket 回连、心跳、任务执行、指标采集 | P0 | ✅ 已实现(HTTP) |
| **应用市场** | 应用模板管理、Helm Chart 仓库 | P0 | ✅ 已实现 |
| **部署编排** | 声明式部署、Drift 检测、调谐 | P0 | ✅ 已实现 |
| **组织项目** | 部门/项目管理、资源归属 | P0 | ⚠️ 需调整(原客户管理) |
| **License** | 配额控制、功能开关、过期管理 | P0 | ❌ 待实现 |
| **监控告警** | 资源监控、应用健康、告警通知 | P1 | ❌ 待实现 |
| **配置中心** | 按组织维度的 Values 管理 | P1 | ❌ 待实现 |
| **权限管理** | JWT 认证、角色权限 | P1 | ✅ 已实现(基础) |
| **隧道代理** | 远程 kubectl exec/logs/port-forward | P2 | ⚠️ 框架已搭 |
| **审计日志** | 操作记录、变更追踪 | P2 | ❌ 待实现 |
| **日志聚合** | 集中查看各集群应用日志 | P2 | ❌ 待实现 |

### 3.2 集群管理

#### 3.2.1 集群注册流程

**方式一：一键脚本安装（推荐，适合新集群）**
```
运营方在控制台创建集群 → 生成注册 Token 和安装命令
     ↓
客户在目标服务器执行安装脚本
     ├── 脚本安装 K3s (如果未安装)
     ├── 脚本部署 Agent (以 K3s Pod 方式运行)
     └── Agent 携带 Token 自动注册
     ↓
API Server 验证 Token → Agent 建立 WebSocket → 状态变为 active
```

**方式二：导入已有 K3s 集群**
```
运营方在控制台创建集群 → 生成注册 YAML
     ↓
客户在已有 K3s 集群上执行 kubectl apply -f registration.yaml
     ↓
Agent 启动 → 携带 Token 注册 → 建立 WebSocket → 状态变为 active
```

#### 3.2.2 集群状态机

```
  registered ──→ provisioning ──→ active
       ↑             │              │
       │             └──→ error     │
       │                            ├──→ unavailable (心跳超时 > 90s)
       │                            └──→ degraded (部分节点异常)
       └────────────────────────────┘
                (删除) → removed
```

#### 3.2.3 集群数据模型

```
Cluster {
  id              string    // 唯一标识
  name            string    // 集群名称 (唯一)
  display_name    string    // 显示名称
  status          string    // registered/provisioning/active/unavailable/degraded/error
  token           string    // 注册令牌 (敏感，注册后可轮换)
  endpoint        string    // K3s API 端点 (可选)
  version         string    // K3s 版本
  node_count      int       // 节点数
  cpu_capacity    string    // CPU 总量
  mem_capacity    string    // 内存总量
  labels          map       // 集群标签 (用于 clusterSelector 批量选择)
  region          string    // 地域
  org_id          string    // 所属组织/项目
  org_name        string    // 组织/项目名称
  last_heartbeat  time      // 最后心跳时间
  ws_connected    bool      // WebSocket 是否连接
}
```

### 3.3 Agent 设计

#### 3.3.1 Agent 架构

```
┌─────────────────────────────────────────────┐
│             KubeNexus Agent                  │
│          (K3s Pod, 单副本)                   │
│                                             │
│ ┌──────────────┐ ┌──────────────────────┐   │
│ │ 连接管理器   │ │   任务执行器         │   │
│ │ - WebSocket  │ │ - helm install       │   │
│ │ - HTTP 心跳  │ │ - helm upgrade       │   │
│ │ - 断线重连   │ │ - helm uninstall     │   │
│ │ - 指数退避   │ │ - kubectl apply      │   │
│ └──────────────┘ └──────────────────────┘   │
│                                             │
│ ┌──────────────┐ ┌──────────────────────┐   │
│ │ 状态采集器   │ │   调谐控制器         │   │
│ │ - 节点信息   │ │ - Drift 检测         │   │
│ │ - 资源指标   │ │ - 期望 vs 实际对比   │   │
│ │ - Pod 状态   │ │ - 自动纠正/告警      │   │
│ │ - 事件收集   │ │ - 定期巡检           │   │
│ └──────────────┘ └──────────────────────┘   │
│                                             │
│ ┌──────────────┐ ┌──────────────────────┐   │
│ │ 隧道代理     │ │   本地缓存           │   │
│ │ - K8s API   │ │ - 期望状态缓存       │   │
│ │   代理      │ │ - 离线自治           │   │
│ │ - exec/log  │ │ - 断网正常运行       │   │
│ └──────────────┘ └──────────────────────┘   │
└─────────────────────────────────────────────┘
```

#### 3.3.2 Agent 通信协议

```
Agent 启动流程:
  1. 读取环境变量 (SERVER_URL, CLUSTER_TOKEN)
  2. 调用 POST /api/v1/clusters/register → 验证身份，获取 cluster_id
  3. 建立 WebSocket 连接: wss://<server>/api/v1/clusters/<id>/tunnel
  4. WebSocket 连接承担三个职责:
     a. 心跳保活 (30s ping/pong)
     b. 接收任务指令 (服务端推送)
     c. 代理 K8s API 请求 (隧道)

断线重连策略:
  重连间隔: 1s → 2s → 4s → 8s → 16s → 30s (最大)
  每次重连前先尝试 HTTP 心跳 (POST /heartbeat) 作为降级通道
  重连成功后同步离线期间的期望状态变更
```

#### 3.3.3 Agent 调谐循环

```
Agent 调谐循环 (每 60s):
  1. 从 API Server 拉取当前集群的期望状态 (所有 Deployment)
  2. 查询本地 K3s 集群的实际状态 (helm list + kubectl get)
  3. 对比期望 vs 实际:
     ├── 缺失 → helm install
     ├── 版本不一致 → helm upgrade
     ├── 多余 → helm uninstall (如果期望状态中已删除)
     └── 配置 Drift → 告警 / 自动纠正 (可配置)
  4. 上报调谐结果
```

### 3.4 应用市场

```
Application {
  id              string    // 唯一标识
  name            string    // 应用标识 (如 erp-system)
  display_name    string    // 显示名称
  description     string    // 应用描述
  icon            string    // 图标 URL
  chart_name      string    // Helm Chart 名称
  chart_repo      string    // Helm 仓库地址
  chart_version   string    // 默认版本
  category        string    // 分类 (saas/business/middleware)
  is_saas         bool      // 是否为平台 SaaS 应用
  default_values  string    // 默认 Helm Values
}
```

### 3.5 部署编排

```
Deployment {
  id              string    // 唯一标识
  cluster_id      string    // 目标集群
  application_id  string    // 应用模板
  name            string    // 部署实例名称
  namespace       string    // K8s 命名空间
  values          string    // Helm Values (YAML, 期望状态)
  status          string    // pending/syncing/synced/drifted/error/stopped
  actual_status   string    // Agent 上报的实际状态
  replicas        int       // 期望副本数
  version         string    // 期望版本
  actual_version  string    // 实际运行版本
  drift_detail    string    // Drift 详情
  message         string    // 状态信息
  last_synced     time      // 最后调谐时间
}
```

### 3.6 组织/项目管理（原客户管理调整）

> 定位变更：从"管理不同企业客户"调整为"管理客户内部的部门/项目/团队"

```
Organization {
  id              string    // 唯一标识
  name            string    // 组织/项目名称
  code            string    // 组织编码 (唯一，用于标签选择)
  contact         string    // 负责人
  phone           string    // 联系电话
  email           string    // 邮箱
  type            string    // 类型 (department/project/team)
  description     string    // 描述
  cluster_ids     []string  // 关联的集群列表
  created_at      time
  updated_at      time
}
```

**用途**：
- 按组织归属管理集群（如"生产部"的3个集群、"测试组"的2个集群）
- 按组织维度查看资源使用、部署状态
- 后续可扩展为组织级别的权限隔离

### 3.7 License 管理

```
License {
  id              string    // 唯一标识
  key             string    // License 密钥 (加密存储)
  product         string    // 产品标识
  customer_name   string    // 客户名称
  issued_at       time      // 签发时间
  expires_at      time      // 过期时间
  max_clusters    int       // 最大集群数
  max_deployments int       // 最大部署数
  features        string    // 功能开关 (JSON, 如 {"monitoring":true,"tunnel":true})
  is_valid        bool      // 是否有效
}
```

**校验策略**：
- 注册新集群时检查 `当前集群数 < max_clusters`
- 创建新部署时检查 `当前部署数 < max_deployments`
- 每次心跳时检查 License 是否过期
- 过期后：禁止新增集群和部署，已有部署继续运行
- 功能开关：未授权的功能在 UI 上置灰或隐藏

### 3.8 监控告警

#### 3.8.1 指标采集

Agent 每次心跳上报以下指标：

```
ClusterMetrics {
  cluster_id      string
  node_count      int
  cpu_usage       float64   // CPU 使用率百分比
  mem_usage       float64   // 内存使用率百分比
  pod_count       int
  cpu_capacity    string    // CPU 总量
  mem_capacity    string    // 内存总量
  version         string    // K3s 版本
}

NodeMetrics {
  cluster_id      string
  node_name       string
  cpu_usage       float64
  mem_usage       float64
  disk_usage      float64
  pod_count       int
  conditions      string    // 节点条件 (Ready/MemoryPressure/DiskPressure)
}
```

#### 3.8.2 告警规则

```
AlertRule {
  id              string
  name            string    // 规则名称
  type            string    // cluster_down / cpu_high / mem_high / drift_detected / license_expiring
  condition       string    // 触发条件 (JSON, 如 {"metric":"cpu_usage","operator":">","threshold":80,"duration":"5m"})
  severity        string    // critical / warning / info
  enabled         bool
  notify_channels string    // 通知渠道 (JSON, 如 {"webhook":"http://...","email":"admin@example.com"})
  last_triggered  time
  created_at      time
}
```

#### 3.8.3 告警记录

```
AlertRecord {
  id              string
  rule_id         string
  rule_name       string
  cluster_id      string
  severity        string
  message         string
  status          string    // firing / resolved
  triggered_at    time
  resolved_at     time
}
```

### 3.9 配置中心

按组织维度管理 Helm Values 模板和覆盖：

```
ConfigTemplate {
  id              string
  name            string    // 配置模板名称
  org_id          string    // 所属组织 (空表示全局)
  application_id  string    // 关联应用
  values          string    // Helm Values (YAML)
  description     string
  created_at      time
  updated_at      time
}
```

部署时可选择配置模板，支持 Values 覆盖层级：
```
应用默认 Values → 组织模板覆盖 → 部署时自定义覆盖
```

### 3.10 审计日志

```
AuditLog {
  id              string
  user_id         string    // 操作用户
  username        string
  action          string    // 操作类型 (create/update/delete/login/deploy)
  resource_type   string    // 资源类型 (cluster/application/deployment/organization)
  resource_id     string    // 资源 ID
  resource_name   string    // 资源名称
  detail          string    // 操作详情 (JSON)
  ip              string    // 来源 IP
  created_at      time
}
```

### 3.11 隧道代理

基于 WebSocket 隧道，支持远程操作集群：

```
隧道请求:
  POST /api/v1/clusters/:id/proxy
  {
    "method": "GET",
    "path": "/api/v1/namespaces/default/pods",
    "headers": {"Authorization": "Bearer ..."},
    "body": ""
  }

隧道响应:
  {
    "status_code": 200,
    "headers": {"Content-Type": "application/json"},
    "body": "{\"items\":[...]}"
  }

支持的操作:
  - kubectl get (查询资源)
  - kubectl logs (查看日志，WebSocket 流式)
  - kubectl exec (远程终端，WebSocket 双向流)
  - kubectl port-forward (端口转发，WebSocket 隧道)
```

---

## 4. API 设计

### 4.1 API 总览

基础路径：`/api/v1`

| 方法 | 路径 | 说明 | 状态 |
|------|------|------|------|
| **认证** | | | |
| POST | /auth/login | 登录 | ✅ |
| POST | /auth/token | 刷新 Token | ✅ |
| GET | /auth/me | 当前用户信息 | ✅ |
| **仪表盘** | | | |
| GET | /dashboard/stats | 统计概览 | ✅ |
| **集群管理** | | | |
| POST | /clusters | 注册集群 | ✅ |
| GET | /clusters | 集群列表 | ✅ |
| GET | /clusters/:id | 集群详情 | ✅ |
| DELETE | /clusters/:id | 删除集群 | ✅ |
| PUT | /clusters/:id/labels | 更新集群标签 | ✅ |
| GET | /clusters/:id/install-script | 获取安装脚本 | ✅ |
| GET | /clusters/:id/registration.yaml | 获取注册 YAML | ✅ |
| POST | /clusters/:id/heartbeat | Agent 心跳 | ✅ |
| GET | /clusters/:id/desired-state | Agent 拉取期望状态 | ✅ |
| POST | /clusters/:id/sync-result | Agent 上报调谐结果 | ✅ |
| GET | /clusters/:id/tunnel | WebSocket 隧道 | ✅ |
| POST | /clusters/:id/token/rotate | 轮换集群 Token | ❌ 待实现 |
| GET | /clusters/:id/metrics | 获取集群指标历史 | ❌ 待实现 |
| GET | /clusters/:id/nodes | 获取节点列表 | ❌ 待实现 |
| POST | /clusters/:id/proxy | 隧道代理 K8s API | ❌ 待实现 |
| **应用市场** | | | |
| POST | /applications | 创建应用 | ✅ |
| GET | /applications | 应用列表 | ✅ |
| GET | /applications/:id | 应用详情 | ✅ |
| PUT | /applications/:id | 更新应用 | ✅ |
| DELETE | /applications/:id | 删除应用 | ✅ |
| **部署管理** | | | |
| POST | /deployments | 创建部署 | ✅ |
| POST | /deployments/batch | 批量部署 | ✅ |
| GET | /deployments | 部署列表 | ✅ |
| GET | /deployments/:id | 部署详情 | ✅ |
| PUT | /deployments/:id | 更新期望状态 | ✅ |
| DELETE | /deployments/:id | 删除部署 | ✅ |
| **组织管理** | | | |
| POST | /organizations | 创建组织 | ❌ 待实现(原 /customers) |
| GET | /organizations | 组织列表 | ❌ 待实现 |
| GET | /organizations/:id | 组织详情 | ❌ 待实现 |
| PUT | /organizations/:id | 更新组织 | ❌ 待实现 |
| DELETE | /organizations/:id | 删除组织 | ❌ 待实现 |
| **License** | | | |
| GET | /license | 获取当前 License 信息 | ❌ 待实现 |
| POST | /license/activate | 激活 License | ❌ 待实现 |
| **监控告警** | | | |
| GET | /alerts/rules | 获取告警规则列表 | ❌ 待实现 |
| POST | /alerts/rules | 创建告警规则 | ❌ 待实现 |
| PUT | /alerts/rules/:id | 更新告警规则 | ❌ 待实现 |
| DELETE | /alerts/rules/:id | 删除告警规则 | ❌ 待实现 |
| GET | /alerts/records | 获取告警记录 | ❌ 待实现 |
| **配置中心** | | | |
| GET | /configs | 获取配置模板列表 | ❌ 待实现 |
| POST | /configs | 创建配置模板 | ❌ 待实现 |
| PUT | /configs/:id | 更新配置模板 | ❌ 待实现 |
| DELETE | /configs/:id | 删除配置模板 | ❌ 待实现 |
| **审计日志** | | | |
| GET | /audit-logs | 获取审计日志列表 | ❌ 待实现 |
| **用户管理** | | | |
| POST | /users | 创建用户 | ❌ 待实现 |
| GET | /users | 用户列表 | ❌ 待实现 |
| PUT | /users/:id | 更新用户 | ❌ 待实现 |
| DELETE | /users/:id | 删除用户 | ❌ 待实现 |

---

## 5. 前端页面设计

### 5.1 页面结构

```
┌─────────────────────────────────────────────────────────────┐
│ KubeNexus                              [用户名] [退出]      │
├──────┬──────────────────────────────────────────────────────┤
│      │                                                      │
│ 仪表盘│                                                      │
│      │             (主内容区)                                │
│ 集群  │                                                      │
│ 管理  │                                                      │
│      │                                                      │
│ 应用  │                                                      │
│ 市场  │                                                      │
│      │                                                      │
│ 部署  │                                                      │
│ 管理  │                                                      │
│      │                                                      │
│ 组织  │                                                      │
│ 管理  │                                                      │
│      │                                                      │
│ 监控  │                                                      │
│ 告警  │                                                      │
│      │                                                      │
│ 配置  │                                                      │
│ 中心  │                                                      │
│      │                                                      │
│ 系统  │                                                      │
│ 设置  │                                                      │
│      │                                                      │
└──────┴──────────────────────────────────────────────────────┘
```

### 5.2 页面清单

| 页面 | 功能 | 状态 |
|------|------|------|
| **登录** | 用户名密码登录 | ✅ |
| **仪表盘** | 集群总数、在线/离线、应用部署数、资源概览、告警摘要 | ✅(需补充告警) |
| **集群管理** | 集群列表、注册新集群、集群详情、标签管理、指标图表 | ✅(需补充指标) |
| **应用市场** | 应用模板列表、添加应用、分类筛选 | ✅ |
| **部署管理** | 部署实例列表、创建部署、批量部署、Drift 状态 | ✅ |
| **组织管理** | 组织/项目列表、创建组织、关联集群 | ❌ 待实现 |
| **监控告警** | 资源趋势图、告警规则管理、告警记录 | ❌ 待实现 |
| **配置中心** | 配置模板管理、Values 编辑器 | ❌ 待实现 |
| **系统设置** | License 管理、用户管理、审计日志 | ❌ 待实现 |

---

## 6. 安全设计

### 6.1 通信安全

| 安全措施 | 说明 |
|---------|------|
| 传输加密 | 全链路 HTTPS |
| Agent 认证 | Token + TLS 证书校验 |
| Token 安全 | 注册后 Token 可轮换，旧 Token 失效 |
| 最小权限 | Agent SA 仅授予必要 RBAC |
| 隧道安全 | WebSocket over TLS |

### 6.2 敏感数据

- Token、License 等敏感字段不通过 API 返回（`json:"-"` 标记）
- 数据库中 License 密钥加密存储（AES-256）
- Helm Values 中可能包含数据库密码等，加密存储

### 6.3 权限模型

| 角色 | 权限 |
|------|------|
| admin | 全部操作权限 |
| operator | 查看集群/应用/部署，创建/管理部署，不能管理用户和 License |
| viewer | 只读查看 |

---

## 7. 开发计划

### Phase 1 — MVP (核心链路) ✅ 已完成

- [x] 后端 API Server (集群/应用/部署 CRUD)
- [x] Agent 注册 + HTTP 心跳 + 任务轮询
- [x] Agent 声明式调谐 (拉取期望状态 → helm install/upgrade)
- [x] 前端核心页面 (仪表盘/集群/应用/部署)
- [x] K3s 一键安装脚本
- [x] 注册 YAML 生成
- [x] 示例 Helm Chart
- [x] JWT 认证 + 角色权限
- [x] WebSocket 隧道框架
- [x] 批量部署 (clusterSelector)

### Phase 2 — 生产可用（当前阶段）

- [ ] 客户管理 → 组织/项目管理调整
- [ ] License 校验 + 配额 + 功能开关
- [ ] Token 一次性使用 + 轮换
- [ ] Agent WebSocket 客户端连接
- [ ] 隧道代理 K8s API 转发
- [ ] 监控指标采集与展示
- [ ] 告警规则与通知
- [ ] 配置中心 (按组织维度的 Values 管理)
- [ ] 审计日志
- [ ] 用户管理 (CRUD)
- [ ] 敏感字段加密存储
- [ ] 前端补充所有新页面

### Phase 3 — 体验提升

- [ ] 灰度发布 / 金丝雀部署
- [ ] 日志聚合查看
- [ ] 集群升级管理
- [ ] GitOps 模式 (可选)
- [ ] Web Terminal
- [ ] 数据库迁移到 PostgreSQL
- [ ] 离线安装包制作

---

## 8. 风险与应对

| 风险 | 影响 | 应对策略 |
|------|------|---------|
| 客户网络不稳定 | Agent 断连 | HTTP 心跳降级 + 指数退避重连 + 本地自治 |
| 客户环境差异大 | 安装失败 | 充分测试主流 OS、提供环境检测脚本 |
| Helm Chart 质量参差 | 部署失败 | 建立 Chart 质量规范、自动化测试 |
| 声明式调谐冲突 | Agent 与手动操作冲突 | Drift 检测告警、可配置 auto-reconcile 开关 |
| License 被破解 | 商业损失 | 密钥加密 + 服务端校验 + 定期在线验证(可选) |
