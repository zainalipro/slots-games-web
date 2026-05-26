package card

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"sort"
	"strings"
)

// Suit represents a card suit.
type Suit byte

const (
	Clubs Suit = iota
	Diamonds
	Hearts
	Spades
)

var SuitNames = [4]string{"♣", "♦", "♥", "♠"}
var SuitWords = [4]string{"clubs", "diamonds", "hearts", "spades"}

// Rank represents a card rank (1=Ace, 2-10, 11=Jack, 12=Queen, 13=King).
type Rank byte

const (
	Ace   Rank = 1
	Two   Rank = 2
	Three Rank = 3
	Four  Rank = 4
	Five  Rank = 5
	Six   Rank = 6
	Seven Rank = 7
	Eight Rank = 8
	Nine  Rank = 9
	Ten   Rank = 10
	Jack  Rank = 11
	Queen Rank = 12
	King  Rank = 13
)

var RankNames = [13]string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}
var RankWords = [13]string{"Ace", "2", "3", "4", "5", "6", "7", "8", "9", "10", "Jack", "Queen", "King"}

// Card represents a single playing card.
type Card struct {
	Suit Suit `json:"suit" yaml:"suit" xml:"suit,attr"`
	Rank Rank `json:"rank" yaml:"rank" xml:"rank,attr"`
}

func (c Card) String() string {
	return RankNames[c.Rank-1] + SuitNames[c.Suit]
}

// Deck is a standard 52-card deck.
type Deck []Card

// NewDeck creates a new shuffled 52-card deck.
func NewDeck() Deck {
	var d = make(Deck, 52)
	var i int
	for s := Clubs; s <= Spades; s++ {
		for r := Ace; r <= King; r++ {
			d[i] = Card{Suit: s, Rank: r}
			i++
		}
	}
	rand.Shuffle(52, func(i, j int) {
		d[i], d[j] = d[j], d[i]
	})
	return d
}

// Hand is a collection of cards.
type Hand []Card

func (h Hand) String() string {
	var s []string
	for _, c := range h {
		s = append(s, c.String())
	}
	return strings.Join(s, " ")
}

// HandValue returns the blackjack-style value of a hand.
// Ace counts as 11 unless it would bust, then as 1.
func HandValue(h Hand) int {
	var total int
	var aces int
	for _, c := range h {
		switch c.Rank {
		case Ace:
			total += 11
			aces++
		case Jack, Queen, King:
			total += 10
		default:
			total += int(c.Rank)
		}
	}
	for total > 21 && aces > 0 {
		total -= 10
		aces--
	}
	return total
}

// BaccaratValue returns the baccarat-style value (0-9).
func BaccaratValue(h Hand) int {
	var total int
	for _, c := range h {
		switch c.Rank {
		case Ace:
			total += 1
		case Two, Three, Four, Five, Six, Seven, Eight, Nine:
			total += int(c.Rank)
		default:
			// 10, J, Q, K = 0
		}
	}
	return total % 10
}

// PokerHandRank represents the rank of a poker hand.
type PokerHandRank int

const (
	HighCard      PokerHandRank = 0
	OnePair       PokerHandRank = 1
	TwoPair       PokerHandRank = 2
	ThreeOfAKind  PokerHandRank = 3
	Straight      PokerHandRank = 4
	Flush         PokerHandRank = 5
	FullHouse     PokerHandRank = 6
	FourOfAKind   PokerHandRank = 7
	StraightFlush PokerHandRank = 8
	RoyalFlush    PokerHandRank = 9
)

var PokerHandNames = map[PokerHandRank]string{
	HighCard:      "High Card",
	OnePair:       "One Pair",
	TwoPair:       "Two Pair",
	ThreeOfAKind:  "Three of a Kind",
	Straight:      "Straight",
	Flush:         "Flush",
	FullHouse:     "Full House",
	FourOfAKind:   "Four of a Kind",
	StraightFlush: "Straight Flush",
	RoyalFlush:    "Royal Flush",
}

// PokerHandResult contains the hand rank and kickers for comparison.
type PokerHandResult struct {
	Rank    PokerHandRank `json:"rank" yaml:"rank" xml:"rank"`
	Cards   Hand          `json:"cards" yaml:"cards" xml:"cards"`
	Kickers []Rank        `json:"kickers" yaml:"kickers" xml:"kickers"`
}

// EvaluatePokerHand evaluates the best 5-card poker hand from 5 cards.
// For video poker, we just evaluate the 5 cards directly.
func EvaluatePokerHand(h Hand) PokerHandResult {
	if len(h) != 5 {
		return PokerHandResult{Rank: HighCard, Cards: h}
	}

	// Sort by rank descending
	var sorted = make(Hand, 5)
	copy(sorted, h)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Rank > sorted[j].Rank
	})

	var ranks [5]Rank
	var suits [5]Suit
	for i, c := range sorted {
		ranks[i] = c.Rank
		suits[i] = c.Suit
	}

	var isFlush = suits[0] == suits[1] && suits[1] == suits[2] && suits[2] == suits[3] && suits[3] == suits[4]

	// Check for straight
	// Since Ace (1) sorts lowest but can be high in Ace-high straights,
	// we check both normal descending sequences and Ace-high/ Ace-low.
	var isStraight bool
	var straightHigh Rank

	// Standard straight: 5 consecutive ranks (works for all except Ace-high)
	if ranks[0]-ranks[4] == 4 && ranks[0] > ranks[1] && ranks[1] > ranks[2] && ranks[2] > ranks[3] && ranks[3] > ranks[4] {
		isStraight = true
		straightHigh = ranks[0]
	}

	// Ace-high straight: A-K-Q-J-T
	// After descending sort: [K(13), Q(12), J(11), 10(10), A(1)]
	if !isStraight && ranks[0] == King && ranks[1] == Queen && ranks[2] == Jack && ranks[3] == Ten && ranks[4] == Ace {
		isStraight = true
		straightHigh = Ace
	}

	// Ace-low straight: A-2-3-4-5 (wheel)
	// After descending sort: [5, 4, 3, 2, A(1)] — already caught by the standard check
	// since 5-1=4 and 5>4>3>2>1. But the straightHigh would be 5 (Five).
	// This is already handled by the standard check above.

	if isFlush && isStraight && straightHigh == Ace {
		return PokerHandResult{Rank: RoyalFlush, Cards: sorted, Kickers: []Rank{Ace}}
	}
	if isFlush && isStraight {
		return PokerHandResult{Rank: StraightFlush, Cards: sorted, Kickers: []Rank{straightHigh}}
	}

	// Count rank occurrences
	var rankCount = make(map[Rank]int)
	for _, r := range ranks {
		rankCount[r]++
	}

	// Find quads, trips, pairs
	var fours, threes []Rank
	var pairs []Rank
	var singles []Rank
	for r, c := range rankCount {
		switch c {
		case 4:
			fours = append(fours, r)
		case 3:
			threes = append(threes, r)
		case 2:
			pairs = append(pairs, r)
		case 1:
			singles = append(singles, r)
		}
	}
	sort.Slice(fours, func(i, j int) bool { return fours[i] > fours[j] })
	sort.Slice(threes, func(i, j int) bool { return threes[i] > threes[j] })
	sort.Slice(pairs, func(i, j int) bool { return pairs[i] > pairs[j] })
	sort.Slice(singles, func(i, j int) bool { return singles[i] > singles[j] })

	if len(fours) == 1 {
		return PokerHandResult{Rank: FourOfAKind, Cards: sorted, Kickers: append([]Rank{fours[0]}, singles...)}
	}
	if len(threes) == 1 && len(pairs) == 1 {
		return PokerHandResult{Rank: FullHouse, Cards: sorted, Kickers: []Rank{threes[0], pairs[0]}}
	}
	if isFlush {
		return PokerHandResult{Rank: Flush, Cards: sorted, Kickers: ranks[:]}
	}
	if isStraight {
		return PokerHandResult{Rank: Straight, Cards: sorted, Kickers: []Rank{straightHigh}}
	}
	if len(threes) == 1 {
		return PokerHandResult{Rank: ThreeOfAKind, Cards: sorted, Kickers: append([]Rank{threes[0]}, singles...)}
	}
	if len(pairs) == 2 {
		var kickers = []Rank{pairs[0], pairs[1]}
		if len(singles) > 0 {
			kickers = append(kickers, singles[0])
		}
		return PokerHandResult{Rank: TwoPair, Cards: sorted, Kickers: kickers}
	}
	if len(pairs) == 1 {
		return PokerHandResult{Rank: OnePair, Cards: sorted, Kickers: append([]Rank{pairs[0]}, singles...)}
	}
	return PokerHandResult{Rank: HighCard, Cards: sorted, Kickers: ranks[:]}
}

// VideoPokerPaytable defines payouts for each hand rank.
type VideoPokerPaytable map[PokerHandRank]float64

// JacksOrBetterPaytable is the standard paytable for Jacks or Better.
var JacksOrBetterPaytable = VideoPokerPaytable{
	RoyalFlush:    800,
	StraightFlush: 50,
	FourOfAKind:   25,
	FullHouse:     9,
	Flush:         6,
	Straight:      4,
	ThreeOfAKind:  3,
	TwoPair:       2,
	OnePair:       1, // Jacks or Better (pair must be J+)
}

// CheckJacksOrBetter checks if a one-pair hand is Jacks or Better
// (pair of Jacks, Queens, Kings, or Aces).
func IsJacksOrBetter(result PokerHandResult) bool {
	if result.Rank != OnePair || len(result.Kickers) == 0 {
		return false
	}
	// Ace (1) is high: J(11), Q(12), K(13), A(1)
	return result.Kickers[0] >= Jack || result.Kickers[0] == Ace
}

// CardGame is the common interface for card-based games.
type CardGame interface {
	Deal() error                    // deal a new round
	GetBet() float64                // returns current bet
	SetBet(float64) error           // set bet
	GetHand() Hand                  // returns player's hand
	GetDealerHand() Hand            // returns dealer's hand (if applicable)
	GetResult() string              // returns round result description
	GetPayout() float64             // returns payout for current round
	GetGameState() string           // returns game state (waiting, dealing, playing, done)
	Scanner(*ShotResult) error      // for compatibility, not typically used
	Spin(float64)                   // for compatibility, deals a round
}

// CardBase provides common card game fields.
type CardBase struct {
	Bet     float64 `json:"bet" yaml:"bet" xml:"bet"`
	Deck    Deck    `json:"deck,omitempty" yaml:"deck,omitempty" xml:"deck,omitempty"`
	State   string  `json:"state" yaml:"state" xml:"state"`
	Payout  float64 `json:"payout" yaml:"payout" xml:"payout"`
	Result  string  `json:"result" yaml:"result" xml:"result"`
}

func (g *CardBase) GetBet() float64 {
	return g.Bet
}

func (g *CardBase) SetBet(bet float64) error {
	if bet <= 0 {
		return ErrBadParam
	}
	g.Bet = bet
	return nil
}

func (g *CardBase) GetPayout() float64 {
	return g.Payout
}

func (g *CardBase) GetResult() string {
	return g.Result
}

func (g *CardBase) GetGameState() string {
	return g.State
}

// ShotResult is a minimal placeholder for Scanner compatibility.
type ShotResult struct {
	Pay float64 `json:"pay" yaml:"pay" xml:"pay,attr"`
}

func (g *CardBase) Scanner(wins *ShotResult) error {
	wins.Pay = g.Payout
	return nil
}

func (g *CardBase) Spin(_ float64) {
	// Spin is used for compatibility; card games use Deal() explicitly.
	// For auto-play scenarios, this does nothing.
}

var (
	ErrBadParam = errors.New("wrong parameter")
	ErrNoGame   = errors.New("game round not started, deal first")
)

// Print_all prints statistics placeholders for card games (scanner not applicable).
func Print_all(sp interface{}) {
	fmt.Println("Card game statistics: analytic calculation not available")
}
