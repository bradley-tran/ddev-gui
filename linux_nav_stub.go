//go:build !linux
// +build !linux

package main

import "context"

// InstallMouseNavHandler is a no-op on non-Linux platforms.
// On Linux, the real implementation intercepts GDK mouse button events.
func InstallMouseNavHandler(ctx context.Context) {}
