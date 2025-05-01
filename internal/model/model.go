package model

import "time"

// Secret representa um segredo individual
type Secret struct {
	Key       string    `json:"key"`
	Value     []byte    `json:"value"` 
	IV        []byte    `json:"iv"`    
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Environment representa um ambiente (dev, prod, etc)
type Environment struct {
	Name    string             `json:"name"`
	Secrets map[string]*Secret `json:"secrets"`
}

// Project representa um projeto
type Project struct {
	Name         string                  `json:"name"`
	Environments map[string]*Environment `json:"environments"`
}

// Store representa o armazenamento completo
type Store struct {
	Version  string              `json:"version"`
	Projects map[string]*Project `json:"projects"`
}

// NewStore cria um novo armazenamento
func NewStore() *Store {
	return &Store{
		Version:  "1.0.0",
		Projects: make(map[string]*Project),
	}
}

// GetOrCreateProject retorna o projeto, criando-o se não existir
func (s *Store) GetOrCreateProject(name string) *Project {
	if project, ok := s.Projects[name]; ok {
		return project
	}

	project := &Project{
		Name:         name,
		Environments: make(map[string]*Environment),
	}
	s.Projects[name] = project
	return project
}

// GetOrCreateEnvironment retorna o ambiente, criando-o se não existir
func (p *Project) GetOrCreateEnvironment(name string) *Environment {
	if env, ok := p.Environments[name]; ok {
		return env
	}

	env := &Environment{
		Name:    name,
		Secrets: make(map[string]*Secret),
	}
	p.Environments[name] = env
	return env
}

// SetSecret adiciona ou atualiza um segredo
func (e *Environment) SetSecret(key string, value []byte, iv []byte) *Secret {
	now := time.Now()

	if s, ok := e.Secrets[key]; ok {
		s.Value = value
		s.IV = iv
		s.UpdatedAt = now
		return s
	}

	secret := &Secret{
		Key:       key,
		Value:     value,
		IV:        iv,
		CreatedAt: now,
		UpdatedAt: now,
	}
	e.Secrets[key] = secret
	return secret
}

// GetSecret retorna um segredo pelo nome
func (e *Environment) GetSecret(key string) (*Secret, bool) {
	secret, ok := e.Secrets[key]
	return secret, ok
}

// RemoveSecret remove um segredo pelo nome
func (e *Environment) RemoveSecret(key string) bool {
	if _, ok := e.Secrets[key]; ok {
		delete(e.Secrets, key)
		return true
	}
	return false
}