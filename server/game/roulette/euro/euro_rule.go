package euro

import (
	"github.com/slotopol/server/game/roulette"
)

// Game implements European Roulette.
type Game struct {
	roulette.RouletteBase `yaml:",inline"`
}

// Compile-time interface check.
var _ roulette.RouletteGame = (*Game)(nil)

// NewGame creates a new European Roulette game.
func NewGame() *Game {
	return &Game{
		RouletteBase: roulette.RouletteBase{
			Bet:       1,
			State:     "waiting",
			BetType:   roulette.BetStraightUp,
			BetNumber: 1,
		},
	}
}
