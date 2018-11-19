package cerebro

import (
	"fmt"
	"github.com/aimeelaplant/externalissuesource"
	"github.com/avast/retry-go"
	"log"
	"net/http"
	"strings"
	"time"
)

// An error returned from the http client. Unfortunately it has no variable associated with it.
const errClientTimeoutString = "Client.Timeout exceeded"

// The default retry delay option.
var retryDelay = retry.Delay(time.Duration(10 * time.Second))

// Retries the URL if a connection error is returned.
// The returned string should be the requested url.
func retryURL(f func() (string, error)) error {
	errCh := make(chan error, 1)
	retry.Do(func() error {
		url, err := f()
		if err != nil {
			if isConnectionError(err) {
				log.Println(fmt.Sprintf("got connection error for %s. retrying.", url))
				return err
			}
			errCh <- err
		}
		close(errCh)
		return nil
	}, retryDelay)
	if err, ok := <-errCh; ok {
		return err
	}
	return nil
}

// isConnectionError checks if the error is a connection-related error or not.
func isConnectionError(err error) bool {
	// 	Wish there was a better way to check the client time out error!
	if strings.Contains(err.Error(), errClientTimeoutString) ||
		err == externalissuesource.ErrConnection ||
		err == http.ErrHandlerTimeout ||
		err == http.ErrServerClosed {
		return true
	}
	return false
}
