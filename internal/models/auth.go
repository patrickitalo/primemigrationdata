package models

// LoginRequest representa a requisição de login
type LoginRequest struct {
	Login int    `json:"login"`
	Senha string `json:"senha"`
	Token string `json:"token"`
}

// LoginResponse representa a resposta da API
type LoginResponse struct {
	StatusCode        int           `json:"statusCode"`
	StatusDescription string        `json:"statusDescription"`
	Results           []interface{} `json:"results"`
	Description       string        `json:"description"`
	APIVersion        string        `json:"apiVersion"`
}

// LoginResult representa os dados do funcionário logado
type LoginResult struct {
	CodigoFuncionario  int    `json:"codigo_funcionario"`
	NomeFuncionario    string `json:"nome_funcionario"`
	NickFuncionario    string `json:"nick_funcionario"`
	CodigoDepartamento int    `json:"codigo_departamento"`
	NomeDepartamento   string `json:"nome_departamento"`
	CodigoFuncao       int    `json:"codigo_funcao"`
	NomeFuncao         string `json:"nome_funcao"`
}

// UserSession representa a sessão do usuário autenticado
type UserSession struct {
	UserID       int
	Nome         string
	Nick         string
	Departamento string
	Funcao       string
}
