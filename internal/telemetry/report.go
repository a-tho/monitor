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

func (o Observer) report(ctx context.Context, metric []*monitor.Metrics) error {
	// Prepare request url
	url := fmt.Sprintf("http://%s/%s/", o.SrvAddr, server.UpdsPath)
	// Prepare request body
	var buf bytes.Buffer
	compressBuf := gzip.NewWriter(&buf)
	enc := json.NewEncoder(compressBuf)
	if err := enc.Encode(metric); err != nil {
		return err
	}
	compressBuf.Close()

	// Prepare and send request
	err := retry.Do(ctx, func(context.Context) error {
		client := resty.New()
		body := buf.Bytes()
		req := client.R().
			SetBody(body).
			SetHeader(contentEncoding, encodingGzip).
			SetHeader(contentType, typeApplicationJSON).
			SetContext(ctx)

		// sign request body if necessary
		if len(o.signKey) > 0 {
			req.SetHeader(bodySignature, o.signature(body))
		}

		_, err := req.Post(url)

		return o.retryIfNetError(err)
	})
	return err
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
