//go:build !linux && !darwin && !windows

package process

import (
	"context"
	"fmt"
)

func hostMemoryBytesPlatform() (uint64, error) {
	return 0, fmt.Errorf("host memory sampling is unsupported on this platform")
}

func sampleTreesPlatform(ctx context.Context, processes []*Process) (TreeUsage, map[*Process]TreeUsage, error) {
	result := make(map[*Process]TreeUsage, len(processes))
	if err := ctx.Err(); err != nil {
		return TreeUsage{}, result, err
	}
	for _, proc := range processes {
		if proc != nil {
			result[proc] = TreeUsage{}
		}
	}
	return TreeUsage{}, result, nil
}
