package prisma5

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "github.com/nakagami/firebirdsql"
	"github.com/primesoftwaresi/prime-migration/pkg/logger"
)

// ============================================================================
// FUNÇÕES AUXILIARES COMPARTILHADAS
// ============================================================================

// tiraAcentos remove acentos de uma string (equivalente à procedure TIRA_ACENTOS)
func tiraAcentos(texto string) string {
	if texto == "" {
		return ""
	}

	acentos := map[rune]rune{
		'à': 'a', 'â': 'a', 'ã': 'a', 'á': 'a',
		'À': 'A', 'Â': 'A', 'Ã': 'A', 'Á': 'A',
		'ê': 'e', 'é': 'e',
		'Ê': 'E', 'É': 'E',
		'ô': 'o', 'õ': 'o', 'ó': 'o',
		'Ô': 'O', 'Ó': 'O', 'Õ': 'O',
		'û': 'u', 'ú': 'u', 'ü': 'u',
		'Û': 'U', 'Ú': 'U', 'Ü': 'U',
		'í': 'i',
		'Í': 'I',
		'ç': 'c',
		'Ç': 'C',
		'ñ': 'n',
		'Ñ': 'N',
	}

	var result strings.Builder
	for _, char := range texto {
		if replacement, ok := acentos[char]; ok {
			result.WriteRune(replacement)
		} else {
			result.WriteRune(char)
		}
	}
	return result.String()
}

// stripNonNumeric remove todos os caracteres não numéricos (equivalente à procedure STRIP_NON_NUMERIC)
func stripNonNumeric(texto string) string {
	re := regexp.MustCompile(`[^0-9]`)
	return re.ReplaceAllString(texto, "")
}

// utilLimpaRTF limpa tags RTF de um texto (equivalente à procedure UTIL_LIMPA_RTF)
// Esta função processa campos BLOB/VARCHAR que podem conter RTF do Firebird
// Versão otimizada usando regex para melhor performance e cobertura completa
func utilLimpaRTF(textoRTF string) string {
	if strings.TrimSpace(textoRTF) == "" {
		return ""
	}

	texto := textoRTF

	// ETAPA 1: Substituir caracteres especiais do RTF (formato \'xx para caracteres acentuados)
	// Usando regex para capturar todos os padrões \'[a-f0-9]{2} de uma vez
	rtfCharRegex := regexp.MustCompile(`\\'([a-f0-9]{2})`)
	texto = rtfCharRegex.ReplaceAllStringFunc(texto, func(match string) string {
		hexCode := match[2:4] // Extrai o código hex (ex: 'c0)
		code, err := strconv.ParseInt(hexCode, 16, 8)
		if err == nil && code >= 192 && code <= 252 {
			// Mapeia códigos ASCII para caracteres acentuados
			charMap := map[int64]rune{
				192: 'À', 193: 'Á', 194: 'Â', 195: 'Ã',
				199: 'Ç', 200: 'È', 201: 'É', 202: 'Ê',
				204: 'Ì', 205: 'Í', 210: 'Ò', 211: 'Ó',
				212: 'Ô', 213: 'Õ', 217: 'Ù', 218: 'Ú',
				220: 'Ü', 224: 'à', 225: 'á', 226: 'â',
				227: 'ã', 231: 'ç', 232: 'è', 233: 'é',
				234: 'ê', 236: 'ì', 237: 'í', 242: 'ò',
				243: 'ó', 244: 'ô', 245: 'õ', 249: 'ù',
				250: 'ú', 252: 'ü',
			}
			if char, ok := charMap[code]; ok {
				return string(char)
			}
			// Se não estiver no mapa, converte diretamente do código ASCII
			return string(rune(code))
		}
		return match
	})

	// ETAPA 2: Remover todas as tags RTF usando regex (muito mais eficiente)
	// Remove tags RTF: \comando ou \comando{número}
	texto = regexp.MustCompile(`\\[a-z]+[0-9]*\*?`).ReplaceAllString(texto, "")

	// Remove blocos de controle RTF: {fonttbl}, {colortbl}, etc
	texto = regexp.MustCompile(`\{[^}]*\}`).ReplaceAllString(texto, "")

	// Remove caracteres de controle isolados
	texto = regexp.MustCompile(`[{}]`).ReplaceAllString(texto, "")

	// Remove escapes de barra invertida restantes
	texto = regexp.MustCompile(`\\[^a-z]`).ReplaceAllString(texto, "")

	// Substituir quebras de linha RTF por espaços
	texto = regexp.MustCompile(`\\par|\\line|ltrpar`).ReplaceAllString(texto, " ")

	// Remover nomes de fontes e estilos específicos
	fontes := []string{`MS Sans Serif`, `Courier New`, `Arial`, `Calibri`}
	for _, fonte := range fontes {
		texto = regexp.MustCompile(regexp.QuoteMeta(fonte)+`;?`).ReplaceAllString(texto, "")
	}

	// ETAPA 3: Limpar caracteres especiais restantes
	texto = regexp.MustCompile(`[\\;]+`).ReplaceAllString(texto, "")

	// ETAPA 4: Remover espaços duplos/múltiplos (otimizado com regex)
	texto = regexp.MustCompile(`\s+`).ReplaceAllString(texto, " ")

	return strings.TrimSpace(texto)
}

// getPharmacieConnection obtém a string de conexão do Pharmacie da tabela CONEXAO
func getPharmacieConnection(prisma5DB *sql.DB) (string, error) {
	var connStr string
	err := prisma5DB.QueryRow(`
		SELECT FIRST 1 IPSERVER || '/' || PORTA || ':' || ALIAS 
		FROM CONEXAO
	`).Scan(&connStr)

	if err != nil {
		return "", fmt.Errorf("erro ao obter conexão Pharmacie: %w", err)
	}

	if connStr == "" {
		return "", fmt.Errorf("CONEXÃO BANCO DE DADOS PHARMACIE NÃO CONFIGURADA")
	}

	return connStr, nil
}

// parsePharmacieConnection parseia a string de conexão no formato "IP/PORTA:ALIAS"
// O ALIAS pode conter ":" (como em C:\path:with\colon), então fazemos split apenas no primeiro ":"
func parsePharmacieConnection(connStr string) (host, port, alias string, err error) {
	// Formato: IP/PORTA:ALIAS
	// Usar SplitN com limite 2 para pegar apenas o primeiro ":" (o ALIAS pode conter ":")
	parts := strings.SplitN(connStr, ":", 2)
	if len(parts) != 2 {
		return "", "", "", fmt.Errorf("formato de conexão inválido: %s", connStr)
	}

	alias = parts[1]

	// Separar IP/PORTA
	parts2 := strings.Split(parts[0], "/")
	if len(parts2) != 2 {
		return "", "", "", fmt.Errorf("formato de conexão inválido: %s", connStr)
	}

	host = parts2[0]
	port = parts2[1]

	return host, port, alias, nil
}

// pharmacieConfig configuração de conexão com o banco Pharmacie
type pharmacieConfig struct {
	Host     string
	Port     string
	Path     string
	User     string
	Password string
}

// connectToFirebird conecta ao banco Firebird usando a configuração fornecida
func connectToFirebird(cfg pharmacieConfig) (*sql.DB, error) {
	// Validar configuração antes de tentar conectar
	if err := validarPharmacieConfig(cfg); err != nil {
		return nil, fmt.Errorf("configuração inválida: %w", err)
	}

	connStr := fmt.Sprintf("%s:%s@%s:%s/%s?charset=WIN1252",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Path)

	db, err := sql.Open("firebirdsql", connStr)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir conexão com banco: %w", err)
	}

	// Configurar timeouts
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	// Testar conexão com timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("erro ao testar conexão com banco: %w", err)
	}

	return db, nil
}

// validarPharmacieConfig valida se a configuração está correta
func validarPharmacieConfig(cfg pharmacieConfig) error {
	if cfg.Path == "" {
		return fmt.Errorf("caminho do banco de dados não pode estar vazio")
	}
	if cfg.User == "" {
		return fmt.Errorf("usuário não pode estar vazio")
	}
	if cfg.Password == "" {
		return fmt.Errorf("senha não pode estar vazia")
	}
	if cfg.Host == "" {
		return fmt.Errorf("host não pode estar vazio")
	}
	if cfg.Port == "" {
		return fmt.Errorf("porta não pode estar vazia")
	}
	return nil
}

// connectToPharmacie conecta ao banco Pharmacie usando as informações de CONEXAO
func connectToPharmacie(prisma5DB *sql.DB) (*sql.DB, error) {
	connStr, err := getPharmacieConnection(prisma5DB)
	if err != nil {
		return nil, err
	}

	host, port, alias, err := parsePharmacieConnection(connStr)
	if err != nil {
		return nil, err
	}

	logger.Info("Conectando ao Pharmacie: %s/%s:%s", host, port, alias)

	// Conectar ao Pharmacie (usar SYSDBA/SySPs_PHARMACIE)
	cfg := pharmacieConfig{
		Host:     host,
		Port:     port,
		Path:     alias,
		User:     "SYSDBA",
		Password: "SySPs_PHARMACIE",
	}

	return connectToFirebird(cfg)
}

func generateCodigoPSInBatch(dbConn *sql.DB, table, codigoField, sequenceName string, offset int) error {
	logger.Info("Gerando CODIGO_PS em lote para %s usando sequence %s...", table, sequenceName)

	// Primeiro, contar quantos registros precisam de CODIGO_PS
	var count int
	err := dbConn.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*) FROM %s WHERE CODIGO_PS IS NULL
	`, table)).Scan(&count)

	if err != nil {
		return fmt.Errorf("erro ao contar registros de %s: %w", table, err)
	}

	if count == 0 {
		logger.Info("Nenhum registro de %s precisa de CODIGO_PS", table)
		return nil
	}

	logger.Info("Gerando CODIGO_PS para %d registros de %s...", count, table)

	// Para Firebird, a melhor abordagem é UPDATE direto com GEN_ID no SET
	// Firebird otimiza isso muito bem quando feito em massa
	// Vamos fazer em um único UPDATE se possível, ou em lotes se necessário

	var updateSQL string
	if offset == 0 {
		updateSQL = fmt.Sprintf(`
			UPDATE %s 
			SET CODIGO_PS = GEN_ID(%s, 1)
			WHERE CODIGO_PS IS NULL
		`, table, sequenceName)
	} else {
		updateSQL = fmt.Sprintf(`
			UPDATE %s 
			SET CODIGO_PS = GEN_ID(%s, 1) + ?
			WHERE CODIGO_PS IS NULL
		`, table, sequenceName)
	}

	result, err := dbConn.Exec(updateSQL, offset)
	if err != nil {
		return fmt.Errorf("erro ao gerar CODIGO_PS em lote para %s: %w", table, err)
	}

	rowsAffected, _ := result.RowsAffected()
	logger.Info("✅ %d registros de %s com CODIGO_PS gerado", rowsAffected, table)

	return nil
}
