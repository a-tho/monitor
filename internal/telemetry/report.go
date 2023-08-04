package telemetry

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"

	"github.com/go-resty/resty/v2"

	monitor "github.com/a-tho/monitor/internal"
	"github.com/a-tho/monitor/internal/retry"
	"github.com/a-tho/monitor/internal/server"
)

const (
	contentEncoding     = "Content-Encoding"
	contentType         = "Content-Type"
	encodingGzip        = "gzip"
	typeApplicationJSON = "application/json"
	bodySignature       = "HashSHA256"
)
}

func (o Observer) signature(body []byte) string {
	hash := hmac.New(sha256.New, o.signKey)
	hash.Write(body)
	sum := hash.Sum(nil)
	return base64.StdEncoding.EncodeToString(sum)
}

func (o Observer) retryIfNetError(err error) error {
	if err != nil {
		var netErr *net.OpError
		if errors.As(err, &netErr) {
			return retry.RetriableError(err)
		}
	}
	return err
}
