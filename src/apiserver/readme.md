# 接口服务 API Server

## 保存策略

文件对象的保存策略由Bucket指定，当Bucket未指定时，将使用配置文件中的配置来决定。

## 身份校验

系统提供两种安全检查模式，通过一种则视为合法

1. Basic Auth：配置固定的账号密码

2. 第三方地址回调

   发送POST请求到指定url地址，并携带以下参数，通过JSON序列化
   - bucket: 请求目标的Bucket
   - name: 请求目标的文件名
   - version: 请求目标的版本号
   - method：待验证请求的方法名
   - extra: 配置中指定的额外参数名，将尝试从请求的url和header中获取
    

## 配置文件参考

```yaml
port: 8080 #部署端口
select-strategy: space-first #对象服务的负载均衡策略 random space-first io-first
log:
  level: debug  #日志等级
  caller: false #是否包含日志发起所在的文件的位置信息
  store-dir: path_to_store/log #日志文件保存目录 为空则不保存
  email: #配置日志警告发送邮箱
    sender: example@qq.com
    smtp-host: smtp.qq.com
    smtp-port: 587
    target: example@gmail.com
    password: password
performance:
  enable: false
  store: local
registry: #服务注册信息
  server-ip: "" #服务器IP 为空则自动检测
  server-id: api-0 #唯一id 可自动生成默认值
  group: "goodfs" #组 默认值
  name: "apiserver" #类 默认值
etcd: #ETCD的配置
  endpoint:
    - example.com:2379
  username: username
  password: password
discovery: #用于发现其他两种服务，指定其‘类’名，默认值
  meta-server-name: "metaserver"
  data-server-name: "objectserver"
object: 
  checksum: false #对象上传后检查其校验值是否一致 （增加上传时间）
  distinct-size: 100mb #对象去重标准，大于此大小则触发全局查重 （增加上传时间）
  distinct-timeout: 150ms #对象去重超时时间，超时则视为不重复
  reed-solomon: #ReedSolomon参数配置
    data-shards: 4
    parity-shards: 2
    block-per-shard: 10000
  replication: #多副本参数配置
    copies-count: 4 #副本数量
    loss-tolerance-rate: 0.1 #可容忍丢失的百分比 越高触发修复的概率越低
    copy-async: true #异步复制副本 false则可能增加上传时间
auth:
  enable: false # 是否开启身份检查 以下任意两种模式有一种通过则视为合法
  password: # basic-auth 检查模式
    enable: true
    username: admin
    password: admin
  callback: # 第三方检查模式
    enable: false
    url: http://localhost:8090/v1/authorize
    params: [ 'access-token' ] #备注的参数将一起回调到指定链接
tls: # tls配置
  enabled: false
  server-cert-file: path_to_cert\example.com+5.pem
  server-key-file: path_to_key\example.com+5-key.pem
```