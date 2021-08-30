package operator

import (
	"github.com/daniilsolovey/BetBotGo/internal/config"
	"github.com/daniilsolovey/BetBotGo/internal/database"
)

type Operator struct {
	config   *config.Config
	database *database.Database
}

func NewOperator(
	config *config.Config,
	database *database.Database,
) *Operator {
	return &Operator{
		config:   config,
		database: database,
	}
}
