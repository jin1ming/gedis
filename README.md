# Gedis
使用go语言重新实现redis

进展：
- [ ]  基本数据结构构建
    - [ ]  动态字符串SDS
    - [ ]  双端链表list
    - [ ]  hash
    - [ ]  set
    - [ ]  zset
- [ ]  Redis协议下的客户端服务端通信
    - [ ]  基于TCP实现通信
    - [ ]  尝试加入 Reactor 网络库 gev
- [ ]  客户端命令解析
    - [ ]  get
    - [ ]  set
    - [ ]  select
    - [ ]  ping
    - [ ]  quit
- [ ]  发布订阅模式
- [ ]  YAML配置设计
- [ ]  持久化
    - [ ]  RDB
    - [ ]  AOF
- [ ]  主从模式
- [ ]  集群