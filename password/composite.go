package password

import (
	"fmt"
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
)

type compositeEncoder struct {
	defaultEncoderId EncoderId
	encoders         map[EncoderId]Encoder
}

// NewCompositeEncoder creates a composite encoder that supports multiple password formats.
// The defaultEncoderId specifies which encoder to use for new passwords.
// Encoders map contains encoder ID to Encoder implementations.
func NewCompositeEncoder(defaultEncoderId EncoderId, encoders map[EncoderId]Encoder) Encoder {
	return &compositeEncoder{
		defaultEncoderId: defaultEncoderId,
		encoders:         encoders,
	}
}

func (c *compositeEncoder) Encode(password string) (string, error) {
	encoder, ok := c.encoders[c.defaultEncoderId]
	if !ok {
		return constants.Empty, fmt.Errorf("%w: %s", ErrDefaultEncoderNotFound, c.defaultEncoderId)
	}

	encoded, err := encoder.Encode(password)
	if err != nil {
		return constants.Empty, err
	}

	return fmt.Sprintf("{%s}%s", c.defaultEncoderId, encoded), nil
}

func (c *compositeEncoder) Matches(password, encodedPassword string) bool {
	encoderId := c.extractEncoderId(encodedPassword)
	if encoderId == EncoderId(constants.Empty) {
		encoderId = c.defaultEncoderId
	}

	encoder, ok := c.encoders[encoderId]
	if !ok {
		return false
	}

	rawEncoded := c.stripPrefix(encodedPassword)
	return encoder.Matches(password, rawEncoded)
}

func (c *compositeEncoder) UpgradeEncoding(encodedPassword string) bool {
	encoderId := c.extractEncoderId(encodedPassword)

	if encoderId != EncoderId(constants.Empty) && encoderId != c.defaultEncoderId {
		return true
	}

	encoder, ok := c.encoders[c.defaultEncoderId]
	if !ok {
		return false
	}

	rawEncoded := c.stripPrefix(encodedPassword)
	return encoder.UpgradeEncoding(rawEncoded)
}

func (c *compositeEncoder) extractEncoderId(encodedPassword string) EncoderId {
	if !strings.HasPrefix(encodedPassword, "{") {
		return EncoderId(constants.Empty)
	}
	end := strings.Index(encodedPassword, "}")
	if end == -1 {
		return EncoderId(constants.Empty)
	}
	return EncoderId(encodedPassword[1:end])
}

func (c *compositeEncoder) stripPrefix(encodedPassword string) string {
	if !strings.HasPrefix(encodedPassword, "{") {
		return encodedPassword
	}
	end := strings.Index(encodedPassword, "}")
	if end == -1 {
		return encodedPassword
	}
	return encodedPassword[end+1:]
}
