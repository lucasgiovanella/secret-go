package cli

import (
	"fmt"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/lucasgiovanella/secret-go/internal/app"
	"github.com/spf13/cobra"
)

func NewRootCmd(a *app.App) *cobra.Command {
	root := &cobra.Command{
		Use:   "secretgo",
		Short: "Gerenciador de segredos seguro",
	}

	root.AddCommand(
		newInitCmd(a),
		newLoginCmd(a),
		newLogoutCmd(a),
		newSetCmd(a),
		newGetCmd(a),
		newRemoveCmd(a),
		newListCmd(a),
	)
	return root
}

func newInitCmd(a *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Inicializa o SecretGo",
		RunE: func(cmd *cobra.Command, args []string) error {
			initialized, err := a.Auth.IsInitialized()
			if err != nil {
				return err
			}
			if initialized {
				fmt.Println("Já inicializado.")
				return nil
			}
			var pwd string
			survey.AskOne(&survey.Password{Message: "Nova senha mestra:"}, &pwd)
			if err := a.Auth.Initialize(pwd); err != nil {
				return err
			}
			fmt.Println("Inicializado com sucesso!")
			return nil
		},
	}
}

func newLoginCmd(a *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Autentica com a senha mestra",
		RunE: func(cmd *cobra.Command, args []string) error {
			var pwd string
			survey.AskOne(&survey.Password{Message: "Senha mestra:"}, &pwd)
			if err := a.Auth.Login(pwd); err != nil {
				return err
			}
			fmt.Println("Login realizado!")
			return nil
		},
	}
}

func newLogoutCmd(a *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Encerra a sessão atual",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.Auth.Logout(); err != nil {
				return err
			}
			fmt.Println("Logout realizado com sucesso.")
			return nil
		},
	}
}

func newSetCmd(a *app.App) *cobra.Command {
	var project, env, key, value string
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Define um segredo",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.Auth.RequireAuth(); err != nil {
				return err
			}
			if project == "" {
				survey.AskOne(&survey.Input{Message: "Projeto:"}, &project)
			}
			if env == "" {
				survey.AskOne(&survey.Input{Message: "Ambiente:"}, &env)
			}
			if key == "" {
				survey.AskOne(&survey.Input{Message: "Chave:"}, &key)
			}
			if value == "" {
				survey.AskOne(&survey.Password{Message: "Valor:"}, &value)
			}
			// Criptografa o valor
			iv, _ := a.Auth.Crypto().GenerateIV()
			masterKey, _ := a.Auth.GetMasterKey()
			enc, _ := a.Auth.Crypto().Encrypt([]byte(value), masterKey, iv)
			proj := a.Store.GetOrCreateProject(project)
			envObj := proj.GetOrCreateEnvironment(env)
			envObj.SetSecret(key, enc, iv)
			if err := a.SaveStore(); err != nil {
				return err
			}
			fmt.Println("Segredo salvo!")
			return nil
		},
	}
	cmd.Flags().StringVarP(&project, "project", "p", "", "Projeto")
	cmd.Flags().StringVarP(&env, "env", "e", "", "Ambiente")
	cmd.Flags().StringVarP(&key, "key", "k", "", "Chave")
	cmd.Flags().StringVarP(&value, "value", "v", "", "Valor")
	return cmd
}

func newGetCmd(a *app.App) *cobra.Command {
	var project, env, key string
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Obtém um segredo",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.Auth.RequireAuth(); err != nil {
				return err
			}
			if project == "" {
				survey.AskOne(&survey.Input{Message: "Projeto:"}, &project)
			}
			if env == "" {
				survey.AskOne(&survey.Input{Message: "Ambiente:"}, &env)
			}
			if key == "" {
				survey.AskOne(&survey.Input{Message: "Chave:"}, &key)
			}
			proj, ok := a.Store.Projects[project]
			if !ok {
				return fmt.Errorf("projeto não encontrado")
			}
			envObj, ok := proj.Environments[env]
			if !ok {
				return fmt.Errorf("ambiente não encontrado")
			}
			secret, ok := envObj.Secrets[key]
			if !ok {
				return fmt.Errorf("segredo não encontrado")
			}
			masterKey, _ := a.Auth.GetMasterKey()
			val, err := a.Auth.Crypto().Decrypt(secret.Value, masterKey, secret.IV)
			if err != nil {
				return err
			}
			fmt.Printf("Valor: %s\n", string(val))
			return nil
		},
	}
	cmd.Flags().StringVarP(&project, "project", "p", "", "Projeto")
	cmd.Flags().StringVarP(&env, "env", "e", "", "Ambiente")
	cmd.Flags().StringVarP(&key, "key", "k", "", "Chave")
	return cmd
}

func newRemoveCmd(a *app.App) *cobra.Command {
	var project, env, key string
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove um segredo",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.Auth.RequireAuth(); err != nil {
				return err
			}
			if project == "" {
				survey.AskOne(&survey.Input{Message: "Projeto:"}, &project)
			}
			if env == "" {
				survey.AskOne(&survey.Input{Message: "Ambiente:"}, &env)
			}
			if key == "" {
				survey.AskOne(&survey.Input{Message: "Chave:"}, &key)
			}
			proj, ok := a.Store.Projects[project]
			if !ok {
				return fmt.Errorf("projeto não encontrado")
			}
			envObj, ok := proj.Environments[env]
			if !ok {
				return fmt.Errorf("ambiente não encontrado")
			}
			if !envObj.RemoveSecret(key) {
				return fmt.Errorf("segredo não encontrado")
			}
			if err := a.SaveStore(); err != nil {
				return err
			}
			fmt.Println("Segredo removido!")
			return nil
		},
	}
	cmd.Flags().StringVarP(&project, "project", "p", "", "Projeto")
	cmd.Flags().StringVarP(&env, "env", "e", "", "Ambiente")
	cmd.Flags().StringVarP(&key, "key", "k", "", "Chave")
	return cmd
}

func newListCmd(a *app.App) *cobra.Command {
	var project, env string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lista segredos",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.Auth.RequireAuth(); err != nil {
				return err
			}
			if project == "" {
				survey.AskOne(&survey.Input{Message: "Projeto:"}, &project)
			}
			if env == "" {
				survey.AskOne(&survey.Input{Message: "Ambiente:"}, &env)
			}
			proj, ok := a.Store.Projects[project]
			if !ok {
				return fmt.Errorf("projeto não encontrado")
			}
			envObj, ok := proj.Environments[env]
			if !ok {
				return fmt.Errorf("ambiente não encontrado")
			}
			fmt.Println("Segredos:")
			for k, s := range envObj.Secrets {
				fmt.Printf("- %s (atualizado: %s)\n", k, s.UpdatedAt.Format(time.RFC3339))
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&project, "project", "p", "", "Projeto")
	cmd.Flags().StringVarP(&env, "env", "e", "", "Ambiente")
	return cmd
}
