package prisma5

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/primesoftwaresi/prime-migration/pkg/logger"
)

// MigratePRISMA5Lotes migra lotes do PRISMA5 para o Pharmacie
func MigratePRISMA5Lotes(prisma5DB *sql.DB) error {
	logger.Info("Iniciando migração de lotes")

	// Conectar ao Pharmacie
	pharmacieDB, err := connectToPharmacie(prisma5DB)
	if err != nil {
		return fmt.Errorf("erro ao conectar ao Pharmacie: %w", err)
	}
	defer pharmacieDB.Close()

	// 1. Gera CODIGO_PS_LOTE e CODIGO_PS_LOTE_LA para lotes com saldo positivo
	logger.Info("Gerando CODIGO_PS para lotes...")
	_, err = prisma5DB.Exec(`
		UPDATE LOTE L
		SET L.CODIGO_PS_LOTE = GEN_ID(GEN_ESTOQUE_LOTE, 1),
			L.CODIGO_PS_LOTE_LA = GEN_ID(GEN_ESTOQUE_LOTE_LA, 1)
		WHERE CAST(L.QUANTIDADELOTE AS NUMERIC(18,4)) - CAST(L.QUANTIDADECOMPROMETIDALOTE AS NUMERIC(18,4)) > 0
		AND EXISTS (
			SELECT 1
			FROM PRODUTO P
			WHERE P.CODIGOGRUPO = L.CODIGOGRUPO 
			AND P.CODIGOPRODUTO = L.CODIGOPRODUTO 
			AND P.CODIGO_PS IS NOT NULL
		)
	`)
	if err != nil {
		logger.Warn("Aviso ao gerar CODIGO_PS para lotes: %v", err)
	}
	logger.Info("✅ CODIGO_PS para lotes gerado")

	// 2. Insere lotes no Pharmacie (ESTOQUE_LOTE)
	logger.Info("Inserindo lotes no Pharmacie...")
	lotesQuery := `
		SELECT 
			L.CODIGO_PS_LOTE AS CODIGO,
			0 AS TIPOLOTE,
			0 AS CODIGO_MOV,
			0 AS CODIGO_MOV_DET,
			P.CODIGO_PS AS CODIGO_PRODUTO,
			CAST(L.CODIGOFORNECEDOR AS INTEGER) AS CODIGO_FORNECEDOR,
			1 AS CODIGO_ESTOQUE_ORIGEM,
			L.LOTEFORNECEDOR AS LOTE,
			IIF(L.FATORCORRECAOLOTE IS NULL OR L.FATORCORRECAOLOTE = 0, 100, 100 / L.FATORCORRECAOLOTE) AS CONCENTRACAO,
			IIF(L.DENSIDADELOTE IS NULL OR (L.DENSIDADELOTE = 0), 1, L.DENSIDADELOTE) AS DENSIDADE,
			0 AS UMIDADE,
			COALESCE(L.FATORUTR, 0) AS UTR,
			CASE 
				WHEN L.FATORUFC >= 1000000000 THEN L.FATORUFC / 1000000000
				ELSE L.FATORUFC
			END AS UFC,
			CASE L.METODOLOTE
				WHEN 'C' THEN 'CH'
				WHEN 'CH-DG' THEN 'CH'
				WHEN 'CH/' THEN 'CH'
				WHEN 'CHM' THEN 'CH'
				WHEN 'CHM/' THEN 'CH'
				WHEN 'CHMM/M' THEN 'CH'
				WHEN 'CHMM/M/' THEN 'CH'
				WHEN 'CMM' THEN 'CH'
				WHEN 'CMM/' THEN 'CH'
				WHEN 'CMM/M' THEN 'CH'
				WHEN 'CMM/M/' THEN 'CH'
				WHEN 'LOTE:' THEN NULL
				WHEN 'LM/' THEN 'LM'
				ELSE L.METODOLOTE
			END AS TIPOESCALA,
			CAST(L.DINAMIZACAOLOTE AS INTEGER) AS DINAMIZACAO,
			CAST(L.DATAFABRICACAOLOTE AS DATE) AS FABRICACAO_DT,
			CAST(L.DATAVALIDADELOTE AS DATE) AS VENCIMENTO_DT,
			CAST(L.QUANTIDADELOTE AS NUMERIC(18,4)) - CAST(L.QUANTIDADECOMPROMETIDALOTE AS NUMERIC(18,4)) AS QUANTIDADE,
			0 AS FECHARLOTE,
			0 AS LOTEUNIFICADO,
			1 AS CADASTRO_LJ,
			1 AS CADASTRO_CF,
			CAST(L.DT_CREATION AS DATE) AS CADASTRO_DT,
			1 AS ALTERACAO_LJ,
			1 AS ALTERACAO_CF,
			L.DT_ALTER AS ALTERACAO_DT,
			0 AS STATUSGERAL,
			COALESCE(L.FATORUI, 0) AS UI,
			0 AS FRACAO_ENTRADA
		FROM LOTE L
		INNER JOIN PRODUTO P ON P.CODIGOPRODUTO = L.CODIGOPRODUTO AND P.CODIGOGRUPO = L.CODIGOGRUPO
		WHERE CAST(L.QUANTIDADELOTE AS NUMERIC(18,4)) - CAST(L.QUANTIDADECOMPROMETIDALOTE AS NUMERIC(18,4)) > 0
		AND P.CODIGO_PS IS NOT NULL
		AND L.CODIGO_PS_LOTE IS NOT NULL
		ORDER BY 1
	`

	rows, err := prisma5DB.Query(lotesQuery)
	if err != nil {
		return fmt.Errorf("erro ao buscar lotes: %w", err)
	}
	defer rows.Close()

	insertLoteSQL := `INSERT INTO ESTOQUE_LOTE (CODIGO, TIPOLOTE, CODIGO_MOV, CODIGO_MOV_DET, CODIGO_PRODUTO, CODIGO_FORNECEDOR, CODIGO_ESTOQUE_ORIGEM, LOTE, CONCENTRACAO, DENSIDADE, UMIDADE, UTR, UFC, TIPOESCALA, DINAMIZACAO, FABRICACAO_DT, VENCIMENTO_DT, QUANTIDADE, FECHARLOTE, LOTEUNIFICADO, CADASTRO_LJ, CADASTRO_CF, CADASTRO_DT, ALTERACAO_LJ, ALTERACAO_CF, ALTERACAO_DT, STATUSGERAL, UI, FRACAO_ENTRADA) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	tx, err := pharmacieDB.Begin()
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação para lotes: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(insertLoteSQL)
	if err != nil {
		return fmt.Errorf("erro ao preparar statement de lotes: %w", err)
	}
	defer stmt.Close()

	contadorLotes := 0
	for rows.Next() {
		var codigo, tipoLote, codigoMov, codigoMovDet, codigoProduto, codigoFornecedor, codigoEstoqueOrigem, umidade, fecharLote, loteUnificado, cadastroLJ, cadastroCF, alteracaoLJ, alteracaoCF, statusGeral, fracaoEntrada sql.NullInt64
		var lote, tipoEscala sql.NullString
		var concentracao, densidade, utr, ufc, quantidade, ui sql.NullFloat64
		var dinamizacao sql.NullInt64
		var fabricacaoDT, vencimentoDT, cadastroDT, alteracaoDT sql.NullTime

		err := rows.Scan(&codigo, &tipoLote, &codigoMov, &codigoMovDet, &codigoProduto, &codigoFornecedor, &codigoEstoqueOrigem, &lote, &concentracao, &densidade, &umidade, &utr, &ufc, &tipoEscala, &dinamizacao, &fabricacaoDT, &vencimentoDT, &quantidade, &fecharLote, &loteUnificado, &cadastroLJ, &cadastroCF, &cadastroDT, &alteracaoLJ, &alteracaoCF, &alteracaoDT, &statusGeral, &ui, &fracaoEntrada)
		if err != nil {
			logger.Warn("Erro ao escanear lote: %v", err)
			continue
		}

		if !codigo.Valid {
			continue
		}

		_, err = stmt.Exec(codigo, tipoLote, codigoMov, codigoMovDet, codigoProduto, codigoFornecedor, codigoEstoqueOrigem, lote, concentracao, densidade, umidade, utr, ufc, tipoEscala, dinamizacao, fabricacaoDT, vencimentoDT, quantidade, fecharLote, loteUnificado, cadastroLJ, cadastroCF, cadastroDT, alteracaoLJ, alteracaoCF, alteracaoDT, statusGeral, ui, fracaoEntrada)
		if err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "unique") && !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				logger.Warn("Erro ao inserir lote %d: %v", codigo.Int64, err)
			}
			continue
		}
		contadorLotes++
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("erro ao commitar transação de lotes: %w", err)
	}
	logger.Info("✅ %d lotes inseridos no Pharmacie", contadorLotes)

	// 3. Insere lotes em ESTOQUE_LOTE_LA (ligação com estoque)
	logger.Info("Inserindo lotes em ESTOQUE_LOTE_LA...")
	lotesLAQuery := `
		SELECT 
			L.CODIGO_PS_LOTE_LA AS CODIGO,
			CASE
				WHEN P.CODIGO_PS < 1000000 THEN 1
				WHEN P.CODIGO_PS > 1000000 AND P.CODIGO_PS < 1999999 THEN 3
				WHEN P.CODIGO_PS > 2000000 AND P.CODIGO_PS < 2999999 THEN 3
				WHEN P.CODIGO_PS > 3000000 AND P.CODIGO_PS < 3999999 THEN 2
				WHEN P.CODIGO_PS > 4000000 AND P.CODIGO_PS < 4999999 THEN 3
			END AS CODIGO_ESTOQUE,
			L.CODIGO_PS_LOTE AS CODIGO_ESTOQUE_LOTE,
			0 AS CODIGO_MOV,
			CAST(L.QUANTIDADELOTE AS NUMERIC(18,4)) - CAST(L.QUANTIDADECOMPROMETIDALOTE AS NUMERIC(18,4)) AS QTDE_ENTRADA,
			CURRENT_TIMESTAMP AS SALDO_DT,
			1 AS STATUSLOTE,
			0 AS ORIGEMLOTE,
			1 AS CADASTRO_LJ,
			1 AS CADASTRO_CF,
			CURRENT_TIMESTAMP AS CADASTRO_DT,
			1 AS ALTERACAO_LJ,
			1 AS ALTERACAO_CF,
			CURRENT_TIMESTAMP AS ALTERACAO_DT
		FROM LOTE L
		INNER JOIN PRODUTO P ON P.CODIGOPRODUTO = L.CODIGOPRODUTO AND P.CODIGOGRUPO = L.CODIGOGRUPO
		WHERE L.CODIGO_PS_LOTE_LA IS NOT NULL
		ORDER BY 1
	`

	rows, err = prisma5DB.Query(lotesLAQuery)
	if err != nil {
		return fmt.Errorf("erro ao buscar lotes LA: %w", err)
	}
	defer rows.Close()

	insertLoteLASQL := `INSERT INTO ESTOQUE_LOTE_LA (CODIGO, CODIGO_ESTOQUE, CODIGO_ESTOQUE_LOTE, CODIGO_MOV, QTDE_ENTRADA, SALDO_DT, STATUSLOTE, ORIGEMLOTE, CADASTRO_LJ, CADASTRO_CF, CADASTRO_DT, ALTERACAO_LJ, ALTERACAO_CF, ALTERACAO_DT) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	tx, err = pharmacieDB.Begin()
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação para lotes LA: %w", err)
	}
	defer tx.Rollback()

	stmt, err = tx.Prepare(insertLoteLASQL)
	if err != nil {
		return fmt.Errorf("erro ao preparar statement de lotes LA: %w", err)
	}
	defer stmt.Close()

	contadorLotesLA := 0
	for rows.Next() {
		var codigo, codigoEstoque, codigoEstoqueLote, codigoMov, statusLote, origemLote, cadastroLJ, cadastroCF, alteracaoLJ, alteracaoCF sql.NullInt64
		var qtdeEntrada sql.NullFloat64
		var saldoDT, cadastroDT, alteracaoDT sql.NullTime

		err := rows.Scan(&codigo, &codigoEstoque, &codigoEstoqueLote, &codigoMov, &qtdeEntrada, &saldoDT, &statusLote, &origemLote, &cadastroLJ, &cadastroCF, &cadastroDT, &alteracaoLJ, &alteracaoCF, &alteracaoDT)
		if err != nil {
			logger.Warn("Erro ao escanear lote LA: %v", err)
			continue
		}

		if !codigo.Valid {
			continue
		}

		_, err = stmt.Exec(codigo, codigoEstoque, codigoEstoqueLote, codigoMov, qtdeEntrada, saldoDT, statusLote, origemLote, cadastroLJ, cadastroCF, cadastroDT, alteracaoLJ, alteracaoCF, alteracaoDT)
		if err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "unique") && !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				logger.Warn("Erro ao inserir lote LA %d: %v", codigo.Int64, err)
			}
			continue
		}
		contadorLotesLA++
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("erro ao commitar transação de lotes LA: %w", err)
	}
	logger.Info("✅ %d lotes LA inseridos no Pharmacie", contadorLotesLA)

	logger.Info("✅ Migração de lotes concluída com sucesso!")
	return nil
}

