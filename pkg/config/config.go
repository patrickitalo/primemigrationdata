package config

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

// Função para solicitar as opções de migração
func SolicitarOpcoes(sistema string) ([]string, *string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Selecione o que deseja migrar (separado por vírgula):")

	// Exibir menu baseado no sistema
	if sistema == "FCERTA" {
		fmt.Println("1 - Clientes, 2 - Médicos, 3 - Fornecedores, 4 - Produtos, 5 - Lotes, 6 - Produção Interna, 7 - Refazer - Historico Vendas")
	} else {
		fmt.Println("1 - Clientes, 2 - Médicos, 3 - Fornecedores, 4 - Produtos, 5 - Forma Farmacêutica, 6 - Lotes, 7 - Produção Interna, 8 - Refazer - Historico Vendas")
	}

	fmt.Print("Opções: ")

	options, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Erro ao ler opções: %v", err)
		return nil, nil
	}

	options = strings.TrimSpace(options)
	if options == "" {
		fmt.Println("Nenhuma opção foi fornecida")
		fmt.Println("Pressione ENTER para sair...")
		fmt.Scanln()
		os.Exit(1)
	}

	selectedOptions := strings.Split(options, ",")

	// Limpar e validar opções
	var opcoesValidas []string
	for _, option := range selectedOptions {
		option = LimparOpcao(option)
		if option == "" {
			continue
		}

		if !ValidarOpcao(option, sistema) {
			log.Printf("Aviso: opção '%s' não é válida e será ignorada", option)
			continue
		}

		opcoesValidas = append(opcoesValidas, option)
	}

	if len(opcoesValidas) == 0 {
		fmt.Println("Nenhuma opção válida foi fornecida")
		fmt.Println("Pressione ENTER para sair...")
		fmt.Scanln()
		os.Exit(1)
	}

	var vVencido *string
	// Só solicita vVencido para FCERTA, não para PRISMA5
	if sistema == "FCERTA" {
		for _, option := range opcoesValidas {
			if option == "5" {
				fmt.Print("Digite o parâmetro para os lotes (-1 para incluir todos, 0 para apenas não vencidos): ")
				vVencidoInput, err := reader.ReadString('\n')
				if err != nil {
					log.Printf("Erro ao ler parâmetro vVencido: %v", err)
					continue
				}
				vVencidoStr := strings.TrimSpace(vVencidoInput)
				if vVencidoStr == "" {
					fmt.Println("Parâmetro vVencido é obrigatório para o tipo 5")
					fmt.Println("Pressione ENTER para sair...")
					fmt.Scanln()
					os.Exit(1)
				}
				vVencido = &vVencidoStr
				break
			}
		}
	}

	return opcoesValidas, vVencido
}

// Função para aguardar o usuário antes de sair
func WaitForUser() {
	fmt.Println("Pressione Enter para sair...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

// LimparOpcao remove espaços em branco e valida a opção
func LimparOpcao(opcao string) string {
	return strings.TrimSpace(opcao)
}

// ValidarOpcao verifica se a opção é válida baseada no sistema
func ValidarOpcao(opcao string, sistema string) bool {
	var opcoesValidas []string

	if sistema == "FCERTA" {
		// FCERTA não tem opção 8 (Forma Farmacêutica)
		opcoesValidas = []string{"1", "2", "3", "4", "5", "6", "7"}
	} else {
		// Outros sistemas têm todas as opções
		opcoesValidas = []string{"1", "2", "3", "4", "5", "6", "7", "8"}
	}

	for _, valida := range opcoesValidas {
		if opcao == valida {
			return true
		}
	}
	return false
}
