package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

const (
	// DefaultSessionTimeout é o tempo padrão de expiração da sessão
	DefaultSessionTimeout = 15 * time.Minute

	// ConfigFileName é o nome do arquivo de configuração
	ConfigFileName = "config.json"

	// MasterKeyFileName é o nome do arquivo que guarda a chave mestra
	MasterKeyFileName = "master.key"

	// StoreDBName é o nome do banco de dados SQLite
	StoreDBName = "store.db"
)

// Config representa a configuração da aplicação
type Config struct {
	// Diretório base onde os arquivos são armazenados
	BaseDir string

	// Timeout da sessão (quanto tempo a senha é válida)
	SessionTimeout time.Duration

	// Caminhos completos para os arquivos
	ConfigFilePath string
	MasterKeyPath  string
	StoreDBPath    string
}

// NewConfig cria uma nova configuração com valores padrão
func NewConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("falha ao obter o diretório home: %w", err)
	}

	baseDir := filepath.Join(homeDir, ".secretgo")

	// Cria o diretório base se não existir
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return nil, fmt.Errorf("falha ao criar diretório base: %w", err)
	}

	config := &Config{
		BaseDir:        baseDir,
		SessionTimeout: DefaultSessionTimeout,
		ConfigFilePath: filepath.Join(baseDir, ConfigFileName),
		MasterKeyPath:  filepath.Join(baseDir, MasterKeyFileName),
		StoreDBPath:    filepath.Join(baseDir, StoreDBName),
	}

	// Carrega configurações do arquivo, se existir
	if err := config.Load(); err != nil {
		return nil, err
	}

	return config, nil
}

// Load carrega as configurações do arquivo
func (c *Config) Load() error {
	// Verifica se o arquivo de configuração existe
	_, err := os.Stat(c.ConfigFilePath)
	fileExists := !os.IsNotExist(err)
	if err != nil && !os.IsNotExist(err) {
		// Erro inesperado ao verificar o arquivo (ex: permissão no diretório)
		return fmt.Errorf("falha ao verificar arquivo de configuração: %w", err)
	}

	// Se o arquivo não existe, cria um com valores padrão
	if !fileExists {
		if saveErr := c.Save(); saveErr != nil {
			return fmt.Errorf("falha ao criar arquivo de configuração padrão: %w", saveErr)
		}
		// Arquivo criado, podemos retornar nil pois os padrões já estão em c.SessionTimeout
		return nil
	}

	// Se o arquivo existe, tenta ler
	viper.SetConfigFile(c.ConfigFilePath)
	viper.SetConfigType("json")
	if readErr := viper.ReadInConfig(); readErr != nil {
		// Se a leitura falhar mesmo o arquivo existindo (ex: formato inválido, permissão de leitura)
		return fmt.Errorf("falha ao ler arquivo de configuração existente: %w", readErr)
	}

	// Lê o timeout da sessão do arquivo lido
	if viper.IsSet("session_timeout") {
		c.SessionTimeout = viper.GetDuration("session_timeout")
	}

	return nil
}

// Save salva as configurações no arquivo
func (c *Config) Save() error {
	// Define as configurações no Viper
	viper.Set("session_timeout", c.SessionTimeout.String())

	// Tenta salvar no arquivo, criando diretórios se necessário
	// Usar SafeWriteConfigAs para garantir que o path/arquivo seja criado.
	if err := viper.SafeWriteConfigAs(c.ConfigFilePath); err != nil {
		// Se SafeWriteConfigAs falhar mesmo assim (ex: permissão), retorna o erro
		return fmt.Errorf("falha ao salvar arquivo de configuração: %w", err)
	}

	return nil
}

// SetSessionTimeout define o timeout da sessão
func (c *Config) SetSessionTimeout(timeout time.Duration) error {
	c.SessionTimeout = timeout
	return c.Save()
}
