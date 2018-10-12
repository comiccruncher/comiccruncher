package cerebro

import (
	"github.com/avast/retry-go"
	"time"
	"strings"
	"github.com/aimeelaplant/externalissuesource"
	"net/http"
	"log"
	"fmt"
)

// The default retry delay option.
var retryDelay = retry.Delay(time.Duration(10 * time.Second))

// Retries the func if a connection error is returned.
// The returned string should be the requested url.
func retryConnectionError(f func() (string, error)) error {
	errCh := make(chan error, 1)
	defer close(errCh)
	retry.Do(func() error {
		url, err := f()
		if err != nil {
			if isConnectionError(err) {
				log.Println(fmt.Sprintf("got connection error for %s. retrying.", url))
				return err
			}
			errCh <- err
			return nil
		}
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
