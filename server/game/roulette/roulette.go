package roulette

import (
	"errors"
	"fmt"
	"math/rand/v2"
)

// BetType represents different roulette bet types.
type BetType string

const (
	BetStraightUp  BetType = "straight_up"   // single number 35:1
	BetSplit       BetType = "split"         // two numbers    17:1
	BetStreet      BetType = "street"        // three numbers  11:1
	BetCorner      BetType = "corner"        // four numbers   8:1
	BetLine        BetType = "line"          // six numbers    5:1
	BetDozen       BetType = "dozen"         // 1-12,13-24,25-36 2:1
	BetColumn      BetType = "column"        // 2:1
	BetRed         BetType = "red"           // 1:1
	BetBlack       BetType = "black"         // 1:1
	BetEven        BetType = "even"          // 1:1
	BetOdd         BetType = "odd"           // 1:1
	BetLow         BetType = "low"           // 1-18 1:1
	BetHigh        BetType = "high"          // 19-36 1:1
)

// RouletteGame is the common interface for roulette games.
type RouletteGame interface {
	Spin(float64)
	GetBet() float64
	SetBet(float64) error
	GetResult() string
	GetPayout() float64
	GetGameState() string
	GetNumber() int
	GetColor() string
	GetBetType() BetType
	SetBetType(BetType) error
	GetBetNumber() int
	SetBetNumber(int) error
}

// European roulette wheel numbers with colors.
type WheelNumber struct {
	Number int
	Color  string
}

var EuropeanWheel = []WheelNumber{
	{0, "green"},
	{32, "red"}, {15, "black"}, {19, "red"}, {4, "black"}, {21, "red"},
	{2, "black"}, {25, "red"}, {17, "black"}, {34, "red"}, {6, "black"},
	{27, "red"}, {13, "black"}, {36, "red"}, {11, "black"}, {30, "red"},
	{8, "black"}, {23, "red"}, {10, "black"}, {5, "red"}, {24, "black"},
	{16, "red"}, {33, "black"}, {1, "red"}, {20, "black"}, {14, "red"},
	{31, "black"}, {9, "red"}, {22, "black"}, {18, "red"}, {29, "black"},
	{7, "red"}, {28, "black"}, {12, "red"}, {35, "black"}, {3, "red"},
	{26, "black"},
}

// GetColor returns the color of a number on the European wheel.
func GetColor(num int) string {
	for _, wn := range EuropeanWheel {
		if wn.Number == num {
			return wn.Color
		}
	}
	return "green"
}

// PayoutMultiplier returns the payout multiplier for a bet type.
func PayoutMultiplier(bt BetType) int {
	switch bt {
	case BetStraightUp:
		return 35
	case BetSplit:
		return 17
	case BetStreet:
		return 11
	case BetCorner:
		return 8
	case BetLine:
		return 5
	case BetDozen, BetColumn:
		return 2
	case BetRed, BetBlack, BetEven, BetOdd, BetLow, BetHigh:
		return 1
	default:
		return 0
	}
}

// CheckWin checks if a bet wins and returns the payout multiplier.
func CheckWin(number int, betType BetType, betNumber int) int {
	var num = number
	var color = GetColor(num)

	switch betType {
	case BetStraightUp:
		if num == betNumber {
			return PayoutMultiplier(BetStraightUp)
		}
	case BetSplit:
		// betNumber is the first of the pair, pair number is betNumber+1
		if num == betNumber || num == betNumber+1 {
			return PayoutMultiplier(BetSplit)
		}
	case BetStreet:
		// betNumber is the first of the street (1-3, 4-6, etc.)
		if num >= betNumber && num <= betNumber+2 {
			return PayoutMultiplier(BetStreet)
		}
	case BetCorner:
		// betNumber is the top-left of the 2x2 corner
		if num == betNumber || num == betNumber+1 || num == betNumber+3 || num == betNumber+4 {
			return PayoutMultiplier(BetCorner)
		}
	case BetLine:
		// betNumber is the first of 6 numbers
		if num >= betNumber && num <= betNumber+5 {
			return PayoutMultiplier(BetLine)
		}
	case BetDozen:
		switch betNumber {
		case 1: // 1-12
			if num >= 1 && num <= 12 {
				return PayoutMultiplier(BetDozen)
			}
		case 2: // 13-24
			if num >= 13 && num <= 24 {
				return PayoutMultiplier(BetDozen)
			}
		case 3: // 25-36
			if num >= 25 && num <= 36 {
				return PayoutMultiplier(BetDozen)
			}
		}
	case BetColumn:
		// betNumber represents column 1, 2, or 3
		if num > 0 && num%3 == betNumber%3 {
			return PayoutMultiplier(BetColumn)
		}
	case BetRed:
		if color == "red" {
			return PayoutMultiplier(BetRed)
		}
	case BetBlack:
		if color == "black" {
			return PayoutMultiplier(BetBlack)
		}
	case BetEven:
		if num > 0 && num%2 == 0 {
			return PayoutMultiplier(BetEven)
		}
	case BetOdd:
		if num > 0 && num%2 == 1 {
			return PayoutMultiplier(BetOdd)
		}
	case BetLow:
		if num >= 1 && num <= 18 {
			return PayoutMultiplier(BetLow)
		}
	case BetHigh:
		if num >= 19 && num <= 36 {
			return PayoutMultiplier(BetHigh)
		}
	}
	return 0
}

// RouletteBase provides common roulette game fields.
type RouletteBase struct {
	Bet       float64 `json:"bet" yaml:"bet" xml:"bet"`
	State     string  `json:"state" yaml:"state" xml:"state"`
	Payout    float64 `json:"payout" yaml:"payout" xml:"payout"`
	Result    string  `json:"result" yaml:"result" xml:"result"`
	Number    int     `json:"number" yaml:"number" xml:"number"`
	Color     string  `json:"color" yaml:"color" xml:"color"`
	BetType   BetType `json:"betType" yaml:"betType" xml:"betType"`
	BetNumber int     `json:"betNumber" yaml:"betNumber" xml:"betNumber"`
}

func (g *RouletteBase) GetBet() float64 {
	return g.Bet
}

func (g *RouletteBase) SetBet(bet float64) error {
	if bet <= 0 {
		return errors.New("wrong parameter")
	}
	g.Bet = bet
	return nil
}

func (g *RouletteBase) GetPayout() float64 {
	return g.Payout
}

func (g *RouletteBase) GetResult() string {
	return g.Result
}

func (g *RouletteBase) GetGameState() string {
	return g.State
}

func (g *RouletteBase) GetNumber() int {
	return g.Number
}

func (g *RouletteBase) GetColor() string {
	return g.Color
}

func (g *RouletteBase) GetBetType() BetType {
	return g.BetType
}

func (g *RouletteBase) SetBetType(bt BetType) error {
	switch bt {
	case BetStraightUp, BetSplit, BetStreet, BetCorner, BetLine,
		BetDozen, BetColumn, BetRed, BetBlack, BetEven, BetOdd, BetLow, BetHigh:
		g.BetType = bt
		return nil
	default:
		return errors.New("invalid bet type")
	}
}

func (g *RouletteBase) GetBetNumber() int {
	return g.BetNumber
}

func (g *RouletteBase) SetBetNumber(n int) error {
	if n < 0 || n > 36 {
		return errors.New("number must be 0-36")
	}
	g.BetNumber = n
	return nil
}

// doSpin spins the roulette wheel.
func (g *RouletteBase) doSpin() {
	g.State = "spinning"
	g.Payout = 0
	g.Result = ""

	// Spin the wheel
	var idx = rand.IntN(len(EuropeanWheel))
	var wn = EuropeanWheel[idx]
	g.Number = wn.Number
	g.Color = wn.Color

	// Check win
	var mult = CheckWin(g.Number, g.BetType, g.BetNumber)
	if mult > 0 {
		g.Payout = g.Bet * float64(mult+1) // stake + winnings
		g.Result = fmt.Sprintf("Win! Number %d %s pays %d:1", g.Number, g.Color, mult)
	} else {
		g.Result = fmt.Sprintf("Lose. Number %d %s", g.Number, g.Color)
	}
	g.State = "done"
}

// Spin is for Gamble interface compatibility.
func (g *RouletteBase) Spin(_ float64) {
	g.doSpin()
}

// ShotResult is for Scanner compatibility.
type ShotResult struct {
	Pay float64 `json:"pay" yaml:"pay" xml:"pay,attr"`
}

func (g *RouletteBase) Scanner(wins *ShotResult) error {
	wins.Pay = g.Payout
	return nil
}

var (
	ErrBadParam = errors.New("wrong parameter")
	ErrNoGame   = errors.New("roulette round not started")
)

// Print_all prints statistics placeholders for roulette.
func Print_all(sp interface{}) {
	fmt.Println("Roulette statistics: European roulette house edge 2.7%")
}
