SET TERM ^ ;

create or alter procedure EXTRACT_DATA_PI (
    V_CONVERSAO integer)
as
declare variable V_USER varchar(20);
declare variable V_SHA varchar(20);
declare variable I_SQL varchar(2000);
declare variable CON varchar(60);
declare variable CODIGO integer;
declare variable MOVIMENTOESTOQUE integer;
declare variable CODIGO_FORMULA integer;
declare variable CODIGO_FORMAFARMACEUTICA integer;
declare variable MODOFAZER varchar(7000);
declare variable POSOLOGIA varchar(7000);
declare variable OPCAO_ATENDIMENTO smallint;
declare variable PRECOSISTEMA smallint;
declare variable OPCAO_ROTULO smallint;
declare variable TIPOCADASTRO smallint;
declare variable PESOFORMULA numeric(18,2);
declare variable INDICECALCULO smallint;
declare variable CADASTRO_LJ smallint;
declare variable CADASTRO_CF smallint;
declare variable CADASTRO_DT timestamp;
declare variable ALTERACAO_LJ smallint;
declare variable ALTERACAO_CF smallint;
declare variable ALTERACAO_DT timestamp;
declare variable CODIGO_ESTOQUEF integer;
declare variable NUMEROLINHA smallint;
declare variable CODIGO_PRODUTO integer;
declare variable QUANTIDADE numeric(18,4);
declare variable UNIDADE varchar(7);
declare variable FASE varchar(2);
declare variable CONCLUIDO smallint;
BEGIN
          V_USER = 'SYSDBA';
          V_SHA = 'masterkey';


        SELECT FIRST 1 CN.IPSERVER || '/' || CN.PORTA || ':' || CN.ALIAS FROM CONEXAO CN INTO :CON;
        
        UPDATE FC05200 FF
        SET FF.CODIGO_PRODUTO_PS = (SELECT FIRST 1 P.CODIGO_PS FROM FC03000 P WHERE P.CDPRO = FF.CDACA);
        
        UPDATE FC05000 F
        SET F.CODIGO_PRODUTO_PS = (SELECT FIRST 1 P.CODIGO_PS FROM FC03000 P WHERE P.CDPRO = F.CDSAC AND P.CODIGO_PS < 4000000)
        WHERE F.CODIGO_PRODUTO_PS IS NULL;
        
        UPDATE FC05000 F
        SET F.CODIGO_PRODUTO_PS = (SELECT FIRST 1 ACA.CODIGO_PRODUTO_PS FROM FC05200 ACA WHERE (ACA.CDFRM = F.CDFRM))
        WHERE F.CODIGO_PRODUTO_PS IS NULL;

        -- ATUALIZA O CAMPO CONVERSAO COM O VALOR FORNECIDO
        UPDATE FC05000 F
        SET F.CONVERSAO = :V_CONVERSAO
        WHERE F.CODIGO_PRODUTO_PS IS NOT NULL AND CONVERSAO IS NULL;

        FOR
            SELECT
            F.CDFRM AS CODIGO,
            CASE
              WHEN F.CODIGO_PRODUTO_PS < 1000000 THEN 0
              WHEN F.CODIGO_PRODUTO_PS > 1000000 AND F.CODIGO_PRODUTO_PS < 1999999 THEN 1
              WHEN F.CODIGO_PRODUTO_PS > 2000000 AND F.CODIGO_PRODUTO_PS < 2999999 THEN 2
              WHEN F.CODIGO_PRODUTO_PS > 3000000 AND F.CODIGO_PRODUTO_PS < 3999999 THEN 3
              WHEN F.CODIGO_PRODUTO_PS > 4000000 AND F.CODIGO_PRODUTO_PS < 4999999 THEN 4
            END AS MOVIMENTOESTOQUE,
            F.CODIGO_PRODUTO_PS AS CODIGO_FORMULA,
            TPFORMAFARMA AS CODIGO_FORMAFARMACEUTICA,
            CAST(UPPER (OBSER) AS VARCHAR(7000)) AS MODOFAZER,
            IIF(F.POSOL2 IS NOT NULL, F.POSOL || '; ' || F.POSOL2, F.POSOL) AS POSOLOGIA,
            0 AS OPCAO_ATENDIMENTO,
            0 AS PRECOSISTEMA,
            0 AS OPCAO_ROTULO,
            IIF(VOLUME = 1, 2, 1) AS TIPOCADASTRO,
            VOLUME AS PESOFORMULA,
            1 AS INDICECALCULO,
            1 AS CADASTRO_LJ,
            1 AS CADASTRO_CF,
            CURRENT_TIMESTAMP AS CADASTRO_DT,
            1 AS ALTERACAO_LJ,
            1 AS ALTERACAO_CF,
            CURRENT_TIMESTAMP AS ALTERACAO_DT,
            -1 AS CONCLUIDO
            FROM FC05000 F
            WHERE F.CODIGO_PRODUTO_PS IS NOT NULL AND F.CONVERSAO = :V_CONVERSAO
            INTO :CODIGO, :MOVIMENTOESTOQUE, :CODIGO_FORMULA, :CODIGO_FORMAFARMACEUTICA, :MODOFAZER, :POSOLOGIA, :OPCAO_ATENDIMENTO, :PRECOSISTEMA, :OPCAO_ROTULO, :TIPOCADASTRO, :PESOFORMULA, :INDICECALCULO, :CADASTRO_LJ, :CADASTRO_CF, :CADASTRO_DT, :ALTERACAO_LJ, :ALTERACAO_CF, :ALTERACAO_DT, :CONCLUIDO
            DO                                                                                                                                                                                                                                                                                                                                                              
            BEGIN
                I_SQL = 'INSERT INTO ESTOQUEF (CODIGO, MOVIMENTOESTOQUE, CODIGO_FORMULA, CODIGO_FORMAFARMACEUTICA, MODOFAZER, POSOLOGIA, OPCAO_ATENDIMENTO, PRECOSISTEMA, OPCAO_ROTULO, TIPOCADASTRO, PESOFORMULA, INDICECALCULO, CADASTRO_LJ, CADASTRO_CF, CADASTRO_DT, ALTERACAO_LJ, ALTERACAO_CF, ALTERACAO_DT, CONCLUIDO) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,?)';
                EXECUTE STATEMENT (:I_SQL) (:CODIGO, :MOVIMENTOESTOQUE, :CODIGO_FORMULA, :CODIGO_FORMAFARMACEUTICA, :MODOFAZER, :POSOLOGIA, :OPCAO_ATENDIMENTO, :PRECOSISTEMA, :OPCAO_ROTULO, :TIPOCADASTRO, :PESOFORMULA, :INDICECALCULO, :CADASTRO_LJ, :CADASTRO_CF, :CADASTRO_DT, :ALTERACAO_LJ, :ALTERACAO_CF, :ALTERACAO_DT, :CONCLUIDO ) ON EXTERNAL :CON AS USER :V_USER PASSWORD :V_SHA WITH COMMON TRANSACTION;
            END

            
       FOR
          SELECT
            GEN_ID(GEN_ESTOQUEF_DETALHE, 1) AS CODIGO,
            FI.CDFRM AS CODIGO_ESTOQUEF,
            FI.ITEMID AS NUMEROLINHA,
            P.CODIGO_PS AS CODIGO_PRODUTO,
            FI.QUANT AS QUANTIDADE,
            LOWER(IIF(FI.UNIDA = 'UN', 'u', FI.UNIDA)) AS UNIDADE,
            IIF(FASEID = '', 1, FASEID) AS FASE,
            1 AS CADASTRO_LJ,
            1 AS CADASTRO_CF,
            CURRENT_TIMESTAMP AS CADASTRO_DT,
            1 AS ALTERACAO_LJ,
            1 AS ALTERACAO_CF,
            CURRENT_TIMESTAMP AS ALTERACAO_DT
            FROM FC05100 FI
            RIGHT JOIN FC05000 F ON F.CDFRM = FI.CDFRM
            INNER JOIN FC03000 P ON P.CDPRO = FI.CDPRIN
            WHERE F.CODIGO_PRODUTO_PS IS NOT NULL AND F.CONVERSAO = :V_CONVERSAO
            INTO :CODIGO, :CODIGO_ESTOQUEF, :NUMEROLINHA, :CODIGO_PRODUTO, :QUANTIDADE, :UNIDADE, :FASE, :CADASTRO_LJ, :CADASTRO_CF, :CADASTRO_DT, :ALTERACAO_LJ, :ALTERACAO_CF, :ALTERACAO_DT
            DO
            BEGIN                                                                                                                                                                                                           
                I_SQL = 'INSERT INTO ESTOQUEF_DETALHE (CODIGO, CODIGO_ESTOQUEF, NUMEROLINHA, CODIGO_PRODUTO, QUANTIDADE, UNIDADE, FASE, CADASTRO_LJ, CADASTRO_CF, CADASTRO_DT, ALTERACAO_LJ, ALTERACAO_CF, ALTERACAO_DT) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)';
                EXECUTE STATEMENT (:I_SQL) (:CODIGO, :CODIGO_ESTOQUEF, :NUMEROLINHA, :CODIGO_PRODUTO, :QUANTIDADE, :UNIDADE, :FASE, :CADASTRO_LJ, :CADASTRO_CF, :CADASTRO_DT, :ALTERACAO_LJ, :ALTERACAO_CF, :ALTERACAO_DT) ON EXTERNAL :CON AS USER :V_USER PASSWORD :V_SHA WITH COMMON TRANSACTION;
            END
END^

SET TERM ; ^
