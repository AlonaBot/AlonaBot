package manage

import (
	"io/fs"
	"path"
	"strconv"
	"strings"
)

// availableBuildMemory 优先取本进程 cgroup v2 到根的最小剩余额度，再回退到全局可用内存。
func availableBuildMemory(root fs.FS) (int64, bool) {
	const cgroups = "sys/fs/cgroup"
	if membership, err := fs.ReadFile(root, "proc/self/cgroup"); err == nil {
		for line := range strings.Lines(string(membership)) {
			group, ok := strings.CutPrefix(strings.TrimSpace(line), "0::/")
			if !ok {
				continue
			}
			var available int64
			found := false
			for dir := path.Join(cgroups, path.Clean("/"+group)); ; dir = path.Dir(dir) {
				limitRaw, _ := fs.ReadFile(root, path.Join(dir, "memory.max"))
				if limit, err := strconv.ParseInt(strings.TrimSpace(string(limitRaw)), 10, 64); err == nil {
					usedRaw, _ := fs.ReadFile(root, path.Join(dir, "memory.current"))
					if used, err := strconv.ParseInt(strings.TrimSpace(string(usedRaw)), 10, 64); err == nil {
						remaining := max(int64(0), limit-used)
						if !found || remaining < available {
							available, found = remaining, true
						}
					}
				}
				if dir == cgroups {
					break
				}
			}
			if found {
				return available, true
			}
			break
		}
	}
	if info, err := fs.ReadFile(root, "proc/meminfo"); err == nil {
		for line := range strings.Lines(string(info)) {
			if value, ok := strings.CutPrefix(line, "MemAvailable:"); ok {
				kb, _ := strings.CutSuffix(strings.TrimSpace(value), " kB")
				n, err := strconv.ParseInt(strings.TrimSpace(kb), 10, 64)
				return n << 10, err == nil
			}
		}
	}
	return 0, false
}
