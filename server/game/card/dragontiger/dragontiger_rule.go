package dragontiger

import (
	"github.com/slotopol/server/game/card"
)

// BetTarget represents where the player bets.
type BetTarget string

const (
	BetDragon BetTarget = "dragon"
	BetTiger  BetTarget = "tiger"
	BetDTie   BetTarget = "tie"
	BetDSuitedTie BetTarget = "suitedtie"
	BetDBig   BetTarget = "big"
	BetDSmall BetTarget = "small"
)

// Game implements a standard Dragon Tiger card game.
type Game struct {
	card.CardBase `yaml:",inline"`
	DragonHand    card.Hand  `json:"dragonHand" yaml:"dragonHand" xml:"dragonHand"`
	TigerHand     card.Hand  `json:"tigerHand" yaml:"tigerHand" xml:"tigerHand"`
	BetTarget     BetTarget  `json:"betTarget" yaml:"betTarget" xml:"betTarget"`
	DragonValue   int        `json:"dragonValue" yaml:"dragonValue" xml:"dragonValue"`
	TigerValue    int        `json:"tigerValue" yaml:"tigerValue" xml:"tigerValue"`
}

// Compile-time interface check.
var _ card.CardGame = (*Game)(nil)

// NewGame creates a new Dragon Tiger game.
func NewGame() *Game {
	return &Game{
		CardBase: card.CardBase{
			Bet:   1,
			State: "waiting",
		},
		BetTarget: BetDragon,
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

// GetHand returns dragon's hand (for CardGame interface).
func (g *Game) GetHand() card.Hand {
	return g.DragonHand
}

// GetDealerHand returns tiger's hand (for CardGame interface).
func (g *Game) GetDealerHand() card.Hand {
	return g.TigerHand
}

// dragonRankValue returns the rank value for comparison (Ace=14, King=13, etc.)
func dragonRankValue(r card.Rank) int {
	switch r {
	case card.Ace:
		return 14
	case card.Two:
		return 2
	case card.Three:
		return 3
	case card.Four:
		return 4
	case card.Five:
		return 5
	case card.Six:
		return 6
	case card.Seven:
		return 7
	case card.Eight:
		return 8
	case card.Nine:
		return 9
	case card.Ten:
		return 10
	case card.Jack:
		return 11
	case card.Queen:
		return 12
	case card.King:
		return 13
	}
	return 0
}

// isHigh checks if the card value is 8 or above (for Big bet)
func isHigh(r card.Rank) bool {
	v := dragonRankValue(r)
	return v >= 8 && v <= 14
}

// isLow checks if the card value is 6 or below (for Small bet)
func isLow(r card.Rank) bool {
	v := dragonRankValue(r)
	return v >= 2 && v <= 6
}

// Deal deals one card each to Dragon and Tiger.
func (g *Game) Deal() error {
	g.DragonHand = nil
	g.TigerHand = nil
	g.Payout = 0
	g.Result = ""
	g.State = "dealing"
	g.Deck = card.NewDeck()

	// Deal one card to Dragon, one to Tiger
	g.DragonHand = append(g.DragonHand, g.dealCard())
	g.TigerHand = append(g.TigerHand, g.dealCard())

	g.DragonValue = dragonRankValue(g.DragonHand[0].Rank)
	g.TigerValue = dragonRankValue(g.TigerHand[0].Rank)

	g.resolve()
	return nil
}

// resolve determines the winner.
func (g *Game) resolve() {
	var dv = g.DragonValue
	var tv = g.TigerValue
	g.State = "done"

	switch g.BetTarget {
	case BetDragon:
		if dv > tv {
			g.Result = "Dragon wins"
			g.Payout = g.Bet * 2 // 1:1
		} else if dv == tv {
			// Tie - player loses half
			g.Result = "Tie - half lost"
			g.Payout = g.Bet * 0.5
		} else {
			g.Result = "Tiger wins"
			g.Payout = 0
		}
	case BetTiger:
		if tv > dv {
			g.Result = "Tiger wins"
			g.Payout = g.Bet * 2 // 1:1
		} else if tv == dv {
			g.Result = "Tie - half lost"
			g.Payout = g.Bet * 0.5
		} else {
			g.Result = "Dragon wins"
			g.Payout = 0
		}
	case BetDTie:
		if dv == tv {
			g.Result = "Tie wins!"
			g.Payout = g.Bet * 9 // 8:1
		} else {
			g.Result = "No tie"
			g.Payout = 0
		}
	case BetDSuitedTie:
		if dv == tv && g.DragonHand[0].Suit == g.TigerHand[0].Suit {
			g.Result = "Suited Tie wins!"
			g.Payout = g.Bet * 51 // 50:1
		} else {
			g.Result = "No suited tie"
			g.Payout = 0
		}
	case BetDBig:
		if dv == tv {
			// On tie, big/small bets lose half
			g.Result = "Tie - half lost"
			g.Payout = g.Bet * 0.5
		} else if isHigh(g.DragonHand[0].Rank) && g.BetTarget == BetDBig {
			g.Result = "Big wins"
			g.Payout = g.Bet * 2
		} else if isHigh(g.TigerHand[0].Rank) {
			g.Result = "Big wins"
			g.Payout = g.Bet * 2
		} else {
			g.Result = "Small"
			g.Payout = 0
		}
	case BetDSmall:
		if dv == tv {
			g.Result = "Tie - half lost"
			g.Payout = g.Bet * 0.5
		} else if isLow(g.DragonHand[0].Rank) {
			g.Result = "Small wins"
			g.Payout = g.Bet * 2
		} else if isLow(g.TigerHand[0].Rank) {
			g.Result = "Small wins"
			g.Payout = g.Bet * 2
		} else {
			g.Result = "Big"
			g.Payout = 0
		}
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
