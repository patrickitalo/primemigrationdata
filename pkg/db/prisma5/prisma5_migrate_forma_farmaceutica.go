package prisma5

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/primesoftwaresi/prime-migration/pkg/logger"
)

// MigratePRISMA5FormaFarmaceutica migra formas farmacêuticas do PRISMA5 para o Pharmacie
// Nota: Esta migração não precisa de parâmetro de conversão
func MigratePRISMA5FormaFarmaceutica(prisma5DB *sql.DB) error {
	logger.Info("Iniciando migração de formas farmacêuticas")

	// Conectar ao Pharmacie
	pharmacieDB, err := connectToPharmacie(prisma5DB)
	if err != nil {
		return fmt.Errorf("erro ao conectar ao Pharmacie: %w", err)
	}
	defer pharmacieDB.Close()

	// 1. Gera CODIGO_PS para formas farmacêuticas usando sequence (OTIMIZADO: UPDATE em massa)
	// Nota: A procedure SQL original usa ORDER BY FF.DESCRICAOFORMA ASC, mas como o Firebird
	// gera os valores de sequence sequencialmente, a ordem final será preservada mesmo sem ORDER BY
	logger.Info("Gerando CODIGO_PS para formas farmacêuticas...")
	_, err = prisma5DB.Exec(`
		UPDATE FORMAFARMACEUTICA FF
		SET FF.CODIGO_PS = GEN_ID(GEN_FORMAFARMACEUTICA, 1)
		WHERE FF.CODIGO_PS IS NULL
		ORDER BY FF.DESCRICAOFORMA ASC
	`)
	if err != nil {
		return fmt.Errorf("erro ao gerar CODIGO_PS para formas farmacêuticas: %w", err)
	}
	logger.Info("✅ CODIGO_PS para formas farmacêuticas gerado")

	// 2. Insere formas farmacêuticas no banco Pharmacie
	logger.Info("Inserindo formas farmacêuticas no Pharmacie...")
	formasQuery := `
		SELECT
			FF.CODIGO_PS AS CODIGO,
			UPPER(FF.DESCRICAOFORMA) AS NOMEFORMAFARMACEUTICA,
			UPPER(FF.DESCRICAOFORMA) AS NOMEROTULO,
			3 AS TIPO_USO,
			FF.VALIDADEFORMA AS VALIDADE,
			CAST(FF.VALORMINIMOFORMA AS NUMERIC(18,2)) AS VALORMINIMOATENDIMENTO,
			1 AS CODIGO_LABORATORIO,
			0 AS TIPOINFORME,
			1 AS CODIGO_FORMAFARMACEUTICA_TU,
			1 AS METODOUTILIZACAO,
			1 AS USOVALIDADE,
			1 AS DESCRICAOROTULO_TIPO,
			1 AS CADASTRO_CF,
			CURRENT_TIMESTAMP AS CADASTRO_DT,
			1 AS CADASTRO_LJ,
			1 AS ALTERACAO_CF,
			CURRENT_TIMESTAMP AS ALTERACAO_DT,
			1 AS ALTERACAO_LJ,
			-1 AS ATIVO
		FROM FORMAFARMACEUTICA FF
		WHERE FF.CODIGO_PS IS NOT NULL
		ORDER BY 1
	`

	rows, err := prisma5DB.Query(formasQuery)
	if err != nil {
		return fmt.Errorf("erro ao buscar formas farmacêuticas para inserir: %w", err)
	}
	defer rows.Close()

	insertFormaSQL := `INSERT INTO FORMAFARMACEUTICA (CODIGO, NOMEFORMAFARMACEUTICA, NOMEROTULO, TIPO_USO, VALIDADE, VALORMINIMOATENDIMENTO, CODIGO_LABORATORIO, TIPOINFORME, CODIGO_FORMAFARMACEUTICA_TU, METODOUTILIZACAO, USOVALIDADE, DESCRICAOROTULO_TIPO, CADASTRO_CF, CADASTRO_DT, CADASTRO_LJ, ALTERACAO_CF, ALTERACAO_DT, ALTERACAO_LJ, ATIVO) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	tx, err := pharmacieDB.Begin()
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação para formas farmacêuticas: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(insertFormaSQL)
	if err != nil {
		return fmt.Errorf("erro ao preparar statement de formas farmacêuticas: %w", err)
	}
	defer stmt.Close()

	contadorFormas := 0
	for rows.Next() {
		var codigo, tipoUso, validade, codigoLaboratorio, tipoInforme, codigoFormaFarmaceuticaTU, metodoUtilizacao, usoValidade, descricaoRotuloTipo, cadastroCF, cadastroLJ, alteracaoCF, alteracaoLJ, ativo sql.NullInt64
		var nomeFormaFarmaceutica, nomeRotulo sql.NullString
		var valorMinimoAtendimento sql.NullFloat64
		var cadastroDT, alteracaoDT sql.NullTime

		err := rows.Scan(&codigo, &nomeFormaFarmaceutica, &nomeRotulo, &tipoUso, &validade, &valorMinimoAtendimento, &codigoLaboratorio, &tipoInforme, &codigoFormaFarmaceuticaTU, &metodoUtilizacao, &usoValidade, &descricaoRotuloTipo, &cadastroCF, &cadastroDT, &cadastroLJ, &alteracaoCF, &alteracaoDT, &alteracaoLJ, &ativo)
		if err != nil {
			logger.Warn("Erro ao escanear forma farmacêutica: %v", err)
			continue
		}

		if !codigo.Valid {
			continue
		}

		_, err = stmt.Exec(codigo, nomeFormaFarmaceutica, nomeRotulo, tipoUso, validade, valorMinimoAtendimento, codigoLaboratorio, tipoInforme, codigoFormaFarmaceuticaTU, metodoUtilizacao, usoValidade, descricaoRotuloTipo, cadastroCF, cadastroDT, cadastroLJ, alteracaoCF, alteracaoDT, alteracaoLJ, ativo)
		if err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "unique") && !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				logger.Warn("Erro ao inserir forma farmacêutica %d: %v", codigo.Int64, err)
			}
			continue
		}
		contadorFormas++
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("erro ao commitar transação de formas farmacêuticas: %w", err)
	}
	logger.Info("✅ %d formas farmacêuticas inseridas no Pharmacie", contadorFormas)

	logger.Info("✅ Migração de formas farmacêuticas concluída com sucesso!")
	return nil
}
