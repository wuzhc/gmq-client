> `gmq`的go版本的客户端,目前`gmq-client`只是一个demo级,建议自己封装一个客户端

## 使用
```bash
git clone https://github.com/wuzhc/gmq-client-go.git $GOPATH/github.com/wuzhc/gmq-client-go
cd gmq-client-go
make
```

## 命令
打开shell,执行如下命令
```bash
# 声明队列
go run main.go -node_addr="127.0.0.1:9503" -cmd="delcare" -topic="ketang" -bind_key="homework"

# 单节点操作
# 推送消息
go run main.go -node_addr="127.0.0.1:9503" -cmd="push" -topic="ketang" -route_key="homework" -push_num=1000 
# 批量推送消息
go run main.go -node_addr="127.0.0.1:9503" -cmd="mpush" -topic="ketang" -route_key="homework" -push_num=1000
# 消费消息
go run main.go -node_addr="127.0.0.1:9503" -cmd="pop" -topic="ketang" -bind_key="homework" -pop_num=1000 
# 轮询消费消息
go run main.go -node_addr="127.0.0.1:9503" -cmd="pop_loop" -bind_key="homework" -topic="ketang" 
# 确认已消费
go run main.go -node_addr="127.0.0.1:9503" -cmd="ack" -topic="ketang" -bind_key="homework" -msg_id="374389276810416130" 
# 消费死信消息
go run main.go -node_addr="127.0.0.1:9503" -cmd="dead" -topic="ketang" -bind_key="homework" -pop_num=1000 

# 多节点操作
# 按权重模式选择节点推送消息
go run main.go -etcd_endpoints="http://127.0.0.1:2379" -cmd="push_by_weight" -topic="ketang" -route_key="homework" -push_num=1000
# 按平均模式选择节点推送消息
go run main.go -etcd_endpoints="http://127.0.0.1:2379" -cmd="push_by_avg" -topic="ketang" -route_key="homework" -push_num=1000
# 按随机模式选择节点推送消息
go run main.go -etcd_endpoints="http://127.0.0.1:2379" -cmd="push_by_rand" -topic="ketang" -route_key="homework" -push_num=1000
# 按权重模式选择节点消费消息
go run main.go -etcd_endpoints="http://127.0.0.1:2379" -cmd="pop_by_weight" -topic="ketang" -bind_key="homework"
```
选项说明:  
- `register_addr` 注册中心http地址
- `node_addr` 节点tcp地址
- `cmd` 执行命令类型
- `topic` 消息所属topic,即消息的分类
- `pop_num` 此次要消费多少条消息
- `push_num` 此次要推送多少条消息
- `msg_Id` 消息ID

注意:  
- 需要先启动etcd服务
- 执行发布和消费命令之前，需要先声明队列，通过绑定键将queue绑定到topic上

## 相关链接
- [https://github.com/wuzhc/gmq](https://github.com/wuzhc/gmq)