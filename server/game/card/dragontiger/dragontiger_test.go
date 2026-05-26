package dragontiger

import (
	"testing"

	"github.com/slotopol/server/game/card"
)

// Helper to create a simple hand with a card of given rank
func makeHand(rankVal int) card.Hand {
	return []card.Card{{Rank: card.Rank(rankVal), Suit: card.Clubs}}
}

// TestNewGame checks that a new Dragon Tiger game has correct defaults.
func TestNewGame(t *testing.T) {
	g := NewGame()
	if g.Bet != 1 {
		t.Errorf("expected default Bet=1, got %g", g.Bet)
	}
	if g.State != "waiting" {
		t.Errorf("expected default State=waiting, got %s", g.State)
	}
	if g.BetTarget != BetDragon {
		t.Errorf("expected default BetTarget=dragon, got %s", g.BetTarget)
	}
	if len(g.Deck) != 0 {
		t.Errorf("expected empty deck initially, got %d cards", len(g.Deck))
	}
}

// TestDeal checks that dealing gives one card each to Dragon and Tiger.
func TestDeal(t *testing.T) {
	g := NewGame()
	if err := g.Deal(); err != nil {
		t.Fatalf("Deal() returned error: %v", err)
	}
	if len(g.DragonHand) != 1 {
		t.Errorf("expected 1 dragon card, got %d", len(g.DragonHand))
	}
	if len(g.TigerHand) != 1 {
		t.Errorf("expected 1 tiger card, got %d", len(g.TigerHand))
	}
	if g.State != "done" {
		t.Errorf("expected State=done after deal, got %s", g.State)
	}
	if g.DragonValue == 0 || g.TigerValue == 0 {
		t.Errorf("DragonValue=%d and TigerValue=%d should be non-zero", g.DragonValue, g.TigerValue)
	}
}

// TestDealResetsHands checks that a new Deal clears previous hands.
func TestDealResetsHands(t *testing.T) {
	g := NewGame()
	g.Deal()
	firstDragon := g.DragonHand[0]
	firstTiger := g.TigerHand[0]

	g.Deal()
	if len(g.DragonHand) != 1 || len(g.TigerHand) != 1 {
		t.Fatal("deal should produce one card each")
	}
	// Check hands changed (very unlikely same card twice from shuffled deck)
	if g.DragonHand[0] == firstDragon && g.TigerHand[0] == firstTiger {
		t.Log("rare: same cards after new deal")
	}
}

// TestDeckShuffle checks that Deck is populated when empty.
func TestDeckShuffle(t *testing.T) {
	g := NewGame()
	g.Deck = nil
	g.Deal()
	if len(g.Deck) == 0 {
		t.Error("Deck should be populated after dealing with empty deck")
	}
}

// TestBetDragonWins checks that betting on Dragon pays correctly when Dragon wins.
func TestBetDragonWins(t *testing.T) {
	g := NewGame()
	g.Bet = 10
	g.BetTarget = BetDragon
	g.DragonValue = 10
	g.TigerValue = 5
	g.DragonHand = makeHand(10)
	g.TigerHand = makeHand(5)
	g.resolve()
	if g.Result != "Dragon wins" {
		t.Errorf("expected 'Dragon wins', got '%s'", g.Result)
	}
	if g.Payout != 20 {
		t.Errorf("expected payout 20 (2x bet), got %g", g.Payout)
	}
}

// TestBetDragonLoses checks that betting on Dragon loses when Tiger wins.
func TestBetDragonLoses(t *testing.T) {
	g := NewGame()
	g.Bet = 10
	g.BetTarget = BetDragon
	g.DragonValue = 5
	g.TigerValue = 10
	g.DragonHand = makeHand(5)
	g.TigerHand = makeHand(10)
	g.resolve()
	if g.Result != "Tiger wins" {
		t.Errorf("expected 'Tiger wins', got '%s'", g.Result)
	}
	if g.Payout != 0 {
		t.Errorf("expected payout 0, got %g", g.Payout)
	}
}

// TestBetDragonTie checks that Dragon bet loses half on tie.
func TestBetDragonTie(t *testing.T) {
	g := NewGame()
	g.Bet = 10
	g.BetTarget = BetDragon
	g.DragonValue = 7
	g.TigerValue = 7
	g.DragonHand = makeHand(7)
	g.TigerHand = makeHand(7)
	g.resolve()
	if g.Payout != 5 {
		t.Errorf("expected payout 5 (half) on tie, got %g", g.Payout)
	}
	if g.Result != "Tie - half lost" {
		t.Errorf("expected 'Tie - half lost', got '%s'", g.Result)
	}
}

// TestBetTigerWins checks that betting on Tiger pays correctly when Tiger wins.
func TestBetTigerWins(t *testing.T) {
	g := NewGame()
	g.Bet = 10
	g.BetTarget = BetTiger
	g.DragonValue = 4
	g.TigerValue = 9
	g.DragonHand = makeHand(4)
	g.TigerHand = makeHand(9)
	g.resolve()
	if g.Result != "Tiger wins" {
		t.Errorf("expected 'Tiger wins', got '%s'", g.Result)
	}
	if g.Payout != 20 {
		t.Errorf("expected payout 20, got %g", g.Payout)
	}
}

// TestBetTigerLoses checks that Tiger bet loses when Dragon wins.
func TestBetTigerLoses(t *testing.T) {
	g := NewGame()
	g.Bet = 10
	g.BetTarget = BetTiger
	g.DragonValue = 13
	g.TigerValue = 2
	g.DragonHand = makeHand(13)
	g.TigerHand = makeHand(2)
	g.resolve()
	if g.Result != "Dragon wins" {
		t.Errorf("expected 'Dragon wins', got '%s'", g.Result)
	}
	if g.Payout != 0 {
		t.Errorf("expected payout 0, got %g", g.Payout)
	}
}

// TestBetTieWins checks that Tie bet pays 8:1 on tie.
func TestBetTieWins(t *testing.T) {
	g := NewGame()
	g.Bet = 10
	g.BetTarget = BetDTie
	g.DragonValue = 8
	g.TigerValue = 8
	g.DragonHand = makeHand(8)
	g.TigerHand = makeHand(8)
	g.resolve()
	if g.Result != "Tie wins!" {
		t.Errorf("expected 'Tie wins!', got '%s'", g.Result)
	}
	if g.Payout != 90 {
		t.Errorf("expected payout 90 (9x), got %g", g.Payout)
	}
}

// TestBetTieLoses checks that Tie bet loses when no tie.
func TestBetTieLoses(t *testing.T) {
	g := NewGame()
	g.Bet = 10
	g.BetTarget = BetDTie
	g.DragonValue = 8
	g.TigerValue = 5
	g.DragonHand = makeHand(8)
	g.TigerHand = makeHand(5)
	g.resolve()
	if g.Payout != 0 {
		t.Errorf("expected payout 0, got %g", g.Payout)
	}
}

// TestBetSuitedTie checks suited tie pays 50:1.
func TestBetSuitedTie(t *testing.T) {
	g := NewGame()
	g.Bet = 10
	g.BetTarget = BetDSuitedTie
	g.DragonValue = 6
	g.TigerValue = 6
	g.DragonHand = []card.Card{{Suit: card.Hearts, Rank: 6}}
	g.TigerHand = []card.Card{{Suit: card.Hearts, Rank: 6}}
	g.resolve()
	if g.Result != "Suited Tie wins!" {
		t.Errorf("expected 'Suited Tie wins!', got '%s'", g.Result)
	}
	if g.Payout != 510 {
		t.Errorf("expected payout 510 (51x), got %g", g.Payout)
	}
}

// TestBetSuitedTieNoMatch checks suited tie loses when suits differ.
func TestBetSuitedTieNoMatch(t *testing.T) {
	g := NewGame()
	g.Bet = 10
	g.BetTarget = BetDSuitedTie
	g.DragonValue = 6
	g.TigerValue = 6
	g.DragonHand = []card.Card{{Suit: card.Hearts, Rank: 6}}
	g.TigerHand = []card.Card{{Suit: card.Clubs, Rank: 6}}
	g.resolve()
	if g.Payout != 0 {
		t.Errorf("expected payout 0 for non-suited tie, got %g", g.Payout)
	}
}

// TestBigBetWin checks Big bet wins on high cards.
func TestBigBetWin(t *testing.T) {
	g := NewGame()
	g.Bet = 10
	g.BetTarget = BetDBig
	g.DragonValue = 9
	g.TigerValue = 3
	g.DragonHand = makeHand(9)
	g.TigerHand = makeHand(3)
	g.resolve()
	// Dragon has high (9 >= 8) -> Big wins
	if g.Payout != 20 {
		t.Errorf("expected payout 20, got %g", g.Payout)
	}
}

// TestBigBetTie checks Big bet loses half on tie.
func TestBigBetTie(t *testing.T) {
	g := NewGame()
	g.Bet = 10
	g.BetTarget = BetDBig
	g.DragonValue = 7
	g.TigerValue = 7
	g.DragonHand = makeHand(7)
	g.TigerHand = makeHand(7)
	g.resolve()
	if g.Payout != 5 {
		t.Errorf("expected payout 5 (half) on tie, got %v", g.Payout)
	}
}

// TestSmallBet checks Small bet wins on low cards.
func TestSmallBet(t *testing.T) {
	g := NewGame()
	g.Bet = 10
	g.BetTarget = BetDSmall
	g.DragonValue = 3
	g.TigerValue = 10
	g.DragonHand = makeHand(3)
	g.TigerHand = makeHand(10)
	g.resolve()
	if g.Payout != 20 {
		t.Errorf("expected payout 20, got %g", g.Payout)
	}
}

// TestScanner checks Scanner returns payout.
func TestScanner(t *testing.T) {
	g := NewGame()
	g.Payout = 42
	var result card.ShotResult
	if err := g.Scanner(&result); err != nil {
		t.Fatalf("Scanner error: %v", err)
	}
	if result.Pay != 42 {
		t.Errorf("expected Pay=42, got %g", result.Pay)
	}
}

// TestSpin checks Spin calls Deal.
func TestSpin(t *testing.T) {
	g := NewGame()
	g.Spin(95.0)
	if g.State != "done" {
		t.Errorf("expected State=done after Spin, got %s", g.State)
	}
	if len(g.DragonHand) != 1 || len(g.TigerHand) != 1 {
		t.Errorf("expected hands dealt after Spin")
	}
}

// TestGetHandDealerHand checks interface methods.
func TestGetHandDealerHand(t *testing.T) {
	g := NewGame()
	g.Deal()
	if len(g.GetHand()) != 1 {
		t.Errorf("GetHand should return DragonHand")
	}
	if len(g.GetDealerHand()) != 1 {
		t.Errorf("GetDealerHand should return TigerHand")
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
}

// TestSetBet checks bet validation.
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

// TestRankValues checks rank value mapping.
func TestRankValues(t *testing.T) {
	tests := []struct {
		rank card.Rank
		val  int
	}{
		{card.Ace, 14},
		{card.Two, 2},
		{card.Three, 3},
		{card.Four, 4},
		{card.Five, 5},
		{card.Six, 6},
		{card.Seven, 7},
		{card.Eight, 8},
		{card.Nine, 9},
		{card.Ten, 10},
		{card.Jack, 11},
		{card.Queen, 12},
		{card.King, 13},
	}
	for _, tt := range tests {
		got := dragonRankValue(tt.rank)
		if got != tt.val {
			t.Errorf("dragonRankValue(%v) = %d, want %d", tt.rank, got, tt.val)
		}
	}
}

// TestIsHighLow checks card classifications.
func TestIsHighLow(t *testing.T) {
	high := []card.Rank{card.Eight, card.Nine, card.Ten, card.Jack, card.Queen, card.King, card.Ace}
	low := []card.Rank{card.Two, card.Three, card.Four, card.Five, card.Six}
	for _, r := range high {
		if !isHigh(r) {
			t.Errorf("isHigh(%v) should be true", r)
		}
		if r >= card.Seven && isLow(r) {
			t.Errorf("isLow(%v) should be false", r)
		}
	}
	for _, r := range low {
		if !isLow(r) {
			t.Errorf("isLow(%v) should be true", r)
		}
		if isHigh(r) {
			t.Errorf("isHigh(%v) should be false", r)
		}
	}
}
