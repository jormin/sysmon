//go:build !darwin
// +build !darwin

package collect

import "errors"

var errCPUFallbackUnsupported = errors.New("cpu fallback 仅支持 darwin")

func fetchCPUFallback() (map[string]cpuRaw, error) {
	return nil, errCPUFallbackUnsupported
}
