# SecretGo

**SecretGo** é um gerenciador de segredos seguro, local e multiplataforma, escrito em Go. Ele permite armazenar, consultar, listar e remover segredos de forma criptografada, organizando-os por projeto e ambiente.

## Funcionalidades

- Armazenamento seguro de segredos usando criptografia AES-GCM.
- Organização dos segredos por projeto e ambiente.
- Persistência local em banco SQLite.
- CLI interativa e amigável.
- Sessão autenticada com expiração configurável.
- Suporte a múltiplos comandos: `init`, `login`, `logout`, `set`, `get`, `remove`, `list`.

## Instalação

1. **Clone o repositório:**
   ```sh
   git clone https://github.com/lucasgiovanella/secret-go.git
   cd secret-go
   ```

2. **Compile o projeto:**
   ```sh
   go build -o secretgo ./cmd/secretgo
   ```

## Como usar

### Inicialização

Antes de usar, inicialize o SecretGo:

```sh
./secretgo init
```
Você será solicitado a criar uma senha mestra.

### Login

Autentique-se para iniciar uma sessão:

```sh
./secretgo login
```

### Logout

Encerre a sessão atual:

```sh
./secretgo logout
```

### Adicionar/Atualizar um segredo

```sh
./secretgo set -p <projeto> -e <ambiente> -k <chave> -v <valor>
```
Se algum parâmetro não for informado, será solicitado interativamente.

### Consultar um segredo

```sh
./secretgo get -p <projeto> -e <ambiente> -k <chave>
```

### Listar segredos

```sh
./secretgo list -p <projeto> -e <ambiente>
```

### Remover um segredo

```sh
./secretgo remove -p <projeto> -e <ambiente> -k <chave>
```

## Estrutura dos dados

- Os segredos são organizados em:
  - **Projeto**
    - **Ambiente** (ex: dev, prod)
      - **Chave** (ex: DATABASE_URL)

- Todos os segredos são criptografados com uma chave derivada da senha mestra.

## Segurança

- A senha mestra nunca é salva em texto claro.
- O arquivo `master.key` armazena apenas o hash da senha e o salt.
- Os segredos são criptografados com AES-GCM e armazenados em um banco SQLite local.
- A sessão expira após 15 minutos (padrão), configurável em `~/.secretgo/config.json`.

## Configuração

O arquivo de configuração fica em `~/.secretgo/config.json` e permite ajustar o tempo de expiração da sessão:

```json
{
  "session_timeout": "15m0s"
}
```

## Dependências

- [Go 1.18+](https://golang.org/)
- [Cobra](https://github.com/spf13/cobra)
- [Survey](https://github.com/AlecAivazis/survey)
- [Viper](https://github.com/spf13/viper)
- [SQLite3](https://github.com/mattn/go-sqlite3)

## Licença

MIT

---

**Autor:** [Lucas Giovanella](https://github.com/lucasgiovanella)
