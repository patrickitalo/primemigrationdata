# Prime Migration Fyne - Guia de Setup

## Pré-requisitos

### 1. Instalar Go
- Download: https://golang.org/dl/
- Versão mínima: Go 1.21+

### 2. Compilador C (Windows)
O Fyne precisa de um compilador C para buildar no Windows. Opções:

**Opção A - MSYS2 (Recomendado):**
1. Instale MSYS2: https://www.msys2.org/
2. Abra MSYS2 UCRT64
3. Execute:
```bash
pacman -S --noconfirm mingw-w64-ucrt-x86_64-gcc
```
4. Adicione `C:\msys64\ucrt64\bin` ao PATH

**Opção B - MinGW-w64:**
- Download: https://www.mingw-w64.org/downloads/
- Adicione o `bin` ao PATH

### 3. Build do Projeto

```bash
# Windows PowerShell
$env:CGO_ENABLED=1
go build -o dist/prime-migration-fyne.exe ./cmd/gui

# Ou use o script
.\build.bat
```

## Estrutura do Projeto

```
prime-migration-v3/
├── cmd/gui/           # Ponto de entrada
├── internal/
│   ├── models/        # Modelos de dados
│   ├── services/      # Lógica de negócio
│   └── ui/            # Interface gráfica
├── pkg/               # Stubs locais (substituir pelo código real)
└── build.bat          # Script de build Windows
```

## Próximos Passos

1. **Copiar código real do projeto original:**
   - Copiar `pkg/` completo do projeto CLI original
   - Copiar pastas `sps-fcerta/` e `sps-prisma5/`

2. **Integrar API de Login:**
   - Quando receber os detalhes da API, atualizar `internal/services/auth_service.go`
   - Fazer chamada HTTP POST para autenticação

## Status Atual

✅ Estrutura criada
✅ Tela de Login implementada
✅ Janela principal com formulários
✅ Validações básicas
⏳ Aguardando código real do projeto original
⏳ Aguardando detalhes da API de autenticação

