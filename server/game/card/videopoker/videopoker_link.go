//go:build !prod || full || card

package videopoker

import (
	"context"
	"fmt"

	"github.com/slotopol/server/game"
)

func CalcStat(ctx context.Context, sp *game.ScanPar) (float64, float64) {
	fmt.Println()
	fmt.Println("Video Poker (Jacks or Better) — card game")
	fmt.Println("Analytic RTP calculation not available via slot scanner.")
	fmt.Println("Standard Jacks or Better RTP with optimal strategy: ~99.54%")
	return 0.9954, 0
}

func init() {
	var gi = game.AlgInfo{
		Aliases: []game.GameAlias{
			{Prov: "slotopol", Name: "Video Poker"},
		},
		AlgDescr: game.AlgDescr{
			GT:  game.GTcard,
			SX:  0,
			SN:  0,
			RTP: []float64{99.54}, // Jacks or Better with optimal strategy
		},
	}
	gi.SetupFactory(func(sel int) game.Gamble { return NewGame() }, CalcStat)
}
