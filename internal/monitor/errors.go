package monitor

import "errors"

var (
	// ErrCpuInfoNotReady indicates CPU information is not yet available from background sampling.
	ErrCpuInfoNotReady = errors.New("cpu info not ready")
	// ErrProcessInfoNotReady indicates process information is not yet available from background sampling.
	ErrProcessInfoNotReady = errors.New("process info not ready")
)
