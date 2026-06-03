package prisma5

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/primesoftwaresi/prime-migration/pkg/logger"
)

// MigratePRISMA5Fornecedores migra fornecedores do PRISMA5 para o Pharmacie
// Parâmetro conversao: valor da conversão (ex: "1", "2")
func MigratePRISMA5Fornecedores(prisma5DB *sql.DB, conversao string) error {
	logger.Info("Iniciando migração de fornecedores (conversão: %s)", conversao)

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

	// 1. Atualiza campo CONVERSAO dos fornecedores
	logger.Info("Atualizando campo CONVERSAO dos fornecedores...")
	_, err = prisma5DB.Exec(`
		UPDATE FORNECEDOR F 
		SET F.CONVERSAO = ? 
		WHERE F.CODIGO_PS IS NULL
	`, conversaoInt)
	if err != nil {
		return fmt.Errorf("erro ao atualizar CONVERSAO: %w", err)
	}

	// 2. Insere fornecedores no banco Pharmacie
	// Nota: Fornecedores usam CODIGOFORNECEDOR diretamente como CODIGO (não gera CODIGO_PS)
	logger.Info("Inserindo fornecedores no Pharmacie...")
	fornecedoresQuery := `
		SELECT
			CAST(F.CODIGOFORNECEDOR AS INTEGER) AS CODIGO,
			UPPER(COALESCE(NULLIF(F.NOMEFORNECEDOR, ''), NULL)) AS NOMEFANTASIA,
			UPPER(COALESCE(NULLIF(F.NOMEFORNECEDOR, ''), NULL)) AS NOMEFORNECEDOR,
			500000 AS CODIGO_GRUPO,
			IIF(NULLIF(TRIM(F.CPFFORNECEDOR), '') IS NOT NULL, 1, 2) AS TIPOPESSOA,
			COALESCE(NULLIF(TRIM(F.CPFFORNECEDOR), ''), NULLIF(TRIM(F.CNPJFORNECEDOR), '')) AS CPF_CNPJ,
			NULLIF(TRIM(F.IEFORNECEDOR), '') AS RG_IE,
			NULLIF(TRIM(F.CONTATOFORNECEDOR), '') AS CONTATO1,
			LOWER(NULLIF(TRIM(F.HOMEPAGEFORNECEDOR), '')) AS HOMEPAGE,
			LOWER(NULLIF(TRIM(F.EMAILFORNECEDOR), '')) AS EMAIL,
			IIF(
				(F.CODIGOBANCO IS NOT NULL OR 
				 NULLIF(TRIM(F.AGENCIAFORNECEDOR), '') IS NOT NULL OR 
				 NULLIF(TRIM(F.CONTAFORNECEDOR), '') IS NOT NULL),
				TRIM(
					COALESCE('CÓDIGO BANCO: ' || CAST(F.CODIGOBANCO AS INTEGER), '') ||
					COALESCE(ASCII_CHAR(13) || ASCII_CHAR(10) || 'AGÊNCIA: ' || NULLIF(F.AGENCIAFORNECEDOR, ''), '') ||
					COALESCE(ASCII_CHAR(13) || ASCII_CHAR(10) || 'CONTA: ' || NULLIF(F.CONTAFORNECEDOR, ''), '')
				),
				NULL
			) AS OBSERVACAO,
			1 AS CADASTRO_LJ,
			1 AS CADASTRO_CF,
			F.DT_CREATION AS CADASTRO_DT,
			1 AS ALTERACAO_LJ,
			1 AS ALTERACAO_CF,
			F.DT_ALTER AS ALTERACAO_DT
		FROM FORNECEDOR F
		WHERE F.CONVERSAO = ?
		ORDER BY 1
	`

	rows, err := prisma5DB.Query(fornecedoresQuery, conversaoInt)
	if err != nil {
		return fmt.Errorf("erro ao buscar fornecedores para inserir: %w", err)
	}
	defer rows.Close()

	insertFornecedorSQL := `INSERT INTO FORNECEDOR (CODIGO, NOMEFANTASIA, NOMEFORNECEDOR, CODIGO_GRUPO, TIPOPESSOA, CPF_CNPJ, RG_IE, CONTATO1, HOMEPAGE, EMAIL, OBSERVACAO, CADASTRO_LJ, CADASTRO_CF, CADASTRO_DT, ALTERACAO_LJ, ALTERACAO_CF, ALTERACAO_DT) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	tx, err := pharmacieDB.Begin()
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação para fornecedores: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(insertFornecedorSQL)
	if err != nil {
		return fmt.Errorf("erro ao preparar statement de fornecedores: %w", err)
	}
	defer stmt.Close()

	contadorFornecedores := 0
	for rows.Next() {
		var codigo, codigoGrupo, tipoPessoa, cadastroLJ, cadastroCF, alteracaoLJ, alteracaoCF sql.NullInt64
		var nomeFantasia, nomeFornecedor, cpfCnpj, rgIe, contato1, homepage, email, observacao sql.NullString
		var cadastroDT, alteracaoDT sql.NullTime

		err := rows.Scan(&codigo, &nomeFantasia, &nomeFornecedor, &codigoGrupo, &tipoPessoa, &cpfCnpj, &rgIe, &contato1, &homepage, &email, &observacao, &cadastroLJ, &cadastroCF, &cadastroDT, &alteracaoLJ, &alteracaoCF, &alteracaoDT)
		if err != nil {
			logger.Warn("Erro ao escanear fornecedor: %v", err)
			continue
		}

		if !codigo.Valid {
			continue
		}

		_, err = stmt.Exec(codigo, nomeFantasia, nomeFornecedor, codigoGrupo, tipoPessoa, cpfCnpj, rgIe, contato1, homepage, email, observacao, cadastroLJ, cadastroCF, cadastroDT, alteracaoLJ, alteracaoCF, alteracaoDT)
		if err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "unique") && !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				logger.Warn("Erro ao inserir fornecedor %d: %v", codigo.Int64, err)
			}
			continue
		}
		contadorFornecedores++
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("erro ao commitar transação de fornecedores: %w", err)
	}
	logger.Info("✅ %d fornecedores inseridos no Pharmacie", contadorFornecedores)

	// 3. Insere endereços dos fornecedores
	logger.Info("Inserindo endereços dos fornecedores...")
	enderecosFornecedoresQuery := `
		SELECT
			CAST(F.CODIGOFORNECEDOR AS INTEGER) AS CODIGO_CADASTRO,
			F.ENDERECOFORNECEDOR,
			F.NUMEROENDERECOFORNECEDOR,
			F.COMPLEMENTOFORNECEDOR,
			F.CEPFORNECEDOR,
			B.NOMEBAIRRO,
			C.NOMECIDADE,
			C.CODIGO_PS AS CODIGO_CIDADE_PS,
			F.DT_CREATION,
			F.DT_ALTER
		FROM FORNECEDOR F
			LEFT JOIN CIDADE C ON C.CODIGOCIDADE = F.CODIGOCIDADE
			LEFT JOIN BAIRRO B ON B.CODIGOBAIRRO = F.CODIGOBAIRRO
		WHERE F.ENDERECOFORNECEDOR <> ''
		AND F.CONVERSAO = ?
	`

	rows, err = prisma5DB.Query(enderecosFornecedoresQuery, conversaoInt)
	if err != nil {
		return fmt.Errorf("erro ao buscar endereços dos fornecedores: %w", err)
	}
	defer rows.Close()

	insertEnderecoSQL := `INSERT INTO CADASTRO_ENDERECO (CODIGO, TIPO_CADASTRO, CODIGO_CADASTRO, ENDERECO, NUMERO, OBSERVACAO, CODIGO_REGIAODETALHE, CODIGO_CIDADEESTADO, CEP, CADASTRO_LJ, CADASTRO_CF, CADASTRO_DT, ALTERACAO_LJ, ALTERACAO_CF, ALTERACAO_DT) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	tx, err = pharmacieDB.Begin()
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação para endereços: %w", err)
	}
	defer tx.Rollback()

	stmt, err = tx.Prepare(insertEnderecoSQL)
	if err != nil {
		return fmt.Errorf("erro ao preparar statement de endereços: %w", err)
	}
	defer stmt.Close()

	contadorEnderecos := 0
	for rows.Next() {
		var codigoCadastro sql.NullInt64
		var enderecoFornecedor, numeroEnderecoFornecedor, complementoFornecedor, cepFornecedor, nomeBairro, nomeCidade sql.NullString
		var codigoCidadePS sql.NullInt64
		var dtCreation, dtAlter sql.NullTime

		if err := rows.Scan(&codigoCadastro, &enderecoFornecedor, &numeroEnderecoFornecedor, &complementoFornecedor, &cepFornecedor, &nomeBairro, &nomeCidade, &codigoCidadePS, &dtCreation, &dtAlter); err != nil {
			logger.Warn("Erro ao escanear endereço do fornecedor: %v", err)
			continue
		}

		if !codigoCadastro.Valid {
			continue
		}

		// Gerar código do endereço
		var codigoEndereco int
		err = prisma5DB.QueryRow(`SELECT GEN_ID(GEN_CADASTRO_ENDERECO, 1) FROM RDB$DATABASE`).Scan(&codigoEndereco)
		if err != nil {
			logger.Warn("Erro ao gerar código do endereço: %v", err)
			continue
		}

		// Montar endereço completo
		enderecoCompleto := ""
		if enderecoFornecedor.Valid && strings.TrimSpace(enderecoFornecedor.String) != "" {
			enderecoCompleto = strings.ToUpper(strings.TrimSpace(enderecoFornecedor.String))
		}
		if nomeBairro.Valid && nomeBairro.String != "" {
			if enderecoCompleto != "" {
				enderecoCompleto += ", " + nomeBairro.String
			} else {
				enderecoCompleto = nomeBairro.String
			}
		}

		numero := "S/N"
		if numeroEnderecoFornecedor.Valid && strings.TrimSpace(numeroEnderecoFornecedor.String) != "" {
			numero = strings.TrimSpace(numeroEnderecoFornecedor.String)
		}

		observacao := ""
		if complementoFornecedor.Valid && strings.TrimSpace(complementoFornecedor.String) != "" {
			observacao = strings.ToUpper(strings.TrimSpace(complementoFornecedor.String))
		}

		cep := ""
		if cepFornecedor.Valid && strings.TrimSpace(cepFornecedor.String) != "" {
			cepLimpo := strings.ReplaceAll(strings.TrimSpace(cepFornecedor.String), "-", "")
			if cepLimpo != "" {
				cep = cepLimpo
			}
		}

		codigoCidadeEstado := int64(1)
		if codigoCidadePS.Valid {
			codigoCidadeEstado = codigoCidadePS.Int64
		}

		_, err = stmt.Exec(codigoEndereco, 4, codigoCadastro.Int64, enderecoCompleto, numero, observacao, 1, codigoCidadeEstado, cep, 1, 1, dtCreation, 1, 1, dtAlter)
		if err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "unique") && !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				logger.Warn("Erro ao inserir endereço do fornecedor %d: %v", codigoEndereco, err)
			}
			continue
		}
		contadorEnderecos++
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("erro ao commitar transação de endereços: %w", err)
	}
	logger.Info("✅ %d endereços de fornecedores inseridos", contadorEnderecos)

	// 4. Insere telefones principais dos fornecedores
	logger.Info("Inserindo telefones principais dos fornecedores...")
	telefonesPrincipaisQuery := `
		SELECT
			CAST(F.CODIGOFORNECEDOR AS INTEGER) AS CODIGO_CADASTRO,
			F.TELEFONEFORNECEDOR,
			F.DDDFORNECEDOR,
			F.DT_CREATION,
			F.DT_ALTER
		FROM FORNECEDOR F
		WHERE F.CONVERSAO = ?
	`

	rows, err = prisma5DB.Query(telefonesPrincipaisQuery, conversaoInt)
	if err != nil {
		return fmt.Errorf("erro ao buscar telefones principais dos fornecedores: %w", err)
	}
	defer rows.Close()

	insertTelefoneSQL := `INSERT INTO CADASTRO_TELEFONE (CODIGO, TIPO_CADASTRO, CODIGO_CADASTRO, TELEFONE_TIPO, TELEFONEPREFIXO, TELEFONE, CADASTRO_LJ, CADASTRO_CF, CADASTRO_DT, ALTERACAO_LJ, ALTERACAO_CF, ALTERACAO_DT) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	tx, err = pharmacieDB.Begin()
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação para telefones principais: %w", err)
	}
	defer tx.Rollback()

	stmt, err = tx.Prepare(insertTelefoneSQL)
	if err != nil {
		return fmt.Errorf("erro ao preparar statement de telefones principais: %w", err)
	}
	defer stmt.Close()

	contadorTelefonesPrincipais := 0
	for rows.Next() {
		var codigoCadastro sql.NullInt64
		var telefoneFornecedor, dddFornecedor sql.NullString
		var dtCreation, dtAlter sql.NullTime

		if err := rows.Scan(&codigoCadastro, &telefoneFornecedor, &dddFornecedor, &dtCreation, &dtAlter); err != nil {
			logger.Warn("Erro ao escanear telefone principal do fornecedor: %v", err)
			continue
		}

		if !codigoCadastro.Valid || !telefoneFornecedor.Valid || telefoneFornecedor.String == "" {
			continue
		}

		// Processar telefone
		telefoneLimpo := stripNonNumeric(telefoneFornecedor.String)
		if telefoneLimpo == "" {
			continue
		}

		// Gerar código do telefone
		var codigoTelefone int
		err = prisma5DB.QueryRow(`SELECT GEN_ID(GEN_CADASTRO_TELEFONE, 1) FROM RDB$DATABASE`).Scan(&codigoTelefone)
		if err != nil {
			logger.Warn("Erro ao gerar código do telefone: %v", err)
			continue
		}

		// Determinar tipo de telefone (se começa com 6,7,8,9 = tipo 3, senão = tipo 2)
		telefoneTipo := 2
		if len(telefoneLimpo) > 0 {
			primeiroDigito := telefoneLimpo[0]
			if primeiroDigito >= '6' && primeiroDigito <= '9' {
				telefoneTipo = 3
			}
		}

		// Processar DDD
		ddd := ""
		if dddFornecedor.Valid && dddFornecedor.String != "" {
			dddLimpo := strings.TrimSpace(dddFornecedor.String)
			if len(dddLimpo) >= 2 {
				if dddLimpo[0] == '0' {
					if len(dddLimpo) > 2 {
						ddd = dddLimpo[1:3]
					}
				} else {
					ddd = dddLimpo[:2]
				}
			}
		}

		// Limitar tamanho do telefone
		if len(telefoneLimpo) > 18 {
			telefoneLimpo = telefoneLimpo[:18]
		}

		_, err = stmt.Exec(codigoTelefone, 4, codigoCadastro.Int64, telefoneTipo, ddd, telefoneLimpo, 1, 1, dtCreation, 1, 1, dtAlter)
		if err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "unique") && !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				logger.Warn("Erro ao inserir telefone principal do fornecedor %d: %v", codigoTelefone, err)
			}
			continue
		}
		contadorTelefonesPrincipais++
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("erro ao commitar transação de telefones principais: %w", err)
	}
	logger.Info("✅ %d telefones principais de fornecedores inseridos", contadorTelefonesPrincipais)

	// 5. Insere telefones de contato dos fornecedores
	logger.Info("Inserindo telefones de contato dos fornecedores...")
	telefonesContatoQuery := `
		SELECT
			CAST(F.CODIGOFORNECEDOR AS INTEGER) AS CODIGO_CADASTRO,
			F.TELEFONECONTATOFORNECEDOR,
			F.DDDFORNECEDOR,
			F.DT_CREATION,
			F.DT_ALTER
		FROM FORNECEDOR F
		WHERE F.CONVERSAO = ?
	`

	rows, err = prisma5DB.Query(telefonesContatoQuery, conversaoInt)
	if err != nil {
		return fmt.Errorf("erro ao buscar telefones de contato dos fornecedores: %w", err)
	}
	defer rows.Close()

	tx, err = pharmacieDB.Begin()
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação para telefones de contato: %w", err)
	}
	defer tx.Rollback()

	stmt, err = tx.Prepare(insertTelefoneSQL)
	if err != nil {
		return fmt.Errorf("erro ao preparar statement de telefones de contato: %w", err)
	}
	defer stmt.Close()

	contadorTelefonesContato := 0
	for rows.Next() {
		var codigoCadastro sql.NullInt64
		var telefoneContatoFornecedor, dddFornecedor sql.NullString
		var dtCreation, dtAlter sql.NullTime

		if err := rows.Scan(&codigoCadastro, &telefoneContatoFornecedor, &dddFornecedor, &dtCreation, &dtAlter); err != nil {
			logger.Warn("Erro ao escanear telefone de contato do fornecedor: %v", err)
			continue
		}

		if !codigoCadastro.Valid || !telefoneContatoFornecedor.Valid || telefoneContatoFornecedor.String == "" {
			continue
		}

		// Processar telefone
		telefoneLimpo := stripNonNumeric(telefoneContatoFornecedor.String)
		if telefoneLimpo == "" {
			continue
		}

		// Gerar código do telefone
		var codigoTelefone int
		err = prisma5DB.QueryRow(`SELECT GEN_ID(GEN_CADASTRO_TELEFONE, 1) FROM RDB$DATABASE`).Scan(&codigoTelefone)
		if err != nil {
			logger.Warn("Erro ao gerar código do telefone de contato: %v", err)
			continue
		}

		// Determinar tipo de telefone
		telefoneTipo := 2
		if len(telefoneLimpo) > 0 {
			primeiroDigito := telefoneLimpo[0]
			if primeiroDigito >= '6' && primeiroDigito <= '9' {
				telefoneTipo = 3
			}
		}

		// Processar DDD
		ddd := ""
		if dddFornecedor.Valid && dddFornecedor.String != "" {
			dddLimpo := strings.TrimSpace(dddFornecedor.String)
			if len(dddLimpo) >= 2 {
				if dddLimpo[0] == '0' {
					if len(dddLimpo) > 2 {
						ddd = dddLimpo[1:3]
					}
				} else {
					ddd = dddLimpo[:2]
				}
			}
		}

		// Limitar tamanho do telefone
		if len(telefoneLimpo) > 18 {
			telefoneLimpo = telefoneLimpo[:18]
		}

		_, err = stmt.Exec(codigoTelefone, 4, codigoCadastro.Int64, telefoneTipo, ddd, telefoneLimpo, 1, 1, dtCreation, 1, 1, dtAlter)
		if err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "unique") && !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				logger.Warn("Erro ao inserir telefone de contato do fornecedor %d: %v", codigoTelefone, err)
			}
			continue
		}
		contadorTelefonesContato++
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("erro ao commitar transação de telefones de contato: %w", err)
	}
	logger.Info("✅ %d telefones de contato de fornecedores inseridos", contadorTelefonesContato)

	// 6. Insere fax dos fornecedores
	logger.Info("Inserindo fax dos fornecedores...")
	faxFornecedoresQuery := `
		SELECT
			CAST(F.CODIGOFORNECEDOR AS INTEGER) AS CODIGO_CADASTRO,
			F.FAXFORNECEDOR,
			F.DDDFORNECEDOR,
			F.DT_CREATION,
			F.DT_ALTER
		FROM FORNECEDOR F
		WHERE F.CONVERSAO = ?
	`

	rows, err = prisma5DB.Query(faxFornecedoresQuery, conversaoInt)
	if err != nil {
		return fmt.Errorf("erro ao buscar fax dos fornecedores: %w", err)
	}
	defer rows.Close()

	tx, err = pharmacieDB.Begin()
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação para fax: %w", err)
	}
	defer tx.Rollback()

	stmt, err = tx.Prepare(insertTelefoneSQL)
	if err != nil {
		return fmt.Errorf("erro ao preparar statement de fax: %w", err)
	}
	defer stmt.Close()

	contadorFax := 0
	for rows.Next() {
		var codigoCadastro sql.NullInt64
		var faxFornecedor, dddFornecedor sql.NullString
		var dtCreation, dtAlter sql.NullTime

		if err := rows.Scan(&codigoCadastro, &faxFornecedor, &dddFornecedor, &dtCreation, &dtAlter); err != nil {
			logger.Warn("Erro ao escanear fax do fornecedor: %v", err)
			continue
		}

		if !codigoCadastro.Valid || !faxFornecedor.Valid || faxFornecedor.String == "" {
			continue
		}

		// Processar fax
		faxLimpo := stripNonNumeric(faxFornecedor.String)
		if faxLimpo == "" {
			continue
		}

		// Gerar código do telefone
		var codigoTelefone int
		err = prisma5DB.QueryRow(`SELECT GEN_ID(GEN_CADASTRO_TELEFONE, 1) FROM RDB$DATABASE`).Scan(&codigoTelefone)
		if err != nil {
			logger.Warn("Erro ao gerar código do fax: %v", err)
			continue
		}

		// Determinar tipo de telefone
		telefoneTipo := 2
		if len(faxLimpo) > 0 {
			primeiroDigito := faxLimpo[0]
			if primeiroDigito >= '6' && primeiroDigito <= '9' {
				telefoneTipo = 3
			}
		}

		// Processar DDD
		ddd := ""
		if dddFornecedor.Valid && dddFornecedor.String != "" {
			dddLimpo := strings.TrimSpace(dddFornecedor.String)
			if len(dddLimpo) >= 2 {
				if dddLimpo[0] == '0' {
					if len(dddLimpo) > 2 {
						ddd = dddLimpo[1:3]
					}
				} else {
					ddd = dddLimpo[:2]
				}
			}
		}

		// Limitar tamanho do fax
		if len(faxLimpo) > 18 {
			faxLimpo = faxLimpo[:18]
		}

		_, err = stmt.Exec(codigoTelefone, 4, codigoCadastro.Int64, telefoneTipo, ddd, faxLimpo, 1, 1, dtCreation, 1, 1, dtAlter)
		if err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "unique") && !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				logger.Warn("Erro ao inserir fax do fornecedor %d: %v", codigoTelefone, err)
			}
			continue
		}
		contadorFax++
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("erro ao commitar transação de fax: %w", err)
	}
	logger.Info("✅ %d fax de fornecedores inseridos", contadorFax)

	logger.Info("✅ Migração de fornecedores concluída com sucesso!")
	return nil
}

