# GoodFS

## 架构

### 协调中心 ETCD

实现注册中心、配置中心、发布订阅的功能

### 接口服务 Api Server

负责处理API请求

### 元数据服务 Meta Server     

复制存储元数据

### 对象数据服务 Obejct Server

负责存储对象数据

## 配置文件

### Api Server

```yaml

```

### Meta Server

```yaml

```

### Object Server

```yaml

```

## 部署

### 单体部署

应用支持单体部署，主要修改 Meta Server 的配置信息

### 伪分布式部署

应用支持伪分布式部署，仅需分配不同的端口

### 分布式部署

应用支持分布式部署

## 后台管理 Admin Server

Admin Server 是独立的控制台应用，内嵌Vue前端，可直接部署Golang可执行文件，也可配置后独立部署前端

### 直接部署

确认配置文件

```yaml

```

### 前后端分离部署

修改后端配置文件

```yaml

```

修改前端配置文件，更改baseUrl为后端地址

```js

```

## 编译构建

 - Linux 环境
 直接执行 `make build-all` 后生成可执行文件至 `bin` 文件夹
 - Windows 环境
 输入指令 `.\sbin\build.cmd` 后生成可执行文件至 `bin` 文件夹

## 局限性
