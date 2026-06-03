create or alter procedure EXTRACT_DATA_REFAZER_2
as
declare variable V_USER varchar(20);
declare variable V_SHA varchar(20);
declare variable I_SQL varchar(2000);
declare variable CON varchar(60);
declare variable CODIGO integer;
declare variable CADASTRO_DT timestamp;
declare variable CODIGO_CLIENTE integer;
declare variable CODIGO_MEDICO integer;
declare variable OBSERVACAO varchar(7000);
declare variable CODIGO_REF_A1 integer;
declare variable NUMEROFORMULA integer;
declare variable CODIGO_FORMAFARMACEUTICA integer;
declare variable QVD numeric(18,2);
declare variable POSOLOGIA varchar(7000);
declare variable OBSERVACAOFORMULA varchar(7000);
declare variable CODIGO_PRODUTO integer;
declare variable QUANTIDADE numeric(18,4);
declare variable UNIDADE varchar(7);
declare variable NUMEROLINHA integer;
declare variable INCLUSAOSISTEMA integer;
BEGIN
     V_USER = 'SYSDBA';
     V_SHA = 'SySPs_PHARMACIE';

      /*OBTEM E MONTA STRING DE CONEXAO DO PHARMACIE*/
      SELECT FIRST 1 CN.IPSERVER || '/' || CN.PORTA || ':' || CN.ALIAS FROM CONEXAO CN INTO :CON;

        BEGIN
              UPDATE FC15000 A1
              SET A1.CODIGO_PS = GEN_ID(GEN_REFAZER_A1, 1)
              WHERE EXISTS (SELECT A2.NRORC
                      FROM FC15100 A2
                      WHERE A2.NRORC = A1.NRORC AND
                            A2.CDFIL = A1.CDFIL)
              AND EXISTS(SELECT C.CDCLI
                     FROM FC07000 C
                     WHERE C.CDCLI = A1.CDCLI)
              AND A1.CODIGO_PS IS NULL
              ORDER BY A1.NRORC ASC, A1.CDFIL ASC;
         END

        BEGIN
              UPDATE FC15100 A2
              SET A2.CODIGO_PS = GEN_ID(GEN_REFAZER_A2, 1),
              A2.CODIGO_PS_A1 = (SELECT FIRST 1 A1.CODIGO_PS
                          FROM FC15000 A1
                          WHERE A1.NRORC = A2.NRORC AND
                                A1.CDFIL = A2.CDFIL AND
                                A1.CODIGO_PS IS NOT NULL)
              WHERE EXISTS (SELECT FIRST 1 A1.CODIGO_PS
              FROM FC15000 A1
              WHERE A1.NRORC = A2.NRORC AND
                    A1.CDFIL = A2.CDFIL AND
                    A1.CODIGO_PS IS NOT NULL)
              AND A2.CODIGO_PS_A1 IS NULL
              ORDER BY A2.NRORC ASC, A2.CDFIL ASC;
        END

        BEGIN
              UPDATE FC15110 A3
              SET A3.CODIGO_PS = GEN_ID(GEN_REFAZER_A3, 1),
              A3.CODIGO_PS_A1 = (SELECT FIRST 1 A1.CODIGO_PS
                          FROM FC15000 A1
                          WHERE A1.NRORC = A3.NRORC AND
                                A1.CDFIL = A3.CDFIL AND
                                A1.CODIGO_PS IS NOT NULL)
              WHERE EXISTS (SELECT FIRST 1 A1.CODIGO_PS
              FROM FC15000 A1
              WHERE A1.NRORC = A3.NRORC AND
                    A1.CDFIL = A3.CDFIL AND
                    A1.CODIGO_PS IS NOT NULL)
              AND EXISTS (SELECT FIRST 1 P.CODIGO_PS
              FROM FC03000 P
              WHERE P.CDPRO = A3.CDPRIN)
              AND A3.CODIGO_PS_A1 IS NULL
              ORDER BY A3.NRORC ASC, A3.CDFIL ASC;
       END

           FOR
               SELECT A1.CODIGO_PS AS CODIGO,
                CAST((MIN(A1.DTENTR) || ' ' || MIN(A2.HRCAD)) AS TIMESTAMP) AS CADASTRO_DT,
                C.CODIGO_PS AS CODIGO_CLIENTE,
                COALESCE((SELECT FIRST 1 M.CODIGO_PS
                         FROM FC04000 M
                         WHERE M.NRCRM = MIN(A2.NRCRM) AND M.PFCRM = MIN(A2.PFCRM)),9999999) AS CODIGO_MEDICO,
                '### VALOR BRUTO: R$ ' || CAST(A1.VRRQU AS NUMERIC(18,2)) || 
                ' | DESCONTO: R$ ' || CAST(A1.VRDSC AS NUMERIC(18,2)) || 
                ' | VALOR LIQUIDO: R$ ' || CAST(A1.VRRQU - A1.VRDSC AS NUMERIC(18,2)) || 
                ' ###' AS OBSERVACAO
                FROM FC15000 A1
                  INNER JOIN FC15100 A2 ON A1.CODIGO_PS = A2.CODIGO_PS_A1
                  INNER JOIN FC07000 C ON C.CDCLI = A1.CDCLI
                WHERE A1.CODIGO_PS IS NOT NULL
                GROUP BY 1,3,5
                INTO :CODIGO, :CADASTRO_DT, :CODIGO_CLIENTE, :CODIGO_MEDICO, :OBSERVACAO
                    DO
                    BEGIN
                        I_SQL = 'INSERT INTO ATENDIMENTO_REF_A1(CODIGO, CADASTRO_DT, CODIGO_CLIENTE, CODIGO_MEDICO, OBSERVACAO) VALUES (?, ?, ?, ?, ?)';
                        EXECUTE STATEMENT (:I_SQL)(:CODIGO, :CADASTRO_DT, :CODIGO_CLIENTE, :CODIGO_MEDICO, :OBSERVACAO) ON EXTERNAL :CON AS USER :V_USER PASSWORD :V_SHA WITH COMMON TRANSACTION;
                    END

           FOR
               SELECT
                    CODIGO_PS AS CODIGO,
                    A2.CODIGO_PS_A1 AS CODIGO_REF_A1,
                    CASE A2.SERIEO
                      WHEN '0' THEN 1
                      WHEN '1' THEN 2
                      WHEN '2' THEN 3
                      WHEN '3' THEN 4
                      WHEN '4' THEN 5
                      WHEN '5' THEN 6
                      WHEN '6' THEN 7
                      WHEN '7' THEN 8
                      WHEN '8' THEN 9
                      WHEN '9' THEN 10
                      WHEN 'A' THEN 11
                      WHEN 'B' THEN 12
                      WHEN 'C' THEN 13
                      WHEN 'D' THEN 14
                      WHEN 'E' THEN 15
                      WHEN 'F' THEN 16
                      WHEN 'G' THEN 19
                      WHEN 'H' THEN 20
                      WHEN 'I' THEN 21
                      WHEN 'J' THEN 22
                      WHEN 'K' THEN 23
                      WHEN 'L' THEN 24
                      WHEN 'M' THEN 25
                      WHEN 'N' THEN 26
                      WHEN 'O' THEN 27
                      WHEN 'P' THEN 28
                      WHEN 'Q' THEN 29
                      WHEN 'R' THEN 30
                      WHEN 'S' THEN 31
                      WHEN 'T' THEN 32
                      WHEN 'U' THEN 33
                      WHEN 'V' THEN 34
                      WHEN 'X' THEN 35
                      WHEN 'Z' THEN 36
                      WHEN 'Y' THEN 37
                      ELSE 38
                    END AS NUMEROFORMULA,
                    A2.TPFORMAFARMA AS CODIGO_FORMAFARMACEUTICA, --PRECISA CADASTRAS AS FORMA FARMACEUTICA E OS LABORATORIOS
                    A2.VOLUME AS QVD,
                    A2.POSOL AS POSOLOGIA,
                    '### VALOR FORMULA: R$ ' || CAST(A2.PRCOBR AS NUMERIC(18,2)) || ' ###' AS OBSERVACAOFORMULA
                    FROM FC15100 A2
                    WHERE A2.CODIGO_PS IS NOT NULL
                    INTO :CODIGO, :CODIGO_REF_A1, :NUMEROFORMULA, :CODIGO_FORMAFARMACEUTICA, :QVD, :POSOLOGIA, :OBSERVACAOFORMULA
                    DO
                      BEGIN
                         I_SQL = 'INSERT INTO ATENDIMENTO_REF_A2(CODIGO, CODIGO_REF_A1, NUMEROFORMULA, CODIGO_FORMAFARMACEUTICA, QVD, POSOLOGIA, OBSERVACAOFORMULA) VALUES(?, ?, ?, ?, ?, ?, ?)';
                         EXECUTE STATEMENT (:I_SQL)(:CODIGO, :CODIGO_REF_A1, :NUMEROFORMULA, :CODIGO_FORMAFARMACEUTICA, :QVD, :POSOLOGIA, :OBSERVACAOFORMULA) ON EXTERNAL :CON AS USER :V_USER PASSWORD :V_SHA WITH COMMON TRANSACTION;
                      END
            FOR
                SELECT
                    A3.CODIGO_PS AS CODIGO,
                    A3.CODIGO_PS_A1 AS CODIGO_REF_A1,
                    P.CODIGO_PS AS CODIGO_PRODUTO,
                    REPLACE(CAST(A3.QUANT AS NUMERIC(18,4)), ',', '.') AS QUANTIDADE,
                    LOWER(IIF(A3.UNIHP IS NULL, IIF(A3.UNIDA = 'UN', 'u', A3.UNIDA), IIF(A3.UNIHP = 'UN', 'u', A3.UNIHP))) AS UNIDADE,
                    CASE A3.SERIEO
                      WHEN '0' THEN 1
                      WHEN '1' THEN 2
                      WHEN '2' THEN 3
                      WHEN '3' THEN 4
                      WHEN '4' THEN 5
                      WHEN '5' THEN 6
                      WHEN '6' THEN 7
                      WHEN '7' THEN 8
                      WHEN '8' THEN 9
                      WHEN '9' THEN 10
                      WHEN 'A' THEN 11
                      WHEN 'B' THEN 12
                      WHEN 'C' THEN 13
                      WHEN 'D' THEN 14
                      WHEN 'E' THEN 15
                      WHEN 'F' THEN 16
                      WHEN 'G' THEN 19
                      WHEN 'H' THEN 20
                      WHEN 'I' THEN 21
                      WHEN 'J' THEN 22
                      WHEN 'K' THEN 23
                      WHEN 'L' THEN 24
                      WHEN 'M' THEN 25
                      WHEN 'N' THEN 26
                      WHEN 'O' THEN 27
                      WHEN 'P' THEN 28
                      WHEN 'Q' THEN 29
                      WHEN 'R' THEN 30
                      WHEN 'S' THEN 31
                      WHEN 'T' THEN 32
                      WHEN 'U' THEN 33
                      WHEN 'V' THEN 34
                      WHEN 'X' THEN 35
                      WHEN 'Z' THEN 36
                      WHEN 'Y' THEN 37
                      ELSE 38
                    END AS NUMEROFORMULA,
                    A3.ITEMID AS NUMEROLINHA,
                    0 AS INCLUSAOSISTEMA
                    FROM FC15110 A3
                      INNER JOIN FC03000 P ON P.CDPRO = A3.CDPRIN
                    WHERE A3.CODIGO_PS IS NOT NULL
                    AND P.CODIGO_PS < 4000000
                    INTO :CODIGO, :CODIGO_REF_A1, :CODIGO_PRODUTO, :QUANTIDADE, :UNIDADE, :NUMEROFORMULA, :NUMEROLINHA, :INCLUSAOSISTEMA
                    DO
                      BEGIN
                         I_SQL = 'INSERT INTO ATENDIMENTO_REF_A3(CODIGO, CODIGO_REF_A1, CODIGO_PRODUTO, QUANTIDADE, UNIDADE, NUMEROFORMULA, NUMEROLINHA, INCLUSAOSISTEMA) VALUES (?, ?, ?, ?, ?, ?, ?, ?)';
                         EXECUTE STATEMENT (:I_SQL)(:CODIGO, :CODIGO_REF_A1, :CODIGO_PRODUTO, :QUANTIDADE, :UNIDADE, :NUMEROFORMULA, :NUMEROLINHA, :INCLUSAOSISTEMA) ON EXTERNAL :CON AS USER :V_USER PASSWORD :V_SHA WITH COMMON TRANSACTION;
                      END

END;