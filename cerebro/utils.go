package cerebro

import (
	"github.com/aimeelaplant/comiccruncher/internal/log"
	"github.com/avast/retry-go"
	"go.uber.org/zap"
	"time"
)

// The default retry delay option.
var retryDelay = retry.Delay(time.Duration(10 * time.Second))

// Retries the func if a connection error is returned.
// The returned string should be the requested url.
func retryConnectionError(f func() (string, error)) error {
	errCh := make(chan error, 1)
	retry.Do(func() error {
		url, err := f()
		if err != nil {
			if isConnectionError(err) {
				log.CEREBRO().Info("got connecting error. retrying.", zap.String("url", url))
				return err
			}
			errCh <- err
			return nil
		}
		close(errCh)
		return nil
	}, retryDelay)
	if err, ok := <-errCh; ok {
		close(errCh)
		return err
	}
	return nil
}
