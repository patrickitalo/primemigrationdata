package models

type ClientConfig struct {
	ID                int    `json:"id"`
	CodigoCliente     string `json:"codigo_cliente"`
	SistemaOrigem     string `json:"sistema_origem"`
	DbPath            string `json:"db_path"`
	DbUser            string `json:"db_user"`
	DbPassword        string `json:"db_password"`
	DbHost            string `json:"db_host"`
	DbPort            string `json:"db_port"`
	AliasPharmacie    string `json:"alias_pharmacie"`
	IpServerPharmacie string `json:"ipserver_pharmacie"`
	PortaPharmacie    string `json:"porta_pharmacie"`
	Conversao         string `json:"conversao"`
}
