package aviator

import (
	"github.com/slotopol/server/game/crash"
)

const HouseEdge = 0.03 // 3% house edge

// Game implements an Aviator-style crash game.
type Game struct {
	crash.CrashBase `yaml:",inline"`
}

// Compile-time interface check.
var _ crash.CrashGame = (*Game)(nil)

// NewGame creates a new Aviator game.
func NewGame() *Game {
	return &Game{
		CrashBase: crash.CrashBase{
			Bet:   1,
			State: "waiting",
		},
	}
}

// Launch starts a new round by determining the crash point.
func (g *Game) Launch() error {
	if g.State == "flying" {
		return crash.ErrNoGame
	}
	g.State = "flying"
	g.CrashPoint = crash.GenerateCrashPoint(HouseEdge)
	g.Multiplier = 1.00
	g.CashedOut = false
	g.Payout = 0
	return nil
}

// CashOut cashes out at the current multiplier.
func (g *Game) CashOut() (float64, error) {
	if g.State != "flying" {
		return 0, crash.ErrNoGame
	}
	if g.CashedOut {
		return 0, crash.ErrAlreadyCashed
	}
	g.CashedOut = true
	g.Payout = g.Bet * g.Multiplier
	g.State = "cashed"
	return g.Payout, nil
}

// Tick advances the multiplier. Returns false if round has crashed.
// In a real implementation, this would be called by a game loop.
// For the API-driven version, the crash check happens on cashout.
func (g *Game) Tick() bool {
	if g.State != "flying" {
		return false
	}
	// Advance multiplier by a small increment
	if g.Multiplier < 2 {
		g.Multiplier += 0.01
	} else if g.Multiplier < 5 {
		g.Multiplier += 0.05
	} else {
		g.Multiplier += 0.10
	}
	g.Multiplier = crash.RoundTo2(g.Multiplier)

	// Check if we've passed the crash point
	if g.Multiplier >= g.CrashPoint {
		g.State = "crashed"
		g.Multiplier = crash.RoundTo2(g.CrashPoint)
		return false
	}
	return true
}

// Scanner is for compatibility with the Gamble interface.
func (g *Game) Scanner(wins *crash.ShotResult) error {
	wins.Pay = g.Payout
	return nil
}

// Spin is for compatibility with the Gamble interface.
func (g *Game) Spin(_ float64) {
	g.Launch()
	// In auto mode, simulate random cashout behavior
	// Real cashout decisions are made via the API
}

// GetAvatars returns dummy avatar count for multiplayer feel.
func (g *Game) GetAvatars() int {
	return 3 // simulate other players
}
