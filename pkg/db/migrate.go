package db

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/primesoftwaresi/prime-migration/pkg/db/prisma5"
	"github.com/primesoftwaresi/prime-migration/pkg/logger"
)

// MigrarExtras opcional: modo incremental e callbacks PRISMA5 (histórico central).
// FCERTA em modo incremental não é suportado (procedures sem retorno por registro).
type MigrarExtras struct {
	Incremental     bool
	PRISMACallbacks *prisma5.MigrationCallbacks
}

func MigrarDados(db *sql.DB, tipo string, conversao string, vVencido *string, sistema string, extra *MigrarExtras) error {
	startTime := time.Now()
	logger.Info("Iniciando migração do tipo %s para sistema %s", tipo, sistema)

	// Validar parâmetros de entrada
	if err := validarParametros(tipo, conversao, vVencido, sistema); err != nil {
		logger.Error("Validação de parâmetros falhou para tipo %s: %v", tipo, err)
		return fmt.Errorf("validação de parâmetros falhou para tipo %s: %w", tipo, err)
	}

	if extra != nil && extra.Incremental && sistema == "FCERTA" {
		return fmt.Errorf("migração incremental para FCERTA não é suportada nesta versão (use migração completa ou evolua procedures/consultas de apoio)")
	}

	// PRISMA5 agora usa código Go direto, não procedures SQL
	if sistema == "PRISMA5" {
		return MigrarDadosPRISMA5(db, tipo, conversao, extra)
	}

	// FCERTA ainda usa procedures SQL
	var query string

	if tipo == "5" && vVencido != nil {
		log.Printf("Executando migração de lotes com parâmetro vVencido: %s", *vVencido)

		// Usar procedureName para obter o nome correto baseado no sistema
		procName := procedureName(tipo, sistema)
		if procName == "" {
			return fmt.Errorf("tipo de migração inválido: %s para sistema %s", tipo, sistema)
		}

		// Converter vVencido para integer se necessário
		var vVencidoInt int
		var err error
		if vVencidoInt, err = strconv.Atoi(*vVencido); err != nil {
			return fmt.Errorf("erro ao converter vVencido para integer: %w", err)
		}

		query = fmt.Sprintf("EXECUTE PROCEDURE %s(%d)", procName, vVencidoInt)
		log.Printf("DEBUG: Query para lotes: %s", query)

		// Executar diretamente sem Prepare (método que funciona)
		_, err = db.Exec(query)
		if err != nil {
			return fmt.Errorf("erro ao executar migração do tipo %s com parâmetro %s: %w", tipo, *vVencido, err)
		}
		log.Printf("Migração de lotes executada com sucesso em %v", time.Since(startTime))

	} else if tipo == "7" {
		log.Printf("Executando migração de histórico de vendas (tipo 7)")

		// FCERTA: executar ambas as procedures sem parâmetros
		procedures := []string{"EXTRACT_DATA_REFAZER_1", "EXTRACT_DATA_REFAZER_2"}

		for i, proc := range procedures {
			log.Printf("Executando procedure %d/%d: %s", i+1, len(procedures), proc)
			procStartTime := time.Now()

			query = fmt.Sprintf("EXECUTE PROCEDURE %s", proc)
			log.Printf("DEBUG: Query para refazer: %s", query)

			// Executar diretamente sem Prepare (método que funciona)
			_, err := db.Exec(query)
			if err != nil {
				return fmt.Errorf("erro ao executar migração %s: %w", proc, err)
			}
			log.Printf("Procedure %s executada com sucesso em %v", proc, time.Since(procStartTime))
		}
		log.Printf("Migração de histórico de vendas FCERTA concluída em %v", time.Since(startTime))
		return nil

	} else {
		procName := procedureName(tipo, sistema)
		if procName == "" {
			return fmt.Errorf("tipo de migração inválido: %s para sistema %s", tipo, sistema)
		}

		log.Printf("Executando migração do tipo %s usando procedure: %s", tipo, procName)

		// Procedures com parâmetros (conversao)
		log.Printf("DEBUG: Executando procedure %s com parâmetro: %s", procName, conversao)

		// Tentar execução direta sem Prepare (pode resolver problemas com EXECUTE STATEMENT)
		if sistema == "FCERTA" && (tipo == "2" || tipo == "3" || tipo == "4" || tipo == "6") {
			// Para FCERTA, usar integer diretamente na query
			if convInt, err := strconv.Atoi(conversao); err == nil {
				query = fmt.Sprintf("EXECUTE PROCEDURE %s(%d)", procName, convInt)
				log.Printf("DEBUG: Query com integer: %s", query)
			} else {
				log.Printf("DEBUG: Erro ao converter conversao para integer: %v", err)
				return fmt.Errorf("erro ao converter conversao para integer: %w", err)
			}
		} else {
			// Para outros casos, usar string
			query = fmt.Sprintf("EXECUTE PROCEDURE %s('%s')", procName, conversao)
			log.Printf("DEBUG: Query com string: %s", query)
		}

		// Executar diretamente sem Prepare
		_, err := db.Exec(query)
		if err != nil {
			log.Printf("DEBUG: Erro ao executar query direta: %v", err)
			return fmt.Errorf("erro ao executar migração do tipo %s: %w", tipo, err)
		}
	}

	log.Printf("Migração do tipo %s concluída com sucesso em %v", tipo, time.Since(startTime))
	return nil
}

// MigrarDadosPRISMA5 executa migração de dados para PRISMA5 usando código Go direto
func MigrarDadosPRISMA5(db *sql.DB, tipo string, conversao string, extra *MigrarExtras) error {
	logger.Info("Iniciando migração PRISMA5 tipo %s com conversão %s", tipo, conversao)

	var cb *prisma5.MigrationCallbacks
	if extra != nil {
		cb = extra.PRISMACallbacks
	}

	switch tipo {
	case "1":
		_, err := prisma5.MigratePRISMA5Clientes(db, conversao, cb)
		return err
	case "2":
		return prisma5.MigratePRISMA5Medicos(db, conversao)
	case "3":
		return prisma5.MigratePRISMA5Fornecedores(db, conversao)
	case "4":
		return prisma5.MigratePRISMA5Produtos(db, conversao)
	case "5":
		return prisma5.MigratePRISMA5FormaFarmaceutica(db)
	case "6":
		return prisma5.MigratePRISMA5Lotes(db)
	case "7":
		return prisma5.MigratePRISMA5ProducaoInterna(db)
	case "8":
		return prisma5.MigratePRISMA5Refazer(db, conversao)
	default:
		return fmt.Errorf("tipo de migração inválido para PRISMA5: %s", tipo)
	}
}

// validarParametros valida os parâmetros de entrada
func validarParametros(tipo, conversao string, vVencido *string, sistema string) error {
	// Validar tipo
	if tipo == "" {
		return fmt.Errorf("tipo de migração não pode estar vazio")
	}

	// Validar se tipo é válido
	tiposValidos := []string{"1", "2", "3", "4", "5", "6", "7", "8"}
	tipoValido := false
	for _, t := range tiposValidos {
		if tipo == t {
			tipoValido = true
			break
		}
	}
	if !tipoValido {
		return fmt.Errorf("tipo de migração '%s' não é válido. Tipos válidos: %v", tipo, tiposValidos)
	}

	// Validar conversão
	if tipo != "5" && conversao == "" {
		return fmt.Errorf("parâmetro de conversão é obrigatório para o tipo %s", tipo)
	}

	if conversao != "" {
		// Validar se conversão é um número válido
		if _, err := strconv.Atoi(conversao); err != nil {
			return fmt.Errorf("parâmetro de conversão '%s' deve ser um número inteiro válido", conversao)
		}
	}

	// Validar sistema
	if sistema == "" {
		return fmt.Errorf("sistema não pode estar vazio")
	}

	sistemasValidos := []string{"FCERTA", "PRISMA5"}
	sistemaValido := false
	for _, s := range sistemasValidos {
		if sistema == s {
			sistemaValido = true
			break
		}
	}
	if !sistemaValido {
		return fmt.Errorf("sistema '%s' não é válido. Sistemas válidos: %v", sistema, sistemasValidos)
	}

	// Só exige vVencido para FCERTA, não para PRISMA5
	if tipo == "5" && sistema == "FCERTA" && vVencido == nil {
		return fmt.Errorf("parâmetro vVencido é obrigatório para o tipo 5 no sistema FCERTA")
	}

	// Validar vVencido se fornecido
	if vVencido != nil && *vVencido != "" {
		if _, err := strconv.Atoi(*vVencido); err != nil {
			return fmt.Errorf("parâmetro vVencido '%s' deve ser um número inteiro válido", *vVencido)
		}
	}

	return nil
}

func procedureName(tipo string, sistema string) string {
	switch sistema {
	case "FCERTA":
		switch tipo {
		case "1":
			return "EXTRACT_DATA_FC07000" // Clientes
		case "2":
			return "EXTRACT_DATA_FC04000" // Médicos
		case "3":
			return "EXTRACT_DATA_FC02000" // Fornecedores
		case "4":
			return "EXTRACT_DATA_FC03000" // Produtos
		case "5":
			return "EXTRACT_DATA_FC03140" // Lotes
		case "6":
			return "EXTRACT_DATA_FC05000" // Produção Interna
		case "7":
			return "EXTRACT_DATA_REFAZER_1" // Refazer - Histórico Vendas
		case "8":
			return "" // Forma Farmacêutica não existe no FCERTA
		default:
			return ""
		}
	case "PRISMA5":
		switch tipo {
		case "1":
			return "EXTRACT_DATA_CLIENTE"
		case "2":
			return "EXTRACT_DATA_MEDICO"
		case "3":
			return "EXTRACT_DATA_FORNECEDOR"
		case "4":
			return "EXTRACT_DATA_PRODUTO"
		case "5":
			return "EXTRACT_DATA_FORMAFARMACEUTICA"
		case "6":
			return "EXTRACT_DATA_LOTE"
		case "7":
			return "EXTRACT_DATA_PROD_INTERNA"
		case "8":
			return "EXTRACT_DATA_REFAZER"
		default:
			return ""
		}
	default:
		// Para outros sistemas, usar os nomes padrão do PRISMA5
		switch tipo {
		case "1":
			return "EXTRACT_DATA_CLIENTE"
		case "2":
			return "EXTRACT_DATA_MEDICO"
		case "3":
			return "EXTRACT_DATA_FORNECEDOR"
		case "4":
			return "EXTRACT_DATA_PRODUTO"
		case "5":
			return "EXTRACT_DATA_FORMAFARMACEUTICA"
		case "6":
			return "EXTRACT_DATA_LOTE"
		case "7":
			return "EXTRACT_DATA_PROD_INTERNA"
		case "8":
			return "EXTRACT_DATA_REFAZER"
		default:
			return ""
		}
	}
}
