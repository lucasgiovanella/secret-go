package main

import (
	"fmt"
	"os"

	"github.com/lucasgiovanella/secret-go/internal/app"
	"github.com/lucasgiovanella/secret-go/internal/cli"
)

func main() {
	// Inicializa a aplicação
	secretApp, err := app.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao inicializar a aplicação: %v\n", err)
		os.Exit(1)
	}
	defer secretApp.Close()

	// Inicializa e executa a CLI
	rootCmd := cli.NewRootCmd(secretApp)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
		os.Exit(1)
	}
}