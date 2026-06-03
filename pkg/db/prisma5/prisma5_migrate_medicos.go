package prisma5

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/primesoftwaresi/prime-migration/pkg/logger"
)

// MigratePRISMA5Medicos migra médicos do PRISMA5 para o Pharmacie
// Parâmetro conversao: valor da conversão (ex: "1", "2")
func MigratePRISMA5Medicos(prisma5DB *sql.DB, conversao string) error {
	logger.Info("Iniciando migração de médicos (conversão: %s)", conversao)

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

	// 1. Atualiza campo CONVERSAO dos médicos
	logger.Info("Atualizando campo CONVERSAO dos médicos...")
	_, err = prisma5DB.Exec(`
		UPDATE MEDICO M 
		SET M.CONVERSAO = ? 
		WHERE M.CODIGO_PS IS NULL
	`, conversaoInt)
	if err != nil {
		return fmt.Errorf("erro ao atualizar CONVERSAO: %w", err)
	}

	// 2. Gera CODIGO_PS para médicos usando sequence (OTIMIZADO: UPDATE em massa)
	// Nota: A procedure SQL original usa ORDER BY M.CODIGOMEDICO, mas como o Firebird
	// gera os valores de sequence sequencialmente e o ORDER BY em UPDATE não é padrão SQL,
	// a ordem final será preservada mesmo sem ORDER BY explicitamente
	logger.Info("Gerando CODIGO_PS para médicos...")
	if err := generateCodigoPSInBatch(prisma5DB, "MEDICO", "CODIGOMEDICO", "GEN_MEDICO", 0); err != nil {
		return fmt.Errorf("erro ao gerar CODIGO_PS para médicos: %w", err)
	}

	// 3. Insere especialidades no banco Pharmacie
	logger.Info("Inserindo especialidades no Pharmacie...")
	especialidadesQuery := `
		SELECT
			CAST(E.CODIGOESPECIALIDADE AS INTEGER) AS CODIGO,
			UPPER(E.DESCRICAOESPECIALIDADE) AS NOMEESPECIALIDADE,
			1 AS CADASTRO_LJ,
			1 AS CADASTRO_CF,
			E.DT_CREATION AS CADASTRO_DT,
			1 AS ALTERACAO_LJ,
			1 AS ALTERACAO_CF,
			E.DT_ALTER AS ALTERACAO_DT
		FROM ESPECIALIDADE E
		WHERE E.CODIGOESPECIALIDADE <> 1
		ORDER BY 1
	`

	rows, err := prisma5DB.Query(especialidadesQuery)
	if err != nil {
		return fmt.Errorf("erro ao buscar especialidades: %w", err)
	}
	defer rows.Close()

	insertEspecialidadeSQL := `INSERT INTO MEDICO_ESPECIALIDADE (CODIGO, NOMEESPECIALIDADE, CADASTRO_LJ, CADASTRO_CF, CADASTRO_DT, ALTERACAO_LJ, ALTERACAO_CF, ALTERACAO_DT) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	tx, err := pharmacieDB.Begin()
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação para especialidades: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(insertEspecialidadeSQL)
	if err != nil {
		return fmt.Errorf("erro ao preparar statement de especialidades: %w", err)
	}
	defer stmt.Close()

	contadorEspecialidades := 0
	for rows.Next() {
		var codigo sql.NullInt64
		var nomeEspecialidade sql.NullString
		var cadastroLJ, cadastroCF, alteracaoLJ, alteracaoCF sql.NullInt64
		var cadastroDT, alteracaoDT sql.NullTime

		err := rows.Scan(&codigo, &nomeEspecialidade, &cadastroLJ, &cadastroCF, &cadastroDT, &alteracaoLJ, &alteracaoCF, &alteracaoDT)
		if err != nil {
			logger.Warn("Erro ao escanear especialidade: %v", err)
			continue
		}

		if !codigo.Valid {
			continue
		}

		_, err = stmt.Exec(codigo.Int64, nomeEspecialidade, cadastroLJ, cadastroCF, cadastroDT, alteracaoLJ, alteracaoCF, alteracaoDT)
		if err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "unique") && !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				logger.Warn("Erro ao inserir especialidade %d: %v", codigo.Int64, err)
			}
			continue
		}
		contadorEspecialidades++
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("erro ao commitar transação de especialidades: %w", err)
	}
	logger.Info("✅ %d especialidades inseridas", contadorEspecialidades)

	// 4. Insere médicos no banco Pharmacie
	logger.Info("Inserindo médicos no Pharmacie...")
	medicosQuery := `
		SELECT
			M.CODIGO_PS AS CODIGO,
			UPPER(M.NOMEMEDICO) AS NOMEMEDICO,
			700000 AS CODIGO_GRUPO,
			COALESCE((SELECT FIRST 1 CAST(EM.CODIGOESPECIALIDADE AS INTEGER) FROM ESPECIALIDADEMEDICO EM WHERE EM.CODIGOMEDICO = M.CODIGOMEDICO), 1) AS CODIGO_MEDICO_ESPECIALIDADE,
			CASE M.TIPOCRMEDICO
				WHEN 'B' THEN 12
				WHEN 'BM' THEN 12
				WHEN 'BR' THEN 1
				WHEN 'C' THEN 1
				WHEN 'CO' THEN 6
				WHEN 'CR' THEN 8
				WHEN 'E' THEN 1
				WHEN 'EF' THEN 1
				WHEN 'EN' THEN 1
				WHEN 'ES' THEN 1
				WHEN 'F' THEN 7
				WHEN 'FI' THEN 9
				WHEN 'FT' THEN 8
				WHEN 'HO' THEN 10
				WHEN 'IT' THEN 1
				WHEN 'M' THEN 1
				WHEN 'MB' THEN 9
				WHEN 'ME' THEN 1
				WHEN 'MV' THEN 2
				WHEN 'N' THEN 5
				WHEN 'O' THEN 3
				WHEN 'P' THEN 4
				WHEN 'R' THEN 9
				WHEN 'RF' THEN 13
				WHEN 'RM' THEN 9
				WHEN 'RN' THEN 5
				WHEN 'RO' THEN 3
				WHEN 'RP' THEN 4
				WHEN 'RT' THEN 6
				WHEN 'T' THEN 6
				WHEN 'TE' THEN 6
				WHEN 'TH' THEN 6
				WHEN 'V' THEN 2
				ELSE 1
			END AS CODIGO_CONSELHO_REGIONAL,
			M.CRMMEDICO AS CRM,
			M.SIGLAESTADOCRMMEDICO AS CR_ESTADO,
			CASE M.GENERO
				WHEN 'M' THEN 1
				WHEN 'F' THEN 2
				ELSE 1
			END AS SEXO,
			IIF(M.EMAILMEDICO = '', NULL, LOWER(M.EMAILMEDICO)) AS EMAIL,
			COALESCE(EXTRACT(DAY FROM M.DT_NASCIMENTO), 0) AS DIANASCIMENTO,
			COALESCE(EXTRACT(MONTH FROM M.DT_NASCIMENTO), 0) AS MESNASCIMENTO,
			COALESCE(EXTRACT(YEAR FROM M.DT_NASCIMENTO), 0) AS ANONASCIMENTO,
			-1 AS ATIVO,
			1 AS CADASTRO_LJ,
			1 AS CADASTRO_CF,
			M.DT_CREATION AS CADASTRO_DT,
			1 AS ALTERACAO_LJ,
			1 AS ALTERACAO_CF,
			M.DT_ALTER AS ALTERACAO_DT
		FROM MEDICO M
		WHERE M.CONVERSAO = ?
		ORDER BY 1
	`

	rows, err = prisma5DB.Query(medicosQuery, conversaoInt)
	if err != nil {
		return fmt.Errorf("erro ao buscar médicos para inserir: %w", err)
	}
	defer rows.Close()

	insertMedicoSQL := `INSERT INTO MEDICO (CODIGO, NOMEMEDICO, CODIGO_GRUPO, CODIGO_MEDICO_ESPECIALIDADE, CODIGO_CONSELHO_REGIONAL, CRM, CR_ESTADO, SEXO, EMAIL, DIANASCIMENTO, MESNASCIMENTO, ANONASCIMENTO, ATIVO, CADASTRO_LJ, CADASTRO_CF, CADASTRO_DT, ALTERACAO_LJ, ALTERACAO_CF, ALTERACAO_DT) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	tx, err = pharmacieDB.Begin()
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação para médicos: %w", err)
	}
	defer tx.Rollback()

	stmt, err = tx.Prepare(insertMedicoSQL)
	if err != nil {
		return fmt.Errorf("erro ao preparar statement de médicos: %w", err)
	}
	defer stmt.Close()

	contadorMedicos := 0
	for rows.Next() {
		var codigo, codigoGrupo, codigoEspecialidade, codigoConselhoRegional, sexo, dianascimento, mesnascimento, anonascimento, ativo, cadastroLJ, cadastroCF, alteracaoLJ, alteracaoCF sql.NullInt64
		var nomeMedico, crm, crEstado, email sql.NullString
		var cadastroDT, alteracaoDT sql.NullTime

		err := rows.Scan(&codigo, &nomeMedico, &codigoGrupo, &codigoEspecialidade, &codigoConselhoRegional, &crm, &crEstado, &sexo, &email, &dianascimento, &mesnascimento, &anonascimento, &ativo, &cadastroLJ, &cadastroCF, &cadastroDT, &alteracaoLJ, &alteracaoCF, &alteracaoDT)
		if err != nil {
			logger.Warn("Erro ao escanear médico: %v", err)
			continue
		}

		if !codigo.Valid {
			continue
		}

		_, err = stmt.Exec(codigo, nomeMedico, codigoGrupo, codigoEspecialidade, codigoConselhoRegional, crm, crEstado, sexo, email, dianascimento, mesnascimento, anonascimento, ativo, cadastroLJ, cadastroCF, cadastroDT, alteracaoLJ, alteracaoCF, alteracaoDT)
		if err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "unique") && !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				logger.Warn("Erro ao inserir médico %d: %v", codigo.Int64, err)
			}
			continue
		}
		contadorMedicos++
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("erro ao commitar transação de médicos: %w", err)
	}
	logger.Info("✅ %d médicos inseridos no Pharmacie", contadorMedicos)

	// 5. Insere endereços dos médicos
	logger.Info("Inserindo endereços dos médicos...")
	enderecosMedicosQuery := `
		SELECT
			M.CODIGO_PS AS CODIGO_CADASTRO,
			M.ENDERECOMEDICO,
			M.NUMEROENDERECOMEDICO,
			M.COMPLEMENTOMEDICO,
			M.CEPMEDICO,
			B.NOMEBAIRRO,
			C.NOMECIDADE,
			C.CODIGO_PS AS CODIGO_CIDADE_PS,
			M.DT_CREATION,
			M.DT_ALTER
		FROM MEDICO M
			LEFT JOIN CIDADE C ON C.CODIGOCIDADE = M.CODIGOCIDADE
			LEFT JOIN BAIRRO B ON B.CODIGOBAIRRO = M.CODIGOBAIRRO
		WHERE M.ENDERECOMEDICO NOT IN ('', '.', ',', '*', '+', '+.')
		AND M.CONVERSAO = ?
	`

	rows, err = prisma5DB.Query(enderecosMedicosQuery, conversaoInt)
	if err != nil {
		return fmt.Errorf("erro ao buscar endereços dos médicos: %w", err)
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
		var enderecoMedico, numeroEnderecoMedico, complementoMedico, cepMedico, nomeBairro, nomeCidade sql.NullString
		var codigoCidadePS sql.NullInt64
		var dtCreation, dtAlter sql.NullTime

		if err := rows.Scan(&codigoCadastro, &enderecoMedico, &numeroEnderecoMedico, &complementoMedico, &cepMedico, &nomeBairro, &nomeCidade, &codigoCidadePS, &dtCreation, &dtAlter); err != nil {
			logger.Warn("Erro ao escanear endereço do médico: %v", err)
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
		if enderecoMedico.Valid && enderecoMedico.String != "" {
			enderecoCompleto = strings.ToUpper(enderecoMedico.String)
		}
		if nomeBairro.Valid && nomeBairro.String != "" {
			if enderecoCompleto != "" {
				enderecoCompleto += ", " + strings.ToUpper(nomeBairro.String)
			} else {
				enderecoCompleto = strings.ToUpper(nomeBairro.String)
			}
		}

		numero := "S/N"
		if numeroEnderecoMedico.Valid && numeroEnderecoMedico.String != "" && strings.TrimSpace(numeroEnderecoMedico.String) != "" {
			numero = numeroEnderecoMedico.String
		}

		observacao := ""
		if complementoMedico.Valid && complementoMedico.String != "" && strings.TrimSpace(complementoMedico.String) != "" {
			observacao = strings.ToUpper(complementoMedico.String)
		}

		cep := "00000000"
		if cepMedico.Valid && cepMedico.String != "" && cepMedico.String != "00000000" {
			cep = cepMedico.String
		}

		codigoCidadeEstado := int64(1)
		if codigoCidadePS.Valid {
			codigoCidadeEstado = codigoCidadePS.Int64
		}

		_, err = stmt.Exec(codigoEndereco, 2, codigoCadastro.Int64, enderecoCompleto, numero, observacao, 1, codigoCidadeEstado, cep, 1, 1, dtCreation, 1, 1, dtAlter)
		if err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "unique") && !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				logger.Warn("Erro ao inserir endereço do médico %d: %v", codigoEndereco, err)
			}
			continue
		}
		contadorEnderecos++
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("erro ao commitar transação de endereços: %w", err)
	}
	logger.Info("✅ %d endereços de médicos inseridos", contadorEnderecos)

	// 6. Insere telefones fixos dos médicos
	logger.Info("Inserindo telefones fixos dos médicos...")
	telefonesMedicosQuery := `
		SELECT
			M.CODIGO_PS AS CODIGO_CADASTRO,
			M.TELEFONEMEDICO,
			M.DDDTELEFONEMEDICO,
			M.DT_CREATION,
			M.DT_ALTER
		FROM MEDICO M
		WHERE M.CONVERSAO = ?
	`

	rows, err = prisma5DB.Query(telefonesMedicosQuery, conversaoInt)
	if err != nil {
		return fmt.Errorf("erro ao buscar telefones dos médicos: %w", err)
	}
	defer rows.Close()

	insertTelefoneSQL := `INSERT INTO CADASTRO_TELEFONE (CODIGO, TIPO_CADASTRO, CODIGO_CADASTRO, TELEFONE_TIPO, TELEFONEPREFIXO, TELEFONE, CADASTRO_LJ, CADASTRO_CF, CADASTRO_DT, ALTERACAO_LJ, ALTERACAO_CF, ALTERACAO_DT) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	tx, err = pharmacieDB.Begin()
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação para telefones: %w", err)
	}
	defer tx.Rollback()

	stmt, err = tx.Prepare(insertTelefoneSQL)
	if err != nil {
		return fmt.Errorf("erro ao preparar statement de telefones: %w", err)
	}
	defer stmt.Close()

	contadorTelefones := 0
	for rows.Next() {
		var codigoCadastro sql.NullInt64
		var telefoneMedico, dddTelefoneMedico sql.NullString
		var dtCreation, dtAlter sql.NullTime

		if err := rows.Scan(&codigoCadastro, &telefoneMedico, &dddTelefoneMedico, &dtCreation, &dtAlter); err != nil {
			logger.Warn("Erro ao escanear telefone do médico: %v", err)
			continue
		}

		if !codigoCadastro.Valid || !telefoneMedico.Valid || telefoneMedico.String == "" {
			continue
		}

		// Processar telefone
		telefoneLimpo := stripNonNumeric(telefoneMedico.String)
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

		// Tipo de telefone fixo = 2
		telefoneTipo := 2

		// Processar DDD
		ddd := ""
		if dddTelefoneMedico.Valid && dddTelefoneMedico.String != "" {
			dddLimpo := strings.TrimSpace(dddTelefoneMedico.String)
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

		_, err = stmt.Exec(codigoTelefone, 2, codigoCadastro.Int64, telefoneTipo, ddd, telefoneLimpo, 1, 1, dtCreation, 1, 1, dtAlter)
		if err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "unique") && !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				logger.Warn("Erro ao inserir telefone do médico %d: %v", codigoTelefone, err)
			}
			continue
		}
		contadorTelefones++
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("erro ao commitar transação de telefones: %w", err)
	}
	logger.Info("✅ %d telefones fixos de médicos inseridos", contadorTelefones)

	// 7. Insere celulares dos médicos
	logger.Info("Inserindo celulares dos médicos...")
	celularesMedicosQuery := `
		SELECT
			M.CODIGO_PS AS CODIGO_CADASTRO,
			M.CELULARMEDICO,
			M.DDDCELULARMEDICO,
			M.DT_CREATION,
			M.DT_ALTER
		FROM MEDICO M
		WHERE M.CONVERSAO = ?
	`

	rows, err = prisma5DB.Query(celularesMedicosQuery, conversaoInt)
	if err != nil {
		return fmt.Errorf("erro ao buscar celulares dos médicos: %w", err)
	}
	defer rows.Close()

	tx, err = pharmacieDB.Begin()
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação para celulares: %w", err)
	}
	defer tx.Rollback()

	stmt, err = tx.Prepare(insertTelefoneSQL)
	if err != nil {
		return fmt.Errorf("erro ao preparar statement de celulares: %w", err)
	}
	defer stmt.Close()

	contadorCelulares := 0
	for rows.Next() {
		var codigoCadastro sql.NullInt64
		var celularMedico, dddCelularMedico sql.NullString
		var dtCreation, dtAlter sql.NullTime

		if err := rows.Scan(&codigoCadastro, &celularMedico, &dddCelularMedico, &dtCreation, &dtAlter); err != nil {
			logger.Warn("Erro ao escanear celular do médico: %v", err)
			continue
		}

		if !codigoCadastro.Valid || !celularMedico.Valid || celularMedico.String == "" {
			continue
		}

		// Processar celular
		celularLimpo := stripNonNumeric(celularMedico.String)
		if celularLimpo == "" {
			continue
		}

		// Gerar código do telefone
		var codigoTelefone int
		err = prisma5DB.QueryRow(`SELECT GEN_ID(GEN_CADASTRO_TELEFONE, 1) FROM RDB$DATABASE`).Scan(&codigoTelefone)
		if err != nil {
			logger.Warn("Erro ao gerar código do celular: %v", err)
			continue
		}

		// Tipo de celular = 3
		telefoneTipo := 3

		// Processar DDD
		ddd := ""
		if dddCelularMedico.Valid && dddCelularMedico.String != "" {
			dddLimpo := strings.TrimSpace(dddCelularMedico.String)
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

		_, err = stmt.Exec(codigoTelefone, 2, codigoCadastro.Int64, telefoneTipo, ddd, celularLimpo, 1, 1, dtCreation, 1, 1, dtAlter)
		if err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "unique") && !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
				logger.Warn("Erro ao inserir celular do médico %d: %v", codigoTelefone, err)
			}
			continue
		}
		contadorCelulares++
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("erro ao commitar transação de celulares: %w", err)
	}
	logger.Info("✅ %d celulares de médicos inseridos", contadorCelulares)

	logger.Info("✅ Migração de médicos concluída com sucesso!")
	return nil
}

