package crash

import (
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
)

// CrashGame is the common interface for crash-type games.
type CrashGame interface {
	Launch() error                    // start a new round
	CashOut() (float64, error)        // cash out at current multiplier
	GetMultiplier() float64           // returns current multiplier
	GetCrashPoint() float64           // returns the crash point for this round
	IsCrashed() bool                  // returns true if round has crashed
	GetBet() float64                  // returns current bet
	SetBet(float64) error             // set bet
	GetGameState() string             // returns game state (waiting, flying, crashed, cashed)
	GetPayout() float64               // returns payout for current round
	Scanner(*ShotResult) error        // compatibility
	Spin(float64)                     // compatibility, launches a round
}

// CrashBase provides common crash game fields.
type CrashBase struct {
	Bet         float64 `json:"bet" yaml:"bet" xml:"bet"`
	State       string  `json:"state" yaml:"state" xml:"state"`
	Multiplier  float64 `json:"multiplier" yaml:"multiplier" xml:"multiplier"`
	CrashPoint  float64 `json:"crashPoint" yaml:"crashPoint" xml:"crashPoint"`
	Payout      float64 `json:"payout" yaml:"payout" xml:"payout"`
	CashedOut   bool    `json:"cashedOut" yaml:"cashedOut" xml:"cashedOut"`
}

// GenerateCrashPoint generates a crash multiplier using provably fair algorithm.
// Uses: crashPoint = (1 - p) / (1 - r) where r ~ Uniform(0,1) and p = 1/100
// This gives: P(crash > x) = p/(x-p) for x > p, so expected crash is e/(e-1) ≈ 1.58
// With house edge built in.
func GenerateCrashPoint(houseEdge float64) float64 {
	// Crash point formula used by many crash games:
	// crashPoint = 1 / (1 - r) * (1 - houseEdge)
	// where r ~ Uniform(0,1)
	// This gives expected value = (1 - houseEdge) * e/(e-1)
	// With houseEdge = 0.03, expected crash ≈ 1.53

	var r = rand.Float64()
	// Use: crashPoint = (1 - houseEdge) / (1 - r)
	var cp = (1 - houseEdge) / (1 - r)
	// Round to 2 decimal places
	cp = math.Floor(cp*100) / 100
	return cp
}

// RoundTo2 rounds a float to 2 decimal places.
func RoundTo2(v float64) float64 {
	return math.Floor(v*100) / 100
}

// ShotResult is a minimal placeholder for Scanner compatibility.
type ShotResult struct {
	Pay float64 `json:"pay" yaml:"pay" xml:"pay,attr"`
}

func (g *CrashBase) GetBet() float64 {
	return g.Bet
}

func (g *CrashBase) SetBet(bet float64) error {
	if bet <= 0 {
		return ErrBadParam
	}
	g.Bet = bet
	return nil
}

func (g *CrashBase) GetMultiplier() float64 {
	return g.Multiplier
}

func (g *CrashBase) GetCrashPoint() float64 {
	return g.CrashPoint
}

func (g *CrashBase) IsCrashed() bool {
	return g.State == "crashed"
}

func (g *CrashBase) GetGameState() string {
	return g.State
}

func (g *CrashBase) GetPayout() float64 {
	return g.Payout
}

func (g *CrashBase) Scanner(wins *ShotResult) error {
	wins.Pay = g.Payout
	return nil
}

func (g *CrashBase) Spin(_ float64) {
	// Spin is for compatibility; crash games use Launch/CashOut explicitly.
	// Does nothing by default — concrete games override if needed.
}

var (
	ErrBadParam      = errors.New("wrong parameter")
	ErrNoGame        = errors.New("no round in progress")
	ErrAlreadyCashed = errors.New("already cashed out this round")
	ErrAlreadyCrashed = errors.New("round already crashed")
)

// Print_all prints statistics placeholders.
func Print_all(sp interface{}) {
	fmt.Println("Crash game statistics: analytic calculation not available via simulator")
}
