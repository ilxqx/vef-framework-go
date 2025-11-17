package password

// Predefined encoder identifiers.
type EncoderId string

const (
	EncoderBcrypt    EncoderId = "bcrypt"
	EncoderArgon2    EncoderId = "argon2"
	EncoderScrypt    EncoderId = "scrypt"
	EncoderPbkdf2    EncoderId = "pbkdf2"
	EncoderMd5       EncoderId = "md5"
	EncoderSha256    EncoderId = "sha256"
	EncoderPlaintext EncoderId = "plaintext"
)
