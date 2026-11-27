//go:build !linux && !darwin && !windows

package process

import "context"

func hostMemoryBytesPlatform() uint64 { return 0 }

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
