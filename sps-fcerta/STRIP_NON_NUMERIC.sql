create or alter procedure STRIP_NON_NUMERIC (
    INPUT_TEXT varchar(100))
returns (
    ONLY_NUMBERS varchar(100))
as
declare variable I integer;
declare variable CHAR_AT_POS char(1);
declare variable ASCII_CODE integer;
BEGIN
  ONLY_NUMBERS = '';
  I = 1;
  WHILE (I <= CHAR_LENGTH(INPUT_TEXT)) DO
  BEGIN
    CHAR_AT_POS = SUBSTRING(INPUT_TEXT FROM I FOR 1);
    ASCII_CODE = ASCII_VAL(CHAR_AT_POS);

    -- Verifica se é número (0-9)
    IF (ASCII_CODE BETWEEN 48 AND 57) THEN
    BEGIN
      ONLY_NUMBERS = ONLY_NUMBERS || CHAR_AT_POS;
    END

    I = I + 1;
  END
  SUSPEND;
END;