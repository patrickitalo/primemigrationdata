package prisma5

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/primesoftwaresi/prime-migration/pkg/logger"
)

// SetupPRISMA5Database prepara o banco de dados PRISMA5 criando tabelas, sequences e colunas necessárias
func SetupPRISMA5Database(db *sql.DB) error {
	logger.Info("Configurando banco de dados PRISMA5 (tabelas, sequences e colunas)")

	// Lista de comandos SQL para criação
	setupCommands := []struct {
		name string
		sql  string
	}{
		{"CIDADEESTADO", "CREATE TABLE CIDADEESTADO (CODIGO INTEGER NOT NULL, CODIGO_IBGE INTEGER, NOMECIDADE VARCHAR(60), UF VARCHAR(2), CODIGO_PAIS INTEGER, CADASTRO_LJ SMALLINT, CADASTRO_DT TIMESTAMP, CADASTRO_CF INTEGER, ALTERACAO_LJ SMALLINT, ALTERACAO_DT TIMESTAMP, ALTERACAO_CF INTEGER, CODIGO_MUNICIPIO INTEGER, REGISTRO_CEP SMALLINT)"},
		{"GRUPO_PHARMACIE", "CREATE TABLE GRUPO_PHARMACIE (CODIGOGRUPO INTEGER, DESCRICAOGRUPO VARCHAR(150), MOVIMENTOESTOQUE SMALLINT)"},
		{"CONEXAO", "CREATE TABLE CONEXAO (ALIAS VARCHAR(60), IPSERVER VARCHAR(60), PORTA VARCHAR(4))"},

		// Sequences
		{"GEN_REFAZER_A1", "CREATE SEQUENCE GEN_REFAZER_A1"},
		{"GEN_REFAZER_A2", "CREATE SEQUENCE GEN_REFAZER_A2"},
		{"GEN_REFAZER_A3", "CREATE SEQUENCE GEN_REFAZER_A3"},
		{"GEN_CADASTRO_ENDERECO", "CREATE SEQUENCE GEN_CADASTRO_ENDERECO"},
		{"GEN_CADASTRO_TELEFONE", "CREATE SEQUENCE GEN_CADASTRO_TELEFONE"},
		{"GEN_GRUPO0", "CREATE SEQUENCE GEN_GRUPO0"},
		{"GEN_GRUPO1", "CREATE SEQUENCE GEN_GRUPO1"},
		{"GEN_GRUPO2", "CREATE SEQUENCE GEN_GRUPO2"},
		{"GEN_GRUPO3", "CREATE SEQUENCE GEN_GRUPO3"},
		{"GEN_GRUPO4", "CREATE SEQUENCE GEN_GRUPO4"},
		{"GEN_ESTOQUE_GERAL0", "CREATE SEQUENCE GEN_ESTOQUE_GERAL0"},
		{"GEN_ESTOQUE_GERAL1", "CREATE SEQUENCE GEN_ESTOQUE_GERAL1"},
		{"GEN_ESTOQUE_GERAL2", "CREATE SEQUENCE GEN_ESTOQUE_GERAL2"},
		{"GEN_ESTOQUE_GERAL3", "CREATE SEQUENCE GEN_ESTOQUE_GERAL3"},
		{"GEN_ESTOQUE_GERAL4", "CREATE SEQUENCE GEN_ESTOQUE_GERAL4"},
		{"GEN_CLIENTE", "CREATE SEQUENCE GEN_CLIENTE"},
		{"GEN_ESTOQUE_LOTE", "CREATE SEQUENCE GEN_ESTOQUE_LOTE"},
		{"GEN_ESTOQUE_LOTE_LA", "CREATE SEQUENCE GEN_ESTOQUE_LOTE_LA"},
		{"GEN_FORMAFARMACEUTICA", "CREATE SEQUENCE GEN_FORMAFARMACEUTICA"},
		{"GEN_ESTOQUEF", "CREATE SEQUENCE GEN_ESTOQUEF"},
		{"GEN_ESTOQUEF_DETALHE", "CREATE SEQUENCE GEN_ESTOQUEF_DETALHE"},
		{"GEN_ESTOQUE_MINMAX", "CREATE SEQUENCE GEN_ESTOQUE_MINMAX"},
		{"GEN_MEDICO", "CREATE SEQUENCE GEN_MEDICO"},

		// Alter sequence
		{"GEN_CADASTRO_ENDERECO_RESTART", "ALTER SEQUENCE GEN_CADASTRO_ENDERECO RESTART WITH 2"},

		// Colunas CODIGO_PS e relacionadas
		{"CIDADE.CODIGO_PS", "ALTER TABLE CIDADE ADD CODIGO_PS INTEGER"},
		{"GRUPO.CODIGO_PS", "ALTER TABLE GRUPO ADD CODIGO_PS INTEGER"},
		{"PRODUTO.CODIGO_PS", "ALTER TABLE PRODUTO ADD CODIGO_PS INTEGER"},
		{"PRODUTO.CONVERSAO", "ALTER TABLE PRODUTO ADD CONVERSAO SMALLINT"},
		{"PRODUTO.NOVO", "ALTER TABLE PRODUTO ADD NOVO SMALLINT"},
		{"SINONIMO.CONVERSAO", "ALTER TABLE SINONIMO ADD CONVERSAO SMALLINT"},
		{"SINONIMO.CODIGO_PRODUTO_PS", "ALTER TABLE SINONIMO ADD CODIGO_PRODUTO_PS INTEGER"},
		{"SINONIMO.CODIGO_SINONIMO_PS", "ALTER TABLE SINONIMO ADD CODIGO_SINONIMO_PS INTEGER"},
		{"SINONIMO.NOVO", "ALTER TABLE SINONIMO ADD NOVO SMALLINT"},
		{"FORMULAPADRAO.CODIGO_PRODUTO_PS", "ALTER TABLE FORMULAPADRAO ADD CODIGO_PRODUTO_PS INTEGER"},
		{"FORMULAPADRAO.CODIGO_ESTOQUEF", "ALTER TABLE FORMULAPADRAO ADD CODIGO_ESTOQUEF INTEGER"},
		{"FORMULAPADRAO.OPCAO_ATENDIMENTO_PS", "ALTER TABLE FORMULAPADRAO ADD OPCAO_ATENDIMENTO_PS SMALLINT"},
		{"ITEMFORMULAPADRAO.CODIGO_ESTOQUEF", "ALTER TABLE ITEMFORMULAPADRAO ADD CODIGO_ESTOQUEF INTEGER"},
		{"ITEMFORMULAVENDA.CODIGO_ESTOQUEF", "ALTER TABLE ITEMFORMULAVENDA ADD CODIGO_ESTOQUEF INTEGER"},
		{"ITEMFORMULAVENDA.CODIGO_PS_A1", "ALTER TABLE ITEMFORMULAVENDA ADD CODIGO_PS_A1 INTEGER"},
		{"ITEMFORMULAVENDA.CODIGO_PS", "ALTER TABLE ITEMFORMULAVENDA ADD CODIGO_PS INTEGER"},
		{"CLIENTE.CODIGO_PS", "ALTER TABLE CLIENTE ADD CODIGO_PS INTEGER"},
		{"CLIENTE.NOVO", "ALTER TABLE CLIENTE ADD NOVO SMALLINT"},
		{"CLIENTE.CONVERSAO", "ALTER TABLE CLIENTE ADD CONVERSAO SMALLINT"},
		{"CLIENTE_ENDERECO_ENTREGA.CODIGO_PS", "ALTER TABLE CLIENTE_ENDERECO_ENTREGA ADD CODIGO_PS INTEGER"},
		{"LOTE.CODIGO_PS_LOTE", "ALTER TABLE LOTE ADD CODIGO_PS_LOTE INTEGER"},
		{"LOTE.CODIGO_PS_LOTE_LA", "ALTER TABLE LOTE ADD CODIGO_PS_LOTE_LA INTEGER"},
		{"VENDA.CODIGO_PS", "ALTER TABLE VENDA ADD CODIGO_PS INTEGER"},
		{"VENDA.CONVERSAO", "ALTER TABLE VENDA ADD CONVERSAO INTEGER"},
		{"VENDA.CODIGO_MEDICO", "ALTER TABLE VENDA ADD CODIGO_MEDICO INTEGER"},
		{"FORMULAVENDA.CONVERSAO", "ALTER TABLE FORMULAVENDA ADD CONVERSAO INTEGER"},
		{"FORMULAVENDA.CODIGO_PS", "ALTER TABLE FORMULAVENDA ADD CODIGO_PS INTEGER"},
		{"VISITADOR.CODIGO_PS", "ALTER TABLE VISITADOR ADD CODIGO_PS INTEGER"},
		{"FORMULAVENDA.CODIGO_PS_A1", "ALTER TABLE FORMULAVENDA ADD CODIGO_PS_A1 INTEGER"},
		{"ITEMFORMULAVENDA.CONVERSAO", "ALTER TABLE ITEMFORMULAVENDA ADD CONVERSAO INTEGER"},
		{"MEDICO.CODIGO_PS", "ALTER TABLE MEDICO ADD CODIGO_PS INTEGER"},
		{"MEDICO.CONVERSAO", "ALTER TABLE MEDICO ADD CONVERSAO SMALLINT"},
		{"MEDICO.ENVIADO", "ALTER TABLE MEDICO ADD ENVIADO SMALLINT"},
		{"MEDICO.NOVO", "ALTER TABLE MEDICO ADD NOVO SMALLINT"},
		{"FORNECEDOR.CODIGO_PS", "ALTER TABLE FORNECEDOR ADD CODIGO_PS INTEGER"},
		{"FORNECEDOR.CONVERSAO", "ALTER TABLE FORNECEDOR ADD CONVERSAO SMALLINT"},
		{"FORNECEDOR.ENVIADO", "ALTER TABLE FORNECEDOR ADD ENVIADO SMALLINT"},
		{"FORNECEDOR.NOVO", "ALTER TABLE FORNECEDOR ADD NOVO SMALLINT"},
		{"VISITADOR.CODIGO_PS", "ALTER TABLE VISITADOR ADD CODIGO_PS INTEGER"},
		{"FORMAFARMACEUTICA.CODIGO_PS", "ALTER TABLE FORMAFARMACEUTICA ADD CODIGO_PS INTEGER"},
		{"GRUPO_PHARMACIE.CODIGO_GRUPO_PHARMACIE", "ALTER TABLE GRUPO_PHARMACIE ADD CODIGO_GRUPO_PHARMACIE INTEGER"},

		// Índices
		{"VENDA_IDX2", "CREATE DESCENDING INDEX VENDA_IDX2 ON VENDA (CODIGO_PS)"},
		{"VENDA_IDX1", "CREATE INDEX VENDA_IDX1 ON VENDA (CODIGO_PS)"},

		// Primary key
		{"PK_CIDADEESTADO", "ALTER TABLE CIDADEESTADO ADD CONSTRAINT PK_CIDADEESTADO PRIMARY KEY (CODIGO)"},
	}

	for _, cmd := range setupCommands {
		logger.Info("Executando setup: %s", cmd.name)
		_, err := db.Exec(cmd.sql)
		if err != nil {
			// Ignorar erros de "already exists" ou "unsuccessful metadata update"
			errStr := strings.ToLower(err.Error())
			if strings.Contains(errStr, "already exists") ||
				strings.Contains(errStr, "unsuccessful metadata update") ||
				strings.Contains(errStr, "duplicate") {
				logger.Info("  %s já existe (ignorando)", cmd.name)
				continue
			} else {
				logger.Error("Erro ao executar %s: %v", cmd.name, err)
				return fmt.Errorf("erro ao executar %s: %w", cmd.name, err)
			}
		} else {
			logger.Info("  %s criado com sucesso", cmd.name)
		}
	}

	logger.Info("Setup do banco PRISMA5 concluído")
	
	// Inserir dados de CIDADEESTADO
	logger.Info("Inserindo dados de CIDADEESTADO...")
	if err := insertCidadeEstado(db); err != nil {
		logger.Warn("Aviso ao inserir dados de CIDADEESTADO: %v", err)
		// Não falhar o setup por causa disso, apenas avisar
	}
	
	return nil
}

// insertCidadeEstado insere os dados de CIDADEESTADO do arquivo SQL
func insertCidadeEstado(db *sql.DB) error {
	logger.Info("Lendo arquivo CIDADEESTADO.sql...")
	
	// Verificar se a tabela já tem dados
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM CIDADEESTADO").Scan(&count)
	if err != nil {
		logger.Warn("Aviso ao verificar dados existentes em CIDADEESTADO: %v", err)
	}
	
	if count > 0 {
		logger.Info("CIDADEESTADO já possui %d registros (pulando inserção)", count)
		return nil
	}
	
	// Ler o arquivo SQL
	content, err := readFileWithEncoding("sps-prisma5/CIDADEESTADO.sql")
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo CIDADEESTADO.sql: %w", err)
	}
	
	logger.Info("Executando INSERTs de CIDADEESTADO...")
	
	// Dividir o conteúdo em statements (separados por ;)
	statements := strings.Split(content, ";")
	batchSize := 1000 // Executar em lotes de 1000 para melhor performance
	totalInserted := 0
	insertCount := 0
	
	// Iniciar transação para melhor performance
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação para CIDADEESTADO: %w", err)
	}
	defer tx.Rollback()
	
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || stmt == "\n" || stmt == "\r" {
			continue
		}
		
		// Apenas processar INSERTs
		if !strings.HasPrefix(strings.ToUpper(stmt), "INSERT INTO CIDADEESTADO") {
			continue
		}
		
		insertCount++
		
		// Executar INSERT
		_, err := tx.Exec(stmt)
		if err != nil {
			// Ignorar erros de duplicate/unique (pode ter sido inserido parcialmente antes)
			errStr := strings.ToLower(err.Error())
			if strings.Contains(errStr, "unique") || strings.Contains(errStr, "duplicate") || 
			   strings.Contains(errStr, "violation of PRIMARY") {
				logger.Info("  Registro já existe (ignorando)")
				continue
			}
			logger.Warn("Erro ao executar INSERT de CIDADEESTADO: %v", err)
			// Rollback e tentar sem transação
			tx.Rollback()
			// Tentar inserir um por um sem transação
			return insertCidadeEstadoOneByOne(db, statements)
		}
		
		totalInserted++
		
		// Commit a cada batchSize registros
		if insertCount%batchSize == 0 {
			if err := tx.Commit(); err != nil {
				logger.Warn("Erro ao commitar transação (tentando novamente): %v", err)
				tx.Rollback()
				// Tentar inserir um por um sem transação
				return insertCidadeEstadoOneByOne(db, statements)
			}
			
			// Iniciar nova transação
			tx, err = db.Begin()
			if err != nil {
				return fmt.Errorf("erro ao reiniciar transação: %w", err)
			}
			logger.Info("  %d registros de CIDADEESTADO inseridos... (total: %d)", batchSize, totalInserted)
		}
	}
	
	// Commit final
	if err := tx.Commit(); err != nil {
		logger.Warn("Erro ao commitar transação final: %v", err)
		return err
	}
	
	logger.Info("✅ %d registros de CIDADEESTADO inseridos com sucesso", totalInserted)
	return nil
}

// insertCidadeEstadoOneByOne insere os dados um por um (fallback quando transação falha)
func insertCidadeEstadoOneByOne(db *sql.DB, statements []string) error {
	logger.Info("Inserindo CIDADEESTADO um por um (modo fallback)...")
	totalInserted := 0
	
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || stmt == "\n" || stmt == "\r" {
			continue
		}
		
		// Apenas processar INSERTs
		if !strings.HasPrefix(strings.ToUpper(stmt), "INSERT INTO CIDADEESTADO") {
			continue
		}
		
		_, err := db.Exec(stmt)
		if err != nil {
			// Ignorar erros de duplicate/unique
			errStr := strings.ToLower(err.Error())
			if strings.Contains(errStr, "unique") || strings.Contains(errStr, "duplicate") || 
			   strings.Contains(errStr, "violation of PRIMARY") {
				continue
			}
			logger.Warn("Erro ao executar INSERT de CIDADEESTADO: %v", err)
			continue
		}
		totalInserted++
		
		if totalInserted%500 == 0 {
			logger.Info("  %d registros de CIDADEESTADO inseridos... (total: %d)", 500, totalInserted)
		}
	}
	
	logger.Info("✅ %d registros de CIDADEESTADO inseridos com sucesso", totalInserted)
	return nil
}

// readFileWithEncoding lê um arquivo com encoding UTF8, procurando em vários caminhos possíveis
func readFileWithEncoding(filePath string) (string, error) {
	// Obter diretórios possíveis
	exeDir := getExecutableDir()
	projectDir := getProjectDir()
	wd, _ := os.Getwd()

	// Tentar caminhos possíveis em ordem de prioridade
	paths := []string{
		filePath,
		"./" + filePath,
	}

	if projectDir != "" {
		paths = append(paths, filepath.Join(projectDir, filePath))
	}

	if exeDir != "" {
		paths = append(paths, filepath.Join(exeDir, filePath))
	}

	if wd != "" {
		paths = append(paths, filepath.Join(wd, filePath))
	}

	// Remover duplicatas
	seen := make(map[string]bool)
	uniquePaths := []string{}
	for _, path := range paths {
		if path == "" {
			continue
		}
		normalizedPath := filepath.Clean(path)
		if !seen[normalizedPath] {
			seen[normalizedPath] = true
			uniquePaths = append(uniquePaths, normalizedPath)
		}
	}

	var lastErr error
	for _, path := range uniquePaths {
		content, err := os.ReadFile(path)
		if err == nil {
			return string(content), nil
		}
		lastErr = err
	}

	return "", fmt.Errorf("erro ao ler arquivo (tentados caminhos: %v): %w", uniquePaths, lastErr)
}

// getExecutableDir retorna o diretório do executável
func getExecutableDir() string {
	exePath, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Dir(exePath)
}

// getProjectDir encontra o diretório raiz do projeto (onde está go.mod).
// Mesma lógica de pkg/db/setup.go: se o processo roda com cwd na pasta do .fdb,
// a subida a partir do wd não acha go.mod; aí entram GO_PROJECT_DIR, exe e caminhos comuns.
func getProjectDir() string {
	if projectDir := os.Getenv("GO_PROJECT_DIR"); projectDir != "" {
		if _, err := os.Stat(filepath.Join(projectDir, "go.mod")); err == nil {
			return projectDir
		}
	}

	wd, err := os.Getwd()
	if err == nil {
		current := wd
		for i := 0; i < 20; i++ {
			goModPath := filepath.Join(current, "go.mod")
			if _, err := os.Stat(goModPath); err == nil {
				return current
			}
			parent := filepath.Dir(current)
			if parent == current {
				break
			}
			current = parent
		}
	}

	exePath, err := os.Executable()
	if err == nil {
		current := filepath.Dir(exePath)
		if !strings.Contains(current, "AppData") || !strings.Contains(current, "go-build") {
			for i := 0; i < 20; i++ {
				goModPath := filepath.Join(current, "go.mod")
				if _, err := os.Stat(goModPath); err == nil {
					return current
				}
				parent := filepath.Dir(current)
				if parent == current {
					break
				}
				current = parent
			}
		}
	}

	homeDir, _ := os.UserHomeDir()
	if homeDir != "" {
		commonPaths := []string{
			filepath.Join(homeDir, "Projetos", "primesoftware", "prime-migration-v3"),
			filepath.Join(homeDir, "Documents", "prime-migration-v3"),
			"C:\\Projetos\\primesoftware\\prime-migration-v3",
		}
		for _, path := range commonPaths {
			if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
				return path
			}
		}
	}

	return ""
}

// Nota: Funções auxiliares estão em prisma5_helpers.go
// Funções de migração estão em arquivos separados:
// - prisma5_migrate_clientes.go (MigratePRISMA5Clientes)
// - prisma5_migrate_medicos.go (MigratePRISMA5Medicos)
// - prisma5_migrate_fornecedores.go (MigratePRISMA5Fornecedores)
// - prisma5_migrate_produtos.go (MigratePRISMA5Produtos)
// - prisma5_migrate_forma_farmaceutica.go (MigratePRISMA5FormaFarmaceutica)
// - prisma5_migrate_lotes.go (MigratePRISMA5Lotes)
// - prisma5_migrate_producao_interna.go (MigratePRISMA5ProducaoInterna)
// - prisma5_migrate_refazer.go (MigratePRISMA5Refazer)
// Importação de planilhas está em excel.go (ImportarPlanilhaGruposPRISMA5)
