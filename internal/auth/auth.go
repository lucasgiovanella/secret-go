package auth

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/lucasgiovanella/secret-go/internal/config"
	"github.com/lucasgiovanella/secret-go/internal/crypto"
)

const (
	// Variável de ambiente para a senha mestra
	EnvMasterPassword = "SECRETGO_MASTER_PASSWORD"
)

var (
	ErrPasswordRequired   = errors.New("senha mestra é necessária")
	ErrInvalidPassword    = errors.New("senha mestra inválida")
	ErrSessionExpired     = errors.New("sessão expirada")
	ErrNotInitialized     = errors.New("SecretGo não inicializado, execute 'secretgo init' primeiro")
	ErrAlreadyInitialized = errors.New("SecretGo já foi inicializado")
)

// Auth gerencia a autenticação e autorização
type Auth struct {
	config       *config.Config
	crypto       *crypto.Crypto
	masterKey    []byte
	lastActivity time.Time
	initialized  bool
}

// AuthSession contém informações sobre a sessão atual
type AuthSession struct {
	Key           []byte
	LastActivity  time.Time
	SessionActive bool
}

// NewAuth cria um novo gerenciador de autenticação
func NewAuth(cfg *config.Config) (*Auth, error) {
	auth := &Auth{
		config:       cfg,
		masterKey:    nil,
		lastActivity: time.Now(),
		initialized:  false,
	}

	// Verifica se o SecretGo está inicializado
	initialized, err := auth.IsInitialized()
	if err != nil {
		return nil, err
	}
	auth.initialized = initialized

	if initialized {
		// Carrega o salt do arquivo de chave mestra
		salt, err := auth.loadSalt()
		if err != nil {
			return nil, err
		}

		// Inicializa o crypto com o salt existente
		crypto, err := crypto.NewCrypto(salt)
		if err != nil {
			return nil, err
		}
		auth.crypto = crypto

		// Tenta carregar a senha do ambiente
		if pwd := os.Getenv(EnvMasterPassword); pwd != "" {
			if err := auth.Login(pwd); err != nil {
				// Ignora erro, apenas não faz login automático
			}
		}
	}

	return auth, nil
}

// IsInitialized verifica se o SecretGo foi inicializado
func (a *Auth) IsInitialized() (bool, error) {
	// Verifica se o arquivo da chave mestra existe
	_, err := os.Stat(a.config.MasterKeyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("falha ao verificar inicialização: %w", err)
	}
	return true, nil
}

// Initialize inicializa o SecretGo com uma nova senha mestra
func (a *Auth) Initialize(password string) error {
	if a.initialized {
		return ErrAlreadyInitialized
	}

	// Cria um novo crypto com salt aleatório
	crypto, err := crypto.NewCrypto(nil)
	if err != nil {
		return fmt.Errorf("falha ao inicializar crypto: %w", err)
	}
	a.crypto = crypto

	// Deriva a chave mestra da senha
	key := crypto.DeriveKey(password)
	a.masterKey = key
	a.lastActivity = time.Now()

	// Salva o hash da senha e o salt
	if err := a.saveMasterKey(password); err != nil {
		return err
	}

	a.initialized = true
	return nil
}

// Login autentica o usuário com a senha mestra
func (a *Auth) Login(password string) error {
	if !a.initialized {
		return ErrNotInitialized
	}

	// Carrega o hash esperado do arquivo
	expectedHash, err := a.loadMasterKeyHash()
	if err != nil {
		return err
	}

	// Verifica a senha
	if !a.crypto.VerifyPassword(password, expectedHash) {
		return ErrInvalidPassword
	}

	// Define a chave mestra e atualiza a última atividade
	a.masterKey = a.crypto.DeriveKey(password)
	a.lastActivity = time.Now()

	// Salva a sessão para futuras invocações
	if err := a.SaveSession(); err != nil {
		// Apenas log, login ainda funciona mesmo se falhar em salvar a sessão
		fmt.Fprintf(os.Stderr, "Aviso: falha ao salvar sessão: %v\n", err)
	}

	return nil
}

// RequireAuth garante que o usuário está autenticado
func (a *Auth) RequireAuth() error {
	if !a.initialized {
		return ErrNotInitialized
	}

	// Se já tem a chave mestra, apenas verifica a expiração
	if a.masterKey != nil {
		// Verifica se a sessão expirou
		if time.Since(a.lastActivity) > a.config.SessionTimeout {
			a.masterKey = nil
			a.ClearSession() // Remove sessão expirada
			return ErrSessionExpired
		}

		// Atualiza o timestamp de atividade
		a.lastActivity = time.Now()
		a.SaveSession() // Atualiza o timestamp no arquivo de sessão
		return nil
	}

	// Tenta carregar a sessão do cache
	if err := a.LoadSession(); err != nil {
		// Se não for possível carregar ou a sessão expirou,
		// precisamos de autenticação
		return ErrPasswordRequired
	}

	// Se carregou com sucesso, a sessão é válida
	return nil
}

// Logout limpa a sessão do usuário
func (a *Auth) Logout() error {
	a.masterKey = nil
	return a.ClearSession()
}

// GetSession retorna a sessão atual
func (a *Auth) GetSession() AuthSession {
	return AuthSession{
		Key:           a.masterKey,
		LastActivity:  a.lastActivity,
		SessionActive: a.masterKey != nil,
	}
}

// GetMasterKey retorna a chave mestra, verificando a autenticação
func (a *Auth) GetMasterKey() ([]byte, error) {
	if err := a.RequireAuth(); err != nil {
		return nil, err
	}
	return a.masterKey, nil
}

// GetCrypto retorna o objeto de criptografia
func (a *Auth) Crypto() *crypto.Crypto {
	return a.crypto
}

// saveMasterKey salva o hash da senha mestra e o salt
func (a *Auth) saveMasterKey(password string) error {
	// Gera o hash da senha
	hash := a.crypto.HashPassword(password)

	// Formato do arquivo: hash:salt
	data := fmt.Sprintf("%s:%x", hash, a.crypto.GetSalt())

	// Salva no arquivo com permissões restritas
	if err := os.WriteFile(a.config.MasterKeyPath, []byte(data), 0600); err != nil {
		return fmt.Errorf("falha ao salvar chave mestra: %w", err)
	}

	return nil
}

// loadMasterKeyHash carrega o hash da senha mestra do arquivo
func (a *Auth) loadMasterKeyHash() (string, error) {
	data, err := os.ReadFile(a.config.MasterKeyPath)
	if err != nil {
		return "", fmt.Errorf("falha ao ler chave mestra: %w", err)
	}

	// Formato do arquivo: hash:salt
	parts := strings.SplitN(string(data), ":", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("formato inválido do arquivo de chave mestra")
	}

	return parts[0], nil
}

// loadSalt carrega o salt do arquivo de chave mestra
func (a *Auth) loadSalt() ([]byte, error) {
	data, err := os.ReadFile(a.config.MasterKeyPath)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler salt: %w", err)
	}

	// Formato do arquivo: hash:salt
	parts := strings.SplitN(string(data), ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("formato inválido do arquivo de chave mestra")
	}

	// Decodifica o salt (formato hex)
	salt, err := hex.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("falha ao decodificar salt: %w", err)
	}

	return salt, nil
}
