# Project1_Shop - 在线书店后端服务

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25-00ADD8?style=flat&logo=go" alt="Go Version"/>
  <img src="https://img.shields.io/badge/MySQL-4479A1?style=flat&logo=mysql&logoColor=white" alt="MySQL"/>
  <img src="https://img.shields.io/badge/Gin-v1.12-00A3E0?style=flat&logo=gin" alt="Gin"/>
  <img src="https://img.shields.io/badge/GORM-v1.31-228B22?style=flat&logo=gorm" alt="GORM"/>
  <img src="https://img.shields.io/badge/Redis-支持-DC382D?style=flat&logo=redis" alt="Redis"/>
  <img src="https://img.shields.io/badge/RabbitMQ-支持-FF6600?style=flat&logo=rabbitmq" alt="RabbitMQ"/>
  <img src="https://img.shields.io/badge/Swagger-文档-85EA2D?style=flat&logo=swagger" alt="Swagger"/>
</p>

基于 **Go + Gin + GORM + MySQL + Redis + RabbitMQ** 构建的在线书店全栈后端服务，提供用户认证、书籍管理、购物车、订单处理及高并发秒杀等核心功能。

---

## 目录

- [功能特性](#功能特性)
- [技术栈](#技术栈)
- [项目结构](#项目结构)
- [快速开始](#快速开始)
- [API 文档（Swagger）](#api-文档swagger)
- [接口总览](#接口总览)
  - [用户模块](#用户模块)
  - [书籍模块](#书籍模块)
  - [购物车模块](#购物车模块)
  - [订单模块](#订单模块)
  - [秒杀模块](#秒杀模块)
  - [管理员模块](#管理员模块)
- [认证机制](#认证机制)
- [错误码说明](#错误码说明)
- [配置说明](#配置说明)

---

## 功能特性

- **用户系统**：注册、登录、个人信息查看，基于 JWT 双 Token（Access Token + Refresh Token）认证，同时实现了单点登录控制（同一账号只能在一个设备登录）
- **书籍管理**：分页查询（按评分/销量）、基于RedisSearch的书名搜索、书籍详情，支持管理员 CRUD
- **评论与评分**：书籍评论（支持嵌套回复/楼中楼）、评分（加权威尔逊算法）、点赞
- **购物车**：添加、查看、修改数量、删除单件、清空购物车
- **订单系统**：创建订单、支付、取消、确认收货，订单超时自动关闭（RabbitMQ 延迟队列）
- **秒杀活动**：高并发秒杀（Redis 预减库存 + RabbitMQ 异步落库），防重复参与
- **管理员后台**：书籍增删改查、订单发货、秒杀活动管理
- **优雅关机**：支持 SIGINT/SIGTERM 信号的优雅停机
- **日志**：基于 zap + lumberjack 的结构化日志与日志切割
- **Swagger API 文档**：集成 swaggo，启动后可在线浏览完整接口文档

---

## 技术栈

| 分层     | 技术/框架                                             | 说明              |
|--------|---------------------------------------------------|-----------------|
| Web 框架 | [Gin](https://github.com/gin-gonic/gin) v1.12     | HTTP 路由与中间件     |
| ORM    | [GORM](https://gorm.io) v1.31                     | MySQL 数据库交互     |
| 缓存     | [go-redis](https://github.com/redis/go-redis) v9  | Redis 缓存与原子操作   |
| 消息队列   | [amqp091-go](https://github.com/rabbitmq/amqp091-go) | RabbitMQ 异步消息处理 |
| 认证     | [golang-jwt/jwt](https://github.com/golang-jwt/jwt) v5 | JWT 双 Token 认证  |
| 配置     | [Viper](https://github.com/spf13/viper)      | YAML 配置文件加载     |
| 日志     | [Zap](https://go.uber.org/zap) + lumberjack       | 高性能结构化日志        |
| API 文档 | [swaggo/swag](https://github.com/swaggo/swag) v1.8 | Swagger 文档自动生成  |
| 数据库    | MySQL 8.0+     | 持久化存储           |
| 搜索引擎   | RedisSearch   | 高性能搜索           |

---

## 项目结构

```
Project1_Shop/
├── main.go                    # 程序入口，初始化各组件，Swagger 全局注释
├── settings/                  # 配置加载（Viper）
├── logger/                    # 日志初始化（Zap）
├── router/
│   └── router.go              # 路由注册（含 Swagger UI 路由）
├── controllers/               # HTTP 控制器层（接口入参校验、响应）
│   ├── userHandle.go          # 用户相关接口
│   ├── bookHandle.go          # 书籍/评论相关接口
│   ├── cartHandle.go          # 购物车接口
│   ├── orderHandle.go         # 订单接口
│   ├── seckillHandle.go       # 秒杀接口
│   ├── adminHandle.go         # 管理员认证接口
│   ├── pageHandle.go          # 页面渲染接口
│   └── responseHandle.go      # 统一响应封装
├── logic/                     # 业务逻辑层
│   ├── user.go
│   ├── book.go
│   ├── cart.go
│   ├── order.go
│   ├── seckill.go
│   ├── comment.go
│   ├── score.go               # 评分逻辑（加权算法 + Worker 池）
│   └── token.go               # Token 刷新逻辑
├── dao/
│   ├── mysql/                 # MySQL DAO 层
│   └── redis/                 # Redis DAO 层
├── models/                    # 数据模型
│   ├── user.go
│   ├── book.go
│   ├── cart.go
│   ├── order.go
│   ├── seckill.go
│   ├── param.go               # 请求参数结构体
│   ├── rating.go
│   ├── comment_view.go
│   ├── faultCode.go           # 业务错误码
│   └── Event.go               # 事件常量与通道
├── pkg/
│   ├── jwt/                   # JWT 生成与解析
│   ├── md5/                   # 密码哈希
│   ├── middlewares/           # Gin 中间件（JWT 认证、管理员校验）
│   ├── mq/                    # RabbitMQ 连接、生产者、消费者
│   └── Worker/                # 评分 Worker 池
├── docs/                      # Swagger 自动生成文档（swag init 生成）
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
└── views/                     # 前端页面（HTML + 静态资源）
```

---

## 快速开始

### 前置依赖

- Go 1.25+
- MySQL 8.0+
- redis-stack-server latest
- RabbitMQ 3.x+

### 克隆与安装

```bash
git clone https://github.com/Frotten/BookShop_Project1.git
cd Project1_Shop
go mod download
```

### 配置文件

修改 `config.yaml`（或项目根目录下的配置文件）：

```yaml
app:
  port: 8080

mysql:
  host: "127.0.0.1"
  port: 3306
  user: "root"
  password: "your_password"
  dbname: "shop"

redis:
  host: "127.0.0.1"
  port: 6379
  password: ""
  db: 0

rabbitmq:
  url: "amqp://guest:guest@127.0.0.1:5672/"
```

### 启动依赖
MySQL启动
```bash
docker run -d \
  --name mysql \
  -p 3306:3306 \
  -e MYSQL_ROOT_PASSWORD=root \
  -e MYSQL_DATABASE=bookstore \
  mysql:8
```
Redis启动
```bash
docker run -d \
  --name redis \
  -p 6379:6379 \
  redis/redis-stack-server:latest
```
RabbitMQ启动
```bash
docker run -d \
  --name rabbitmq \
  -p 5672:5672 \
  -p 15672:15672 \
  rabbitmq:3-management
```

### 生成 Swagger 文档

```bash
# 安装 swag 工具（仅首次需要）
go install github.com/swaggo/swag/cmd/swag@v1.8.12

# 生成文档
swag init --generalInfo main.go --output docs
```

### 启动服务

```bash
go run main.go
```

服务默认监听 `:8080`，启动后访问：

- **Swagger UI**：`http://localhost:8080/swagger/index.html`
- **健康检查**：`http://localhost:8080/hello`

---

## API 文档（Swagger）

启动服务后，访问 [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html) 即可查看完整的交互式 API 文档。

**认证方式**：在 Swagger UI 右上角点击 **Authorize**，输入：
```
Bearer <your_access_token>
```

---

## 接口总览

> 所有接口响应均为 JSON 格式，通用响应结构如下：
>
> ```json
> {
>   "code": 1000,
>   "msg": "success",
>   "data": { ... }
> }
> ```

### 用户模块

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/api/register` | 用户注册 | 无 |
| POST | `/api/login` | 用户登录 | 无 |
| POST | `/refreshtoken` | 刷新 Access Token | Cookie |
| GET  | `/api/userInfo` | 获取当前用户信息 | JWT |
| GET  | `/api/userComments` | 获取当前用户评论列表 | JWT |
| GET  | `/api/userRatings` | 获取当前用户评分记录 | JWT |

### 书籍模块

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/api/getBooksJSON` | 获取书籍列表（第1页，按评分） | 无 |
| GET | `/api/getBooksJSON/:page` | 按评分分页获取书籍 | 无 |
| GET | `/api/getBooksBySaleJSON` | 获取书籍列表（第1页，按销量） | 无 |
| GET | `/api/getBooksBySaleJSON/:page` | 按销量分页获取书籍 | 无 |
| GET | `/api/getBookDetail/:book_id` | 根据 ID 获取书籍详情 | 无 |
| GET | `/api/getBookTitle/:title` | 根据书名搜索书籍 | 无 |
| GET | `/api/topScore` | 获取评分 Top 书籍 | 无 |
| GET | `/api/topSale` | 获取销量 Top 书籍 | 无 |
| GET | `/api/comments?book_id=` | 获取书籍评论列表 | 无 |
| POST | `/api/rateBook` | 对书籍评分 | JWT |
| POST | `/api/comment` | 发表书籍评论 | JWT |
| POST | `/api/comment/like` | 点赞评论 | JWT |

### 购物车模块

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST   | `/api/cart` | 添加书籍到购物车 | JWT |
| GET    | `/api/cart` | 获取购物车列表 | JWT |
| PUT    | `/api/cart` | 更新购物车商品数量 | JWT |
| DELETE | `/api/cart/:book_id` | 删除购物车中指定书籍 | JWT |
| DELETE | `/api/cart` | 清空购物车 | JWT |

### 订单模块

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/api/orderCreate` | 创建订单 | JWT |
| GET  | `/api/userOrders` | 获取当前用户订单列表 | JWT |
| GET  | `/api/orderDetail/:order_id` | 获取订单详情 | JWT |
| POST | `/api/orderPay` | 支付订单 | JWT |
| POST | `/api/orderCancel` | 取消订单 | JWT |
| POST | `/api/orderConfirm` | 确认收货 | JWT |

**订单状态说明**：`-1` 已删除 / `0` 未支付 / `1` 已支付待发货 / `2` 已发货待收货 / `3` 已收货

### 秒杀模块

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET  | `/api/seckill/list` | 获取进行中的秒杀活动列表 | 无 |
| GET  | `/api/seckill/:id` | 获取秒杀活动详情 | 无 |
| POST | `/api/seckill/do` | 参与秒杀抢购 | JWT |

### 管理员模块

> 管理员接口均需要使用管理员 JWT Token（`/api/AdminLogin` 登录获取）。

#### 认证

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/api/AdminRegister` | 管理员注册 | 无 |
| POST | `/api/AdminLogin` | 管理员登录 | 无 |

#### 书籍管理

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST   | `/admin/book/add` | 添加书籍 | Admin JWT |
| GET    | `/admin/book/getbook/:book_id` | 获取书籍详情 | Admin JWT |
| DELETE | `/admin/book/delete/:book_id` | 删除书籍 | Admin JWT |
| POST   | `/admin/book/update` | 更新书籍信息 | Admin JWT |
| GET    | `/admin/book/getBookTitle/:title` | 搜索书籍 | Admin JWT |

#### 订单管理

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET  | `/admin/order/list` | 获取待发货订单列表 | Admin JWT |
| POST | `/admin/order/orderShip` | 订单发货 | Admin JWT |

#### 秒杀管理

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET  | `/admin/seckill/list` | 获取秒杀活动列表 | Admin JWT |
| POST | `/admin/seckill/create` | 创建秒杀活动 | Admin JWT |
| POST | `/admin/seckill/down/:id` | 下架秒杀活动 | Admin JWT |

---

## 认证机制

项目采用 **JWT 双 Token 方案**：

| Token | 存储方式 | 有效期 | 用途 |
|-------|---------|--------|------|
| Access Token | Cookie `access_token` + 响应体 | 短（如 2h） | 接口请求认证 |
| Refresh Token | Cookie `refresh_token` (HttpOnly) | 长（如 7d） | 刷新 Access Token |

**使用流程**：
1. 调用 `/api/login` 获取 `access_token`
2. 在请求头中携带：`Authorization: Bearer <access_token>`
3. Access Token 过期时，前端自动调用 `/refreshtoken` 通过 Refresh Token 换取新 Token

---

## 错误码说明

| 错误码 | 含义 |
|--------|------|
| 1000 | success - 成功 |
| 1001 | 请求参数错误 |
| 1002 | 用户已存在 |
| 1003 | 用户不存在 |
| 1004 | 用户名或密码错误 |
| 1005 | 服务器繁忙，请稍后再试 |
| 1006 | 无效的 Token |
| 1007 | 需要登录 |
| 1008 | 书籍已存在 |
| 1009 | 书籍不存在 |
| 1010 | 列表存在问题 |
| 1011 | 数据库错误 |
| 1012 | 缓存错误 |
| 1013 | 库存不足 |
| 1014 | 订单不存在或无权访问 |
| 1015 | 订单已确认，请勿重复操作 |
| 1016 | 秒杀活动不存在或未开始 |
| 1017 | 秒杀活动已结束 |
| 1018 | 您已参与该秒杀，请勿重复抢购 |
| 1019 | 秒杀商品已抢完 |

---

## 配置说明

项目通过 Viper 加载 YAML 配置文件，支持以下配置项：

```yaml
app:
  port: 8080           # 服务监听端口
  mode: "debug"        # Gin 运行模式：debug / release

log:
  level: "debug"       # 日志级别
  filename: "shop.log" # 日志文件路径
  max_size: 200        # 单个日志文件最大大小（MB）
  max_age: 30          # 日志保留天数
  max_backups: 7       # 最多保留备份数量

mysql:
  host: "127.0.0.1"
  port: 3306
  user: "root"
  password: "password"
  dbname: "shop"
  max_open_conns: 200
  max_idle_conns: 50

redis:
  host: "127.0.0.1"
  port: 6379
  password: ""
  db: 0
  pool_size: 100

rabbitmq:
  url: "amqp://guest:guest@127.0.0.1:5672/"
```

---

## 架构设计要点

### 秒杀高并发方案

```
用户请求 → Redis 预减库存（原子 DECR）
         → 写入 RabbitMQ 秒杀队列
         → Consumer 异步创建 SeckillOrder + 普通 Order
         → 前端轮询/通知结果
```

### 订单超时关闭

订单创建后，通过 RabbitMQ 延迟队列（TTL + 死信交换机）实现超时自动取消，超时订单归还库存。

### 评分加权算法

使用**贝叶斯加权平均**避免评分数量少的书籍因偶然性排名虚高：

```
加权评分 = (实际均分 × 评分人数 + 全局均分 × 权重因子) / (评分人数 + 权重因子)
```

---

> 项目持续迭代中，如有问题欢迎提 Issue 或 PR。
