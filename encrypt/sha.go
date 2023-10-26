package encrypt

import (
	"crypto/sha256"
	"encoding/base64"
)

func SHA256(data []byte) string {
	hd := sha256.Sum256(data)
	return base64.StdEncoding.EncodeToString(hd[:])
}
