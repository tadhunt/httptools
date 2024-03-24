package httptools

import (
	"context"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/tadhunt/retry"
	"github.com/tadhunt/logger"
)

const (
	DL_InitialDelay = 500 * time.Millisecond
	DL_MaxDelay     = 2 * time.Second
	DL_MaxTries     = 5
)

func Download(srcUrl string, dstPath string) error {
	log := logger.NewCompatLogWriter(logger.LogLevel_DEBUG)

	f, err := os.Create(dstPath)
	if err != nil {
		return log.ErrFmt("%s: create %s: %v", srcUrl, dstPath, err)
	}

	defer f.Close()

	return DownloadTo(srcUrl, f)
}

func DownloadTo(srcUrl string, dst io.Writer) error {
	log := logger.NewCompatLogWriter(logger.LogLevel_DEBUG)

	retrier := retry.NewRetrier(DL_MaxTries, DL_InitialDelay, DL_MaxDelay)

	start := time.Now()
	last := start
	attempt := 0

	ctx := context.Background()

	err := retrier.RunContext(ctx, func(c context.Context) error {
		now := time.Now()
		total := now.Sub(start)
		sinceLast := now.Sub(last)
		last = now

		attempt++

		log.Debugf("%s: attempt %d (elapsed %s, since last %s)", srcUrl, attempt, total, sinceLast)

		r, err := http.Get(srcUrl)
		if err != nil {
			return retry.Stop(log.ErrFmt("%s: %v", srcUrl, err))
		}
		defer r.Body.Close()

		switch {
		default:
			return retry.Stop(log.ErrFmt("%s: %v", srcUrl, r.Status))
		case r.StatusCode == http.StatusOK:
			_, err = io.Copy(dst, r.Body)
			if err != nil {
				return retry.Stop(log.ErrFmt("%s: %v", srcUrl, err))
			}
			return nil
		case r.StatusCode >= 500:
			return log.ErrFmt("retry")
		}
	})

	if err != nil {
		return err
	}

	return nil
}
