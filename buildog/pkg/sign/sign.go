package sign

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type Payload struct {
	NotaryGun string `json:"notaryGun"`
	SHA256    string `json:"sha256"`
	ByteSize  int    `json:"byteSize"`
	Version   string `json:"version"`
}

// curl -L -X POST -H "Authorization: Bearer sdjfhjsdfhsjdf" https://signing-dev.repositories.cloud.sap/signingsvc/sign $payload

func Sign(payload Payload) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	r := bytes.NewReader(b)
	http.Post("https://signing-dev.repositories.cloud.sap/signingsvc/sign", "application/json", r)
}
