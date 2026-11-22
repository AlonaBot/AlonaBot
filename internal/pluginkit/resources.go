package pluginkit

import (
	"encoding/json"
	"fmt"
	"os"
)

// Resources 保存与一次成功构建同源的包内配置与文案。
type Resources struct {
	Manifest, ConfigSchema, AdapterConfigSchema []byte
	I18n                                        map[string][]byte
}

// ReadResources 读取构建摘要绑定的旁车，供离线装配与打包使用。
func ReadResources(build Build) (Resources, error) {
	var resources Resources
	raw, err := os.ReadFile(fmt.Sprintf("%s.%x.json", build.Executable, build.Digest))
	if err != nil {
		return resources, err
	}
	err = json.Unmarshal(raw, &resources)
	return resources, err
}
