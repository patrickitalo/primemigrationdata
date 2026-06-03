create or alter procedure UTIL_LIMPA_RTF (
    TEXTO_RTF varchar(8000))
returns (
    TEXTO_LIMPO varchar(8000))
as
declare variable TEXTO_TEMP varchar(8000);
BEGIN
    TEXTO_TEMP = :TEXTO_RTF;

    IF (TRIM(COALESCE(:TEXTO_TEMP, '')) = '') THEN
    BEGIN
        TEXTO_LIMPO = NULL;
        SUSPEND;
    END
    ELSE
    BEGIN
        -- ETAPA 1: Substitui caracteres especiais (formato \'xx)
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\''c0', ASCII_CHAR(192)); -- À
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\''c1', ASCII_CHAR(193)); -- Á
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\''c2', ASCII_CHAR(194)); -- Â
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\''c3', ASCII_CHAR(195)); -- Ã
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\''c7', ASCII_CHAR(199)); -- Ç
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\''c8', ASCII_CHAR(200)); -- È
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\''c9', ASCII_CHAR(201)); -- É
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\''ca', ASCII_CHAR(202)); -- Ê
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\''cc', ASCII_CHAR(204)); -- Ì
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\''cd', ASCII_CHAR(205)); -- Í
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\''d2', ASCII_CHAR(210)); -- Ò
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\''d3', ASCII_CHAR(211)); -- Ó
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\''d4', ASCII_CHAR(212)); -- Ô
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\''d5', ASCII_CHAR(213)); -- Õ
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\''d9', ASCII_CHAR(217)); -- Ù
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\''da', ASCII_CHAR(218)); -- Ú
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\''dc', ASCII_CHAR(220)); -- Ü
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\''e0', ASCII_CHAR(224)); -- à
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\''e1', ASCII_CHAR(225)); -- á
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\''e2', ASCII_CHAR(226)); -- â
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\''e3', ASCII_CHAR(227)); -- ã
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\''e7', ASCII_CHAR(231)); -- ç
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\''e8', ASCII_CHAR(232)); -- è
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\''e9', ASCII_CHAR(233)); -- é
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\''ea', ASCII_CHAR(234)); -- ê
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\''ec', ASCII_CHAR(236)); -- ì
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\''ed', ASCII_CHAR(237)); -- í
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\''f2', ASCII_CHAR(242)); -- ò
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\''f3', ASCII_CHAR(243)); -- ó
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\''f4', ASCII_CHAR(244)); -- ô
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\''f5', ASCII_CHAR(245)); -- õ
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\''f9', ASCII_CHAR(249)); -- ù
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\''fa', ASCII_CHAR(250)); -- ú
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\''fc', ASCII_CHAR(252)); -- ü

        -- ETAPA 2: Remove tags e cabeçalhos RTF
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\ansicpg1252', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\rtf1', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\ansi', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\deff0', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\deflang1046', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\lang1046', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '{\fonttbl', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '{\colortbl', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\viewkind4', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\uc1', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\pard', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\b0', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\b', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\i0', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\i', '');
        
        -- Tags de fonte e família
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\f0', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\f1', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\f2', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\f3', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\fswiss', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\fbidis', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\fprq2', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\fs16', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\fs18', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\fs20', '');

        -- Tags de cor
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\cf0', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\cf1', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\red0', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\green0', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\blue0', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, 'lue0', '');         /* NOVO - Para corrigir dados malformados */

        -- Tags de parágrafo e quebra de linha
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\ltrpar', ' ');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\par', ' ');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\line', ' ');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\sa160', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\sl252', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\slmult1', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\qj', '');

        -- ETAPA 3: Limpeza de resíduos de tags
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\fnil', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\fcharset0', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, 'MS Sans Serif;', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, 'Courier New;', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, 'Arial;', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, 'Calibri;', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, 'Arial', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, 'Calibri', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '{', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '}', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '\', '');
        TEXTO_TEMP = REPLACE(:TEXTO_TEMP, ';', '');

        -- ETAPA 4: Limpeza final do texto
        WHILE (POSITION('  ' IN :TEXTO_TEMP) > 0) DO
            TEXTO_TEMP = REPLACE(:TEXTO_TEMP, '  ', ' ');

        TEXTO_LIMPO = TRIM(:TEXTO_TEMP);
        SUSPEND;
    END
END;