package middleware

import (
	"crypto/hmac"
	"encoding/base64"
	"bytes"
	"io"
	"crypto/sha256"
	"net/http"
)

const (
	bodySignature = "HashSHA256"

	errReadBody = "unreadable request body"
	errWriteBody  = "failed to write response body"
	errSignature = "invalid signature"
)

type hashResponseWriter struct {
	body bytes.Buffer
	http.ResponseWriter
}

func (hrw hashResponseWriter) Write(p []byte) (n int, err error) {
	return hrw.body.Write(p)
}

func WithSigning(handler func(w http.ResponseWriter, r *http.Request), key string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if key != "" {
			// Check signature if body is not empty
			// 1. recreate signature if body is not empty
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, errReadBody, http.StatusBadRequest)
				return
			}
			if len(body) != 0 {
				signWant := hash(body, key)
				// 2. retrieve received signature
				signGotEncoded := r.Header.Values(bodySignature)
				if len(signGotEncoded) == 0 {
					http.Error(w, errSignature, http.StatusBadRequest)
					return
				}
				signGot, err := base64.StdEncoding.DecodeString(signGotEncoded[0])
				if err != nil {
					http.Error(w, errSignature, http.StatusBadRequest)
					return
				}
				// 3. compare the two
				if !hmac.Equal(signGot, signWant) {
					http.Error(w, errReadBody, http.StatusBadRequest)
					return
				}
			}
			// 4. put the body back to let the handler use it
			r.Body = io.NopCloser(bytes.NewBuffer(body))

			// Prepare response body to be able to sign it
			hashW := hashResponseWriter{ResponseWriter: w}

			handler(hashW, r)

			// Sign response body
			// 1. add response body signature
			body = hashW.body.Bytes()
			sign := hash(body, key)
			signEncoded := base64.StdEncoding.EncodeToString(sign)
			w.Header().Add(bodySignature, signEncoded)
			// 2. send the body that we have been caching so far
			_, err = io.Copy(w, &hashW.body)
			if err != nil {
				http.Error(w, errWriteBody, http.StatusInternalServerError)
				return
			}
		} else {
			handler(w, r)
		}
	}
}

func hash(value []byte, key string) []byte {
	hash := hmac.New(sha256.New, []byte(key))
	hash.Write(value)
	return hash.Sum(nil)
}
