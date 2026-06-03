package historydb

import (
	"strconv"
	"strings"
)

// EntityTypeFromOption mapeia opção de migração ("1".."8") para ENTITY_TYPE do banco central.
func EntityTypeFromOption(option string) int {
	n, err := strconv.Atoi(strings.TrimSpace(option))
	if err != nil || n < 1 || n > 8 {
		return 0
	}
	return n
}
