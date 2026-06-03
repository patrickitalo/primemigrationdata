package prisma5

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/primesoftwaresi/prime-migration/pkg/logger"
)

// MigratePRISMA5Produtos migra produtos do PRISMA5 para o Pharmacie
// Parâmetro conversao: valor da conversão (ex: "1", "2")
func MigratePRISMA5Produtos(prisma5DB *sql.DB, conversao string) error {
	logger.Info("Iniciando migração de produtos (conversão: %s)", conversao)

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

	// 1. Atualiza campo CONVERSAO dos produtos
	logger.Info("Atualizando campo CONVERSAO dos produtos...")
	_, err = prisma5DB.Exec(`
		UPDATE PRODUTO P 
		SET P.CONVERSAO = ? 
		WHERE P.CODIGO_PS IS NULL
	`, conversaoInt)
	if err != nil {
		logger.Warn("Aviso ao atualizar CONVERSAO de produtos: %v", err)
	}

	// 2. Atualiza campo CONVERSAO dos sinônimos
	logger.Info("Atualizando campo CONVERSAO dos sinônimos...")
	_, err = prisma5DB.Exec(`
		UPDATE SINONIMO S 
		SET S.CONVERSAO = ? 
		WHERE S.CODIGO_PRODUTO_PS IS NULL
	`, conversaoInt)
	if err != nil {
		logger.Warn("Aviso ao atualizar CONVERSAO de sinônimos: %v", err)
	}

	// 3. Gera CODIGO_PS para produtos baseado no MOVIMENTOESTOQUE do GRUPO_PHARMACIE
	logger.Info("Gerando CODIGO_PS para produtos...")

	// MOVIMENTOESTOQUE = 0 (GEN_ESTOQUE_GERAL0 - sem offset)
	_, err = prisma5DB.Exec(`
		UPDATE PRODUTO P
		SET P.CODIGO_PS = GEN_ID(GEN_ESTOQUE_GERAL0, 1)
		WHERE P.CODIGO_PS IS NULL
		AND P.CONVERSAO = ?
		AND EXISTS (
			SELECT 1 FROM GRUPO_PHARMACIE G
			WHERE G.CODIGOGRUPO = P.CODIGOGRUPO
			AND G.MOVIMENTOESTOQUE = 0
		)
	`, conversaoInt)
	if err != nil {
		logger.Warn("Aviso ao gerar CODIGO_PS para MOVIMENTOESTOQUE=0: %v", err)
	}

	// MOVIMENTOESTOQUE = 1 (GEN_ESTOQUE_GERAL1 + 1000000)
	_, err = prisma5DB.Exec(`
		UPDATE PRODUTO P
		SET P.CODIGO_PS = GEN_ID(GEN_ESTOQUE_GERAL1, 1) + 1000000
		WHERE P.CODIGO_PS IS NULL
		AND P.CONVERSAO = ?
		AND EXISTS (
			SELECT 1 FROM GRUPO_PHARMACIE G
			WHERE G.CODIGOGRUPO = P.CODIGOGRUPO
			AND G.MOVIMENTOESTOQUE = 1
		)
	`, conversaoInt)
	if err != nil {
		logger.Warn("Aviso ao gerar CODIGO_PS para MOVIMENTOESTOQUE=1: %v", err)
	}

	// MOVIMENTOESTOQUE = 2 (GEN_ESTOQUE_GERAL2 + 2000000)
	_, err = prisma5DB.Exec(`
		UPDATE PRODUTO P
		SET P.CODIGO_PS = GEN_ID(GEN_ESTOQUE_GERAL2, 1) + 2000000
		WHERE P.CODIGO_PS IS NULL
		AND P.CONVERSAO = ?
		AND EXISTS (
			SELECT 1 FROM GRUPO_PHARMACIE G
			WHERE G.CODIGOGRUPO = P.CODIGOGRUPO
			AND G.MOVIMENTOESTOQUE = 2
		)
	`, conversaoInt)
	if err != nil {
		logger.Warn("Aviso ao gerar CODIGO_PS para MOVIMENTOESTOQUE=2: %v", err)
	}

	// MOVIMENTOESTOQUE = 3 (GEN_ESTOQUE_GERAL3 + 3000000)
	_, err = prisma5DB.Exec(`
		UPDATE PRODUTO P
		SET P.CODIGO_PS = GEN_ID(GEN_ESTOQUE_GERAL3, 1) + 3000000
		WHERE P.CODIGO_PS IS NULL
		AND P.CONVERSAO = ?
		AND EXISTS (
			SELECT 1 FROM GRUPO_PHARMACIE G
			WHERE G.CODIGOGRUPO = P.CODIGOGRUPO
			AND G.MOVIMENTOESTOQUE = 3
		)
	`, conversaoInt)
	if err != nil {
		logger.Warn("Aviso ao gerar CODIGO_PS para MOVIMENTOESTOQUE=3: %v", err)
	}

	// MOVIMENTOESTOQUE = 4 (GEN_ESTOQUE_GERAL4 + 4000000)
	_, err = prisma5DB.Exec(`
		UPDATE PRODUTO P
		SET P.CODIGO_PS = GEN_ID(GEN_ESTOQUE_GERAL4, 1) + 4000000
		WHERE P.CODIGO_PS IS NULL
		AND P.CONVERSAO = ?
		AND EXISTS (
			SELECT 1 FROM GRUPO_PHARMACIE G
			WHERE G.CODIGOGRUPO = P.CODIGOGRUPO
			AND G.MOVIMENTOESTOQUE = 4
		)
	`, conversaoInt)
	if err != nil {
		logger.Warn("Aviso ao gerar CODIGO_PS para MOVIMENTOESTOQUE=4: %v", err)
	}

	// 4. Atualiza CODIGO_PRODUTO_PS dos sinônimos
	logger.Info("Atualizando CODIGO_PRODUTO_PS dos sinônimos...")
	_, err = prisma5DB.Exec(`
		UPDATE SINONIMO S
		SET S.CODIGO_PRODUTO_PS = (
			SELECT P.CODIGO_PS FROM PRODUTO P 
			WHERE P.CODIGOPRODUTO = S.CODIGOPRODUTO 
			AND P.CODIGOGRUPO = S.CODIGOGRUPO
		)
		WHERE S.CODIGO_PRODUTO_PS IS NULL 
		AND S.CONVERSAO = ?
	`, conversaoInt)
	if err != nil {
		logger.Warn("Aviso ao atualizar CODIGO_PRODUTO_PS dos sinônimos: %v", err)
	}

	// 5. Atualiza CODIGO_SINONIMO_PS dos sinônimos (apenas para ranges 1 e 2)
	logger.Info("Atualizando CODIGO_SINONIMO_PS dos sinônimos...")
	
	// Para range 1 (1000000-1999999)
	_, err = prisma5DB.Exec(`
		UPDATE SINONIMO S
		SET S.CODIGO_SINONIMO_PS = GEN_ID(GEN_ESTOQUE_GERAL1, 1) + 1000000
		WHERE S.CODIGO_PRODUTO_PS IS NOT NULL
		AND S.CODIGO_PRODUTO_PS > 1000000
		AND S.CODIGO_PRODUTO_PS < 1999999
		AND S.CONVERSAO = ?
	`, conversaoInt)
	if err != nil {
		logger.Warn("Aviso ao atualizar CODIGO_SINONIMO_PS para range 1: %v", err)
	}

	// Para range 2 (2000000-2999999)
	_, err = prisma5DB.Exec(`
		UPDATE SINONIMO S
		SET S.CODIGO_SINONIMO_PS = GEN_ID(GEN_ESTOQUE_GERAL2, 1) + 2000000
		WHERE S.CODIGO_PRODUTO_PS IS NOT NULL
		AND S.CODIGO_PRODUTO_PS > 2000000
		AND S.CODIGO_PRODUTO_PS < 2999999
		AND S.CONVERSAO = ?
	`, conversaoInt)
	if err != nil {
		logger.Warn("Aviso ao atualizar CODIGO_SINONIMO_PS para range 2: %v", err)
	}

	// 6. Insere produtos no banco Pharmacie
	logger.Info("Inserindo produtos no Pharmacie...")
	produtosQuery := `
		SELECT
			P.CODIGO_PS AS CODIGO,
			CASE
				WHEN P.CODIGO_PS < 1000000 THEN 0
				WHEN P.CODIGO_PS > 1000000 AND P.CODIGO_PS < 1999999 THEN 1
				WHEN P.CODIGO_PS > 2000000 AND P.CODIGO_PS < 2999999 THEN 2
				WHEN P.CODIGO_PS > 3000000 AND P.CODIGO_PS < 3999999 THEN 3
				WHEN P.CODIGO_PS > 4000000 AND P.CODIGO_PS < 4999999 THEN 4
			END AS MOVIMENTOESTOQUE,
			CASE
				WHEN P.CODIGO_PS < 1000000 THEN 1
				WHEN P.CODIGO_PS > 1000000 AND P.CODIGO_PS < 1999999 THEN 100000
				WHEN P.CODIGO_PS > 2000000 AND P.CODIGO_PS < 2999999 THEN 200000
				WHEN P.CODIGO_PS > 3000000 AND P.CODIGO_PS < 3999999 THEN 300000
				WHEN P.CODIGO_PS > 4000000 AND P.CODIGO_PS < 4999999 THEN 400000
			END AS CODIGO_GRUPO,
			CASE
				WHEN P.CODIGO_PS < 1000000 THEN 1
				WHEN P.CODIGO_PS > 1000000 AND P.CODIGO_PS < 1999999 THEN 100000
				WHEN P.CODIGO_PS > 2000000 AND P.CODIGO_PS < 2999999 THEN 200000
				WHEN P.CODIGO_PS > 3000000 AND P.CODIGO_PS < 3999999 THEN 300000
				WHEN P.CODIGO_PS > 4000000 AND P.CODIGO_PS < 4999999 THEN 400000
			END AS CODIGO_ESTOQUE_CLASSIFICACAO,
			CASE
				WHEN P.CODIGO_PS < 1000000 THEN 1
				WHEN P.CODIGO_PS > 1000000 AND P.CODIGO_PS < 1999999 THEN 100000
				WHEN P.CODIGO_PS > 2000000 AND P.CODIGO_PS < 2999999 THEN 200000
				WHEN P.CODIGO_PS > 3000000 AND P.CODIGO_PS < 3999999 THEN 300000
				WHEN P.CODIGO_PS > 4000000 AND P.CODIGO_PS < 4999999 THEN 400000
			END AS CODIGO_ESTOQUE_TIPO,
			IIF(P.INATIVOPRODUTO = 1, 'ZZZ-' || P.DESCRICAOPRODUTO, P.DESCRICAOPRODUTO) AS NOMEPRODUTO,
			IIF(
				P.INATIVOPRODUTO = 1,
				'ZZZ-' || IIF(
					P.DESCRICAOROTULOPRODUTO IS NULL OR P.DESCRICAOROTULOPRODUTO = '',
					P.DESCRICAOPRODUTO,
					P.DESCRICAOROTULOPRODUTO
				),
				IIF(
					P.DESCRICAOROTULOPRODUTO IS NULL OR P.DESCRICAOROTULOPRODUTO = '',
					P.DESCRICAOPRODUTO,
					P.DESCRICAOROTULOPRODUTO
				)
			) AS NOMEPRODUTO_ROTULO,
			P.DESCRICAOPRODUTO AS NOMEPRODUTO_FRACAO,
			COALESCE(P.CODIGONCM, NULL) AS CODIGO_NCM,
			COALESCE(P.CODIGODCB, NULL) AS CODIGO_DCB,
			IIF(P.CODIGOBARRAPRODUTO = '' OR P.CODIGOBARRAPRODUTO IS NULL, NULL, P.CODIGOBARRAPRODUTO) AS CODIGOBARRA,
			CASE P.CODIGOLISTACONTROLADO
				WHEN 'A1' THEN 100001
				WHEN 'A2' THEN 100002
				WHEN 'A3' THEN 100003
				WHEN 'AM' THEN 100004
				WHEN 'B1' THEN 100005
				WHEN 'B2' THEN 100006
				WHEN 'C1' THEN 100007
				WHEN 'C2' THEN 100008
				WHEN 'C3' THEN 100009
				WHEN 'C4' THEN 100010
				WHEN 'C5' THEN 100011
				WHEN 'D1' THEN 100012
				WHEN 'D2' THEN 100013
				WHEN 'F1' THEN 100014
				WHEN 'F2' THEN 100015
				WHEN 'F4' THEN 100016
			END AS CODIGO_ESTOQUE_PORTARIA,
			COALESCE(P.FRACIONAMENTOPRODUTO, 1) AS FRACAOVENDA,
			IIF(P.SIGLAUNIDADEVOLUME IS NULL OR P.SIGLAUNIDADEVOLUME = '', P.SIGLAUNIDADEESTOQUE, P.SIGLAUNIDADEVOLUME) AS UNIDADE,
			IIF(P.SIGLAUNIDADE IS NULL OR P.SIGLAUNIDADE = '', P.SIGLAUNIDADEESTOQUE, P.SIGLAUNIDADE) AS UNIDADE_CALCULO,
			CAST(P.VALORCUSTOPRODUTO AS NUMERIC(18,4)) AS VALORCOMPRA,
			CAST(P.CUSTOREFERENCIAPRODUTO AS NUMERIC(18,4)) AS VALORCUSTO,
			CAST(
				CAST(
					IIF(P.CUSTOREFERENCIAPRODUTO > 10000000, 0, P.CUSTOREFERENCIAPRODUTO) 
					AS NUMERIC(18,4)
				) *
				(
					CAST(1.0 AS NUMERIC(18,4)) +
					CAST(COALESCE(P.FATORREFERENCIAPRODUTO, 800) AS NUMERIC(18,4)) /
					CAST(100.0 AS NUMERIC(18,4))
				)
			AS NUMERIC(18,4)) AS VALORVENDA,
			IIF(P.FATORUTRPRODUTO > 0, P.FATORUTRPRODUTO, NULL) AS UTR,
			IIF(P.FATORUIPRODUTO > 0, P.FATORUIPRODUTO, NULL) AS UI,
			IIF(P.DENSIDADEPRODUTO > 0, P.DENSIDADEPRODUTO, 1) AS DENSIDADE,
			IIF(P.FATORCORRECAOPRODUTO > 0, (100/P.FATORCORRECAOPRODUTO), 100) AS CONCENTRACAO,
			COALESCE(P.FATOREQUIVALENCIAPRODUTO, 1) AS FATOREQUIV,
			2 AS FATORCALCULO,
			-P.ESTOQUEMINIMOPRODUTO AS ESTOQUEMINIMO,
			P.DOSEMAXIMAPRODUTO AS QTDEMAX,
			P.DOSEMAXIMAPERCENTUALPRODUTO AS QTDEMAXP,
			IIF(P.DOSEMAXIMAPRODUTO IS NOT NULL OR P.DOSEMAXIMAPRODUTO <> '', -1, 0) AS CHECARMAX,
			IIF(P.CONSERVACAOPRODUTO IS NOT NULL OR P.CONSERVACAOPRODUTO <> '',
				P.CONSERVACAOPRODUTO || ASCII_CHAR(13) || ASCII_CHAR(10) || COALESCE(P.CATEGORIATERAPEUTICAPRODUTO, ''),
				COALESCE(P.CATEGORIATERAPEUTICAPRODUTO, '')) AS CARACTERISTICA,
			IIF(P.OBSERVACAOPRODUTO IS NOT NULL OR P.OBSERVACAOPRODUTO <> '',
				P.OBSERVACAOPRODUTO || ASCII_CHAR(13) || ASCII_CHAR(10) || COALESCE(P.OBSERVACAOFICHATECNICA, '') || ASCII_CHAR(13) || ASCII_CHAR(10) || COALESCE(P.OBSERVACAOVENDAPRODUTO, ''),
				IIF(P.OBSERVACAOFICHATECNICA IS NOT NULL OR P.OBSERVACAOFICHATECNICA <> '',
					P.OBSERVACAOFICHATECNICA || ASCII_CHAR(13) || ASCII_CHAR(10) || COALESCE(P.OBSERVACAOVENDAPRODUTO, ''),
					COALESCE(P.OBSERVACAOVENDAPRODUTO, ''))
			) AS OBSERVACAO,
			-1 AS USO_ATENDIMENTO,
			CASE P.INATIVOPRODUTO
				WHEN 0 THEN -1
				WHEN 1 THEN 0
			END AS ATIVO,
			1 AS CADASTRO_LJ,
			1 AS CADASTRO_CF,
			P.DT_CREATION AS CADASTRO_DT,
			1 AS ALTERACAO_LJ,
			1 AS ALTERACAO_CF,
			P.DT_ALTER AS ALTERACAO_DT
		FROM PRODUTO P
		WHERE P.CODIGO_PS IS NOT NULL 
		AND P.CONVERSAO = ?
		ORDER BY 1
	`

	rows, err := prisma5DB.Query(produtosQuery, conversaoInt)
	if err != nil {
		return fmt.Errorf("erro ao buscar produtos para inserir: %w", err)
	}
	defer rows.Close()

	insertProdutoSQL := `INSERT INTO ESTOQUE_GERAL (CODIGO, MOVIMENTOESTOQUE, CODIGO_GRUPO, CODIGO_ESTOQUE_CLASSIFICACAO, CODIGO_ESTOQUE_TIPO, NOMEPRODUTO, NOMEPRODUTO_ROTULO, NOMEPRODUTO_FRACAO, CODIGO_NCM, CODIGO_DCB, CODIGOBARRA, CODIGO_ESTOQUE_PORTARIA, FRACAOVENDA, UNIDADE, UNIDADE_CALCULO, VALORCOMPRA, VALORCUSTO, VALORVENDA, UTR, UI, DENSIDADE, CONCENTRACAO, FATOREQUIV, FATORCALCULO, ESTOQUEMINIMO, QTDEMAX, QTDEMAXP, CHECARMAX, CARACTERISTICA, OBSERVACAO, USO_ATENDIMENTO, ATIVO, CADASTRO_LJ, CADASTRO_CF, CADASTRO_DT, ALTERACAO_LJ, ALTERACAO_CF, ALTERACAO_DT) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	tx, err := pharmacieDB.Begin()
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação para produtos: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(insertProdutoSQL)
	if err != nil {
		return fmt.Errorf("erro ao preparar statement de produtos: %w", err)
	}
	defer stmt.Close()

	contadorProdutos := 0
	for rows.Next() {
		var codigo, movimentoEstoque, codigoGrupo, codigoEstoqueClassificacao, codigoEstoqueTipo, fracaoVenda, fatorCalculo, checarMax, usoAtendimento, ativo, cadastroLJ, cadastroCF, alteracaoLJ, alteracaoCF sql.NullInt64
		var codigoEstoquePortaria sql.NullInt64
		var nomeProduto, nomeProdutoRotulo, nomeProdutoFracao, codigoNCM, codigoDCB, codigoBarra, unidade, unidadeCalculo, caracteristica, observacao sql.NullString
		var valorCompra, valorCusto, valorVenda, utr, ui, densidade, concentracao, fatorEquiv, estoqueMinimo, qtdMax, qtdMaxP sql.NullFloat64
		var cadastroDT, alteracaoDT sql.NullTime

		err := rows.Scan(&codigo, &movimentoEstoque, &codigoGrupo, &codigoEstoqueClassificacao, &codigoEstoqueTipo, &nomeProduto, &nomeProdutoRotulo, &nomeProdutoFracao, &codigoNCM, &codigoDCB, &codigoBarra, &codigoEstoquePortaria, &fracaoVenda, &unidade, &unidadeCalculo, &valorCompra, &valorCusto, &valorVenda, &utr, &ui, &densidade, &concentracao, &fatorEquiv, &fatorCalculo, &estoqueMinimo, &qtdMax, &qtdMaxP, &checarMax, &caracteristica, &observacao, &usoAtendimento, &ativo, &cadastroLJ, &cadastroCF, &cadastroDT, &alteracaoLJ, &alteracaoCF, &alteracaoDT)
		if err != nil {
			logger.Warn("Erro ao escanear produto: %v", err)
			continue
		}

		if !codigo.Valid {
			continue
		}

		_, err = stmt.Exec(codigo, movimentoEstoque, codigoGrupo, codigoEstoqueClassificacao, codigoEstoqueTipo, nomeProduto, nomeProdutoRotulo, nomeProdutoFracao, codigoNCM, codigoDCB, codigoBarra, codigoEstoquePortaria, fracaoVenda, unidade, unidadeCalculo, valorCompra, valorCusto, valorVenda, utr, ui, densidade, concentracao, fatorEquiv, fatorCalculo, estoqueMinimo, qtdMax, qtdMaxP, checarMax, caracteristica, observacao, usoAtendimento, ativo, cadastroLJ, cadastroCF, cadastroDT, alteracaoLJ, alteracaoCF, alteracaoDT)
		if err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "unique") && !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				logger.Warn("Erro ao inserir produto %d: %v", codigo.Int64, err)
			}
			continue
		}
		contadorProdutos++
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("erro ao commitar transação de produtos: %w", err)
	}
	logger.Info("✅ %d produtos inseridos no Pharmacie", contadorProdutos)

	// 7. Insere sinônimos no banco Pharmacie (como produtos secundários)
	logger.Info("Inserindo sinônimos no Pharmacie...")
	// Descrições: CAST + SUBSTRING evitam isc 335544321 (truncamento / BLOB); texto como no banco (sem UPPER).
	// FATOREQUIV em DOUBLE evita overflow em NUMERIC grande. CONVERSAO comparada como INTEGER.
	sinonimosQuery := `
		SELECT
			S.CODIGO_SINONIMO_PS AS CODIGO,
			S.CODIGO_PRODUTO_PS AS CODIGOPRINCIPAL,
			CASE
				WHEN S.CODIGO_SINONIMO_PS < 1000000 THEN 0
				WHEN S.CODIGO_SINONIMO_PS > 1000000 AND S.CODIGO_SINONIMO_PS < 1999999 THEN 1
				WHEN S.CODIGO_SINONIMO_PS > 2000000 AND S.CODIGO_SINONIMO_PS < 2999999 THEN 2
				WHEN S.CODIGO_SINONIMO_PS > 3000000 AND S.CODIGO_SINONIMO_PS < 3999999 THEN 3
				WHEN S.CODIGO_SINONIMO_PS > 4000000 AND S.CODIGO_SINONIMO_PS < 4999999 THEN 4
			END AS MOVIMENTOESTOQUE,
			CASE
				WHEN S.CODIGO_SINONIMO_PS < 1000000 THEN 1
				WHEN S.CODIGO_SINONIMO_PS > 1000000 AND S.CODIGO_SINONIMO_PS < 1999999 THEN 100000
				WHEN S.CODIGO_SINONIMO_PS > 2000000 AND S.CODIGO_SINONIMO_PS < 2999999 THEN 200000
				WHEN S.CODIGO_SINONIMO_PS > 3000000 AND S.CODIGO_SINONIMO_PS < 3999999 THEN 300000
				WHEN S.CODIGO_SINONIMO_PS > 4000000 AND S.CODIGO_SINONIMO_PS < 4999999 THEN 400000
			END AS CODIGO_GRUPO,
			CASE
				WHEN S.CODIGO_SINONIMO_PS < 1000000 THEN 1
				WHEN S.CODIGO_SINONIMO_PS > 1000000 AND S.CODIGO_SINONIMO_PS < 1999999 THEN 100000
				WHEN S.CODIGO_SINONIMO_PS > 2000000 AND S.CODIGO_SINONIMO_PS < 2999999 THEN 200000
				WHEN S.CODIGO_SINONIMO_PS > 3000000 AND S.CODIGO_SINONIMO_PS < 3999999 THEN 300000
				WHEN S.CODIGO_SINONIMO_PS > 4000000 AND S.CODIGO_SINONIMO_PS < 4999999 THEN 400000
			END AS CODIGO_ESTOQUE_CLASSIFICACAO,
			CASE
				WHEN S.CODIGO_SINONIMO_PS < 1000000 THEN 1
				WHEN S.CODIGO_SINONIMO_PS > 1000000 AND S.CODIGO_SINONIMO_PS < 1999999 THEN 100000
				WHEN S.CODIGO_SINONIMO_PS > 2000000 AND S.CODIGO_SINONIMO_PS < 2999999 THEN 200000
				WHEN S.CODIGO_SINONIMO_PS > 3000000 AND S.CODIGO_SINONIMO_PS < 3999999 THEN 300000
				WHEN S.CODIGO_SINONIMO_PS > 4000000 AND S.CODIGO_SINONIMO_PS < 4999999 THEN 400000
			END AS CODIGO_ESTOQUE_TIPO,
			SUBSTRING(COALESCE(CAST(S.DESCRICAOSINONIMO AS VARCHAR(8000)), '') FROM 1 FOR 240) AS NOMEPRODUTO,
			SUBSTRING(COALESCE(CAST(S.DESCRICAOROTULOSINONIMO AS VARCHAR(8000)), '') FROM 1 FOR 240) AS NOMEPRODUTO_ROTULO,
			CAST(IIF(COALESCE(S.FATOREQUIVALENCIASINONIMO, 0) = 0, 1, S.FATOREQUIVALENCIASINONIMO) AS DOUBLE PRECISION) AS FATOREQUIV,
			CASE
				WHEN S.FATOREQUIVALENCIASINONIMO = 1 THEN -1
				WHEN S.FATOREQUIVALENCIASINONIMO IS NULL THEN -1
				WHEN S.FATOREQUIVALENCIASINONIMO <> 1 THEN -1
			END AS FATOREQUIV_UC,
			-1 AS ATIVO,
			-1 AS USO_ATENDIMENTO,
			1 AS CADASTRO_LJ,
			1 AS CADASTRO_CF,
			S.DT_CREATION AS CADASTRO_DT,
			1 AS ALTERACAO_LJ,
			1 AS ALTERACAO_CF,
			S.DT_ALTER AS ALTERACAO_DT
		FROM SINONIMO S
		WHERE S.CODIGO_PRODUTO_PS IS NOT NULL
		AND S.CODIGO_SINONIMO_PS IS NOT NULL
		AND CAST(S.CONVERSAO AS INTEGER) = ?
	`

	rows, err = prisma5DB.Query(sinonimosQuery, conversaoInt)
	if err != nil {
		return fmt.Errorf("erro ao buscar sinônimos para inserir: %w", err)
	}
	defer rows.Close()

	insertSinonimoSQL := `INSERT INTO ESTOQUE_GERAL (CODIGO, CODIGOPRINCIPAL, MOVIMENTOESTOQUE, CODIGO_GRUPO, CODIGO_ESTOQUE_CLASSIFICACAO, CODIGO_ESTOQUE_TIPO, NOMEPRODUTO, NOMEPRODUTO_ROTULO, FATOREQUIV, FATOREQUIV_UC, ATIVO, USO_ATENDIMENTO, CADASTRO_LJ, CADASTRO_CF, CADASTRO_DT, ALTERACAO_LJ, ALTERACAO_CF, ALTERACAO_DT) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	tx, err = pharmacieDB.Begin()
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação para sinônimos: %w", err)
	}
	defer tx.Rollback()

	stmt, err = tx.Prepare(insertSinonimoSQL)
	if err != nil {
		return fmt.Errorf("erro ao preparar statement de sinônimos: %w", err)
	}
	defer stmt.Close()

	contadorSinonimos := 0
	for rows.Next() {
		var codigo, codigoPrincipal, movimentoEstoque, codigoGrupo, codigoEstoqueClassificacao, codigoEstoqueTipo, fatorEquivUC, ativo, usoAtendimento, cadastroLJ, cadastroCF, alteracaoLJ, alteracaoCF sql.NullInt64
		var nomeProduto, nomeProdutoRotulo sql.NullString
		var fatorEquiv sql.NullFloat64
		var cadastroDT, alteracaoDT sql.NullTime

		err := rows.Scan(&codigo, &codigoPrincipal, &movimentoEstoque, &codigoGrupo, &codigoEstoqueClassificacao, &codigoEstoqueTipo, &nomeProduto, &nomeProdutoRotulo, &fatorEquiv, &fatorEquivUC, &ativo, &usoAtendimento, &cadastroLJ, &cadastroCF, &cadastroDT, &alteracaoLJ, &alteracaoCF, &alteracaoDT)
		if err != nil {
			logger.Warn("Erro ao escanear sinônimo: %v", err)
			continue
		}

		if !codigo.Valid || !codigoPrincipal.Valid {
			continue
		}

		// Verificar se o produto principal existe no Pharmacie antes de inserir o sinônimo
		var produtoExiste int
		errCheck := pharmacieDB.QueryRow("SELECT COUNT(*) FROM ESTOQUE_GERAL WHERE CODIGO = ?", codigoPrincipal.Int64).Scan(&produtoExiste)
		if errCheck != nil || produtoExiste == 0 {
			logger.Warn("Produto principal %d não encontrado no Pharmacie, pulando sinônimo %d", codigoPrincipal.Int64, codigo.Int64)
			continue
		}

		_, err = stmt.Exec(codigo, codigoPrincipal, movimentoEstoque, codigoGrupo, codigoEstoqueClassificacao, codigoEstoqueTipo, nomeProduto, nomeProdutoRotulo, fatorEquiv, fatorEquivUC, ativo, usoAtendimento, cadastroLJ, cadastroCF, cadastroDT, alteracaoLJ, alteracaoCF, alteracaoDT)
		if err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "unique") && !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				logger.Warn("Erro ao inserir sinônimo %d: %v", codigo.Int64, err)
			}
			continue
		}
		contadorSinonimos++
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("erro ao commitar transação de sinônimos: %w", err)
	}
	logger.Info("✅ %d sinônimos inseridos no Pharmacie", contadorSinonimos)

	logger.Info("✅ Migração de produtos concluída com sucesso!")
	return nil
}

