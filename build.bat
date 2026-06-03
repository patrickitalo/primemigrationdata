@echo off
REM Script para build do projeto com CGO habilitado

REM Mudar para o diretório onde o script está localizado
cd /d "%~dp0"

echo Configurando CGO e compiladores...
set CGO_ENABLED=1
set CC=gcc
set CXX=g++
set PATH=C:\msys64\ucrt64\bin;%PATH%

echo.
echo Diretorio atual: %CD%
echo.

echo Verificando GCC...
gcc --version
if errorlevel 1 (
    echo ERRO: GCC nao encontrado! Verifique se o MSYS2 esta instalado.
    pause
    exit /b 1
)

echo.
echo Fazendo build do projeto...
go build -o prime-migration.exe ./cmd/gui/main.go

if errorlevel 1 (
    echo.
    echo ERRO: Build falhou!
    pause
    exit /b 1
)

echo.
echo Build concluido com sucesso!
echo Executavel criado: prime-migration.exe
echo.
pause

