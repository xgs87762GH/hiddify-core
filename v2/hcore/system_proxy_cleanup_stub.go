//go:build !windows

package hcore

// clearStaleSystemProxy is a no-op outside Windows.
func clearStaleSystemProxy() {}

// rememberSystemProxy is a no-op outside Windows.
func rememberSystemProxy() {}
