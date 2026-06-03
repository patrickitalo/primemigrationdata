create or alter procedure IS_NUMERIC (
    INPUT_STR varchar(255))
returns (
    IS_NUM smallint)
as
declare variable I integer;
declare variable CHAR_DIGIT varchar(1);
BEGIN
    IS_NUM = 1;
    I = 1;
    WHILE (I <= CHAR_LENGTH(INPUT_STR)) DO
    BEGIN
    CHAR_DIGIT = SUBSTRING(INPUT_STR FROM I FOR 1);
    IF (CHAR_DIGIT < '0' OR CHAR_DIGIT > '9') THEN
    BEGIN
      IS_NUM = 0;
      LEAVE;
    END
    I = I + 1;
    END
  SUSPEND;
    END;