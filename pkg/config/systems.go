package config

// Lista de sistemas disponíveis
var SistemasDisponiveis = []string{
	"FCERTA",
	"PRISMA5",
	"AFRAM",
	"ALQUIMISTA",
	"FORMULARIUM",
	"HOSFARMA",
	"INTUITIVE",
	"MEDICATOR",
	"MULTIDATA",
	"PHARMACONTROL",
	"TRIER",
	"VITORIASOFT",
	"VSM",
}

// Sequência de arquivos SQL para o sistema FCERTA
// Ordem específica: procedures auxiliares primeiro, depois inserts, depois procedures de extração
var SequenciaSQLFCERTA = []string{
	// 1. UTIL_EXTRACT_FCERTA.sql - criada e executada separadamente no setup
	// 2-6. Procedures auxiliares (criadas antes das que dependem delas)
	"sps-fcerta/FN_TIRA_ACENTO.sql",
	"sps-fcerta/IS_NUMERIC.sql",
	"sps-fcerta/SP_REMOVE_NON_NUMERIC.SQL",
	"sps-fcerta/STRIP_NON_NUMERIC.sql",
	"sps-fcerta/UTIL_FORMATA_TEL_FC04400.sql",
	// 7. CIDADEESTADO.sql - inserts de dados
	"sps-fcerta/CIDADEESTADO.sql",
	// 8-15. Procedures de extração de dados
	"sps-fcerta/EXTRACT_DATA_FC02000.sql",
	"sps-fcerta/EXTRACT_DATA_FC03000.sql",
	"sps-fcerta/EXTRACT_DATA_FC03140.sql",
	"sps-fcerta/EXTRACT_DATA_FC04000.sql",
	"sps-fcerta/EXTRACT_DATA_FC05000.sql",
	"sps-fcerta/EXTRACT_DATA_FC07000.sql",
	"sps-fcerta/EXTRACT_DATA_REFAZER_1.sql",
	"sps-fcerta/EXTRACT_DATA_REFAZER_2.sql",
}

// Sequência de arquivos SQL para o sistema PRISMA5
var SequenciaSQLPRISMA5 = []string{
	"sps-prisma5/UTIL_EXTRACT_PRISMA5.sql",
	"sps-prisma5/TIRA_ACENTOS.sql",
	"sps-prisma5/STRIP_NON_NUMERIC.sql",
	"sps-prisma5/EXTRACT_DATA_CLIENTE.sql",
	"sps-prisma5/EXTRACT_DATA_MEDICO.sql",
	"sps-prisma5/EXTRACT_DATA_FORNECEDOR.sql",
	"sps-prisma5/EXTRACT_DATA_FORMAFARMACEUTICA.sql",
	"sps-prisma5/EXTRACT_DATA_PRODUTO.sql",
	"sps-prisma5/EXTRACT_DATA_LOTE.sql",
	"sps-prisma5/EXTRACT_DATA_PROD_INTERNA.sql",
	"sps-prisma5/UTIL_LIMPA_RTF.sql",
	"sps-prisma5/EXTRACT_DATA_REFAZER.sql",
}
