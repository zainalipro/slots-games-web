package api

import (
	"encoding/json"
	"encoding/xml"
	"log"

	"github.com/gin-gonic/gin"

	cfg "github.com/slotopol/server/config"
	"github.com/slotopol/server/game/crash"
	"github.com/slotopol/server/game/crash/aviator"
	"github.com/slotopol/server/util"
)

// Returns bet value.
func ApiAviatorBetGet(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr" form:"gid" binding:"required"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_avi_betget_nobind, err)
		return
	}

	var scene *Scene
	if scene, err = GetScene(arg.GID); err != nil {
		Ret404(c, AEC_avi_betget_noscene, err)
		return
	}
	var game crash.CrashGame
	if game, ok = scene.Game.(crash.CrashGame); !ok {
		Ret403(c, AEC_avi_betget_notcrash, ErrNotCrash)
		return
	}

	var admin, al = MustAdmin(c, scene.CID)
	if admin.UID != scene.UID && al&ALdealer == 0 {
		Ret403(c, AEC_avi_betget_noaccess, ErrNoAccess)
		return
	}

	var ret struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"ret"`
		Bet     float64  `json:"bet" yaml:"bet" xml:"bet"`
	}
	ret.Bet = game.GetBet()
	RetOk(c, ret)
}

// Set bet value.
func ApiAviatorBetSet(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr" binding:"required"`
		Bet     float64  `json:"bet" yaml:"bet" xml:"bet" binding:"required"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_avi_betset_nobind, err)
		return
	}

	var scene *Scene
	if scene, err = GetScene(arg.GID); err != nil {
		Ret404(c, AEC_avi_betset_noscene, err)
		return
	}
	var game crash.CrashGame
	if game, ok = scene.Game.(crash.CrashGame); !ok {
		Ret403(c, AEC_avi_betset_notcrash, ErrNotCrash)
		return
	}

	var admin, al = MustAdmin(c, scene.CID)
	if admin.UID != scene.UID && al&ALdealer == 0 {
		Ret403(c, AEC_avi_betset_noaccess, ErrNoAccess)
		return
	}

	if err = game.SetBet(arg.Bet); err != nil {
		Ret403(c, AEC_avi_betset_badbet, err)
		return
	}

	Ret204(c)
}

// Launch starts a new Aviator round.
func ApiAviatorLaunch(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr" form:"gid" binding:"required"`
		Bet     float64  `json:"bet,omitempty" yaml:"bet,omitempty" xml:"bet,omitempty"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_avi_launch_nobind, err)
		return
	}

	var scene *Scene
	if scene, err = GetScene(arg.GID); err != nil {
		Ret404(c, AEC_avi_launch_noscene, err)
		return
	}
	var g crash.CrashGame
	if g, ok = scene.Game.(crash.CrashGame); !ok {
		Ret403(c, AEC_avi_launch_notcrash, ErrNotCrash)
		return
	}

	var club *Club
	if club, ok = Clubs.Get(scene.CID); !ok {
		Ret500(c, AEC_avi_launch_noclub, ErrNoClub)
		return
	}

	var user *User
	if user, ok = Users.Get(scene.UID); !ok {
		Ret500(c, AEC_avi_launch_nouser, ErrNoUser)
		return
	}

	var admin, al = MustAdmin(c, scene.CID)
	if admin.UID != scene.UID && al&ALdealer == 0 {
		Ret403(c, AEC_avi_launch_noaccess, ErrNoAccess)
		return
	}

	if arg.Bet != 0 {
		if err = g.SetBet(arg.Bet); err != nil {
			Ret403(c, AEC_avi_launch_badbet, err)
			return
		}
	}

	var cost = g.GetBet()

	var props *Props
	if props, ok = user.props.Get(scene.CID); !ok {
		Ret500(c, AEC_avi_launch_noprops, ErrNoProps)
		return
	}
	if props.Wallet < cost {
		Ret403(c, AEC_avi_launch_nomoney, ErrNoMoney)
		return
	}

	var bank = club.Bank()
	_ = bank

	if err = g.Launch(); err != nil {
		Ret500(c, AEC_avi_launch_badbank, err)
		return
	}

	// Debit the bet immediately
	var debit = cost

	if Cfg.ClubUpdateBuffer > 1 {
		go BankBat[scene.CID].Put(cfg.XormStorage, scene.UID, debit)
	} else if err = BankBat[scene.CID].Put(cfg.XormStorage, scene.UID, debit); err != nil {
		Ret500(c, AEC_avi_launch_sqlbank, err)
		return
	}

	club.AddBank(debit)
	props.Wallet -= debit

	var sid = SpinCounter.Inc()
	scene.SID = sid
	var rec = Spinlog{
		SID:    sid,
		GID:    arg.GID,
		MRTP:   0,
		Gain:   0,
		Wallet: props.Wallet,
	}
	var b []byte
	if b, err = json.Marshal(scene.Game); err != nil {
		return
	}
	rec.Game = util.B2S(b)

	if Cfg.UseSpinLog {
		go func() {
			if err = SpinBuf.Put(cfg.XormSpinlog, rec); err != nil {
				log.Printf("can not write to spin log: %s", err.Error())
			}
		}()
	}

	var ret struct {
		XMLName    xml.Name        `json:"-" yaml:"-" xml:"ret"`
		SID        uint64          `json:"sid" yaml:"sid" xml:"sid,attr"`
		Game       crash.CrashGame `json:"game" yaml:"game" xml:"game"`
		CrashPoint float64         `json:"crashPoint" yaml:"crashPoint" xml:"crashPoint"`
		Multiplier float64         `json:"multiplier" yaml:"multiplier" xml:"multiplier"`
		Wallet     float64         `json:"wallet" yaml:"wallet" xml:"wallet"`
		State      string          `json:"state" yaml:"state" xml:"state"`
	}
	ret.SID = sid
	ret.Game = g
	ret.CrashPoint = g.GetCrashPoint()
	ret.Multiplier = g.GetMultiplier()
	ret.Wallet = props.Wallet
	ret.State = g.GetGameState()

	RetOk(c, ret)
}

// CashOut cashes out the current Aviator round.
func ApiAviatorCashOut(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr" form:"gid" binding:"required"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_avi_cashout_nobind, err)
		return
	}

	var scene *Scene
	if scene, err = GetScene(arg.GID); err != nil {
		Ret404(c, AEC_avi_cashout_noscene, err)
		return
	}
	var g crash.CrashGame
	if g, ok = scene.Game.(crash.CrashGame); !ok {
		Ret403(c, AEC_avi_cashout_notcrash, ErrNotCrash)
		return
	}

	var admin, al = MustAdmin(c, scene.CID)
	if admin.UID != scene.UID && al&ALdealer == 0 {
		Ret403(c, AEC_avi_cashout_noaccess, ErrNoAccess)
		return
	}

	var payout float64
	if payout, err = g.CashOut(); err != nil {
		Ret403(c, AEC_avi_cashout_nogame, err)
		return
	}

	// Credit the payout back
	var user *User
	if user, ok = Users.Get(scene.UID); !ok {
		Ret500(c, AEC_avi_cashout_nogame, ErrNoUser)
		return
	}
	var props *Props
	if props, ok = user.props.Get(scene.CID); !ok {
		Ret500(c, AEC_avi_cashout_nogame, ErrNoProps)
		return
	}
	var club *Club
	if club, ok = Clubs.Get(scene.CID); !ok {
		Ret500(c, AEC_avi_cashout_nogame, ErrNoClub)
		return
	}

	// Credit the winnings (negative debit = credit to wallet)
	var debit = -payout

	if Cfg.ClubUpdateBuffer > 1 {
		go BankBat[scene.CID].Put(cfg.XormStorage, scene.UID, debit)
	} else if err = BankBat[scene.CID].Put(cfg.XormStorage, scene.UID, debit); err != nil {
		Ret500(c, AEC_avi_cashout_nogame, err)
		return
	}

	club.AddBank(debit)
	props.Wallet -= debit

	var ret struct {
		XMLName    xml.Name        `json:"-" yaml:"-" xml:"ret"`
		Game       crash.CrashGame `json:"game" yaml:"game" xml:"game"`
		Multiplier float64         `json:"multiplier" yaml:"multiplier" xml:"multiplier"`
		Payout     float64         `json:"payout" yaml:"payout" xml:"payout"`
		Wallet     float64         `json:"wallet" yaml:"wallet" xml:"wallet"`
		State      string          `json:"state" yaml:"state" xml:"state"`
	}
	ret.Game = g
	ret.Multiplier = g.GetMultiplier()
	ret.Payout = payout
	ret.Wallet = props.Wallet
	ret.State = g.GetGameState()

	RetOk(c, ret)
}

// State returns the current game state (useful for polling during a round).
func ApiAviatorState(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr" form:"gid" binding:"required"`
		Advance bool     `json:"advance,omitempty" yaml:"advance,omitempty" xml:"advance,omitempty"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_avi_betget_nobind, err)
		return
	}

	var scene *Scene
	if scene, err = GetScene(arg.GID); err != nil {
		Ret404(c, AEC_avi_betget_noscene, err)
		return
	}
	var g crash.CrashGame
	if g, ok = scene.Game.(crash.CrashGame); !ok {
		Ret403(c, AEC_avi_betget_notcrash, ErrNotCrash)
		return
	}

	// Advance the multiplier if requested
	if arg.Advance {
		if avGame, ok2 := g.(*aviator.Game); ok2 {
			avGame.Tick()
		}
	}

	var ret struct {
		XMLName    xml.Name        `json:"-" yaml:"-" xml:"ret"`
		Multiplier float64         `json:"multiplier" yaml:"multiplier" xml:"multiplier"`
		CrashPoint float64         `json:"crashPoint" yaml:"crashPoint" xml:"crashPoint"`
		State      string          `json:"state" yaml:"state" xml:"state"`
		Payout     float64         `json:"payout" yaml:"payout" xml:"payout"`
		Game       crash.CrashGame `json:"game" yaml:"game" xml:"game"`
	}
	ret.Multiplier = g.GetMultiplier()
	ret.CrashPoint = g.GetCrashPoint()
	ret.State = g.GetGameState()
	ret.Payout = g.GetPayout()
	ret.Game = g

	RetOk(c, ret)
}
