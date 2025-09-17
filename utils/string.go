package utils

import (
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
)

const (
	defaultKey = "default" // defaultKey is the default key for tag attributes
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
			if _, ok := attrs[defaultKey]; ok {
				logger.Warnf("Ignoring duplicate default attribute [%s] of tag: %s", attr, tag)
				continue
			}

			attrs[defaultKey] = attr
			continue
		}

		attrs[attr[:idx]] = attr[idx+1:]
	}

	return attrs
}

// ParseQueryString parses the query string.
func ParseQueryString(query string) map[string]string {
	kvs := make(map[string]string)
	for kv := range strings.SplitSeq(query, constants.Ampersand) {
		kv = strings.TrimSpace(kv)
		if kv == constants.Empty {
			continue // Skip empty parameters
		}

		idx := strings.IndexByte(kv, constants.ByteEquals)
		if idx == -1 {
			if _, ok := kvs[defaultKey]; ok {
				logger.Warnf("Ignoring duplicate default key [%s] of query: %s", kv, query)
				continue
			}

			kvs[defaultKey] = kv
			continue
		}

		kvs[kv[:idx]] = kv[idx+1:]
	}

	return kvs
}
