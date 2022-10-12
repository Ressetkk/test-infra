package sign

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	TypeNotaryBackend = "notary"
)

type NotaryConfig struct {
	// Set URL to Notary server signing endpoint
	Endpoint string `yaml:"endpoint" json:"endpoint"`
	// SecretPath contains path to the authentication credentials used for specific notary server
	Secret *AuthSecretConfig `yaml:"secret,omitempty" json:"secret,omitempty"`
	// Time after connection to notary server should time out
	Timeout time.Duration `yaml:"timeout" json:"timeout"`
	// RetryTimeout is time between each signing request to notary in case something fails
	// Default is 10 seconds
	RetryTimeout time.Duration `yaml:"retry-timeout" json:"retry-timeout"`
}

// AuthSecretConfig contains auth information for notary server
type AuthSecretConfig struct {
	Path string `yaml:"path" json:"path"`
	Type string `yaml:"type" json:"type"`
}

// SigningRequest contains information about all images with tags to sign using Notary
type SigningRequest struct {
	// Global unique name, e.g. full image name with registry URL
	NotaryGun string `json:"notaryGun"`
	// SHA sum of manifest.json
	SHA256 string `json:"sha256"`
	// size of manifest.json
	ByteSize int64 `json:"byteSize"`
	// Image tag
	Version string `json:"version"`
}

type AuthFunc func(r *http.Request) *http.Request

type NotarySigner struct {
	authFunc     AuthFunc
	c            http.Client
	url          string
	retryTimeout time.Duration
}

// BearerToken mutates created request, so it contains bearer token of authorized user
// Serves as middleware function before sending request
func BearerToken(token string) AuthFunc {
	return func(r *http.Request) *http.Request {
		r.Header.Add("Authorization", "Bearer "+token)
		return r
	}
}

func (ns NotarySigner) buildSigningRequest(images []string) ([]SigningRequest, error) {
	var sr []SigningRequest
	for _, i := range images {
		var base, tag string
		// Split on ":"
		parts := strings.Split(i, tagDelim)
		// Verify that we aren't confusing a tag for a hostname w/ port for the purposes of weak validation.
		if len(parts) > 1 && !strings.Contains(parts[len(parts)-1], regRepoDelimiter) {
			base = strings.Join(parts[:len(parts)-1], tagDelim)
			tag = parts[len(parts)-1]
		}
		ref, err := name.ParseReference(i)
		if err != nil {
			return nil, fmt.Errorf("ref parse: %w", err)
		}
		i, err := remote.Image(ref)
		if err != nil {
			return nil, fmt.Errorf("get image: %w", err)
		}
		m, err := i.Manifest()
		if err != nil {
			return nil, fmt.Errorf("image manifest: %w", err)
		}
		sha := m.Config.Digest.String()
		size := m.Config.Size
		sr = append(sr, SigningRequest{NotaryGun: base, Version: tag, ByteSize: size, SHA256: sha})
	}
	return sr, nil
}

func (ns NotarySigner) Sign(images []string) error {
	sr, err := ns.buildSigningRequest(images)
	b, err := json.Marshal(sr)
	if err != nil {
		return fmt.Errorf("marshal signing request: %w", err)
	}
	req, err := http.NewRequest("POST", ns.url, bytes.NewReader(b))
	if err != nil {
		return err
	}

	if ns.authFunc != nil {
		req = ns.authFunc(req)
	}

	retries := 5
	var respBody []byte
	var statusCode int
	w := time.NewTicker(ns.retryTimeout)
	defer w.Stop()
	for retries > 0 {
		select {
		case <-w.C:
			resp, err := ns.c.Do(req)
			if err != nil {
				return fmt.Errorf("notary request: %w", err)
			}
			respBody, err = io.ReadAll(resp.Body)
			statusCode = resp.StatusCode
			if err != nil {
				return fmt.Errorf("body read: %w", err)
			}
			switch resp.StatusCode {
			case http.StatusOK:
				// response was fine. Do not need anything else
				return nil
			case http.StatusUnauthorized, http.StatusForbidden, http.StatusBadRequest:
				return fmt.Errorf("notary response: %v %s", resp.StatusCode, resp.Status)
			}
			retries--
		}
	}
	fmt.Println("Reached all retries. Stopping.")
	fmt.Println(respBody)
	return fmt.Errorf("other notary error: %v", statusCode)
}

func (nc NotaryConfig) NewSigner() (Signer, error) {
	var ns NotarySigner

	// (@Ressetkk): Should this be loaded before calling signer?
	if nc.Secret != nil {
		f, err := os.ReadFile(nc.Secret.Path)
		if err != nil {
			return nil, err
		}
		switch nc.Secret.Type {
		case "bearer":
			ns.authFunc = BearerToken(string(f))
		}
	}
	ns.retryTimeout = 10 * time.Second
	if nc.RetryTimeout > 0 {
		ns.retryTimeout = nc.RetryTimeout
	}
	ns.c.Timeout = nc.Timeout
	ns.url = nc.Endpoint
	return ns, nil
}
