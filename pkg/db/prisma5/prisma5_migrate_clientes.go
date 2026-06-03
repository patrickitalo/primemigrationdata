package prisma5

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/primesoftwaresi/prime-migration/pkg/logger"
)

// MigratePRISMA5Clientes migra clientes do PRISMA5 para o Pharmacie
// Parâmetro conversao: valor da conversão (ex: "1", "2").
// cb opcional: histórico incremental e gravação no banco central (tipo 1 / clientes).
func MigratePRISMA5Clientes(prisma5DB *sql.DB, conversao string, cb *MigrationCallbacks) (*OptionStats, error) {
	logger.Info("Iniciando migração de clientes (conversão: %s)", conversao)

	var outStats *OptionStats
	if cb != nil {
		if cb.Stats == nil {
			cb.Stats = &OptionStats{}
		}
		outStats = cb.Stats
	}

	conversaoInt, err := strconv.Atoi(conversao)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter conversão para inteiro: %w", err)
	}

	// Conectar ao Pharmacie
	pharmacieDB, err := connectToPharmacie(prisma5DB)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao Pharmacie: %w", err)
	}
	defer pharmacieDB.Close()

	// 1. Remove acentos das cidades (CIDADEESTADO)
	logger.Info("Removendo acentos das cidades (CIDADEESTADO)...")
	rows, err := prisma5DB.Query("SELECT CODIGO, NOMECIDADE FROM CIDADEESTADO WHERE NOMECIDADE <> ''")
	if err != nil {
		logger.Warn("Aviso ao buscar cidades de CIDADEESTADO: %v", err)
	} else {
		defer rows.Close()

		// Iniciar transação para melhor performance
		tx, err := prisma5DB.Begin()
		if err != nil {
			logger.Warn("Aviso ao iniciar transação para CIDADEESTADO: %v", err)
		} else {
			defer tx.Rollback()

			// Preparar UPDATE statement dentro da transação
			updateStmt, err := tx.Prepare("UPDATE CIDADEESTADO SET NOMECIDADE = ? WHERE CODIGO = ?")
			if err != nil {
				logger.Warn("Aviso ao preparar UPDATE de CIDADEESTADO: %v", err)
				tx.Rollback()
			} else {
				defer updateStmt.Close()

				contadorAtualizados := 0
				batchSize := 1000

				for rows.Next() {
					var codigo int
					var nomeCidade string
					if err := rows.Scan(&codigo, &nomeCidade); err == nil {
						nomeSemAcentos := tiraAcentos(nomeCidade)
						if nomeSemAcentos != nomeCidade {
							_, err := updateStmt.Exec(nomeSemAcentos, codigo)
							if err != nil {
								logger.Warn("Erro ao atualizar cidade %d (CODIGO=%d): %v", codigo, codigo, err)
								continue
							}
							contadorAtualizados++

							// Commit a cada batchSize registros
							if contadorAtualizados%batchSize == 0 {
								if err := tx.Commit(); err != nil {
									logger.Warn("Erro ao commitar transação (tentando continuar): %v", err)
									tx.Rollback()
								}
								// Reiniciar transação
								tx, err = prisma5DB.Begin()
								if err != nil {
									logger.Warn("Erro ao reiniciar transação: %v", err)
									break
								}
								updateStmt.Close()
								updateStmt, err = tx.Prepare("UPDATE CIDADEESTADO SET NOMECIDADE = ? WHERE CODIGO = ?")
								if err != nil {
									logger.Warn("Erro ao preparar UPDATE na nova transação: %v", err)
									break
								}
								logger.Info("  %d cidades atualizadas... (total: %d)", batchSize, contadorAtualizados)
							}
						}
					}
				}

				// Commit final
				if err := tx.Commit(); err != nil {
					logger.Warn("Erro ao commitar transação final de CIDADEESTADO: %v", err)
				} else {
					logger.Info("✅ %d cidades de CIDADEESTADO atualizadas (acentos removidos)", contadorAtualizados)
				}
			}
		}
	}

	// 2. Remove acentos das cidades (CIDADE)
	logger.Info("Removendo acentos das cidades (CIDADE)...")
	rows, err = prisma5DB.Query("SELECT CODIGOCIDADE, NOMECIDADE FROM CIDADE WHERE NOMECIDADE <> ''")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var codigoCidade int
			var nomeCidade string
			if err := rows.Scan(&codigoCidade, &nomeCidade); err == nil {
				nomeSemAcentos := tiraAcentos(nomeCidade)
				if nomeSemAcentos != nomeCidade {
					prisma5DB.Exec("UPDATE CIDADE SET NOMECIDADE = ? WHERE CODIGOCIDADE = ?", nomeSemAcentos, codigoCidade)
				}
			}
		}
	}

	// 3. Atualiza CODIGO_PS das cidades por CODIGO_IBGE
	logger.Info("Atualizando CODIGO_PS das cidades por CODIGO_IBGE...")
	_, err = prisma5DB.Exec(`
		UPDATE CIDADE C SET 
		C.CODIGO_PS = (
			SELECT CE.CODIGO FROM CIDADEESTADO CE 
			WHERE CE.CODIGO_IBGE = C.CODIGOIBGE AND CE.CODIGO_IBGE IS NOT NULL
		)
		WHERE C.CODIGO_PS IS NULL
	`)
	if err != nil {
		logger.Warn("Aviso ao atualizar CODIGO_PS por CODIGO_IBGE: %v", err)
	}

	// 3. Atualiza CODIGO_PS das cidades por NOME
	logger.Info("Atualizando CODIGO_PS das cidades por NOME...")
	_, err = prisma5DB.Exec(`
		UPDATE CIDADE C SET 
		C.CODIGO_PS = (
			SELECT FIRST 1 CE.CODIGO FROM CIDADEESTADO CE 
			WHERE CE.NOMECIDADE = C.NOMECIDADE
		)
		WHERE C.CODIGO_PS IS NULL
	`)
	if err != nil {
		logger.Warn("Aviso ao atualizar CODIGO_PS por NOME: %v", err)
	}

	// 4. Atualiza campo CONVERSAO dos clientes
	logger.Info("Atualizando campo CONVERSAO dos clientes...")
	_, err = prisma5DB.Exec(`
		UPDATE CLIENTE C 
		SET C.CONVERSAO = ? 
		WHERE C.CODIGO_PS IS NULL
	`, conversaoInt)
	if err != nil {
		return nil, fmt.Errorf("erro ao atualizar CONVERSAO: %w", err)
	}

	// 5. Gera CODIGO_PS para clientes usando sequence (OTIMIZADO: UPDATE em massa)
	logger.Info("Gerando CODIGO_PS para clientes...")
	if err := generateCodigoPSInBatch(prisma5DB, "CLIENTE", "CODIGOCLIENTE", "GEN_CLIENTE", 0); err != nil {
		return nil, fmt.Errorf("erro ao gerar CODIGO_PS para clientes: %w", err)
	}

	// 6. Atualiza CODIGO_PS dos endereços de entrega
	logger.Info("Atualizando CODIGO_PS dos endereços de entrega...")
	_, err = prisma5DB.Exec(`
		UPDATE CLIENTE_ENDERECO_ENTREGA CEC 
		SET CEC.CODIGO_PS = (
			SELECT FIRST 1 C.CODIGO_PS FROM CLIENTE C 
			WHERE C.CODIGOCLIENTE = CEC.CODIGOCLIENTE
		)
	`)
	if err != nil {
		logger.Warn("Aviso ao atualizar CODIGO_PS dos endereços: %v", err)
	}

	// 7. Insere clientes no banco Pharmacie
	logger.Info("Inserindo clientes no Pharmacie...")
	clientesQuery := `
		SELECT
			C.CODIGO_PS AS CODIGO,
			800000 AS CODIGO_GRUPO,
			IIF(CHAR_LENGTH(REPLACE(REPLACE(REPLACE(CPFCNPJCLIENTE, '-', ''), '/', ''), '.', '')) > 11, 2, 1) AS TIPOPESSOA,
			UPPER(IIF(NULLIF(TRIM(NOMECLIENTE), '') IS NULL, NULL, NOMECLIENTE)) AS NOMECLIENTE,
			UPPER(IIF(NULLIF(TRIM(NOMEROTULOCLIENTE), '') IS NULL, NOMECLIENTE, NOMEROTULOCLIENTE)) AS NOMEROTULO,
			IIF(NULLIF(TRIM(GENERO), '') IS NULL, 1, GENERO) AS SEXO,
			IIF(NULLIF(TRIM(CPFCNPJCLIENTE), '') IS NULL, NULL, REPLACE(REPLACE(REPLACE(CPFCNPJCLIENTE, '-', ''), '/', ''), '.', '')) AS CPF_CNPJ,
			IIF(NULLIF(TRIM(RGCLIENTE), '') IS NULL, 19, 2) AS CODIGO_TIPODOCUMENTO,
			IIF(NULLIF(TRIM(RGCLIENTE), '') IS NULL, 
				IIF(NULLIF(TRIM(IECLIENTE), '') IS NULL, NULL, IECLIENTE), 
				RGCLIENTE) AS RG_IE,
			IIF(NULLIF(TRIM(ORGAOEXPEDIDORCLIENTE), '') IS NULL, NULL, ORGAOEXPEDIDORCLIENTE) AS ORGAOEXPEDIDOR,
			IIF(NULLIF(TRIM(UFEXPEDIDORCLIENTE), '') IS NULL, NULL, UFEXPEDIDORCLIENTE) AS ORGAOEXPEDIDOR_UF,
			IIF(NULLIF(TRIM(EMAILCLIENTE), '') IS NULL, NULL, EMAILCLIENTE) AS EMAIL1,
			COALESCE(EXTRACT(DAY FROM DATANASCIMENTOCLIENTE), 0) AS DIANASCIMENTO,
			COALESCE(EXTRACT(MONTH FROM DATANASCIMENTOCLIENTE), 0) AS MESNASCIMENTO,
			COALESCE(EXTRACT(YEAR FROM DATANASCIMENTOCLIENTE), 0) AS ANONASCIMENTO,
			CASE IE_ATIVO
				WHEN 'A' THEN -1
				ELSE 0
			END AS ATIVO,
			-1 AS CONSUMIDORFINAL,
			1 AS TIPODESCONTO,
			3 AS NAOCONTRIBUINTE,
			CAST(SUBSTRING(
				IIF(NULLIF(TRIM(MENSAGEMVENDACLIENTE), '') IS NOT NULL,
					MENSAGEMVENDACLIENTE || ASCII_CHAR(13) || ASCII_CHAR(10),
					'') ||
				IIF(NULLIF(TRIM(OBSGERALCLIENTE), '') IS NOT NULL,
					OBSGERALCLIENTE || ASCII_CHAR(13) || ASCII_CHAR(10),
					'') ||
				IIF(NULLIF(TRIM(OBSERVACAOOPCLIENTE), '') IS NOT NULL,
					OBSERVACAOOPCLIENTE,
					'')
				FROM 1 FOR 7000) AS VARCHAR(7000)) AS OBSERVACAO,
			1 AS CADASTRO_LJ,
			1 AS CADASTRO_CF,
			DT_CREATION AS CADASTRO_DT,
			1 AS ALTERACAO_LJ,
			1 AS ALTERACAO_CF,
			DT_ALTER AS ALTERACAO_DT
		FROM CLIENTE C
		WHERE C.CODIGO_PS IS NOT NULL
		AND C.CONVERSAO = ?
		ORDER BY 1
	`

	rows, err = prisma5DB.Query(clientesQuery, conversaoInt)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar clientes para inserir: %w", err)
	}
	defer rows.Close()

	insertClienteSQL := `INSERT INTO CLIENTE (CODIGO, CODIGO_GRUPO, TIPOPESSOA, NOMECLIENTE, NOMEROTULO, SEXO, CPF_CNPJ, CODIGO_TIPODOCUMENTO, RG_IE, ORGAOEXPEDIDOR, ORGAOEXPEDIDOR_UF, EMAIL1, DIANASCIMENTO, MESNASCIMENTO, ANONASCIMENTO, ATIVO, CONSUMIDORFINAL, TIPODESCONTO, NAOCONTRIBUINTE, OBSERVACAO, CADASTRO_LJ, CADASTRO_CF, CADASTRO_DT, ALTERACAO_LJ, ALTERACAO_CF, ALTERACAO_DT) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	tx, err := pharmacieDB.Begin()
	if err != nil {
		return nil, fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(insertClienteSQL)
	if err != nil {
		return nil, fmt.Errorf("erro ao preparar statement: %w", err)
	}
	defer stmt.Close()

	contadorClientes := 0
	for rows.Next() {
		var codigo, codigoGrupo, tipoPessoa, sexo, codigoTipoDocumento, dianascimento, mesnascimento, anonascimento, ativo, consumidorFinal, tipoDesconto, naoContribuinte, cadastroLJ, cadastroCF, alteracaoLJ, alteracaoCF sql.NullInt64
		var nomeCliente, nomeRotulo, cpfCnpj, rgIe, orgaoExpedidor, orgaoExpedidorUF, email1, observacao sql.NullString
		var cadastroDT, alteracaoDT sql.NullTime

		err := rows.Scan(&codigo, &codigoGrupo, &tipoPessoa, &nomeCliente, &nomeRotulo, &sexo, &cpfCnpj, &codigoTipoDocumento, &rgIe, &orgaoExpedidor, &orgaoExpedidorUF, &email1, &dianascimento, &mesnascimento, &anonascimento, &ativo, &consumidorFinal, &tipoDesconto, &naoContribuinte, &observacao, &cadastroLJ, &cadastroCF, &cadastroDT, &alteracaoLJ, &alteracaoCF, &alteracaoDT)
		if err != nil {
			logger.Warn("Erro ao escanear cliente: %v", err)
			continue
		}

		if cb != nil && cb.Stats != nil {
			cb.Stats.TotalOrigem++
		}

		var sourceKey, rowHash string
		if codigo.Valid {
			sourceKey = strconv.FormatInt(codigo.Int64, 10)
			rowHash = hashClienteMigracao(nomeCliente, cpfCnpj, codigo)
		}

		if cb != nil && cb.Incremental && cb.RecordMeta != nil && sourceKey != "" {
			if prev, ok := cb.RecordMeta[sourceKey]; ok && prev.SourceHash == rowHash {
				cb.Stats.Skipped++
				continue
			}
		}

		_, err = stmt.Exec(codigo, codigoGrupo, tipoPessoa, nomeCliente, nomeRotulo, sexo, cpfCnpj, codigoTipoDocumento, rgIe, orgaoExpedidor, orgaoExpedidorUF, email1, dianascimento, mesnascimento, anonascimento, ativo, consumidorFinal, tipoDesconto, naoContribuinte, observacao, cadastroLJ, cadastroCF, cadastroDT, alteracaoLJ, alteracaoCF, alteracaoDT)
		if err != nil {
			// Ignorar erros de duplicata
			if !strings.Contains(strings.ToLower(err.Error()), "unique") && !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				logger.Warn("Erro ao inserir cliente %d: %v", codigo.Int64, err)
				if cb != nil && cb.Stats != nil {
					cb.Stats.Erros++
				}
			}
			continue
		}
		contadorClientes++
		if cb != nil && cb.SaveRecord != nil && sourceKey != "" {
			if err := cb.SaveRecord(cb.RunID, cb.ClientID, cb.EntityType, sourceKey, sourceKey, rowHash); err != nil {
				if cb.SaveError != nil {
					_ = cb.SaveError(cb.RunID, cb.EntityType, sourceKey, err.Error())
				}
			} else if cb.Stats != nil {
				cb.Stats.Novos++
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("erro ao commitar transação: %w", err)
	}

	logger.Info("✅ %d clientes inseridos no Pharmacie", contadorClientes)

	// 8. Insere endereços de entrega
	logger.Info("Inserindo endereços de entrega...")
	insertEnderecoSQL := `INSERT INTO CADASTRO_ENDERECO (CODIGO, TIPO_CADASTRO, CODIGO_CADASTRO, ENDERECO, NUMERO, OBSERVACAO, CODIGO_REGIAODETALHE, CODIGO_CIDADEESTADO, CEP, CADASTRO_LJ, CADASTRO_CF, CADASTRO_DT, ALTERACAO_LJ, ALTERACAO_CF, ALTERACAO_DT) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	enderecosEntregaQuery := `
		SELECT
			GEN_ID(GEN_CADASTRO_ENDERECO, 1) AS CODIGO,
			1 AS TIPO_CADASTRO,
			CAST(CL.CODIGO_PS AS INTEGER) AS CODIGO_CADASTRO,
			SUBSTRING(
				UPPER(
					COALESCE(CEE.DS_ENDERECO, '') ||
					CASE 
						WHEN CEE.DS_ENDERECO IS NOT NULL AND CEE.DS_ENDERECO <> '' 
							 AND B.NOMEBAIRRO IS NOT NULL AND B.NOMEBAIRRO <> ''
							THEN ', ' || B.NOMEBAIRRO
						WHEN B.NOMEBAIRRO IS NOT NULL AND B.NOMEBAIRRO <> ''
							THEN B.NOMEBAIRRO
						ELSE ''
					END ||
					CASE 
						WHEN (
							(CEE.DS_ENDERECO IS NOT NULL AND CEE.DS_ENDERECO <> '') OR 
							(B.NOMEBAIRRO IS NOT NULL AND B.NOMEBAIRRO <> '')
						) AND CC.NOMECIDADE IS NOT NULL AND CC.NOMECIDADE <> ''
							THEN ', ' || CC.NOMECIDADE
						WHEN CC.NOMECIDADE IS NOT NULL AND CC.NOMECIDADE <> ''
							THEN CC.NOMECIDADE
						ELSE ''
					END
				) FROM 1 FOR 100
			) AS ENDERECO,
			COALESCE(CEE.NR_ENDERECO, 'S/N') AS NUMERO,
			UPPER(COALESCE(CEE.DS_COMPLEMENTO, '') || IIF(CEE.DS_COMPLEMENTO <> '', ', ', '') || COALESCE(CEE.DS_PROXIMIDADE, '')) AS OBSERVACAO,
			1 AS CODIGO_REGIAODETALHE,
			IIF(CC.CODIGO_PS IS NOT NULL, CC.CODIGO_PS, 1) AS CODIGO_CIDADEESTADO,
			COALESCE(CEE.CD_CEP, '00000000') AS CEP,
			1 AS CADASTRO_LJ,
			1 AS CADASTRO_CF,
			CEE.DT_CREATION AS CADASTRO_DT,
			1 AS ALTERACAO_LJ,
			1 AS ALTERACAO_CF,
			CEE.DT_ALTER AS ALTERACAO_DT
		FROM CLIENTE_ENDERECO_ENTREGA CEE
			INNER JOIN CLIENTE CL ON CL.CODIGOCLIENTE = CEE.CODIGOCLIENTE
			LEFT JOIN CIDADE CC ON CC.CODIGOCIDADE = CEE.CD_CIDADE
			LEFT JOIN BAIRRO B ON B.CODIGOBAIRRO = CEE.CD_BAIRRO
		WHERE CEE.DS_ENDERECO <> ''
		AND CL.CONVERSAO = ?
	`

	tx, err = pharmacieDB.Begin()
	if err != nil {
		return nil, fmt.Errorf("erro ao iniciar transação para endereços de entrega: %w", err)
	}
	defer tx.Rollback()

	stmt, err = tx.Prepare(insertEnderecoSQL)
	if err != nil {
		return nil, fmt.Errorf("erro ao preparar statement de endereços de entrega: %w", err)
	}
	defer stmt.Close()

	rows, err = prisma5DB.Query(enderecosEntregaQuery, conversaoInt)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar endereços de entrega: %w", err)
	}
	defer rows.Close()

	contadorEnderecos := 0
	for rows.Next() {
		var codigo, tipoCadastro, codigoCadastro, codigoRegiaoDetalhe, codigoCidadeEstado, cadastroLJ, cadastroCF, alteracaoLJ, alteracaoCF sql.NullInt64
		var endereco, numero, observacao, cep sql.NullString
		var cadastroDT, alteracaoDT sql.NullTime

		err := rows.Scan(&codigo, &tipoCadastro, &codigoCadastro, &endereco, &numero, &observacao, &codigoRegiaoDetalhe, &codigoCidadeEstado, &cep, &cadastroLJ, &cadastroCF, &cadastroDT, &alteracaoLJ, &alteracaoCF, &alteracaoDT)
		if err != nil {
			logger.Warn("Erro ao escanear endereço de entrega: %v", err)
			continue
		}

		_, err = stmt.Exec(codigo, tipoCadastro, codigoCadastro, endereco, numero, observacao, codigoRegiaoDetalhe, codigoCidadeEstado, cep, cadastroLJ, cadastroCF, cadastroDT, alteracaoLJ, alteracaoCF, alteracaoDT)
		if err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "unique") && !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				logger.Warn("Erro ao inserir endereço de entrega: %v", err)
			}
			continue
		}
		contadorEnderecos++
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("erro ao commitar transação de endereços de entrega: %w", err)
	}
	logger.Info("✅ %d endereços de entrega inseridos", contadorEnderecos)

	// 9. Insere endereços principais dos clientes
	logger.Info("Inserindo endereços principais dos clientes...")
	enderecosPrincipaisQuery := `
		SELECT
			GEN_ID(GEN_CADASTRO_ENDERECO, 1) AS CODIGO,
			1 AS TIPO_CADASTRO,
			CAST(C.CODIGO_PS AS INTEGER) AS CODIGO_CADASTRO,
			SUBSTRING(
				UPPER(
					COALESCE(C.ENDERECOCLIENTE, '') ||
					CASE
						WHEN C.ENDERECOCLIENTE IS NOT NULL AND C.ENDERECOCLIENTE <> '' 
							 AND B.NOMEBAIRRO IS NOT NULL AND B.NOMEBAIRRO <> ''
							THEN ', ' || B.NOMEBAIRRO
						WHEN B.NOMEBAIRRO IS NOT NULL AND B.NOMEBAIRRO <> ''
							THEN B.NOMEBAIRRO
						ELSE ''
					END ||
					CASE
						WHEN (C.ENDERECOCLIENTE IS NOT NULL AND C.ENDERECOCLIENTE <> '') 
							 OR (B.NOMEBAIRRO IS NOT NULL AND B.NOMEBAIRRO <> '')
							THEN
								CASE
									WHEN CC.NOMECIDADE IS NOT NULL AND CC.NOMECIDADE <> '' 
										THEN ', ' || CC.NOMECIDADE
									ELSE ''
								END
						WHEN CC.NOMECIDADE IS NOT NULL AND CC.NOMECIDADE <> '' 
							THEN CC.NOMECIDADE
						ELSE ''
					END
				) FROM 1 FOR 100
			) AS ENDERECO,
			COALESCE(C.NUMEROENDERECOCLIENTE, 'S/N') AS NUMERO,
			UPPER(COALESCE(C.COMPLEMENTOCLIENTE, '') || IIF(C.COMPLEMENTOCLIENTE <> '', ', ', '') || COALESCE(C.PROXIMIDADECLIENTE, '')) AS OBSERVACAO,
			1 AS CODIGO_REGIAODETALHE,
			IIF(CC.CODIGO_PS IS NOT NULL, CC.CODIGO_PS, 1) AS CODIGO_CIDADEESTADO,
			COALESCE(C.CEPCLIENTE, '00000000') AS CEP,
			1 AS CADASTRO_LJ,
			1 AS CADASTRO_CF,
			C.DT_CREATION AS CADASTRO_DT,
			1 AS ALTERACAO_LJ,
			1 AS ALTERACAO_CF,
			C.DT_ALTER AS ALTERACAO_DT
		FROM CLIENTE C
			LEFT JOIN CIDADE CC ON CC.CODIGOCIDADE = C.CODIGOCIDADE
			LEFT JOIN BAIRRO B ON B.CODIGOBAIRRO = C.CODIGOBAIRRO
		WHERE C.ENDERECOCLIENTE <> ''
			AND NOT EXISTS (
				SELECT 1
				FROM CLIENTE_ENDERECO_ENTREGA CEE2
				WHERE CEE2.CODIGOCLIENTE = C.CODIGOCLIENTE
					AND CEE2.DS_ENDERECO <> ''
			)
			AND C.CONVERSAO = ?
	`

	tx, err = pharmacieDB.Begin()
	if err != nil {
		return nil, fmt.Errorf("erro ao iniciar transação para endereços principais: %w", err)
	}
	defer tx.Rollback()

	stmt, err = tx.Prepare(insertEnderecoSQL)
	if err != nil {
		return nil, fmt.Errorf("erro ao preparar statement de endereços principais: %w", err)
	}
	defer stmt.Close()

	rows, err = prisma5DB.Query(enderecosPrincipaisQuery, conversaoInt)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar endereços principais: %w", err)
	}
	defer rows.Close()

	contadorEnderecosPrincipais := 0
	for rows.Next() {
		var codigo, tipoCadastro, codigoCadastro, codigoRegiaoDetalhe, codigoCidadeEstado, cadastroLJ, cadastroCF, alteracaoLJ, alteracaoCF sql.NullInt64
		var endereco, numero, observacao, cep sql.NullString
		var cadastroDT, alteracaoDT sql.NullTime

		err := rows.Scan(&codigo, &tipoCadastro, &codigoCadastro, &endereco, &numero, &observacao, &codigoRegiaoDetalhe, &codigoCidadeEstado, &cep, &cadastroLJ, &cadastroCF, &cadastroDT, &alteracaoLJ, &alteracaoCF, &alteracaoDT)
		if err != nil {
			logger.Warn("Erro ao escanear endereço principal: %v", err)
			continue
		}

		_, err = stmt.Exec(codigo, tipoCadastro, codigoCadastro, endereco, numero, observacao, codigoRegiaoDetalhe, codigoCidadeEstado, cep, cadastroLJ, cadastroCF, cadastroDT, alteracaoLJ, alteracaoCF, alteracaoDT)
		if err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "unique") && !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				logger.Warn("Erro ao inserir endereço principal: %v", err)
			}
			continue
		}
		contadorEnderecosPrincipais++
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("erro ao commitar transação de endereços principais: %w", err)
	}
	logger.Info("✅ %d endereços principais inseridos", contadorEnderecosPrincipais)

	// 10. Insere telefones dos clientes
	insertTelefoneSQL := `INSERT INTO CADASTRO_TELEFONE (CODIGO, TIPO_CADASTRO, CODIGO_CADASTRO, TELEFONE_TIPO, TELEFONEPREFIXO, TELEFONE, CADASTRO_LJ, CADASTRO_CF, CADASTRO_DT, ALTERACAO_LJ, ALTERACAO_CF, ALTERACAO_DT) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	// 10.1. Telefones de entrega
	logger.Info("Inserindo telefones de entrega...")
	telefonesEntregaQuery := `
		SELECT
			CAST(CC.CODIGO_PS AS INTEGER) AS CODIGO_CADASTRO,
			CEE.NR_TELEFONE,
			CEE.DDD_TELEFONE,
			CEE.DT_CREATION AS CADASTRO_DT,
			CEE.DT_ALTER AS ALTERACAO_DT
		FROM CLIENTE_ENDERECO_ENTREGA CEE
			INNER JOIN CLIENTE CC ON CC.CODIGOCLIENTE = CEE.CODIGOCLIENTE
		WHERE CEE.NR_TELEFONE <> ''
			AND CC.CONVERSAO = ?
	`

	tx, err = pharmacieDB.Begin()
	if err != nil {
		return nil, fmt.Errorf("erro ao iniciar transação para telefones de entrega: %w", err)
	}
	defer tx.Rollback()

	stmt, err = tx.Prepare(insertTelefoneSQL)
	if err != nil {
		return nil, fmt.Errorf("erro ao preparar statement de telefones de entrega: %w", err)
	}
	defer stmt.Close()

	rows, err = prisma5DB.Query(telefonesEntregaQuery, conversaoInt)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar telefones de entrega: %w", err)
	}
	defer rows.Close()

	contadorTelefonesEntrega := 0
	for rows.Next() {
		var codigoCadastro sql.NullInt64
		var nrTelefone, dddTelefone sql.NullString
		var cadastroDT, alteracaoDT sql.NullTime

		err := rows.Scan(&codigoCadastro, &nrTelefone, &dddTelefone, &cadastroDT, &alteracaoDT)
		if err != nil {
			logger.Warn("Erro ao escanear telefone de entrega: %v", err)
			continue
		}

		if !codigoCadastro.Valid || !nrTelefone.Valid || nrTelefone.String == "" {
			continue
		}

		// Processar telefone
		telefoneLimpo := stripNonNumeric(nrTelefone.String)
		if telefoneLimpo == "" {
			continue
		}

		// Gerar código do telefone
		var codigoTelefone int
		err = prisma5DB.QueryRow(`SELECT GEN_ID(GEN_CADASTRO_TELEFONE, 1) FROM RDB$DATABASE`).Scan(&codigoTelefone)
		if err != nil {
			logger.Warn("Erro ao gerar código do telefone de entrega: %v", err)
			continue
		}

		// Determinar tipo de telefone (se primeiro dígito entre '0' e '5' = tipo 1, senão = tipo 3)
		telefoneTipo := 3
		if len(telefoneLimpo) > 0 {
			primeiroDigito := telefoneLimpo[0]
			if primeiroDigito >= '0' && primeiroDigito <= '5' {
				telefoneTipo = 1
			}
		}

		// Processar DDD
		ddd := ""
		if dddTelefone.Valid && dddTelefone.String != "" {
			dddLimpo := strings.TrimSpace(dddTelefone.String)
			if len(dddLimpo) > 0 && dddLimpo[0] == '0' {
				if len(dddLimpo) > 2 {
					ddd = dddLimpo[1:3]
				}
			} else if len(dddLimpo) >= 2 {
				ddd = dddLimpo[:2]
			}
		}

		// Limitar tamanho do telefone
		if len(telefoneLimpo) > 18 {
			telefoneLimpo = telefoneLimpo[:18]
		}

		_, err = stmt.Exec(codigoTelefone, 1, codigoCadastro.Int64, telefoneTipo, ddd, telefoneLimpo, 1, 1, cadastroDT, 1, 1, alteracaoDT)
		if err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "unique") && !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				logger.Warn("Erro ao inserir telefone de entrega: %v", err)
			}
			continue
		}
		contadorTelefonesEntrega++
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("erro ao commitar transação de telefones de entrega: %w", err)
	}
	logger.Info("✅ %d telefones de entrega inseridos", contadorTelefonesEntrega)

	// 10.2. Celulares de entrega
	logger.Info("Inserindo celulares de entrega...")
	celularesEntregaQuery := `
		SELECT
			CAST(CC.CODIGO_PS AS INTEGER) AS CODIGO_CADASTRO,
			CEE.NR_CELULAR,
			CEE.DDD_CELULAR,
			CEE.DT_CREATION AS CADASTRO_DT,
			CEE.DT_ALTER AS ALTERACAO_DT
		FROM CLIENTE_ENDERECO_ENTREGA CEE
			INNER JOIN CLIENTE CC ON CC.CODIGOCLIENTE = CEE.CODIGOCLIENTE
		WHERE CEE.NR_CELULAR <> ''
			AND CC.CONVERSAO = ?
	`

	tx, err = pharmacieDB.Begin()
	if err != nil {
		return nil, fmt.Errorf("erro ao iniciar transação para celulares de entrega: %w", err)
	}
	defer tx.Rollback()

	stmt, err = tx.Prepare(insertTelefoneSQL)
	if err != nil {
		return nil, fmt.Errorf("erro ao preparar statement de celulares de entrega: %w", err)
	}
	defer stmt.Close()

	rows, err = prisma5DB.Query(celularesEntregaQuery, conversaoInt)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar celulares de entrega: %w", err)
	}
	defer rows.Close()

	contadorCelularesEntrega := 0
	for rows.Next() {
		var codigoCadastro sql.NullInt64
		var nrCelular, dddCelular sql.NullString
		var cadastroDT, alteracaoDT sql.NullTime

		err := rows.Scan(&codigoCadastro, &nrCelular, &dddCelular, &cadastroDT, &alteracaoDT)
		if err != nil {
			logger.Warn("Erro ao escanear celular de entrega: %v", err)
			continue
		}

		if !codigoCadastro.Valid || !nrCelular.Valid || nrCelular.String == "" {
			continue
		}

		// Processar celular
		celularLimpo := stripNonNumeric(nrCelular.String)
		if celularLimpo == "" {
			continue
		}

		// Gerar código do telefone
		var codigoTelefone int
		err = prisma5DB.QueryRow(`SELECT GEN_ID(GEN_CADASTRO_TELEFONE, 1) FROM RDB$DATABASE`).Scan(&codigoTelefone)
		if err != nil {
			logger.Warn("Erro ao gerar código do celular de entrega: %v", err)
			continue
		}

		// Determinar tipo de telefone (se primeiro dígito entre '0' e '5' = tipo 1, senão = tipo 3)
		telefoneTipo := 3
		if len(celularLimpo) > 0 {
			primeiroDigito := celularLimpo[0]
			if primeiroDigito >= '0' && primeiroDigito <= '5' {
				telefoneTipo = 1
			}
		}

		// Processar DDD
		ddd := ""
		if dddCelular.Valid && dddCelular.String != "" {
			dddLimpo := strings.TrimSpace(dddCelular.String)
			if len(dddLimpo) > 0 && dddLimpo[0] == '0' {
				if len(dddLimpo) > 2 {
					ddd = dddLimpo[1:3]
				}
			} else if len(dddLimpo) >= 2 {
				ddd = dddLimpo[:2]
			}
		}

		// Limitar tamanho do celular
		if len(celularLimpo) > 18 {
			celularLimpo = celularLimpo[:18]
		}

		_, err = stmt.Exec(codigoTelefone, 1, codigoCadastro.Int64, telefoneTipo, ddd, celularLimpo, 1, 1, cadastroDT, 1, 1, alteracaoDT)
		if err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "unique") && !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				logger.Warn("Erro ao inserir celular de entrega: %v", err)
			}
			continue
		}
		contadorCelularesEntrega++
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("erro ao commitar transação de celulares de entrega: %w", err)
	}
	logger.Info("✅ %d celulares de entrega inseridos", contadorCelularesEntrega)

	// 10.3. Telefones principais dos clientes
	logger.Info("Inserindo telefones principais dos clientes...")
	telefonesPrincipaisQuery := `
		SELECT
			CAST(CC.CODIGO_PS AS INTEGER) AS CODIGO_CADASTRO,
			CC.TELEFONECLIENTE,
			CC.DDDTELEFONECLIENTE,
			CC.DT_CREATION AS CADASTRO_DT,
			CC.DT_ALTER AS ALTERACAO_DT
		FROM CLIENTE CC
		WHERE CC.TELEFONECLIENTE <> ''
			AND NOT EXISTS (
				SELECT 1
				FROM CLIENTE_ENDERECO_ENTREGA CEE2
				WHERE CEE2.CODIGOCLIENTE = CC.CODIGOCLIENTE
					AND CEE2.NR_TELEFONE <> ''
			)
			AND CC.CONVERSAO = ?
	`

	tx, err = pharmacieDB.Begin()
	if err != nil {
		return nil, fmt.Errorf("erro ao iniciar transação para telefones principais: %w", err)
	}
	defer tx.Rollback()

	stmt, err = tx.Prepare(insertTelefoneSQL)
	if err != nil {
		return nil, fmt.Errorf("erro ao preparar statement de telefones principais: %w", err)
	}
	defer stmt.Close()

	rows, err = prisma5DB.Query(telefonesPrincipaisQuery, conversaoInt)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar telefones principais: %w", err)
	}
	defer rows.Close()

	contadorTelefonesPrincipais := 0
	for rows.Next() {
		var codigoCadastro sql.NullInt64
		var telefoneCliente, dddTelefoneCliente sql.NullString
		var cadastroDT, alteracaoDT sql.NullTime

		err := rows.Scan(&codigoCadastro, &telefoneCliente, &dddTelefoneCliente, &cadastroDT, &alteracaoDT)
		if err != nil {
			logger.Warn("Erro ao escanear telefone principal: %v", err)
			continue
		}

		if !codigoCadastro.Valid || !telefoneCliente.Valid || telefoneCliente.String == "" {
			continue
		}

		// Processar telefone
		telefoneLimpo := stripNonNumeric(telefoneCliente.String)
		if telefoneLimpo == "" {
			continue
		}

		// Gerar código do telefone
		var codigoTelefone int
		err = prisma5DB.QueryRow(`SELECT GEN_ID(GEN_CADASTRO_TELEFONE, 1) FROM RDB$DATABASE`).Scan(&codigoTelefone)
		if err != nil {
			logger.Warn("Erro ao gerar código do telefone principal: %v", err)
			continue
		}

		// Determinar tipo de telefone (se primeiro dígito entre '0' e '5' = tipo 1, senão = tipo 3)
		telefoneTipo := 3
		if len(telefoneLimpo) > 0 {
			primeiroDigito := telefoneLimpo[0]
			if primeiroDigito >= '0' && primeiroDigito <= '5' {
				telefoneTipo = 1
			}
		}

		// Processar DDD
		ddd := ""
		if dddTelefoneCliente.Valid && dddTelefoneCliente.String != "" {
			dddLimpo := strings.TrimSpace(dddTelefoneCliente.String)
			if len(dddLimpo) > 0 && dddLimpo[0] == '0' {
				if len(dddLimpo) > 2 {
					ddd = dddLimpo[1:3]
				}
			} else if len(dddLimpo) >= 2 {
				ddd = dddLimpo[:2]
			}
		}

		// Limitar tamanho do telefone
		if len(telefoneLimpo) > 18 {
			telefoneLimpo = telefoneLimpo[:18]
		}

		_, err = stmt.Exec(codigoTelefone, 1, codigoCadastro.Int64, telefoneTipo, ddd, telefoneLimpo, 1, 1, cadastroDT, 1, 1, alteracaoDT)
		if err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "unique") && !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				logger.Warn("Erro ao inserir telefone principal: %v", err)
			}
			continue
		}
		contadorTelefonesPrincipais++
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("erro ao commitar transação de telefones principais: %w", err)
	}
	logger.Info("✅ %d telefones principais inseridos", contadorTelefonesPrincipais)

	// 10.4. Celulares principais dos clientes
	logger.Info("Inserindo celulares principais dos clientes...")
	celularesPrincipaisQuery := `
		SELECT
			CAST(CC.CODIGO_PS AS INTEGER) AS CODIGO_CADASTRO,
			CC.CELULARCLIENTE,
			CC.DDDCELULARCLIENTE,
			CC.DDDTELEFONECLIENTE,
			CC.DT_CREATION AS CADASTRO_DT,
			CC.DT_ALTER AS ALTERACAO_DT
		FROM CLIENTE CC
		WHERE CC.CELULARCLIENTE <> ''
			AND NOT EXISTS (
				SELECT 1
				FROM CLIENTE_ENDERECO_ENTREGA CEE3
				WHERE CEE3.CODIGOCLIENTE = CC.CODIGOCLIENTE
					AND CEE3.NR_CELULAR <> ''
			)
			AND CC.CONVERSAO = ?
	`

	tx, err = pharmacieDB.Begin()
	if err != nil {
		return nil, fmt.Errorf("erro ao iniciar transação para celulares principais: %w", err)
	}
	defer tx.Rollback()

	stmt, err = tx.Prepare(insertTelefoneSQL)
	if err != nil {
		return nil, fmt.Errorf("erro ao preparar statement de celulares principais: %w", err)
	}
	defer stmt.Close()

	rows, err = prisma5DB.Query(celularesPrincipaisQuery, conversaoInt)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar celulares principais: %w", err)
	}
	defer rows.Close()

	contadorCelularesPrincipais := 0
	for rows.Next() {
		var codigoCadastro sql.NullInt64
		var celularCliente, dddCelularCliente, dddTelefoneCliente sql.NullString
		var cadastroDT, alteracaoDT sql.NullTime

		err := rows.Scan(&codigoCadastro, &celularCliente, &dddCelularCliente, &dddTelefoneCliente, &cadastroDT, &alteracaoDT)
		if err != nil {
			logger.Warn("Erro ao escanear celular principal: %v", err)
			continue
		}

		if !codigoCadastro.Valid || !celularCliente.Valid || celularCliente.String == "" {
			continue
		}

		// Processar celular
		celularLimpo := stripNonNumeric(celularCliente.String)
		if celularLimpo == "" {
			continue
		}

		// Gerar código do telefone
		var codigoTelefone int
		err = prisma5DB.QueryRow(`SELECT GEN_ID(GEN_CADASTRO_TELEFONE, 1) FROM RDB$DATABASE`).Scan(&codigoTelefone)
		if err != nil {
			logger.Warn("Erro ao gerar código do celular principal: %v", err)
			continue
		}

		// Determinar tipo de telefone (se primeiro dígito entre '0' e '5' = tipo 1, senão = tipo 3)
		telefoneTipo := 3
		if len(celularLimpo) > 0 {
			primeiroDigito := celularLimpo[0]
			if primeiroDigito >= '0' && primeiroDigito <= '5' {
				telefoneTipo = 1
			}
		}

		// Processar DDD (usar DDDCELULARCLIENTE se existir e começar com '0', senão usar DDDTELEFONECLIENTE)
		ddd := ""
		if dddCelularCliente.Valid && dddCelularCliente.String != "" {
			dddLimpo := strings.TrimSpace(dddCelularCliente.String)
			if len(dddLimpo) > 0 && dddLimpo[0] == '0' {
				if len(dddLimpo) > 2 {
					ddd = dddLimpo[1:3]
				}
			}
		}
		// Se DDD não foi definido, usar DDDTELEFONECLIENTE
		if ddd == "" && dddTelefoneCliente.Valid && dddTelefoneCliente.String != "" {
			dddLimpo := strings.TrimSpace(dddTelefoneCliente.String)
			if len(dddLimpo) >= 2 {
				ddd = dddLimpo[:2]
			}
		}

		// Limitar tamanho do celular
		if len(celularLimpo) > 18 {
			celularLimpo = celularLimpo[:18]
		}

		_, err = stmt.Exec(codigoTelefone, 1, codigoCadastro.Int64, telefoneTipo, ddd, celularLimpo, 1, 1, cadastroDT, 1, 1, alteracaoDT)
		if err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "unique") && !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				logger.Warn("Erro ao inserir celular principal: %v", err)
			}
			continue
		}
		contadorCelularesPrincipais++
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("erro ao commitar transação de celulares principais: %w", err)
	}
	logger.Info("✅ %d celulares principais inseridos", contadorCelularesPrincipais)
	logger.Info("✅ Migração de clientes concluída com sucesso!")

	return outStats, nil
}

func hashClienteMigracao(nomeCliente, cpfCnpj sql.NullString, codigo sql.NullInt64) string {
	var b strings.Builder
	if codigo.Valid {
		fmt.Fprintf(&b, "%d|", codigo.Int64)
	}
	if nomeCliente.Valid {
		b.WriteString(strings.TrimSpace(nomeCliente.String))
		b.WriteByte('|')
	}
	if cpfCnpj.Valid {
		b.WriteString(strings.TrimSpace(cpfCnpj.String))
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}
