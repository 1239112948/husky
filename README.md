# husky

## 说明
包括服务（进程）之间通信，消息事件注册，配置热更新，日志等。游戏场景剥离出来的

以当前生产业务下的场景拓扑结构为例

![avatar](https://github.com/guogeer/husky/blob/master/doc/service.png)

目录说明

```
cmd              网络消息处理
router           路由服，服务注册，数据转发等全局功能
gateway          网关服，负责客户端消息转发、负载均衡
config.xml  相关配置，如数据库账号密码，路由服地址等
...                  配置热更新，待整理
```
