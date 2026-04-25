# 衡牧KubeNexusK3s多集群管理系统 API 参考文档

## 概述

本文档描述衡牧KubeNexusK3s多集群管理系统的全部 API 接口。所有接口的基础路径为 `/api/v1`，数据格式统一使用 JSON。

## 认证方式

### 用户认证

通过 JWT Token 进行身份验证。在请求头中携带 Token：

```
Authorization: Bearer <JWT Token>
```

Token 有效期为 24 小时，过期后需重新登录或刷新 Token。

### Agent 认证

集群 Agent 通过集群 Token 进行身份验证。在请求头中携带：

```
X-Cluster-Token: <Cluster Token>
```

集群 Token 在注册集群时自动生成，格式为 `cn-<uuid>`。

### 角色说明

| 角色 | 说明 |
|------|------|
| `admin` | 管理员，拥有全部操作权限 |
| `viewer` | 普通用户，仅拥有只读权限 |

---

## 认证模块

### POST /auth/login

用户登录，获取 JWT Token。

**认证要求**：无

**请求体**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名 |
| password | string | 是 | 密码 |

**请求示例**：

```json
{
  "username": "admin",
  "password": "admin123"
}
```

**响应示例**：

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "username": "admin",
    "role": "admin"
  }
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 401 | 用户名或密码错误 |

---

### POST /auth/token

刷新 JWT Token，获取新的有效 Token。

**认证要求**：需用户认证

**请求体**：无

**响应示例**：

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 401 | Token 无效或已过期 |
| 500 | Token 生成失败 |

---

### GET /auth/me

获取当前登录用户信息。

**认证要求**：需用户认证

**请求参数**：无

**响应示例**：

```json
{
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "username": "admin",
  "role": "admin"
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 401 | Token 无效或已过期 |
| 404 | 用户不存在 |

---

## 仪表盘模块

### GET /dashboard/stats

获取仪表盘统计概览数据。

**认证要求**：需用户认证

**请求参数**：无

**响应示例**：

```json
{
  "total_clusters": 12,
  "active_clusters": 10,
  "unavailable_clusters": 2,
  "total_applications": 8,
  "total_deployments": 35,
  "total_organizations": 5,
  "recent_alerts": 3
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 500 | 服务器内部错误 |

---

## 集群管理模块

### POST /clusters

注册新的集群。

**认证要求**：需用户认证

**请求体**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 集群名称，需唯一 |
| display_name | string | 否 | 集群显示名称 |
| org_id | string | 否 | 所属组织 ID |
| region | string | 否 | 集群所在区域 |
| labels | map[string]string | 否 | 集群标签键值对 |

**请求示例**：

```json
{
  "name": "prod-cluster-01",
  "display_name": "生产集群01",
  "org_id": "org-001",
  "region": "cn-east-1",
  "labels": {
    "env": "production",
    "team": "backend"
  }
}
```

**响应示例**：

```json
{
  "id": "c1d2e3f4-a5b6-7890-cdef-1234567890ab",
  "name": "prod-cluster-01",
  "display_name": "生产集群01",
  "status": "registered",
  "endpoint": "",
  "version": "",
  "node_count": 0,
  "cpu_capacity": "",
  "mem_capacity": "",
  "labels": {
    "env": "production",
    "team": "backend"
  },
  "region": "cn-east-1",
  "org_id": "org-001",
  "org_name": "研发部",
  "last_heartbeat": "0001-01-01T00:00:00Z",
  "ws_connected": false,
  "created_at": "2026-04-26T10:00:00Z",
  "updated_at": "2026-04-26T10:00:00Z"
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 500 | License 不存在、已过期或集群数量已达上限 |

---

### GET /clusters

列出所有集群。

**认证要求**：需用户认证

**请求参数**：无

**响应示例**：

```json
{
  "items": [
    {
      "id": "c1d2e3f4-a5b6-7890-cdef-1234567890ab",
      "name": "prod-cluster-01",
      "display_name": "生产集群01",
      "status": "active",
      "endpoint": "https://10.0.0.1:6443",
      "version": "v1.28.4+k3s1",
      "node_count": 3,
      "cpu_capacity": "12",
      "mem_capacity": "48Gi",
      "labels": {
        "env": "production"
      },
      "region": "cn-east-1",
      "org_id": "org-001",
      "org_name": "研发部",
      "last_heartbeat": "2026-04-26T10:30:00Z",
      "ws_connected": true,
      "created_at": "2026-04-26T10:00:00Z",
      "updated_at": "2026-04-26T10:30:00Z"
    }
  ]
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 500 | 服务器内部错误 |

---

### GET /clusters/:id

获取指定集群的详细信息。

**认证要求**：需用户认证

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 集群 ID |

**响应示例**：

```json
{
  "id": "c1d2e3f4-a5b6-7890-cdef-1234567890ab",
  "name": "prod-cluster-01",
  "display_name": "生产集群01",
  "status": "active",
  "endpoint": "https://10.0.0.1:6443",
  "version": "v1.28.4+k3s1",
  "node_count": 3,
  "cpu_capacity": "12",
  "mem_capacity": "48Gi",
  "labels": {
    "env": "production"
  },
  "region": "cn-east-1",
  "org_id": "org-001",
  "org_name": "研发部",
  "last_heartbeat": "2026-04-26T10:30:00Z",
  "ws_connected": true,
  "created_at": "2026-04-26T10:00:00Z",
  "updated_at": "2026-04-26T10:30:00Z"
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 404 | 集群不存在 |

---

### DELETE /clusters/:id

删除指定集群。

**认证要求**：需管理员

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 集群 ID |

**响应示例**：

```json
{
  "message": "deleted"
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 403 | 非管理员无权操作 |
| 500 | 服务器内部错误 |

---

### PUT /clusters/:id/labels

更新指定集群的标签。

**认证要求**：需用户认证

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 集群 ID |

**请求体**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| labels | map[string]string | 是 | 新的标签键值对，会整体替换原有标签 |

**请求示例**：

```json
{
  "labels": {
    "env": "staging",
    "team": "devops",
    "priority": "high"
  }
}
```

**响应示例**：

```json
{
  "id": "c1d2e3f4-a5b6-7890-cdef-1234567890ab",
  "name": "prod-cluster-01",
  "display_name": "生产集群01",
  "status": "active",
  "labels": {
    "env": "staging",
    "team": "devops",
    "priority": "high"
  },
  "region": "cn-east-1",
  "org_id": "org-001",
  "org_name": "研发部",
  "last_heartbeat": "2026-04-26T10:30:00Z",
  "ws_connected": true,
  "created_at": "2026-04-26T10:00:00Z",
  "updated_at": "2026-04-26T10:35:00Z"
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 500 | 服务器内部错误 |

---

### POST /clusters/:id/token/rotate

轮换指定集群的 Agent Token。轮换后旧 Token 立即失效，需使用新 Token 重新连接 Agent。

**认证要求**：需管理员

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 集群 ID |

**请求体**：无

**响应示例**：

```json
{
  "message": "token rotated",
  "token": "cn-f1e2d3c4-b5a6-7890-fedc-ba0987654321"
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 403 | 非管理员无权操作 |
| 500 | 服务器内部错误 |

---

### GET /clusters/:id/install-script

获取指定集群的 Agent 安装脚本。

**认证要求**：需用户认证

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 集群 ID |

**响应**：返回 Shell 脚本内容，Content-Type 为 `text/x-shellscript`。

**响应示例**：

```sh
#!/bin/sh
set -e

echo "=== 衡牧KubeNexusK3s多集群管理系统 Agent Installer ==="

if ! command -v k3s >/dev/null 2>&1; then
    echo "Installing K3s..."
    curl -sfL https://get.k3s.io | sh -
    echo "K3s installed successfully."
else
    echo "K3s already installed, skipping."
fi

echo "Deploying 衡牧KubeNexusK3s多集群管理系统 Agent..."
k3s kubectl apply -f "https://kubenexus.example.com/api/v1/clusters/c1d2e3f4/registration.yaml"

echo "Agent deployed. Waiting for registration..."
echo "You can check status with: k3s kubectl get pods -n kubenexus-system"
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 404 | 集群不存在 |

---

### GET /clusters/:id/registration.yaml

获取指定集群的 Agent 注册 YAML，包含 Secret、ServiceAccount、ClusterRole 及 Deployment 等资源定义。

**认证要求**：需用户认证

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 集群 ID |

**响应**：返回 YAML 内容，Content-Type 为 `application/yaml`。

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 404 | 集群不存在 |

---

### POST /clusters/:id/heartbeat

Agent 上报心跳数据。系统根据心跳时间判断集群状态，超过 90 秒未收到心跳的集群将被标记为 `unavailable`。

**认证要求**：需 Agent 认证

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 集群 ID |

**请求体**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| token | string | 是 | 集群 Token |
| node_count | int | 否 | 节点数量 |
| cpu_usage | float64 | 否 | CPU 使用率（0-100） |
| mem_usage | float64 | 否 | 内存使用率（0-100） |
| pod_count | int | 否 | Pod 数量 |
| version | string | 否 | K3s 版本号 |
| cpu_capacity | string | 否 | CPU 总容量 |
| mem_capacity | string | 否 | 内存总容量 |
| info | string | 否 | 附加信息（JSON 格式） |

**请求示例**：

```json
{
  "token": "cn-a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "node_count": 3,
  "cpu_usage": 45.2,
  "mem_usage": 62.8,
  "pod_count": 28,
  "version": "v1.28.4+k3s1",
  "cpu_capacity": "12",
  "mem_capacity": "48Gi",
  "info": "{\"kubernetes_version\":\"v1.28.4+k3s1\"}"
}
```

**响应示例**：

```json
{
  "message": "ok"
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 401 | 集群 Token 无效 |
| 500 | 服务器内部错误 |

---

### GET /clusters/:id/desired-state

Agent 获取集群的期望状态，包括需要部署、升级或同步的应用列表，以及需要移除的部署列表。

**认证要求**：需 Agent 认证

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 集群 ID |

**响应示例**：

```json
{
  "cluster_id": "c1d2e3f4-a5b6-7890-cdef-1234567890ab",
  "deployments": [
    {
      "deployment_id": "d1e2f3a4-b5c6-7890-defa-1234567890bc",
      "name": "nginx-ingress",
      "namespace": "ingress-nginx",
      "chart_name": "ingress-nginx",
      "chart_repo": "https://kubernetes.github.io/ingress-nginx",
      "chart_version": "4.8.3",
      "values": "{\"controller\":{\"replicaCount\":2}}",
      "action": "install"
    },
    {
      "deployment_id": "d2e3f4a5-b6c7-7890-defa-2345678901cd",
      "name": "prometheus",
      "namespace": "monitoring",
      "chart_name": "prometheus",
      "chart_repo": "https://prometheus-community.github.io/helm-charts",
      "chart_version": "25.0.0",
      "values": "{}",
      "action": "upgrade"
    }
  ],
  "removed": [
    "d3e4f5a6-b7c8-7890-defa-3456789012de"
  ]
}
```

**action 取值说明**：

| 值 | 说明 |
|----|------|
| install | 首次安装 |
| upgrade | 版本升级 |
| sync | 同步配置（版本未变） |

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 401 | 集群 Token 无效 |
| 500 | 服务器内部错误 |

---

### POST /clusters/:id/sync-result

Agent 上报同步结果，包括每个部署的实际状态和集群指标数据。

**认证要求**：需 Agent 认证

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 集群 ID |

**请求体**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| results | array | 是 | 同步结果列表 |
| results[].deployment_id | string | 是 | 部署 ID |
| results[].status | string | 是 | 同步状态：synced / error / drifted / syncing |
| results[].actual_version | string | 否 | 实际运行版本 |
| results[].actual_replicas | int | 否 | 实际副本数 |
| results[].message | string | 否 | 状态消息 |
| results[].drift_detail | string | 否 | 配置漂移详情 |
| cluster_metrics | object | 否 | 集群指标 |
| cluster_metrics.node_count | int | 否 | 节点数量 |
| cluster_metrics.cpu_usage | float64 | 否 | CPU 使用率 |
| cluster_metrics.mem_usage | float64 | 否 | 内存使用率 |
| cluster_metrics.pod_count | int | 否 | Pod 数量 |
| cluster_metrics.version | string | 否 | K3s 版本号 |

**请求示例**：

```json
{
  "results": [
    {
      "deployment_id": "d1e2f3a4-b5c6-7890-defa-1234567890bc",
      "status": "synced",
      "actual_version": "4.8.3",
      "actual_replicas": 2,
      "message": "release installed successfully"
    },
    {
      "deployment_id": "d2e3f4a5-b6c7-7890-defa-2345678901cd",
      "status": "drifted",
      "actual_version": "24.0.0",
      "actual_replicas": 1,
      "message": "configuration drift detected",
      "drift_detail": "replicas mismatch: desired 2, actual 1"
    }
  ],
  "cluster_metrics": {
    "node_count": 3,
    "cpu_usage": 45.2,
    "mem_usage": 62.8,
    "pod_count": 28,
    "version": "v1.28.4+k3s1"
  }
}
```

**响应示例**：

```json
{
  "message": "ok"
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 401 | 集群 Token 无效 |
| 500 | 服务器内部错误 |

---

### GET /clusters/:id/tunnel

建立 WebSocket 隧道连接，用于服务端与集群 Agent 之间的实时双向通信。通过隧道可代理 K8s API 请求、下发任务等。

**认证要求**：需 Agent 认证

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 集群 ID |

**协议**：WebSocket

**消息格式**：

```json
{
  "type": "heartbeat | task | tunnel_request | tunnel_response",
  "id": "消息唯一标识",
  "payload": {}
}
```

**消息类型说明**：

| 类型 | 方向 | 说明 |
|------|------|------|
| heartbeat | Agent -> 服务端 | 心跳保活 |
| task | 服务端 -> Agent | 下发任务 |
| tunnel_request | 服务端 -> Agent | 代理请求 |
| tunnel_response | Agent -> 服务端 | 代理响应 |

**tunnel_request 载荷格式**：

```json
{
  "method": "GET",
  "path": "/api/v1/nodes",
  "headers": {"Authorization": "Bearer xxx"},
  "body": ""
}
```

**tunnel_response 载荷格式**：

```json
{
  "status_code": 200,
  "headers": {"Content-Type": "application/json"},
  "body": "{\"items\":[]}"
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 401 | 集群 Token 无效 |

---

### GET /clusters/:id/metrics

获取指定集群的历史指标数据（心跳记录）。

**认证要求**：需用户认证

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 集群 ID |

**查询参数**：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| limit | int | 否 | 60 | 返回记录数量，最大 500 |

**响应示例**：

```json
{
  "items": [
    {
      "id": "h1a2b3c4-d5e6-7890-fabc-123456789012",
      "cluster_id": "c1d2e3f4-a5b6-7890-cdef-1234567890ab",
      "node_count": 3,
      "cpu_usage": 45.2,
      "mem_usage": 62.8,
      "pod_count": 28,
      "version": "v1.28.4+k3s1",
      "info": "",
      "reported_at": "2026-04-26T10:30:00Z"
    }
  ]
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 500 | 服务器内部错误 |

---

### POST /clusters/:id/proxy

通过 WebSocket 隧道代理 K8s API 请求到目标集群。仅允许 GET 方法，且路径受白名单限制。

**认证要求**：需用户认证

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 集群 ID |

**请求体**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| method | string | 是 | HTTP 方法，仅支持 GET |
| path | string | 是 | K8s API 路径 |
| headers | map[string]string | 否 | 请求头 |
| body | string | 否 | 请求体 |

**允许的 K8s API 路径白名单**：

- `/api/v1/nodes`
- `/api/v1/pods`
- `/api/v1/services`
- `/api/v1/namespaces`
- `/apis/apps/v1/deployments`
- `/api/v1/events`

**请求示例**：

```json
{
  "method": "GET",
  "path": "/api/v1/nodes",
  "headers": {},
  "body": ""
}
```

**响应示例**：

```json
{
  "status_code": 200,
  "headers": {
    "Content-Type": "application/json"
  },
  "body": "{\"kind\":\"NodeList\",\"apiVersion\":\"v1\",\"items\":[]}"
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 403 | 路径不在白名单中或方法不是 GET |
| 502 | 集群未连接或代理请求超时 |

---

### GET /clusters/:id/nodes

获取指定集群的节点列表。通过 WebSocket 隧道代理请求到集群的 K8s API。

**认证要求**：需用户认证

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 集群 ID |

**响应示例**：

```json
{
  "items": [
    {
      "metadata": {
        "name": "node-01",
        "creationTimestamp": "2026-04-20T08:00:00Z"
      },
      "status": {
        "conditions": [
          {
            "type": "Ready",
            "status": "True"
          }
        ],
        "capacity": {
          "cpu": "4",
          "memory": "16384Mi"
        }
      }
    }
  ]
}
```

**说明**：如果集群未通过 WebSocket 连接，或代理请求失败，将返回空列表。

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 404 | 集群不存在 |

---

## 应用管理模块

### POST /applications

创建新的应用。

**认证要求**：需管理员

**请求体**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 应用名称 |
| display_name | string | 否 | 显示名称 |
| description | string | 否 | 应用描述 |
| icon | string | 否 | 图标 URL |
| chart_name | string | 是 | Helm Chart 名称 |
| chart_repo | string | 否 | Helm 仓库地址 |
| chart_version | string | 否 | Chart 版本 |
| category | string | 否 | 应用分类 |
| is_saas | bool | 否 | 是否为 SaaS 应用，默认 true |
| default_values | string | 否 | 默认 Values 配置（YAML 字符串） |

**请求示例**：

```json
{
  "name": "nginx-ingress",
  "display_name": "Nginx Ingress Controller",
  "description": "Kubernetes 入口控制器",
  "chart_name": "ingress-nginx",
  "chart_repo": "https://kubernetes.github.io/ingress-nginx",
  "chart_version": "4.8.3",
  "category": "networking",
  "is_saas": true,
  "default_values": "controller:\n  replicaCount: 2\n  service:\n    type: LoadBalancer"
}
```

**响应示例**：

```json
{
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "name": "nginx-ingress",
  "display_name": "Nginx Ingress Controller",
  "description": "Kubernetes 入口控制器",
  "icon": "",
  "chart_name": "ingress-nginx",
  "chart_repo": "https://kubernetes.github.io/ingress-nginx",
  "chart_version": "4.8.3",
  "category": "networking",
  "is_saas": true,
  "default_values": "controller:\n  replicaCount: 2\n  service:\n    type: LoadBalancer",
  "created_at": "2026-04-26T10:00:00Z",
  "updated_at": "2026-04-26T10:00:00Z"
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 403 | 非管理员无权操作 |
| 500 | 服务器内部错误 |

---

### GET /applications

列出所有应用。

**认证要求**：需用户认证

**请求参数**：无

**响应示例**：

```json
{
  "items": [
    {
      "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "name": "nginx-ingress",
      "display_name": "Nginx Ingress Controller",
      "description": "Kubernetes 入口控制器",
      "icon": "",
      "chart_name": "ingress-nginx",
      "chart_repo": "https://kubernetes.github.io/ingress-nginx",
      "chart_version": "4.8.3",
      "category": "networking",
      "is_saas": true,
      "default_values": "",
      "created_at": "2026-04-26T10:00:00Z",
      "updated_at": "2026-04-26T10:00:00Z"
    }
  ]
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 500 | 服务器内部错误 |

---

### GET /applications/:id

获取指定应用的详细信息。

**认证要求**：需用户认证

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 应用 ID |

**响应示例**：

```json
{
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "name": "nginx-ingress",
  "display_name": "Nginx Ingress Controller",
  "description": "Kubernetes 入口控制器",
  "icon": "",
  "chart_name": "ingress-nginx",
  "chart_repo": "https://kubernetes.github.io/ingress-nginx",
  "chart_version": "4.8.3",
  "category": "networking",
  "is_saas": true,
  "default_values": "",
  "created_at": "2026-04-26T10:00:00Z",
  "updated_at": "2026-04-26T10:00:00Z"
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 404 | 应用不存在 |

---

### PUT /applications/:id

更新指定应用的信息。

**认证要求**：需管理员

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 应用 ID |

**请求体**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| display_name | string | 否 | 显示名称 |
| description | string | 否 | 应用描述 |
| icon | string | 否 | 图标 URL |
| chart_name | string | 否 | Helm Chart 名称 |
| chart_repo | string | 否 | Helm 仓库地址 |
| chart_version | string | 否 | Chart 版本 |
| category | string | 否 | 应用分类 |
| default_values | string | 否 | 默认 Values 配置 |

**请求示例**：

```json
{
  "display_name": "Nginx Ingress Controller v2",
  "chart_version": "4.9.0",
  "description": "更新后的入口控制器"
}
```

**响应示例**：

```json
{
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "name": "nginx-ingress",
  "display_name": "Nginx Ingress Controller v2",
  "description": "更新后的入口控制器",
  "icon": "",
  "chart_name": "ingress-nginx",
  "chart_repo": "https://kubernetes.github.io/ingress-nginx",
  "chart_version": "4.9.0",
  "category": "networking",
  "is_saas": true,
  "default_values": "",
  "created_at": "2026-04-26T10:00:00Z",
  "updated_at": "2026-04-26T10:30:00Z"
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 403 | 非管理员无权操作 |
| 500 | 服务器内部错误 |

---

### DELETE /applications/:id

删除指定应用。

**认证要求**：需管理员

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 应用 ID |

**响应示例**：

```json
{
  "message": "deleted"
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 403 | 非管理员无权操作 |
| 500 | 服务器内部错误 |

---

## 部署管理模块

### POST /deployments

创建单个部署，将应用部署到指定集群。

**认证要求**：需用户认证

**请求体**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| cluster_id | string | 是 | 目标集群 ID |
| application_id | string | 是 | 应用 ID |
| name | string | 是 | 部署名称 |
| namespace | string | 否 | 命名空间，默认 default |
| values | string | 否 | 自定义 Values 配置（YAML 字符串） |
| replicas | int | 否 | 副本数，默认 1 |
| version | string | 否 | 部署版本，默认使用应用 Chart 版本 |

**请求示例**：

```json
{
  "cluster_id": "c1d2e3f4-a5b6-7890-cdef-1234567890ab",
  "application_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "name": "nginx-ingress-prod",
  "namespace": "ingress-nginx",
  "values": "controller:\n  replicaCount: 3",
  "replicas": 3,
  "version": "4.8.3"
}
```

**响应示例**：

```json
{
  "id": "d1e2f3a4-b5c6-7890-defa-1234567890bc",
  "cluster_id": "c1d2e3f4-a5b6-7890-cdef-1234567890ab",
  "application_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "name": "nginx-ingress-prod",
  "namespace": "ingress-nginx",
  "values": "controller:\n  replicaCount: 3",
  "status": "pending",
  "actual_status": "",
  "replicas": 3,
  "version": "4.8.3",
  "actual_version": "",
  "drift_detail": "",
  "message": "",
  "last_synced": "0001-01-01T00:00:00Z",
  "created_at": "2026-04-26T10:00:00Z",
  "updated_at": "2026-04-26T10:00:00Z"
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 500 | License 不存在、已过期、部署数量达上限、集群或应用不存在 |

---

### POST /deployments/batch

批量部署应用到多个集群。支持通过集群 ID 列表或标签选择器指定目标集群。

**认证要求**：需管理员

**请求体**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| application_id | string | 是 | 应用 ID |
| name | string | 是 | 部署名称 |
| namespace | string | 否 | 命名空间，默认 default |
| cluster_ids | array[string] | 否 | 目标集群 ID 列表 |
| cluster_selector | object | 否 | 集群标签选择器 |
| cluster_selector.labels | map[string]string | 否 | 标签匹配条件 |
| values_overrides | string | 否 | 覆盖的 Values 配置（YAML 字符串） |
| replicas | int | 否 | 副本数，默认 1 |

**请求示例**：

```json
{
  "application_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "name": "monitoring-stack",
  "namespace": "monitoring",
  "cluster_selector": {
    "labels": {
      "env": "production"
    }
  },
  "cluster_ids": ["c1d2e3f4-a5b6-7890-cdef-1234567890ab"],
  "values_overrides": "prometheus:\n  retention: 30d",
  "replicas": 1
}
```

**响应示例**：

```json
{
  "items": [
    {
      "id": "d1e2f3a4-b5c6-7890-defa-1234567890bc",
      "cluster_id": "c1d2e3f4-a5b6-7890-cdef-1234567890ab",
      "application_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "name": "monitoring-stack",
      "namespace": "monitoring",
      "values": "prometheus:\n  retention: 30d",
      "status": "pending",
      "actual_status": "",
      "replicas": 1,
      "version": "25.0.0",
      "actual_version": "",
      "drift_detail": "",
      "message": "",
      "last_synced": "0001-01-01T00:00:00Z",
      "created_at": "2026-04-26T10:00:00Z",
      "updated_at": "2026-04-26T10:00:00Z"
    }
  ]
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 403 | 非管理员无权操作 |
| 500 | 应用不存在 |

---

### GET /deployments

列出部署列表。

**认证要求**：需用户认证

**查询参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| cluster_id | string | 否 | 按集群 ID 过滤 |

**响应示例**：

```json
{
  "items": [
    {
      "id": "d1e2f3a4-b5c6-7890-defa-1234567890bc",
      "cluster_id": "c1d2e3f4-a5b6-7890-cdef-1234567890ab",
      "application_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "name": "nginx-ingress-prod",
      "namespace": "ingress-nginx",
      "values": "",
      "status": "synced",
      "actual_status": "deployed",
      "replicas": 3,
      "version": "4.8.3",
      "actual_version": "4.8.3",
      "drift_detail": "",
      "message": "",
      "last_synced": "2026-04-26T10:30:00Z",
      "created_at": "2026-04-26T10:00:00Z",
      "updated_at": "2026-04-26T10:30:00Z"
    }
  ]
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 500 | 服务器内部错误 |

---

### GET /deployments/:id

获取指定部署的详细信息。

**认证要求**：需用户认证

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 部署 ID |

**响应示例**：

```json
{
  "id": "d1e2f3a4-b5c6-7890-defa-1234567890bc",
  "cluster_id": "c1d2e3f4-a5b6-7890-cdef-1234567890ab",
  "application_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "name": "nginx-ingress-prod",
  "namespace": "ingress-nginx",
  "values": "",
  "status": "synced",
  "actual_status": "deployed",
  "replicas": 3,
  "version": "4.8.3",
  "actual_version": "4.8.3",
  "drift_detail": "",
  "message": "",
  "last_synced": "2026-04-26T10:30:00Z",
  "created_at": "2026-04-26T10:00:00Z",
  "updated_at": "2026-04-26T10:30:00Z"
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 404 | 部署不存在 |

---

### PUT /deployments/:id

更新指定部署的配置。更新后部署状态将重置为 `pending`，等待 Agent 同步。

**认证要求**：需用户认证

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 部署 ID |

**请求体**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| values | string | 否 | 自定义 Values 配置 |
| replicas | int | 否 | 副本数 |
| version | string | 否 | 部署版本 |

**请求示例**：

```json
{
  "values": "controller:\n  replicaCount: 5",
  "replicas": 5,
  "version": "4.9.0"
}
```

**响应示例**：

```json
{
  "id": "d1e2f3a4-b5c6-7890-defa-1234567890bc",
  "cluster_id": "c1d2e3f4-a5b6-7890-cdef-1234567890ab",
  "application_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "name": "nginx-ingress-prod",
  "namespace": "ingress-nginx",
  "values": "controller:\n  replicaCount: 5",
  "status": "pending",
  "actual_status": "deployed",
  "replicas": 5,
  "version": "4.9.0",
  "actual_version": "4.8.3",
  "drift_detail": "",
  "message": "",
  "last_synced": "2026-04-26T10:30:00Z",
  "created_at": "2026-04-26T10:00:00Z",
  "updated_at": "2026-04-26T10:35:00Z"
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 500 | 服务器内部错误 |

---

### DELETE /deployments/:id

删除指定部署。

**认证要求**：需用户认证

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 部署 ID |

**响应示例**：

```json
{
  "message": "deleted"
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 500 | 服务器内部错误 |

---

## 组织管理模块

### POST /organizations

创建新的组织。

**认证要求**：需管理员

**请求体**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 组织名称 |
| code | string | 是 | 组织编码，需唯一 |
| contact | string | 否 | 联系人 |
| phone | string | 否 | 联系电话 |
| email | string | 否 | 邮箱 |
| type | string | 否 | 组织类型，默认 department |
| description | string | 否 | 组织描述 |

**请求示例**：

```json
{
  "name": "研发部",
  "code": "RD",
  "contact": "张三",
  "phone": "13800138000",
  "email": "rd@example.com",
  "type": "department",
  "description": "研发部门"
}
```

**响应示例**：

```json
{
  "id": "o1a2b3c4-d5e6-7890-fabc-123456789012",
  "name": "研发部",
  "code": "RD",
  "contact": "张三",
  "phone": "13800138000",
  "email": "rd@example.com",
  "type": "department",
  "description": "研发部门",
  "created_at": "2026-04-26T10:00:00Z",
  "updated_at": "2026-04-26T10:00:00Z"
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 403 | 非管理员无权操作 |
| 500 | 服务器内部错误 |

---

### GET /organizations

列出所有组织。

**认证要求**：需用户认证

**请求参数**：无

**响应示例**：

```json
{
  "items": [
    {
      "id": "o1a2b3c4-d5e6-7890-fabc-123456789012",
      "name": "研发部",
      "code": "RD",
      "contact": "张三",
      "phone": "13800138000",
      "email": "rd@example.com",
      "type": "department",
      "description": "研发部门",
      "created_at": "2026-04-26T10:00:00Z",
      "updated_at": "2026-04-26T10:00:00Z"
    }
  ]
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 500 | 服务器内部错误 |

---

### GET /organizations/:id

获取指定组织的详细信息。

**认证要求**：需用户认证

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 组织 ID |

**响应示例**：

```json
{
  "id": "o1a2b3c4-d5e6-7890-fabc-123456789012",
  "name": "研发部",
  "code": "RD",
  "contact": "张三",
  "phone": "13800138000",
  "email": "rd@example.com",
  "type": "department",
  "description": "研发部门",
  "created_at": "2026-04-26T10:00:00Z",
  "updated_at": "2026-04-26T10:00:00Z"
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 404 | 组织不存在 |

---

### PUT /organizations/:id

更新指定组织的信息。

**认证要求**：需管理员

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 组织 ID |

**请求体**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 否 | 组织名称 |
| code | string | 否 | 组织编码 |
| contact | string | 否 | 联系人 |
| phone | string | 否 | 联系电话 |
| email | string | 否 | 邮箱 |
| type | string | 否 | 组织类型 |
| description | string | 否 | 组织描述 |

**请求示例**：

```json
{
  "name": "研发中心",
  "contact": "李四",
  "description": "研发中心（原研发部）"
}
```

**响应示例**：

```json
{
  "id": "o1a2b3c4-d5e6-7890-fabc-123456789012",
  "name": "研发中心",
  "code": "RD",
  "contact": "李四",
  "phone": "13800138000",
  "email": "rd@example.com",
  "type": "department",
  "description": "研发中心（原研发部）",
  "created_at": "2026-04-26T10:00:00Z",
  "updated_at": "2026-04-26T10:30:00Z"
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 403 | 非管理员无权操作 |
| 500 | 服务器内部错误 |

---

### DELETE /organizations/:id

删除指定组织。

**认证要求**：需管理员

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 组织 ID |

**响应示例**：

```json
{
  "message": "deleted"
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 403 | 非管理员无权操作 |
| 500 | 服务器内部错误 |

---

## License 模块

### GET /license

获取当前 License 信息。

**认证要求**：需用户认证

**请求参数**：无

**响应示例**：

```json
{
  "id": "license-default",
  "product": "KubeNexus",
  "customer_name": "衡牧科技",
  "issued_at": "2026-01-01T00:00:00Z",
  "expires_at": "2027-01-01T00:00:00Z",
  "max_clusters": 20,
  "max_deployments": 200,
  "features": "proxy,tunnel,alert",
  "is_valid": true,
  "created_at": "2026-01-01T00:00:00Z"
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 404 | License 不存在 |

---

### POST /license/activate

激活或更新 License。

**认证要求**：需管理员

**请求体**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| key | string | 是 | License 密钥 |
| product | string | 否 | 产品名称 |
| customer_name | string | 否 | 客户名称 |
| max_clusters | int | 否 | 最大集群数 |
| max_deployments | int | 否 | 最大部署数 |
| features | string | 否 | 功能特性（逗号分隔） |
| expires_at | datetime | 否 | 过期时间 |
| issued_at | datetime | 否 | 签发时间 |

**请求示例**：

```json
{
  "key": "KNX-LICENSE-XXXX-YYYY-ZZZZ",
  "product": "KubeNexus",
  "customer_name": "衡牧科技",
  "max_clusters": 50,
  "max_deployments": 500,
  "features": "proxy,tunnel,alert,batch_deploy",
  "expires_at": "2027-12-31T23:59:59Z",
  "issued_at": "2026-04-26T00:00:00Z"
}
```

**响应示例**：

```json
{
  "id": "license-default",
  "product": "KubeNexus",
  "customer_name": "衡牧科技",
  "issued_at": "2026-04-26T00:00:00Z",
  "expires_at": "2027-12-31T23:59:59Z",
  "max_clusters": 50,
  "max_deployments": 500,
  "features": "proxy,tunnel,alert,batch_deploy",
  "is_valid": true,
  "created_at": "2026-01-01T00:00:00Z"
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 403 | 非管理员无权操作 |
| 500 | 服务器内部错误 |

---

### GET /license/quota

获取 License 配额使用情况。

**认证要求**：需用户认证

**请求参数**：无

**响应示例**：

```json
{
  "clusters": {
    "current": 12,
    "max": 20
  },
  "deployments": {
    "current": 35,
    "max": 200
  }
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 500 | 服务器内部错误 |

---

## 告警模块

### GET /alerts/rules

列出所有告警规则。

**认证要求**：需用户认证

**请求参数**：无

**响应示例**：

```json
{
  "items": [
    {
      "id": "r1a2b3c4-d5e6-7890-fabc-123456789012",
      "name": "节点离线告警",
      "type": "node",
      "condition": "node_count < expected_count",
      "severity": "critical",
      "enabled": true,
      "notify_channels": "webhook,email",
      "last_triggered": "2026-04-26T09:00:00Z",
      "created_at": "2026-04-20T10:00:00Z",
      "updated_at": "2026-04-26T09:00:00Z"
    }
  ]
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 500 | 服务器内部错误 |

---

### POST /alerts/rules

创建告警规则。

**认证要求**：需管理员

**请求体**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 规则名称 |
| type | string | 是 | 告警类型 |
| condition | string | 是 | 触发条件表达式 |
| severity | string | 否 | 严重级别，默认 warning。可选：info / warning / critical |
| enabled | bool | 否 | 是否启用 |
| notify_channels | string | 否 | 通知渠道（逗号分隔） |

**请求示例**：

```json
{
  "name": "CPU 使用率过高",
  "type": "resource",
  "condition": "cpu_usage > 80",
  "severity": "warning",
  "enabled": true,
  "notify_channels": "webhook,email"
}
```

**响应示例**：

```json
{
  "id": "r2a3b4c5-d6e7-7890-fabc-234567890123",
  "name": "CPU 使用率过高",
  "type": "resource",
  "condition": "cpu_usage > 80",
  "severity": "warning",
  "enabled": true,
  "notify_channels": "webhook,email",
  "last_triggered": "0001-01-01T00:00:00Z",
  "created_at": "2026-04-26T10:00:00Z",
  "updated_at": "2026-04-26T10:00:00Z"
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 403 | 非管理员无权操作 |
| 500 | 服务器内部错误 |

---

### PUT /alerts/rules/:id

更新告警规则。

**认证要求**：需管理员

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 规则 ID |

**请求体**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 否 | 规则名称 |
| type | string | 否 | 告警类型 |
| condition | string | 否 | 触发条件表达式 |
| severity | string | 否 | 严重级别 |
| enabled | bool | 否 | 是否启用 |
| notify_channels | string | 否 | 通知渠道 |

**请求示例**：

```json
{
  "severity": "critical",
  "condition": "cpu_usage > 90",
  "enabled": true
}
```

**响应示例**：

```json
{
  "id": "r2a3b4c5-d6e7-7890-fabc-234567890123",
  "name": "CPU 使用率过高",
  "type": "resource",
  "condition": "cpu_usage > 90",
  "severity": "critical",
  "enabled": true,
  "notify_channels": "webhook,email",
  "last_triggered": "0001-01-01T00:00:00Z",
  "created_at": "2026-04-26T10:00:00Z",
  "updated_at": "2026-04-26T10:30:00Z"
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 403 | 非管理员无权操作 |
| 500 | 服务器内部错误 |

---

### DELETE /alerts/rules/:id

删除告警规则。

**认证要求**：需管理员

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 规则 ID |

**响应示例**：

```json
{
  "message": "deleted"
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 403 | 非管理员无权操作 |
| 500 | 服务器内部错误 |

---

### GET /alerts/records

列出告警记录。

**认证要求**：需用户认证

**查询参数**：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| cluster_id | string | 否 | - | 按集群 ID 过滤 |
| status | string | 否 | - | 按状态过滤。可选：firing / resolved |
| limit | int | 否 | 50 | 返回记录数量，最大 500 |

**响应示例**：

```json
{
  "items": [
    {
      "id": "ar1a2b3c4-d5e6-7890-fabc-123456789012",
      "rule_id": "r1a2b3c4-d5e6-7890-fabc-123456789012",
      "rule_name": "节点离线告警",
      "cluster_id": "c1d2e3f4-a5b6-7890-cdef-1234567890ab",
      "severity": "critical",
      "message": "集群节点数量低于预期",
      "status": "firing",
      "triggered_at": "2026-04-26T09:00:00Z",
      "resolved_at": "0001-01-01T00:00:00Z"
    }
  ]
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 500 | 服务器内部错误 |

---

### PUT /alerts/records/:id/acknowledge

确认告警记录，将状态从 `firing` 变更为 `resolved`。

**认证要求**：需管理员

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 告警记录 ID |

**请求体**：无

**响应示例**：

```json
{
  "message": "acknowledged"
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 403 | 非管理员无权操作 |
| 404 | 告警记录不存在 |

---

## 配置中心模块

### GET /configs

列出配置模板。

**认证要求**：需用户认证

**查询参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| org_id | string | 否 | 按组织 ID 过滤 |

**响应示例**：

```json
{
  "items": [
    {
      "id": "cfg1a2b3c4-d5e6-7890-fabc-123456789012",
      "name": "生产环境默认配置",
      "org_id": "o1a2b3c4-d5e6-7890-fabc-123456789012",
      "application_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "values": "replicaCount: 3\nresources:\n  limits:\n    cpu: 500m\n    memory: 512Mi",
      "description": "生产环境标准配置模板",
      "created_at": "2026-04-26T10:00:00Z",
      "updated_at": "2026-04-26T10:00:00Z"
    }
  ]
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 500 | 服务器内部错误 |

---

### POST /configs

创建配置模板。

**认证要求**：需用户认证

**请求体**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 模板名称 |
| org_id | string | 否 | 所属组织 ID |
| application_id | string | 否 | 关联应用 ID |
| values | string | 是 | 配置值（YAML 字符串） |
| description | string | 否 | 模板描述 |

**请求示例**：

```json
{
  "name": "生产环境默认配置",
  "org_id": "o1a2b3c4-d5e6-7890-fabc-123456789012",
  "application_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "values": "replicaCount: 3\nresources:\n  limits:\n    cpu: 500m\n    memory: 512Mi",
  "description": "生产环境标准配置模板"
}
```

**响应示例**：

```json
{
  "id": "cfg1a2b3c4-d5e6-7890-fabc-123456789012",
  "name": "生产环境默认配置",
  "org_id": "o1a2b3c4-d5e6-7890-fabc-123456789012",
  "application_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "values": "replicaCount: 3\nresources:\n  limits:\n    cpu: 500m\n    memory: 512Mi",
  "description": "生产环境标准配置模板",
  "created_at": "2026-04-26T10:00:00Z",
  "updated_at": "2026-04-26T10:00:00Z"
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 500 | 服务器内部错误 |

---

### GET /configs/:id

获取指定配置模板的详细信息。

**认证要求**：需用户认证

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 配置模板 ID |

**响应示例**：

```json
{
  "id": "cfg1a2b3c4-d5e6-7890-fabc-123456789012",
  "name": "生产环境默认配置",
  "org_id": "o1a2b3c4-d5e6-7890-fabc-123456789012",
  "application_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "values": "replicaCount: 3\nresources:\n  limits:\n    cpu: 500m\n    memory: 512Mi",
  "description": "生产环境标准配置模板",
  "created_at": "2026-04-26T10:00:00Z",
  "updated_at": "2026-04-26T10:00:00Z"
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 404 | 配置模板不存在 |

---

### PUT /configs/:id

更新指定配置模板。

**认证要求**：需用户认证

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 配置模板 ID |

**请求体**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 否 | 模板名称 |
| values | string | 否 | 配置值 |
| description | string | 否 | 模板描述 |

**请求示例**：

```json
{
  "name": "生产环境高可用配置",
  "values": "replicaCount: 5\nresources:\n  limits:\n    cpu: 1000m\n    memory: 1Gi",
  "description": "生产环境高可用配置模板"
}
```

**响应示例**：

```json
{
  "id": "cfg1a2b3c4-d5e6-7890-fabc-123456789012",
  "name": "生产环境高可用配置",
  "org_id": "o1a2b3c4-d5e6-7890-fabc-123456789012",
  "application_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "values": "replicaCount: 5\nresources:\n  limits:\n    cpu: 1000m\n    memory: 1Gi",
  "description": "生产环境高可用配置模板",
  "created_at": "2026-04-26T10:00:00Z",
  "updated_at": "2026-04-26T10:30:00Z"
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 500 | 服务器内部错误 |

---

### DELETE /configs/:id

删除指定配置模板。

**认证要求**：需用户认证

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 配置模板 ID |

**响应示例**：

```json
{
  "message": "deleted"
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 500 | 服务器内部错误 |

---

## 审计日志模块

### GET /audit-logs

列出审计日志。

**认证要求**：需用户认证

**查询参数**：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| resource_type | string | 否 | - | 按资源类型过滤（如 cluster / application / deployment / organization / user / alert_rule / config_template / license） |
| username | string | 否 | - | 按用户名过滤 |
| action | string | 否 | - | 按操作类型过滤（如 create / update / delete / login / deploy / batch_deploy / rotate_token / activate / acknowledge） |
| limit | int | 否 | 100 | 返回记录数量，最大 500 |

**响应示例**：

```json
{
  "items": [
    {
      "id": "al1a2b3c4-d5e6-7890-fabc-123456789012",
      "user_id": "u1a2b3c4-d5e6-7890-fabc-123456789012",
      "username": "admin",
      "action": "create",
      "resource_type": "cluster",
      "resource_id": "c1d2e3f4-a5b6-7890-cdef-1234567890ab",
      "resource_name": "prod-cluster-01",
      "detail": "",
      "ip": "192.168.1.100",
      "created_at": "2026-04-26T10:00:00Z"
    }
  ]
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 500 | 服务器内部错误 |

---

## 用户管理模块

### POST /users

创建新用户。

**认证要求**：需管理员

**请求体**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名，需唯一 |
| password | string | 是 | 密码 |
| role | string | 否 | 角色，默认 viewer。可选：admin / viewer |

**请求示例**：

```json
{
  "username": "operator01",
  "password": "secure_password",
  "role": "viewer"
}
```

**响应示例**：

```json
{
  "id": "u2a3b4c5-d6e7-7890-fabc-234567890123",
  "username": "operator01",
  "role": "viewer",
  "created_at": "2026-04-26T10:00:00Z",
  "updated_at": "2026-04-26T10:00:00Z"
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 403 | 非管理员无权操作 |
| 500 | 服务器内部错误 |

---

### GET /users

列出所有用户。

**认证要求**：需管理员

**请求参数**：无

**响应示例**：

```json
{
  "items": [
    {
      "id": "u1a2b3c4-d5e6-7890-fabc-123456789012",
      "username": "admin",
      "role": "admin",
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-01-01T00:00:00Z"
    },
    {
      "id": "u2a3b4c5-d6e7-7890-fabc-234567890123",
      "username": "operator01",
      "role": "viewer",
      "created_at": "2026-04-26T10:00:00Z",
      "updated_at": "2026-04-26T10:00:00Z"
    }
  ]
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 403 | 非管理员无权操作 |
| 500 | 服务器内部错误 |

---

### PUT /users/:id

更新用户信息。

**认证要求**：需管理员

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 用户 ID |

**请求体**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| role | string | 否 | 角色。可选：admin / viewer |
| password | string | 否 | 新密码 |

**请求示例**：

```json
{
  "role": "admin",
  "password": "new_secure_password"
}
```

**响应示例**：

```json
{
  "id": "u2a3b4c5-d6e7-7890-fabc-234567890123",
  "username": "operator01",
  "role": "admin",
  "created_at": "2026-04-26T10:00:00Z",
  "updated_at": "2026-04-26T10:30:00Z"
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 403 | 非管理员无权操作 |
| 500 | 服务器内部错误 |

---

### DELETE /users/:id

删除用户。

**认证要求**：需管理员

**路径参数**：

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 用户 ID |

**响应示例**：

```json
{
  "message": "deleted"
}
```

**错误响应**：

| 状态码 | 说明 |
|--------|------|
| 403 | 非管理员无权操作 |
| 500 | 服务器内部错误 |

---

## 通用错误格式

所有接口在发生错误时返回统一的 JSON 格式：

```json
{
  "error": "错误描述信息"
}
```

## 状态码汇总

| 状态码 | 说明 |
|--------|------|
| 200 | 请求成功 |
| 201 | 资源创建成功 |
| 400 | 请求参数错误 |
| 401 | 未认证或认证失败 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |
| 502 | 网关错误（集群未连接或代理超时） |

## 集群状态说明

| 状态 | 说明 |
|------|------|
| registered | 已注册，Agent 尚未连接 |
| active | 正常运行，Agent 心跳正常 |
| unavailable | 不可用，超过 90 秒未收到 Agent 心跳 |

## 部署状态说明

| 状态 | 说明 |
|------|------|
| pending | 待同步，等待 Agent 拉取 |
| syncing | 同步中 |
| synced | 已同步 |
| error | 同步失败 |
| drifted | 配置漂移 |
| stopped | 已停止 |
