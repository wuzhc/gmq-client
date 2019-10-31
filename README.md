> `gmq`的go版本的客户端,目前`gmq-client`只是一个demo级,建议自己封装一个客户端

## 使用
```bash
git clone https://github.com/wuzhc/gmq-client-go.git
cd gmq-client-go
make
```

## 命令
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
默认会编译后的文件会存放在`$GOPATH/bin`,如果你的`$GOPATH/bin`没有添加环境变量,可以直接用`./build/gclient`代替`gclient`,或者直接运行源码,用`go run main.go`代替`glicent`