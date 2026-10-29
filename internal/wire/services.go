package wire

// ServiceSend 指定 Core 的消息发送服务；固定服务名不含 /，与插件路由目标分离。
const ServiceSend = "core.send"

const (
	// ServiceAdapterInstanceStart 启动一个账号实例。
	ServiceAdapterInstanceStart = "adapter.instance.start"
	// ServiceAdapterInstanceStop 停止一个账号实例。
	ServiceAdapterInstanceStop = "adapter.instance.stop"
	// ServiceAdapterSend 向账号实例提交消息发送。
	ServiceAdapterSend = "adapter.send"
	// ServiceCoreAdapterDeliver 向 Core 交付实例事件与游标。
	ServiceCoreAdapterDeliver = "core.adapter.deliver"
	// ServiceCoreAdapterCursor 保存实例的接收游标。
	ServiceCoreAdapterCursor = "core.adapter.cursor"
	// ServiceCoreAdapterStatus 上报实例致命错误。
	ServiceCoreAdapterStatus = "core.adapter.status"
)

// 会话服务按调用插件身份隔离数据，允许无父调用链的请求。
const (
	// ServiceConversationAppend 指定会话条目追加服务。
	ServiceConversationAppend = "core.conversation.append"
	// ServiceConversationList 指定会话条目分页读取服务。
	ServiceConversationList = "core.conversation.list"
	// ServiceConversationClear 指定会话清除服务。
	ServiceConversationClear = "core.conversation.clear"
)
