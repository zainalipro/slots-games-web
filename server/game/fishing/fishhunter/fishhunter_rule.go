package fishhunter

import (
	"github.com/slotopol/server/game/fishing"
)

// Game implements the Fish Feast fishing game.
type Game struct {
	fishing.FishingBase `yaml:",inline"`
}

// Declare conformity with FishingGame interface.
var _ fishing.FishingGame = (*Game)(nil)

// NewGame creates a new Fish Feast game instance.
func NewGame() *Game {
	return &Game{
		FishingBase: fishing.FishingBase{
			Bet:    1,
			Cannon: fishing.Cannon1,
		},
	}
}

// Scanner fires a shot at the current aim position.
// If the fish is caught, it's removed from the pool and a new one spawns.
// After each shot, some fish may swim away (despawn) and be replaced.
// Bomb fish (Octopus) trigger chain catches of nearby small fish.
func (g *Game) Scanner(result *fishing.ShotResult) error {
	// Resolve the shot at the aimed position
	g.ResolveShot(g.Aim, result)

	// After each shot, some fish may swim away
	g.Pool.ApplySwimAway()

	return nil
}

// SetCannon validates and sets the cannon level.
func (g *Game) SetCannon(cannon fishing.CannonLevel) error {
	return g.FishingBase.SetCannon(cannon)
}

// SetAim validates and sets the aim position.
func (g *Game) SetAim(aim int) error {
	return g.FishingBase.SetAim(aim)
}
