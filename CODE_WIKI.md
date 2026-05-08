# Nightingale (夜莺) Code Wiki

> 版本: v6 | 模块路径: `github.com/ccfos/nightingale/v6` | 语言: Go 1.25+ | 协议: Apache-2.0

---

## 目录

- [1. 项目概述](#1-项目概述)
- [2. 整体架构](#2-整体架构)
- [3. 部署模式](#3-部署模式)
- [4. 目录结构](#4-目录结构)
- [5. 核心模块详解](#5-核心模块详解)
  - [5.1 Center — 中心服务](#51-center--中心服务)
  - [5.2 Alert — 告警引擎](#52-alert--告警引擎)
  - [5.3 Pushgw — 数据推送网关](#53-pushgw--数据推送网关)
  - [5.4 AIAgent — AI 智能体](#54-aiagent--ai-智能体)
  - [5.5 Models — 数据模型层](#55-models--数据模型层)
  - [5.6 Memsto — 内存缓存层](#56-memsto--内存缓存层)
  - [5.7 Datasource — 数据源抽象层](#57-datasource--数据源抽象层)
  - [5.8 DSKit — 时序查询工具集](#58-dskit--时序查询工具集)
  - [5.9 Storage — 存储层](#59-storage--存储层)
  - [5.10 DSCache — 数据源缓存](#510-dscache--数据源缓存)
  - [5.11 Conf — 配置系统](#511-conf--配置系统)
  - [5.12 Pkg — 公共工具包](#512-pkg--公共工具包)
- [6. 关键数据流](#6-关键数据流)
- [7. 关键类与接口说明](#7-关键类与接口说明)
- [8. 依赖关系](#8-依赖关系)
- [9. 构建与运行](#9-构建与运行)
- [10. 配置说明](#10-配置说明)

---

## 1. 项目概述

Nightingale（夜莺）是一个开源的云原生监控告警系统，专注于告警引擎和告警事件的处理与分发。与 Grafana 偏重可视化不同，Nightingale 更强调告警规则的评估、告警事件的降噪、升级与协作。

**核心定位：**
- 告警规则引擎：支持 Prometheus PromQL、主机监控、Loki 日志等多种规则类型
- 告警事件处理：事件 Pipeline、静默规则、订阅规则、通知规则
- 多通道通知：内置 20+ 通知媒介（邮件、钉钉、企微、飞书、Telegram、Webhook 等）
- AI 智能体：集成 LLM（OpenAI/Claude/Gemini），支持 ReAct/Plan-ReAct 推理模式
- 多数据源：支持 Prometheus、ElasticSearch、ClickHouse、MySQL、PostgreSQL、TDengine 等
- 多协议接入：Remote Write、OpenTSDB、Datadog、Falcon 等协议

---

## 2. 整体架构

```
┌──────────────────────────────────────────────────────────────────┐
│                        Nightingale 系统架构                       │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────┐   ┌─────────────┐   ┌─────────────┐            │
│  │  Categraf   │   │  Telegraf   │   │  Grafana    │            │
│  │  (采集器)    │   │  (采集器)    │   │  Agent      │            │
│  └──────┬──────┘   └──────┬──────┘   └──────┬──────┘            │
│         │                 │                 │                    │
│         ▼                 ▼                 ▼                    │
│  ┌──────────────────────────────────────────────────┐            │
│  │              Pushgw (数据推送网关)                 │            │
│  │  Remote Write / OpenTSDB / Datadog / Falcon      │            │
│  └──────────────────────┬───────────────────────────┘            │
│                         │                                        │
│                         ▼                                        │
│  ┌──────────────────────────────────────────────────┐            │
│  │        时序数据库 (Prometheus / VictoriaMetrics)   │            │
│  └──────────────────────┬───────────────────────────┘            │
│                         │                                        │
│                         ▼                                        │
│  ┌──────────────────────────────────────────────────┐            │
│  │              Center (中心服务)                     │            │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────────────┐  │            │
│  │  │ REST API │ │ 前端静态  │ │ AI Agent / A2A   │  │            │
│  │  └──────────┘ └──────────┘ └──────────────────┘  │            │
│  └──────────────────────┬───────────────────────────┘            │
│                         │                                        │
│                         ▼                                        │
│  ┌──────────────────────────────────────────────────┐            │
│  │              Alert (告警引擎)                      │            │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────────────┐  │            │
│  │  │ 规则评估  │ │ 事件分发  │ │ 通知发送          │  │            │
│  │  │ (Eval)   │ │(Dispatch)│ │ (Sender)         │  │            │
│  │  └──────────┘ └──────────┘ └──────────────────┘  │            │
│  └──────────────────────┬───────────────────────────┘            │
│                         │                                        │
│                         ▼                                        │
│  ┌──────────────────────────────────────────────────┐            │
│  │    通知渠道 (邮件/钉钉/企微/飞书/Telegram/Webhook) │            │
│  └──────────────────────────────────────────────────┘            │
│                                                                  │
│  ┌──────────────────────────────────────────────────┐            │
│  │         基础设施 (MySQL/PostgreSQL/Redis)          │            │
│  └──────────────────────────────────────────────────┘            │
└──────────────────────────────────────────────────────────────────┘
```

---

## 3. 部署模式

Nightingale 支持三种部署模式，对应三个独立的可执行文件：

| 模式 | 可执行文件 | 入口 | 说明 |
|------|-----------|------|------|
| **中心模式** | `n9e` | `cmd/center/main.go` | 集成 Center + Alert + Pushgw，单进程全功能 |
| **告警独立模式** | `n9e-alert` | `cmd/alert/main.go` | 独立告警引擎，通过 CenterApi 获取配置 |
| **推送网关模式** | `n9e-pushgw` | `cmd/pushgw/main.go` | 独立数据推送网关 |
| **边缘模式** | `n9e-edge` | `cmd/edge/main.go` | 边缘数据中心部署，网络不佳时本地告警 |

---

## 4. 目录结构

```
nightingale/v6/
├── aiagent/            # AI 智能体模块 (LLM + ReAct + MCP + A2A)
│   ├── a2a/            # Agent-to-Agent 协议实现
│   ├── chat/           # 聊天交互处理
│   ├── llm/            # LLM 客户端 (OpenAI/Claude/Gemini)
│   ├── llmconfig/      # LLM 配置探测
│   ├── mcp/            # Model Context Protocol 客户端
│   ├── prompts/        # 系统提示词模板
│   ├── skill/          # 技能管理
│   └── tools/          # 内置工具集
├── alert/              # 告警引擎模块
│   ├── aconf/          # 告警配置
│   ├── astats/         # 告警统计
│   ├── common/         # 公共工具
│   ├── dispatch/       # 事件分发
│   ├── eval/           # 规则评估
│   ├── mute/           # 静默规则
│   ├── naming/         # HashRing 命名与主备选举
│   ├── pipeline/       # 事件 Pipeline 处理
│   ├── process/        # 告警事件处理器
│   ├── queue/          # 告警队列
│   ├── record/         # Recording Rule 调度
│   ├── router/         # 告警 HTTP 路由
│   └── sender/         # 通知发送器
├── center/             # 中心服务模块
│   ├── cconf/          # 中心配置
│   ├── cstats/         # 中心统计
│   ├── integration/    # 内置集成
│   ├── metas/          # 元数据管理
│   ├── router/         # 中心 HTTP 路由 (60+ API)
│   └── sso/            # SSO 单点登录
├── cli/                # 命令行工具
│   └── upgrade/        # 数据库升级工具
├── cmd/                # 程序入口
│   ├── a2a-cli/        # A2A 命令行客户端
│   ├── aichat-cli/     # AI 聊天命令行
│   ├── alert/          # 告警引擎入口
│   ├── center/         # 中心服务入口
│   ├── cli/            # CLI 工具入口
│   ├── edge/           # 边缘模式入口
│   └── pushgw/         # 推送网关入口
├── conf/               # 全局配置定义
├── cron/               # 定时任务 (清理通知记录等)
├── datasource/         # 数据源实现
│   ├── ck/             # ClickHouse
│   ├── doris/          # Apache Doris
│   ├── es/             # Elasticsearch
│   ├── mysql/          # MySQL
│   ├── opensearch/     # OpenSearch
│   ├── postgresql/     # PostgreSQL
│   ├── prom/           # Prometheus
│   ├── tdengine/       # TDengine
│   └── victorialogs/   # VictoriaLogs
├── docker/             # Docker 部署文件
├── dscache/            # 数据源实例缓存
├── dskit/              # 时序查询工具集
│   ├── clickhouse/     # ClickHouse 查询
│   ├── doris/          # Doris 查询
│   ├── mysql/          # MySQL 查询
│   ├── pool/           # 连接池
│   ├── postgres/       # PostgreSQL 查询
│   ├── sqlbase/        # SQL 基础抽象
│   ├── tdengine/       # TDengine 查询
│   ├── types/          # 公共类型
│   └── victorialogs/   # VictoriaLogs 查询
├── dumper/             # 调试数据导出
├── etc/                # 配置文件模板
├── front/              # 前端静态资源 (statik 嵌入)
├── memsto/             # 内存缓存模块
├── models/             # 数据模型 (GORM)
├── pkg/                # 公共工具包
├── prom/               # Prometheus 客户端封装
├── pushgw/             # 数据推送网关
│   ├── idents/         # 标识符管理
│   ├── kafka/          # Kafka 生产者
│   ├── pconf/          # 网关配置
│   ├── pstat/          # 网关统计
│   ├── router/         # 协议路由
│   └── writer/         # 数据写入器
└── storage/            # 存储层 (Redis/DB)
```

---

## 5. 核心模块详解

### 5.1 Center — 中心服务

**职责：** Nightingale 的主服务，提供 REST API、前端页面、SSO 登录、AI 助手等功能。在中心模式下，它同时内嵌 Alert 和 Pushgw 模块。

**入口函数：** `center.Initialize(configDir, cryptoKey)`

**初始化流程：**

```
Initialize()
  ├── conf.InitConfig()           # 加载配置
  ├── cconf.LoadMetricsYaml()     # 加载指标描述
  ├── cconf.LoadOpsYaml()         # 加载运维配置
  ├── logx.Init()                 # 初始化日志
  ├── i18nx.Init()                # 初始化国际化
  ├── storage.New()               # 初始化数据库 (MySQL/PostgreSQL/SQLite)
  ├── models.Migrate()            # 数据库迁移
  ├── storage.NewRedis()          # 初始化 Redis
  ├── memsto.New*Cache()          # 初始化所有内存缓存
  ├── alert.Start()               # 启动告警引擎
  ├── writer.NewWriters()         # 初始化数据写入器
  ├── centerrt.New()              # 创建中心路由
  ├── alertrt.New()               # 创建告警路由
  ├── pushgwrt.New()              # 创建推送网关路由
  ├── httpx.GinEngine()           # 创建 Gin 引擎
  └── httpx.Init()                # 启动 HTTP 服务
```

**API 路由分组（60+ 路由文件）：**

| 路由文件 | 功能 |
|---------|------|
| `router.go` | 主路由注册入口 |
| `router_alert_rule.go` | 告警规则 CRUD |
| `router_alert_cur_event.go` | 活跃告警事件查询 |
| `router_alert_his_event.go` | 历史告警事件查询 |
| `router_dashboard.go` | 仪表盘管理 |
| `router_datasource.go` | 数据源管理 |
| `router_target.go` | 监控目标管理 |
| `router_user.go` | 用户管理 |
| `router_busi_group.go` | 业务组管理 |
| `router_notify_rule.go` | 通知规则管理 |
| `router_ai_assistant.go` | AI 助手对话 |
| `router_ai_agent.go` | AI Agent 管理 |
| `router_ai_skill.go` | AI 技能管理 |
| `router_mcp_server.go` | MCP 服务器管理 |
| `router_login.go` | 登录认证 |
| `router_role.go` | 角色权限管理 |

---

### 5.2 Alert — 告警引擎

**职责：** 告警规则评估、事件生成、事件分发、通知发送。是 Nightingale 的核心引擎。

**核心子模块：**

#### 5.2.1 Eval — 规则评估

**关键结构体：** `AlertRuleWorker`

```go
type AlertRuleWorker struct {
    DatasourceId int64
    Rule         *models.AlertRule
    Processor    *process.Processor
    PromClients  *prom.PromClientMap
    Scheduler    *cron.Cron
    // ...
}
```

**评估流程：**

```
AlertRuleWorker.Eval()
  ├── 获取规则类型 (PROMETHEUS / HOST / LOKI / 其他)
  ├── 根据类型调用对应评估方法:
  │   ├── GetPromAnomalyPoint()     # Prometheus/Loki 规则
  │   ├── GetHostAnomalyPoint()     # 主机监控规则 (target_miss/offset/pct_target_miss)
  │   └── GetAnomalyPoint()         # 通用数据源规则 (MySQL/CK/ES 等)
  ├── 处理恢复事件
  └── Processor.Handle(anomalyPoints)  # 交由处理器处理
```

**规则类型：**
- **PROMETHEUS**: 使用 PromQL 查询，支持变量填充（VarFillingBeforeQuery / VarFillingAfterQuery）
- **HOST**: 主机存活检测、时钟偏移检测、存活率检测
- **LOKI**: 日志查询规则
- **通用**: 通过 DSCache 查询任意数据源，支持多查询 Join（inner/left/right/cartesian/exclude）

**Scheduler 调度：** 每条规则使用独立的 `cron.Scheduler`，支持 `@every Ns` 和标准 cron 表达式。

#### 5.2.2 Naming — 分布式调度

**关键机制：** 基于 HashRing 的告警规则分片和主备选举

```
naming.NewNaming()
  ├── HashRing: 将告警规则按哈希分配到不同节点
  ├── Heartbeat: 节点心跳上报（存入 DB）
  └── Leader: 主节点选举（用于 DingTalk 等需要单点回调的通道）
```

#### 5.2.3 Dispatch — 事件分发

**关键结构体：** `Dispatch`

**分发流程：**

```
Dispatch.HandleEventNotify(event)
  ├── HandleEventWithNotifyRule()       # 通知规则路径
  │   ├── HandleEventPipeline()         # 执行事件 Pipeline
  │   ├── ShouldSkipNotify()            # 检查是否跳过通知
  │   └── SendByNotifyRule()            # 按通知规则发送
  ├── fillUsers()                       # 填充通知用户
  ├── NotifyTarget 组装                  # 通知目标分发
  └── Send()                            # 发送通知
      ├── Sender.Send()                 # 按通道发送 (Email/Dingtalk/Wecom/...)
      ├── SendCallbacks()               # 发送回调
      ├── BatchSendWebhooks()           # 批量发送 Webhook
      ├── MayPluginNotify()             # 插件通知
      └── HandleIbex()                  # 故障自愈
```

#### 5.2.4 Sender — 通知发送

**支持的通知渠道：**

| 发送器 | 文件 | 说明 |
|--------|------|------|
| Email | `email.go` | SMTP 邮件发送 |
| Dingtalk | `dingtalk.go` | 钉钉机器人/应用 |
| Wecom | `wecom.go` | 企业微信 |
| Feishu | `feishu.go` | 飞书 |
| Lark | `lark.go` | 国际版 Lark |
| FeishuCard | `feishucard.go` | 飞书卡片消息 |
| LarkCard | `larkcard.go` | Lark 卡片消息 |
| Mm | `mm.go` | MatterMost |
| Telegram | `telegram.go` | Telegram |
| Webhook | `webhook.go` | 自定义 Webhook |
| GlobalWebhook | `global_webhook.go` | 全局 Webhook |
| Ibex | `ibex.go` | 故障自愈回调 |
| Plugin | `plugin.go` | 脚本插件通知 |

#### 5.2.5 Pipeline — 事件流水线

事件 Pipeline 支持对告警事件进行自动化处理，如追加元数据、标签重写、事件丢弃等。使用工作流引擎 `engine.WorkflowEngine` 执行。

---

### 5.3 Pushgw — 数据推送网关

**职责：** 接收多种协议的监控数据，转发到时序数据库。

**支持的协议：**

| 协议 | 路由文件 | 说明 |
|------|---------|------|
| Prometheus Remote Write | `router_remotewrite.go` | 原生 Remote Write 协议 |
| OpenTSDB | `router_opentsdb.go` | OpenTSDB HTTP 协议 |
| Datadog | `router_datadog.go` | Datadog DogStatsD 协议 |
| OpenFalcon | `router_openfalcon.go` | Falcon 推送协议 |
| Proxy Remote Write | `router_proxy_remotewrite.go` | 代理转发 Remote Write |

**数据写入流程：**

```
HTTP Request
  ├── Router.ParseRequest()       # 解析协议格式
  ├── Target.UpdateTarget()       # 更新监控目标心跳
  ├── Writer.Forward()            # 转发到后端
  │   ├── HTTP Writer             # 直接 Remote Write
  │   └── Kafka Writer            # 写入 Kafka 队列
  └── Queue management            # 内存队列管理
```

**Writer 类型：**
- `writer.go`: HTTP Remote Write 写入器，支持多个后端实例
- `kafka_writer.go`: Kafka 异步写入器
- `queue.go`: 内存队列，缓冲写入请求
- `relabel.go`: 指标 Relabel 处理

---

### 5.4 AIAgent — AI 智能体

**职责：** 提供基于 LLM 的智能告警分析、对话式运维、自动化操作能力。

#### 5.4.1 Agent 核心架构

**关键结构体：** `Agent`

```go
type Agent struct {
    cfg                *AgentConfig
    llmClient          llm.LLM
    skillRegistry      *SkillRegistry
    skillSelector      *LLMSkillSelector
    mcpClientManager   *mcp.ClientManager
    mcpServers         map[string]*mcp.ServerConfig
    externalToolHandler ExternalToolHandler
    toolDeps           *ToolDeps
}
```

**执行模式：**

| 模式 | 常量 | 说明 |
|------|------|------|
| ReAct | `AgentModeReAct` | 思考→行动→观察循环，适合简单任务 |
| Plan-ReAct | `AgentModePlanReAct` | 先规划→分步执行→综合结果，适合复杂多步任务 |

**ReAct 执行流程：**

```
Agent.Run()
  ├── selectAndLoadSkills()        # 选择并加载技能
  ├── appendSkillTools()           # 追加技能工具
  ├── appendMCPTools()             # 追加 MCP 工具
  └── executeReAct() / executePlanReAct()
      ├── callLLM()                # 调用 LLM
      ├── parseAction()            # 解析 LLM 输出的 Action
      ├── executeTool()            # 执行工具
      └── 循环直到 Final Answer 或超时
```

**Plan-ReAct 执行流程：**

```
executePlanReAct()
  ├── Phase 1: 生成执行计划 (Plan)
  │   └── LLM 生成 ExecutionPlan (多步 PlanStep)
  ├── Phase 2: 逐步执行 (Step Execution)
  │   └── 每个 PlanStep 内部运行 ReAct 循环
  └── Phase 3: 综合结果 (Synthesis)
      └── LLM 汇总所有步骤结果
```

#### 5.4.2 LLM 客户端

**接口定义：** `llm.LLM`

| 实现 | 文件 | 说明 |
|------|------|------|
| OpenAI | `llm/openai.go` | OpenAI GPT 系列，兼容 OpenAI API 格式的模型 |
| Claude | `llm/claude.go` | Anthropic Claude 系列 |
| Gemini | `llm/gemini.go` | Google Gemini 系列 |

**客户端缓存：** `llm.ClientCache` — 避免重复创建 LLM 客户端实例

#### 5.4.3 MCP (Model Context Protocol)

**关键组件：**
- `mcp.ClientManager`: MCP 客户端管理器
- `mcp.Client`: MCP 客户端，支持 SSE 和 Stdio 两种传输方式
- `mcp.ServerConfig`: MCP 服务器配置

#### 5.4.4 A2A (Agent-to-Agent)

**关键组件：**
- `a2a.Handler`: A2A 协议请求处理器
- `a2a.Executor`: A2A 任务执行器
- `a2a.Bridge`: A2A 与 Nightingale Agent 的桥接

#### 5.4.5 内置工具集

| 工具 | 文件 | 功能 |
|------|------|------|
| alert | `tools/alert.go` | 查询告警事件、告警规则 |
| alert_rule | `tools/alert_rule.go` | 管理告警规则 |
| busi_group | `tools/busi_group.go` | 查询业务组 |
| dashboard | `tools/dashboard.go` | 查询仪表盘 |
| dashboard_builder | `tools/dashboard_builder.go` | 构建仪表盘 |
| datasource | `tools/datasource.go` | 查询数据源 |
| datasource_query | `tools/datasource_query.go` | 执行数据源查询 |
| metric | `tools/metric.go` | 查询指标描述 |
| mute | `tools/mute.go` | 管理静默规则 |
| notify_rule | `tools/notify_rule.go` | 查询通知规则 |
| target | `tools/target.go` | 查询监控目标 |
| user | `tools/user.go` | 查询用户信息 |
| team | `tools/team.go` | 查询团队信息 |
| file | `tools/file.go` | 读取技能文件 |
| sql | `tools/sql.go` | 执行 SQL 查询 |
| subscribe | `tools/subscribe.go` | 管理订阅规则 |
| task_tpl | `tools/task_tpl.go` | 管理自愈模板 |

---

### 5.5 Models — 数据模型层

**职责：** 定义所有数据库模型，封装数据库 CRUD 操作。使用 GORM ORM。

**核心模型：**

| 模型 | 文件 | 说明 |
|------|------|------|
| `AlertRule` | `alert_rule.go` | 告警规则 |
| `AlertCurEvent` | `alert_cur_event.go` | 活跃告警事件 |
| `AlertHisEvent` | `alert_his_event.go` | 历史告警事件 |
| `AlertMute` | `alert_mute.go` | 静默规则 |
| `AlertSubscribe` | `alert_subscribe.go` | 订阅规则 |
| `Dashboard` | `dashboard.go` | 仪表盘 |
| `Datasource` | `datasource.go` | 数据源配置 |
| `Target` | `target.go` | 监控目标 |
| `User` | `user.go` | 用户 |
| `UserGroup` | `user_group.go` | 用户组 |
| `BusiGroup` | `busi_group.go` | 业务组 |
| `NotifyRule` | `notify_rule.go` | 通知规则 |
| `NotifyChannel` | `notify_channel.go` | 通知通道 |
| `NotifyConfig` | `notify_config.go` | 通知配置 |
| `MessageTpl` | `message_tpl.go` | 消息模板 |
| `Role` | `role.go` | 角色 |
| `RoleOperation` | `role_operation.go` | 角色操作权限 |
| `RecordingRule` | `recording_rule.go` | Recording Rule |
| `EventPipeline` | `event_pipeline.go` | 事件 Pipeline |
| `EventProcessor` | `event_processor.go` | 事件处理器 |
| `Board` | `board.go` | 监控看板 |
| `TaskTpl` | `task_tpl.go` | 自愈任务模板 |
| `AIAgent` | `ai_agent.go` | AI Agent 配置 |
| `AIAssistant` | `ai_assistant.go` | AI 助手 |
| `AISkill` | `ai_skill.go` | AI 技能 |
| `AILLMConfig` | `ai_llm_config.go` | LLM 配置 |
| `AIMCPServer` | `ai_mcp_server.go` | MCP 服务器配置 |
| `SourceToken` | `source_token.go` | API Token |
| `SSOConfig` | `sso_config.go` | SSO 配置 |

**AlertRule 关键字段：**

```go
type AlertRule struct {
    Id               int64
    GroupId          int64       // 业务组 ID
    Cate             string      // 数据源类别
    Cluster          string      // 数据源集群
    RuleConfig       string      // 规则配置 JSON
    PromEvalInterval int         // 评估间隔（秒）
    CronPattern      string      // Cron 表达式
    Severity         int         // 告警级别
    NotifyGroupsJSON []string    // 通知组
    NotifyRuleIds    []int64     // 通知规则 IDs
    // ...
}
```

---

### 5.6 Memsto — 内存缓存层

**职责：** 将数据库中的配置数据缓存到内存，避免频繁查询数据库。所有缓存均支持定时同步。

**缓存类型：**

| 缓存 | 文件 | 说明 |
|------|------|------|
| `AlertRuleCacheType` | `alert_rule_cache.go` | 告警规则缓存 |
| `AlertMuteCacheType` | `alert_mute_cache.go` | 静默规则缓存 |
| `AlertSubscribeCacheType` | `alert_subscribe_cache.go` | 订阅规则缓存 |
| `BusiGroupCacheType` | `busi_group_cache.go` | 业务组缓存 |
| `TargetCacheType` | `target_cache.go` | 监控目标缓存 |
| `DatasourceCacheType` | `datasource_cache.go` | 数据源缓存 |
| `UserCacheType` | `user_cache.go` | 用户缓存 |
| `UserGroupCacheType` | `user_group_cache.go` | 用户组缓存 |
| `UserTokenCacheType` | `user_token_cache.go` | 用户 Token 缓存 |
| `NotifyConfigCacheType` | `notify_config.go` | 通知配置缓存 |
| `NotifyRuleCacheType` | `notify_rule_cache.go` | 通知规则缓存 |
| `NotifyChannelCacheType` | `notify_channel_cache.go` | 通知通道缓存 |
| `MessageTemplateCacheType` | `message_template_cache.go` | 消息模板缓存 |
| `ConfigCacheType` | `config_cache.go` | 系统配置缓存 |
| `CvalCacheType` | `config_cval_cache.go` | 配置值缓存 |
| `TaskTplCache` | `task_tpl_cache.go` | 自愈模板缓存 |
| `EventProcessorCacheType` | `event_processor_cache.go` | 事件处理器缓存 |
| `RecordingRuleCacheType` | `recording_rule_cache.go` | Recording Rule 缓存 |
| `ESIndexPatternCacheType` | `es_index_pattern.go` | ES 索引模式缓存 |

**同步机制：** 每个缓存类型内部启动 goroutine，定期从数据库全量或增量同步数据。`SyncStats` 统计同步状态。

---

### 5.7 Datasource — 数据源抽象层

**职责：** 统一不同数据源的查询接口，支持插件式注册。

**核心接口：** `datasource.Datasource`

```go
type Datasource interface {
    Init(settings map[string]interface{}) (Datasource, error)
    InitClient() error
    Validate(ctx context.Context) error
    Equal(p Datasource) bool
    MakeLogQuery(ctx context.Context, query interface{}, eventTags []string, start, end int64) (interface{}, error)
    MakeTSQuery(ctx context.Context, query interface{}, eventTags []string, start, end int64) (interface{}, error)
    QueryData(ctx context.Context, query interface{}) ([]models.DataResp, error)
    QueryLog(ctx context.Context, query interface{}) ([]interface{}, int64, error)
    QueryMapData(ctx context.Context, query interface{}) ([]map[string]string, error)
}
```

**注册机制：** `RegisterDatasource(typ, instance)` — 每种数据源在 init() 时自注册

**支持的数据源类型：**

| ID | 类型 | 类别 | 说明 |
|----|------|------|------|
| 1 | prometheus | timeseries | Prometheus 兼容 (VictoriaMetrics 等) |
| 2 | elasticsearch | logging | Elasticsearch |
| 3 | aliyun-sls | logging | 阿里云 SLS |
| 4 | ck | timeseries | ClickHouse |
| 5 | mysql | timeseries | MySQL |
| 6 | pgsql | timeseries | PostgreSQL |
| 7 | victorialogs | logging | VictoriaLogs |

**额外支持：** TDengine、OpenSearch、Doris（通过 dskit 模块）

---

### 5.8 DSKit — 时序查询工具集

**职责：** 为各类数据库提供统一的时序数据查询抽象层，供告警评估和 Dashboard 查询使用。

**核心类型：**

```go
// dskit/types/types.go
type Timeseries interface {
    QueryData(ctx context.Context, query interface{}) ([]models.DataResp, error)
}
```

**子模块：**

| 子模块 | 说明 |
|--------|------|
| `sqlbase` | SQL 类数据源的基础查询抽象 |
| `mysql` | MySQL 时序查询实现 |
| `postgres` | PostgreSQL 时序查询实现 |
| `clickhouse` | ClickHouse 时序查询实现 |
| `doris` | Doris 时序查询实现 |
| `tdengine` | TDengine 查询实现 |
| `victorialogs` | VictoriaLogs 日志查询实现 |
| `pool` | 连接池管理 |
| `types` | 公共类型定义 (CallContext, Timeseries 等) |

---

### 5.9 Storage — 存储层

**职责：** 数据库和 Redis 的初始化与管理。

**组件：**
- `storage.New()`: 初始化 GORM 数据库连接（支持 MySQL/PostgreSQL/SQLite）
- `storage.NewRedis()`: 初始化 Redis 客户端
- `storage.PubSubBus`: 基于 Redis Pub/Sub 的消息总线，用于跨节点通信

---

### 5.10 DSCache — 数据源缓存

**职责：** 缓存已初始化的数据源客户端实例，避免每次查询都重新创建连接。

**关键方法：**
- `dscache.Init()`: 初始化缓存
- `DsCache.Get(cate, dsId)`: 获取数据源实例
- `DsCache.Set(cate, dsId, instance)`: 设置数据源实例

---

### 5.11 Conf — 配置系统

**职责：** 定义全局配置结构，加载和解析 TOML 配置文件。

**核心结构体：** `ConfigType`

```go
type ConfigType struct {
    Global    GlobalConfig
    Log       logx.Config
    HTTP      httpx.Config
    DB        ormx.DBConfig
    Redis     storage.RedisConfig
    CenterApi CenterApi
    Pushgw    pconf.Pushgw
    Alert     aconf.Alert
    Center    cconf.Center
    Ibex      Ibex
}
```

**配置加载流程：**

```
conf.InitConfig(configDir, cryptoKey)
  ├── cfg.LoadConfigByDir()       # 从目录加载 TOML 文件
  ├── Pushgw.PreCheck()           # 预检查推送网关配置
  ├── Alert.PreCheck()            # 预检查告警配置
  ├── Center.PreCheck()           # 预检查中心配置
  ├── decryptConfig()             # 解密敏感配置
  └── 自动检测 Alert.Heartbeat.IP # 自动获取出口 IP
```

**配置文件：** `etc/config.toml` (TOML 格式)

---

### 5.12 Pkg — 公共工具包

| 包 | 文件 | 功能 |
|---|------|------|
| `ginx` | `auth.go`, `funcs.go`, `param.go`, `render.go`, `errorx.go` | Gin 框架辅助（认证、参数绑定、错误处理） |
| `ormx` | `ormx.go`, `database_init.go` | GORM 数据库初始化（MySQL/PostgreSQL/SQLite） |
| `poster` | `post.go` | HTTP 客户端封装（重试、负载均衡） |
| `prom` | `reader.go`, `writer.go` | Prometheus 查询/写入客户端 |
| `ctx` | `ctx.go` | 请求上下文（封装 DB 和 IsCenter 标识） |
| `hash` | `hash.go`, `hash_fnv.go`, `hash_md5.go` | 哈希工具 |
| `logx` | `logx.go` | 日志初始化 |
| `httpx` | `httpx.go` | HTTP 服务初始化 |
| `cfg` | `cfg.go`, `scan.go` | 配置文件加载 |
| `i18nx` | `i18n.go` | 国际化 |
| `secu` | `aes.go`, `rsa.go` | 加密解密（AES/RSA） |
| `tplx` | `tplx.go`, `fns.go` | Go 模板引擎扩展 |
| `parser` | `calc.go` | 表达式计算器（用于告警触发条件判断） |
| `unit` | `unit_convert.go` | 单位转换（B/KB/MB/GB 等） |
| `promql` | `parser.go`, `promql.go` | PromQL 解析工具 |
| `fasttime` | `fasttime.go` | 高性能时间获取 |
| `flashduty` | `post.go`, `sync_user.go` | FlashDuty 集成 |
| `ldapx` | `ldapx.go`, `user_sync.go` | LDAP 用户同步 |
| `oauth2x` | `oauth2x.go` | OAuth2 认证 |
| `oidcx` | `oidc.go` | OIDC 认证 |
| `cas` | `cas.go` | CAS 认证 |
| `ibex` | `ibex.go` | Ibex 故障自愈客户端 |
| `dingtalk` | `dingtalk.go` | 钉钉 SDK 封装 |
| `feishu` | `feishu.go` | 飞书 SDK 封装 |
| `cmdx` | `cmdx.go` | 命令执行工具 |
| `version` | `version.go` | 版本信息 |
| `macros` | `macros.go` | 宏变量注册 |
| `slice` | `contains.go` | 切片工具 |
| `strx` | `verify.go` | 字符串验证 |
| `osx` | `osx.go` | 操作系统工具 |
| `loggrep` | `loggrep.go` | 日志过滤 |
| `choice` | `choice.go` | 选择工具 |
| `aop` | `log.go`, `rec.go` | AOP 切面 |
| `pool` | `pool.go` | 连接池 |

---

## 6. 关键数据流

### 6.1 告警评估数据流

```
AlertRule (DB)
    │
    ▼
AlertRuleCache (Memsto)
    │
    ▼
Scheduler (cron) ─── 定时触发 ───▶ AlertRuleWorker.Eval()
    │                                      │
    │                                      ├── 获取规则类型
    │                                      ├── 查询数据源
    │                                      │   ├── Prometheus: PromClients.GetCli().Query()
    │                                      │   ├── Host: TargetCache 查询
    │                                      │   └── 通用: DSCache.Get().QueryData()
    │                                      │
    │                                      ├── 判断触发条件
    │                                      │   ├── PromQL 结果判断
    │                                      │   └── 表达式计算 parser.Calc()
    │                                      │
    │                                      └── 生成 AnomalyPoint
    │                                              │
    ▼                                              ▼
Processor.Handle(anomalyPoints) ───▶ 生成 AlertCurEvent
    │                                              │
    │                                              ▼
    │                                     写入 DB (alert_cur_event)
    │                                              │
    ▼                                              ▼
Consumer.LoopConsume() ───▶ Dispatch.HandleEventNotify()
    │                                      │
    │                                      ├── HandleEventPipeline() (事件处理)
    │                                      ├── HandleEventWithNotifyRule() (通知规则)
    │                                      └── Send() (通知发送)
    │                                              │
    ▼                                              ▼
Sender.Send() ───▶ Email / Dingtalk / Wecom / Feishu / ...
```

### 6.2 监控数据接入流

```
Categraf / Telegraf / Datadog Agent
    │
    │  Prometheus Remote Write / OpenTSDB / Datadog / Falcon
    ▼
Pushgw Router
    │
    ├── 解析请求 (JSON / Protobuf)
    ├── 提取 ident (主机标识)
    ├── 更新 Target 心跳
    ├── Relabel 处理
    │
    ▼
Writer (内存队列)
    │
    ├── HTTP Writer ───▶ Prometheus / VictoriaMetrics (Remote Write)
    └── Kafka Writer ───▶ Kafka ───▶ 后端消费写入
```

### 6.3 AI 对话数据流

```
用户消息 (HTTP/SSE)
    │
    ▼
center/router_ai_assistant.go
    │
    ▼
aiagent.Adapter (适配层)
    │
    ├── 构建 AgentConfig
    ├── 注入 LLM Client
    ├── 注入 ToolDeps
    │
    ▼
Agent.Run()
    │
    ├── 选择技能 (SkillSelector)
    ├── 装配工具表 (builtin + skill + MCP)
    │
    ▼
ReAct Loop / Plan-ReAct Loop
    │
    ├── callLLM() ───▶ OpenAI / Claude / Gemini
    │                       │
    │                       ▼
    │                  LLM Response (Action + ActionInput)
    │
    ├── parseAction() ───▶ 解析工具调用
    │
    ├── executeTool()
    │   ├── builtin tool ───▶ 直接执行 (查询告警/仪表盘等)
    │   ├── MCP tool ───▶ MCP Client 调用外部服务
    │   ├── HTTP tool ───▶ HTTP 请求
    │   └── processor/skill tool ───▶ ExternalToolHandler
    │
    └── 循环直到 Final Answer
    │
    ▼
AgentResponse / StreamChunk (SSE)
```

---

## 7. 关键类与接口说明

### 7.1 核心接口

| 接口 | 位置 | 说明 |
|------|------|------|
| `datasource.Datasource` | `datasource/datasource.go` | 数据源统一查询接口 |
| `llm.LLM` | `aiagent/llm/llm.go` | LLM 客户端接口 |
| `sender.Sender` | `alert/sender/sender.go` | 通知发送器接口 |
| `sender.CallBacker` | `alert/sender/sender.go` | 通知回调接口 |
| `provider.NotifyChannelProvider` | `alert/sender/provider/` | 通知通道 Provider 接口 |
| `dskit/types.Timeseries` | `dskit/types/types.go` | 时序数据查询接口 |
| `storage.Redis` | `storage/redis.go` | Redis 操作接口 |
| `ExternalToolHandler` | `aiagent/types.go` | AI Agent 外部工具处理函数 |
| `BuiltinToolFunc` | `aiagent/types.go` | AI Agent 内置工具处理函数 |

### 7.2 核心结构体

| 结构体 | 位置 | 说明 |
|--------|------|------|
| `AlertRuleWorker` | `alert/eval/eval.go` | 告警规则评估 Worker |
| `Processor` | `alert/process/process.go` | 告警事件处理器 |
| `Dispatch` | `alert/dispatch/dispatch.go` | 事件分发器 |
| `Consumer` | `alert/dispatch/consume.go` | 事件消费器 |
| `Agent` | `aiagent/agent.go` | AI Agent 核心 |
| `AgentConfig` | `aiagent/types.go` | Agent 配置 |
| `AgentRequest` | `aiagent/types.go` | Agent 请求 |
| `AgentResponse` | `aiagent/types.go` | Agent 响应 |
| `ExecutionPlan` | `aiagent/types.go` | 执行计划 |
| `ConfigType` | `conf/conf.go` | 全局配置 |
| `AlertRule` | `models/alert_rule.go` | 告警规则模型 |
| `AlertCurEvent` | `models/alert_cur_event.go` | 活跃告警事件模型 |
| `Datasource` | `models/datasource.go` | 数据源模型 |
| `Target` | `models/target.go` | 监控目标模型 |
| `User` | `models/user.go` | 用户模型 |
| `NotifyRule` | `models/notify_rule.go` | 通知规则模型 |

### 7.3 关键函数

| 函数 | 位置 | 说明 |
|------|------|------|
| `center.Initialize()` | `center/center.go` | 中心服务初始化入口 |
| `alert.Initialize()` | `alert/alert.go` | 告警引擎初始化入口 |
| `alert.Start()` | `alert/alert.go` | 启动告警引擎所有组件 |
| `pushgw.Initialize()` | `pushgw/pushgw.go` | 推送网关初始化入口 |
| `Agent.Run()` | `aiagent/agent.go` | AI Agent 执行入口 |
| `AlertRuleWorker.Eval()` | `alert/eval/eval.go` | 告警规则评估 |
| `Dispatch.HandleEventNotify()` | `alert/dispatch/dispatch.go` | 事件分发处理 |
| `conf.InitConfig()` | `conf/conf.go` | 配置初始化 |
| `storage.New()` | `storage/storage.go` | 数据库初始化 |
| `memsto.New*Cache()` | `memsto/*.go` | 各类缓存初始化 |
| `datasource.RegisterDatasource()` | `datasource/datasource.go` | 数据源注册 |
| `parser.Calc()` | `pkg/parser/calc.go` | 表达式计算 |

---

## 8. 依赖关系

### 8.1 模块间依赖关系图

```
                    ┌──────────┐
                    │   cmd/   │
                    │ (入口层)  │
                    └────┬─────┘
                         │
            ┌────────────┼────────────┐
            ▼            ▼            ▼
      ┌──────────┐ ┌──────────┐ ┌──────────┐
      │  center  │ │  alert   │ │  pushgw  │
      │ (中心服务)│ │(告警引擎) │ │(推送网关) │
      └────┬─────┘ └────┬─────┘ └────┬─────┘
           │            │            │
           ├────────────┼────────────┤
           ▼            ▼            ▼
      ┌──────────┐ ┌──────────┐ ┌──────────┐
      │ aiagent  │ │  memsto  │ │  models  │
      │(AI智能体) │ │(内存缓存) │ │(数据模型) │
      └────┬─────┘ └────┬─────┘ └────┬─────┘
           │            │            │
           ▼            ▼            ▼
      ┌──────────┐ ┌──────────┐ ┌──────────┐
      │datasource│ │  dscache │ │ storage  │
      │(数据源层) │ │(DS缓存)  │ │(存储层)   │
      └────┬─────┘ └────┬─────┘ └────┬─────┘
           │            │            │
           ▼            ▼            ▼
      ┌──────────┐ ┌──────────┐ ┌──────────┐
      │  dskit   │ │   prom   │ │   pkg    │
      │(查询工具) │ │(Prom客户端)│ │(公共工具) │
      └──────────┘ └──────────┘ └──────────┘
                         │
                         ▼
                   ┌──────────┐
                   │   conf   │
                   │ (配置层)  │
                   └──────────┘
```

### 8.2 主要外部依赖

| 依赖 | 版本 | 用途 |
|------|------|------|
| `gin-gonic/gin` | v1.9.1 | HTTP Web 框架 |
| `gorm.io/gorm` | v1.25.10 | ORM 框架 |
| `redis/go-redis` | v9.0.2 | Redis 客户端 |
| `prometheus/prometheus` | v0.47.1 | Prometheus API 兼容 |
| `prometheus/client_golang` | v1.20.5 | Prometheus 客户端库 |
| `IBM/sarama` | v1.45.0 | Kafka 客户端 |
| `ClickHouse/clickhouse-go` | v2.23.2 | ClickHouse 客户端 |
| `olivere/elastic` | v7.0.32 | Elasticsearch 客户端 |
| `opensearch-project/opensearch-go` | v2.3.0 | OpenSearch 客户端 |
| `larksuite/oapi-sdk-go` | v3.5.1 | 飞书 SDK |
| `alibabacloud-go/dingtalk` | v1.6.95 | 钉钉 SDK |
| `modelcontextprotocol/go-sdk` | v1.4.0 | MCP Go SDK |
| `a2aproject/a2a-go` | v2.2.1 | A2A 协议 SDK |
| `expr-lang/expr` | v1.16.1 | 表达式引擎 |
| `dgrijalva/jwt-go` | v3.2.0 | JWT 认证 |
| `go-ldap/ldap` | v3.4.4 | LDAP 客户端 |
| `coreos/go-oidc` | v2.2.1 | OIDC 认证 |
| `robfig/cron` | (via prometheus) | Cron 调度 |
| `rakyll/statik` | v0.1.7 | 静态文件嵌入 |
| `flashcatcloud/ibex` | v1.3.6 | 故障自愈 |
| `glebarez/sqlite` | v1.11.0 | SQLite 驱动 (纯 Go) |
| `VictoriaMetrics/metricsql` | v0.81.1 | MetricsQL 解析 |
| `pingcap/tidb/pkg/parser` | (latest) | SQL 解析器 |

### 8.3 数据库依赖

| 数据库 | 用途 | 必需 |
|--------|------|------|
| MySQL / PostgreSQL / SQLite | 主数据库（存储规则、事件、用户等） | 是 |
| Redis | 缓存、会话、Pub/Sub、分布式锁 | 是 |
| Prometheus / VictoriaMetrics | 时序数据存储 | 是（作为数据源） |
| Elasticsearch / OpenSearch | 日志数据存储 | 否（可选数据源） |
| ClickHouse | 时序/日志数据存储 | 否（可选数据源） |
| Kafka | 监控数据异步写入 | 否（可选） |

---

## 9. 构建与运行

### 9.1 构建命令

```bash
# 完整构建（含前端下载）
make all

# 仅构建后端
make build

# 构建各独立组件
make build-edge      # 边缘模式
make build-alert     # 独立告警引擎
make build-pushgw    # 独立推送网关
make build-cli       # CLI 工具

# 发布构建
make release
```

### 9.2 运行方式

```bash
# 中心模式（默认，全功能）
./n9e

# 独立告警引擎
./n9e-alert

# 独立推送网关
./n9e-pushgw

# 后台运行
make run            # nohup ./n9e > n9e.log 2>&1 &
make run-alert      # nohup ./n9e-alert > n9e-alert.log 2>&1 &
make run-pushgw     # nohup ./n9e-pushgw > n9e-pushgw.log 2>&1 &
```

### 9.3 命令行参数

| 参数 | 环境变量 | 说明 |
|------|---------|------|
| `-configs` | `N9E_CONFIGS` | 配置文件目录（默认 `etc`） |
| `-crypto-key` | — | 配置文件加密密钥 |
| `-version` | — | 显示版本号 |

### 9.4 Docker 部署

项目提供多种 Docker Compose 配置：

| 文件 | 说明 |
|------|------|
| `docker/compose-host-network/docker-compose.yaml` | 主机网络模式 |
| `docker/compose-postgres/docker-compose.yaml` | PostgreSQL 后端 |
| `docker/compose-bridge/docker-compose.yaml` | 桥接网络模式 |
| `docker/compose-host-network-metric-log/docker-compose.yaml` | 含日志监控 |

### 9.5 数据库初始化

- MySQL/PostgreSQL: 使用 `docker/initsql/` 下的 SQL 脚本初始化
- SQLite: 使用 `docker/sqlite.sql` 初始化
- 程序启动时自动执行 `models.Migrate()` 进行数据库迁移

---

## 10. 配置说明

配置文件位于 `etc/config.toml`，采用 TOML 格式。主要配置段：

| 配置段 | 说明 |
|--------|------|
| `[Global]` | 全局配置（RunMode: release/debug） |
| `[Log]` | 日志配置（目录、级别、输出方式） |
| `[HTTP]` | HTTP 服务配置（端口、超时、认证等） |
| `[DB]` | 数据库配置（类型、连接串） |
| `[Redis]` | Redis 配置 |
| `[Center]` | 中心服务配置（SSO、集成等） |
| `[Alert]` | 告警引擎配置（心跳、评估间隔等） |
| `[Pushgw]` | 推送网关配置（写入后端、Kafka 等） |
| `[Ibex]` | 故障自愈配置 |

**默认端口：** 17000

**默认数据库：** SQLite（零依赖启动，生产环境建议 MySQL/PostgreSQL）

**默认登录：** root / root.2020（首次初始化时自动创建）

---

> 本文档基于 Nightingale v6 源码自动生成，最后更新时间：2026-05-08
