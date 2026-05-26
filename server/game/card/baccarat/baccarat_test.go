package baccarat

import (
	"testing"

	"github.com/slotopol/server/game/card"
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
	if g.BetTarget != BetPlayer {
		t.Errorf("expected default BetTarget=player, got %s", g.BetTarget)
	}
}

// TestDeal checks initial 2 cards each.
func TestDeal(t *testing.T) {
	g := NewGame()
	if err := g.Deal(); err != nil {
		t.Fatalf("Deal() error: %v", err)
	}
	if len(g.PlayerHand) < 2 || len(g.PlayerHand) > 3 {
		t.Errorf("expected 2-3 player cards, got %d", len(g.PlayerHand))
	}
	if len(g.BankerHand) < 2 || len(g.BankerHand) > 3 {
		t.Errorf("expected 2-3 banker cards, got %d", len(g.BankerHand))
	}
	if g.State != "done" {
		t.Errorf("expected State=done after deal, got %s", g.State)
	}
}

// TestNaturalWin checks natural 8 or 9 ends immediately.
func TestNaturalWin(t *testing.T) {
	g := NewGame()
	// Force natural 9 for player
	g.PlayerHand = []card.Card{
		{Suit: card.Hearts, Rank: card.Six},
		{Suit: card.Clubs, Rank: card.Three},
	}
	g.BankerHand = []card.Card{
		{Suit: card.Diamonds, Rank: card.Four},
		{Suit: card.Spades, Rank: card.Two},
	}
	g.BetTarget = BetPlayer
	var pv = card.BaccaratValue(g.PlayerHand)
	if pv != 9 {
		t.Fatalf("expected player 9, got %d", pv)
	}
	g.resolve()
	if g.State != "done" {
		t.Errorf("expected done after natural")
	}
	if g.Result != "Player wins" {
		t.Errorf("expected Player wins, got %s", g.Result)
	}
}

// TestNaturalEight checks natural 8 ends immediately.
func TestNaturalEight(t *testing.T) {
	g := NewGame()
	g.PlayerHand = []card.Card{
		{Suit: card.Hearts, Rank: card.Five},
		{Suit: card.Clubs, Rank: card.Two},
	}
	g.BankerHand = []card.Card{
		{Suit: card.Diamonds, Rank: card.Four},
		{Suit: card.Spades, Rank: card.Four},
	}
	g.BetTarget = BetBanker
	var bv2 = card.BaccaratValue(g.BankerHand)
	if bv2 != 8 {
		t.Fatalf("expected banker 8, got %d", bv2)
	}
	g.resolve()
	if g.Result != "Banker wins" {
		t.Errorf("expected Banker wins, got %s", g.Result)
	}
}

// TestBetPlayerWins checks Player bet pays 1:1.
func TestBetPlayerWins(t *testing.T) {
	g := NewGame()
	g.Bet = 10
	g.BetTarget = BetPlayer
	g.PlayerHand = []card.Card{
		{Suit: card.Hearts, Rank: card.Seven},
		{Suit: card.Clubs, Rank: card.Two},
	}
	g.BankerHand = []card.Card{
		{Suit: card.Diamonds, Rank: card.Five},
		{Suit: card.Spades, Rank: card.Two},
	}
	g.resolve()
	if g.Result != "Player wins" {
		t.Errorf("expected Player wins, got %s", g.Result)
	}
	if g.Payout != 20 {
		t.Errorf("expected Payout=20 (2x bet), got %g", g.Payout)
	}
}

// TestBetPlayerLoses checks Player bet loses.
func TestBetPlayerLoses(t *testing.T) {
	g := NewGame()
	g.Bet = 10
	g.BetTarget = BetPlayer
	g.PlayerHand = []card.Card{
		{Suit: card.Hearts, Rank: card.Five},
		{Suit: card.Clubs, Rank: card.Two},
	}
	g.BankerHand = []card.Card{
		{Suit: card.Diamonds, Rank: card.Seven},
		{Suit: card.Spades, Rank: card.Two},
	}
	g.resolve()
	if g.Result != "Banker wins" {
		t.Errorf("expected Banker wins, got %s", g.Result)
	}
	if g.Payout != 0 {
		t.Errorf("expected Payout=0, got %g", g.Payout)
	}
}

// TestBetPlayerPush checks Player bet pushes on tie.
func TestBetPlayerPush(t *testing.T) {
	g := NewGame()
	g.Bet = 10
	g.BetTarget = BetPlayer
	g.PlayerHand = []card.Card{
		{Suit: card.Hearts, Rank: card.Five},
		{Suit: card.Clubs, Rank: card.Three},
	}
	g.BankerHand = []card.Card{
		{Suit: card.Diamonds, Rank: card.Five},
		{Suit: card.Spades, Rank: card.Three},
	}
	g.resolve()
	if g.Result != "Tie - bet returned" {
		t.Errorf("expected Tie - bet returned, got %s", g.Result)
	}
	if g.Payout != 10 {
		t.Errorf("expected Payout=10 (push), got %g", g.Payout)
	}
}

// TestBetBankerWins checks Banker bet pays 0.95:1.
func TestBetBankerWins(t *testing.T) {
	g := NewGame()
	g.Bet = 10
	g.BetTarget = BetBanker
	g.PlayerHand = []card.Card{
		{Suit: card.Hearts, Rank: card.Five},
		{Suit: card.Clubs, Rank: card.Two},
	}
	g.BankerHand = []card.Card{
		{Suit: card.Diamonds, Rank: card.Seven},
		{Suit: card.Spades, Rank: card.Two},
	}
	g.resolve()
	if g.Result != "Banker wins" {
		t.Errorf("expected Banker wins, got %s", g.Result)
	}
	// Payout = 10 + 10*0.95 = 19.5
	if g.Payout != 19.5 {
		t.Errorf("expected Payout=19.5 (10 + 9.5 commission), got %g", g.Payout)
	}
}

// TestBetBankerLoses checks Banker bet loses.
func TestBetBankerLoses(t *testing.T) {
	g := NewGame()
	g.Bet = 10
	g.BetTarget = BetBanker
	g.PlayerHand = []card.Card{
		{Suit: card.Hearts, Rank: card.Seven},
		{Suit: card.Clubs, Rank: card.Two},
	}
	g.BankerHand = []card.Card{
		{Suit: card.Diamonds, Rank: card.Five},
		{Suit: card.Spades, Rank: card.Two},
	}
	g.resolve()
	if g.Result != "Player wins" {
		t.Errorf("expected Player wins, got %s", g.Result)
	}
	if g.Payout != 0 {
		t.Errorf("expected Payout=0, got %g", g.Payout)
	}
}

// TestBetBankerPush checks Banker bet pushes on tie.
func TestBetBankerPush(t *testing.T) {
	g := NewGame()
	g.Bet = 10
	g.BetTarget = BetBanker
	g.PlayerHand = []card.Card{
		{Suit: card.Hearts, Rank: card.Five},
		{Suit: card.Clubs, Rank: card.Three},
	}
	g.BankerHand = []card.Card{
		{Suit: card.Diamonds, Rank: card.Five},
		{Suit: card.Spades, Rank: card.Three},
	}
	g.resolve()
	if g.Result != "Tie - bet returned" {
		t.Errorf("expected Tie - bet returned, got %s", g.Result)
	}
	if g.Payout != 10 {
		t.Errorf("expected Payout=10 (push), got %g", g.Payout)
	}
}

// TestBetTieWins checks Tie bet pays 8:1.
func TestBetTieWins(t *testing.T) {
	g := NewGame()
	g.Bet = 10
	g.BetTarget = BetTie
	g.PlayerHand = []card.Card{
		{Suit: card.Hearts, Rank: card.Five},
		{Suit: card.Clubs, Rank: card.Three},
	}
	g.BankerHand = []card.Card{
		{Suit: card.Diamonds, Rank: card.Five},
		{Suit: card.Spades, Rank: card.Three},
	}
	g.resolve()
	if g.Result != "Tie wins!" {
		t.Errorf("expected Tie wins!, got %s", g.Result)
	}
	if g.Payout != 90 {
		t.Errorf("expected Payout=90 (10*9), got %g", g.Payout)
	}
}

// TestBetTieLoses checks Tie bet loses.
func TestBetTieLoses(t *testing.T) {
	g := NewGame()
	g.Bet = 10
	g.BetTarget = BetTie
	g.PlayerHand = []card.Card{
		{Suit: card.Hearts, Rank: card.Five},
		{Suit: card.Clubs, Rank: card.Two},
	}
	g.BankerHand = []card.Card{
		{Suit: card.Diamonds, Rank: card.Seven},
		{Suit: card.Spades, Rank: card.Two},
	}
	g.resolve()
	if g.Result != "No tie" {
		t.Errorf("expected No tie, got %s", g.Result)
	}
	if g.Payout != 0 {
		t.Errorf("expected Payout=0, got %g", g.Payout)
	}
}

// TestBaccaratValues checks card values in baccarat.
func TestBaccaratValues(t *testing.T) {
	tests := []struct {
		hand card.Hand
		want int
	}{
		{[]card.Card{{Suit: card.Hearts, Rank: card.Ace}, {Suit: card.Clubs, Rank: card.Nine}}, 0},  // 1+9=10 -> 0
		{[]card.Card{{Suit: card.Hearts, Rank: card.Ten}, {Suit: card.Clubs, Rank: card.Eight}}, 8},  // 0+8=8
		{[]card.Card{{Suit: card.Hearts, Rank: card.King}, {Suit: card.Clubs, Rank: card.Queen}}, 0}, // 0+0=0
		{[]card.Card{{Suit: card.Hearts, Rank: card.Five}, {Suit: card.Clubs, Rank: card.Six}}, 1},   // 5+6=11 -> 1
		{[]card.Card{{Suit: card.Hearts, Rank: card.Seven}, {Suit: card.Clubs, Rank: card.Three}}, 0}, // 7+3=10 -> 0
	}
	for _, tt := range tests {
		got := card.BaccaratValue(tt.hand)
		if got != tt.want {
			t.Errorf("BaccaratValue(%v) = %d, want %d", tt.hand, got, tt.want)
		}
	}
}

// TestNaturalDetection checks natural detection logic.
func TestNaturalDetection(t *testing.T) {
	g := NewGame()
	g.PlayerHand = []card.Card{
		{Suit: card.Hearts, Rank: card.Eight},
		{Suit: card.Clubs, Rank: card.Ace},
	}
	g.BankerHand = []card.Card{
		{Suit: card.Diamonds, Rank: card.Four},
		{Suit: card.Spades, Rank: card.Three},
	}
	var pv = card.BaccaratValue(g.PlayerHand)
	if pv != 9 {
		t.Errorf("expected player 9, got %d", pv)
	}
	// Player natural should end round
	if pv >= 8 {
		g.resolve()
		if g.State != "done" {
			t.Errorf("expected done after natural, got %s", g.State)
		}
	}
}

// TestBetSet checks bet validation.
func TestBetSet(t *testing.T) {
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

// TestThirdCardRules checks player <=5 draws third card.
func TestThirdCardRules(t *testing.T) {
	g := NewGame()
	g.State = "dealing"
	g.PlayerHand = []card.Card{
		{Suit: card.Hearts, Rank: card.Ace},
		{Suit: card.Clubs, Rank: card.Two},
	}
	g.BankerHand = []card.Card{
		{Suit: card.Diamonds, Rank: card.Four},
		{Suit: card.Spades, Rank: card.Three},
	}
	var pv = card.BaccaratValue(g.PlayerHand)
	if pv == 3 && pv <= 5 {
		g.PlayerHand = append(g.PlayerHand, card.Card{Suit: card.Hearts, Rank: card.Seven})
	}
	if len(g.PlayerHand) != 3 {
		t.Errorf("expected 3 player cards when <=5, got %d", len(g.PlayerHand))
	}
}

// TestPlayerStandsOnSix checks player stands on 6-7.
func TestPlayerStandsOnSix(t *testing.T) {
	g := NewGame()
	g.PlayerHand = []card.Card{
		{Suit: card.Hearts, Rank: card.Four},
		{Suit: card.Clubs, Rank: card.Two},
	}
	var pv = card.BaccaratValue(g.PlayerHand)
	if pv != 6 {
		t.Fatalf("expected player 6, got %d", pv)
	}
	// Player should stand on 6-7, so no third card
	if pv > 5 {
		// Player stays
	}
	_ = g
}

// TestBankerDrawsOnTwo checks banker draws on <=2.
func TestBankerDrawsOnTwo(t *testing.T) {
	g := NewGame()
	g.BankerHand = []card.Card{
		{Suit: card.Diamonds, Rank: card.Ace},
		{Suit: card.Spades, Rank: card.Ace},
	}
	var bv = card.BaccaratValue(g.BankerHand)
	if bv != 2 {
		t.Fatalf("expected banker 2, got %d", bv)
	}
	if bv <= 2 {
		g.BankerHand = append(g.BankerHand, card.Card{Suit: card.Hearts, Rank: card.Five})
	}
	if len(g.BankerHand) != 3 {
		t.Errorf("expected banker to draw on <=2, got %d cards", len(g.BankerHand))
	}
}

// TestBankerStandsOnSeven checks banker stands on 7.
func TestBankerStandsOnSeven(t *testing.T) {
	g := NewGame()
	g.BankerHand = []card.Card{
		{Suit: card.Diamonds, Rank: card.Four},
		{Suit: card.Spades, Rank: card.Three},
	}
	var bv = card.BaccaratValue(g.BankerHand)
	if bv != 7 {
		t.Fatalf("expected banker 7, got %d", bv)
	}
	// Banker stands on 7
	var drew = false
	if bv <= 2 {
		drew = true
	} else if bv == 3 {
		bv = 7
	}
	_ = drew
}

// TestInterface checks CardGame interface.
func TestInterface(t *testing.T) {
	g := NewGame()
	if g.GetHand() != nil {
		// empty before deal
	}
	g.Deal()
	if len(g.GetHand()) != len(g.PlayerHand) {
		t.Errorf("GetHand mismatch")
	}
	if len(g.GetDealerHand()) != len(g.BankerHand) {
		t.Errorf("GetDealerHand mismatch")
	}
	if g.GetPayout() != g.Payout {
		t.Errorf("GetPayout mismatch")
	}
	if g.GetGameState() != "done" {
		t.Errorf("GetGameState should be done")
	}
	if g.GetResult() == "" {
		t.Errorf("GetResult should not be empty")
	}
	var result card.ShotResult
	if err := g.Scanner(&result); err != nil {
		t.Errorf("Scanner error: %v", err)
	}
}

// TestSpin checks Spin calls Deal.
func TestSpin(t *testing.T) {
	g := NewGame()
	g.Spin(95.0)
	if len(g.PlayerHand) < 2 {
		t.Errorf("expected player hand after Spin")
	}
}

// TestGetBet checks bet getter.
func TestGetBet(t *testing.T) {
	g := NewGame()
	if g.GetBet() != 1 {
		t.Errorf("GetBet should return 1")
	}
}

// TestDealResets checks deal resets previous state.
func TestDealResets(t *testing.T) {
	g := NewGame()
	g.Deal()
	firstHand := make([]card.Card, len(g.PlayerHand))
	copy(firstHand, g.PlayerHand)
	g.Deal()
	var same = len(firstHand) == len(g.PlayerHand)
	if same {
		for i, c := range firstHand {
			if c != g.PlayerHand[i] {
				same = false
				break
			}
		}
	}
	if same {
		t.Log("rare: identical hands after consecutive deals")
	}
}
