create or alter procedure SP_REMOVE_NON_NUMERIC (
    "INPUT" varchar(255))
returns (
    "OUTPUT" varchar(255))
as
declare variable I integer;
declare variable CURRENT_CHAR char(1);
BEGIN
    OUTPUT = '';
    i = 1;
    WHILE (i <= CHAR_LENGTH(INPUT)) DO
    BEGIN
        current_char = SUBSTRING(INPUT FROM i FOR 1);
        IF (current_char BETWEEN '0' AND '9') THEN
            OUTPUT = OUTPUT || current_char;
        i = i + 1;
    END
    SUSPEND; -- Retorna o resultado
END;