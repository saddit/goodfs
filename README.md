[![Go](https://github.com/oss-go/goodfs/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/oss-go/goodfs/actions/workflows/go.yml)

[Looking for English?](./docs/README_en.md)

# GoodFS

GoodFS 是一种高度专注于读写文件对象的分布式系统，它具有优秀的高可用性、最终一致性以及强大的水平扩展能力。特别适用于读多写少的场景。在对象存储方面，它采用了多副本策略和纠删码文件修复技术，支持Bucket分类管理和灵活细粒度的对象配置，使得用户能够轻松地管理和维护存储对象。

GoodFS 的数据迁移功能包括元数据和对象数据两部分。元数据迁移涉及文件系统中的文件元信息，包括文件名、大小、创建时间等。对象数据迁移则包括文件本身的数据内容。这两部分的数据迁移过程是相互独立的，并且GoodFS 具有严格的数据一致性保证，保证在迁移过程中数据不丢失或不重复。此外，GoodFS还支持定时迁移和手动迁移两种方式，使得用户能够根据自己的需求进行数据迁移。

## 架构

GoodFS 架构包括接口服务、元数据服务、对象数据服务以及作为协调中心的ETCD。接口服务负责处理用户请求，元数据服务负责维护文件元信息，对象数据服务负责维护文件数据，ETCD作为协调中心负责维护集群状态和信息。

- [API Server](./src/apiserver)
- [Meta Server](./src/metaserver)
- [Object Server](./src/objectserver)

### 控制台

- [Admin Server](./src/adminserver)

   能够通过网页的方式，方便的进行集群和数据管理，观察服务器运行状况和性能监控

## 部署

GoodFS 提供了多种部署方式，包括单体部署、伪分布式部署和完全分布式部署。

单体部署是最简单的部署方式，它将所有服务部署在一台机器上。这种方式适用于小规模或测试环境。

伪分布式部署将元数据服务和对象数据服务部署在不同的机器上，但是接口服务仍然部署在单台机器上。这种方式适用于中小规模环境。

完全分布式部署将所有服务都部署在不同的机器上，这种方式适用于大规模环境。

### 单体部署

应用支持单体部署，主要修改 Meta Server 的配置信息

### 伪分布式部署

应用支持伪分布式部署，仅需分配不同的端口

### 分布式部署

应用支持分布式部署

## 编译构建

- Linux 环境

输入指令 `make build-all` 后生成可执行文件至 `bin` 文件夹

- Windows 环境

输入指令 `.\sbin\build.cmd` 后生成可执行文件至 `bin` 文件夹

## 局限性

1. 不支持 Amazon S3 协议
2. 不支持 IAM 身份认证

