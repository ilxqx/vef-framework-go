package password

// Predefined encoder identifiers.
type EncoderID string

const (
	EncoderBcrypt    EncoderID = "bcrypt"
	EncoderArgon2    EncoderID = "argon2"
	EncoderScrypt    EncoderID = "scrypt"
	EncoderPbkdf2    EncoderID = "pbkdf2"
	EncoderMd5       EncoderID = "md5"
	EncoderSha256    EncoderID = "sha256"
	EncoderPlaintext EncoderID = "plaintext"
)
