package blackjack

import (
	"github.com/slotopol/server/game/card"
)

// Game implements a standard Blackjack game.
type Game struct {
	card.CardBase `yaml:",inline"`
	Hand          card.Hand `json:"hand" yaml:"hand" xml:"hand"`
	DealerHand    card.Hand `json:"dealerHand" yaml:"dealerHand" xml:"dealerHand"`
	DealerHidden  bool      `json:"dealerHidden" yaml:"dealerHidden" xml:"dealerHidden"`
	Doubled       bool      `json:"doubled" yaml:"doubled" xml:"doubled"`
}

// Compile-time interface check.
var _ card.CardGame = (*Game)(nil)

// NewGame creates a new Blackjack game.
func NewGame() *Game {
	return &Game{
		CardBase: card.CardBase{
			Bet:   1,
			State: "waiting",
		},
		DealerHidden: true,
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

// resetHand prepares for a new round.
func (g *Game) resetHand() {
	g.Hand = nil
	g.DealerHand = nil
	g.DealerHidden = true
	g.Doubled = false
	g.Payout = 0
	g.Result = ""
	g.State = "dealing"
}

// GetHand returns player's hand.
func (g *Game) GetHand() card.Hand {
	return g.Hand
}

// GetDealerHand returns dealer's hand.
func (g *Game) GetDealerHand() card.Hand {
	return g.DealerHand
}

// Deal deals initial two cards to player and dealer.
func (g *Game) Deal() error {
	g.resetHand()
	g.Deck = card.NewDeck()

	// Player gets two cards
	g.Hand = append(g.Hand, g.dealCard())
	g.Hand = append(g.Hand, g.dealCard())

	// Dealer gets two cards (one hidden)
	g.DealerHand = append(g.DealerHand, g.dealCard())
	g.DealerHand = append(g.DealerHand, g.dealCard())

	// Check for player blackjack
	if card.HandValue(g.Hand) == 21 {
		g.DealerHidden = false
		// Dealer draws
		for card.HandValue(g.DealerHand) < 17 {
			g.DealerHand = append(g.DealerHand, g.dealCard())
		}
		var dv = card.HandValue(g.DealerHand)
		if dv == 21 && len(g.DealerHand) == 2 {
			// Both have blackjack - push
			g.Payout = g.Bet
			g.Result = "Push - both have Blackjack"
		} else {
			// Player blackjack pays 3:2
			g.Payout = g.Bet + g.Bet*1.5
			g.Result = "Blackjack!"
		}
		g.State = "done"
		return nil
	}

	// Check for dealer blackjack (shows if upcard is A or 10)
	if g.DealerHand[0].Rank == card.Ace || g.DealerHand[0].Rank >= card.Ten {
		g.DealerHidden = false
		if card.HandValue(g.DealerHand) == 21 {
			// Dealer has blackjack
			g.Result = "Dealer Blackjack"
			g.Payout = 0
			g.State = "done"
			return nil
		}
		g.DealerHidden = true
	}

	g.State = "playing"
	return nil
}

// Hit deals another card to the player.
func (g *Game) Hit() error {
	if g.State != "playing" {
		return card.ErrNoGame
	}
	g.Hand = append(g.Hand, g.dealCard())
	if card.HandValue(g.Hand) > 21 {
		// Bust
		g.DealerHidden = false
		g.Result = "Bust"
		g.Payout = 0
		g.State = "done"
	}
	return nil
}

// Stand ends the player's turn and plays the dealer's hand.
func (g *Game) Stand() error {
	if g.State != "playing" {
		return card.ErrNoGame
	}
	g.DealerHidden = false
	// Dealer draws to 17
	for card.HandValue(g.DealerHand) < 17 {
		g.DealerHand = append(g.DealerHand, g.dealCard())
	}
	g.resolve()
	return nil
}

// Double doubles the bet and takes exactly one more card, then stands.
func (g *Game) Double() error {
	if g.State != "playing" {
		return card.ErrNoGame
	}
	if len(g.Hand) != 2 {
		return card.ErrNoGame
	}
	g.Doubled = true
	g.Hand = append(g.Hand, g.dealCard())
	if card.HandValue(g.Hand) > 21 {
		g.DealerHidden = false
		g.Result = "Bust"
		g.Payout = 0
		g.State = "done"
		return nil
	}
	// Stand after double
	g.DealerHidden = false
	for card.HandValue(g.DealerHand) < 17 {
		g.DealerHand = append(g.DealerHand, g.dealCard())
	}
	g.resolve()
	return nil
}

// resolve compares hands after player stands.
// Payout is total return (stake + winnings combined).
func (g *Game) resolve() {
	var pv = card.HandValue(g.Hand)
	var dv = card.HandValue(g.DealerHand)

	g.DealerHidden = false
	g.State = "done"

	var stake = g.Bet
	if g.Doubled {
		stake = g.Bet * 2
	}

	if dv > 21 {
		g.Result = "Dealer Bust"
		// Return stake + stake (1:1 on the actual stake)
		g.Payout = stake * 2
		return
	}
	if pv > dv {
		g.Result = "Win"
		g.Payout = stake * 2
	} else if pv == dv {
		g.Result = "Push"
		g.Payout = stake
	} else {
		g.Result = "Lose"
		g.Payout = 0
	}
}

// GetDealerUpcard returns the dealer's visible card.
func (g *Game) GetDealerUpcard() card.Card {
	if len(g.DealerHand) > 0 {
		return g.DealerHand[0]
	}
	return card.Card{}
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
