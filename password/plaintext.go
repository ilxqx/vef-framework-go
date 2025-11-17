package password

type plaintextEncoder struct{}

// NewPlaintextEncoder creates a new plaintext password encoder.
// WARNING: Provides NO security. Use only for testing/development.
func NewPlaintextEncoder() Encoder {
	return new(plaintextEncoder)
}

func (e *plaintextEncoder) Encode(password string) (string, error) {
	return password, nil
}

func (e *plaintextEncoder) Matches(password, encodedPassword string) bool {
	return password == encodedPassword
}

func (e *plaintextEncoder) UpgradeEncoding(encodedPassword string) bool {
	return true
}
