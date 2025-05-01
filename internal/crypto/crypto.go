package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

const (
	// Parâmetros para Argon2
	argon2Time    = 3
	argon2Memory  = 64 * 1024
	argon2Threads = 4
	argon2KeyLen  = 32 // 256 bits para AES-256
	
	// Tamanho do salt para derivação de chave
	saltSize = 16
	
	// Tamanho do IV para AES
	ivSize = 12 // Para GCM
)

// Crypto lida com todas as operações criptográficas
type Crypto struct {
	salt []byte
}

// NewCrypto cria uma nova instância do gerenciador de criptografia
func NewCrypto(salt []byte) (*Crypto, error) {
	// Se não for fornecido um salt, gera um novo
	if salt == nil {
		salt = make([]byte, saltSize)
		if _, err := io.ReadFull(rand.Reader, salt); err != nil {
			return nil, fmt.Errorf("falha ao gerar salt: %w", err)
		}
	}

	return &Crypto{
		salt: salt,
	}, nil
}

// GetSalt retorna o salt usado para derivação de chave
func (c *Crypto) GetSalt() []byte {
	return c.salt
}

// DeriveKey deriva uma chave criptográfica da senha usando Argon2id
func (c *Crypto) DeriveKey(password string) []byte {
	return argon2.IDKey([]byte(password), c.salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)
}

// HashPassword gera um hash da senha para verificação
func (c *Crypto) HashPassword(password string) string {
	key := c.DeriveKey(password)
	return hex.EncodeToString(key)
}

// VerifyPassword verifica se a senha está correta
func (c *Crypto) VerifyPassword(password, expectedHash string) bool {
	actualHash := c.HashPassword(password)
	return actualHash == expectedHash
}

// GenerateIV gera um novo vetor de inicialização aleatório
func (c *Crypto) GenerateIV() ([]byte, error) {
	iv := make([]byte, ivSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("falha ao gerar IV: %w", err)
	}
	return iv, nil
}

// Encrypt criptografa dados com a chave e IV fornecidos
func (c *Crypto) Encrypt(data []byte, key []byte, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar cifra AES: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar GCM: %w", err)
	}

	// Criptografa e aplica autenticação
	ciphertext := aesgcm.Seal(nil, iv, data, nil)
	return ciphertext, nil
}

// Decrypt descriptografa dados com a chave e IV fornecidos
func (c *Crypto) Decrypt(ciphertext []byte, key []byte, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar cifra AES: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar GCM: %w", err)
	}

	// Descriptografa e verifica autenticação
	plaintext, err := aesgcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("falha ao descriptografar: %w", err)
	}

	return plaintext, nil
}

// ComputeHMAC calcula um HMAC-SHA256 dos dados
func (c *Crypto) ComputeHMAC(data []byte, key []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// VerifyHMAC verifica a integridade dos dados usando HMAC
func (c *Crypto) VerifyHMAC(data []byte, key []byte, expectedMAC []byte) bool {
	actualMAC := c.ComputeHMAC(data, key)
	return hmac.Equal(actualMAC, expectedMAC)
}