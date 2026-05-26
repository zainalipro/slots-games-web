//go:build !prod || full || card

package blackjack

import (
	"context"
	"fmt"

	"github.com/slotopol/server/game"
)

func CalcStat(ctx context.Context, sp *game.ScanPar) (float64, float64) {
	fmt.Println()
	fmt.Println("Blackjack — card game")
	fmt.Println("Analytic RTP calculation not available via slot scanner.")
	fmt.Println("Standard Blackjack RTP with basic strategy: ~99.5%")
	return 0.995, 0
}

func init() {
	var gi = game.AlgInfo{
		Aliases: []game.GameAlias{
			{Prov: "slotopol", Name: "Blackjack"},
		},
		AlgDescr: game.AlgDescr{
			GT:  game.GTcard,
			SX:  0, // not a grid game
			SN:  0, // not a symbol game
			RTP: []float64{99.5}, // standard blackjack RTP with basic strategy
		},
	}
	gi.SetupFactory(func(sel int) game.Gamble { return NewGame() }, CalcStat)
}
