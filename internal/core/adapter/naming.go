package adapter

// Named 提供平台账号用户名，供命令匹配 /cmd@name；Instance 可选实现。
type Named interface {
	AccountName() string
}
