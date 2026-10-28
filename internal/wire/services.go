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
