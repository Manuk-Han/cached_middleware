package infrautil

import (
	"context"
	"go.uber.org/zap"
	"os"
	"time"
)

func LogConnectionResult(log *zap.SugaredLogger, name string, err error) {
	if err != nil {
		log.Errorf("❌ %s connection failed: %v", name, err)
	} else {
		log.Infof("✅ %s connected", name)
	}
}

func RunMessageLoop(
	ctx context.Context,
	log *zap.SugaredLogger,
	maxFails int,
	readFn func() (string, error),
	handler func(string),
) {
	failCount := 0
	ready := false

	for {
		msg, err := readFn()
		if err != nil {
			failCount++
			log.Errorf("📉 message read failed (attempt %d/%d): %v", failCount, maxFails, err)
			if failCount >= maxFails {
				log.Error("❌ listener aborted after max retry limit")
				os.Exit(1)
			}
			time.Sleep(2 * time.Second)
			continue
		}

		if !ready {
			log.Infof("✅ listener ready")
			ready = true
		}

		failCount = 0
		log.Infof("📩 message received: %s", msg)
		handler(msg)
	}
}
