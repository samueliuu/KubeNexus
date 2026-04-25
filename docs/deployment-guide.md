# 衡牧KubeNexusK3s多集群管理系统 - 部署指南

## 环境要求

| 依赖 | 最低版本 | 说明 |
|------|---------|------|
| Go | 1.22+ | 后端编译 |
| Node.js | 18+ | 前端构建 |
| SQLite | 3 | 数据库（CGO 内置，无需单独安装） |

操作系统支持 Linux、macOS、Windows。生产环境推荐 Linux。

## 构建步骤

### 后端构建

```bash
cd server
go build -o kubenexus-server ./cmd/server/
```

编译产物为 `kubenexus-server` 可执行文件。

如需构建密码哈希工具：

```bash
go build -o hashpw ./cmd/hashpw/
```

### 前端构建

```bash
cd web
npm install
npm run build
```

构建产物输出到 `web/dist/` 目录。后端启动时会自动从 `./web/dist/` 提供前端静态文件。

## 环境变量配置

| 变量名 | 必填 | 默认值 | 说明 |
|--------|------|--------|------|
| `JWT_SECRET` | 生产环境必填 | 随机生成（每次重启变化） | JWT 签名密钥，未设置时服务重启后所有 Token 失效 |
| `SERVER_URL` | 否 | `http://localhost:8080` | 服务对外访问地址，用于生成 Agent 安装脚本和注册 YAML |
| `SERVER_PORT` | 否 | `8080` | 服务监听端口 |
| `DB_DSN` | 否 | `kubenexus.db` | SQLite 数据库文件路径 |
| `CORS_ORIGINS` | 否 | `http://localhost:3000,http://localhost:3001` | 允许的跨域来源，多个用逗号分隔 |

### 环境变量示例

```bash
export JWT_SECRET="your-secure-random-secret-at-least-32-chars"
export SERVER_URL="https://kubenexus.example.com"
export SERVER_PORT="8080"
export DB_DSN="/var/lib/kubenexus/kubenexus.db"
export CORS_ORIGINS="https://kubenexus.example.com"
```

## 启动服务

### 前置条件

确保前端已构建完成，`web/dist/` 目录存在且包含 `index.html` 和 `assets/`。

### 启动命令

```bash
# 设置环境变量后直接运行
./kubenexus-server
```

服务启动后日志输出示例：

```
========================================
默认管理员账号已创建
用户名: admin
密码: xxxxxxxxxxxx
请立即登录并修改密码！
========================================
衡牧KubeNexusK3s多集群管理系统 Server starting on :8080
```

### 开发模式

后端开发：

```bash
cd server
go run ./cmd/server/
```

前端开发（独立启动开发服务器，自动代理 API 请求到后端）：

```bash
cd web
npm run dev
```

前端开发服务器默认监听 `http://localhost:3000`，API 请求自动代理到 `http://localhost:8080`。

## Agent 安装

在管理界面注册集群后，可通过以下两种方式将 Agent 部署到目标 K3s 集群。

### 安装脚本方式

在集群详情页获取安装脚本，或使用 `scripts/install-k3s.sh`：

```bash
chmod +x scripts/install-k3s.sh
./scripts/install-k3s.sh <KUBENEXUS_SERVER_URL> <CLUSTER_TOKEN>
```

示例：

```bash
./scripts/install-k3s.sh https://kubenexus.example.com cn-abc123-def456
```

脚本执行以下操作：

1. 检测并安装 K3s（如未安装）
2. 等待 K3s 就绪
3. 在 `kubenexus-system` 命名空间部署 Agent（包含 Secret、ServiceAccount、ClusterRole、ClusterRoleBinding、Deployment）

### Helm 方式

使用项目提供的 Helm Chart 部署：

```bash
cd charts/kubenexus-agent

# 编辑 values.yaml 设置连接参数
# serverUrl: "https://kubenexus.example.com"
# clusterToken: "cn-abc123-def456"
# clusterId: ""

helm install kubenexus-agent . \
  --namespace kubenexus-system \
  --create-namespace \
  --set serverUrl="https://kubenexus.example.com" \
  --set clusterToken="cn-abc123-def456"
```

Helm Chart 配置项（`values.yaml`）：

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `serverUrl` | 管理服务器地址 | `https://kubenexus.example.com` |
| `clusterToken` | 集群注册 Token | 空（必填） |
| `clusterId` | 集群 ID（留空自动注册） | 空 |
| `image.repository` | Agent 镜像仓库 | `kubenexus/agent` |
| `image.tag` | Agent 镜像标签 | `latest` |
| `image.pullPolicy` | 镜像拉取策略 | `IfNotPresent` |
| `resources.limits.cpu` | CPU 限制 | `200m` |
| `resources.limits.memory` | 内存限制 | `256Mi` |
| `resources.requests.cpu` | CPU 请求 | `50m` |
| `resources.requests.memory` | 内存请求 | `64Mi` |
| `heartbeatInterval` | 心跳间隔（秒） | `30` |
| `syncInterval` | 同步间隔（秒） | `60` |

### 验证 Agent 状态

```bash
kubectl get pods -n kubenexus-system
kubectl logs -n kubenexus-system -l app=kubenexus-agent
```

Agent 正常运行后，管理界面中集群状态应变为"活跃"。

## 生产部署注意事项

### JWT_SECRET 必须设置

未设置 `JWT_SECRET` 时，系统会随机生成一个密钥，但每次服务重启后所有已颁发的 Token 将失效，所有用户需要重新登录。生产环境务必设置固定且足够复杂的密钥：

```bash
# 生成随机密钥
openssl rand -base64 48
# 或
head -c 48 /dev/urandom | base64
```

将生成的密钥设置到环境变量：

```bash
export JWT_SECRET="生成的密钥"
```

### CORS_ORIGINS 配置

生产环境必须配置 `CORS_ORIGINS` 为实际的前端访问域名，否则浏览器会拦截跨域请求：

```bash
export CORS_ORIGINS="https://kubenexus.example.com"
```

多个域名用逗号分隔：

```bash
export CORS_ORIGINS="https://kubenexus.example.com,https://kubenet.internal.company.com"
```

### 数据库备份

系统使用 SQLite 数据库，备份步骤：

```bash
# 停止服务
systemctl stop kubenexus

# 备份数据库文件
cp /var/lib/kubenexus/kubenexus.db /var/lib/kubenexus/kubenexus.db.bak.$(date +%Y%m%d%H%M%S)

# 重启服务
systemctl start kubenexus
```

建议设置定时任务自动备份：

```bash
# 每天凌晨 2 点备份
0 2 * * * cp /var/lib/kubenexus/kubenexus.db /var/backups/kubenexus/kubenexus-$(date +\%Y\%m\%d).db
```

### 反向代理配置

生产环境推荐使用 Nginx 作为反向代理，配置示例：

```nginx
server {
    listen 443 ssl http2;
    server_name kubenexus.example.com;

    ssl_certificate     /etc/ssl/certs/kubenexus.crt;
    ssl_certificate_key /etc/ssl/private/kubenexus.key;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # WebSocket 代理（Agent 隧道连接需要）
    location /api/v1/clusters/*/tunnel {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
    }
}
```

注意事项：

- 必须配置 WebSocket 代理支持，Agent 通过 WebSocket 与服务器建立隧道连接
- `proxy_read_timeout` 和 `proxy_send_timeout` 建议设置为 3600 秒以上，避免长连接被断开
- `SERVER_URL` 环境变量应设置为 `https://kubenexus.example.com`，确保 Agent 安装脚本中的地址正确

## 默认管理员账号

系统首次启动时，如果数据库中不存在任何用户，会自动创建默认管理员账号：

- **用户名**：`admin`
- **密码**：随机生成，输出到服务启动日志

```
========================================
默认管理员账号已创建
用户名: admin
密码: xxxxxxxxxxxx
请立即登录并修改密码！
========================================
```

首次登录后请立即修改默认密码。密码仅在此处输出一次，不会再次显示，请妥善保存。

如需手动创建管理员密码哈希，可使用 `hashpw` 工具：

```bash
./hashpw "新密码"
```

输出可直接用于数据库操作。
