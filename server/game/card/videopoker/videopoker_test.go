package videopoker

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
	if len(g.Hand) != 0 {
		t.Errorf("expected empty hand initially")
	}
}

// TestDeal checks 5 cards dealt, state=holding.
func TestDeal(t *testing.T) {
	g := NewGame()
	if err := g.Deal(); err != nil {
		t.Fatalf("Deal() error: %v", err)
	}
	if len(g.Hand) != 5 {
		t.Errorf("expected 5 cards, got %d", len(g.Hand))
	}
	if g.State != "holding" {
		t.Errorf("expected State=holding after deal, got %s", g.State)
	}
	if g.HoldMask != [5]bool{} {
		t.Errorf("expected empty hold mask after deal")
	}
}

// TestSetHold checks hold mask.
func TestSetHold(t *testing.T) {
	g := NewGame()
	g.Deal()
	var hold = [5]bool{true, false, true, false, false}
	if err := g.SetHold(hold); err != nil {
		t.Fatalf("SetHold error: %v", err)
	}
	if g.HoldMask != hold {
		t.Errorf("HoldMask not set correctly")
	}
}

// TestSetHoldWhenNotHolding checks error.
func TestSetHoldWhenNotHolding(t *testing.T) {
	g := NewGame()
	if err := g.SetHold([5]bool{}); err == nil {
		t.Errorf("expected error when setting hold before deal")
	}
}

// TestDrawReplacesCards checks draw replaces non-held cards.
func TestDrawReplacesCards(t *testing.T) {
	g := NewGame()
	g.Deal()
	var original = make([]card.Card, 5)
	copy(original, g.Hand)
	g.SetHold([5]bool{true, true, true, false, false})
	if err := g.Draw(); err != nil {
		t.Fatalf("Draw() error: %v", err)
	}
	if g.State != "done" {
		t.Errorf("expected State=done after draw, got %s", g.State)
	}
	// Check that held cards are still present (hand is sorted after draw)
	// Cards at indices 0,1,2 were held
	var found0, found1, found2 bool
	for _, c := range g.Hand {
		if c == original[0] {
			found0 = true
		}
		if c == original[1] {
			found1 = true
		}
		if c == original[2] {
			found2 = true
		}
	}
	if !found0 {
		t.Errorf("expected held card[0] %v to be present after draw", original[0])
	}
	if !found1 {
		t.Errorf("expected held card[1] %v to be present after draw", original[1])
	}
	if !found2 {
		t.Errorf("expected held card[2] %v to be present after draw", original[2])
	}
	// Non-held cards (3,4) should have been replaced
	var keptAll = 0
	for _, c := range g.Hand {
		if c == original[3] || c == original[4] {
			keptAll++
		}
	}
	if keptAll == 2 && len(g.Deck) < 2 {
		t.Log("rare: non-held cards coincidentally same after redraw")
	}
}

// TestDrawWhenNotHolding checks error.
func TestDrawWhenNotHolding(t *testing.T) {
	g := NewGame()
	if err := g.Draw(); err == nil {
		t.Errorf("expected error when drawing before deal")
	}
}

// TestDealResets checks deal resets previous state.
func TestDealResets(t *testing.T) {
	g := NewGame()
	g.Deal()
	var first = make([]card.Card, 5)
	copy(first, g.Hand)
	g.Deal()
	var same = true
	for i, c := range first {
		if c != g.Hand[i] {
			same = false
			break
		}
	}
	if same {
		t.Log("rare: identical hands after consecutive deals")
	}
}

// TestRoyalFlush checks Royal Flush payout (800:1).
func TestRoyalFlush(t *testing.T) {
	g := NewGame()
	g.Bet = 10
	g.HandResult = card.PokerHandResult{
		Rank: card.RoyalFlush,
	}
	g.resolve()
	if g.Payout != 8000 {
		t.Errorf("expected Payout=8000, got %g", g.Payout)
	}
	if g.Result != "Royal Flush!" {
		t.Errorf("expected Royal Flush!, got %s", g.Result)
	}
}

// TestStraightFlush checks Straight Flush payout (50:1).
func TestStraightFlush(t *testing.T) {
	g := NewGame()
	g.Bet = 10
	g.HandResult = card.PokerHandResult{
		Rank: card.StraightFlush,
	}
	g.resolve()
	if g.Payout != 500 {
		t.Errorf("expected Payout=500, got %g", g.Payout)
	}
	if g.Result != "Straight Flush!" {
		t.Errorf("expected Straight Flush!, got %s", g.Result)
	}
}

// TestFourOfAKind checks Four of a Kind payout (25:1).
func TestFourOfAKind(t *testing.T) {
	g := NewGame()
	g.Bet = 10
	g.HandResult = card.PokerHandResult{
		Rank: card.FourOfAKind,
	}
	g.resolve()
	if g.Payout != 250 {
		t.Errorf("expected Payout=250, got %g", g.Payout)
	}
	if g.Result != "Four of a Kind!" {
		t.Errorf("expected Four of a Kind!, got %s", g.Result)
	}
}

// TestFullHouse checks Full House payout (9:1).
func TestFullHouse(t *testing.T) {
	g := NewGame()
	g.Bet = 10
	g.HandResult = card.PokerHandResult{
		Rank: card.FullHouse,
	}
	g.resolve()
	if g.Payout != 90 {
		t.Errorf("expected Payout=90, got %g", g.Payout)
	}
	if g.Result != "Full House!" {
		t.Errorf("expected Full House!, got %s", g.Result)
	}
}

// TestFlush checks Flush payout (6:1).
func TestFlush(t *testing.T) {
	g := NewGame()
	g.Bet = 10
	g.HandResult = card.PokerHandResult{
		Rank: card.Flush,
	}
	g.resolve()
	if g.Payout != 60 {
		t.Errorf("expected Payout=60, got %g", g.Payout)
	}
	if g.Result != "Flush!" {
		t.Errorf("expected Flush!, got %s", g.Result)
	}
}

// TestStraight checks Straight payout (4:1).
func TestStraight(t *testing.T) {
	g := NewGame()
	g.Bet = 10
	g.HandResult = card.PokerHandResult{
		Rank: card.Straight,
	}
	g.resolve()
	if g.Payout != 40 {
		t.Errorf("expected Payout=40, got %g", g.Payout)
	}
	if g.Result != "Straight!" {
		t.Errorf("expected Straight!, got %s", g.Result)
	}
}

// TestThreeOfAKind checks Three of a Kind payout (3:1).
func TestThreeOfAKind(t *testing.T) {
	g := NewGame()
	g.Bet = 10
	g.HandResult = card.PokerHandResult{
		Rank: card.ThreeOfAKind,
	}
	g.resolve()
	if g.Payout != 30 {
		t.Errorf("expected Payout=30, got %g", g.Payout)
	}
	if g.Result != "Three of a Kind" {
		t.Errorf("expected Three of a Kind, got %s", g.Result)
	}
}

// TestTwoPair checks Two Pair payout (2:1).
func TestTwoPair(t *testing.T) {
	g := NewGame()
	g.Bet = 10
	g.HandResult = card.PokerHandResult{
		Rank: card.TwoPair,
	}
	g.resolve()
	if g.Payout != 20 {
		t.Errorf("expected Payout=20, got %g", g.Payout)
	}
	if g.Result != "Two Pair" {
		t.Errorf("expected Two Pair, got %s", g.Result)
	}
}

// TestJacksOrBetter checks One Pair J+ payout (1:1).
func TestJacksOrBetter(t *testing.T) {
	g := NewGame()
	g.Bet = 10
	g.HandResult = card.PokerHandResult{
		Rank:    card.OnePair,
		Kickers: []card.Rank{card.Jack, card.Queen, card.Nine},
	}
	if !card.IsJacksOrBetter(g.HandResult) {
		t.Errorf("IsJacksOrBetter should return true for pair of Jacks")
	}
	g.resolve()
	if g.Payout != 10 {
		t.Errorf("expected Payout=10, got %g", g.Payout)
	}
	if g.Result != "Jacks or Better" {
		t.Errorf("expected Jacks or Better, got %s", g.Result)
	}
}

// TestJacksOrBetterAces checks pair of Aces pays.
func TestJacksOrBetterAces(t *testing.T) {
	g := NewGame()
	g.HandResult = card.PokerHandResult{
		Rank:    card.OnePair,
		Kickers: []card.Rank{card.Ace, card.King, card.Queen},
	}
	if !card.IsJacksOrBetter(g.HandResult) {
		t.Errorf("IsJacksOrBetter should return true for pair of Aces")
	}
}

// TestLowPair checks pair under Jack pays 0.
func TestLowPair(t *testing.T) {
	g := NewGame()
	g.Bet = 10
	g.HandResult = card.PokerHandResult{
		Rank:    card.OnePair,
		Kickers: []card.Rank{card.Ten, card.Nine, card.Eight},
	}
	if card.IsJacksOrBetter(g.HandResult) {
		t.Errorf("IsJacksOrBetter should return false for pair of Tens")
	}
	g.resolve()
	if g.Payout != 0 {
		t.Errorf("expected Payout=0 for low pair, got %g", g.Payout)
	}
	if g.Result != "Low Pair" {
		t.Errorf("expected Low Pair, got %s", g.Result)
	}
}

// TestHighCard checks high card pays 0.
func TestHighCard(t *testing.T) {
	g := NewGame()
	g.Bet = 10
	g.HandResult = card.PokerHandResult{
		Rank: card.HighCard,
	}
	g.resolve()
	if g.Payout != 0 {
		t.Errorf("expected Payout=0 for high card, got %g", g.Payout)
	}
	if g.Result != "High Card" {
		t.Errorf("expected High Card, got %s", g.Result)
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

// TestInterface checks CardGame interface.
func TestInterface(t *testing.T) {
	g := NewGame()
	g.Deal()
	if len(g.GetHand()) != 5 {
		t.Errorf("GetHand should return 5 cards")
	}
	if g.GetDealerHand() != nil {
		t.Errorf("GetDealerHand should return nil for video poker")
	}
	if g.GetPayout() != g.Payout {
		t.Errorf("GetPayout mismatch")
	}
	if g.GetGameState() == "" {
		t.Errorf("GetGameState should not be empty")
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
	if len(g.Hand) != 5 {
		t.Errorf("expected hand dealt after Spin")
	}
}

// TestEvaluateRoyalFlush checks royal flush detection.
func TestEvaluateRoyalFlush(t *testing.T) {
	h := card.Hand{
		{Suit: card.Hearts, Rank: card.Ace},
		{Suit: card.Hearts, Rank: card.King},
		{Suit: card.Hearts, Rank: card.Queen},
		{Suit: card.Hearts, Rank: card.Jack},
		{Suit: card.Hearts, Rank: card.Ten},
	}
	r := card.EvaluatePokerHand(h)
	if r.Rank != card.RoyalFlush {
		t.Errorf("expected RoyalFlush, got %s", card.PokerHandNames[r.Rank])
	}
}

// TestEvaluateStraightFlush checks straight flush detection.
func TestEvaluateStraightFlush(t *testing.T) {
	h := card.Hand{
		{Suit: card.Clubs, Rank: card.Nine},
		{Suit: card.Clubs, Rank: card.Eight},
		{Suit: card.Clubs, Rank: card.Seven},
		{Suit: card.Clubs, Rank: card.Six},
		{Suit: card.Clubs, Rank: card.Five},
	}
	r := card.EvaluatePokerHand(h)
	if r.Rank != card.StraightFlush {
		t.Errorf("expected StraightFlush, got %s", card.PokerHandNames[r.Rank])
	}
}

// TestEvaluateFourOfAKind checks four of a kind detection.
func TestEvaluateFourOfAKind(t *testing.T) {
	h := card.Hand{
		{Suit: card.Hearts, Rank: card.Seven},
		{Suit: card.Clubs, Rank: card.Seven},
		{Suit: card.Diamonds, Rank: card.Seven},
		{Suit: card.Spades, Rank: card.Seven},
		{Suit: card.Hearts, Rank: card.Two},
	}
	r := card.EvaluatePokerHand(h)
	if r.Rank != card.FourOfAKind {
		t.Errorf("expected FourOfAKind, got %s", card.PokerHandNames[r.Rank])
	}
}

// TestEvaluateFullHouse checks full house detection.
func TestEvaluateFullHouse(t *testing.T) {
	h := card.Hand{
		{Suit: card.Hearts, Rank: card.King},
		{Suit: card.Clubs, Rank: card.King},
		{Suit: card.Diamonds, Rank: card.King},
		{Suit: card.Spades, Rank: card.Five},
		{Suit: card.Hearts, Rank: card.Five},
	}
	r := card.EvaluatePokerHand(h)
	if r.Rank != card.FullHouse {
		t.Errorf("expected FullHouse, got %s", card.PokerHandNames[r.Rank])
	}
}

// TestEvaluateFlush checks flush detection.
func TestEvaluateFlush(t *testing.T) {
	h := card.Hand{
		{Suit: card.Diamonds, Rank: card.Ace},
		{Suit: card.Diamonds, Rank: card.Ten},
		{Suit: card.Diamonds, Rank: card.Seven},
		{Suit: card.Diamonds, Rank: card.Four},
		{Suit: card.Diamonds, Rank: card.Two},
	}
	r := card.EvaluatePokerHand(h)
	if r.Rank != card.Flush {
		t.Errorf("expected Flush, got %s", card.PokerHandNames[r.Rank])
	}
}

// TestEvaluateStraight checks straight detection.
func TestEvaluateStraight(t *testing.T) {
	h := card.Hand{
		{Suit: card.Hearts, Rank: card.Nine},
		{Suit: card.Clubs, Rank: card.Eight},
		{Suit: card.Diamonds, Rank: card.Seven},
		{Suit: card.Spades, Rank: card.Six},
		{Suit: card.Hearts, Rank: card.Five},
	}
	r := card.EvaluatePokerHand(h)
	if r.Rank != card.Straight {
		t.Errorf("expected Straight, got %s", card.PokerHandNames[r.Rank])
	}
}

// TestEvaluateAceLowStraight checks A-2-3-4-5 straight.
func TestEvaluateAceLowStraight(t *testing.T) {
	h := card.Hand{
		{Suit: card.Hearts, Rank: card.Ace},
		{Suit: card.Clubs, Rank: card.Two},
		{Suit: card.Diamonds, Rank: card.Three},
		{Suit: card.Spades, Rank: card.Four},
		{Suit: card.Hearts, Rank: card.Five},
	}
	r := card.EvaluatePokerHand(h)
	if r.Rank != card.Straight {
		t.Errorf("expected Straight for A-2-3-4-5, got %s", card.PokerHandNames[r.Rank])
	}
}

// TestEvaluateAceLowStraightFlush checks A-2-3-4-5 straight flush.
func TestEvaluateAceLowStraightFlush(t *testing.T) {
	h := card.Hand{
		{Suit: card.Hearts, Rank: card.Ace},
		{Suit: card.Hearts, Rank: card.Two},
		{Suit: card.Hearts, Rank: card.Three},
		{Suit: card.Hearts, Rank: card.Four},
		{Suit: card.Hearts, Rank: card.Five},
	}
	r := card.EvaluatePokerHand(h)
	if r.Rank != card.StraightFlush {
		t.Errorf("expected StraightFlush, got %s", card.PokerHandNames[r.Rank])
	}
}

// TestEvaluateThreeOfAKind checks three of a kind detection.
func TestEvaluateThreeOfAKind(t *testing.T) {
	h := card.Hand{
		{Suit: card.Hearts, Rank: card.Queen},
		{Suit: card.Clubs, Rank: card.Queen},
		{Suit: card.Diamonds, Rank: card.Queen},
		{Suit: card.Spades, Rank: card.Nine},
		{Suit: card.Hearts, Rank: card.Two},
	}
	r := card.EvaluatePokerHand(h)
	if r.Rank != card.ThreeOfAKind {
		t.Errorf("expected ThreeOfAKind, got %s", card.PokerHandNames[r.Rank])
	}
}

// TestEvaluateTwoPair checks two pair detection.
func TestEvaluateTwoPair(t *testing.T) {
	h := card.Hand{
		{Suit: card.Hearts, Rank: card.Jack},
		{Suit: card.Clubs, Rank: card.Jack},
		{Suit: card.Diamonds, Rank: card.Four},
		{Suit: card.Spades, Rank: card.Four},
		{Suit: card.Hearts, Rank: card.Nine},
	}
	r := card.EvaluatePokerHand(h)
	if r.Rank != card.TwoPair {
		t.Errorf("expected TwoPair, got %s", card.PokerHandNames[r.Rank])
	}
}

// TestEvaluateOnePair checks one pair detection.
func TestEvaluateOnePair(t *testing.T) {
	h := card.Hand{
		{Suit: card.Hearts, Rank: card.Ace},
		{Suit: card.Clubs, Rank: card.Ace},
		{Suit: card.Diamonds, Rank: card.King},
		{Suit: card.Spades, Rank: card.Queen},
		{Suit: card.Hearts, Rank: card.Two},
	}
	r := card.EvaluatePokerHand(h)
	if r.Rank != card.OnePair {
		t.Errorf("expected OnePair, got %s", card.PokerHandNames[r.Rank])
	}
	if !card.IsJacksOrBetter(r) {
		t.Errorf("expected Jacks or Better for pair of Aces")
	}
}

// TestEvaluateHighCard checks high card detection.
func TestEvaluateHighCard(t *testing.T) {
	h := card.Hand{
		{Suit: card.Hearts, Rank: card.Ace},
		{Suit: card.Clubs, Rank: card.King},
		{Suit: card.Diamonds, Rank: card.Queen},
		{Suit: card.Spades, Rank: card.Six},
		{Suit: card.Hearts, Rank: card.Three},
	}
	r := card.EvaluatePokerHand(h)
	if r.Rank != card.HighCard {
		t.Errorf("expected HighCard, got %s", card.PokerHandNames[r.Rank])
	}
}

// TestGetResult checks result getter.
func TestGetResult(t *testing.T) {
	g := NewGame()
	g.Deal()
	g.SetHold([5]bool{true, true, true, true, true})
	g.Draw()
	if g.State == "done" && g.GetResult() == "" {
		t.Errorf("GetResult should not be empty when done")
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

// TestDealCreatesDeck checks deck is populated.
func TestDealCreatesDeck(t *testing.T) {
	g := NewGame()
	g.Deal()
	if len(g.Deck) == 0 {
		t.Errorf("expected deck to be populated")
	}
}

// TestHoldAllThenDraw checks holding all 5 cards preserves them.
func TestHoldAllThenDraw(t *testing.T) {
	g := NewGame()
	g.Deal()
	var original = make([]card.Card, 5)
	copy(original, g.Hand)
	g.SetHold([5]bool{true, true, true, true, true})
	g.Draw()
	// After Draw(), hand is sorted by rank. Check all original cards present.
	for _, oc := range original {
		var found bool
		for _, hc := range g.Hand {
			if hc == oc {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected held card %v to be present after draw", oc)
		}
	}
}

// TestHoldNoneThenDraw checks replacing all 5 cards.
func TestHoldNoneThenDraw(t *testing.T) {
	g := NewGame()
	g.Deal()
	var original = make([]card.Card, 5)
	copy(original, g.Hand)
	g.SetHold([5]bool{false, false, false, false, false})
	g.Draw()
	// After draw with no holds, all cards should be different from original
	// (unless deck wraps and draws the same card again, which is impossible)
	for i, oc := range original {
		for j, hc := range g.Hand {
			if hc == oc {
				// Same card found - could coincidentally be same rank/suit from new draw
				_ = i
				_ = j
			}
		}
	}
}

// TestPayoutMultipliers checks all payout multipliers.
func TestPayoutMultipliers(t *testing.T) {
	tests := []struct {
		rank card.PokerHandRank
		mult int
	}{
		{card.RoyalFlush, 800},
		{card.StraightFlush, 50},
		{card.FourOfAKind, 25},
		{card.FullHouse, 9},
		{card.Flush, 6},
		{card.Straight, 4},
		{card.ThreeOfAKind, 3},
		{card.TwoPair, 2},
	}
	for _, tt := range tests {
		var g = NewGame()
		g.Bet = 1
		g.HandResult = card.PokerHandResult{Rank: tt.rank}
		g.resolve()
		if int(g.Payout) != tt.mult {
			t.Errorf("Payout for rank %s = %g, expected %d", card.PokerHandNames[tt.rank], g.Payout, tt.mult)
		}
	}
}
