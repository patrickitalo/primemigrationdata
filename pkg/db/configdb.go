package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/primesoftwaresi/prime-migration/pkg/config"
	_ "modernc.org/sqlite"
)

// Estrutura para configuração do banco de dados
type DbConfig struct {
	Path      string
	User      string
	Password  string
	Host      string
	Port      string
	Conversao string
	LogLevel  string
	LogFile   string
}

// Estrutura para persistir configurações do cliente no SQLite
type ClientConfig struct {
	ID                int    `json:"id"`
	CodigoCliente     string `json:"codigo_cliente"`
	SistemaOrigem     string `json:"sistema_origem"`
	DbPath            string `json:"db_path"`
	DbUser            string `json:"db_user"`
	DbPassword        string `json:"db_password"`
	DbHost            string `json:"db_host"`
	DbPort            string `json:"db_port"`
	AliasPharmacie    string `json:"alias_pharmacie"`
	IpServerPharmacie string `json:"ipserver_pharmacie"`
	PortaPharmacie    string `json:"porta_pharmacie"`
	Conversao         string `json:"conversao"`
	UltimaAtualizacao string `json:"ultima_atualizacao"`
}

// Estrutura para controle de estado da migração no SQLite
type MigrationState struct {
	ID                           int      `json:"id"`
	CodigoCliente                string   `json:"codigo_cliente"`
	SistemaOrigem                string   `json:"sistema_origem"`
	ConfiguracaoInicialExecutada bool     `json:"configuracao_inicial_executada"`
	ProceduresCriadas            []string `json:"procedures_criadas"`
	InsertsExecutados            []string `json:"inserts_executados"`
	UltimaAtualizacao            string   `json:"ultima_atualizacao"`
}

var (
	configDB   *sql.DB
	logDBMutex sync.Mutex // Mutex para serializar acesso ao SQLite em logs
)

// configSQLiteDSN retorna o DSN do SQLite (config local). PRIME_CONFIG_DB_PATH = caminho do arquivo
// ou DSN completo (se já contiver '?').
func configSQLiteDSN() string {
	const suffix = "?_journal_mode=WAL&_busy_timeout=5000"
	p := strings.TrimSpace(os.Getenv("PRIME_CONFIG_DB_PATH"))
	if p == "" {
		return "./config.db" + suffix
	}
	if strings.Contains(p, "?") {
		return p
	}
	return p + suffix
}

// Inicializa o banco de dados SQLite para configurações
func InitConfigDB() error {
	var err error
	configDB, err = sql.Open("sqlite", configSQLiteDSN())
	if err != nil {
		return fmt.Errorf("erro ao abrir banco de configurações: %w", err)
	}
	
	// Configurar modo WAL e busy timeout
	_, err = configDB.Exec("PRAGMA journal_mode=WAL; PRAGMA busy_timeout=5000;")
	if err != nil {
		return fmt.Errorf("erro ao configurar SQLite: %w", err)
	}
	
	// Configurar connection pool
	configDB.SetMaxOpenConns(1) // SQLite funciona melhor com 1 conexão
	configDB.SetMaxIdleConns(1)
	configDB.SetConnMaxLifetime(0) // Sem expiração

	// Criar tabela de configurações se não existir
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS client_configs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		codigo_cliente TEXT NOT NULL,
		sistema_origem TEXT NOT NULL,
		db_path TEXT NOT NULL,
		db_user TEXT NOT NULL,
		db_password TEXT NOT NULL,
		db_host TEXT NOT NULL,
		db_port TEXT NOT NULL,
		alias_pharmacie TEXT,
		ipserver_pharmacie TEXT,
		porta_pharmacie TEXT,
		conversao TEXT,
		ultima_atualizacao TEXT NOT NULL,
		UNIQUE(codigo_cliente, sistema_origem)
	);`

	_, err = configDB.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("erro ao criar tabela de configurações: %w", err)
	}

	// Criar tabela de estado da migração se não existir
	createMigrationStateTableSQL := `
	CREATE TABLE IF NOT EXISTS migration_states (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		codigo_cliente TEXT NOT NULL,
		sistema_origem TEXT NOT NULL,
		configuracao_inicial_executada BOOLEAN DEFAULT 0,
		procedures_criadas TEXT DEFAULT '[]',
		inserts_executados TEXT DEFAULT '[]',
		ultima_atualizacao TEXT NOT NULL,
		UNIQUE(codigo_cliente, sistema_origem)
	);`

	_, err = configDB.Exec(createMigrationStateTableSQL)
	if err != nil {
		return fmt.Errorf("erro ao criar tabela de estado da migração: %w", err)
	}

	// Criar tabela de logs se não existir
	createLogsTableSQL := `
	CREATE TABLE IF NOT EXISTS migration_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp TEXT NOT NULL,
		level TEXT NOT NULL,
		message TEXT NOT NULL,
		codigo_cliente TEXT,
		sistema_origem TEXT,
		tipo_migracao TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`

	_, err = configDB.Exec(createLogsTableSQL)
	if err != nil {
		return fmt.Errorf("erro ao criar tabela de logs: %w", err)
	}

	// Criar índices separadamente
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON migration_logs(timestamp)",
		"CREATE INDEX IF NOT EXISTS idx_logs_cliente ON migration_logs(codigo_cliente, sistema_origem)",
		"CREATE INDEX IF NOT EXISTS idx_logs_level ON migration_logs(level)",
	}

	for _, idxSQL := range indexes {
		_, err = configDB.Exec(idxSQL)
		if err != nil {
			// Não falhar se o índice já existir
			log.Printf("Aviso ao criar índice de logs: %v", err)
		}
	}

	// Log silencioso para aplicação GUI - logs vão para arquivo se configurado
	// logger.Info("Banco de configurações SQLite inicializado com sucesso")
	return nil
}

// Fecha a conexão com o banco de configurações
func CloseConfigDB() error {
	if configDB != nil {
		return configDB.Close()
	}
	return nil
}

// Salva configuração do cliente no SQLite
func SaveClientConfig(config *ClientConfig) error {
	if configDB == nil {
		return fmt.Errorf("banco de configurações não inicializado")
	}

	// Verificar se já existe configuração para este cliente e sistema
	var count int
	err := configDB.QueryRow(
		"SELECT COUNT(*) FROM client_configs WHERE codigo_cliente = ? AND sistema_origem = ?",
		config.CodigoCliente, config.SistemaOrigem,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("erro ao verificar configuração existente: %w", err)
	}

	if count > 0 {
		// Atualizar configuração existente
		_, err = configDB.Exec(`
			UPDATE client_configs SET 
				db_path = ?, db_user = ?, db_password = ?, db_host = ?, db_port = ?,
				alias_pharmacie = ?, ipserver_pharmacie = ?, porta_pharmacie = ?,
				conversao = ?, ultima_atualizacao = ?
			WHERE codigo_cliente = ? AND sistema_origem = ?`,
			config.DbPath, config.DbUser, config.DbPassword, config.DbHost, config.DbPort,
			config.AliasPharmacie, config.IpServerPharmacie, config.PortaPharmacie,
			config.Conversao, config.UltimaAtualizacao,
			config.CodigoCliente, config.SistemaOrigem,
		)
		if err != nil {
			return fmt.Errorf("erro ao atualizar configuração: %w", err)
		}
		log.Printf("Configuração atualizada para cliente %s sistema %s", config.CodigoCliente, config.SistemaOrigem)
	} else {
		// Inserir nova configuração
		_, err = configDB.Exec(`
			INSERT INTO client_configs (
				codigo_cliente, sistema_origem, db_path, db_user, db_password, 
				db_host, db_port, alias_pharmacie, ipserver_pharmacie, porta_pharmacie,
				conversao, ultima_atualizacao
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			config.CodigoCliente, config.SistemaOrigem, config.DbPath, config.DbUser, config.DbPassword,
			config.DbHost, config.DbPort, config.AliasPharmacie, config.IpServerPharmacie, config.PortaPharmacie,
			config.Conversao, config.UltimaAtualizacao,
		)
		if err != nil {
			return fmt.Errorf("erro ao inserir configuração: %w", err)
		}
		log.Printf("Configuração salva para cliente %s sistema %s", config.CodigoCliente, config.SistemaOrigem)
	}

	return nil
}

// Carrega configuração do cliente do SQLite
func LoadClientConfig(codigoCliente, sistemaOrigem string) (*ClientConfig, error) {
	if configDB == nil {
		return nil, fmt.Errorf("banco de configurações não inicializado")
	}

	var config ClientConfig
	err := configDB.QueryRow(`
		SELECT id, codigo_cliente, sistema_origem, db_path, db_user, db_password,
		       db_host, db_port, alias_pharmacie, ipserver_pharmacie, porta_pharmacie,
		       conversao, ultima_atualizacao
		FROM client_configs 
		WHERE codigo_cliente = ? AND sistema_origem = ?`,
		codigoCliente, sistemaOrigem,
	).Scan(
		&config.ID, &config.CodigoCliente, &config.SistemaOrigem, &config.DbPath,
		&config.DbUser, &config.DbPassword, &config.DbHost, &config.DbPort,
		&config.AliasPharmacie, &config.IpServerPharmacie, &config.PortaPharmacie,
		&config.Conversao, &config.UltimaAtualizacao,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Configuração não encontrada
		}
		return nil, fmt.Errorf("erro ao carregar configuração: %w", err)
	}

	return &config, nil
}

// Lista todas as configurações salvas
func ListClientConfigs() ([]ClientConfig, error) {
	if configDB == nil {
		return nil, fmt.Errorf("banco de configurações não inicializado")
	}

	rows, err := configDB.Query(`
		SELECT id, codigo_cliente, sistema_origem, db_path, db_user, db_password,
		       db_host, db_port, alias_pharmacie, ipserver_pharmacie, porta_pharmacie,
		       conversao, ultima_atualizacao
		FROM client_configs 
		ORDER BY codigo_cliente, sistema_origem`)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar configurações: %w", err)
	}
	defer rows.Close()

	var configs []ClientConfig
	for rows.Next() {
		var config ClientConfig
		err := rows.Scan(
			&config.ID, &config.CodigoCliente, &config.SistemaOrigem, &config.DbPath,
			&config.DbUser, &config.DbPassword, &config.DbHost, &config.DbPort,
			&config.AliasPharmacie, &config.IpServerPharmacie, &config.PortaPharmacie,
			&config.Conversao, &config.UltimaAtualizacao,
		)
		if err != nil {
			return nil, fmt.Errorf("erro ao escanear configuração: %w", err)
		}
		configs = append(configs, config)
	}

	return configs, nil
}

// Remove configuração do cliente
func DeleteClientConfig(codigoCliente, sistemaOrigem string) error {
	if configDB == nil {
		return fmt.Errorf("banco de configurações não inicializado")
	}

	result, err := configDB.Exec(
		"DELETE FROM client_configs WHERE codigo_cliente = ? AND sistema_origem = ?",
		codigoCliente, sistemaOrigem,
	)
	if err != nil {
		return fmt.Errorf("erro ao deletar configuração: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erro ao verificar linhas afetadas: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("configuração não encontrada para cliente %s sistema %s", codigoCliente, sistemaOrigem)
	}

	log.Printf("Configuração removida para cliente %s sistema %s", codigoCliente, sistemaOrigem)
	return nil
}

// Converte ClientConfig para DbConfig (usado pelo sistema existente)
func (c *ClientConfig) ToDbConfig() DbConfig {
	return DbConfig{
		Path:      c.DbPath,
		User:      c.DbUser,
		Password:  c.DbPassword,
		Host:      c.DbHost,
		Port:      c.DbPort,
		Conversao: c.Conversao,
		LogLevel:  "INFO",
		LogFile:   "migration.log",
	}
}

// Função para obter configurações do cliente com persistência SQLite
func GetClientConfig(codigoCliente, sistemaNome string) (*ClientConfig, error) {
	// Tentar carregar configurações existentes do SQLite
	existingConfig, err := LoadClientConfig(codigoCliente, sistemaNome)
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar configurações existentes: %v", err)
	}

	// Se existe configuração, perguntar se quer reutilizar
	if existingConfig != nil {
		fmt.Printf("Configurações encontradas para o cliente %s no sistema %s:\n", codigoCliente, sistemaNome)
		fmt.Printf("Banco: %s:%s/%s\n", existingConfig.DbHost, existingConfig.DbPort, existingConfig.DbPath)
		fmt.Printf("Conversão: %s\n", existingConfig.Conversao)
		fmt.Print("Deseja reutilizar essas configurações? (s/n): ")

		var resposta string
		fmt.Scanln(&resposta)
		resposta = strings.TrimSpace(strings.ToLower(resposta))

		if resposta == "s" || resposta == "sim" || resposta == "y" || resposta == "yes" {
			return existingConfig, nil
		}
	}

	// Se não existe ou usuário não quer reutilizar, criar nova configuração
	fmt.Printf("Configurando cliente %s para sistema %s...\n", codigoCliente, sistemaNome)

	// Obter configurações do banco de origem
	dbPath, user, password, host, port, _ := config.LoadConfigSetup(sistemaNome)

	// Se for FCERTA ou PRISMA5, solicitar dados do Pharmacie
	var aliasPharmacie, ipServerPharmacie, portaPharmacie, conversao string
	if sistemaNome == "FCERTA" || sistemaNome == "PRISMA5" {
		fmt.Println("Informe os dados de conexão do sistema Pharmacie (destino):")
		fmt.Print("ALIAS: ")
		fmt.Scanln(&aliasPharmacie)
		aliasPharmacie = removeQuotes(strings.TrimSpace(aliasPharmacie))

		fmt.Print("IPSERVER: ")
		fmt.Scanln(&ipServerPharmacie)
		ipServerPharmacie = removeQuotes(strings.TrimSpace(ipServerPharmacie))

		fmt.Print("PORTA: ")
		fmt.Scanln(&portaPharmacie)
		portaPharmacie = removeQuotes(strings.TrimSpace(portaPharmacie))

		// Solicitar conversão
		fmt.Print("Digite o valor da conversão (1, 2, etc.): ")
		fmt.Scanln(&conversao)
		conversao = removeQuotes(strings.TrimSpace(conversao))
	}

	// Criar nova configuração
	config := NewClientConfigFromDbConfig(codigoCliente, sistemaNome, dbPath, user, password, host, port, aliasPharmacie, ipServerPharmacie, portaPharmacie, conversao)

	// Salvar configuração no SQLite
	if err := SaveClientConfig(config); err != nil {
		return nil, fmt.Errorf("erro ao salvar configurações: %v", err)
	}

	return config, nil
}

// Cria ClientConfig a partir dos parâmetros individuais
func NewClientConfigFromDbConfig(codigoCliente, sistemaOrigem, dbPath, user, password, host, port, aliasPharmacie, ipServerPharmacie, portaPharmacie, conversao string) *ClientConfig {
	return &ClientConfig{
		CodigoCliente:     codigoCliente,
		SistemaOrigem:     sistemaOrigem,
		DbPath:            dbPath,
		DbUser:            user,
		DbPassword:        password,
		DbHost:            host,
		DbPort:            port,
		AliasPharmacie:    aliasPharmacie,
		IpServerPharmacie: ipServerPharmacie,
		PortaPharmacie:    portaPharmacie,
		Conversao:         conversao,
		UltimaAtualizacao: fmt.Sprintf("%d", time.Now().Unix()),
	}
}

// Carrega o estado da migração do SQLite
func LoadMigrationState(codigoCliente, sistemaOrigem string) (*MigrationState, error) {
	if configDB == nil {
		return nil, fmt.Errorf("banco de configurações não inicializado")
	}

	query := `SELECT id, codigo_cliente, sistema_origem, configuracao_inicial_executada, 
			  procedures_criadas, inserts_executados, ultima_atualizacao 
			  FROM migration_states 
			  WHERE codigo_cliente = ? AND sistema_origem = ?`

	var state MigrationState
	var proceduresCriadasJSON, insertsExecutadosJSON string

	err := configDB.QueryRow(query, codigoCliente, sistemaOrigem).Scan(
		&state.ID,
		&state.CodigoCliente,
		&state.SistemaOrigem,
		&state.ConfiguracaoInicialExecutada,
		&proceduresCriadasJSON,
		&insertsExecutadosJSON,
		&state.UltimaAtualizacao,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Se não existe, retorna estado vazio
			return &MigrationState{
				CodigoCliente:                codigoCliente,
				SistemaOrigem:                sistemaOrigem,
				ConfiguracaoInicialExecutada: false,
				ProceduresCriadas:            []string{},
				InsertsExecutados:            []string{},
				UltimaAtualizacao:            fmt.Sprintf("%d", time.Now().Unix()),
			}, nil
		}
		return nil, fmt.Errorf("erro ao carregar estado da migração: %w", err)
	}

	// Converter JSON strings para slices
	state.ProceduresCriadas = parseJSONStringArray(proceduresCriadasJSON)
	state.InsertsExecutados = parseJSONStringArray(insertsExecutadosJSON)

	return &state, nil
}

// Salva o estado da migração no SQLite
func SaveMigrationState(state *MigrationState) error {
	if configDB == nil {
		return fmt.Errorf("banco de configurações não inicializado")
	}

	// Converter slices para JSON strings
	proceduresCriadasJSON := stringsToJSONStringArray(state.ProceduresCriadas)
	insertsExecutadosJSON := stringsToJSONStringArray(state.InsertsExecutados)

	// Atualizar timestamp
	state.UltimaAtualizacao = fmt.Sprintf("%d", time.Now().Unix())

	// Usar UPSERT (INSERT OR REPLACE)
	query := `INSERT OR REPLACE INTO migration_states 
			  (codigo_cliente, sistema_origem, configuracao_inicial_executada, 
			   procedures_criadas, inserts_executados, ultima_atualizacao)
			  VALUES (?, ?, ?, ?, ?, ?)`

	_, err := configDB.Exec(query,
		state.CodigoCliente,
		state.SistemaOrigem,
		state.ConfiguracaoInicialExecutada,
		proceduresCriadasJSON,
		insertsExecutadosJSON,
		state.UltimaAtualizacao,
	)

	if err != nil {
		return fmt.Errorf("erro ao salvar estado da migração: %w", err)
	}

	return nil
}

// Função auxiliar para converter slice de strings para JSON string
func stringsToJSONStringArray(slice []string) string {
	if len(slice) == 0 {
		return "[]"
	}
	result := "["
	for i, s := range slice {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf(`"%s"`, s)
	}
	result += "]"
	return result
}

// Função auxiliar para converter JSON string para slice de strings
func parseJSONStringArray(jsonStr string) []string {
	if jsonStr == "" || jsonStr == "[]" {
		return []string{}
	}

	// Remove colchetes e aspas
	jsonStr = strings.Trim(jsonStr, "[]")
	if jsonStr == "" {
		return []string{}
	}

	// Split por vírgula e remove aspas
	parts := strings.Split(jsonStr, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		part = strings.Trim(part, `"`)
		if part != "" {
			result = append(result, part)
		}
	}

	return result
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

// SaveLog salva um log no banco SQLite com mutex para evitar concorrência
func SaveLog(level, message, codigoCliente, sistemaOrigem, tipoMigracao string) error {
	if configDB == nil {
		// Se o banco não estiver inicializado, apenas ignorar (não falhar)
		return nil
	}

	// Serializar acesso ao SQLite com mutex
	logDBMutex.Lock()
	defer logDBMutex.Unlock()

	// Verificar se o banco ainda está aberto
	if err := configDB.Ping(); err != nil {
		// Tentar reabrir se necessário
		var reopenErr error
		configDB, reopenErr = sql.Open("sqlite", configSQLiteDSN())
		if reopenErr != nil {
			return nil // Ignorar erro silenciosamente
		}
		configDB.SetMaxOpenConns(1)
		configDB.SetMaxIdleConns(1)
		configDB.SetConnMaxLifetime(0)
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// Usar valores NULL se os campos opcionais estiverem vazios
	var codigoClienteNull, sistemaOrigemNull, tipoMigracaoNull interface{}
	if codigoCliente != "" {
		codigoClienteNull = codigoCliente
	}
	if sistemaOrigem != "" {
		sistemaOrigemNull = sistemaOrigem
	}
	if tipoMigracao != "" {
		tipoMigracaoNull = tipoMigracao
	}

	// Truncar mensagem se muito longa (limite do SQLite TEXT é grande, mas por segurança)
	if len(message) > 10000 {
		message = message[:10000] + "... [truncado]"
	}

	// Tentar inserir com retry em caso de busy
	maxRetries := 3
	var err error
	for i := 0; i < maxRetries; i++ {
		_, err = configDB.Exec(`
			INSERT INTO migration_logs (timestamp, level, message, codigo_cliente, sistema_origem, tipo_migracao)
			VALUES (?, ?, ?, ?, ?, ?)`,
			timestamp, level, message, codigoClienteNull, sistemaOrigemNull, tipoMigracaoNull,
		)
		
		if err == nil {
			return nil
		}
		
		errStr := err.Error()
		
		// Se não for erro de busy/locked, logar e retornar imediatamente
		if !strings.Contains(errStr, "locked") && !strings.Contains(errStr, "busy") && !strings.Contains(errStr, "SQLITE_BUSY") {
			// Logar apenas na primeira tentativa para não poluir
			if i == 0 {
				// Tentar garantir que o erro seja visível, mas não quebrar o fluxo
				// Usar stderr para garantir que aparece no console mesmo em produção
				fmt.Fprintf(os.Stderr, "[SQLite Log Error] Erro ao salvar log: %v\n", err)
				fmt.Fprintf(os.Stderr, "[SQLite Log Error] Mensagem (primeiros 200 chars): %s\n", 
					func() string {
						if len(message) > 200 {
							return message[:200] + "..."
						}
						return message
					}())
			}
			return nil // Ignorar erro silenciosamente para não quebrar o fluxo
		}
		
		// Aguardar antes de tentar novamente (backoff exponencial)
		time.Sleep(time.Duration(i+1) * 10 * time.Millisecond)
	}
	
	// Se chegou aqui, todas as tentativas falharam (provavelmente SQLITE_BUSY)
	// Logar apenas uma vez para debug
	fmt.Fprintf(os.Stderr, "[SQLite Log Error] Falha após %d tentativas (SQLITE_BUSY/locked)\n", maxRetries)
	return nil // Ignorar erro silenciosamente para não quebrar o fluxo
}

// GetLogs retorna os logs do SQLite com filtros opcionais
func GetLogs(codigoCliente, sistemaOrigem string, limit int) ([]map[string]interface{}, error) {
	if configDB == nil {
		return nil, fmt.Errorf("banco de configurações não inicializado")
	}

	var query string
	var args []interface{}

	if codigoCliente != "" && sistemaOrigem != "" {
		query = `
			SELECT id, timestamp, level, message, codigo_cliente, sistema_origem, tipo_migracao, created_at
			FROM migration_logs
			WHERE codigo_cliente = ? AND sistema_origem = ?
			ORDER BY timestamp DESC, id DESC
			LIMIT ?
		`
		args = []interface{}{codigoCliente, sistemaOrigem, limit}
	} else {
		query = `
			SELECT id, timestamp, level, message, codigo_cliente, sistema_origem, tipo_migracao, created_at
			FROM migration_logs
			ORDER BY timestamp DESC, id DESC
			LIMIT ?
		`
		args = []interface{}{limit}
	}

	rows, err := configDB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar logs: %w", err)
	}
	defer rows.Close()

	var logs []map[string]interface{}
	for rows.Next() {
		var id int
		var timestamp, level, message string
		var codigoCliente, sistemaOrigem, tipoMigracao sql.NullString
		var createdAt sql.NullTime

		err := rows.Scan(&id, &timestamp, &level, &message, &codigoCliente, &sistemaOrigem, &tipoMigracao, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("erro ao escanear log: %w", err)
		}

		logEntry := map[string]interface{}{
			"id":             id,
			"timestamp":      timestamp,
			"level":          level,
			"message":        message,
			"codigo_cliente": codigoCliente.String,
			"sistema_origem": sistemaOrigem.String,
			"tipo_migracao":  tipoMigracao.String,
			"created_at":     createdAt.Time,
		}

		logs = append(logs, logEntry)
	}

	return logs, nil
}
