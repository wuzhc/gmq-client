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
# 单节点操作
# 推送消息
gclient -node_addr="127.0.0.1:9503" -cmd="push" -topic="gmq-topic-1" -push_num=1000
# 批量推送消息
gclient -node_addr="127.0.0.1:9503" -cmd="mpush" -topic="gmq-topic-1" -push_num=1000
# 消费消息
gclient -node_addr="127.0.0.1:9503" -cmd="pop" -topic="gmq-topic-1" -pop_num=1000
# 轮询消费消息
gclient -node_addr="127.0.0.1:9503" -cmd="pop_loop" -topic="gmq-topic-1"
# 确认已消费
gclient -node_addr="127.0.0.1:9503" -cmd="ack" -topic="gmq-topic-1" -msg_id="374389276810416130"
# 消费死信消息
gclient -node_addr="127.0.0.1:9503" -cmd="dead" -topic="gmq-topic-1" -pop_num=1000

# 多节点操作
# 按权重模式选择节点推送消息
gclient -register_addr="http://127.0.0.1:9595" -cmd="push_by_weight" -topic="gmq-topic-1" -push_num=1000
# 按平均模式选择节点推送消息
gclient -register_addr="http://127.0.0.1:9595" -cmd="push_by_avg" -topic="gmq-topic-1" -push_num=1000
# 按随机模式选择节点推送消息
gclient -register_addr="http://127.0.0.1:9595" -cmd="push_by_rand" -topic="gmq-topic-1" -push_num=1000
# 按权重模式选择节点消费消息
gclient -register_addr="http://127.0.0.1:9595" -cmd="pop_by_weight" -topic="gmq-topic-1"
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
默认会把编译后的文件存放在`$GOPATH/bin`,如果你的`$GOPATH/bin`没有添加到环境变量,可以直接用`./build/gclient`代替`gclient`,或者直接运行源码,用`go run main.go`代替`glicent`

## 相关链接
- [https://github.com/wuzhc/gmq](https://github.com/wuzhc/gmq)