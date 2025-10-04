package crypto

// Encryptor defines the interface for encrypting plaintext data.
type Encryptor interface {
	// Encrypt encrypts the plaintext string and returns the encrypted string.
	// The returned string is typically base64-encoded or hex-encoded.
	// Returns an error if encryption fails.
	Encrypt(plaintext string) (string, error)
}

// Decryptor defines the interface for decrypting ciphertext data.
type Decryptor interface {
	// Decrypt decrypts the encrypted string and returns the plaintext string.
	// The encrypted string is typically base64-encoded or hex-encoded.
	// Returns an error if decryption fails (e.g., invalid format, wrong key, corrupted data).
	Decrypt(ciphertext string) (string, error)
}

// Cipher defines the interface for both encryption and decryption operations.
type Cipher interface {
	Encryptor
	Decryptor
}
