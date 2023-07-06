[![Go](https://github.com/oss-go/goodfs/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/oss-go/goodfs/actions/workflows/go.yml)

# GoodFS

GoodFS 是一种高度专注于读写文件对象的分布式系统，具有优秀的高可用性、最终一致性以及强大的水平扩展能力。特别适用于读多写少的场景。在对象存储方面，它采用了多副本策略和纠删码文件修复技术，支持Bucket分类管理和灵活细粒度的对象配置。

GoodFS 的数据迁移功能包括元数据和对象数据两部分。元数据迁移涉及文件系统中的文件元信息，包括文件名、大小、创建时间等。对象数据迁移则包括文件本身的数据内容。这两部分的数据迁移过程是相互独立的。

## 架构

GoodFS 架构包括接口服务、元数据服务、对象数据服务以及作为协调中心的ETCD。接口服务负责处理用户请求，元数据服务负责维护文件元信息，对象数据服务负责维护文件数据，ETCD作为协调中心负责维护集群状态和信息。

- [API Server](./src/apiserver)
- [Meta Server](./src/metaserver)
- [Object Server](./src/objectserver)

### 控制台

- [Admin Server](./src/adminserver) 能够通过网页的方式，方便的进行集群和数据管理，观察服务器运行状况和性能监控。

## 部署

具体参考每个服务目录下的readme文档。

系统默认读取根目录下conf文件夹内的对应文件作为配置文件。

## 编译构建

- Linux 环境

输入指令 `make build-all` 后生成可执行文件至 `bin` 文件夹

- Windows 环境

输入指令 `.\sbin\build.cmd` 后生成可执行文件至 `bin` 文件夹

## 局限性

1. 不支持 Amazon S3 协议
2. 不支持 IAM 身份认证
3. 测试规模有限，可能存在一致性风险

