package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/primesoftwaresi/prime-migration-fyne/internal/models"
)

const (
	// Token fixo da API
	apiToken = "QVBJX1BSSU1FX05VQ0xFT19CNDQ"
	// URL da API de login
	loginAPIURL = "https://nucleo.primesoftware.com.br/login/usuario"
	// Timeout para requisições HTTP
	httpTimeout = 10 * time.Second
)

type AuthService struct {
	httpClient *http.Client
}

func NewAuthService() *AuthService {
	return &AuthService{
		httpClient: &http.Client{
			Timeout: httpTimeout,
		},
	}
}

// Authenticate faz login via API e retorna a sessão do usuário
func (as *AuthService) Authenticate(username, password string) (*models.UserSession, error) {
	if username == "" || password == "" {
		return nil, errors.New("usuário e senha são obrigatórios")
	}

	// Converter username para número (login é int na API)
	loginNum, err := strconv.Atoi(username)
	if err != nil {
		return nil, errors.New("login deve ser um número")
	}

	// Preparar requisição
	reqData := models.LoginRequest{
		Login: loginNum,
		Senha: password,
		Token: apiToken,
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("erro ao preparar requisição: %w", err)
	}

	// Criar requisição HTTP
	req, err := http.NewRequest("POST", loginAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Executar requisição
	resp, err := as.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar com a API: %w", err)
	}
	defer resp.Body.Close()

	// Ler resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta da API: %w", err)
	}

	// Parsear resposta JSON
	var loginResp models.LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return nil, fmt.Errorf("erro ao processar resposta da API: %w", err)
	}

	// Verificar status
	if loginResp.StatusCode == 403 {
		return nil, errors.New("dados informados estão incorretos")
	}

	if loginResp.StatusCode != 200 {
		return nil, fmt.Errorf("erro no login: %s", loginResp.Description)
	}

	// Verificar se há resultados
	if len(loginResp.Results) == 0 {
		return nil, errors.New("nenhum dado retornado pela API")
	}

	// Obter primeiro resultado
	resultData := loginResp.Results[0]

	// Verificar se é um array vazio (erro 403 retorna [[]])
	if resultArray, ok := resultData.([]interface{}); ok && len(resultArray) == 0 {
		return nil, errors.New("dados informados estão incorretos")
	}

	// Converter para JSON e depois para LoginResult
	resultJSON, err := json.Marshal(resultData)
	if err != nil {
		return nil, fmt.Errorf("erro ao processar dados do usuário: %w", err)
	}

	var result models.LoginResult
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("erro ao processar dados do usuário: %w", err)
	}

	// Criar sessão do usuário
	session := &models.UserSession{
		UserID:       result.CodigoFuncionario,
		Nome:         result.NomeFuncionario,
		Nick:         result.NickFuncionario,
		Departamento: result.NomeDepartamento,
		Funcao:       result.NomeFuncao,
	}

	return session, nil
}
