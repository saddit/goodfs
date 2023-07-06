# 元数据服务 Meta Server

## Raft集群配置

## 哈希槽配置

## 配置文件参考

```yaml
port: "8090" #部署端口
data-dir: path_to_store/temp  #数据存储的目录
max-concurrent-streams: 100 #grpc最大并发数
log:
  level: debug  #日志等级
  store-dir: path_to_store/log #日志文件保存目录 为空则不保存
  caller: false #是否包含日志发起所在的文件的位置信息
  email: #配置日志警告发送邮箱
    sender: example@qq.com
    smtp-host: smtp.qq.com
    smtp-port: 587
    target: example@gmail.com
    password: password
cluster: 
  enable: false # 是否启用Raft同步集群
  bootstrap: false # 是否作为启动节点 （仅第一次有效）
  group-id: raft-1 # Raft集群的编号 需要集群内所有单位保持一致
  election-timeout: 900ms # 选举超时时间
  heartbeat-timeout: 800ms # 心跳超时时间
  nodes: # 集群内的所有节点
    - meta-0,localhost:8090
    - meta-1,localhost:8091
    - meta-2,localhost:8092
registry: #服务注册信息
  server-ip: "" #服务器IP 为空则自动检测
  server-id: api-0 #唯一id 可自动生成默认值
  group: "goodfs" #组 默认值
  name: "metaserver" #类 默认值
etcd: #ETCD的配置
  endpoint:
    - example.com:2379
  username: username
  password: password
hash-slot: #哈希槽配置
  slots:
    - 0-16384
  prepare-timeout: 1m0s
cache: # 缓存配置
  ttl: 20m0s  #生命周期
  clean-interval: 10m0s #检测周期
  max-size: 1GB #最大缓存空间
```