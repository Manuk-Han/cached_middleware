package strategy

import (
	"cache/config"
	"cache/interface"
	"fmt"
)

type versionedStrategy struct {
	delimiter      string
	defaultVersion int
}

func NewVersionedKeyStrategy(cfg config.VersionedStrategy) _interface.IInvalidationStrategy {
	return &versionedStrategy{
		delimiter:      cfg.Delimiter,
		defaultVersion: cfg.DefaultVersion,
	}
}

func (v *versionedStrategy) GenerateKey(topic string, key string) string {
	return topic + ":" + key + v.delimiter + toString(v.defaultVersion)
}

func (v *versionedStrategy) ComputeTTL(baseTTL int) int {
	return baseTTL
}

func toString(i int) string {
	return fmt.Sprintf("%d", i)
}
