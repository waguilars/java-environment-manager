//go:build windows

package platform

// NewPlatform creates the appropriate Platform implementation based on the OS
func NewPlatform() Platform {
	return NewWindowsPlatform()
}
