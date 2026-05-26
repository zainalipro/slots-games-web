//go:build !prod || full || card

package baccarat

import (
	"context"
	"fmt"

	"github.com/slotopol/server/game"
)

func CalcStat(ctx context.Context, sp *game.ScanPar) (float64, float64) {
	fmt.Println()
	fmt.Println("Baccarat — card game")
	fmt.Println("Analytic RTP calculation not available via slot scanner.")
	fmt.Println("Standard Baccarat RTP: Player bet ~98.76%, Banker bet ~98.94%, Tie bet ~85.64%")
	return 0.9894, 0
}

func init() {
	var gi = game.AlgInfo{
		Aliases: []game.GameAlias{
			{Prov: "slotopol", Name: "Baccarat"},
		},
		AlgDescr: game.AlgDescr{
			GT:  game.GTcard,
			SX:  0, // not a grid game
			SN:  0, // not a symbol game
			RTP: []float64{98.94}, // banker bet RTP
		},
	}
	gi.SetupFactory(func(sel int) game.Gamble { return NewGame() }, CalcStat)
}
