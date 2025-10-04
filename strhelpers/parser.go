package strhelpers

import (
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/internal/log"
)

var logger = log.Named("strhelpers")

const (
	// TagAttrDefaultKey is the default key for tag attributes
	TagAttrDefaultKey = "__default"
)

// ParseTagAttrs parses the tag attributes.
func ParseTagAttrs(tag string) map[string]string {
	attrs := make(map[string]string)
	for attr := range strings.SplitSeq(tag, constants.Comma) {
		attr = strings.TrimSpace(attr)
		if attr == constants.Empty {
			continue // Skip empty attributes
		}

		idx := strings.IndexByte(attr, constants.ByteEquals)
		if idx == -1 {
			if _, ok := attrs[TagAttrDefaultKey]; ok {
				logger.Warnf("Ignoring duplicate default attribute [%s] of tag: %s", attr, tag)
				continue
			}

			attrs[TagAttrDefaultKey] = attr
			continue
		}

		attrs[attr[:idx]] = attr[idx+1:]
	}

	return attrs
}

// ParseTagArgs parses the tag args.
func ParseTagArgs(args string) map[string]string {
	kvs := make(map[string]string)
	for kv := range strings.SplitSeq(args, constants.Space) {
		kv = strings.TrimSpace(kv)
		if kv == constants.Empty {
			continue // Skip empty parameters
		}

		idx := strings.IndexByte(kv, constants.ByteColon)
		if idx == -1 {
			if _, ok := kvs[TagAttrDefaultKey]; ok {
				logger.Warnf("Ignoring duplicate default key [%s] of arg: %s", kv, args)
				continue
			}

			kvs[TagAttrDefaultKey] = kv
			continue
		}

		kvs[kv[:idx]] = kv[idx+1:]
	}

	return kvs
}
