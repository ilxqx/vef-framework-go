package password

// Encoder defines the interface for password encoding and verification.
type Encoder interface {
	// Encode encodes the raw password (e.g., hashing, encrypting).
	// Returns the encoded password or an error if encoding fails.
	Encode(password string) (string, error)
	// Matches verifies whether the raw password matches the encoded password.
	// Returns true if the passwords match, false otherwise.
	Matches(password, encodedPassword string) bool
	// UpgradeEncoding determines whether the encoded password should be re-encoded.
	// This is useful for algorithm migration or cost factor upgrades.
	// Returns true if the password should be upgraded, false otherwise.
	UpgradeEncoding(encodedPassword string) bool
}
