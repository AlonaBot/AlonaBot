//go:build !linux && !darwin && !windows

package process

func hostMemoryBytesPlatform() uint64 { return 0 }

func sampleTreesPlatform(processes []*Process) (TreeUsage, map[*Process]TreeUsage, error) {
	result := make(map[*Process]TreeUsage, len(processes))
	for _, proc := range processes {
		if proc != nil {
			result[proc] = TreeUsage{}
		}
	}
	return TreeUsage{}, result, nil
}
