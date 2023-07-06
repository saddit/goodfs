# 对象服务 Object Server

## 配置文件参考

```yaml
port: 8100 #部署端口
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
exclude-mount-points: ["C:"] #排除挂载点
base-mount-point: "E:" #包含的挂载点，排除与包含同时配置则只有包含生效
storage-path: /file/temp #保存路径，将在每个挂载点下的该路径保存数据
cache: #缓存配置
  max-size: 1GB #缓存空间
  ttl: 3h #生命周期
  clean-interval: 12m #检测周期
  max-item-size: 16MB #最大缓存对象的大小
registry: #服务注册信息
  server-ip: "" #服务器IP 为空则自动检测
  server-id: api-0 #唯一id 可自动生成默认值
  group: "goodfs" #组 默认值
  name: "objectserver" #类 默认值
etcd: #ETCD的配置
  endpoint:
    - example.com:2379
  username: username
  password: password
discovery:
  meta-server-name: "metaserver"
```