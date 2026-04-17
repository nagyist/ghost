package common

import (
	"fmt"
)

// FormatStorageSize formats storage size in MiB or GiB as appropriate
func FormatStorageSize(storageMib *int) string {
	if storageMib == nil {
		return "-"
	}
	mib := *storageMib
	if mib < 1024 {
		return fmt.Sprintf("%dMiB", mib)
	}
	gib := float64(mib) / 1024
	// Show one decimal place if not a whole number
	if gib == float64(int(gib)) {
		return fmt.Sprintf("%dGiB", int(gib))
	}
	return fmt.Sprintf("%.1fGiB", gib)
}
