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

// ServiceToolList 指定调用插件可调用的工具目录服务。
const ServiceToolList = "core.tool.list"

// ServiceEdit 指定 Core 的消息编辑服务。
const ServiceEdit = "core.edit"

// ServiceAdapterEdit 向账号实例提交消息编辑。
const ServiceAdapterEdit = "adapter.edit"

// ServiceDelete 指定 Core 的消息撤回服务。
const ServiceDelete = "core.delete"

// ServiceAdapterDelete 向账号实例提交消息撤回。
const ServiceAdapterDelete = "adapter.delete"

// ServiceAction 指定 Core 的平台动作服务。
const ServiceAction = "core.action"

// ServiceAdapterAction 向账号实例提交已声明的平台动作。
const ServiceAdapterAction = "adapter.action"

// ServiceMediaURL 为本次发送的只读媒体令牌签发临时 URL。
const ServiceMediaURL = "core.media.url"

// ServiceAdapterCommandsSync 替换账号实例的命令菜单。
const ServiceAdapterCommandsSync = "adapter.commands.sync"

// ServiceContinuationStart 登记父 Handler 成功结束后的延续调用。
const ServiceContinuationStart = "core.continuation.start"

// ServiceWait 等待延续调用所属用户在原会话的下一条消息。
const ServiceWait = "core.wait"

// ServiceToolInvoke 指定沿父调用链执行工具的服务。
const ServiceToolInvoke = "core.tool.invoke"

// ServiceToolConfirm 指定操作的用户确认服务。
const ServiceToolConfirm = "core.tool.confirm"
