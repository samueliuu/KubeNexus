# 衡牧KubeNexusK3s多集群管理系统 架构设计文档

## 1. 系统概述

### 1.1 产品定位

衡牧KubeNexusK3s多集群管理系统是面向企业的轻量级 K3s 多集群统一管理平台，采用本地部署模式。系统通过中心管控面与集群代理的协同架构，实现对多个 K3s 集群的注册纳管、应用分发、状态监控、告警管理和配置治理等核心能力。

### 1.2 核心设计理念

- **管控面-代理分离**：中心管控面负责状态维护与策略决策，Agent 负责本地执行与状态上报，职责清晰
- **声明式部署**：管控面维护期望状态，Agent 定期拉取期望状态并执行调谐，自动检测配置漂移
- **双通道通信**：WebSocket 长连接提供实时双向通信隧道，HTTP 短连接承载心跳与状态同步
- **零信任安全**：JWT 用户认证 + Agent Token 机器认证双重机制，K8s API 代理采用白名单模式
- **轻量级部署**：管控面使用 SQLite 单文件存储，Agent 使用 Go 标准库零外部依赖

### 1.3 技术栈

| 层级 | 技术选型 |
|------|----------|
| 管控面后端 | Go 1.22+ / Gin / GORM / SQLite |
| 管控面前端 | React 18 / TypeScript / Ant Design 5 / ProComponents / Vite |
| 集群代理 | Go 标准库（零外部依赖） |
| 通信协议 | WebSocket + HTTP 双通道 |
| 认证机制 | JWT（用户）+ Agent Token（代理）双重认证 |
| 包管理器 | Helm（Agent 端执行） |

---

## 2. 系统架构

### 2.1 整体架构图

```
                           +---------------------------+
                           |       运维管理员           |
                           +------------+--------------+
                                        |
                                   浏览器访问
                                        |
                                        v
+-----------------------------------------------------------------------+
|                          管控面（控制平面）                              |
|                                                                       |
|  +------------------+    +------------------+    +-----------------+  |
|  |   前端 Web UI    |    |   后端 API 服务   |    |   后台任务引擎   |  |
|  |                  |    |                  |    |                 |  |
|  | React 18         |    | Gin Router       |    | 集群状态检查器   |  |
|  | Ant Design 5     |    | JWT 认证中间件    |    | 告警评估引擎     |  |
|  | ProComponents    |    | Agent Token 认证  |    | 心跳数据清理     |  |
|  | TypeScript       |    | CORS 中间件      |    |                 |  |
|  +--------+---------+    +--------+---------+    +--------+--------+  |
|           |                       |                       |           |
|           |    REST API           |                       |           |
|           +-------><--------------+                       |           |
|                                   |                       |           |
|                           +-------+--------+              |           |
|                           |   Service 层    |<-------------+           |
|                           |                |                          |
|                           | ClusterService |                          |
|                           | AppService     |                          |
|                           | DeployService  |                          |
|                           | OrgService     |                          |
|                           | LicenseService |                          |
|                           | AlertService   |                          |
|                           | ConfigService  |                          |
|                           | AuditService   |                          |
|                           | UserService    |                          |
|                           | AuthService    |                          |
|                           | DashboardSvc   |                          |
|                           +-------+--------+                          |
|                                   |                                   |
|                           +-------+--------+                          |
|                           |   Store 层      |                          |
|                           |  GORM + SQLite  |                          |
|                           +----------------+                          |
|                                                                       |
|  +----------------------------------------------------------------+  |
|  |                    WebSocket 隧道管理器                          |  |
|  |                                                                |  |
|  |  连接池管理  |  消息路由  |  K8s API 代理  |  任务下发          |  |
|  +----------------------------------------------------------------+  |
+-----------------------------------------------------------------------+
            |                    |                    |
       WebSocket            HTTP 心跳            HTTP 同步
       长连接隧道           (30秒间隔)           (60秒间隔)
            |                    |                    |
            v                    v                    v
+-------------------+  +-------------------+  +-------------------+
|   K3s 集群 A      |  |   K3s 集群 B      |  |   K3s 集群 C      |
|                   |  |                   |  |                   |
| +---------------+ |  | +---------------+ |  | +---------------+ |
| | KubeNexus     | |  | | KubeNexus     | |  | | KubeNexus     | |
| | Agent         | |  | | Agent         | |  | | Agent         | |
| |               | |  | |               | |  | |               | |
| | 心跳采集      | |  | | 心跳采集      | |  | | 心跳采集      | |
| | 状态同步      | |  | | 状态同步      | |  | | 状态同步      | |
| | Helm 执行     | |  | | Helm 执行     | |  | | Helm 执行     | |
| | K8s API 代理  | |  | | K8s API 代理  | |  | | K8s API 代理  | |
| +-------+-------+ |  | +-------+-------+ |  | +-------+-------+ |
|         |         |  |         |         |  |         |         |
|   K3s 集群资源    |  |   K3s 集群资源    |  |   K3s 集群资源    |
|   Helm Charts     |  |   Helm Charts     |  |   Helm Charts     |
+-------------------+  +-------------------+  +-------------------+
```

### 2.2 分层架构

管控面后端采用严格的四层架构设计，各层职责明确，单向依赖：

```
+--------------------------------------------------+
|                    API 层                         |
|  路由注册 | 请求校验 | 参数绑定 | 响应序列化       |
|  认证中间件 | 权限中间件 | 审计日志               |
+--------------------------------------------------+
                        |
                        v
+--------------------------------------------------+
|                  Service 层                       |
|  业务逻辑 | 状态编排 | License 配额检查            |
|  期望状态计算 | 同步结果处理 | 告警触发            |
+--------------------------------------------------+
                        |
                        v
+--------------------------------------------------+
|                  Store 层                         |
|  数据持久化 | 查询构建 | 事务管理                  |
|  GORM 模型映射 | 软删除 | 数据初始化              |
+--------------------------------------------------+
                        |
                        v
+--------------------------------------------------+
|                  SQLite                           |
|  单文件数据库 | 零运维 | 嵌入式部署               |
+--------------------------------------------------+
```

| 层级 | 职责 | 关键约束 |
|------|------|----------|
| API 层 | 路由注册、请求校验、参数绑定、响应序列化、认证鉴权 | 不包含业务逻辑，仅做参数转换和错误映射 |
| Service 层 | 业务逻辑、状态编排、配额检查、期望状态计算 | 可跨 Store 操作，但不直接操作数据库连接 |
| Store 层 | 数据持久化、查询构建、事务管理、模型映射 | 仅操作 GORM，不包含业务判断逻辑 |
| SQLite | 数据存储 | 单文件，零配置，嵌入式 |

---

## 3. 核心组件

### 3.1 管控面后端

#### 3.1.1 API 层

API 层基于 Gin 框架实现，统一注册在 `/api/v1` 路径下，按资源域划分路由组：

| 路由组 | 路径前缀 | 认证方式 | 说明 |
|--------|----------|----------|------|
| 认证 | `/api/v1/auth` | JWT | 登录、令牌刷新、当前用户 |
| 仪表盘 | `/api/v1/dashboard` | JWT | 全局统计数据 |
| 集群 | `/api/v1/clusters` | JWT / Agent Token | 集群管理、心跳、同步、隧道 |
| 应用 | `/api/v1/applications` | JWT | 应用模板管理 |
| 部署 | `/api/v1/deployments` | JWT | 部署实例管理 |
| 组织 | `/api/v1/organizations` | JWT | 组织机构管理 |
| 许可证 | `/api/v1/license` | JWT | 许可证与配额 |
| 告警 | `/api/v1/alerts` | JWT | 告警规则与记录 |
| 配置 | `/api/v1/configs` | JWT | 配置模板管理 |
| 审计 | `/api/v1/audit-logs` | JWT | 操作审计日志 |
| 用户 | `/api/v1/users` | JWT + Admin | 用户管理 |

权限控制分为两个层级：

- **角色权限**：`admin` 角色拥有全部操作权限，`viewer` 角色仅拥有只读权限
- **写入保护**：创建、修改、删除等写操作通过 `AdminMiddleware` 限制为管理员角色

#### 3.1.2 认证中间件

系统采用双重认证机制，分别服务于不同场景：

**JWT 用户认证**（`AuthMiddleware`）：

- 用于浏览器端用户访问
- 请求头携带 `Authorization: Bearer <token>`
- Token 包含 `user_id`、`username`、`role` 声明
- Token 有效期 24 小时，通过 `JWT_SECRET` 环境变量签名
- 未设置 `JWT_SECRET` 时自动生成随机密钥（重启后失效，生产环境必须配置）

**Agent Token 机器认证**（`AgentAuthMiddleware`）：

- 用于集群 Agent 访问管控面
- 请求头携带 `X-Cluster-Token: <token>`
- Token 格式为 `cn-<uuid>`，注册集群时自动生成
- 通过数据库查询 Token 对应的集群记录进行验证
- 支持通过 API 轮换 Token

#### 3.1.3 Service 层

Service 层是业务逻辑的核心，包含以下服务：

| 服务 | 职责 |
|------|------|
| ClusterService | 集群注册、心跳处理、期望状态计算、同步结果处理、安装脚本生成 |
| ApplicationService | 应用模板的增删改查 |
| DeploymentService | 部署创建、批量部署（支持标签选择器）、部署更新与删除 |
| OrganizationService | 组织机构管理 |
| LicenseService | 许可证管理、配额检查 |
| AlertService | 告警规则管理、告警记录查询与确认 |
| ConfigService | 配置模板管理，支持按组织过滤 |
| AuditService | 审计日志记录与查询 |
| UserService | 用户管理、密码哈希 |
| AuthService | 用户登录验证 |
| DashboardService | 仪表盘统计数据聚合 |

#### 3.1.4 Store 层

Store 层基于 GORM 实现，使用 SQLite 作为存储引擎：

- 自动迁移：启动时通过 `AutoMigrate` 自动创建和更新表结构
- 软删除：核心业务模型（Cluster、Application、Deployment、Organization、User）使用 `gorm.DeletedAt` 实现软删除
- 数据初始化：首次启动时自动创建默认管理员账号、默认许可证和默认告警规则
- 标签查询：集群标签使用 JSON 存储，通过 `json_extract` 函数实现标签过滤查询
- 心跳清理：每小时自动清理 7 天前的历史心跳数据

#### 3.1.5 WebSocket 隧道管理器

隧道管理器是管控面与 Agent 之间实时通信的核心组件：

```
+------------------------------------------------------------------+
|                     隧道管理器（Manager）                          |
|                                                                  |
|  +------------------------------------------------------------+  |
|  |                   连接池（connections）                      |  |
|  |                                                            |  |
|  |  map[string]*ClusterConnection                             |  |
|  |    ├── cluster_id_1 → ClusterConnection                    |  |
|  |    │     ├── WebSocket 连接                                 |  |
|  |    │     ├── 待响应请求映射（pendingReqs）                   |  |
|  |    │     └── 互斥锁（mu）                                   |  |
|  |    ├── cluster_id_2 → ClusterConnection                    |  |
|  |    └── cluster_id_3 → ClusterConnection                    |  |
|  +------------------------------------------------------------+  |
|                                                                  |
|  消息类型：                                                      |
|    heartbeat       → 心跳保活                                    |
|    task            → 任务下发（Helm 安装/升级/卸载）              |
|    tunnel_request  → 隧道请求（K8s API 代理转发）               |
|    tunnel_response → 隧道响应                                    |
|                                                                  |
|  回调机制：                                                      |
|    onConnect    → 集群上线时更新 ws_connected = true             |
|    onDisconnect → 集群下线时更新 ws_connected = false            |
+------------------------------------------------------------------+
```

关键设计：

- **连接管理**：同一集群重复连接时，自动关闭旧连接，保留新连接
- **请求-响应匹配**：通过请求 ID 和 channel 实现异步请求的同步等待，超时时间 30 秒
- **并发安全**：每个连接独立持有互斥锁，保护 WebSocket 写操作和 pendingReqs 映射
- **K8s API 代理**：白名单模式，仅允许 GET 请求访问安全路径

K8s API 代理白名单路径：

| 允许的路径 | 说明 |
|------------|------|
| `/api/v1/nodes` | 节点信息 |
| `/api/v1/pods` | Pod 信息 |
| `/api/v1/services` | 服务信息 |
| `/api/v1/namespaces` | 命名空间信息 |
| `/apis/apps/v1/deployments` | Deployment 信息 |
| `/api/v1/events` | 事件信息 |

#### 3.1.6 后台任务引擎

系统启动后自动运行三个后台任务：

| 任务 | 执行间隔 | 职责 |
|------|----------|------|
| 集群状态检查器 | 60 秒 | 检查活跃集群心跳超时（90 秒），标记为不可用 |
| 告警评估引擎 | 60 秒 | 遍历启用的告警规则，对每个集群评估条件，触发或恢复告警 |
| 心跳数据清理 | 1 小时 | 清理 7 天前的历史心跳记录 |

### 3.2 集群代理（Agent）

#### 3.2.1 Agent 架构

Agent 是部署在每个 K3s 集群中的轻量级代理程序，使用 Go 标准库实现，零外部依赖：

```
+----------------------------------------------+
|              KubeNexus Agent                  |
|                                              |
|  +------------------+  +------------------+  |
|  |   心跳采集模块    |  |   状态同步模块    |  |
|  |                  |  |                  |  |
|  | 节点数量采集     |  | 拉取期望状态     |  |
|  | Pod 数量采集     |  | Helm 安装执行    |  |
|  | 版本信息采集     |  | Helm 升级执行    |  |
|  | CPU/内存采集     |  | 卸载执行         |  |
|  +--------+---------+  +--------+---------+  |
|           |                       |          |
|           | 30秒间隔              | 60秒间隔  |
|           v                       v          |
|  +-----------------------------------------+ |
|  |           HTTP 通信模块                   | |
|  |                                         | |
|  |  X-Cluster-Token 认证                   | |
|  |  心跳上报 POST /clusters/:id/heartbeat  | |
|  |  期望状态 GET /clusters/:id/desired-state| |
|  |  同步结果 POST /clusters/:id/sync-result | |
|  +-----------------------------------------+ |
|                                              |
|  +-----------------------------------------+ |
|  |           本地执行引擎                    | |
|  |                                         | |
|  |  kubectl 命令执行                        | |
|  |  helm install / upgrade 命令执行         | |
|  |  临时 Values 文件管理                    | |
|  +-----------------------------------------+ |
+----------------------------------------------+
```

#### 3.2.2 Agent 运行流程

```
启动
  |
  v
检查 CLUSTER_ID 是否配置
  |
  +-- 未配置 --> 调用注册接口 --> 获取 CLUSTER_ID
  |
  v
发送初始心跳
  |
  v
执行初始同步
  |
  v
进入主循环
  |
  +-- 心跳定时器触发 (30秒) --> 采集指标 --> 上报心跳
  |
  +-- 同步定时器触发 (60秒) --> 拉取期望状态 --> 调谐部署 --> 上报同步结果
  |
  +-- 上下文取消 --> 优雅退出
```

#### 3.2.3 声明式调谐逻辑

Agent 拉取期望状态后，对每个部署项执行调谐：

| 期望动作 | Agent 行为 | 成功结果 | 失败结果 |
|----------|-----------|----------|----------|
| `install` | 执行 `helm install` | 状态标记为 `synced` | 状态标记为 `error` |
| `upgrade` | 执行 `helm upgrade` | 状态标记为 `synced` | 状态标记为 `error` |
| `sync` | 无需操作 | 状态标记为 `synced` | - |

管控面计算期望动作的逻辑：

- 部署的 `ActualStatus` 为空 → 动作为 `install`
- 部署的 `Version` 与 `ActualVersion` 不一致 → 动作为 `upgrade`
- 其他情况 → 动作为 `sync`

已停止（`status=stopped`）的部署会被放入 `removed` 列表，Agent 执行卸载操作。

#### 3.2.4 Agent 部署方式

Agent 通过 Kubernetes 原生资源部署到目标集群，管控面提供自动生成的注册 YAML：

```
kubenexus-system 命名空间
  |
  +-- Secret: kubenexus-agent-config
  |     ├── SERVER_URL
  |     ├── CLUSTER_TOKEN
  |     └── CLUSTER_ID
  |
  +-- ServiceAccount: kubenexus-agent
  |
  +-- ClusterRole: kubenexus-agent (全部资源权限)
  |
  +-- ClusterRoleBinding: kubenexus-agent
  |
  +-- Deployment: kubenexus-agent
        └── 容器: agent
              ├── 资源限制: CPU 200m / 内存 256Mi
              ├── 资源请求: CPU 50m / 内存 64Mi
              └── 环境变量从 Secret 注入
```

### 3.3 管控面前端

#### 3.3.1 前端架构

前端基于 React 18 + TypeScript + Ant Design 5 构建，采用 Vite 作为构建工具：

```
+----------------------------------------------+
|              前端应用                          |
|                                              |
|  路由层（React Router v6）                    |
|    /login           → 登录页                  |
|    /dashboard       → 仪表盘                  |
|    /clusters        → 集群列表                |
|    /clusters/:id    → 集群详情                |
|    /applications    → 应用管理                |
|    /deployments     → 部署管理                |
|    /organizations   → 组织管理                |
|    /alerts          → 告警中心                |
|    /configs         → 配置中心                |
|    /settings        → 系统设置                |
|                                              |
|  认证层（AuthContext）                         |
|    登录状态管理                                |
|    Token 持久化                                |
|    路由守卫（ProtectedRoute）                  |
|                                              |
|  布局层（AppLayout）                           |
|    侧边栏导航                                  |
|    顶部栏                                      |
|    内容区域                                     |
|                                              |
|  API 层（api/index.ts）                        |
|    统一请求封装                                |
|    Token 注入                                  |
|    错误处理                                    |
+----------------------------------------------+
```

#### 3.3.2 前端部署模式

前端构建产物嵌入后端服务，通过 Gin 静态文件服务提供：

- `/assets/*` → 前端静态资源（JS、CSS）
- 其他路径 → `index.html`（SPA 路由兜底）

---

## 4. 数据模型

### 4.1 实体关系图

```
+---------------+       +------------------+       +------------------+
| Organization  |       |    Cluster       |       |   Application    |
+---------------+       +------------------+       +------------------+
| id            |<--+   | id               |   +-->| id               |
| name          |   |   | name             |   |   | name             |
| code          |   |   | display_name     |   |   | display_name     |
| contact       |   |   | status           |   |   | description      |
| phone         |   |   | token            |   |   | icon             |
| email         |   |   | endpoint         |   |   | chart_name       |
| type          |   |   | version          |   |   | chart_repo       |
| description   |   |   | node_count       |   |   | chart_version    |
+---------------+   |   | cpu_capacity     |   |   | category         |
                    |   | mem_capacity     |   |   | is_saas          |
                    |   | labels (JSON)    |   |   | default_values   |
                    |   | region           |   |   +------------------+
                    +---| org_id           |   |
                        | org_name         |   |
                        | last_heartbeat   |   |
                        | ws_connected     |   |
                        +--------+---------+   |
                                 |             |
                                 | 1:N         |
                                 v             |
                        +------------------+   |
                        |   Deployment     |   |
                        +------------------+   |
                        | id               |   |
                        | cluster_id   ----+   |
                        | application_id --+---+
                        | name             |
                        | namespace        |
                        | values           |
                        | status           |
                        | actual_status    |
                        | replicas         |
                        | version          |
                        | actual_version   |
                        | drift_detail     |
                        | message          |
                        | last_synced      |
                        +------------------+

+------------------+       +------------------+
|    AlertRule     |       |   AlertRecord    |
+------------------+       +------------------+
| id               |<--+   | id               |
| name             |   |   | rule_id      ----+
| type             |   |   | rule_name        |
| condition (JSON) |   |   | cluster_id       |
| severity         |   |   | severity         |
| enabled          |   |   | message          |
| notify_channels  |   |   | status           |
| last_triggered   |   |   | triggered_at     |
+------------------+   |   | resolved_at      |
                       |   +------------------+
                       |
+------------------+   |
| ConfigTemplate   |   |
+------------------+   |
| id               |   |
| name             |   |
| org_id           |   |
| application_id   |   |
| values           |   |
| description      |   |
+------------------+

+------------------+       +------------------+
|   AuditLog       |       |    Heartbeat     |
+------------------+       +------------------+
| id               |       | id               |
| user_id          |       | cluster_id       |
| username         |       | node_count       |
| action           |       | cpu_usage        |
| resource_type    |       | mem_usage        |
| resource_id      |       | pod_count        |
| resource_name    |       | version          |
| detail           |       | info             |
| ip               |       | reported_at      |
+------------------+       +------------------+

+------------------+       +------------------+
|     User         |       |     License      |
+------------------+       +------------------+
| id               |       | id               |
| username         |       | key              |
| password (bcrypt)|       | product          |
| role             |       | customer_name    |
+------------------+       | issued_at        |
                           | expires_at       |
                           | max_clusters     |
                           | max_deployments  |
                           | features (JSON)  |
                           | is_valid         |
                           +------------------+
```

### 4.2 数据模型详细定义

#### Cluster（集群）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | string(64) | 主键 | 集群唯一标识 |
| name | string(128) | 唯一索引 | 集群名称 |
| display_name | string(256) | - | 显示名称 |
| status | string(32) | 默认: registered | 集群状态：registered / active / unavailable |
| token | string(256) | - | Agent 认证令牌，JSON 响应中隐藏 |
| endpoint | string(512) | - | 集群访问地址 |
| version | string(64) | - | K3s 版本 |
| node_count | int | 默认: 0 | 节点数量 |
| cpu_capacity | string(32) | - | CPU 总容量 |
| mem_capacity | string(32) | - | 内存总容量 |
| labels | text (JSON) | - | 集群标签，键值对 |
| region | string(64) | - | 区域 |
| org_id | string(64) | 索引 | 所属组织标识 |
| org_name | string(256) | - | 所属组织名称（冗余） |
| last_heartbeat | timestamp | - | 最后心跳时间 |
| ws_connected | bool | 默认: false | WebSocket 连接状态 |
| created_at | timestamp | - | 创建时间 |
| updated_at | timestamp | - | 更新时间 |
| deleted_at | timestamp | 索引 | 软删除时间 |

#### Application（应用模板）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | string(64) | 主键 | 应用唯一标识 |
| name | string(128) | - | 应用名称 |
| display_name | string(256) | - | 显示名称 |
| description | string(1024) | - | 应用描述 |
| icon | string(512) | - | 应用图标地址 |
| chart_name | string(128) | - | Helm Chart 名称 |
| chart_repo | string(512) | - | Helm Chart 仓库地址 |
| chart_version | string(64) | - | Chart 默认版本 |
| category | string(64) | - | 应用分类 |
| is_saas | bool | 默认: true | 是否为 SaaS 应用 |
| default_values | text | - | 默认 Values 配置（YAML） |
| created_at | timestamp | - | 创建时间 |
| updated_at | timestamp | - | 更新时间 |
| deleted_at | timestamp | 索引 | 软删除时间 |

#### Deployment（部署实例）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | string(64) | 主键 | 部署唯一标识 |
| cluster_id | string(64) | 索引 | 目标集群标识 |
| application_id | string(64) | 索引 | 关联应用标识 |
| name | string(128) | - | 部署名称 |
| namespace | string(64) | 默认: default | 目标命名空间 |
| values | text | - | 自定义 Values 配置（YAML） |
| status | string(32) | 默认: pending | 期望状态：pending / syncing / synced / error / drifted / stopped |
| actual_status | string(32) | - | 实际状态 |
| replicas | int | 默认: 1 | 副本数 |
| version | string(64) | - | 期望版本 |
| actual_version | string(64) | - | 实际版本 |
| drift_detail | text | - | 配置漂移详情 |
| message | string(1024) | - | 状态消息 |
| last_synced | timestamp | - | 最后同步时间 |
| created_at | timestamp | - | 创建时间 |
| updated_at | timestamp | - | 更新时间 |
| deleted_at | timestamp | 索引 | 软删除时间 |

#### Organization（组织）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | string(64) | 主键 | 组织唯一标识 |
| name | string(256) | 唯一索引 | 组织名称 |
| code | string(64) | 唯一索引 | 组织编码 |
| contact | string(128) | - | 联系人 |
| phone | string(32) | - | 联系电话 |
| email | string(128) | - | 联系邮箱 |
| type | string(32) | 默认: department | 组织类型 |
| description | string(1024) | - | 组织描述 |
| created_at | timestamp | - | 创建时间 |
| updated_at | timestamp | - | 更新时间 |
| deleted_at | timestamp | 索引 | 软删除时间 |

#### License（许可证）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | string(64) | 主键 | 许可证标识 |
| key | string(1024) | - | 许可证密钥，JSON 响应中隐藏 |
| product | string(128) | - | 产品名称 |
| customer_name | string(256) | - | 客户名称 |
| issued_at | timestamp | - | 签发时间 |
| expires_at | timestamp | - | 过期时间 |
| max_clusters | int | 默认: 5 | 最大集群数 |
| max_deployments | int | 默认: 50 | 最大部署数 |
| features | text (JSON) | - | 功能特性开关 |
| is_valid | bool | 默认: true | 是否有效 |
| created_at | timestamp | - | 创建时间 |

#### AlertRule（告警规则）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | string(64) | 主键 | 规则唯一标识 |
| name | string(128) | - | 规则名称 |
| type | string(64) | - | 规则类型：cluster_down / cpu_high / mem_high / drift_detected / license_expiring |
| condition | text (JSON) | - | 触发条件 |
| severity | string(32) | 默认: warning | 严重级别：critical / warning |
| enabled | bool | 默认: true | 是否启用 |
| notify_channels | text | - | 通知渠道 |
| last_triggered | timestamp | - | 最后触发时间 |
| created_at | timestamp | - | 创建时间 |
| updated_at | timestamp | - | 更新时间 |

#### AlertRecord（告警记录）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | string(64) | 主键 | 记录唯一标识 |
| rule_id | string(64) | 索引 | 关联规则标识 |
| rule_name | string(128) | - | 规则名称（冗余） |
| cluster_id | string(64) | 索引 | 关联集群标识 |
| severity | string(32) | - | 严重级别 |
| message | string(1024) | - | 告警消息 |
| status | string(32) | 默认: firing | 告警状态：firing / resolved |
| triggered_at | timestamp | - | 触发时间 |
| resolved_at | timestamp | - | 恢复时间 |

#### ConfigTemplate（配置模板）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | string(64) | 主键 | 模板唯一标识 |
| name | string(128) | - | 模板名称 |
| org_id | string(64) | 索引 | 所属组织标识 |
| application_id | string(64) | 索引 | 关联应用标识 |
| values | text | - | Values 配置（YAML） |
| description | string(1024) | - | 模板描述 |
| created_at | timestamp | - | 创建时间 |
| updated_at | timestamp | - | 更新时间 |

#### AuditLog（审计日志）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | string(64) | 主键 | 日志唯一标识 |
| user_id | string(64) | 索引 | 操作用户标识 |
| username | string(128) | - | 操作用户名 |
| action | string(64) | - | 操作类型：login / create / update / delete / deploy / rotate_token / activate |
| resource_type | string(64) | - | 资源类型：cluster / application / deployment / organization / license / alert_rule / user / config_template |
| resource_id | string(64) | - | 资源标识 |
| resource_name | string(256) | - | 资源名称 |
| detail | text | - | 操作详情 |
| ip | string(64) | - | 客户端 IP |
| created_at | timestamp | - | 创建时间 |

#### Heartbeat（心跳记录）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | string(64) | 主键 | 心跳唯一标识 |
| cluster_id | string(64) | 索引 | 集群标识 |
| node_count | int | - | 节点数量 |
| cpu_usage | float64 | - | CPU 使用率（百分比） |
| mem_usage | float64 | - | 内存使用率（百分比） |
| pod_count | int | - | Pod 数量 |
| version | string(64) | - | K3s 版本 |
| info | text | - | 附加信息（JSON） |
| reported_at | timestamp | - | 上报时间 |

#### User（用户）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | string(64) | 主键 | 用户唯一标识 |
| username | string(128) | 唯一索引 | 用户名 |
| password | string(256) | - | 密码哈希（bcrypt），JSON 响应中隐藏 |
| role | string(32) | 默认: viewer | 角色：admin / viewer |
| created_at | timestamp | - | 创建时间 |
| updated_at | timestamp | - | 更新时间 |
| deleted_at | timestamp | 索引 | 软删除时间 |

---

## 5. 核心数据流

### 5.1 集群注册与纳管流程

```
管理员                管控面                  Agent                 K3s 集群
  |                    |                      |                      |
  | 1. 注册集群        |                      |                      |
  |------------------->|                      |                      |
  |    (名称/组织/区域)|                      |                      |
  |                    |                      |                      |
  |                    | 2. 生成 Agent Token  |                      |
  |                    |    创建集群记录       |                      |
  |                    |    (status=registered)|                     |
  |                    |                      |                      |
  | 3. 返回集群信息     |                      |                      |
  |<-------------------|                      |                      |
  |    (含 Token)      |                      |                      |
  |                    |                      |                      |
  | 4. 获取安装脚本     |                      |                      |
  |------------------->|                      |                      |
  |                    |                      |                      |
  | 5. 返回安装脚本     |                      |                      |
  |<-------------------|                      |                      |
  |                    |                      |                      |
  | 6. 在目标集群执行   |                      |                      |
  |---------------------------------------->|                      |
  |    安装脚本         |                      |                      |
  |                    |                      | 7. 创建 kubenexus    |
  |                    |                      |----->| system 命名空间 |
  |                    |                      |      | 部署 Agent      |
  |                    |                      |<-----|                 |
  |                    |                      |                      |
  |                    | 8. Agent 启动        |                      |
  |                    |    发送心跳          |                      |
  |                    |<---------------------|                      |
  |                    |                      |                      |
  |                    | 9. 更新集群状态       |                      |
  |                    |    (status=active)   |                      |
  |                    |                      |                      |
  |                    | 10. 建立 WebSocket   |                      |
  |                    |<====================>|                      |
  |                    |     长连接隧道        |                      |
```

### 5.2 声明式部署同步流程

```
管理员                管控面                  Agent                 K3s 集群
  |                    |                      |                      |
  | 1. 创建部署        |                      |                      |
  |------------------->|                      |                      |
  |    (集群/应用/配置) |                      |                      |
  |                    |                      |                      |
  |                    | 2. 创建部署记录       |                      |
  |                    |    (status=pending)   |                      |
  |                    |                      |                      |
  | 3. 部署创建成功     |                      |                      |
  |<-------------------|                      |                      |
  |                    |                      |                      |
  |                    | 4. 定时拉取期望状态   |                      |
  |                    |<---------------------|                      |
  |                    |                      |                      |
  |                    | 5. 返回期望状态       |                      |
  |                    |    (action=install)   |                      |
  |                    |---------------------->|                      |
  |                    |                      |                      |
  |                    |                      | 6. 执行 helm install |
  |                    |                      |---------------------->|
  |                    |                      |                      |
  |                    |                      | 7. 安装完成           |
  |                    |                      |<----------------------|
  |                    |                      |                      |
  |                    | 8. 上报同步结果       |                      |
  |                    |<---------------------|                      |
  |                    |    (status=synced)   |                      |
  |                    |                      |                      |
  |                    | 9. 更新部署状态       |                      |
  |                    |    (status=synced)   |                      |
  |                    |    (actual_version)  |                      |
```

### 5.3 配置漂移检测流程

```
Agent                 管控面                  告警引擎
  |                    |                      |
  | 1. 拉取期望状态     |                      |
  |------------------->|                      |
  |                    |                      |
  | 2. 返回期望状态     |                      |
  |<-------------------|                      |
  |    (action=sync)   |                      |
  |                    |                      |
  | 3. 检测到实际状态   |                      |
  |    与期望状态不一致  |                      |
  |                    |                      |
  | 4. 上报同步结果     |                      |
  |------------------->|                      |
  |    (status=drifted)|                      |
  |    (drift_detail)  |                      |
  |                    |                      |
  |                    | 5. 更新部署状态       |
  |                    |    (status=drifted)   |
  |                    |                      |
  |                    |    6. 定时评估规则     |
  |                    |---------------------->|
  |                    |                      |
  |                    |    7. 检测到 drifted  |
  |                    |    状态的部署          |
  |                    |<----------------------|
  |                    |                      |
  |                    | 8. 创建告警记录       |
  |                    |    (status=firing)    |
```

### 5.4 K8s API 代理转发流程

```
管理员              前端                管控面             隧道管理器         Agent        K3s API
  |                  |                    |                   |                |             |
  | 查看集群节点      |                    |                   |                |             |
  |----------------->|                    |                   |                |             |
  |                  | 1. POST /proxy     |                   |                |             |
  |                  |------------------->|                   |                |             |
  |                  |  (method=GET,      |                   |                |             |
  |                  |   path=/api/v1/    |                   |                |             |
  |                  |   nodes)           |                   |                |             |
  |                  |                    |                   |                |             |
  |                  |                    | 2. 白名单校验      |                |             |
  |                  |                    |    GET + 允许路径  |                |             |
  |                  |                    |                   |                |             |
  |                  |                    | 3. 发送隧道请求    |                |             |
  |                  |                    |------------------->|                |             |
  |                  |                    |                   |                |             |
  |                  |                    |                   | 4. WebSocket   |             |
  |                  |                    |                   | 转发请求       |             |
  |                  |                    |                   |--------------->|             |
  |                  |                    |                   |                |             |
  |                  |                    |                   |                | 5. 请求     |
  |                  |                    |                   |                | K8s API    |
  |                  |                    |                   |                |------------>|
  |                  |                    |                   |                |             |
  |                  |                    |                   |                | 6. 响应     |
  |                  |                    |                   |                |<------------|
  |                  |                    |                   |                |             |
  |                  |                    |                   | 7. WebSocket   |             |
  |                  |                    |                   | 返回响应       |             |
  |                  |                    |                   |<---------------|             |
  |                  |                    |                   |                |             |
  |                  |                    | 8. 返回代理响应    |                |             |
  |                  |                    |<-------------------|                |             |
  |                  |                    |                   |                |             |
  |                  | 9. 返回节点数据     |                   |                |             |
  |                  |<-------------------|                   |                |             |
  |                  |                    |                   |                |             |
  | 10. 展示节点列表  |                    |                   |                |             |
  |<-----------------|                    |                   |                |             |
```

### 5.5 告警评估流程

```
告警引擎                          Store
   |                               |
   | 1. 每 60 秒触发评估            |
   |                               |
   | 2. 获取所有启用的告警规则       |
   |------------------------------>|
   |                               |
   | 3. 获取所有集群列表            |
   |------------------------------>|
   |                               |
   | 4. 对每条规则 × 每个集群评估   |
   |                               |
   | +-- cluster_down 类型         |
   | |   检查集群状态是否为         |
   | |   unavailable               |
   | |                             |
   | +-- cpu_high 类型             |
   | |   获取最新心跳记录           |
   | |   检查 CPU 使用率阈值        |
   | |                             |
   | +-- mem_high 类型             |
   | |   获取最新心跳记录           |
   | |   检查内存使用率阈值         |
   | |                             |
   | +-- drift_detected 类型       |
   | |   获取集群部署列表           |
   | |   检查是否存在 drifted 状态  |
   | |                             |
   | +-- license_expiring 类型     |
   |     获取许可证信息             |
   |     检查过期天数阈值           |
   |                               |
   | 5. 条件满足 → 触发告警         |
   |    检查是否已有 firing 记录    |
   |    避免重复触发                |
   |------------------------------>|
   |    创建 AlertRecord            |
   |    (status=firing)            |
   |                               |
   | 6. 条件不满足 → 恢复告警       |
   |    查找 firing 状态的记录      |
   |------------------------------>|
   |    更新 AlertRecord            |
   |    (status=resolved)          |
```

---

## 6. 安全设计

### 6.1 认证体系

系统采用双轨认证机制，针对不同访问主体使用不同的认证方式：

```
+----------------------------------------------+
|                  认证体系                      |
|                                              |
|  用户认证（JWT）          代理认证（Token）    |
|  +----------------+      +----------------+  |
|  | 登录获取 Token  |      | 注册时生成      |  |
|  | Bearer 认证     |      | X-Cluster-Token | |
|  | 24小时有效期    |      | 支持轮换        |  |
|  | 含角色信息      |      | 数据库校验      |  |
|  +----------------+      +----------------+  |
|         |                        |            |
|         v                        v            |
|  +------------------------------------------+ |
|  |              访问控制                      | |
|  |  admin  → 全部操作权限                     | |
|  |  viewer → 只读权限                         | |
|  |  Agent  → 心跳/同步/隧道（仅限本集群）      | |
|  +------------------------------------------+ |
+----------------------------------------------+
```

### 6.2 敏感数据保护

| 数据 | 保护措施 |
|------|----------|
| 用户密码 | bcrypt 哈希存储，JSON 响应中通过 `json:"-"` 标签隐藏 |
| 集群 Token | JSON 响应中通过 `json:"-"` 标签隐藏，仅注册和轮换时返回 |
| 许可证密钥 | JSON 响应中通过 `json:"-"` 标签隐藏 |
| JWT 签名密钥 | 通过 `JWT_SECRET` 环境变量配置，未配置时自动生成（重启失效） |

### 6.3 K8s API 代理安全

- **白名单模式**：仅允许预定义的安全路径，默认拒绝所有未授权路径
- **只读限制**：仅允许 GET 请求，禁止 POST/PUT/DELETE/PATH 等写操作
- **隧道超时**：代理请求超时时间 30 秒，防止长时间阻塞
- **连接校验**：WebSocket 连接时校验 Origin 头，防止跨站攻击

### 6.4 CORS 安全

- 默认仅允许 `localhost:3000` 和 `localhost:3001` 来源
- 可通过 `CORS_ORIGINS` 环境变量配置允许的来源列表
- 允许的请求头：`Content-Type`、`Authorization`、`X-Cluster-Token`

### 6.5 License 配额控制

系统在以下关键操作前执行 License 配额检查：

| 操作 | 检查项 |
|------|--------|
| 注册集群 | 集群数量是否超过 `max_clusters` |
| 创建部署 | 部署数量是否超过 `max_deployments` |
| License 有效性 | 是否过期、是否有效 |

---

## 7. 部署架构

### 7.1 单机部署模式

系统采用单机本地部署模式，所有组件运行在同一台服务器上：

```
+----------------------------------------------------------+
|                      服务器                                |
|                                                          |
|  +----------------------------------------------------+  |
|  |              KubeNexus 管控面进程                    |  |
|  |                                                    |  |
|  |  +----------+  +----------+  +----------+         |  |
|  |  | Gin HTTP |  | WebSocket|  | 后台任务  |         |  |
|  |  | 服务     |  | 隧道服务  |  | 引擎     |         |  |
|  |  +----+-----+  +----+-----+  +----+-----+         |  |
|  |       |              |             |               |  |
|  |       +------+-------+-------------+               |  |
|  |              |                                     |  |
|  |       +------+------+                              |  |
|  |       |   SQLite   |                              |  |
|  |       | kubenexus.db|                              |  |
|  |       +------------+                              |  |
|  |                                                    |  |
|  |  +----------------------------------------------+ |  |
|  |  |          前端静态资源                          | |  |
|  |  |          ./web/dist/                          | |  |
|  |  +----------------------------------------------+ |  |
|  +----------------------------------------------------+  |
|                                                          |
|  环境变量：                                               |
|    DB_DSN       → SQLite 数据库路径（默认: kubenexus.db） |
|    SERVER_URL   → 管控面访问地址（默认: http://localhost:8080）|
|    SERVER_PORT  → 监听端口（默认: 8080）                  |
|    JWT_SECRET   → JWT 签名密钥（生产环境必须配置）         |
|    CORS_ORIGINS → 允许的跨域来源（逗号分隔）               |
+----------------------------------------------------------+
```

### 7.2 Agent 部署模式

Agent 以 Pod 形式部署在每个 K3s 集群中：

```
+----------------------------------------------------------+
|                    K3s 集群                                |
|                                                          |
|  命名空间: kubenexus-system                               |
|                                                          |
|  +----------------------------------------------------+  |
|  |              kubenexus-agent Pod                    |  |
|  |                                                    |  |
|  |  环境变量（从 Secret 注入）：                        |  |
|  |    SERVER_URL    → 管控面访问地址                    |  |
|  |    CLUSTER_TOKEN → 集群认证令牌                      |  |
|  |    CLUSTER_ID    → 集群标识                          |  |
|  |    KUBECONFIG    → kubeconfig 路径（可选）           |  |
|  |                                                    |  |
|  |  资源配置：                                         |  |
|  |    请求: CPU 50m / 内存 64Mi                        |  |
|  |    限制: CPU 200m / 内存 256Mi                      |  |
|  |                                                    |  |
|  |  RBAC 权限：                                        |  |
|  |    ServiceAccount: kubenexus-agent                  |  |
|  |    ClusterRole: 全部资源读写权限                     |  |
|  +----------------------------------------------------+  |
+----------------------------------------------------------+
```

### 7.3 网络通信矩阵

| 源 | 目标 | 协议 | 端口 | 用途 |
|----|------|------|------|------|
| 浏览器 | 管控面 | HTTP | 8080 | Web UI 访问 |
| Agent | 管控面 | HTTP | 8080 | 心跳上报、状态同步 |
| Agent | 管控面 | WebSocket | 8080 | 隧道连接、K8s API 代理 |
| Agent | K3s API | HTTP | 6443 | 集群资源查询与操作 |
| Agent | Helm Repo | HTTPS | 443 | Chart 仓库访问 |

---

## 8. 状态机

### 8.1 集群状态机

```
                    注册集群
                       |
                       v
               +---------------+
               |  registered   |  初始状态，Agent 尚未连接
               +-------+-------+
                       |
                       | Agent 首次心跳
                       v
               +---------------+
          +--->|    active     |  正常运行，心跳正常
          |    +-------+-------+
          |            |
          |            | 心跳超时 (>90秒)
          |            v
          |    +---------------+
          |    |  unavailable  |  集群离线
          |    +-------+-------+
          |            |
          |            | Agent 恢复心跳
          +------------+
```

### 8.2 部署状态机

```
                    创建部署
                       |
                       v
               +---------------+
               |    pending    |  等待 Agent 拉取
               +-------+-------+
                       |
                       | Agent 开始调谐
                       v
               +---------------+
               |   syncing     |  正在执行 Helm 操作
               +-------+-------+
                       |
            +----------+----------+
            |                     |
            v                     v
    +---------------+     +---------------+
    |    synced     |     |    error      |  Helm 执行失败
    |  (稳定状态)   |     +---------------+
    +-------+-------+
            |
            | 检测到配置漂移
            v
    +---------------+
    |   drifted     |  实际状态与期望不一致
    +-------+-------+
            |
            | 更新部署配置
            v
    回到 pending 状态

    任何状态 --删除--> stopped（逻辑删除，Agent 执行卸载）
```

### 8.3 告警状态机

```
                    条件满足
                       |
                       v
               +---------------+
               |    firing     |  告警触发中
               +-------+-------+
                       |
            +----------+----------+
            |                     |
            | 条件不再满足         | 人工确认
            v                     v
    +---------------+     +---------------+
    |   resolved    |<----|   resolved    |
    +---------------+     +---------------+
```

---

## 9. 默认初始化数据

### 9.1 默认管理员

系统首次启动时自动创建默认管理员账号：

| 字段 | 值 |
|------|-----|
| 用户名 | admin |
| 密码 | 自动生成的 12 位随机字符串 |
| 角色 | admin |

密码在启动日志中输出，管理员需首次登录后立即修改。

### 9.2 默认许可证

| 字段 | 值 |
|------|-----|
| 产品 | kubenexus |
| 客户名称 | 默认 |
| 最大集群数 | 10 |
| 最大部署数 | 100 |
| 功能特性 | monitoring: true, tunnel: true, config_center: true |
| 有效期 | 365 天 |
| 是否有效 | true |

### 9.3 默认告警规则

| 规则名称 | 类型 | 条件 | 严重级别 |
|----------|------|------|----------|
| 集群离线 | cluster_down | status == unavailable | critical |
| CPU使用率过高 | cpu_high | cpu_usage > 80% | warning |
| 内存使用率过高 | mem_high | mem_usage > 80% | warning |
| 配置漂移检测 | drift_detected | drift == true | warning |
| License即将过期 | license_expiring | license_days < 30 | warning |

---

## 10. 项目目录结构

```
KubeNexus/
├── agent/                          集群代理
│   ├── cmd/agent/
│   │   └── main.go                 Agent 入口
│   ├── internal/agent/
│   │   └── agent.go                Agent 核心逻辑
│   └── go.mod                      Agent 依赖（零外部依赖）
│
├── server/                         管控面后端
│   ├── cmd/
│   │   ├── server/
│   │   │   └── main.go             服务入口
│   │   └── hashpw/
│   │       └── main.go             密码哈希工具
│   ├── internal/
│   │   ├── api/
│   │   │   └── api.go              API 路由与处理器
│   │   ├── middleware/
│   │   │   └── auth.go             认证与权限中间件
│   │   ├── model/
│   │   │   └── model.go            数据模型定义
│   │   ├── service/
│   │   │   ├── alert_engine.go     告警评估引擎
│   │   │   ├── application.go      应用与部署服务
│   │   │   ├── auth.go             认证服务
│   │   │   ├── cluster.go          集群服务
│   │   │   └── services.go         其他服务
│   │   ├── store/
│   │   │   └── store.go            数据访问层
│   │   └── tunnel/
│   │       └── tunnel.go           WebSocket 隧道管理
│   ├── go.mod
│   └── go.sum
│
├── web/                            管控面前端
│   ├── src/
│   │   ├── api/
│   │   │   └── index.ts            API 请求封装
│   │   ├── components/
│   │   │   └── AppLayout.tsx       应用布局
│   │   ├── contexts/
│   │   │   └── AuthContext.tsx      认证上下文
│   │   ├── pages/
│   │   │   ├── Dashboard.tsx       仪表盘
│   │   │   ├── Clusters.tsx        集群列表
│   │   │   ├── ClusterDetail.tsx   集群详情
│   │   │   ├── Applications.tsx    应用管理
│   │   │   ├── Deployments.tsx     部署管理
│   │   │   ├── Organizations.tsx   组织管理
│   │   │   ├── Alerts.tsx          告警中心
│   │   │   ├── ConfigCenter.tsx    配置中心
│   │   │   ├── Settings.tsx        系统设置
│   │   │   └── Login.tsx           登录页
│   │   ├── App.tsx                 路由配置
│   │   ├── main.tsx                入口文件
│   │   └── index.css               全局样式
│   ├── index.html
│   ├── package.json
│   ├── tsconfig.json
│   └── vite.config.ts
│
├── charts/                         Helm Chart 模板
│   ├── kubenexus-agent/            Agent 部署 Chart
│   └── saas-app/                   SaaS 应用 Chart
│
├── scripts/
│   └── install-k3s.sh              K3s 安装脚本
│
└── docs/                           文档
    └── architecture.md             架构设计文档
```
