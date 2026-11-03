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

// ServiceMediaBeginUpload 指定媒体上传的声明与预留服务。
const ServiceMediaBeginUpload = "core.media.begin_upload"

// ServiceMediaWriteChunk 指定媒体顺序分块服务。
const ServiceMediaWriteChunk = "core.media.write_chunk"

// ServiceMediaCommitUpload 指定媒体校验与引用签发服务。
const ServiceMediaCommitUpload = "core.media.commit_upload"

// 会话服务按调用插件身份隔离数据，允许无父调用链的请求。
const (
	// ServiceConversationAppend 指定会话条目追加服务。
	ServiceConversationAppend = "core.conversation.append"
	// ServiceConversationList 指定会话条目分页读取服务。
	ServiceConversationList = "core.conversation.list"
	// ServiceConversationClear 指定会话清除服务。
	ServiceConversationClear = "core.conversation.clear"
)

// ServiceBusPublish 指定 Event Bus 的发布准入服务。
const ServiceBusPublish = "core.bus.publish"

// ServiceAdapterMediaRead 指定 Adapter 平台文件的分块读取服务。
const ServiceAdapterMediaRead = "adapter.media.read"

// ServiceMediaRead 指定 Core 媒体令牌的分块读取服务。
const ServiceMediaRead = "core.media.read"

// ServiceMediaGrant 指定 Core 媒体读取权限的受限授权服务。
const ServiceMediaGrant = "core.media.grant"
