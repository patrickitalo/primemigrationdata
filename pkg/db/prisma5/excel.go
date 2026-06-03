package prisma5

import (
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/primesoftwaresi/prime-migration/pkg/logger"
	"github.com/xuri/excelize/v2"
)

// ImportarPlanilhaGruposPRISMA5 importa dados da planilha Excel para a tabela GRUPO_PHARMACIE
func ImportarPlanilhaGruposPRISMA5(db *sql.DB, caminhoPlanilha string) error {
	logger.Info("Iniciando importação da planilha de grupos PRISMA5: %s", caminhoPlanilha)

	// Verificar se o arquivo existe
	if _, err := os.Stat(caminhoPlanilha); os.IsNotExist(err) {
		return fmt.Errorf("arquivo não encontrado: %s", caminhoPlanilha)
	}

	// Abrir a planilha Excel
	f, err := excelize.OpenFile(caminhoPlanilha)
	if err != nil {
		return fmt.Errorf("erro ao abrir planilha: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			logger.Error("Erro ao fechar planilha: %v", err)
		}
	}()

	// Obter a primeira planilha
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return fmt.Errorf("planilha vazia ou sem abas")
	}

	// Obter todas as linhas da planilha
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("erro ao ler linhas da planilha: %w", err)
	}

	if len(rows) < 2 {
		return fmt.Errorf("planilha deve ter pelo menos uma linha de cabeçalho e uma linha de dados")
	}

	// Pular a primeira linha (cabeçalho) - linha 1 é o cabeçalho
	// Começar da linha 2 (índice 1)
	registrosInseridos := 0
	registrosComErro := 0

	logger.Info("Processando %d linhas da planilha (excluindo cabeçalho)", len(rows)-1)

	for i := 1; i < len(rows); i++ {
		row := rows[i]

		// Verificar se a linha tem pelo menos 3 colunas
		if len(row) < 3 {
			logger.Warn("Linha %d ignorada: possui menos de 3 colunas", i+1)
			continue
		}

		codigoGrupo := strings.TrimSpace(row[0])
		descricaoGrupo := strings.TrimSpace(row[1])
		movimentoEstoque := strings.TrimSpace(row[2])

		// Validar campos obrigatórios
		if codigoGrupo == "" {
			logger.Warn("Linha %d ignorada: CODIGOGRUPO está vazio", i+1)
			registrosComErro++
			continue
		}

		// Extrair apenas números do campo MOVIMENTOESTOQUE
		movimentoEstoqueNumeros := extrairApenasNumeros(movimentoEstoque)

		// Inserir na tabela GRUPO_PHARMACIE
		query := "INSERT INTO GRUPO_PHARMACIE (CODIGOGRUPO, DESCRICAOGRUPO, MOVIMENTOESTOQUE) VALUES (?, ?, ?)"
		_, err := db.Exec(query, codigoGrupo, descricaoGrupo, movimentoEstoqueNumeros)
		if err != nil {
			// Verificar se é erro de duplicidade (registro já existe)
			if strings.Contains(err.Error(), "violation") || strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique constraint") {
				logger.Warn("Linha %d ignorada: registro com CODIGOGRUPO '%s' já existe", i+1, codigoGrupo)
				registrosComErro++
				continue
			}
			logger.Error("Erro ao inserir linha %d (CODIGOGRUPO: %s): %v", i+1, codigoGrupo, err)
			registrosComErro++
			continue
		}

		registrosInseridos++

		// Log de progresso a cada 100 registros
		if registrosInseridos%100 == 0 {
			fmt.Printf("%d registros importados...\n", registrosInseridos)
			logger.Info("%d registros importados até o momento", registrosInseridos)
		}
	}

	logger.Info("Importação concluída: %d registros inseridos, %d registros com erro ou ignorados", registrosInseridos, registrosComErro)
	fmt.Printf("✅ Importação concluída!\n")
	fmt.Printf("   Registros inseridos: %d\n", registrosInseridos)
	if registrosComErro > 0 {
		fmt.Printf("   Registros ignorados/erros: %d\n", registrosComErro)
	}

	return nil
}

// extrairApenasNumeros extrai apenas os dígitos numéricos de uma string
func extrairApenasNumeros(s string) string {
	// Remover tudo que não for número
	re := regexp.MustCompile(`[^\d]`)
	return re.ReplaceAllString(s, "")
}
