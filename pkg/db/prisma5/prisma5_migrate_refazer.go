package prisma5

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/primesoftwaresi/prime-migration/pkg/logger"
)

// MigratePRISMA5Refazer migra vendas (refazer) do PRISMA5 para o Pharmacie
// Parâmetro conversao: valor da conversão (ex: "1", "2")
func MigratePRISMA5Refazer(prisma5DB *sql.DB, conversao string) error {
	logger.Info("Iniciando migração de refazer (vendas) com conversão: %s", conversao)

	conversaoInt, err := strconv.Atoi(conversao)
	if err != nil {
		return fmt.Errorf("erro ao converter conversão para inteiro: %w", err)
	}

	// Conectar ao Pharmacie
	pharmacieDB, err := connectToPharmacie(prisma5DB)
	if err != nil {
		return fmt.Errorf("erro ao conectar ao Pharmacie: %w", err)
	}
	defer pharmacieDB.Close()

	// 1. Atualiza campo CONVERSAO para VENDA (registros sem CODIGO_PS)
	logger.Info("Atualizando campo CONVERSAO para vendas...")
	updateConversaoVendaSQL := `
		UPDATE VENDA A1 SET A1.CONVERSAO = ? 
		WHERE A1.CODIGO_PS IS NULL AND EXISTS (
			SELECT A2.NUMEROVENDA FROM FORMULAVENDA A2 
			WHERE A2.NUMEROVENDA = A1.NUMEROVENDA
		) AND EXISTS(
			SELECT C.CODIGOCLIENTE FROM CLIENTE C 
			WHERE C.CODIGOCLIENTE = A1.CODIGOCLIENTE
		)
	`
	_, err = prisma5DB.Exec(updateConversaoVendaSQL, conversaoInt)
	if err != nil {
		// Detalhar erro com query formatada
		logger.Warn("❌ ERRO ao atualizar CONVERSAO de vendas:")
		logger.Warn("   =============================================")

		// Mostrar query SQL com numeração de linhas
		lines := strings.Split(updateConversaoVendaSQL, "\n")
		logger.Warn("   Query SQL (com numeração de linhas):")
		for i, line := range lines {
			lineNum := i + 1
			logger.Warn("   [%2d] %s", lineNum, line)
		}

		logger.Warn("   Parâmetro conversaoInt: %d", conversaoInt)
		logger.Warn("   Erro: %v", err)
		logger.Warn("   =============================================")
		logger.Warn("   Análise: O erro indica que A1.CODIGOCLIENTE não foi reconhecido")
		logger.Warn("   Isso pode ocorrer se o Firebird não reconhecer o alias A1 no contexto do EXISTS")
		logger.Warn("   Verifique se a tabela VENDA tem a coluna CODIGOCLIENTE")
	}

	// 2. Atualiza campo CONVERSAO para FORMULAVENDA (registros sem CODIGO_PS_A1)
	logger.Info("Atualizando campo CONVERSAO para fórmulas de venda...")
	_, err = prisma5DB.Exec(`
		UPDATE FORMULAVENDA A2 SET A2.CONVERSAO = ? 
		WHERE A2.CODIGO_PS_A1 IS NULL AND EXISTS (
			SELECT A1.CODIGO_PS FROM VENDA A1 
			WHERE A1.NUMEROVENDA = A2.NUMEROVENDA AND A1.CONVERSAO = ?
		)
	`, conversaoInt, conversaoInt)
	if err != nil {
		logger.Warn("Aviso ao atualizar CONVERSAO de fórmulas de venda: %v", err)
	}

	// 3. Atualiza campo CONVERSAO para ITEMFORMULAVENDA (registros sem CODIGO_PS_A1)
	logger.Info("Atualizando campo CONVERSAO para itens de fórmula de venda...")
	_, err = prisma5DB.Exec(`
		UPDATE ITEMFORMULAVENDA A3 SET A3.CONVERSAO = ? 
		WHERE A3.CODIGO_PS_A1 IS NULL AND EXISTS (
			SELECT A1.CODIGO_PS FROM VENDA A1 
			WHERE A1.NUMEROVENDA = A3.NUMEROVENDA AND A1.CONVERSAO = ?
		) AND EXISTS (
			SELECT P.CODIGO_PS FROM PRODUTO P 
			WHERE P.CODIGOPRODUTO = A3.CODIGOPRODUTO AND P.CODIGOGRUPO = A3.CODIGOGRUPO
		)
	`, conversaoInt, conversaoInt)
	if err != nil {
		logger.Warn("Aviso ao atualizar CONVERSAO de itens de fórmula de venda: %v", err)
	}

	// 4. Gera CODIGO_PS para VENDA (registros com CONVERSAO)
	logger.Info("Gerando CODIGO_PS para vendas...")
	updateCodigoPSVendaSQL := `
		UPDATE VENDA A1
		SET A1.CODIGO_PS = GEN_ID(GEN_REFAZER_A1, 1),
			A1.CODIGO_MEDICO = (SELECT FIRST 1 M.CODIGO_PS FROM FORMULAVENDA A22 INNER JOIN MEDICO M ON M.CODIGOMEDICO = A22.CODIGOMEDICO WHERE A22.NUMEROVENDA = A1.NUMEROVENDA)
		WHERE A1.CONVERSAO = ?
		AND A1.CODIGO_PS IS NULL
		AND EXISTS (SELECT A2.NUMEROVENDA
			FROM FORMULAVENDA A2
			WHERE A2.NUMEROVENDA = A1.NUMEROVENDA)
		AND EXISTS(SELECT C.CODIGOCLIENTE
			FROM CLIENTE C
			WHERE C.CODIGOCLIENTE = A1.CODIGOCLIENTE)
	`
	_, err = prisma5DB.Exec(updateCodigoPSVendaSQL, conversaoInt)
	if err != nil {
		// Detalhar erro com query formatada
		logger.Warn("❌ ERRO ao gerar CODIGO_PS para vendas:")
		logger.Warn("   =============================================")

		// Mostrar query SQL com numeração de linhas
		lines := strings.Split(updateCodigoPSVendaSQL, "\n")
		logger.Warn("   Query SQL (com numeração de linhas):")
		for i, line := range lines {
			lineNum := i + 1
			// Destacar a linha onde provavelmente está o erro (linha com A1.CODIGOCLIENTE)
			if strings.Contains(line, "A1.CODIGOCLIENTE") {
				logger.Warn("   [%2d] %s  <-- ⚠️ LINHA COM ERRO", lineNum, line)
			} else {
				logger.Warn("   [%2d] %s", lineNum, line)
			}
		}

		logger.Warn("   Parâmetro conversaoInt: %d", conversaoInt)
		logger.Warn("   Erro: %v", err)
		logger.Warn("   =============================================")
		logger.Warn("   Análise: O erro indica que A1.CODIGOCLIENTE não foi reconhecido")
		logger.Warn("   Isso pode ocorrer se o Firebird não reconhecer o alias A1 no contexto do EXISTS")
		logger.Warn("   Verifique se a tabela VENDA tem a coluna CODIGOCLIENTE")
		logger.Warn("   Verifique também se o alias A1 está sendo reconhecido no contexto do EXISTS")
		logger.Warn("   Sugestão: Pode ser necessário usar VENDA.CODIGOCLIENTE ao invés de A1.CODIGOCLIENTE")
	}

	// 5. Gera CODIGO_PS para FORMULAVENDA (registros com CONVERSAO)
	logger.Info("Gerando CODIGO_PS para fórmulas de venda...")
	_, err = prisma5DB.Exec(`
		UPDATE FORMULAVENDA A2
		SET A2.CODIGO_PS = GEN_ID(GEN_REFAZER_A2, 1),
			A2.CODIGO_PS_A1 = (
				SELECT FIRST 1 A1.CODIGO_PS
				FROM VENDA A1
				WHERE A1.NUMEROVENDA = A2.NUMEROVENDA
				AND A1.CODIGO_PS IS NOT NULL
			)
		WHERE A2.CONVERSAO = ?
		AND A2.CODIGO_PS IS NULL
		AND EXISTS (
			SELECT FIRST 1 A1.CODIGO_PS
			FROM VENDA A1
			WHERE A1.NUMEROVENDA = A2.NUMEROVENDA
			AND A1.CODIGO_PS IS NOT NULL
		)
	`, conversaoInt)
	if err != nil {
		logger.Warn("Aviso ao gerar CODIGO_PS para fórmulas de venda: %v", err)
	}

	// 6. Gera CODIGO_PS para ITEMFORMULAVENDA (registros com CONVERSAO)
	logger.Info("Gerando CODIGO_PS para itens de fórmula de venda...")
	_, err = prisma5DB.Exec(`
		UPDATE ITEMFORMULAVENDA A3
		SET A3.CODIGO_PS = GEN_ID(GEN_REFAZER_A3, 1),
			A3.CODIGO_PS_A1 = (
				SELECT FIRST 1 A1.CODIGO_PS
				FROM VENDA A1
				WHERE A1.NUMEROVENDA = A3.NUMEROVENDA
				AND A1.CODIGO_PS IS NOT NULL
			)
		WHERE A3.CONVERSAO = ?
		AND A3.CODIGO_PS IS NULL
		AND EXISTS (
			SELECT FIRST 1 A1.CODIGO_PS
			FROM VENDA A1
			WHERE A1.NUMEROVENDA = A3.NUMEROVENDA
			AND A1.CODIGO_PS IS NOT NULL
		)
		AND EXISTS (
			SELECT FIRST 1 P.CODIGO_PS
			FROM PRODUTO P
			WHERE P.CODIGOPRODUTO = A3.CODIGOPRODUTO 
			AND P.CODIGOGRUPO = A3.CODIGOGRUPO 
			AND P.CODIGO_PS < 4000000
		)
		ORDER BY A3.NUMEROVENDA ASC
	`, conversaoInt)
	if err != nil {
		logger.Warn("Aviso ao gerar CODIGO_PS para itens de fórmula de venda: %v", err)
	}

	// 7. Insere vendas no Pharmacie (ATENDIMENTO_REF_A1)
	logger.Info("Inserindo vendas no Pharmacie...")
	vendasQuery := `
		SELECT
			A1.CODIGO_PS AS CODIGO,
			A1.DATAEMISSAOVENDA AS CADASTRO_DT,
			C.CODIGO_PS AS CODIGO_CLIENTE,
			'### VALOR BRUTO: R$ ' || 
				CAST(COALESCE(A1.VALORBRUTOVENDA, 0) AS NUMERIC(18,2)) ||
			' | DESCONTO: R$ ' || 
				CAST(COALESCE(A1.VALORDESCONTOVENDA, 0) AS NUMERIC(18,2)) ||
			' | VALOR LIQUIDO: R$ ' || 
				CAST(COALESCE(A1.VALORLIQUIDOVENDA, 0) AS NUMERIC(18,2)) ||
			' ###' AS OBSERVACAO,
			COALESCE(A1.CODIGO_MEDICO, 9999999) AS CODIGO_MEDICO
		FROM VENDA A1
		INNER JOIN CLIENTE C ON C.CODIGOCLIENTE = A1.CODIGOCLIENTE
		WHERE A1.CODIGO_PS IS NOT NULL 
		AND A1.CONVERSAO = ?
		ORDER BY 1
	`

	rows, err := prisma5DB.Query(vendasQuery, conversaoInt)
	if err != nil {
		return fmt.Errorf("erro ao buscar vendas: %w", err)
	}
	defer rows.Close()

	insertVendaSQL := `INSERT INTO ATENDIMENTO_REF_A1 (CODIGO, CADASTRO_DT, CODIGO_CLIENTE, OBSERVACAO, CODIGO_MEDICO) VALUES (?, ?, ?, ?, ?)`

	tx, err := pharmacieDB.Begin()
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação para vendas: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(insertVendaSQL)
	if err != nil {
		return fmt.Errorf("erro ao preparar statement de vendas: %w", err)
	}
	defer stmt.Close()

	contadorVendas := 0
	for rows.Next() {
		var codigo, codigoCliente, codigoMedico sql.NullInt64
		var observacao sql.NullString
		var cadastroDT sql.NullTime

		err := rows.Scan(&codigo, &cadastroDT, &codigoCliente, &observacao, &codigoMedico)
		if err != nil {
			logger.Warn("Erro ao escanear venda: %v", err)
			continue
		}

		if !codigo.Valid || !codigoCliente.Valid {
			continue
		}

		_, err = stmt.Exec(codigo, cadastroDT, codigoCliente, observacao, codigoMedico)
		if err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "unique") && !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				logger.Warn("Erro ao inserir venda %d: %v", codigo.Int64, err)
			}
			continue
		}
		contadorVendas++
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("erro ao commitar transação de vendas: %w", err)
	}
	logger.Info("✅ %d vendas inseridas no Pharmacie", contadorVendas)

	// 8. Insere fórmulas no Pharmacie (ATENDIMENTO_REF_A2)
	logger.Info("Inserindo fórmulas de venda no Pharmacie...")
	formulasQuery := `
		SELECT
			A2.CODIGO_PS AS CODIGO,
			A2.CODIGO_PS_A1 AS CODIGO_REF_A1,
			A2.NUMEROFORMULA AS NUMEROFORMULA,
			FF.CODIGO_PS AS CODIGO_FORMAFARMACEUTICA,
			CAST(
				CASE
					WHEN A2.QUANTIDADEFORMULA IS NOT NULL THEN A2.QUANTIDADEFORMULA
					WHEN A2.QUANTIDADECAPSULASFORMULA IS NOT NULL THEN A2.QUANTIDADECAPSULASFORMULA
					ELSE A2.VOLUMEBASEFORMULA
				END 
				AS NUMERIC(18,4)) AS QVD,
			TRIM(
				COALESCE(
					'### VALOR FÓRMULA: R$ ' || 
					CAST(A2.VALORLIQUIDOFORMULA AS NUMERIC(18,2)) || ' ###', ''
				)
				|| 
				CASE 
					WHEN A2.OBSERVACAOFORMULA IS NOT NULL AND A2.OBSERVACAOFORMULA <> '' 
					THEN ASCII_CHAR(10) || CAST(A2.OBSERVACAOFORMULA AS VARCHAR(7000))
					ELSE '' 
				END
			) AS OBSERVACAOFORMULA,
			COALESCE(UPPER(PS.DESCRICAOPOSOLOGIA), '') AS POSOLOGIA,
			A2.ETIQUETAFORMULA AS TEXTOROTULO_RAW
		FROM FORMULAVENDA A2
		INNER JOIN POSOLOGIA PS ON PS.CODIGOPOSOLOGIA = A2.CODIGOPOSOLOGIA
		INNER JOIN FORMAFARMACEUTICA FF ON FF.CODIGOFORMA = A2.CODIGOFORMA
		WHERE A2.CODIGO_PS IS NOT NULL 
		AND A2.CONVERSAO = ?
	`

	rows, err = prisma5DB.Query(formulasQuery, conversaoInt)
	if err != nil {
		return fmt.Errorf("erro ao buscar fórmulas de venda: %w", err)
	}
	defer rows.Close()

	insertFormulaSQL := `INSERT INTO ATENDIMENTO_REF_A2 (CODIGO, CODIGO_REF_A1, NUMEROFORMULA, CODIGO_FORMAFARMACEUTICA, QVD, OBSERVACAOFORMULA, POSOLOGIA, TEXTOROTULO) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	tx, err = pharmacieDB.Begin()
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação para fórmulas de venda: %w", err)
	}
	defer tx.Rollback()

	stmt, err = tx.Prepare(insertFormulaSQL)
	if err != nil {
		return fmt.Errorf("erro ao preparar statement de fórmulas de venda: %w", err)
	}
	defer stmt.Close()

	contadorFormulas := 0
	for rows.Next() {
		var codigo, codigoRefA1, numeroFormula, codigoFormaFarmaceutica sql.NullInt64
		var qvd sql.NullFloat64
		var observacaoFormula, posologia, textoRotuloRaw sql.NullString

		err := rows.Scan(&codigo, &codigoRefA1, &numeroFormula, &codigoFormaFarmaceutica, &qvd, &observacaoFormula, &posologia, &textoRotuloRaw)
		if err != nil {
			logger.Warn("Erro ao escanear fórmula de venda: %v", err)
			continue
		}

		if !codigo.Valid || !codigoRefA1.Valid {
			continue
		}

		// Processar texto do rótulo usando função Go
		var textoRotulo sql.NullString
		if textoRotuloRaw.Valid && textoRotuloRaw.String != "" {
			textoLimpo := utilLimpaRTF(textoRotuloRaw.String)
			textoRotulo = sql.NullString{String: textoLimpo, Valid: true}
		}

		_, err = stmt.Exec(codigo, codigoRefA1, numeroFormula, codigoFormaFarmaceutica, qvd, observacaoFormula, posologia, textoRotulo)
		if err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "unique") && !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				logger.Warn("Erro ao inserir fórmula de venda %d: %v", codigo.Int64, err)
			}
			continue
		}
		contadorFormulas++
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("erro ao commitar transação de fórmulas de venda: %w", err)
	}
	logger.Info("✅ %d fórmulas de venda inseridas no Pharmacie", contadorFormulas)

	// 9. Insere itens no Pharmacie (ATENDIMENTO_REF_A3)
	logger.Info("Inserindo itens de fórmula de venda no Pharmacie...")
	itensQuery := `
		SELECT
			A3.CODIGO_PS AS CODIGO,
			A3.CODIGO_PS_A1 AS CODIGO_REF_A1,
			P.CODIGO_PS AS CODIGO_PRODUTO,
			CAST(COALESCE(A3.QUANTIDADE, A3.DINAMIZACAO) AS NUMERIC(18,4)) AS QUANTIDADE,
			LOWER(
            IIF(
                A3.CALCULOITEM = 0, '%',
                IIF(A3.CALCULOITEM = 2, 'qsp(g)',
                    CASE
                        WHEN P.SIGLAUNIDADE = 'UN' THEN 'u'
                        WHEN A3.SIGLAUNIDADE IS NULL THEN COALESCE(A3.METODO, P.SIGLAUNIDADE)
                        ELSE  A3.SIGLAUNIDADE
                    END
                )
            )
        	) AS UNIDADE,
			A3.NUMEROFORMULA AS NUMEROFORMULA,
			0 AS INCLUSAOSISTEMA
		FROM ITEMFORMULAVENDA A3
		INNER JOIN PRODUTO P ON P.CODIGOPRODUTO = A3.CODIGOPRODUTO AND P.CODIGOGRUPO = A3.CODIGOGRUPO
		WHERE A3.CODIGO_PS IS NOT NULL 
		AND A3.CONVERSAO = ?
		ORDER BY 1
	`

	rows, err = prisma5DB.Query(itensQuery, conversaoInt)
	if err != nil {
		return fmt.Errorf("erro ao buscar itens de fórmula de venda: %w", err)
	}
	defer rows.Close()

	insertItemSQL := `INSERT INTO ATENDIMENTO_REF_A3 (CODIGO, CODIGO_REF_A1, CODIGO_PRODUTO, QUANTIDADE, UNIDADE, NUMEROFORMULA, INCLUSAOSISTEMA) VALUES (?, ?, ?, ?, ?, ?, ?)`

	tx, err = pharmacieDB.Begin()
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação para itens de fórmula de venda: %w", err)
	}
	defer tx.Rollback()

	stmt, err = tx.Prepare(insertItemSQL)
	if err != nil {
		return fmt.Errorf("erro ao preparar statement de itens de fórmula de venda: %w", err)
	}
	defer stmt.Close()

	contadorItens := 0
	for rows.Next() {
		var codigo, codigoRefA1, codigoProduto, numeroFormula, inclusaoSistema sql.NullInt64
		var quantidade sql.NullFloat64
		var unidade sql.NullString

		err := rows.Scan(&codigo, &codigoRefA1, &codigoProduto, &quantidade, &unidade, &numeroFormula, &inclusaoSistema)
		if err != nil {
			logger.Warn("Erro ao escanear item de fórmula de venda: %v", err)
			continue
		}

		if !codigo.Valid || !codigoRefA1.Valid || !codigoProduto.Valid {
			continue
		}

		_, err = stmt.Exec(codigo, codigoRefA1, codigoProduto, quantidade, unidade, numeroFormula, inclusaoSistema)
		if err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "unique") && !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				logger.Warn("Erro ao inserir item de fórmula de venda %d: %v", codigo.Int64, err)
			}
			continue
		}
		contadorItens++
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("erro ao commitar transação de itens de fórmula de venda: %w", err)
	}
	logger.Info("✅ %d itens de fórmula de venda inseridos no Pharmacie", contadorItens)

	logger.Info("✅ Migração de refazer (vendas) concluída com sucesso!")
	return nil
}
