# 衡牧KubeNexusK3s多集群管理系统

<p align="center">
  <strong>KubeNexus</strong> — 衡牧K3s多集群管理系统
</p>

<p align="center">
  <a href="./LICENSE"><img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="License: Apache 2.0"></a>
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8" alt="Go Version">
  <img src="https://img.shields.io/badge/React-18+-61DAFB" alt="React Version">
  <img src="https://img.shields.io/badge/TypeScript-5.4+-3178C6" alt="TypeScript Version">
</p>

---

## 项目简介

衡牧KubeNexusK3s多集群管理系统是一款面向企业的轻量级 K3s 多集群统一管理平台，采用**本地部署**模式。系统通过 Agent 代理架构实现集群注册、状态监控、应用分发、声明式部署等核心功能，帮助企业实现多 K3s 集群的集中管控和自动化运维。

### 核心特性

- 🚀 **多集群管理** — 支持 K3s 集群注册、状态监控、心跳检测、在线/离线状态实时展示
- 📦 **应用市场** — 提供应用模板管理，支持 Helm Chart 配置，统一管理和版本控制
- ⚙️ **声明式部署** — 期望状态与实际状态对比自动同步，支持配置漂移检测和自动修复
- 🔗 **WebSocket 隧道** — 管控面与集群之间安全隧道，支持 Kubernetes API 代理转发
- 🏢 **组织管理** — 支持部门、项目、团队等多层级组织架构
- 🔔 **监控告警** — 内置集群离线、CPU/内存过高、配置漂移、License 过期等告警规则
- 📋 **配置中心** — 配置模板管理，按组织维度管理 Helm Values 配置
- 🔑 **License 管理** — 许可证激活和配额管理，控制最大集群数和部署数
- 👥 **用户管理** — 多角色用户管理，JWT 身份认证
- 📊 **审计日志** — 自动记录用户操作行为
- 📈 **仪表盘** — 多维度统计概览

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.22+ / Gin / GORM / SQLite |
| 前端 | React 18 / TypeScript / Ant Design 5 / ProComponents / Vite |
| Agent | Go (标准库，零外部依赖) |
| 通信 | WebSocket + HTTP 双通道 |
| 认证 | JWT + Agent Token 双重认证 |

## 系统架构

```
┌─────────────────────────────────────────────┐
│              KubeNexus 管控面                 │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐    │
│  │  Web UI  │ │ REST API │ │ WebSocket│    │
│  │ (React)  │ │  (Gin)   │ │  Tunnel  │    │
│  └──────────┘ └──────────┘ └──────────┘    │
│        │            │            │           │
│  ┌─────────────────────────────────────┐    │
│  │           Service Layer             │    │
│  └─────────────────────────────────────┘    │
│        │                                    │
│  ┌─────────────────────────────────────┐    │
│  │           SQLite Database           │    │
│  └─────────────────────────────────────┘    │
└─────────────────────────────────────────────┘
         │                    │
    HTTP │             WebSocket
         │                    │
┌────────┴────────┐  ┌───────┴────────┐
│  K3s Cluster 1  │  │  K3s Cluster N │
│  ┌────────────┐ │  │ ┌────────────┐ │
│  │   Agent    │ │  │ │   Agent    │ │
│  └────────────┘ │  │ └────────────┘ │
└─────────────────┘  └────────────────┘
```

## 快速开始

### 环境要求

- Go 1.22+
- Node.js 18+
- SQLite 3

### 构建后端

```bash
cd server
go build -o kubenexus ./cmd/server
```

### 构建前端

```bash
cd web
npm install
npm run build
```

### 启动服务

```bash
# 设置环境变量
export JWT_SECRET=your-secret-key
export SERVER_URL=http://localhost:8080

# 启动后端（前端静态文件已内嵌）
./kubenexus
```

访问 `http://localhost:8080`，首次启动自动创建，密码输出到控制台日志。

### 安装 Agent

在目标 K3s 集群上安装 Agent：

```bash
# 方式一：安装脚本（在集群详情页获取）
bash <install-script>

# 方式二：Helm
helm install kubenexus-agent ./charts/kubenexus-agent \
  --set config.serverUrl=http://<管控面地址>:8080 \
  --set config.clusterToken=<集群Token>
```

## 项目结构

```
KubeNexus/
├── server/                    # 后端服务
│   ├── cmd/server/            # 入口
│   └── internal/
│       ├── api/               # API 路由与 Handler
│       ├── middleware/        # 认证中间件
│       ├── model/             # 数据模型
│       ├── service/           # 业务逻辑
│       ├── store/             # 数据访问层
│       └── tunnel/            # WebSocket 隧道
├── agent/                     # 集群 Agent
│   ├── cmd/agent/             # Agent 入口
│   └── internal/agent/        # Agent 核心逻辑
├── web/                       # 前端
│   └── src/
│       ├── api/               # API 客户端
│       ├── components/        # 公共组件
│       ├── contexts/          # React Context
│       └── pages/             # 页面组件
├── charts/                    # Helm Charts
│   ├── kubenexus-agent/       # Agent Chart
│   └── saas-app/              # SaaS 应用示例 Chart
├── scripts/                   # 运维脚本
│   └── install-k3s.sh         # K3s + Agent 安装脚本
└── docs/                      # 文档
    ├── architecture.md        # 架构设计
    ├── api-reference.md       # API参考
    ├── deployment-guide.md    # 部署指南
    ├── development-guide.md   # 开发指南
    └── changelog.md           # 版本变更日志
```

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `JWT_SECRET` | JWT 认证密钥（生产环境必须设置） | 随机生成 |
| `SERVER_URL` | 服务端对外访问地址 | `http://localhost:8080` |
| `SERVER_PORT` | 服务端监听端口 | `8080` |
| `DB_DSN` | 数据库连接字符串 | `kubenexus.db` |

## 联系方式

- 📧 邮箱：support@pricenexus.cn

## 许可证

本项目基于 [Apache License 2.0](./LICENSE) 开源。
