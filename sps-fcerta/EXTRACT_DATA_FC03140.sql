create or alter procedure EXTRACT_DATA_FC03140 (
    V_VENCIDO integer)
as
declare variable V_USER varchar(20);
declare variable V_SHA varchar(20);
declare variable I_SQL varchar(2000);
declare variable CON varchar(60);
declare variable VENCIDO_CHECK integer;
declare variable CODIGO integer;
declare variable TIPOLOTE integer;
declare variable CODIGO_MOV integer;
declare variable CODIGO_MOV_DET integer;
declare variable CODIGO_PRODUTO integer;
declare variable CODIGO_FORNECEDOR integer;
declare variable CODIGO_ESTOQUE_ORIGEM integer;
declare variable LOTE varchar(30);
declare variable CONCENTRACAO numeric(18,4);
declare variable DENSIDADE numeric(18,4);
declare variable UMIDADE numeric(18,4);
declare variable UTR numeric(18,4);
declare variable UFC numeric(18,4);
declare variable FABRICACAO_DT date;
declare variable VENCIMENTO_DT date;
declare variable QUANTIDADE numeric(18,4);
declare variable FECHARLOTE integer;
declare variable LOTEUNIFICADO integer;
declare variable CADASTRO_LJ smallint;
declare variable CADASTRO_CF smallint;
declare variable CADASTRO_DT timestamp;
declare variable ALTERACAO_LJ smallint;
declare variable ALTERACAO_CF smallint;
declare variable ALTERACAO_DT timestamp;
declare variable STATUSGERAL integer;
declare variable UI numeric(18,4);
declare variable FRACAO_ENTRADA integer;
declare variable CODIGO_ESTOQUE integer;
declare variable CODIGO_ESTOQUE_LOTE integer;
declare variable QTDE_ENTRADA numeric(18,4);
declare variable SALDO_DT timestamp;
declare variable STATUSLOTE smallint;
declare variable ORIGEMLOTE smallint;
BEGIN

    /*
    ###################################################################################################
    ###        VARIAVEL V_VENCIDO                                                                  ####
    ###       -1 = PEGAR LOTES VENCIDOS E NAO VENCIDOS POREM COM QUANTIDADE MAIOR QUE "0"         ####
    ###       0 = PEGAR LOTES COM DATA VALIDADE MAIOR QUE DATA CORRENTE E QUANTIDADE MAIOR QUE "0" ####
    ###################################################################################################
     */

    VENCIDO_CHECK = V_VENCIDO;
    V_USER = 'SYSDBA';
    V_SHA = 'SySPs_PHARMACIE';

    /*OBTEM E MONTA STRING DE CONEXAO DO PHARMACIE*/
    SELECT FIRST 1 CN.IPSERVER || '/' || CN.PORTA || ':' || CN.ALIAS FROM CONEXAO CN INTO :CON;

    UPDATE FC03140 L
    SET L.CODIGO_PS_LOTE = GEN_ID(GEN_ESTOQUE_LOTE, 1)
    WHERE
        (:VENCIDO_CHECK = -1 AND L.ESTAT > 0) OR -- Verificar vencidos
        (:VENCIDO_CHECK = 0 AND L.DTVAL > CURRENT_DATE AND L.ESTAT > 0);

    UPDATE FC03140 L
    SET L.CODIGO_PS_LOTE_LA = GEN_ID(GEN_ESTOQUE_LOTE_LA, 1)
    WHERE 
        (:VENCIDO_CHECK = -1 AND L.ESTAT > 0) OR -- Verificar vencidos
        (:VENCIDO_CHECK = 0 AND L.DTVAL > CURRENT_DATE AND L.ESTAT > 0);


     FOR
        SELECT 
        CODIGO_PS_LOTE AS CODIGO,
        0 AS TIPOLOTE,
        0 AS CODIGO_MOV,
        0 AS CODIGO_MOV_DET,
        P.CODIGO_PS AS CODIGO_PRODUTO,
        (SELECT IIF(F.FORNECID IS NULL, NULL, L.FORNECID)
        FROM FC02000 F 
        WHERE L.FORNECID = F.FORNECID) AS CODIGO_FORNECEDOR,
        1 AS CODIGO_ESTOQUE_ORIGEM,
        L.NRLOT AS LOTE,
        L.TEOR  AS CONCENTRACAO,
        L.DENSI AS DENSIDADE,
        0 AS UMIDADE,
        COALESCE(L.QTUTR, 0) AS UTR,
        COALESCE(L.QTUFC /1000000000, 0) AS UFC,
        L.DTFAB AS FABRICACAO_DT,
        L.DTVAL AS VENCIMENTO_DT,
        CAST(L.ESTAT AS NUMERIC(18,4)) AS QUANTIDADE,
        0 AS FECHARLOTE,
        0 AS LOTEUNIFICADO,
        1 AS CADASTRO_LJ,
        1 AS CADASTRO_CF,
        L.DTENT AS CADASTRO_DT,
        1 AS ALTERACAO_LJ,
        1 AS ALTERACAO_CF,
        CAST((L.DTALT || ' ' || L.HRALT) AS TIMESTAMP)  AS ALTERACAO_DT,
        0 AS STATUSGERAL,
        COALESCE(L.QTUI, 0) AS UI,
        0 AS FRACAO_ENTRADA
        FROM FC03140 L
          INNER JOIN FC03000 P ON L.CDPRO = P.CDPRO
        WHERE L.CODIGO_PS_LOTE IS NOT NULL
        AND L.CODIGO_PS_LOTE_LA IS NOT NULL
        INTO :CODIGO, :TIPOLOTE, :CODIGO_MOV, :CODIGO_MOV_DET, :CODIGO_PRODUTO, :CODIGO_FORNECEDOR, :CODIGO_ESTOQUE_ORIGEM, :LOTE, :CONCENTRACAO, :DENSIDADE, :UMIDADE, :UTR, :UFC, :FABRICACAO_DT, :VENCIMENTO_DT, :QUANTIDADE, :FECHARLOTE, :LOTEUNIFICADO, :CADASTRO_LJ, :CADASTRO_CF, :CADASTRO_DT, :ALTERACAO_LJ, :ALTERACAO_CF, :ALTERACAO_DT, :STATUSGERAL, :UI, :FRACAO_ENTRADA
        DO
          BEGIN                                                                                                                                                                                                                                                                                                                                                                     --27
          I_SQL = 'INSERT INTO ESTOQUE_LOTE(CODIGO, TIPOLOTE, CODIGO_MOV, CODIGO_MOV_DET, CODIGO_PRODUTO, CODIGO_FORNECEDOR, CODIGO_ESTOQUE_ORIGEM, LOTE, CONCENTRACAO, DENSIDADE, UMIDADE, UTR, UFC, FABRICACAO_DT, VENCIMENTO_DT, QUANTIDADE, FECHARLOTE, LOTEUNIFICADO,  CADASTRO_LJ, CADASTRO_CF, CADASTRO_DT, ALTERACAO_LJ, ALTERACAO_CF, ALTERACAO_DT, STATUSGERAL, UI, FRACAO_ENTRADA ) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)';
          EXECUTE STATEMENT (:I_SQL) (:CODIGO, :TIPOLOTE, :CODIGO_MOV, :CODIGO_MOV_DET, :CODIGO_PRODUTO, :CODIGO_FORNECEDOR, :CODIGO_ESTOQUE_ORIGEM, :LOTE, :CONCENTRACAO, :DENSIDADE, :UMIDADE, :UTR, :UFC, :FABRICACAO_DT, :VENCIMENTO_DT, :QUANTIDADE, :FECHARLOTE, :LOTEUNIFICADO, :CADASTRO_LJ, :CADASTRO_CF, :CADASTRO_DT, :ALTERACAO_LJ, :ALTERACAO_CF, :ALTERACAO_DT, :STATUSGERAL, :UI, :FRACAO_ENTRADA) ON EXTERNAL :CON AS USER :V_USER PASSWORD :V_SHA WITH COMMON TRANSACTION;
          END

    FOR
        SELECT 
            CODIGO_PS_LOTE_LA AS CODIGO,
            CASE
              WHEN CODIGO_PS < 1000000 THEN 1
              WHEN CODIGO_PS > 1000000 AND CODIGO_PS < 1999999 THEN 3
              WHEN CODIGO_PS > 2000000 AND CODIGO_PS < 2999999 THEN 3
              WHEN CODIGO_PS > 3000000 AND CODIGO_PS < 3999999 THEN 2
              WHEN CODIGO_PS > 4000000 AND CODIGO_PS < 4999999 THEN 3
            END AS CODIGO_ESTOQUE,
            CODIGO_PS_LOTE AS CODIGO_ESTOQUE_LOTE,
            0 AS CODIGO_MOV,
            CAST(L.ESTAT AS NUMERIC(18,4)) AS QTDE_ENTRADA,
            CURRENT_TIMESTAMP AS SALDO_DT,
            CASE STLOT
              WHEN 'P' THEN 3
              WHEN 'L' THEN 1
              WHEN 'B' THEN 0
              ELSE 1
            END AS STATUSLOTE,
            0 AS ORIGEMLOTE,
            1 AS CADASTRO_LJ,
            1 AS CADASTRO_CF,
            CURRENT_TIMESTAMP AS CADASTRO_DT,
            1 AS ALTERACAO_LJ,
            1 AS ALTERACAO_CF,
            CURRENT_TIMESTAMP  AS ALTERACAO_DT
            FROM FC03140 L
            INNER JOIN FC03000 P ON L.CDPRO = P.CDPRO
            WHERE L.CODIGO_PS_LOTE IS NOT NULL
            AND L.CODIGO_PS_LOTE_LA IS NOT NULL
            INTO :CODIGO, :CODIGO_ESTOQUE, :CODIGO_ESTOQUE_LOTE, :CODIGO_MOV, :QTDE_ENTRADA, :SALDO_DT, :STATUSLOTE, :ORIGEMLOTE, :CADASTRO_LJ, :CADASTRO_CF, :CADASTRO_DT, :ALTERACAO_LJ, :ALTERACAO_CF, :ALTERACAO_DT
            DO
              BEGIN                                                                                                                                                                                                                              --14
              I_SQL = 'INSERT INTO ESTOQUE_LOTE_LA (CODIGO, CODIGO_ESTOQUE, CODIGO_ESTOQUE_LOTE, CODIGO_MOV, QTDE_ENTRADA, SALDO_DT,  STATUSLOTE, ORIGEMLOTE, CADASTRO_LJ, CADASTRO_CF, CADASTRO_DT, ALTERACAO_LJ, ALTERACAO_CF, ALTERACAO_DT) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)';
              EXECUTE STATEMENT (:I_SQL) (:CODIGO, :CODIGO_ESTOQUE, :CODIGO_ESTOQUE_LOTE, :CODIGO_MOV, :QTDE_ENTRADA, :SALDO_DT, :STATUSLOTE, :ORIGEMLOTE, :CADASTRO_LJ, :CADASTRO_CF, :CADASTRO_DT, :ALTERACAO_LJ, :ALTERACAO_CF, :ALTERACAO_DT) ON EXTERNAL :CON AS USER :V_USER PASSWORD :V_SHA WITH COMMON TRANSACTION;
              END
END;