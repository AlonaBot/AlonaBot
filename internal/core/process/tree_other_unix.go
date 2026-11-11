//go:build unix && !darwin

package process

func (t *tree) signalPermissionError(err error) error { return err }
