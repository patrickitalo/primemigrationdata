package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/primesoftwaresi/prime-migration/pkg/config"
	"github.com/primesoftwaresi/prime-migration/pkg/logger"
)

// getProjectDir encontra o diretório raiz do projeto (onde está go.mod)
// Procura a partir do diretório do executável ou do diretório de trabalho
func getProjectDir() string {
	// Quando executado com go run, o executável está em AppData/go-build
	// Nesse caso, usar a variável de ambiente ou o diretório onde go.mod deveria estar
	
	// 1. Tentar variável de ambiente GOWORK ou GOENV
	if projectDir := os.Getenv("GO_PROJECT_DIR"); projectDir != "" {
		if _, err := os.Stat(filepath.Join(projectDir, "go.mod")); err == nil {
			return projectDir
		}
	}
	
	// 2. Tentar a partir do diretório de trabalho primeiro (mais comum com go run)
	wd, err := os.Getwd()
	if err == nil {
		current := wd
		for i := 0; i < 20; i++ { // Aumentar para 20 níveis
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

	// 3. Tentar a partir do executável
	exePath, err := os.Executable()
	if err == nil {
		current := filepath.Dir(exePath)
		// Se executável está em diretório temporário do Go, não confiar nele
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

	// 4. Tentar caminhos comuns de desenvolvimento
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

	// Se não encontrou, retornar string vazia
	return ""
}

// getExecutableDir retorna o diretório onde o executável está localizado
func getExecutableDir() string {
	exePath, err := os.Executable()
	if err != nil {
		// Se não conseguir obter o caminho do executável, usar diretório atual
		wd, _ := os.Getwd()
		return wd
	}
	return filepath.Dir(exePath)
}

// Lê arquivo com encoding UTF8
func readFileWithEncoding(filePath string) (string, error) {
	// Obter diretórios possíveis - as pastas sps-fcerta e sps-prisma5
	// podem estar no diretório do executável OU no diretório do projeto
	exeDir := getExecutableDir()
	projectDir := getProjectDir()
	wd, _ := os.Getwd()

	// Tentar caminhos possíveis em ordem de prioridade
	paths := []string{}

	// 1. Caminho relativo atual e explícito
	paths = append(paths, filePath)
	paths = append(paths, "./"+filePath)

	// 2. Diretório do projeto (prioridade máxima se encontrado)
	if projectDir != "" {
		paths = append(paths, filepath.Join(projectDir, filePath))
	}

	// 3. Se executável está em diretório temporário do Go (go run),
	// procurar o projeto a partir do diretório de trabalho atual
	if strings.Contains(exeDir, "AppData") && strings.Contains(exeDir, "go-build") {
		// Quando executado com go run, procurar go.mod a partir do diretório onde foi executado
		if projectDir == "" {
			// Se não encontrou go.mod, tentar subir a partir do wd
			current := wd
			for i := 0; i < 20; i++ {
				testPath := filepath.Join(current, filePath)
				if _, err := os.Stat(testPath); err == nil {
					paths = append(paths, testPath)
				}
				goModPath := filepath.Join(current, "go.mod")
				if _, err := os.Stat(goModPath); err == nil {
					paths = append(paths, filepath.Join(current, filePath))
					break
				}
				parent := filepath.Dir(current)
				if parent == current {
					break
				}
				current = parent
			}
		}
	}

	// 4. Diretório do executável
	paths = append(paths, filepath.Join(exeDir, filePath))

	// 5. Diretório de trabalho
	paths = append(paths, filepath.Join(wd, filePath))

	// Remover duplicatas mantendo ordem
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
			// Retorna o conteúdo como UTF8 diretamente
			return string(content), nil
		}
		lastErr = err
	}

	// Se nenhum caminho funcionou, retornar o último erro com informações úteis
	return "", fmt.Errorf("erro ao ler arquivo (tentados caminhos: %v, diretório projeto: %s, diretório executável: %s, diretório de trabalho: %s): %w", uniquePaths, projectDir, exeDir, wd, lastErr)
}

// Executa um arquivo SQL como um único statement (usado para procedures)
// fcertaProcedureExists verifica se uma procedure de usuário existe no catálogo Firebird.
func fcertaProcedureExists(db *sql.DB, name string) bool {
	var n int
	q := `SELECT COUNT(*) FROM RDB$PROCEDURES WHERE TRIM(RDB$PROCEDURE_NAME) = ? AND RDB$SYSTEM_FLAG = 0`
	if err := db.QueryRow(q, strings.TrimSpace(strings.ToUpper(name))).Scan(&n); err != nil {
		return false
	}
	return n > 0
}

func executeSQLFileAsSingleStatement(db *sql.DB, filePath string) error {
	content, err := readFileWithEncoding(filePath)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo %s: %w", filePath, err)
	}
	stmt := strings.TrimSpace(content)
	if stmt == "" {
		return nil
	}
	_, err = db.Exec(stmt)
	if err != nil {
		return fmt.Errorf("erro ao executar statement único: %w", err)
	}
	return nil
}

func SetupProcedures(db *sql.DB, sistema string, clienteCodigo string) error {
	state, _ := LoadMigrationState(clienteCodigo, sistema)
	logger.Info("Iniciando configuração das procedures para sistema %s, cliente %s", sistema, clienteCodigo)

	switch strings.ToUpper(sistema) {
	case "PRISMA5":
		// PRISMA5 não usa mais procedures SQL - tudo é feito em Go
		// O setup do banco é feito via SetupPRISMA5Database (tabelas, colunas, sequences)
		logger.Info("Sistema PRISMA5 não requer setup de procedures SQL (tudo é feito em Go)")
		return nil

	case "FCERTA":
		fmt.Printf("Iniciando criação das procedures do %s...\n", sistema)
		logger.Info("Configurando procedures para sistema FCERTA")
		// O estado fica no SQLite por cliente+sistema; o .FDB pode ser outro arquivo (restauração,
		// cópia nova) sem procedures — aí não podemos pular o setup só porque o flag está true.
		if state.ConfiguracaoInicialExecutada {
			if fcertaProcedureExists(db, "EXTRACT_DATA_FC07000") {
				fmt.Println("Configuração inicial já executada anteriormente. Pulando esta etapa.")
				return nil
			}
			fmt.Println("Aviso: o app marcou setup como feito, mas este Firebird não tem as procedures EXTRACT_* (banco novo ou outro .FDB). Recriando procedures...")
			logger.Warn("FCERTA: configuracao_inicial_executada no SQLite, mas EXTRACT_DATA_FC07000 ausente no Firebird; repetindo setup.")
			state.ConfiguracaoInicialExecutada = false
			state.ProceduresCriadas = nil
			state.InsertsExecutados = nil
			_ = SaveMigrationState(state)
		}
		// Executar a procedure UTIL_EXTRACT_FCERTA.sql como statement único
		if !contains(state.ProceduresCriadas, "UTIL_EXTRACT_FCERTA") {
			if err := executeSQLFileAsSingleStatement(db, "sps-fcerta/UTIL_EXTRACT_FCERTA.sql"); err != nil {
				if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "unsuccessful metadata update") {
					fmt.Println("Procedure UTIL_EXTRACT_FCERTA já existe ou já foi criada anteriormente.")
					log.Printf("Aviso: %v", err)
				} else {
					fmt.Printf("Erro ao criar procedure UTIL_EXTRACT_FCERTA: %v\n", err)
					log.Printf("Erro ao criar procedure UTIL_EXTRACT_FCERTA: %v", err)
					return err
				}
			}
			state.ProceduresCriadas = append(state.ProceduresCriadas, "UTIL_EXTRACT_FCERTA")
			_ = SaveMigrationState(state)
		}
		// Executar a procedure UTIL_EXTRACT_FCERTA
		fmt.Println("Executando procedure UTIL_EXTRACT_FCERTA...")
		if _, err := db.Exec("EXECUTE PROCEDURE UTIL_EXTRACT_FCERTA"); err != nil {
			if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "unsuccessful metadata update") {
				fmt.Println("Procedure UTIL_EXTRACT_FCERTA já foi executada anteriormente e os campos já existem.")
				log.Printf("Aviso: %v", err)
			} else {
				fmt.Printf("Erro ao executar procedure UTIL_EXTRACT_FCERTA: %v\n", err)
				log.Printf("Erro ao executar procedure UTIL_EXTRACT_FCERTA: %v", err)
				return err
			}
		}
		// Executar os demais scripts
		for _, sqlFile := range config.SequenciaSQLFCERTA {
			if sqlFile == "sps-fcerta/UTIL_EXTRACT_FCERTA.sql" {
				continue // já executado acima
			}
			name := extractName(sqlFile)
			if contains(state.ProceduresCriadas, name) || contains(state.InsertsExecutados, name) {
				fmt.Printf("%s já executado anteriormente. Pulando.\n", name)
				continue
			}
			fmt.Printf("Executando: %s...\n", sqlFile)
			if isProcedureFile(sqlFile) {
				fmt.Printf("  - Identificado como procedure\n")
				if err := executeSQLFileAsSingleStatement(db, sqlFile); err != nil {
					// Verificar se é erro "Malformed string" - isso indica que a procedure não foi criada corretamente
					if strings.Contains(err.Error(), "Malformed string") {
						fmt.Printf("ERRO: Procedure %s não foi criada corretamente (Malformed string). Tentando fazer DROP e recriar...\n", name)
						log.Printf("ERRO: Procedure %s com erro Malformed string: %v", name, err)
						
						// Tentar fazer DROP da procedure
						procedureName := strings.ToUpper(name)
						dropQuery := fmt.Sprintf("DROP PROCEDURE %s", procedureName)
						_, dropErr := db.Exec(dropQuery)
						if dropErr != nil {
							fmt.Printf("Aviso ao fazer DROP: %v\n", dropErr)
						} else {
							fmt.Printf("Procedure %s removida. Tentando recriar...\n", procedureName)
						}
						
						// Tentar recriar
						if retryErr := executeSQLFileAsSingleStatement(db, sqlFile); retryErr != nil {
							fmt.Printf("ERRO: Não foi possível recriar procedure %s: %v\n", name, retryErr)
							log.Printf("ERRO ao recriar procedure %s: %v", name, retryErr)
							return fmt.Errorf("erro ao criar procedure %s: %w", sqlFile, retryErr)
						}
						fmt.Printf("  - Procedure %s recriada com sucesso após erro Malformed string\n", name)
						state.ProceduresCriadas = append(state.ProceduresCriadas, name)
						_ = SaveMigrationState(state)
					} else if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "unsuccessful metadata update") {
						// Verificar se a procedure realmente existe no banco antes de ignorar o erro
						procedureName := strings.ToUpper(name)
						var count int
						checkQuery := `SELECT COUNT(*) FROM RDB$PROCEDURES WHERE UPPER(RDB$PROCEDURE_NAME) = ?`
						checkErr := db.QueryRow(checkQuery, procedureName).Scan(&count)
						
						if checkErr == nil && count > 0 {
							fmt.Printf("Procedure %s já existe no banco (erro ignorado): %s\n", name, sqlFile)
							log.Printf("Aviso: %v", err)
							state.ProceduresCriadas = append(state.ProceduresCriadas, name)
							_ = SaveMigrationState(state)
						} else {
							// Procedure não existe, mas deu erro - não ignorar
							fmt.Printf("ERRO: Procedure %s não existe no banco, mas houve erro ao criar: %v\n", name, err)
							log.Printf("ERRO: Procedure %s não existe no banco: %v", name, err)
							return fmt.Errorf("erro ao criar procedure %s: %w", sqlFile, err)
						}
					} else {
						fmt.Printf("Erro ao executar %s: %v\n", sqlFile, err)
						log.Printf("Erro ao executar %s: %v", sqlFile, err)
						return err
					}
				} else {
					fmt.Printf("  - Procedure normal criada com sucesso: %s\n", sqlFile)
					state.ProceduresCriadas = append(state.ProceduresCriadas, name)
					_ = SaveMigrationState(state)
				}
			} else {
				fmt.Printf("  - Identificado como script INSERT\n")
				if err := executeSQLFile(db, sqlFile); err != nil {
					fmt.Printf("Erro ao executar %s: %v\n", sqlFile, err)
					log.Printf("Erro ao executar %s: %v", sqlFile, err)
					return err
				} else {
					fmt.Printf("  - Script INSERT executado com sucesso: %s\n", sqlFile)
					state.InsertsExecutados = append(state.InsertsExecutados, name)
					_ = SaveMigrationState(state)
				}
			}
		}
		fmt.Println("Procedures e scripts criados com sucesso!")
		state.ConfiguracaoInicialExecutada = true
		_ = SaveMigrationState(state)
		return nil

	default:
		logger.Warn("Sistema %s não possui configuração de procedures definida.", sistema)
		return nil
	}
}

// Verifica se o arquivo é uma procedure baseado no conteúdo
func isProcedureFile(filePath string) bool {
	content, err := readFileWithEncoding(filePath)
	if err != nil {
		return false
	}

	// Verificar se contém "create or alter procedure" ou "create or alter function" no início
	contentStr := strings.ToLower(content)
	return strings.Contains(contentStr, "create or alter procedure") || strings.Contains(contentStr, "create or alter function")
}

// Verifica se é uma procedure de tabela (que tem SUSPEND)
// IMPORTANTE: Procedures de extração (EXTRACT_DATA_*) não devem ser executadas durante o setup,
// apenas criadas. Elas serão executadas durante a migração.
func isTableProcedure(filePath string) bool {
	content, err := readFileWithEncoding(filePath)
	if err != nil {
		return false
	}

	contentStr := strings.ToLower(content)
	fileName := strings.ToLower(filepath.Base(filePath))

	// Procedures de extração (EXTRACT_DATA_*) não são table procedures para execução no setup
	// Elas são apenas criadas, não executadas
	if strings.Contains(fileName, "extract_data_") {
		return false
	}

	// Verificar se é procedure (não função) e tem SUSPEND
	return strings.Contains(contentStr, "create or alter procedure") &&
		!strings.Contains(contentStr, "create or alter function") &&
		strings.Contains(contentStr, "suspend")
}

func executeSQLFile(db *sql.DB, filePath string) error {
	content, err := readFileWithEncoding(filePath)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo %s: %w", filePath, err)
	}
	statements := strings.Split(string(content), ";")
	batchSize := 1000
	batch := make([]string, 0, batchSize)
	count := 0
	for i, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || stmt == "\n" || stmt == "\r" {
			continue
		}
		batch = append(batch, stmt)
		if len(batch) == batchSize || i == len(statements)-1 {
			for _, s := range batch {
				if s == "" || s == "\n" || s == "\r" {
					continue
				}
				_, err := db.Exec(s)
				if err != nil {
					fmt.Printf("Erro ao executar statement: %v\nStatement problemático: [%s]\n", err, s)
					log.Printf("Erro ao executar statement: %v\nStatement problemático: [%s]", err, s)
					return fmt.Errorf("erro ao executar statement: %w", err)
				}
				count++
			}
			fmt.Printf("%d statements executados do arquivo %s...\n", count, filePath)
			batch = batch[:0]
		}
	}
	return nil
}

// Executa uma procedure de tabela (mantém o SUSPEND)
func executeTableProcedure(db *sql.DB, filePath string) error {
	content, err := readFileWithEncoding(filePath)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo %s: %w", filePath, err)
	}

	// Manter o SUSPEND para procedures de tabela
	stmt := strings.TrimSpace(content)
	if stmt == "" {
		return nil
	}
	logger.Info("Executando procedure de tabela: %s", filePath)
	_, err = db.Exec(stmt)
	if err != nil {
		return fmt.Errorf("erro ao executar procedure de tabela: %w", err)
	}
	return nil
}

func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

func extractName(filePath string) string {
	_, fileName := filepath.Split(filePath)
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}
