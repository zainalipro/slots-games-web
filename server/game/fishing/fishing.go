package fishing

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"sort"

	"github.com/slotopol/server/game"
)

// FishType represents fish species.
type FishType byte

const (
	FishGuppy   FishType = 1 // small common fish, 4× payout
	FishPerch   FishType = 2 // medium fish, 12× payout
	FishPike    FishType = 3 // large fish, 28× payout
	FishShark   FishType = 4 // very large, 70× payout
	FishOctopus FishType = 5 // bomb fish, 50× + chain catches nearby fish
	FishDragon  FishType = 6 // jackpot fish, 350× payout
	FishNone    FishType = 0
)

// SpawnWeights defines spawn probability for each fish type (index = FishType-1).
var SpawnWeights = [6]int{350, 250, 180, 100, 70, 50} // total = 1000

// FishMult returns the payout multiplier for each fish type (index = FishType-1).
var FishMult = [6]float64{4, 12, 28, 70, 50, 350}

// CannonLevel defines the cannon power.
type CannonLevel int

const (
	Cannon1 CannonLevel = 1 + iota // 1× cost
	Cannon2                        // 2× cost
	Cannon3                        // 3× cost
	Cannon4                        // 4× cost
	Cannon5                        // 5× cost
)

// CannonCost returns the bet cost multiplier for each cannon level.
var CannonCost = [5]float64{1, 2, 3, 4, 5}

// CannonCatch defines catch probability per fish type for each cannon level.
// Index: [cannonLevel-1][fishType-1]
var CannonCatch = [5][6]float64{
	{0.30, 0.10, 0.025, 0.005, 0.014, 0.0017}, // C1
	{0.45, 0.18, 0.06, 0.015, 0.035, 0.005},    // C2
	{0.53, 0.28, 0.10, 0.035, 0.07, 0.012},     // C3
	{0.62, 0.35, 0.15, 0.05, 0.10, 0.022},      // C4
	{0.68, 0.42, 0.20, 0.07, 0.14, 0.035},      // C5
}

// Fish represents a single fish in the pool.
type Fish struct {
	Type  FishType `json:"type" yaml:"type" xml:"type,attr"`
	HP    int      `json:"hp" yaml:"hp" xml:"hp,attr"`
	MaxHP int      `json:"maxHp" yaml:"maxHp" xml:"maxHp,attr"`
	Pos   int      `json:"pos" yaml:"pos" xml:"pos,attr"`
}

// Pool is a collection of fish on screen.
type Pool []Fish

const PoolSize = 12

// Fill populates the pool with random fish.
func (p *Pool) Fill() {
	*p = make(Pool, PoolSize)
	for i := range PoolSize {
		(*p)[i] = newFish(i)
	}
}

func newFish(pos int) Fish {
	var ft = randomFishType()
	return Fish{
		Type:  ft,
		HP:    fishHP(ft),
		MaxHP: fishHP(ft),
		Pos:   pos,
	}
}

// fishHP returns the visual HP for a fish type.
// HP is cosmetic but gives the player a sense of progress.
func fishHP(ft FishType) int {
	switch ft {
	case FishGuppy:
		return 1
	case FishPerch:
		return 2
	case FishPike:
		return 4
	case FishShark:
		return 8
	case FishOctopus:
		return 6
	case FishDragon:
		return 15
	default:
		return 1
	}
}

func randomFishType() FishType {
	var r = rand.IntN(1000)
	var cum int
	for ft := FishGuppy; ft <= FishDragon; ft++ {
		cum += SpawnWeights[ft-1]
		if r < cum {
			return ft
		}
	}
	return FishGuppy
}

// SwimAwayChance is the probability per fish per shot that it swims away.
const SwimAwayChance = 0.04

// ApplySwimAway removes some fish from the pool randomly and replaces them.
func (p *Pool) ApplySwimAway() {
	for i := range *p {
		if (*p)[i].Type == FishNone {
			continue
		}
		if rand.Float64() < SwimAwayChance {
			(*p)[i] = newFish(i)
		}
	}
}

// ReplaceFish replaces a fish at the given position with a new random fish.
func (p *Pool) ReplaceFish(pos int) {
	if pos >= 0 && pos < len(*p) {
		(*p)[pos] = newFish(pos)
	}
}

// ShotResult represents the outcome of a shot.
type ShotResult struct {
	Fish     FishType     `json:"fish" yaml:"fish" xml:"fish,attr"`
	Hit      bool         `json:"hit" yaml:"hit" xml:"hit,attr"`
	Catch    bool         `json:"catch" yaml:"catch" xml:"catch,attr"`
	Pay      float64      `json:"pay" yaml:"pay" xml:"pay,attr"`
	Pos      int          `json:"pos" yaml:"pos" xml:"pos,attr"`
	Chain    []ShotResult `json:"chain,omitempty" yaml:"chain,omitempty" xml:"chain,omitempty"`
}

// FishingGame is the common interface for all fishing games.
type FishingGame interface {
	Scanner(*ShotResult) error  // fire a shot at the aimed position
	Spin(float64)               // populate the pool with fresh fish
	GetBet() float64            // returns current base bet
	SetBet(float64) error       // set base bet
	GetCannon() CannonLevel     // returns current cannon level
	SetCannon(CannonLevel) error // set cannon level
	GetPool() Pool              // returns current pool state
	GetAim() int                // returns current aimed position
	SetAim(int) error           // set aimed position
}

var (
	ErrBadParam       = errors.New("wrong parameter")
	ErrBadCannon      = errors.New("invalid cannon level")
	ErrAimOutOfBounds = errors.New("aim position is out of pool bounds")
)

// FishingBase provides common fields for fishing games.
type FishingBase struct {
	Pool   Pool        `json:"pool" yaml:"pool" xml:"pool"`
	Bet    float64     `json:"bet" yaml:"bet" xml:"bet"`
	Aim    int         `json:"aim" yaml:"aim" xml:"aim"`
	Cannon CannonLevel `json:"cannon" yaml:"cannon" xml:"cannon"`
}

func (g *FishingBase) Spin(_ float64) {
	g.Pool.Fill()
}

func (g *FishingBase) GetBet() float64 {
	return g.Bet
}

func (g *FishingBase) SetBet(bet float64) error {
	if bet <= 0 {
		return ErrBadParam
	}
	g.Bet = bet
	return nil
}

func (g *FishingBase) GetCannon() CannonLevel {
	return g.Cannon
}

func (g *FishingBase) SetCannon(cannon CannonLevel) error {
	if cannon < Cannon1 || cannon > Cannon5 {
		return ErrBadCannon
	}
	g.Cannon = cannon
	return nil
}

func (g *FishingBase) GetPool() Pool {
	return g.Pool
}

func (g *FishingBase) GetAim() int {
	return g.Aim
}

func (g *FishingBase) SetAim(aim int) error {
	if aim < 0 || aim >= len(g.Pool) {
		return ErrAimOutOfBounds
	}
	g.Aim = aim
	return nil
}

// ResolveShot attempts to catch the fish at the given pool position.
// Returns the shot result including chain catches for bomb (Octopus) fish.
func (g *FishingBase) ResolveShot(pos int, result *ShotResult) {
	result.Pos = pos
	if pos < 0 || pos >= len(g.Pool) {
		result.Hit = false
		result.Catch = false
		result.Pay = 0
		return
	}

	var fish = g.Pool[pos]
	if fish.Type == FishNone {
		result.Hit = false
		result.Catch = false
		result.Pay = 0
		return
	}

	result.Fish = fish.Type

	// Determine catch based on cannon level and fish type
	var catchProb float64
	if g.Cannon >= Cannon1 && g.Cannon <= Cannon5 {
		catchProb = CannonCatch[g.Cannon-1][fish.Type-1]
	} else {
		catchProb = 0
	}

	var caught = rand.Float64() < catchProb
	result.Hit = true
	result.Catch = caught

	if caught {
		result.Pay = FishMult[fish.Type-1] * g.Bet
		g.Pool.ReplaceFish(pos)

		// Octopus (bomb): chain-catch 2 random small fish
		if fish.Type == FishOctopus {
			result.Chain = g.doBombChain()
		}
	} else {
		result.Pay = 0
		// Decrease HP visually on a miss (fish gets "worn down")
		if g.Pool[pos].HP > 0 {
			g.Pool[pos].HP--
		}
	}
}

// doBombChain catches 2 random Guppy/Perch fish from the pool as chain reaction.
func (g *FishingBase) doBombChain() []ShotResult {
	var chain []ShotResult
	for range 2 {
		// Find small fish in pool
		var targets []int
		for i, f := range g.Pool {
			if f.Type == FishGuppy || f.Type == FishPerch {
				targets = append(targets, i)
			}
		}
		if len(targets) == 0 {
			break
		}
		var idx = targets[rand.IntN(len(targets))]
		var f = g.Pool[idx]
		var pay = FishMult[f.Type-1] * g.Bet
		chain = append(chain, ShotResult{
			Fish:   f.Type,
			Hit:    true,
			Catch:  true,
			Pay:    pay,
			Pos:    idx,
		})
		g.Pool.ReplaceFish(idx)
	}
	return chain
}

// CalcStat computes the theoretical RTP and variance for each cannon level,
// then reports the average across all cannons.
func CalcStat(ctx context.Context, sp *game.ScanPar) (float64, float64) {
	fmt.Println()
	fmt.Println("Fishing game — Fish Feast")
	fmt.Printf("Pool size: %d fish, %d fish types\n", PoolSize, 6)

	var totalEV, totalE2 float64
	for cannon := Cannon1; cannon <= Cannon5; cannon++ {
		var ev, e2 float64
		var cost = CannonCost[cannon-1]
		for ft := FishGuppy; ft <= FishDragon; ft++ {
			var prob = float64(SpawnWeights[ft-1]) / 1000
			var catch = CannonCatch[cannon-1][ft-1]
			var mult = FishMult[ft-1]
			// Expected return per shot = spawn_prob * catch_prob * mult / cost
			var evFt = prob * catch * mult / cost
			ev += evFt
			e2 += prob * catch * mult * mult / (cost * cost)
		}
		totalEV += ev
		totalE2 += e2
		fmt.Printf("Cannon %d (%.0f× cost): RTP = %.4g%%\n", cannon, cost, ev*100)
	}

	var avgEV = totalEV / 5
	var avgE2 = totalE2 / 5
	var avgD = avgE2 - avgEV*avgEV

	fmt.Println()
	fmt.Printf("Average RTP = %.6g%%\n", avgEV*100)
	Print_all(sp, avgEV, avgD)
	return avgEV, avgD
}

// Print_all prints statistics for the given RTP and variance.
func Print_all(sp *game.ScanPar, ev, D float64) {
	if sp.IsMain() {
		fmt.Printf("RTP = %.8g%%\n", ev*100)
	}
	if sp.IsVI() {
		var sigma = math.Sqrt(D)
		var vi = game.GetZ(sp.Conf) * sigma
		fmt.Printf("sigma = %.6g, VI[%.4g%%] = %.6g (%s)\n", sigma, sp.Conf*100, vi, game.VIname5[game.VIclass5(sigma)])
	}
	if sp.IsCI() && ev < game.RTPconv {
		var sigma = math.Sqrt(D)
		var ci = game.CI(sp.Conf, ev, sigma)
		var BRci = game.BankrollPlayer(sp.Conf, ev, sigma, ci)
		fmt.Printf("CI[%.4g%%] = %d, bankroll[CI] = %.6g\n", sp.Conf*100, int(ci+0.5), BRci)
	}
	if sp.IsSpread() {
		fmt.Println()
		fmt.Printf("RTP spread for spins number with confidence %.4g%%:\n", sp.Conf*100)
		var N = []int{1e3, 1e4, 1e5, 1e6, 1e7}
		var sigma = math.Sqrt(D)
		var vi = game.GetZ(sp.Conf) * sigma
		var ci = game.CI(sp.Conf, ev, sigma)
		if ci < 1e7 {
			N = append(N, int(ci+0.5))
			sort.Ints(N)
		}
		for _, n := range N {
			var Δ = vi / math.Sqrt(float64(n))
			fmt.Printf("%8d: %.2f%% ... %.2f%%\n", n, (ev-Δ)*100, (ev+Δ)*100)
		}
	}
}
