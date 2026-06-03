create or alter procedure TIRA_ACENTOS (
    DADO varchar(512) = '')
returns (
    RETORNO varchar(512))
as
declare variable COM_ACENTO varchar(40) = 'àâêôûãõáéíóúçüÀÂÊÔÛÃÕÁÉÍÓÚÇÜÑñ';
declare variable SEM_ACENTO varchar(40) = 'aaeouaoaeioucuAAEOUAOAEIOUCUNn';
declare variable LETRA varchar(1) = '';
BEGIN
    RETORNO = '';
    WHILE (CHAR_LENGTH(DADO) > 0) DO
    BEGIN
        SELECT CASE SUBSTRING(:DADO FROM 1 FOR 1)
            WHEN 'à' THEN 'a'
            WHEN 'â' THEN 'a'
            WHEN 'ã' THEN 'a'
            WHEN 'á' THEN 'a'
            WHEN 'À' THEN 'A'
            WHEN 'Â' THEN 'A'
            WHEN 'Ã' THEN 'A'
            WHEN 'Á' THEN 'A'
            WHEN 'ê' THEN 'e'
            WHEN 'é' THEN 'e'
            WHEN 'Ê' THEN 'E'
            WHEN 'É' THEN 'E'
            WHEN 'ô' THEN 'o'
            WHEN 'õ' THEN 'o'
            WHEN 'ó' THEN 'o'
            WHEN 'Ô' THEN 'O'
            WHEN 'Ó' THEN 'O'
            WHEN 'Õ' THEN 'O'
            WHEN 'û' THEN 'u'
            WHEN 'ú' THEN 'u'
            WHEN 'ü' THEN 'u'
            WHEN 'Û' THEN 'U'
            WHEN 'Ú' THEN 'U'
            WHEN 'Ü' THEN 'U'
            WHEN 'í' THEN 'i'
            WHEN 'Í' THEN 'I'
            WHEN 'ç' THEN 'c'
            WHEN 'Ç' THEN 'C'
            WHEN 'ñ' THEN 'n'
            WHEN 'Ñ' THEN 'N'
            ELSE SUBSTRING(:DADO FROM 1 FOR 1)
        END
        FROM rdb$database INTO :LETRA;

        RETORNO = RETORNO || LETRA;

        DADO = SUBSTRING(DADO FROM 2 FOR 512);
    END

    SUSPEND;
END;