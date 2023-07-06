## 后台管理 Admin Server

Admin Server 是独立的控制台应用，内嵌Vue前端。

### 直接部署

确认配置文件

修改admin-server配置文件，以下为示例

```yaml
port: 9898  # 部署端口号
log:
  level: debug  #日志等级
enabled-api-tls: false  #API Server 是否启用了tls
discovery:
  group: "goodfs" #服务注册的组名
etcd: #ETCD的配置
  endpoint:
    - example.com:2379
  username: username
  password: password
```

修改前端配置文件，更改backendPort为后端监听的端口号

文件位于[此目录](./ui/public/config.js)

```js
window.backendPort = 80
```