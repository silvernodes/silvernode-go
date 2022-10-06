Silvernode-Go 是一个轻量级、可扩展、简洁易用的即时通讯框架

# 初心
服务的本质在于交互，每一个服务节点都好比一个活生生的人，每个人会记录自己感兴趣的人，并与之产生联系，人与人之间的交流构建了社会关系，众多节点之间的有序交互则组建了稳定运转的服务集群。

常在思考，如何才能降低即时通讯服务器编写的难度？

传统的企业级应用，前后端交互多基于Http协议，逻辑形式多为请求应答模式，辅之于成熟的技术框架，开发难度会成倍下降。

而对于有后端实时推送需求的业务(比如聊天或者游戏)，我们则需要基于更底层的即时通讯协议去实现。此时，底层网络事件的处理、链接的管理、收发操作带来的逻辑割裂、内部节点异步通讯等等方面的处理，无形中都会增加编码的难度。

如果细心对比我们便不难发现，其实相较于企业级应用开发，即时通讯服务仅是在前者请求应答模式的基础之上，新增了后端主动推送的能力，其他并无区别。

那么我们能否通过对底层业务的简化，达到大幅降低即时通讯服务器编码难度的目的？当然！

—— 此为Silvernode-Go构建之初心

# 特性
	
## 注册中心及服务发现

- Silvernode-Go支持将etcd、consul、nacos作为注册中心，并提供服务发现、共享配置等机制

## 多种即时网络通讯协议支持
- TCP

	Silvernode-Go在底层解决了TCP的分包问题，并加入心跳检测机制，可以敏感捕获底层的网络异常
	
- UDP/KCP

	Silvernode-Go基于KCP协议实现可靠的UDP传输，并通过心跳检测机制弥补单纯依靠KCP无法察觉断线及其他网络异常的问题
	
- WebSocket
	
	Silvernode-Go集成了官方go-net扩展包的ws协议支持，兼容web、小程序、h5等各种应用场景
	

## 节点化及集群扩展
### 节点化设计
- Silvernode-Go将每个独立进程抽象为集群中的一个节点(node)，众多的节点构建一个完整的微服务集群
- 每个节点具备自己的实例id、名称(也是种类，诸如gate、login等等)
- 每个节点启动时，会先将自身信息写入到注册中心；当注册中心检测到节点不可用时，信息会被删除

### 集群扩展
- Silvernode-Go允许每个节点设置自身的backend(后端服务)
- 当借由服务发现检测到对应种类的节点时，会自发向目标节点发起连接请求，并自动维护其可用状态，以实现整个集群的动态扩展及按需连接
	
### 身份验证
- Silvernode-Go中每个节点都具备自己独一无二的sig(身份识别码)
- 当目标节点监听到链接请求时，会根据发起节点的id查阅其身份信息，若验证异常则视为外来节点
	
### 简化操作
- Silvernode-Go隐藏了底层的链接管理，整个集群内的所有节点实现按需连接
- Silvernode-Go封装了套接字等基础操作，只要知道对应的节点id，并且已建立链接(按需或是外部主动发起)，就可以直接向对方发送数据
	
## 协程的使用及管控
- Silvernode-Go结合goroutine和channel二者特性，封装并约束其使用
- Processor (执行器)
	
	使用方法等同于线程安全的任务队列，可以便捷且安全的实现并发任务的调度
	
	自身可以只包含1个协程，构建安全的同步执行上下文环境；也可包含多个协程，高效并发的执行某项调度任务
	
- Scheduler (调度器)

	非常方便的实现诸如延迟执行、定时调度、重复调度等类型的调度任务
	
- Service (服务)

	后台不间断的执行某项操作，直至强制中断
	
## 位于应用层的Peer

- Silvernode-Go选择将Peer作为应用层建制，这一点与其他框架将其视为网络层产物的设计思想不同
- Silvernode-Go通过Peer将相对复杂的网络层数据收发，抽象为极简的应用层RPC调用
- Silvernode-Go中的每个Peer分别对应一个独立的逻辑单元，定位类似Web框架中的Controller
- Peer默认使用了Go语言的反射机制(reflect)，但可以通过自身的发布操作实现代码自动化生成，从而规避反射带来的效率损失
- Peer可自由指定一个专属Processor，从而使得应用层的任意逻辑单元，均可便捷的实现同步/并发/单协程/多协程等业务处理模型
- Peer与底层框架之间保持松耦合，可以自由选择使用

## 关于HTTP
- Silvernode-Go在原生Http网络库的基础上作了简化，用以赋予自身快速构建基础Http服务的能力
- Http服务是一套独立体系，一般只针对集群外部的网络请求，不具备前边提到的节点化、服务发现、按需链接等诸多特性
- 作用于应用层，与底层框架之间保持松耦合，支持选择性使用，这一点与Peer类似
- 如果你想搭建稳定可靠的企业级应用服务，go-zero会是更好的选择，这里强烈推荐

## 自定义插件扩展
- Silvernode-Go提供了一套标准的插件整合机制，可方便快捷的整合其他三方库及组件

## 日志系统
- Silvernode-Go实现了一个可定制的日志系统，使用者可重写Writter自由选择日志的输出端，比如控制台、本地文件或者更加体系化的日志统计系统(Prometheus、ELK等，需自行实现)

## 较为丰富的前端SDK
- Silvernode-Go目前提供了js、ts、c#等前端SDK支持，涵盖了web、小程序、h5、游戏开发等领域，后续会继续提供其他语言版本的SDK

## 限流、熔断、服务降级

- 构建中，敬请期待。。


# 间架结构

![alt framework](https://images.cnblogs.com/cnblogs_com/itfantasy/2226315/o_221006170127_framework.png)

# 处理流程(Pipeline)

![alt framework](https://images.cnblogs.com/cnblogs_com/itfantasy/2226315/o_221006170113_pipeline.png)

# Hello Silvernode-Go!
- app.yml (工程配置文件)
``` 
# 注册中心配置
cluster: 
  type: etcd
  nacos:
    namespace: silvernode
    ipaddress: 127.0.0.1
    port: 8848
  consul:
    namespace: silvernode
    ipaddress: 127.0.0.1
    port: 8500
  etcd:
    namespace: silvernode
    ipaddress: 127.0.0.1
    port: 2379
# 节点配置
node:
  nodeid: room#1 # 节点id
  name: room # 节点名称
  backends: # 后端服务，可以是多种
    - lobby # 代表一旦发现name为lobby的节点会自动发起链接
  ispub: true # 允许外部访问
  endpoints: # 可访问终端，可以多个，支持多种协议
    - ws://127.0.0.1:33056/room # 主endpoint
    - udp://127.0.0.1:33066
```
- main.go
```
package main

import (
	"fmt"

	silvernode "github.com/silvernodes/silvernode-go"
	"github.com/silvernodes/silvernode-go/peers"
)

func main() {
	// 流水线设定
	silvernode.BindPipeline(&silvernode.Pipeline{
		Init: func() {
			// TODO: 调用peers.Register注册外部可访问的Peer
            peers.Register(&proc.TestProc{}, nil)
		},
	})
	peers.Boot() // 启用Peer系统
	// 开启服务
	if err := silvernode.Serve(); err != nil {
		fmt.Println("服务节点启动失败:" + err.Error())
	}
}
```

- proc/test_proc.go
```
package proc

import (
	"github.com/silvernodes/silvernode-go/peers"
)

type Arg struct {
}

type Reply struct {
	Msg string
}

type TestProc struct {
	Peer *peers.Peer
}

func (t *TestProc) Hello(arg *Arg, reply *Reply) error {
	reply.Msg = "Hello Silvernode !!"
	return nil
}

```

- client
```
peer.Call("ServerNode Id", "TestProc.Hello", args, (ret, err: error) => {
    if (err != null) {
        console.log(err);
        return;
    }
    console.log(ret.Msg);
});
```

# 前端SDK列表
- [silvernode-sdks-cs](https://github.com/silvernodes/silvernode-sdks-cs)
- [silvernode-sdks-ts(ws only)](https://github.com/silvernodes/silvernode-sdks-ts)
- [silvernode-sdks-js(ws only)](https://github.com/silvernodes/silvernode-sdks-js)
- more ...

# 历史版本
- [0.1.0-勇者](https://github.com/silvernodes/silvernode-go) [ 致那黑夜中的呜咽与怒吼！ ——[《孤勇者》（陈奕迅）](https://music.163.com/#/song?id=1901371647)]

# 技术交流

![alt framework](https://images.cnblogs.com/cnblogs_com/itfantasy/2226315/o_221006172249_wx.png)

# 备注
- Silvernode-Go 遵循 [MIT](https://github.com/silvernodes/silvernode-go/blob/main/LICENSE) 开源许可协议