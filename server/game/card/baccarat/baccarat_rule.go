package baccarat

import (
	"github.com/slotopol/server/game/card"
)

// BetTarget represents where the player bets.
type BetTarget string

const (
	BetPlayer BetTarget = "player"
	BetBanker BetTarget = "banker"
	BetTie    BetTarget = "tie"
)

// Game implements a standard Baccarat game.
type Game struct {
	card.CardBase `yaml:",inline"`
	PlayerHand    card.Hand `json:"playerHand" yaml:"playerHand" xml:"playerHand"`
	BankerHand    card.Hand `json:"bankerHand" yaml:"bankerHand" xml:"bankerHand"`
	BetTarget     BetTarget `json:"betTarget" yaml:"betTarget" xml:"betTarget"`
}

// Compile-time interface check.
var _ card.CardGame = (*Game)(nil)

// NewGame creates a new Baccarat game.
func NewGame() *Game {
	return &Game{
		CardBase: card.CardBase{
			Bet:   1,
			State: "waiting",
		},
		BetTarget: BetPlayer,
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

// GetHand returns player's hand (for CardGame interface).
func (g *Game) GetHand() card.Hand {
	return g.PlayerHand
}

// GetDealerHand returns banker's hand (for CardGame interface).
func (g *Game) GetDealerHand() card.Hand {
	return g.BankerHand
}

// Deal deals the initial cards according to Baccarat rules.
func (g *Game) Deal() error {
	g.PlayerHand = nil
	g.BankerHand = nil
	g.Payout = 0
	g.Result = ""
	g.State = "dealing"
	g.Deck = card.NewDeck()

	// Deal initial 2 cards each
	g.PlayerHand = append(g.PlayerHand, g.dealCard())
	g.BankerHand = append(g.BankerHand, g.dealCard())
	g.PlayerHand = append(g.PlayerHand, g.dealCard())
	g.BankerHand = append(g.BankerHand, g.dealCard())

	// Check for natural (8 or 9)
	var pv = card.BaccaratValue(g.PlayerHand)
	var bv = card.BaccaratValue(g.BankerHand)

	if pv >= 8 || bv >= 8 {
		g.resolve()
		return nil
	}

	// Player third card rules
	if pv <= 5 {
		g.PlayerHand = append(g.PlayerHand, g.dealCard())
	}

	// Banker third card rules (depends on player's third card)
	var playerThird card.Card
	var hasPlayerThird = len(g.PlayerHand) == 3
	if hasPlayerThird {
		playerThird = g.PlayerHand[2]
	}

	if bv <= 2 {
		// Banker draws if <= 2
		g.BankerHand = append(g.BankerHand, g.dealCard())
	} else if bv == 3 && (!hasPlayerThird || playerThird.Rank != 8) {
		g.BankerHand = append(g.BankerHand, g.dealCard())
	} else if bv == 4 && hasPlayerThird && playerThird.Rank >= 2 && playerThird.Rank <= 7 {
		g.BankerHand = append(g.BankerHand, g.dealCard())
	} else if bv == 5 && hasPlayerThird && playerThird.Rank >= 4 && playerThird.Rank <= 7 {
		g.BankerHand = append(g.BankerHand, g.dealCard())
	} else if bv == 6 && hasPlayerThird && playerThird.Rank >= 6 && playerThird.Rank <= 7 {
		g.BankerHand = append(g.BankerHand, g.dealCard())
	}
	// Banker stands on 7

	g.resolve()
	return nil
}

// resolve determines the winner.
func (g *Game) resolve() {
	var pv = card.BaccaratValue(g.PlayerHand)
	var bv = card.BaccaratValue(g.BankerHand)
	g.State = "done"

	switch g.BetTarget {
	case BetPlayer:
		if pv > bv {
			g.Result = "Player wins"
			g.Payout = g.Bet * 2 // 1:1 payout
		} else if pv == bv {
			g.Result = "Tie - bet returned"
			g.Payout = g.Bet // push
		} else {
			g.Result = "Banker wins"
			g.Payout = 0
		}
	case BetBanker:
		if bv > pv {
			g.Result = "Banker wins"
			// Banker pays 0.95:1 (5% commission)
			g.Payout = g.Bet + g.Bet*0.95
		} else if bv == pv {
			g.Result = "Tie - bet returned"
			g.Payout = g.Bet // push
		} else {
			g.Result = "Player wins"
			g.Payout = 0
		}
	case BetTie:
		if pv == bv {
			g.Result = "Tie wins!"
			g.Payout = g.Bet * 9 // 8:1 payout
		} else {
			g.Result = "No tie"
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
