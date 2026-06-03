package prisma5

import (
	"database/sql"
	"fmt"
	"math"
	"strings"

	"github.com/primesoftwaresi/prime-migration/pkg/logger"
)

// MigratePRISMA5ProducaoInterna migra produção interna do PRISMA5 para o Pharmacie
func MigratePRISMA5ProducaoInterna(prisma5DB *sql.DB) error {
	logger.Info("Iniciando migração de produção interna")

	// Conectar ao Pharmacie
	pharmacieDB, err := connectToPharmacie(prisma5DB)
	if err != nil {
		return fmt.Errorf("erro ao conectar ao Pharmacie: %w", err)
	}
	defer pharmacieDB.Close()

	// 1. Atualiza CODIGO_PRODUTO_PS em FORMULAPADRAO
	logger.Info("Atualizando CODIGO_PRODUTO_PS em FORMULAPADRAO...")
	_, err = prisma5DB.Exec(`
		UPDATE FORMULAPADRAO F
		SET F.CODIGO_PRODUTO_PS = (
			SELECT FIRST 1 P.CODIGO_PS 
			FROM PRODUTO P 
			WHERE P.CODIGOPRODUTO = F.CODIGOPRODUTO 
			AND P.CODIGOGRUPO = F.CODIGOGRUPO
		)
		WHERE F.CODIGO_PRODUTO_PS IS NULL
	`)
	if err != nil {
		logger.Warn("Aviso ao atualizar CODIGO_PRODUTO_PS em FORMULAPADRAO: %v", err)
	}

	// 2. Define OPCAO_ATENDIMENTO_PS = 2 para fórmulas padrão tipo 2
	logger.Info("Definindo OPCAO_ATENDIMENTO_PS = 2 para fórmulas tipo 2...")
	_, err = prisma5DB.Exec(`
		UPDATE FORMULAPADRAO F
		SET F.OPCAO_ATENDIMENTO_PS = 2
		WHERE F.CODIGO_PRODUTO_PS IS NULL
		AND F.TIPOFORMULAPADRAO = 2
	`)
	if err != nil {
		logger.Warn("Aviso ao definir OPCAO_ATENDIMENTO_PS: %v", err)
	}

	// 3. Gera CODIGO_PRODUTO_PS para fórmulas com OPCAO_ATENDIMENTO_PS = 2
	logger.Info("Gerando CODIGO_PRODUTO_PS para fórmulas de produção interna...")
	_, err = prisma5DB.Exec(`
		UPDATE FORMULAPADRAO F
		SET F.CODIGO_PRODUTO_PS = GEN_ID(GEN_ESTOQUE_GERAL1, 1) + 1000000
		WHERE F.OPCAO_ATENDIMENTO_PS = 2
	`)
	if err != nil {
		logger.Warn("Aviso ao gerar CODIGO_PRODUTO_PS para produção interna: %v", err)
	}

	// 4. Gera CODIGO_ESTOQUEF para todas as fórmulas com CODIGO_PRODUTO_PS não nulo
	logger.Info("Gerando CODIGO_ESTOQUEF para fórmulas...")
	_, err = prisma5DB.Exec(`
		UPDATE FORMULAPADRAO F
		SET F.CODIGO_ESTOQUEF = GEN_ID(GEN_ESTOQUEF, 1)
		WHERE F.CODIGO_PRODUTO_PS IS NOT NULL
	`)
	if err != nil {
		logger.Warn("Aviso ao gerar CODIGO_ESTOQUEF: %v", err)
	}

	// 5. Insere produtos de produção interna em ESTOQUE_GERAL
	logger.Info("Inserindo produtos de produção interna no Pharmacie...")
	produtosQuery := `
		SELECT
			F.CODIGO_PRODUTO_PS AS CODIGO,
			IIF(F.CODIGO_PRODUTO_PS < 1999999, 1, 3) AS MOVIMENTOESTOQUE,
			IIF(F.CODIGO_PRODUTO_PS < 1999999, 100000, 300000) AS CODIGO_GRUPO,
			IIF(F.CODIGO_PRODUTO_PS < 1999999, 100000, 300000) AS CODIGO_ESTOQUE_CLASSIFICACAO,
			IIF(F.CODIGO_PRODUTO_PS < 1999999, 100000, 300000) AS CODIGO_ESTOQUE_TIPO,
			UPPER(F.DESCRICAOFORMULAPADRAO) || ' (PRODUCAO INTERNA)' AS NOMEPRODUTO,
			UPPER(F.DESCRICAOFORMULAPADRAO) || ' (PRODUCAO INTERNA)' AS NOMEPRODUTO_FRACAO,
			UPPER(F.DESCRICAOFORMULAPADRAO) AS NOMEPRODUTO_ROTULO,
			-1 AS ATIVO,
			100 AS CONCENTRACAO,
			1 AS DENSIDADE,
			1 AS INDICELUCROCVM,
			-1 AS USO_ATENDIMENTO,
			1 AS CADASTRO_LJ,
			1 AS CADASTRO_CF,
			CURRENT_TIMESTAMP AS CADASTRO_DT,
			1 AS ALTERACAO_LJ,
			1 AS ALTERACAO_CF,
			CURRENT_TIMESTAMP AS ALTERACAO_DT
		FROM FORMULAPADRAO F
		WHERE F.OPCAO_ATENDIMENTO_PS = 2
	`

	rows, err := prisma5DB.Query(produtosQuery)
	if err != nil {
		return fmt.Errorf("erro ao buscar produtos de produção interna: %w", err)
	}
	defer rows.Close()

	insertProdutoSQL := `INSERT INTO ESTOQUE_GERAL (CODIGO, MOVIMENTOESTOQUE, CODIGO_GRUPO, CODIGO_ESTOQUE_CLASSIFICACAO, CODIGO_ESTOQUE_TIPO, NOMEPRODUTO, NOMEPRODUTO_FRACAO, NOMEPRODUTO_ROTULO, ATIVO, CONCENTRACAO, DENSIDADE, INDICELUCROCVM, USO_ATENDIMENTO, CADASTRO_LJ, CADASTRO_CF, CADASTRO_DT, ALTERACAO_LJ, ALTERACAO_CF, ALTERACAO_DT) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	tx, err := pharmacieDB.Begin()
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação para produtos de produção interna: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(insertProdutoSQL)
	if err != nil {
		return fmt.Errorf("erro ao preparar statement de produtos de produção interna: %w", err)
	}
	defer stmt.Close()

	contadorProdutos := 0
	for rows.Next() {
		var codigo, movimentoEstoque, codigoGrupo, codigoEstoqueClassificacao, codigoEstoqueTipo, ativo, concentracao, densidade, indicelucrocvm, usoAtendimento, cadastroLJ, cadastroCF, alteracaoLJ, alteracaoCF sql.NullInt64
		var nomeProduto, nomeProdutoFracao, nomeProdutoRotulo sql.NullString
		var cadastroDT, alteracaoDT sql.NullTime

		err := rows.Scan(&codigo, &movimentoEstoque, &codigoGrupo, &codigoEstoqueClassificacao, &codigoEstoqueTipo, &nomeProduto, &nomeProdutoFracao, &nomeProdutoRotulo, &ativo, &concentracao, &densidade, &indicelucrocvm, &usoAtendimento, &cadastroLJ, &cadastroCF, &cadastroDT, &alteracaoLJ, &alteracaoCF, &alteracaoDT)
		if err != nil {
			logger.Warn("Erro ao escanear produto de produção interna: %v", err)
			continue
		}

		if !codigo.Valid {
			continue
		}

		_, err = stmt.Exec(codigo, movimentoEstoque, codigoGrupo, codigoEstoqueClassificacao, codigoEstoqueTipo, nomeProduto, nomeProdutoFracao, nomeProdutoRotulo, ativo, concentracao, densidade, indicelucrocvm, usoAtendimento, cadastroLJ, cadastroCF, cadastroDT, alteracaoLJ, alteracaoCF, alteracaoDT)
		if err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "unique") && !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				logger.Warn("Erro ao inserir produto de produção interna %d: %v", codigo.Int64, err)
			}
			continue
		}
		contadorProdutos++
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("erro ao commitar transação de produtos de produção interna: %w", err)
	}
	logger.Info("✅ %d produtos de produção interna inseridos", contadorProdutos)

	// 6. Insere fórmulas em ESTOQUEF
	logger.Info("Inserindo fórmulas no Pharmacie...")
	formulasQuery := `
		SELECT
			F.CODIGO_ESTOQUEF AS CODIGO,
			CASE
				WHEN F.CODIGO_PRODUTO_PS < 1000000 THEN 0
				WHEN F.CODIGO_PRODUTO_PS > 1000000 AND F.CODIGO_PRODUTO_PS < 1999999 THEN 1
				WHEN F.CODIGO_PRODUTO_PS > 2000000 AND F.CODIGO_PRODUTO_PS < 2999999 THEN 2
				WHEN F.CODIGO_PRODUTO_PS > 3000000 AND F.CODIGO_PRODUTO_PS < 3999999 THEN 3
				WHEN F.CODIGO_PRODUTO_PS > 4000000 AND F.CODIGO_PRODUTO_PS < 4999999 THEN 4
			END AS MOVIMENTOESTOQUE,
			F.CODIGO_PRODUTO_PS AS CODIGO_FORMULA,
			FF.CODIGO_PS AS CODIGO_FORMAFARMACEUTICA,
			UPPER(
				CASE
					WHEN COALESCE(TRIM(F.OBSERVACAOFORMULAPADRAO), '') <> '' AND COALESCE(TRIM(F.OBSERVACAOETIQUETA), '') <> ''
					THEN CAST(F.OBSERVACAOFORMULAPADRAO AS VARCHAR(7000)) || (ASCII_CHAR(13) || ASCII_CHAR(10)) || CAST(F.OBSERVACAOETIQUETA AS VARCHAR(7000))
					WHEN COALESCE(TRIM(F.OBSERVACAOFORMULAPADRAO), '') <> ''
					THEN CAST(F.OBSERVACAOFORMULAPADRAO AS VARCHAR(7000))
					WHEN COALESCE(TRIM(F.OBSERVACAOETIQUETA), '') <> ''
					THEN CAST(F.OBSERVACAOETIQUETA AS VARCHAR(7000))
					ELSE ''
				END
			) AS MODOFAZER,
			COALESCE(F.OPCAO_ATENDIMENTO_PS, 0) AS OPCAO_ATENDIMENTO,
			0 AS PRECOSISTEMA,
			0 AS OPCAO_ROTULO,
			CASE
				WHEN F.CODIGO_PRODUTO_PS > 1000000 AND F.CODIGO_PRODUTO_PS < 1999999 THEN 1
				WHEN F.CODIGO_PRODUTO_PS > 3000000 AND F.CODIGO_PRODUTO_PS < 3999999 THEN 2
				ELSE 1
			END AS TIPOCADASTRO,
			COALESCE(F.VOLUMEBASEFORMULAPADRAO, 100) AS PESOFORMULA,
			1 AS INDICECALCULO,
			-1 AS CONCLUIDO,
			1 AS CADASTRO_LJ,
			1 AS CADASTRO_CF,
			CURRENT_TIMESTAMP AS CADASTRO_DT,
			1 AS ALTERACAO_LJ,
			1 AS ALTERACAO_CF,
			CURRENT_TIMESTAMP AS ALTERACAO_DT
		FROM FORMULAPADRAO F
		INNER JOIN FORMAFARMACEUTICA FF ON FF.CODIGOFORMA = F.CODIGOFORMA
		WHERE F.CODIGO_ESTOQUEF IS NOT NULL
	`

	rows, err = prisma5DB.Query(formulasQuery)
	if err != nil {
		return fmt.Errorf("erro ao buscar fórmulas: %w", err)
	}
	defer rows.Close()

	insertFormulaSQL := `INSERT INTO ESTOQUEF (CODIGO, MOVIMENTOESTOQUE, CODIGO_FORMULA, CODIGO_FORMAFARMACEUTICA, MODOFAZER, OPCAO_ATENDIMENTO, PRECOSISTEMA, OPCAO_ROTULO, TIPOCADASTRO, PESOFORMULA, INDICECALCULO, CONCLUIDO, CADASTRO_LJ, CADASTRO_CF, CADASTRO_DT, ALTERACAO_LJ, ALTERACAO_CF, ALTERACAO_DT) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	tx, err = pharmacieDB.Begin()
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação para fórmulas: %w", err)
	}
	defer tx.Rollback()

	stmt, err = tx.Prepare(insertFormulaSQL)
	if err != nil {
		return fmt.Errorf("erro ao preparar statement de fórmulas: %w", err)
	}
	defer stmt.Close()

	contadorFormulas := 0
	for rows.Next() {
		var codigo, movimentoEstoque, codigoFormula, codigoFormaFarmaceutica, opcaoAtendimento, precoSistema, opcaoRotulo, tipoCadastro, indiceCalculo, concluido, cadastroLJ, cadastroCF, alteracaoLJ, alteracaoCF sql.NullInt64
		var pesoFormula sql.NullFloat64 // PESOFORMULA pode ser decimal
		var modoFazer sql.NullString
		var cadastroDT, alteracaoDT sql.NullTime

		err := rows.Scan(&codigo, &movimentoEstoque, &codigoFormula, &codigoFormaFarmaceutica, &modoFazer, &opcaoAtendimento, &precoSistema, &opcaoRotulo, &tipoCadastro, &pesoFormula, &indiceCalculo, &concluido, &cadastroLJ, &cadastroCF, &cadastroDT, &alteracaoLJ, &alteracaoCF, &alteracaoDT)
		if err != nil {
			logger.Warn("Erro ao escanear fórmula: %v", err)
			continue
		}

		if !codigo.Valid {
			continue
		}

		// Converter pesoFormula de float64 para int64 (arredondando se necessário)
		var pesoFormulaInt sql.NullInt64
		if pesoFormula.Valid {
			pesoFormulaInt.Int64 = int64(math.Round(pesoFormula.Float64)) // Arredonda para inteiro
			pesoFormulaInt.Valid = true
		}

		_, err = stmt.Exec(codigo, movimentoEstoque, codigoFormula, codigoFormaFarmaceutica, modoFazer, opcaoAtendimento, precoSistema, opcaoRotulo, tipoCadastro, pesoFormulaInt, indiceCalculo, concluido, cadastroLJ, cadastroCF, cadastroDT, alteracaoLJ, alteracaoCF, alteracaoDT)
		if err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "unique") && !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				logger.Warn("Erro ao inserir fórmula %d: %v", codigo.Int64, err)
			}
			continue
		}
		contadorFormulas++
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("erro ao commitar transação de fórmulas: %w", err)
	}
	logger.Info("✅ %d fórmulas inseridas no Pharmacie", contadorFormulas)

	// 7. Atualiza CODIGO_ESTOQUEF em ITEMFORMULAPADRAO
	logger.Info("Atualizando CODIGO_ESTOQUEF em ITEMFORMULAPADRAO...")
	_, err = prisma5DB.Exec(`
		UPDATE ITEMFORMULAPADRAO FI
		SET FI.CODIGO_ESTOQUEF = (
			SELECT FIRST 1 F.CODIGO_ESTOQUEF 
			FROM FORMULAPADRAO F 
			WHERE F.CODIGOFORMULAPADRAO = FI.CODIGOFORMULAPADRAO
		)
	`)
	if err != nil {
		logger.Warn("Aviso ao atualizar CODIGO_ESTOQUEF em ITEMFORMULAPADRAO: %v", err)
	}

	// 8. Insere itens da fórmula em ESTOQUEF_DETALHE (com 3 queries UNION: itens da fórmula, embalagem, cápsula)
	logger.Info("Inserindo itens de fórmula no Pharmacie...")
	itensQuery := `
		-- Query 1: Itens da Formula
		SELECT
			GEN_ID(GEN_ESTOQUEF_DETALHE, 1) AS CODIGO,
			FI.CODIGO_ESTOQUEF AS CODIGO_ESTOQUEF,
			FI.FASEITEMFORMULA AS FASE,
			P.CODIGO_PS AS CODIGO_PRODUTO,
			CAST(
				IIF(
					FI.CALCULOITEMFORMULA = 2, COALESCE(FP.VOLUMEBASEFORMULAPADRAO, 100), FI.QUANTIDADEITEMFORMULA
				) AS NUMERIC(18,4)
			) AS QUANTIDADE,
			LOWER(
				IIF(
					FI.CALCULOITEMFORMULA = 0, '%',
					IIF(FI.CALCULOITEMFORMULA = 2, 'qsp(g)',
						CASE
							WHEN P.SIGLAUNIDADE = 'UN' THEN 'u'
							WHEN FI.SIGLAUNIDADE IS NULL THEN P.SIGLAUNIDADE
							ELSE FI.SIGLAUNIDADE
						END
					)
				)
			) AS UNIDADE,
			1 AS CADASTRO_LJ,
			1 AS CADASTRO_CF,
			FI.DT_CREATION AS CADASTRO_DT,
			1 AS ALTERACAO_LJ,
			1 AS ALTERACAO_CF,
			FI.DT_ALTER AS ALTERACAO_DT
		FROM ITEMFORMULAPADRAO FI
		INNER JOIN PRODUTO P
			ON P.CODIGOPRODUTO = FI.CODIGOPRODUTO
			AND P.CODIGOGRUPO = FI.CODIGOGRUPO
		INNER JOIN FORMULAPADRAO FP
			ON FP.CODIGOFORMULAPADRAO = FI.CODIGOFORMULAPADRAO
		WHERE FI.CODIGO_ESTOQUEF IS NOT NULL
		AND P.CODIGO_PS IS NOT NULL

		UNION ALL

		-- Query 2: Embalagem
		SELECT
			GEN_ID(GEN_ESTOQUEF_DETALHE, 1) AS CODIGO,
			FP.CODIGO_ESTOQUEF AS CODIGO_ESTOQUEF,
			NULL AS FASE,
			P.CODIGO_PS AS CODIGO_PRODUTO,
			FP.QTDEEMBALAGEMFORMULAPADRAO AS QUANTIDADE,
			'u' AS UNIDADE,
			1 AS CADASTRO_LJ,
			1 AS CADASTRO_CF,
			FP.DT_CREATION AS CADASTRO_DT,
			1 AS ALTERACAO_LJ,
			1 AS ALTERACAO_CF,
			FP.DT_ALTER AS ALTERACAO_DT
		FROM FORMULAPADRAO FP
		INNER JOIN PRODUTO P ON P.CODIGOPRODUTO = FP.CODIGOPRODUTOEMBALAGEM AND P.CODIGOGRUPO = FP.CODIGOGRUPOEMBALAGEM
		WHERE FP.CODIGO_ESTOQUEF IS NOT NULL
			AND P.CODIGO_PS IS NOT NULL

		UNION ALL

		-- Query 3: Capsula
		SELECT
			GEN_ID(GEN_ESTOQUEF_DETALHE, 1) AS CODIGO,
			FP.CODIGO_ESTOQUEF AS CODIGO_ESTOQUEF,
			NULL AS FASE,
			P.CODIGO_PS AS CODIGO_PRODUTO,
			FP.CAPSULASBASEFORMULAPADRAO AS QUANTIDADE,
			'u' AS UNIDADE,
			1 AS CADASTRO_LJ,
			1 AS CADASTRO_CF,
			FP.DT_CREATION AS CADASTRO_DT,
			1 AS ALTERACAO_LJ,
			1 AS ALTERACAO_CF,
			FP.DT_ALTER AS ALTERACAO_DT
		FROM FORMULAPADRAO FP
		INNER JOIN PRODUTO P ON P.CODIGOPRODUTO = FP.CODIGOPRODUTOCAPSULA AND P.CODIGOGRUPO = FP.CODIGOGRUPOCAPSULA
		WHERE FP.CODIGO_ESTOQUEF IS NOT NULL
			AND P.CODIGO_PS IS NOT NULL
	`

	rows, err = prisma5DB.Query(itensQuery)
	if err != nil {
		return fmt.Errorf("erro ao buscar itens de fórmula: %w", err)
	}
	defer rows.Close()

	insertItemSQL := `INSERT INTO ESTOQUEF_DETALHE (CODIGO, CODIGO_ESTOQUEF, FASE, CODIGO_PRODUTO, QUANTIDADE, UNIDADE, CADASTRO_LJ, CADASTRO_CF, CADASTRO_DT, ALTERACAO_LJ, ALTERACAO_CF, ALTERACAO_DT) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	tx, err = pharmacieDB.Begin()
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação para itens de fórmula: %w", err)
	}
	defer tx.Rollback()

	stmt, err = tx.Prepare(insertItemSQL)
	if err != nil {
		return fmt.Errorf("erro ao preparar statement de itens de fórmula: %w", err)
	}
	defer stmt.Close()

	contadorItens := 0
	for rows.Next() {
		var codigo, codigoEstoqueF, fase, codigoProduto, cadastroLJ, cadastroCF, alteracaoLJ, alteracaoCF sql.NullInt64
		var quantidade sql.NullFloat64
		var unidade sql.NullString
		var cadastroDT, alteracaoDT sql.NullTime

		err := rows.Scan(&codigo, &codigoEstoqueF, &fase, &codigoProduto, &quantidade, &unidade, &cadastroLJ, &cadastroCF, &cadastroDT, &alteracaoLJ, &alteracaoCF, &alteracaoDT)
		if err != nil {
			logger.Warn("Erro ao escanear item de fórmula: %v", err)
			continue
		}

		if !codigo.Valid || !codigoEstoqueF.Valid {
			continue
		}

		// Validar campos obrigatórios antes de inserir
		if !codigoProduto.Valid {
			logger.Warn("Código do produto inválido para item de fórmula %d, pulando...", codigo.Int64)
			continue
		}

		_, err = stmt.Exec(codigo, codigoEstoqueF, fase, codigoProduto, quantidade, unidade, cadastroLJ, cadastroCF, cadastroDT, alteracaoLJ, alteracaoCF, alteracaoDT)
		if err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "unique") && !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				logger.Warn("Erro ao inserir item de fórmula %d: %v", codigo.Int64, err)
			}
			continue
		}
		contadorItens++
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("erro ao commitar transação de itens de fórmula: %w", err)
	}
	logger.Info("✅ %d itens de fórmula inseridos no Pharmacie", contadorItens)

	logger.Info("✅ Migração de produção interna concluída com sucesso!")
	return nil
}

