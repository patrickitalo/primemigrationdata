package config

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

// Função para obter as configurações do usuário
func LoadConfig(modoSetup bool, sistemaNome string) (string, string, string, string, string, string) {
	reader := bufio.NewReader(os.Stdin)

	// Mensagem dinâmica baseada no sistema
	var mensagemBanco string
	switch sistemaNome {
	case "FCERTA":
		mensagemBanco = "Digite o caminho do banco de dados FormulaCerta: "
	case "PRISMA5":
		mensagemBanco = "Digite o caminho do banco de dados PRISMA5: "
	default:
		mensagemBanco = fmt.Sprintf("Digite o caminho do banco de dados %s: ", sistemaNome)
	}

	fmt.Print(mensagemBanco)
	dbPath, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Erro ao ler caminho do banco: %v", err)
		dbPath = ""
	}
	dbPath = strings.TrimSpace(dbPath)
	// Remover aspas duplas do caminho do banco
	dbPath = removeQuotes(dbPath)
	if dbPath == "" {
		fmt.Println("Caminho do banco de dados é obrigatório")
		fmt.Println("Pressione ENTER para sair...")
		fmt.Scanln()
		os.Exit(1)
	}

	fmt.Print("Digite o nome de usuário do banco de dados (padrão: sysdba): ")
	user, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Erro ao ler usuário: %v", err)
		user = ""
	}
	user = strings.TrimSpace(user)
	if user == "" {
		user = "sysdba"
	}

	fmt.Print("Digite a senha do banco de dados (padrão: pharmacie): ")
	password, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Erro ao ler senha: %v", err)
		password = ""
	}
	password = strings.TrimSpace(password)
	if password == "" {
		password = "pharmacie"
	}

	fmt.Print("Digite o endereço do host do banco de dados (padrão: localhost): ")
	host, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Erro ao ler host: %v", err)
		host = ""
	}
	host = strings.TrimSpace(host)
	if host == "" {
		host = "localhost"
	}

	// Validar formato do host
	if !isValidHost(host) {
		fmt.Printf("❌ Host '%s' não é válido! Use formato IP ou nome de domínio.\n", host)
		fmt.Println("Pressione ENTER para tentar novamente...")
		fmt.Scanln()
		return LoadConfig(modoSetup, sistemaNome) // Recursão para tentar novamente
	}

	fmt.Print("Digite a porta do banco de dados (padrão: 3050): ")
	port, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Erro ao ler porta: %v", err)
		port = ""
	}
	port = strings.TrimSpace(port)
	if port == "" {
		port = "3050"
	}

	// Validar formato da porta
	if !isValidPort(port) {
		fmt.Printf("❌ Porta '%s' não é válida! Use um número entre 1 e 65535.\n", port)
		fmt.Println("Pressione ENTER para tentar novamente...")
		fmt.Scanln()
		return LoadConfig(modoSetup, sistemaNome) // Recursão para tentar novamente
	}

	conversao := ""
	if !modoSetup {
		fmt.Print("Digite o valor da conversão (1, 2, etc.): ")
		conversao, err = reader.ReadString('\n')
		if err != nil {
			log.Printf("Erro ao ler conversão: %v", err)
			conversao = ""
		}
		conversao = strings.TrimSpace(conversao)
		if conversao == "" {
			fmt.Println("Valor de conversão é obrigatório")
			fmt.Println("Pressione ENTER para sair...")
			fmt.Scanln()
			os.Exit(1)
		}
	}

	return dbPath, user, password, host, port, conversao
}

// Função para ser usada no modo de configuração (setup)
func LoadConfigSetup(sistemaNome string) (string, string, string, string, string, string) {
	return LoadConfig(true, sistemaNome)
}

// removeQuotes remove aspas duplas do início e fim da string
func removeQuotes(s string) string {
	s = strings.TrimSpace(s)
	// Remove aspas duplas do início e fim
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

// isValidHost valida se o host é válido (IP ou nome de domínio)
func isValidHost(host string) bool {
	if host == "" {
		return false
	}

	// Padrão para IP (IPv4)
	ipPattern := `^(\d{1,3}\.){3}\d{1,3}$`
	ipRegex := regexp.MustCompile(ipPattern)

	// Padrão para nome de domínio
	domainPattern := `^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`
	domainRegex := regexp.MustCompile(domainPattern)

	// Aceita localhost, IP ou nome de domínio
	return host == "localhost" || ipRegex.MatchString(host) || domainRegex.MatchString(host)
}

// isValidPort valida se a porta é válida (1-65535)
func isValidPort(port string) bool {
	if port == "" {
		return false
	}

	// Padrão para número de porta
	portPattern := `^\d+$`
	portRegex := regexp.MustCompile(portPattern)

	if !portRegex.MatchString(port) {
		return false
	}

	// Converter para número e verificar range
	portNum := 0
	for _, char := range port {
		portNum = portNum*10 + int(char-'0')
	}

	return portNum >= 1 && portNum <= 65535
}
