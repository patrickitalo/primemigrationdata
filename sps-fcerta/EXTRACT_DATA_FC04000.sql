create or alter procedure EXTRACT_DATA_FC04000 (
    V_CONVERSAO integer)
as
declare variable V_USER varchar(20);
declare variable V_SHA varchar(20);
declare variable CON varchar(60);
declare variable CODIGO integer;
declare variable I_SQL varchar(2000);
declare variable NOMEMEDICO varchar(255);
declare variable CRM varchar(15);
declare variable CR_ESTADO varchar(2);
declare variable SEXO varchar(2);
declare variable CODIGO_MEDICO_ESPECIALIDADE integer;
declare variable CODIGO_CONSELHO_REGIONAL integer;
declare variable CODIGO_GRUPO integer;
declare variable DIANASCIMENTO varchar(2);
declare variable MESNASCIMENTO varchar(2);
declare variable ANONASCIMENTO varchar(4);
declare variable OBSERVACAO_MED varchar(300);
declare variable ATIVO integer;
declare variable CADASTRO_LJ integer;
declare variable CADASTRO_CF integer;
declare variable CADASTRO_DT timestamp;
declare variable ALTERACAO_LJ integer;
declare variable ALTERACAO_CF integer;
declare variable ALTERACAO_DT timestamp;
declare variable CODIGO_CADASTRO integer;
declare variable CODIGO_CIDADE_PS integer;
declare variable TIPO_CADASTRO integer;
declare variable ENDERECO varchar(500);
declare variable NUMERO varchar(30);
declare variable CEP varchar(30);
declare variable OBSERVACAO_END varchar(2000);
declare variable CODIGO_REGIAODETALHE integer;
declare variable CODIGO_CIDADEESTADO integer;
declare variable TELEFONE_TIPO integer;
declare variable TELEFONEPREFIXO varchar(4);
declare variable TELEFONE varchar(30);
declare variable OBSERVACAO varchar(300);
BEGIN
    V_USER = 'SYSDBA';
     V_SHA = 'SySPs_PHARMACIE';

    /*OBTEM E MONTA STRING DE CONEXAO DO PHARMACIE*/

    SELECT FIRST 1 CN.IPSERVER || '/' || CN.PORTA || ':' || CN.ALIAS FROM CONEXAO CN INTO :CON;

    -- ATUALIZA O CAMPO CONVERSAO COM O VALOR FORNECIDO
    UPDATE FC04000 SET CONVERSAO = :V_CONVERSAO WHERE CODIGO_PS IS NULL AND CONVERSAO IS NULL;

    /* UPDATE NO QUAL GERA O CODIGO_PS NA TABELA DO MEDICO*/

    UPDATE FC04000 M
    SET M.CODIGO_PS = GEN_ID(GEN_MEDICO, 1)
    WHERE M.CODIGO_PS IS NULL AND M.CONVERSAO = :V_CONVERSAO;

    /*QUERY QUE BUSCA E FORMATA INFORMACOES DE CADASTRO DE MEDICO */
    FOR
        SELECT DISTINCT
                M.CODIGO_PS AS CODIGO,
                M.NOMEMED AS NOMEMEDICO,
                M.NRCRM AS CRM,
                M.UFCRM AS CR_ESTADO,
                CASE M.TPSEX
                    WHEN 'M' THEN 1
                    WHEN 'F' THEN 2
                END AS SEXO,
                COALESCE((SELECT FIRST 1
                    CASE ME.CDESP
                        WHEN '001' THEN 2
                        WHEN '002' THEN 3
                        WHEN '003' THEN 4
                        WHEN '004' THEN 5
                        WHEN '005' THEN 6
                        WHEN '006' THEN 7
                        WHEN '007' THEN 8
                        WHEN '008' THEN 9
                        WHEN '009' THEN 10
                        WHEN '010' THEN 11
                        WHEN '011' THEN 12
                        WHEN '012' THEN 13
                        WHEN '013' THEN 14
                        WHEN '014' THEN 15
                        WHEN '015' THEN 16
                        WHEN '016' THEN 17
                        WHEN '017' THEN 18
                        WHEN '018' THEN 19
                        WHEN '019' THEN 20
                        WHEN '020' THEN 21
                        WHEN '021' THEN 22
                        WHEN '022' THEN 23
                        WHEN '023' THEN 24
                        WHEN '024' THEN 25
                        WHEN '025' THEN 26
                        WHEN '026' THEN 27
                        WHEN '027' THEN 28
                        WHEN '028' THEN 29
                        WHEN '029' THEN 30
                        WHEN '030' THEN 31
                        WHEN '031' THEN 32
                        WHEN '032' THEN 33
                        WHEN '033' THEN 34
                        WHEN '034' THEN 35
                        WHEN '035' THEN 36
                        WHEN '036' THEN 37
                        WHEN '037' THEN 38
                        WHEN '038' THEN 39
                        WHEN '039' THEN 40
                        WHEN '040' THEN 41
                        WHEN '041' THEN 42
                        WHEN '042' THEN 43
                        WHEN '043' THEN 44
                        WHEN '044' THEN 45
                        WHEN '045' THEN 46
                        WHEN '046' THEN 47
                        WHEN '047' THEN 48
                        WHEN '048' THEN 49
                        WHEN '049' THEN 50
                        WHEN '050' THEN 51
                        WHEN '051' THEN 52
                        WHEN '052' THEN 53
                        WHEN '053' THEN 54
                        WHEN '054' THEN 55
                        WHEN '055' THEN 56
                        WHEN '056' THEN 57
                        WHEN '057' THEN 58
                        WHEN '058' THEN 59
                        WHEN '059' THEN 60
                        WHEN '060' THEN 61
                        WHEN '061' THEN 62
                        WHEN '062' THEN 63
                        WHEN '063' THEN 64
                        WHEN '064' THEN 65
                        WHEN '065' THEN 66
                        WHEN '066' THEN 67
                        WHEN '067' THEN 68
                        WHEN '068' THEN 69
                        WHEN '069' THEN 70
                        WHEN '070' THEN 71
                        WHEN '071' THEN 72
                        WHEN '072' THEN 73
                        WHEN '073' THEN 74
                        WHEN '074' THEN 75
                        WHEN '075' THEN 76
                        WHEN '076' THEN 77
                        WHEN '077' THEN 78
                        WHEN '078' THEN 79
                        WHEN '079' THEN 80
                        WHEN '080' THEN 81
                        WHEN '081' THEN 82
                        WHEN '082' THEN 83
                        WHEN '083' THEN 84
                        WHEN '084' THEN 85
                        WHEN '085' THEN 86
                        WHEN '086' THEN 87
                        WHEN '087' THEN 88
                        WHEN '088' THEN 89
                        WHEN '089' THEN 90
                        WHEN '090' THEN 91
                        WHEN '091' THEN 92
                        WHEN '092' THEN 93
                        WHEN '093' THEN 94
                        WHEN '094' THEN 95
                        WHEN '095' THEN 96
                        WHEN '096' THEN 97
                        WHEN '097' THEN 98
                        WHEN '098' THEN 99
                        WHEN '099' THEN 100
                        WHEN '100' THEN 101
                        ELSE 1
                    END
                    FROM FC04300 ME
                    WHERE M.NRCRM = ME.NRCRM AND M.UFCRM = ME.UFCRM
                   ), 1) AS CODIGO_MEDICO_ESPECIALIDADE,
                CASE M.PFCRM
                    WHEN '1' THEN 1
                    WHEN '2' THEN 3
                    WHEN '3' THEN 2
                    WHEN '5' THEN 4
                    WHEN '6' THEN 7
                    WHEN '7' THEN 12
                    WHEN '8' THEN 14
                    WHEN '9' THEN 5
                    WHEN 'A' THEN 8
                    WHEN 'B' THEN 13
                    WHEN 'C' THEN 14
                    WHEN 'D' THEN 9
                    ELSE 1
                END AS CODIGO_CONSELHO_REGIONAL,
                700000 AS CODIGO_GRUPO,
                COALESCE(EXTRACT(DAY FROM M.DTNAS), 0) AS DIANASCIMENTO,
                COALESCE(EXTRACT(MONTH FROM M.DTNAS), 0) AS MESNASCIMENTO,
                COALESCE(EXTRACT(YEAR FROM M.DTNAS), 0) AS ANONASCIMENTO,
                M.OBSERV AS OBSERVACAO,
                -1 AS ATIVO,
                1 AS CADASTRO_LJ,
                1 AS CADASTRO_CF,
                M.DTCAD AS CADASTRO_DT,
                1 AS ALTERACAO_LJ,
                1 AS ALTERACAO_CF,
                CURRENT_TIMESTAMP AS ALTERACAO_DT
            FROM FC04000 M
            LEFT JOIN FC04300 ME ON M.NRCRM = ME.NRCRM AND M.UFCRM = ME.UFCRM
            WHERE M.CODIGO_PS IS NOT NULL AND M.CONVERSAO = :V_CONVERSAO
            INTO :CODIGO, :NOMEMEDICO, :CRM, :CR_ESTADO, :SEXO, :CODIGO_MEDICO_ESPECIALIDADE, :CODIGO_CONSELHO_REGIONAL, :CODIGO_GRUPO, :DIANASCIMENTO, :MESNASCIMENTO, :ANONASCIMENTO, :OBSERVACAO_MED, :ATIVO, :CADASTRO_LJ, :CADASTRO_CF, :CADASTRO_DT, :ALTERACAO_LJ, :ALTERACAO_CF, :ALTERACAO_DT
            DO BEGIN

            /*BLOCO ONDE RECEBE AS INFORMACOES DA QUERY DE CADASTRO DE MEDICO E VAI INSERINDO OS DADOS NO BANCO DE DADOS PHARMACIE*/

            I_SQL = 'INSERT INTO MEDICO (CODIGO, NOMEMEDICO, CRM, CR_ESTADO, SEXO, CODIGO_MEDICO_ESPECIALIDADE, CODIGO_CONSELHO_REGIONAL, CODIGO_GRUPO, DIANASCIMENTO, MESNASCIMENTO, ANONASCIMENTO, OBSERVACAO, ATIVO, CADASTRO_LJ, CADASTRO_CF, CADASTRO_DT, ALTERACAO_LJ, ALTERACAO_CF, ALTERACAO_DT) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)';
            EXECUTE STATEMENT (:I_SQL) (:CODIGO, :NOMEMEDICO, :CRM, :CR_ESTADO, :SEXO, :CODIGO_MEDICO_ESPECIALIDADE, :CODIGO_CONSELHO_REGIONAL, :CODIGO_GRUPO, :DIANASCIMENTO, :MESNASCIMENTO, :ANONASCIMENTO, :OBSERVACAO_MED, :ATIVO, :CADASTRO_LJ, :CADASTRO_CF, :CADASTRO_DT, :ALTERACAO_LJ, :ALTERACAO_CF, :ALTERACAO_DT) ON EXTERNAL :CON AS USER :V_USER PASSWORD :V_SHA WITH COMMON TRANSACTION;
                END
    /*UPDATE QUE VINCULA OS CODIGOS DE CIDADES DO PHARMACIE NA TABELA DE ENDERECOS DE MEDICO*/

    BEGIN
        UPDATE FC04400 FC
        SET FC.CODIGO_CIDADE_PS = (SELECT FIRST 1 CID.CODIGO
                                  FROM CIDADEESTADO CID
                                  WHERE FC.MUNIC = CID.NOMECIDADE
                                    AND FC.UNFED = CID.UF)
        WHERE FC.CODIGO_CIDADE_PS IS NULL;

        UPDATE FC04400 FC
        SET FC.MUNIC = (SELECT RETORNO FROM FN_TIRA_ACENTO(FC.MUNIC));
    END
    BEGIN
        UPDATE FC04400 ME
        SET ME.CODIGO_PS = (SELECT M.CODIGO_PS
                       FROM FC04000 M
                       WHERE M.PFCRM = ME.PFCRM AND
                             M.UFCRM = ME.UFCRM AND
                             M.NRCRM = ME.NRCRM)
        WHERE ME.CODIGO_PS IS NULL;

    END
        FOR
           SELECT GEN_ID(GEN_CADASTRO_ENDERECO, 1) AS CODIGO,
                2 AS TIPO_CADASTRO,
                M.CODIGO_PS AS CODIGO_CADASTRO,
                SUBSTRING(IIF(MUNIC IS NOT NULL AND CODIGO_CIDADE_PS IS NULL, 
                          IIF(BAIRR IS NOT NULL AND BAIRR <> '', 
                              ENDER || ', ' || BAIRR || 
                              IIF(MUNIC IS NOT NULL AND MUNIC <> '', ', ' || MUNIC, '') || 
                              IIF(UNFED IS NOT NULL AND UNFED <> '', '-' || UNFED, ''), 
                              ENDER || 
                              IIF(MUNIC IS NOT NULL AND MUNIC <> '', ', ' || MUNIC, '') || 
                              IIF(UNFED IS NOT NULL AND UNFED <> '', '-' || UNFED, '')), 
                          IIF(BAIRR IS NOT NULL AND BAIRR <> '', ENDER || ', ' || BAIRR, ENDER)) 
                          FROM 1 FOR 100) AS ENDERECO,
                ENDNR AS NUMERO,
                NRCEP AS CEP,
                ENDCP AS OBSERVACAO,
                1 AS CODIGO_REGIAODETALHE,
                COALESCE(CODIGO_CIDADE_PS, 1) AS CODIGO_CIDADEESTADO,
                1 AS CADASTRO_LJ,
                1 AS CADASTRO_CF,
                CURRENT_TIMESTAMP AS CADASTRO_DT,
                1 AS ALTERACAO_LJ,
                1 AS ALTERACAO_CF,
                CURRENT_TIMESTAMP AS ALTERACAO_DT
           FROM FC04400 E
           INNER JOIN FC04000 M ON M.PFCRM = E.PFCRM AND  M.UFCRM = E.UFCRM AND M.NRCRM = E.NRCRM
           WHERE ENDER IS NOT NULL AND
                    ENDER <> '.' AND
                    ENDER <> '' AND
                    E.CODIGO_PS IS NOT NULL
                    AND M.CONVERSAO = :V_CONVERSAO
           INTO :CODIGO, :TIPO_CADASTRO, :CODIGO_CADASTRO, :ENDERECO, :NUMERO, :CEP, :OBSERVACAO_END, :CODIGO_REGIAODETALHE, :CODIGO_CIDADEESTADO, :CADASTRO_LJ, :CADASTRO_CF, :CADASTRO_DT, :ALTERACAO_LJ, :ALTERACAO_CF, :ALTERACAO_DT
           DO
           BEGIN
                I_SQL = 'INSERT INTO CADASTRO_ENDERECO (CODIGO, TIPO_CADASTRO, CODIGO_CADASTRO, ENDERECO, NUMERO, CEP, OBSERVACAO, CODIGO_REGIAODETALHE, CODIGO_CIDADEESTADO, CADASTRO_LJ, CADASTRO_CF, CADASTRO_DT, ALTERACAO_LJ, ALTERACAO_CF, ALTERACAO_DT) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)';
                EXECUTE STATEMENT (:I_SQL) (:CODIGO, :TIPO_CADASTRO, :CODIGO_CADASTRO, :ENDERECO, :NUMERO, :CEP, :OBSERVACAO_END, :CODIGO_REGIAODETALHE, :CODIGO_CIDADEESTADO, :CADASTRO_LJ, :CADASTRO_CF, :CADASTRO_DT, :ALTERACAO_LJ, :ALTERACAO_CF, :ALTERACAO_DT) ON EXTERNAL :CON AS USER :V_USER PASSWORD :V_SHA WITH COMMON TRANSACTION;
           END

           BEGIN
               EXECUTE PROCEDURE UTIL_FORMATA_TEL_FC04400;
            END

    -- Limpeza de dados vazios/inválidos
    BEGIN
        UPDATE FC04400 SET NRDDD = NULL WHERE NRDDD = '' OR NRDDD = 0;
        UPDATE FC04400 SET NRDDD2 = NULL WHERE NRDDD2 = '' OR NRDDD2 = 0;
        UPDATE FC04400 SET NRDDDFAX = NULL WHERE NRDDDFAX = '' OR NRDDDFAX = 0;
        UPDATE FC04400 SET NRTEL = NULL WHERE NRTEL = '';
        UPDATE FC04400 SET NRTEL2 = NULL WHERE NRTEL2 = '';
        UPDATE FC04400 SET NRFAX = NULL WHERE NRFAX = '';
    END

       FOR
            SELECT GEN_ID(GEN_CADASTRO_TELEFONE, 1) AS CODIGO,
                2 AS TIPO_CADASTRO,
                M.CODIGO_PS AS CODIGO_CADASTRO,
                IIF((SELECT ONLY_NUMBERS FROM STRIP_NON_NUMERIC(NRTEL)) <= 5, 2, 3) AS TELEFONE_TIPO,
                SUBSTRING((SELECT ONLY_NUMBERS FROM STRIP_NON_NUMERIC(REPLACE(NRDDD, ' ', ''))) FROM 1 FOR 2) AS TELEFONEPREFIXO,
                SUBSTRING((SELECT ONLY_NUMBERS FROM STRIP_NON_NUMERIC(NRTEL)) FROM 1 FOR 18) AS TELEFONE,
                '' AS OBSERVACAO,
                1 AS CADASTRO_LJ,
                1 AS CADASTRO_CF,
                CURRENT_TIMESTAMP AS CADASTRO_DT,
                1 AS ALTERACAO_LJ,
                1 AS ALTERACAO_CF,
                CURRENT_TIMESTAMP AS ALTERACAO_DT
                FROM FC04400 T
                INNER JOIN FC04000 M ON M.PFCRM = T.PFCRM AND  T.UFCRM = T.UFCRM AND M.NRCRM = T.NRCRM
                WHERE (SELECT ONLY_NUMBERS FROM STRIP_NON_NUMERIC(NRTEL)) <> ''
                AND M.CODIGO_PS IS NOT NULL
                AND M.CONVERSAO = :V_CONVERSAO

            UNION
                SELECT GEN_ID(GEN_CADASTRO_TELEFONE, 1) AS CODIGO,
                2 AS TIPO_CADASTRO,
                M.CODIGO_PS AS CODIGO_CADASTRO,
                IIF((SELECT ONLY_NUMBERS FROM STRIP_NON_NUMERIC(NRTEL2)) <= 5, 2, 3) AS TELEFONE_TIPO,
                SUBSTRING((SELECT ONLY_NUMBERS FROM STRIP_NON_NUMERIC(REPLACE(NRDDD2, ' ', ''))) FROM 1 FOR 2) AS TELEFONEPREFIXO,
                SUBSTRING((SELECT ONLY_NUMBERS FROM STRIP_NON_NUMERIC(NRTEL2)) FROM 1 FOR 18) AS TELEFONE,
                (SELECT ONLY_NUMBERS FROM STRIP_NON_NUMERIC(RAMAL2)) AS OBSERVACAO,
                1 AS CADASTRO_LJ,
                1 AS CADASTRO_CF,
                CURRENT_TIMESTAMP AS CADASTRO_DT,
                1 AS ALTERACAO_LJ,
                1 AS ALTERACAO_CF,
                CURRENT_TIMESTAMP AS ALTERACAO_DT
                FROM FC04400 T
                INNER JOIN FC04000 M ON M.PFCRM = T.PFCRM AND  T.UFCRM = T.UFCRM AND M.NRCRM = T.NRCRM
                WHERE (SELECT ONLY_NUMBERS FROM STRIP_NON_NUMERIC(NRTEL2)) <> ''
                AND M.CODIGO_PS IS NOT NULL
                AND M.CONVERSAO = :V_CONVERSAO
            UNION
                SELECT GEN_ID(GEN_CADASTRO_TELEFONE, 1) AS CODIGO,
                2 AS TIPO_CADASTRO,
                M.CODIGO_PS AS CODIGO_CADASTRO,
                IIF((SELECT ONLY_NUMBERS FROM STRIP_NON_NUMERIC(NRFAX)) <= 5, 2, 3) AS TELEFONE_TIPO,
                SUBSTRING((SELECT ONLY_NUMBERS FROM STRIP_NON_NUMERIC(REPLACE(NRDDDFAX, ' ', ''))) FROM 1 FOR 2) AS TELEFONEPREFIXO,
                SUBSTRING((SELECT ONLY_NUMBERS FROM STRIP_NON_NUMERIC(NRFAX)) FROM 1 FOR 18) AS TELEFONE,
                (SELECT ONLY_NUMBERS FROM STRIP_NON_NUMERIC(RAMALFAX)) AS OBSERVACAO,
                1 AS CADASTRO_LJ,
                1 AS CADASTRO_CF,
                CURRENT_TIMESTAMP AS CADASTRO_DT,
                1 AS ALTERACAO_LJ,
                1 AS ALTERACAO_CF,
                CURRENT_TIMESTAMP AS ALTERACAO_DT
                FROM FC04400 T
                INNER JOIN FC04000 M ON M.PFCRM = T.PFCRM AND  T.UFCRM = T.UFCRM AND M.NRCRM = T.NRCRM
                WHERE (SELECT ONLY_NUMBERS FROM STRIP_NON_NUMERIC(NRFAX)) <> ''
                AND M.CODIGO_PS IS NOT NULL
                AND M.CONVERSAO = :V_CONVERSAO

                INTO :CODIGO, :TIPO_CADASTRO, :CODIGO_CADASTRO, :TELEFONE_TIPO, :TELEFONEPREFIXO, :TELEFONE, :OBSERVACAO, :CADASTRO_LJ, :CADASTRO_CF, :CADASTRO_DT, :ALTERACAO_LJ, :ALTERACAO_CF, :ALTERACAO_DT
                DO
                BEGIN
                    -- FAZ A INSERCAO DOS DADOS NO BANCO DE DADOS PHARMACIE
                    I_SQL = 'INSERT INTO CADASTRO_TELEFONE (CODIGO,TIPO_CADASTRO,CODIGO_CADASTRO,TELEFONE_TIPO,TELEFONEPREFIXO,TELEFONE,OBSERVACAO,CADASTRO_LJ,CADASTRO_CF,CADASTRO_DT,ALTERACAO_LJ,ALTERACAO_CF,ALTERACAO_DT) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)';
            
                    EXECUTE STATEMENT (:I_SQL) (:CODIGO, :TIPO_CADASTRO, :CODIGO_CADASTRO, :TELEFONE_TIPO, :TELEFONEPREFIXO, :TELEFONE, :OBSERVACAO, :CADASTRO_LJ, :CADASTRO_CF, :CADASTRO_DT, :ALTERACAO_LJ, :ALTERACAO_CF, :ALTERACAO_DT) ON EXTERNAL :CON AS USER :V_USER PASSWORD :V_SHA WITH COMMON TRANSACTION;
                END
END;