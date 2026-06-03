create or alter procedure UPDATE_CODIGOS_PS (
    CONVERSAO_IP smallint,
    TIPO_ATUALIZACAO smallint)
returns (
    MENSAGEM_RETORNO varchar(1000))
as
declare variable V_USER varchar(20);
declare variable V_SHA varchar(20);
declare variable V_SQL varchar(3000);
declare variable V_CON varchar(200);
declare variable CODIGO_PS integer;
declare variable CODIGO_PRODUTO_PS integer;
declare variable CODIGO_SINONIMO_PS integer;
declare variable CONV integer;
declare variable COD_CLI integer;
declare variable COD_FOR integer;
declare variable COD_PRO integer;
declare variable COD_GRU integer;
declare variable COD_SIN integer;
declare variable CODIGO_MED integer;
declare variable NUMEROVENDA decimal(18,5);
declare variable NUMEROFORMULA decimal(18,5);
declare variable CODIGO_PS_A1 integer;
declare variable C1 integer;
declare variable C2 integer;
declare variable C3 integer;
declare variable C4 integer;
declare variable C5 integer;
BEGIN
    -- Inicializa contadores
    C1 = 0;
    C2 = 0;
    C3 = 0;
    C4 = 0;
    C5 = 0;

    MENSAGEM_RETORNO = '';

    -- Monta string de conexão do banco PHARMACIE
    SELECT FIRST 1 CN.IPSERVER || '/' || CN.PORTA || ':' || CN.ALIAS
    FROM CONEXAO CN
    WHERE CN.TIPO_CONEXAO = 2
    INTO :V_CON;

    -- Credenciais do banco externo
    V_USER = 'SYSDBA';
    V_SHA  = 'SySPs_PHARMACIE';

    ----------------------------------------------------------------------
    -- 1) CLIENTE (CLIENTE)
    ----------------------------------------------------------------------
    IF (:TIPO_ATUALIZACAO = 0 OR :TIPO_ATUALIZACAO = 1) THEN
    BEGIN
        FOR SELECT C.CODIGOCLIENTE, C.CODIGO_PS, C.CONVERSAO
            FROM CLIENTE C
            WHERE C.CODIGO_PS IS NOT NULL
              AND C.CONVERSAO = :CONVERSAO_IP
        INTO :COD_CLI, :CODIGO_PS, :CONV
        DO
        BEGIN
            V_SQL =
                'UPDATE CLIENTE SET ' ||
                ' CODIGO_PS = ' || :CODIGO_PS ||
                ', CONVERSAO = ' || :CONV ||
                ' WHERE CODIGOCLIENTE = ' || :COD_CLI;

            EXECUTE STATEMENT (:V_SQL)
            ON EXTERNAL :V_CON
            AS USER :V_USER PASSWORD :V_SHA
            WITH COMMON TRANSACTION;

            C1 = C1 + 1;
        END

        MENSAGEM_RETORNO = MENSAGEM_RETORNO ||
                           'CLIENTES ATUALIZADOS: ' || C1 || ASCII_CHAR(13) || ASCII_CHAR(10);
    END

    ----------------------------------------------------------------------
    -- 2) MÉDICO
    ----------------------------------------------------------------------
    IF (:TIPO_ATUALIZACAO = 0 OR :TIPO_ATUALIZACAO = 2) THEN
    BEGIN
        FOR SELECT 
                M.CODIGOMEDICO,
                M.CODIGO_PS,
                M.CONVERSAO
            FROM MEDICO M
            WHERE M.CODIGO_PS IS NOT NULL
              AND M.CONVERSAO = :CONVERSAO_IP
        INTO :CODIGO_MED, :CODIGO_PS, :CONV
        DO
        BEGIN
            V_SQL =
                'UPDATE MEDICO SET ' ||
                ' CODIGO_PS = ' || :CODIGO_PS ||
                ', CONVERSAO = ' || :CONV ||
                ' WHERE CODIGOMEDICO = ' || :CODIGO_MED;
            EXECUTE STATEMENT (:V_SQL)
            ON EXTERNAL :V_CON
            AS USER :V_USER PASSWORD :V_SHA
            WITH COMMON TRANSACTION;

            C2 = C2 + 1;
        END

        MENSAGEM_RETORNO = MENSAGEM_RETORNO ||
                           'MEDICOS ATUALIZADOS: ' || C2 || ASCII_CHAR(13) || ASCII_CHAR(10);
    END

    ----------------------------------------------------------------------
    -- 3) FORNECEDOR
    ----------------------------------------------------------------------
    IF (:TIPO_ATUALIZACAO = 0 OR :TIPO_ATUALIZACAO = 3) THEN
    BEGIN
        FOR SELECT F.CODIGOFORNECEDOR,
                   F.CONVERSAO
            FROM FORNECEDOR F
            WHERE F.CONVERSAO IS NOT NULL
              AND F.CONVERSAO = :CONVERSAO_IP
        INTO :COD_FOR,:CONV
        DO
        BEGIN
            V_SQL =
                'UPDATE FORNECEDOR SET ' ||
                'CONVERSAO = ' || :CONV ||
                ' WHERE CODIGOFORNECEDOR = ' || :COD_FOR;

            EXECUTE STATEMENT (:V_SQL)
            ON EXTERNAL :V_CON
            AS USER :V_USER PASSWORD :V_SHA
            WITH COMMON TRANSACTION;

            C3 = C3 + 1;
        END

        MENSAGEM_RETORNO = MENSAGEM_RETORNO ||
                           'FORNECEDORES ATUALIZADOS: ' || C3 || ASCII_CHAR(13) || ASCII_CHAR(10);
    END

    ----------------------------------------------------------------------
    -- 4) PRODUTO
    ----------------------------------------------------------------------
    IF (:TIPO_ATUALIZACAO = 0 OR :TIPO_ATUALIZACAO = 4) THEN
    BEGIN
        FOR SELECT P.CODIGOPRODUTO,P.CODIGOGRUPO,P.CODIGO_PS, P.CONVERSAO
            FROM PRODUTO P
            WHERE P.CODIGO_PS IS NOT NULL
              AND P.CONVERSAO = :CONVERSAO_IP
        INTO :COD_PRO,:COD_GRU,:CODIGO_PS, :CONV
        DO
        BEGIN
            V_SQL =
                'UPDATE PRODUTO SET ' ||
                ' CODIGO_PS = ' || :CODIGO_PS ||
                ', CONVERSAO = ' || :CONV ||
                ' WHERE CODIGOPRODUTO = ' || :COD_PRO ||
                ' AND CODIGOGRUPO = ' || :COD_GRU;

            EXECUTE STATEMENT (:V_SQL)
            ON EXTERNAL :V_CON
            AS USER :V_USER PASSWORD :V_SHA
            WITH COMMON TRANSACTION;

            C4 = C4 + 1;
        END

        MENSAGEM_RETORNO = MENSAGEM_RETORNO ||
                           'PRODUTOS ATUALIZADOS: ' || C4 || ASCII_CHAR(13) || ASCII_CHAR(10);
    END

    ----------------------------------------------------------------------
    -- 5) SINONIMO
    ----------------------------------------------------------------------
    IF (:TIPO_ATUALIZACAO = 0 OR :TIPO_ATUALIZACAO = 5) THEN
    BEGIN
        FOR SELECT S.CODIGOPRODUTO,S.CODIGOGRUPO,S.CODIGOSINONIMO,
                    S.CODIGO_PRODUTO_PS, S.CODIGO_SINONIMO_PS, S.CONVERSAO
            FROM SINONIMO S
            WHERE S.CODIGO_SINONIMO_PS IS NOT NULL
              AND S.CONVERSAO = :CONVERSAO_IP
        INTO :COD_PRO,:COD_GRU, :COD_SIN,
             :CODIGO_PRODUTO_PS, :CODIGO_SINONIMO_PS, :CONV
        DO
        BEGIN
            V_SQL =
                'UPDATE SINONIMO SET ' ||
                ' CODIGO_PRODUTO_PS = ' || :CODIGO_PRODUTO_PS ||
                ', CODIGO_SINONIMO_PS = ' || :CODIGO_SINONIMO_PS ||
                ', CONVERSAO = ' || :CONV ||
                ' WHERE CODIGOPRODUTO = ' || :COD_PRO ||
                ' AND CODIGOGRUPO = ' || :COD_GRU ||
                ' AND CODIGOSINONIMO = ' || :COD_SIN;

            EXECUTE STATEMENT (:V_SQL)
            ON EXTERNAL :V_CON
            AS USER :V_USER PASSWORD :V_SHA
            WITH COMMON TRANSACTION;

            C5 = C5 + 1;
        END

        MENSAGEM_RETORNO = MENSAGEM_RETORNO ||
                           'SINONIMOS ATUALIZADOS: ' || C5 || ASCII_CHAR(13) || ASCII_CHAR(10);
    END

     ----------------------------------------------------------------------
    -- 5) REFAZER
    ----------------------------------------------------------------------
    IF (:TIPO_ATUALIZACAO = 0 OR :TIPO_ATUALIZACAO = 6) THEN
    BEGIN
        FOR SELECT
            A1.NUMEROVENDA,A1.CODIGO_PS, A1.CONVERSAO
            FROM VENDA A1
            WHERE A1.CODIGO_PS IS NOT NULL AND A1.CONVERSAO = :CONVERSAO_IP
        INTO :NUMEROVENDA, :CODIGO_PS, :CONV
        DO
        BEGIN
            V_SQL =
                'UPDATE VENDA SET ' ||
                ' CODIGO_PS = ' || :CODIGO_PS ||
                ', CONVERSAO = ' || :CONV ||
                ' WHERE NUMEROVENDA = ' || :NUMEROVENDA;

            EXECUTE STATEMENT (:V_SQL)
            ON EXTERNAL :V_CON
            AS USER :V_USER PASSWORD :V_SHA
            WITH COMMON TRANSACTION;

            C5 = C5 + 1;
        END

        FOR SELECT
            A2.NUMEROVENDA,
            A2.NUMEROFORMULA,
            A2.CODIGO_PS,
            A2.CODIGO_PS_A1,
            A2.CONVERSAO
            FROM FORMULAVENDA A2
            WHERE A2.CODIGO_PS IS NOT NULL AND A2.CONVERSAO = :CONVERSAO_IP
        INTO :NUMEROVENDA,:NUMEROFORMULA, :CODIGO_PS,:CODIGO_PS_A1, :CONV
        DO
        BEGIN
            V_SQL =
                'UPDATE FORMULAVENDA  SET ' ||
                ' CODIGO_PS = ' || :CODIGO_PS ||
                ', CONVERSAO = ' || :CONV ||
                ', CODIGO_PS_A1 = ' || :CODIGO_PS_A1 ||
                ' WHERE NUMEROVENDA = ' || :NUMEROVENDA ||
                ' AND NUMEROFORMULA = ' || :NUMEROFORMULA ;

            EXECUTE STATEMENT (:V_SQL)
            ON EXTERNAL :V_CON
            AS USER :V_USER PASSWORD :V_SHA
            WITH COMMON TRANSACTION;

            C5 = C5 + 1;
        END

        FOR SELECT
            A3.NUMEROVENDA,
            A3.NUMEROFORMULA,
            A3.CODIGO_PS,
            A3.CODIGO_PS_A1,
            A3.CONVERSAO
            FROM ITEMFORMULAVENDA A3
            WHERE A3.CODIGO_PS IS NOT NULL AND A3.CONVERSAO = :CONVERSAO_IP
        INTO :NUMEROVENDA,:NUMEROFORMULA, :CODIGO_PS,:CODIGO_PS_A1, :CONV
        DO
        BEGIN
            V_SQL =
                'UPDATE ITEMFORMULAVENDA  SET ' ||
                ' CODIGO_PS = ' || :CODIGO_PS ||
                ', CONVERSAO = ' || :CONV ||
                ', CODIGO_PS_A1 = ' || :CODIGO_PS_A1 ||
                ' WHERE NUMEROVENDA = ' || :NUMEROVENDA ||
                ' AND NUMEROFORMULA = ' || :NUMEROFORMULA ;

            EXECUTE STATEMENT (:V_SQL)
            ON EXTERNAL :V_CON
            AS USER :V_USER PASSWORD :V_SHA
            WITH COMMON TRANSACTION;

            C5 = C5 + 1;
        END



        MENSAGEM_RETORNO = MENSAGEM_RETORNO || 'VENDAS ATUALIZADOS: ' || C5 || ASCII_CHAR(13) || ASCII_CHAR(10);
    END

    SUSPEND;
END;