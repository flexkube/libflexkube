package flexkube

import (
	"crypto/sha256"
	"fmt"
)

func sha256sum(data []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(data))
}
