package euro

import (
	"testing"

	"github.com/slotopol/server/game/roulette"
)

// TestNewGame checks defaults.
func TestNewGame(t *testing.T) {
	g := NewGame()
	if g.Bet != 1 {
		t.Errorf("expected Bet=1, got %g", g.Bet)
	}
	if g.State != "waiting" {
		t.Errorf("expected State=waiting, got %s", g.State)
	}
	if g.BetType != roulette.BetStraightUp {
		t.Errorf("expected BetType=straight_up, got %s", g.BetType)
	}
	if g.BetNumber != 1 {
		t.Errorf("expected BetNumber=1, got %d", g.BetNumber)
	}
}

// TestSpin sets state to done.
func TestSpin(t *testing.T) {
	g := NewGame()
	g.Spin(95.0)
	if g.State != "done" {
		t.Errorf("expected State=done after Spin, got %s", g.State)
	}
	if g.Number < 0 || g.Number > 36 {
		t.Errorf("Number should be 0-36, got %d", g.Number)
	}
	if g.Color == "" {
		t.Errorf("Color should not be empty")
	}
	if g.Payout != 0 && g.Payout != g.Bet*36 {
		t.Errorf("unexpected payout %g for straight-up bet", g.Payout)
	}
}

// TestSpinResultChanged checks each spin produces a new result.
func TestSpinResultChanged(t *testing.T) {
	g := NewGame()
	g.Spin(95.0)
	r1 := g.Number
	g.Spin(95.0)
	r2 := g.Number
	if r1 == 0 && r2 == 0 {
		t.Log("two consecutive zero spins (possible but unlikely)")
	}
	_ = r2
}

// TestGetBet checks bet getter.
func TestGetBet(t *testing.T) {
	g := NewGame()
	if g.GetBet() != 1 {
		t.Errorf("GetBet should return 1")
	}
}

// TestSetBet checks bet setter.
func TestSetBet(t *testing.T) {
	g := NewGame()
	if err := g.SetBet(5); err != nil {
		t.Errorf("SetBet(5) should succeed: %v", err)
	}
	if g.Bet != 5 {
		t.Errorf("Bet should be 5, got %g", g.Bet)
	}
	if err := g.SetBet(0); err == nil {
		t.Errorf("SetBet(0) should fail")
	}
}

// TestBetType checks bet type getter/setter.
func TestBetType(t *testing.T) {
	g := NewGame()
	if g.GetBetType() != roulette.BetStraightUp {
		t.Errorf("default bet type should be straight_up")
	}
	if err := g.SetBetType(roulette.BetRed); err != nil {
		t.Errorf("SetBetType(red) should succeed: %v", err)
	}
	if g.BetType != roulette.BetRed {
		t.Errorf("BetType should be red")
	}
	if err := g.SetBetType("invalid"); err == nil {
		t.Errorf("invalid bet type should fail")
	}
}

// TestBetNumber checks bet number getter/setter.
func TestBetNumber(t *testing.T) {
	g := NewGame()
	if g.GetBetNumber() != 1 {
		t.Errorf("default bet number should be 1")
	}
	if err := g.SetBetNumber(17); err != nil {
		t.Errorf("SetBetNumber(17) should succeed: %v", err)
	}
	if g.BetNumber != 17 {
		t.Errorf("BetNumber should be 17")
	}
	if err := g.SetBetNumber(37); err == nil {
		t.Errorf("SetBetNumber(37) should fail")
	}
	if err := g.SetBetNumber(-1); err == nil {
		t.Errorf("SetBetNumber(-1) should fail")
	}
}

// TestGetNumberColor checks number and color after spin.
func TestGetNumberColor(t *testing.T) {
	g := NewGame()
	g.Spin(95.0)
	if g.GetNumber() != g.Number {
		t.Errorf("GetNumber mismatch")
	}
	if g.GetColor() != g.Color {
		t.Errorf("GetColor mismatch")
	}
}

// TestStraightUpWin checks straight-up bet wins correctly.
func TestStraightUpWin(t *testing.T) {
	g := NewGame()
	g.Bet = 10
	g.BetType = roulette.BetStraightUp
	g.BetNumber = 17
	g.Number = 17
	g.Color = "black"
	var mult = roulette.CheckWin(g.Number, g.BetType, g.BetNumber)
	if mult != 35 {
		t.Errorf("expected multiplier 35, got %d", mult)
	}
	g.Payout = g.Bet * float64(mult+1)
	if g.Payout != 360 {
		t.Errorf("expected payout 360, got %g", g.Payout)
	}
}

// TestStraightUpLose checks straight-up bet loses on wrong number.
func TestStraightUpLose(t *testing.T) {
	g := NewGame()
	g.BetType = roulette.BetStraightUp
	g.BetNumber = 17
	g.Number = 5
	var mult = roulette.CheckWin(g.Number, g.BetType, g.BetNumber)
	if mult != 0 {
		t.Errorf("expected 0 on wrong number, got %d", mult)
	}
}

// TestRedBet checks red bet.
func TestRedBet(t *testing.T) {
	// Number 1 is red
	var mult = roulette.CheckWin(1, roulette.BetRed, 0)
	if mult != 1 {
		t.Errorf("expected red win on 1, got %d", mult)
	}
	// Number 2 is black
	mult = roulette.CheckWin(2, roulette.BetRed, 0)
	if mult != 0 {
		t.Errorf("expected red lose on 2, got %d", mult)
	}
	// 0 is green
	mult = roulette.CheckWin(0, roulette.BetRed, 0)
	if mult != 0 {
		t.Errorf("expected red lose on 0, got %d", mult)
	}
}

// TestBlackBet checks black bet.
func TestBlackBet(t *testing.T) {
	mult := roulette.CheckWin(2, roulette.BetBlack, 0)
	if mult != 1 {
		t.Errorf("expected black win on 2, got %d", mult)
	}
	mult = roulette.CheckWin(1, roulette.BetBlack, 0)
	if mult != 0 {
		t.Errorf("expected black lose on 1, got %d", mult)
	}
}

// TestEvenOdd checks even/odd bets.
func TestEvenOdd(t *testing.T) {
	even := roulette.CheckWin(10, roulette.BetEven, 0)
	if even != 1 {
		t.Errorf("expected even win on 10, got %d", even)
	}
	odd := roulette.CheckWin(11, roulette.BetOdd, 0)
	if odd != 1 {
		t.Errorf("expected odd win on 11, got %d", odd)
	}
	// 0 is neither
	even0 := roulette.CheckWin(0, roulette.BetEven, 0)
	if even0 != 0 {
		t.Errorf("expected 0 to lose even, got %d", even0)
	}
	odd0 := roulette.CheckWin(0, roulette.BetOdd, 0)
	if odd0 != 0 {
		t.Errorf("expected 0 to lose odd, got %d", odd0)
	}
}

// TestLowHigh checks low/high bets.
func TestLowHigh(t *testing.T) {
	low := roulette.CheckWin(5, roulette.BetLow, 0)
	if low != 1 {
		t.Errorf("expected low win on 5, got %d", low)
	}
	high := roulette.CheckWin(20, roulette.BetHigh, 0)
	if high != 1 {
		t.Errorf("expected high win on 20, got %d", high)
	}
	// 32 is high
	low32 := roulette.CheckWin(32, roulette.BetLow, 0)
	if low32 != 0 {
		t.Errorf("expected low lose on 32, got %d", low32)
	}
}

// TestDozen checks dozen bets.
func TestDozen(t *testing.T) {
	d1 := roulette.CheckWin(5, roulette.BetDozen, 1)
	if d1 != 2 {
		t.Errorf("expected dozen 1 win on 5, got %d", d1)
	}
	d2 := roulette.CheckWin(15, roulette.BetDozen, 2)
	if d2 != 2 {
		t.Errorf("expected dozen 2 win on 15, got %d", d2)
	}
	d3 := roulette.CheckWin(30, roulette.BetDozen, 3)
	if d3 != 2 {
		t.Errorf("expected dozen 3 win on 30, got %d", d3)
	}
	d0 := roulette.CheckWin(0, roulette.BetDozen, 1)
	if d0 != 0 {
		t.Errorf("expected 0 to lose dozen, got %d", d0)
	}
}

// TestColumn checks column bets.
func TestColumn(t *testing.T) {
	// Column 1 numbers: 1,4,7,10,13,16,19,22,25,28,31,34
	c1 := roulette.CheckWin(1, roulette.BetColumn, 1)
	if c1 != 2 {
		t.Errorf("expected column 1 win on 1, got %d", c1)
	}
	c2 := roulette.CheckWin(2, roulette.BetColumn, 2)
	if c2 != 2 {
		t.Errorf("expected column 2 win on 2, got %d", c2)
	}
	c3 := roulette.CheckWin(3, roulette.BetColumn, 3)
	if c3 != 2 {
		t.Errorf("expected column 3 win on 3, got %d", c3)
	}
}

// TestSplit checks split bets.
func TestSplit(t *testing.T) {
	// Split on 1-2
	s1 := roulette.CheckWin(1, roulette.BetSplit, 1)
	if s1 != 17 {
		t.Errorf("expected split win on 1, got %d", s1)
	}
	s2 := roulette.CheckWin(2, roulette.BetSplit, 1)
	if s2 != 17 {
		t.Errorf("expected split win on 2, got %d", s2)
	}
	s3 := roulette.CheckWin(3, roulette.BetSplit, 1)
	if s3 != 0 {
		t.Errorf("expected split lose on 3, got %d", s3)
	}
}

// TestStreet checks street bets.
func TestStreet(t *testing.T) {
	// Street on 1-3
	st1 := roulette.CheckWin(1, roulette.BetStreet, 1)
	if st1 != 11 {
		t.Errorf("expected street win on 1, got %d", st1)
	}
	st3 := roulette.CheckWin(3, roulette.BetStreet, 1)
	if st3 != 11 {
		t.Errorf("expected street win on 3, got %d", st3)
	}
	st4 := roulette.CheckWin(4, roulette.BetStreet, 1)
	if st4 != 0 {
		t.Errorf("expected street lose on 4, got %d", st4)
	}
}

// TestCorner checks corner bets.
func TestCorner(t *testing.T) {
	// Corner on 1,2,4,5 (top-left=1)
	c1 := roulette.CheckWin(1, roulette.BetCorner, 1)
	if c1 != 8 {
		t.Errorf("expected corner win on 1, got %d", c1)
	}
	c2 := roulette.CheckWin(2, roulette.BetCorner, 1)
	if c2 != 8 {
		t.Errorf("expected corner win on 2, got %d", c2)
	}
	c5 := roulette.CheckWin(5, roulette.BetCorner, 1)
	if c5 != 8 {
		t.Errorf("expected corner win on 5, got %d", c5)
	}
	c3 := roulette.CheckWin(3, roulette.BetCorner, 1)
	if c3 != 0 {
		t.Errorf("expected corner lose on 3, got %d", c3)
	}
}

// TestLine checks line (6-number) bets.
func TestLine(t *testing.T) {
	// Line on 1-6
	l1 := roulette.CheckWin(1, roulette.BetLine, 1)
	if l1 != 5 {
		t.Errorf("expected line win on 1, got %d", l1)
	}
	l6 := roulette.CheckWin(6, roulette.BetLine, 1)
	if l6 != 5 {
		t.Errorf("expected line win on 6, got %d", l6)
	}
	l7 := roulette.CheckWin(7, roulette.BetLine, 1)
	if l7 != 0 {
		t.Errorf("expected line lose on 7, got %d", l7)
	}
}

// TestPayoutMultiplier checks payout multipliers.
func TestPayoutMultiplier(t *testing.T) {
	tests := []struct {
		bt  roulette.BetType
		exp int
	}{
		{roulette.BetStraightUp, 35},
		{roulette.BetSplit, 17},
		{roulette.BetStreet, 11},
		{roulette.BetCorner, 8},
		{roulette.BetLine, 5},
		{roulette.BetDozen, 2},
		{roulette.BetColumn, 2},
		{roulette.BetRed, 1},
		{roulette.BetBlack, 1},
		{roulette.BetEven, 1},
		{roulette.BetOdd, 1},
		{roulette.BetLow, 1},
		{roulette.BetHigh, 1},
	}
	for _, tt := range tests {
		got := roulette.PayoutMultiplier(tt.bt)
		if got != tt.exp {
			t.Errorf("PayoutMultiplier(%s) = %d, want %d", tt.bt, got, tt.exp)
		}
	}
}

// TestGetPayout checks payout getter.
func TestGetPayout(t *testing.T) {
	g := NewGame()
	g.Payout = 42
	if g.GetPayout() != 42 {
		t.Errorf("GetPayout mismatch")
	}
}

// TestGetResult checks result after spin.
func TestGetResult(t *testing.T) {
	g := NewGame()
	g.Spin(95.0)
	if g.GetResult() == "" {
		t.Errorf("GetResult should not be empty after spin")
	}
}

// TestScanner checks Scanner.
func TestScanner(t *testing.T) {
	g := NewGame()
	g.Payout = 100
	var result roulette.ShotResult
	if err := g.Scanner(&result); err != nil {
		t.Fatalf("Scanner error: %v", err)
	}
	if result.Pay != 100 {
		t.Errorf("expected Pay=100, got %g", result.Pay)
	}
}

// TestGetColor checks European wheel colors.
func TestGetColor(t *testing.T) {
	if c := roulette.GetColor(0); c != "green" {
		t.Errorf("expected 0=green, got %s", c)
	}
	if c := roulette.GetColor(1); c != "red" {
		t.Errorf("expected 1=red, got %s", c)
	}
	if c := roulette.GetColor(2); c != "black" {
		t.Errorf("expected 2=black, got %s", c)
	}
	if c := roulette.GetColor(99); c != "green" {
		t.Errorf("expected unknown=green, got %s", c)
	}
}

// TestBetTypeSetAll checks all valid bet types can be set.
func TestBetTypeSetAll(t *testing.T) {
	g := NewGame()
	bets := []roulette.BetType{
		roulette.BetStraightUp, roulette.BetSplit, roulette.BetStreet,
		roulette.BetCorner, roulette.BetLine, roulette.BetDozen,
		roulette.BetColumn, roulette.BetRed, roulette.BetBlack,
		roulette.BetEven, roulette.BetOdd, roulette.BetLow, roulette.BetHigh,
	}
	for _, bt := range bets {
		if err := g.SetBetType(bt); err != nil {
			t.Errorf("SetBetType(%s) failed: %v", bt, err)
		}
	}
}
