//go:build !prod || full || fish

package fishhunter

import (
	"github.com/slotopol/server/game"
	"github.com/slotopol/server/game/fishing"
)

var Info = game.AlgInfo{
	Aliases: []game.GameAlias{
		{Prov: "Slotopol", Name: "Fish Feast", Date: game.Date(2025, 5, 26)},
	},
	AlgDescr: game.AlgDescr{
		GT:  game.GTfish,
		GP:  0,
		SX:  12, // pool size
		SY:  0,
		SN:  6,  // number of fish types
		LN:  0,
		BN:  0,
		RTP: []float64{94.6, 87.8, 91.1, 92.0, 94.1},
	},
	Update: func(ai *game.AlgInfo) {
		ai.RTP = game.MakeRtpList(ReelsMap)
	},
}

func init() {
	Info.SetupFactory(func(sel int) game.Gamble { return NewGame() }, fishing.CalcStat)
}
