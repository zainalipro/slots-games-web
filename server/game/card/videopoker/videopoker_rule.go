package videopoker

import (
	"sort"

	"github.com/slotopol/server/game/card"
)

// Game implements a Video Poker (Jacks or Better) game.
type Game struct {
	card.CardBase `yaml:",inline"`
	Hand          card.Hand   `json:"hand" yaml:"hand" xml:"hand"`
	HoldMask      [5]bool     `json:"holdMask" yaml:"holdMask" xml:"holdMask"`
	HandResult    card.PokerHandResult `json:"handResult,omitempty" yaml:"handResult,omitempty" xml:"handResult,omitempty"`
}

// Compile-time interface check.
var _ card.CardGame = (*Game)(nil)

// NewGame creates a new Video Poker game.
func NewGame() *Game {
	return &Game{
		CardBase: card.CardBase{
			Bet:   1,
			State: "waiting",
		},
	}
}

// dealCard draws a card from the deck.
func (g *Game) dealCard() card.Card {
	if len(g.Deck) == 0 {
		g.Deck = card.NewDeck()
	}
	var c = g.Deck[0]
	g.Deck = g.Deck[1:]
	return c
}

// GetHand returns the player's hand.
func (g *Game) GetHand() card.Hand {
	return g.Hand
}

// GetDealerHand returns nil (no dealer in video poker).
func (g *Game) GetDealerHand() card.Hand {
	return nil
}

// Deal deals 5 cards to the player.
func (g *Game) Deal() error {
	g.Hand = nil
	g.HoldMask = [5]bool{}
	g.Payout = 0
	g.Result = ""
	g.State = "dealing"
	g.Deck = card.NewDeck()

	// Deal 5 cards
	for i := 0; i < 5; i++ {
		g.Hand = append(g.Hand, g.dealCard())
	}

	g.State = "holding"
	return nil
}

// SetHold sets which cards to hold (true = keep, false = discard).
func (g *Game) SetHold(hold [5]bool) error {
	if g.State != "holding" {
		return card.ErrNoGame
	}
	g.HoldMask = hold
	return nil
}

// Draw replaces the non-held cards and evaluates the hand.
func (g *Game) Draw() error {
	if g.State != "holding" {
		return card.ErrNoGame
	}

	// Replace non-held cards
	for i := 0; i < 5; i++ {
		if !g.HoldMask[i] {
			g.Hand[i] = g.dealCard()
		}
	}

	// Sort hand for evaluation
	sort.Slice(g.Hand, func(i, j int) bool {
		return g.Hand[i].Rank > g.Hand[j].Rank
	})

	// Evaluate hand
	g.HandResult = card.EvaluatePokerHand(g.Hand)
	g.resolve()
	g.State = "done"
	return nil
}

// resolve calculates payout based on poker hand ranking.
func (g *Game) resolve() {
	var bet = g.Bet

	switch g.HandResult.Rank {
	case card.RoyalFlush:
		g.Payout = bet * 800
		g.Result = "Royal Flush!"
	case card.StraightFlush:
		g.Payout = bet * 50
		g.Result = "Straight Flush!"
	case card.FourOfAKind:
		g.Payout = bet * 25
		g.Result = "Four of a Kind!"
	case card.FullHouse:
		g.Payout = bet * 9
		g.Result = "Full House!"
	case card.Flush:
		g.Payout = bet * 6
		g.Result = "Flush!"
	case card.Straight:
		g.Payout = bet * 4
		g.Result = "Straight!"
	case card.ThreeOfAKind:
		g.Payout = bet * 3
		g.Result = "Three of a Kind"
	case card.TwoPair:
		g.Payout = bet * 2
		g.Result = "Two Pair"
	case card.OnePair:
		if card.IsJacksOrBetter(g.HandResult) {
			g.Payout = bet * 1
			g.Result = "Jacks or Better"
		} else {
			g.Payout = 0
			g.Result = "Low Pair"
		}
	default:
		g.Payout = 0
		g.Result = "High Card"
	}
}

// Scanner is for compatibility with the Gamble interface.
func (g *Game) Scanner(wins *card.ShotResult) error {
	wins.Pay = g.Payout
	return nil
}

// Spin is for compatibility with the Gamble interface.
func (g *Game) Spin(_ float64) {
	g.Deal()
}
