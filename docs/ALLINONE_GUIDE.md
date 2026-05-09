# Nightingale All-in-One 版本使用指南

Nightingale All-in-One 是一个零依赖、开箱即用的监控系统版本，将所有组件打包为单个可执行文件，内置 SQLite 数据库和内存版 Redis，适合小型环境、快速部署和测试场景。

## 主要特性

- **零依赖部署**：无需安装数据库或 Redis，解压即用
- **单文件打包**：所有组件集成在单个可执行文件中
- **内置 Web 界面**：前端页面已嵌入可执行文件
- **快速启动**：下载后运行一条命令即可启动
- **数据持久化**：使用 SQLite 存储数据，重启后数据不丢失

## 系统要求

- 操作系统：Linux（amd64/arm64）、Windows、macOS
- 内存：至少 512MB RAM
- 磁盘：至少 1GB 可用空间
- 网络：需要访问目标监控数据源（如 Prometheus）

## 快速开始

### 下载并解压

从发布页面下载对应操作系统的压缩包并解压：

```bash
# Linux/macOS
tar -xzf n9e-allinone-linux-amd64.tar.gz
cd n9e-allinone-linux-amd64

# Windows
# 使用 PowerShell 或 7-Zip 解压
```

### 启动服务

直接运行可执行文件：

```bash
# Linux/macOS
./n9e-allinone

# Windows
n9e-allinone.exe
```

服务将在 `http://localhost:17000` 启动。

### 访问 Web 界面

打开浏览器访问 `http://localhost:17000`，使用以下凭据登录：

- 用户名：`root`
- 密码：`root.2020`

## 部署配置

### 命令行参数

All-in-One 版本支持以下命令行参数：

| 参数 | 环境变量 | 默认值 | 说明 |
|------|---------|--------|------|
| `--port` | `N9E_PORT` | `17000` | HTTP 服务端口 |
| `--data-dir` | `N9E_DATA_DIR` | `./n9e-data` | 数据存储目录 |
| `--config` | `N9E_CONFIG` | 自动生成 | 配置文件路径 |
| `--password` | `N9E_PASSWORD` | `root.2020` | 管理员密码 |
| `--version` | - | - | 显示版本号 |
| `--help` | - | - | 显示帮助信息 |

### 数据目录

指定数据存储目录：

```bash
./n9e-allinone --data-dir /opt/n9e-data
```

数据目录结构：

```
n9e-data/
├── config.toml          # 配置文件
├── n9e.db             # SQLite 数据库
└── logs/              # 日志文件
```

### 自定义端口

修改服务监听端口：

```bash
./n9e-allinone --port 18000
```

## 配置说明

All-in-One 版本使用自动生成的配置文件，位于 `{data-dir}/config.toml`。

### 关键配置项

#### 数据库配置

```toml
[DB]
DBType = "sqlite"
SqliteFile = "./n9e-data/n9e.db"
MaxLifetime = 12
MaxOpenConns = 100
MaxIdleConns = 10
```

#### Redis 配置

```toml
[Redis]
Address = "./n9e-data"
RedisType = "miniredis"
DB = 0
```

#### HTTP 服务配置

```toml
[HTTP]
Host = "0.0.0.0"
Port = 17000
```

#### 告警引擎配置

```toml
[Alert]
[Alert.Heartbeat]
IP = "127.0.0.1"
Port = 17000
EngineName = "default"
```

### 手动配置文件

如需使用自定义配置，可手动创建配置文件：

```bash
# 创建配置目录
mkdir -p /opt/n9e/etc

# 创建配置文件
cat > /opt/n9e/etc/config.toml << 'EOF'
[Global]
RunMode = "release"

[Log]
Dir = "/opt/n9e/logs"
Level = "INFO"

[HTTP]
Host = "0.0.0.0"
Port = 17000

[DB]
DBType = "sqlite"
SqliteFile = "/opt/n9e/data/n9e.db"

[Redis]
Address = "/opt/n9e/data"
RedisType = "miniredis"
EOF

# 启动服务
./n9e-allinone --config /opt/n9e/etc/config.toml
```

## 后台运行

### 使用 systemd（Linux）

创建 systemd 服务文件：

```bash
sudo tee /etc/systemd/system/n9e-allinone.service << 'EOF'
[Unit]
Description=Nightingale All-in-One
After=network.target

[Service]
Type=simple
User=n9e
Group=n9e
WorkingDirectory=/opt/n9e
ExecStart=/opt/n9e/n9e-allinone --data-dir /opt/n9e/data
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# 启用并启动服务
sudo systemctl daemon-reload
sudo systemctl enable n9e-allinone
sudo systemctl start n9e-allinone

# 查看状态
sudo systemctl status n9e-allinone
```

### 使用 nohup

```bash
# 创建数据目录
mkdir -p ~/n9e-data

# 后台运行
nohup ./n9e-allinone --data-dir ~/n9e-data > n9e.log 2>&1 &

# 查看日志
tail -f ~/n9e-data/logs/n9e.log

# 查看进程
ps aux | grep n9e-allinone
```

## Docker 部署

### 使用 Docker

```bash
# 创建数据目录
mkdir -p ~/n9e-data

# 运行容器
docker run -d \
  --name n9e-allinone \
  -p 17000:17000 \
  -v ~/n9e-data:/data \
  flashcatcloud/nightingale-allinone:latest

# 查看日志
docker logs -f n9e-allinone
```

### 使用 Docker Compose

创建 `docker-compose.yml`：

```yaml
version: '3.8'

services:
  n9e:
    image: flashcatcloud/nightingale-allinone:latest
    container_name: n9e-allinone
    ports:
      - "17000:17000"
    volumes:
      - ./n9e-data:/data
    restart: unless-stopped
```

启动服务：

```bash
docker-compose up -d
```

## 数据备份与恢复

### 备份数据

All-in-One 版本的数据存储在 SQLite 数据库中：

```bash
# 停止服务
sudo systemctl stop n9e-allinone

# 备份数据目录
tar -czvf n9e-backup-$(date +%Y%m%d).tar.gz ~/n9e-data/
```

### 恢复数据

```bash
# 停止服务
sudo systemctl stop n9e-allinone

# 解压备份
tar -xzvf n9e-backup-20240101.tar.gz

# 恢复数据目录
cp -r n9e-data/* ~/n9e-data/

# 启动服务
sudo systemctl start n9e-allinone
```

## 性能优化

### 资源限制

根据服务器配置调整数据库连接池：

```toml
[DB]
MaxOpenConns = 200
MaxIdleConns = 20
```

### 日志配置

调整日志级别减少磁盘 IO：

```toml
[Log]
Level = "WARNING"  # 生产环境建议使用 WARNING
Output = "file"
```

## 故障排查

### 服务无法启动

1. 检查端口是否被占用：

```bash
netstat -tlnp | grep 17000
```

2. 检查数据目录权限：

```bash
ls -la ~/n9e-data/
```

3. 查看日志获取详细错误信息：

```bash
cat ~/n9e-data/logs/n9e.log
```

### 数据库连接失败

1. 检查 SQLite 文件是否存在：

```bash
ls -la ~/n9e-data/n9e.db
```

2. 修复数据库文件权限：

```bash
chmod 644 ~/n9e-data/n9e.db
```

### 内存不足

1. 监控内存使用：

```bash
free -h
```

2. 减少 miniredis 内存使用：

```toml
[Redis]
MaxMemory = "256mb"
```

## 升级说明

### 数据迁移

1. 停止旧版本服务：

```bash
sudo systemctl stop n9e-allinone
```

2. 备份数据目录：

```bash
cp -r ~/n9e-data ~/n9e-data.bak
```

3. 替换可执行文件：

```bash
mv n9e-allinone n9e-allinone.old
tar -xzf n9e-allinone-new.tar.gz
```

4. 启动新版本：

```bash
sudo systemctl start n9e-allinone
```

## 与标准版对比

| 特性 | All-in-One | 标准版 |
|------|-----------|--------|
| 数据库 | 内置 SQLite | 需要 MySQL/PostgreSQL |
| Redis | 内置 miniredis | 需要 Redis |
| 部署复杂度 | 低（单文件） | 高（多组件） |
| 适用场景 | 测试/小规模生产 | 大规模生产 |
| 性能 | 受限于单机资源 | 可水平扩展 |
| 数据持久性 | SQLite 文件 | 专业数据库集群 |
| 维护成本 | 低 | 高 |

## 安全建议

1. **修改默认密码**：首次使用后立即修改 `root` 用户密码
2. **限制网络访问**：配置防火墙限制对 17000 端口的访问
3. **启用 HTTPS**：生产环境建议使用反向代理启用 HTTPS
4. **定期备份**：制定数据备份策略
5. **监控资源**：监控 CPU、内存、磁盘使用情况

## 技术支持

- 文档：`https://n9e.github.io/`
- GitHub Issue：`https://github.com/ccfos/nightingale/issues`
- 社区论坛：`https://bbs.flashcat.cloud/`
- Slack：`n9e-talk.slack.com`

## 许可证

Nightingale 使用 Apache License 2.0 开源许可证。
