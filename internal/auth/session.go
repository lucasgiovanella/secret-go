package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SessionCache representa o cache da sessão no sistema de arquivos
type SessionCache struct {
	MasterKey  []byte    `json:"master_key"`
	LastActive time.Time `json:"last_active"`
}

// SaveSession salva os dados da sessão atual em um arquivo cache
func (a *Auth) SaveSession() error {
	if a.masterKey == nil {
		return ErrPasswordRequired
	}

	// Dados da sessão para salvar
	cache := SessionCache{
		MasterKey:  a.masterKey,
		LastActive: a.lastActivity,
	}

	// Transforma em JSON
	data, err := json.Marshal(cache)
	if err != nil {
		return fmt.Errorf("falha ao serializar sessão: %w", err)
	}

	// Salva no arquivo de sessão
	sessionPath := filepath.Join(a.config.BaseDir, "session.json")
	if err := os.WriteFile(sessionPath, data, 0600); err != nil {
		return fmt.Errorf("falha ao salvar sessão: %w", err)
	}

	return nil
}

// LoadSession tenta carregar uma sessão salva
func (a *Auth) LoadSession() error {
	sessionPath := filepath.Join(a.config.BaseDir, "session.json")

	// Verifica se existe uma sessão salva
	data, err := os.ReadFile(sessionPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrPasswordRequired
		}
		return fmt.Errorf("falha ao ler sessão: %w", err)
	}

	// Deserializa a sessão
	var cache SessionCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return fmt.Errorf("falha ao carregar sessão: %w", err)
	}

	// Verifica se a sessão expirou
	if time.Since(cache.LastActive) > a.config.SessionTimeout {
		if err := os.Remove(sessionPath); err != nil {
			// Apenas log, não falha se não conseguir remover
			fmt.Fprintf(os.Stderr, "Aviso: falha ao remover sessão expirada: %v\n", err)
		}
		return ErrSessionExpired
	}

	// Restaura a sessão
	a.masterKey = cache.MasterKey
	a.lastActivity = time.Now()

	// Salva a sessão com o tempo atualizado
	return a.SaveSession()
}

// ClearSession limpa a sessão salva
func (a *Auth) ClearSession() error {
	sessionPath := filepath.Join(a.config.BaseDir, "session.json")

	// Remove o arquivo de sessão se existir
	if _, err := os.Stat(sessionPath); err == nil {
		if err := os.Remove(sessionPath); err != nil {
			return fmt.Errorf("falha ao limpar sessão: %w", err)
		}
	}

	a.masterKey = nil
	return nil
}
