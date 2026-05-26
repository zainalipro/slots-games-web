package api

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"log"

	"github.com/gin-gonic/gin"

	cfg "github.com/slotopol/server/config"
	"github.com/slotopol/server/game/fishing"
	"github.com/slotopol/server/util"
)

// POST /fishing/room/create - creates a new multiplayer fishing room
func ApiFishRoomCreate(c *gin.Context) {
	var room = fishing.RoomManager.CreateRoom()
	RetOk(c, room)
}

// POST /fishing/room/join - joins an existing fishing room
func ApiFishRoomJoin(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		RoomID  string   `json:"roomId" yaml:"roomId" xml:"roomId,attr" form:"roomId" binding:"required"`
		CID     uint64   `json:"cid" yaml:"cid" xml:"cid,attr" form:"cid" binding:"required"`
		Bet     float64  `json:"bet,omitempty" yaml:"bet,omitempty" xml:"bet,omitempty"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_fish_room_join_nobind, err)
		return
	}

	var room = fishing.RoomManager.GetRoom(arg.RoomID)
	if room == nil {
		Ret404(c, AEC_fish_room_join_noroom, errors.New("room not found"))
		return
	}

	var club *Club
	if club, ok = Clubs.Get(arg.CID); !ok {
		Ret500(c, AEC_fish_room_join_noclub, ErrNoClub)
		return
	}
	_ = club

	var user *User
	var cu = c.MustGet("user").(*User)
	if user, ok = Users.Get(cu.UID); !ok {
		Ret500(c, AEC_fish_room_join_nouser, ErrNoUser)
		return
	}

	var admin, al = MustAdmin(c, arg.CID)
	if admin.UID != user.UID && al&ALdealer == 0 {
		Ret403(c, AEC_fish_room_join_noaccess, ErrNoAccess)
		return
	}

	var bet = arg.Bet
	if bet <= 0 {
		bet = 1
	}

	var wallet = user.GetWallet(arg.CID)
	var player = room.JoinPlayer(user.UID, bet, wallet)

	var ret struct {
		XMLName xml.Name          `json:"-" yaml:"-" xml:"ret"`
		RoomID  string            `json:"roomId" yaml:"roomId" xml:"roomId,attr"`
		Player  *fishing.RoomPlayer `json:"player" yaml:"player" xml:"player"`
	}
	ret.RoomID = room.ID
	ret.Player = player

	RetOk(c, ret)
}

// POST /fishing/room/leave - leaves a fishing room
func ApiFishRoomLeave(c *gin.Context) {
	var err error
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		RoomID  string   `json:"roomId" yaml:"roomId" xml:"roomId,attr" form:"roomId" binding:"required"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_fish_room_leave_nobind, err)
		return
	}

	var room = fishing.RoomManager.GetRoom(arg.RoomID)
	if room == nil {
		Ret404(c, AEC_fish_room_leave_noroom, errors.New("room not found"))
		return
	}

	var cu = c.MustGet("user").(*User)
	room.LeavePlayer(cu.UID)
	Ret204(c)
}

// POST /fishing/room/state - gets current room state (pool + players)
func ApiFishRoomState(c *gin.Context) {
	var err error
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		RoomID  string   `json:"roomId" yaml:"roomId" xml:"roomId,attr" form:"roomId" binding:"required"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_fish_room_state_nobind, err)
		return
	}

	var room = fishing.RoomManager.GetRoom(arg.RoomID)
	if room == nil {
		Ret404(c, AEC_fish_room_state_noroom, errors.New("room not found"))
		return
	}

	var pool, players = room.GetRoomState()

	var ret struct {
		XMLName xml.Name            `json:"-" yaml:"-" xml:"ret"`
		Pool    fishing.Pool        `json:"pool" yaml:"pool" xml:"pool"`
		Players []*fishing.RoomPlayer `json:"players" yaml:"players" xml:"players"`
	}
	ret.Pool = pool
	ret.Players = players

	RetOk(c, ret)
}

// POST /fishing/room/fire - fires at the shared pool in a room
func ApiFishRoomFire(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		RoomID  string   `json:"roomId" yaml:"roomId" xml:"roomId,attr" form:"roomId" binding:"required"`
		CID     uint64   `json:"cid" yaml:"cid" xml:"cid,attr" form:"cid" binding:"required"`
		Bet     float64  `json:"bet,omitempty" yaml:"bet,omitempty" xml:"bet,omitempty"`
		Cannon  *int     `json:"cannon,omitempty" yaml:"cannon,omitempty" xml:"cannon,omitempty"`
		Aim     *int     `json:"aim,omitempty" yaml:"aim,omitempty" xml:"aim,omitempty"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_fish_room_fire_nobind, err)
		return
	}

	var room = fishing.RoomManager.GetRoom(arg.RoomID)
	if room == nil {
		Ret404(c, AEC_fish_room_fire_noroom, errors.New("room not found"))
		return
	}

	var club *Club
	if club, ok = Clubs.Get(arg.CID); !ok {
		Ret500(c, AEC_fish_room_fire_noclub, ErrNoClub)
		return
	}

	var cu = c.MustGet("user").(*User)
	var user *User
	if user, ok = Users.Get(cu.UID); !ok {
		Ret500(c, AEC_fish_room_fire_nouser, ErrNoUser)
		return
	}

	var admin, al = MustAdmin(c, arg.CID)
	if admin.UID != user.UID && al&ALdealer == 0 {
		Ret403(c, AEC_fish_room_fire_noaccess, ErrNoAccess)
		return
	}

	var props *Props
	if props, ok = user.props.Get(arg.CID); !ok {
		Ret500(c, AEC_fish_room_fire_noprops, ErrNoProps)
		return
	}

	var uid = user.UID
	var aim = 0
	var cannon = fishing.Cannon1
	var bet = arg.Bet

	// Get player's current settings from room
	var existingPlayer = room.GetPlayer(uid)
	if existingPlayer != nil {
		if bet <= 0 {
			bet = existingPlayer.Bet
		}
		cannon = existingPlayer.Cannon
		aim = existingPlayer.Aim
	}

	if arg.Cannon != nil {
		cannon = fishing.CannonLevel(*arg.Cannon)
	}
	if arg.Aim != nil {
		aim = *arg.Aim
	}
	if bet <= 0 {
		bet = 1
	}

	var cost = bet * fishing.CannonCost[cannon-1]

	if props.Wallet < cost {
		Ret403(c, AEC_fish_room_fire_nomoney, ErrNoMoney)
		return
	}

	var bank = club.Bank()
	var mrtp = GetRTP(user, club)

	// Fire at the shared pool
	var shotResult *fishing.ShotInRoom
	var attempts = 0
	for {
		attempts++
		if attempts > cfg.Cfg.MaxSpinAttempts {
			Ret500(c, AEC_fish_room_fire_badbank, ErrBadBank)
			return
		}
		shotResult, err = room.FireInRoom(uid, aim, cannon, bet)
		if err != nil {
			Ret404(c, AEC_fish_room_fire_noplayer, err)
			return
		}
		var debit = cost - shotResult.TotalPay
		if bank+debit >= 0 || (bank < 0 && debit > 0) {
			break
		}
	}

	var debit = cost - shotResult.TotalPay

	// Update wallet via bank bat
	if Cfg.ClubUpdateBuffer > 1 {
		go BankBat[arg.CID].Put(cfg.XormStorage, uid, debit)
	} else if err = BankBat[arg.CID].Put(cfg.XormStorage, uid, debit); err != nil {
		Ret500(c, AEC_fish_room_fire_sqlbank, err)
		return
	}

	club.AddBank(debit)
	props.Wallet -= debit

	// Log the spin
	var sid = SpinCounter.Inc()
	var rec = Spinlog{
		SID:    sid,
		GID:    0, // room-based, no scene GID
		MRTP:   mrtp,
		Gain:   shotResult.TotalPay,
		Wallet: props.Wallet,
	}
	var b []byte
	if b, err = json.Marshal(shotResult); err != nil {
		return
	}
	rec.Wins = util.B2S(b)
	if Cfg.UseSpinLog {
		go func() {
			if err = SpinBuf.Put(cfg.XormSpinlog, rec); err != nil {
				log.Printf("can not write to spin log: %s", err.Error())
			}
		}()
	}

	var ret struct {
		XMLName  xml.Name              `json:"-" yaml:"-" xml:"ret"`
		SID      uint64                `json:"sid" yaml:"sid" xml:"sid,attr"`
		Result   *fishing.ShotInRoom   `json:"result" yaml:"result" xml:"result"`
		Wallet   float64               `json:"wallet" yaml:"wallet" xml:"wallet"`
	}
	ret.SID = sid
	ret.Result = shotResult
	ret.Wallet = props.Wallet

	RetOk(c, ret)
}

// Returns bet value.
func ApiFishBetGet(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr" form:"gid" binding:"required"`
	}
	var ret struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"ret"`
		Bet     float64  `json:"bet" yaml:"bet" xml:"bet"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_fish_betget_nobind, err)
		return
	}

	var scene *Scene
	if scene, err = GetScene(arg.GID); err != nil {
		Ret404(c, AEC_fish_betget_noscene, err)
		return
	}
	var game fishing.FishingGame
	if game, ok = scene.Game.(fishing.FishingGame); !ok {
		Ret403(c, AEC_fish_betget_notfish, ErrNotFish)
		return
	}

	var admin, al = MustAdmin(c, scene.CID)
	if admin.UID != scene.UID && al&ALdealer == 0 {
		Ret403(c, AEC_fish_betget_noaccess, ErrNoAccess)
		return
	}

	ret.Bet = game.GetBet()

	RetOk(c, ret)
}

// Set bet value.
func ApiFishBetSet(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr" binding:"required"`
		Bet     float64  `json:"bet" yaml:"bet" xml:"bet" binding:"required"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_fish_betset_nobind, err)
		return
	}

	var scene *Scene
	if scene, err = GetScene(arg.GID); err != nil {
		Ret404(c, AEC_fish_betset_noscene, err)
		return
	}
	var game fishing.FishingGame
	if game, ok = scene.Game.(fishing.FishingGame); !ok {
		Ret403(c, AEC_fish_betset_notfish, ErrNotFish)
		return
	}

	var admin, al = MustAdmin(c, scene.CID)
	if admin.UID != scene.UID && al&ALdealer == 0 {
		Ret403(c, AEC_fish_betset_noaccess, ErrNoAccess)
		return
	}

	if err = game.SetBet(arg.Bet); err != nil {
		Ret403(c, AEC_fish_betset_badbet, err)
		return
	}

	Ret204(c)
}

// Returns current cannon level.
func ApiFishCannonGet(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr" form:"gid" binding:"required"`
	}
	var ret struct {
		XMLName xml.Name           `json:"-" yaml:"-" xml:"ret"`
		Cannon fishing.CannonLevel `json:"cannon" yaml:"cannon" xml:"cannon"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_fish_cannonget_nobind, err)
		return
	}

	var scene *Scene
	if scene, err = GetScene(arg.GID); err != nil {
		Ret404(c, AEC_fish_cannonget_noscene, err)
		return
	}
	var game fishing.FishingGame
	if game, ok = scene.Game.(fishing.FishingGame); !ok {
		Ret403(c, AEC_fish_cannonget_notfish, ErrNotFish)
		return
	}

	var admin, al = MustAdmin(c, scene.CID)
	if admin.UID != scene.UID && al&ALdealer == 0 {
		Ret403(c, AEC_fish_cannonget_noaccess, ErrNoAccess)
		return
	}

	ret.Cannon = game.GetCannon()

	RetOk(c, ret)
}

// Set cannon level.
func ApiFishCannonSet(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name           `json:"-" yaml:"-" xml:"arg"`
		GID     uint64             `json:"gid" yaml:"gid" xml:"gid,attr" binding:"required"`
		Cannon fishing.CannonLevel `json:"cannon" yaml:"cannon" xml:"cannon" binding:"required"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_fish_cannonset_nobind, err)
		return
	}

	var scene *Scene
	if scene, err = GetScene(arg.GID); err != nil {
		Ret404(c, AEC_fish_cannonset_noscene, err)
		return
	}
	var game fishing.FishingGame
	if game, ok = scene.Game.(fishing.FishingGame); !ok {
		Ret403(c, AEC_fish_cannonset_notfish, ErrNotFish)
		return
	}

	var admin, al = MustAdmin(c, scene.CID)
	if admin.UID != scene.UID && al&ALdealer == 0 {
		Ret403(c, AEC_fish_cannonset_noaccess, ErrNoAccess)
		return
	}

	if err = game.SetCannon(arg.Cannon); err != nil {
		Ret403(c, AEC_fish_cannonset_badcannon, err)
		return
	}

	Ret204(c)
}

// Returns current aim position.
func ApiFishAimGet(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr" form:"gid" binding:"required"`
	}
	var ret struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"ret"`
		Aim     int      `json:"aim" yaml:"aim" xml:"aim"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_fish_aimget_nobind, err)
		return
	}

	var scene *Scene
	if scene, err = GetScene(arg.GID); err != nil {
		Ret404(c, AEC_fish_aimget_noscene, err)
		return
	}
	var game fishing.FishingGame
	if game, ok = scene.Game.(fishing.FishingGame); !ok {
		Ret403(c, AEC_fish_aimget_notfish, ErrNotFish)
		return
	}

	var admin, al = MustAdmin(c, scene.CID)
	if admin.UID != scene.UID && al&ALdealer == 0 {
		Ret403(c, AEC_fish_aimget_noaccess, ErrNoAccess)
		return
	}

	ret.Aim = game.GetAim()

	RetOk(c, ret)
}

// Set aim position (which fish in the pool to target).
func ApiFishAimSet(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr" binding:"required"`
		Aim     int      `json:"aim" yaml:"aim" xml:"aim" binding:"required"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_fish_aimset_nobind, err)
		return
	}

	var scene *Scene
	if scene, err = GetScene(arg.GID); err != nil {
		Ret404(c, AEC_fish_aimset_noscene, err)
		return
	}
	var game fishing.FishingGame
	if game, ok = scene.Game.(fishing.FishingGame); !ok {
		Ret403(c, AEC_fish_aimset_notfish, ErrNotFish)
		return
	}

	var admin, al = MustAdmin(c, scene.CID)
	if admin.UID != scene.UID && al&ALdealer == 0 {
		Ret403(c, AEC_fish_aimset_noaccess, ErrNoAccess)
		return
	}

	if err = game.SetAim(arg.Aim); err != nil {
		Ret403(c, AEC_fish_aimset_badaim, err)
		return
	}

	Ret204(c)
}

// Fire a shot at the current aim position.
// Returns pool state, shot result (including chain catches for bomb fish),
// and wallet balance.
func ApiFishFire(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr" form:"gid" binding:"required"`
		Bet     float64  `json:"bet,omitempty" yaml:"bet,omitempty" xml:"bet,omitempty"`
		Cannon  *int     `json:"cannon,omitempty" yaml:"cannon,omitempty" xml:"cannon,omitempty"`
		Aim     *int     `json:"aim,omitempty" yaml:"aim,omitempty" xml:"aim,omitempty"`
	}
	var ret struct {
		XMLName xml.Name             `json:"-" yaml:"-" xml:"ret"`
		SID     uint64               `json:"sid" yaml:"sid" xml:"sid,attr"`
		Game    fishing.FishingGame  `json:"game" yaml:"game" xml:"game"`
		Result  fishing.ShotResult   `json:"result" yaml:"result" xml:"result"`
		Wallet  float64              `json:"wallet" yaml:"wallet" xml:"wallet"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_fish_fire_nobind, err)
		return
	}

	var scene *Scene
	if scene, err = GetScene(arg.GID); err != nil {
		Ret404(c, AEC_fish_fire_noscene, err)
		return
	}
	var game fishing.FishingGame
	if game, ok = scene.Game.(fishing.FishingGame); !ok {
		Ret403(c, AEC_fish_fire_notfish, ErrNotFish)
		return
	}

	var club *Club
	if club, ok = Clubs.Get(scene.CID); !ok {
		Ret500(c, AEC_fish_fire_noclub, ErrNoClub)
		return
	}

	var user *User
	if user, ok = Users.Get(scene.UID); !ok {
		Ret500(c, AEC_fish_fire_nouser, ErrNoUser)
		return
	}

	var admin, al = MustAdmin(c, scene.CID)
	if admin.UID != scene.UID && al&ALdealer == 0 {
		Ret403(c, AEC_fish_fire_noaccess, ErrNoAccess)
		return
	}

	if arg.Bet != 0 {
		if err = game.SetBet(arg.Bet); err != nil {
			Ret403(c, AEC_fish_fire_badbet, err)
			return
		}
	}
	if arg.Cannon != nil {
		if err = game.SetCannon(fishing.CannonLevel(*arg.Cannon)); err != nil {
			Ret403(c, AEC_fish_fire_badcannon, err)
			return
		}
	}
	if arg.Aim != nil {
		if err = game.SetAim(*arg.Aim); err != nil {
			Ret403(c, AEC_fish_fire_badaim, err)
			return
		}
	}

	var cost = game.GetBet() * fishing.CannonCost[game.GetCannon()-1]

	var props *Props
	if props, ok = user.props.Get(scene.CID); !ok {
		Ret500(c, AEC_fish_fire_noprops, ErrNoProps)
		return
	}
	if props.Wallet < cost {
		Ret403(c, AEC_fish_fire_nomoney, ErrNoMoney)
		return
	}

	var bank = club.Bank()
	var mrtp = GetRTP(user, club)

	// Fire until gain less than bank value
	var result fishing.ShotResult
	var totalPay float64
	var debit float64
	var n = 0
	for { // repeat until valid shot and bank can cover it
		n++
		if n > cfg.Cfg.MaxSpinAttempts {
			Ret500(c, AEC_fish_fire_badbank, ErrBadBank)
			return
		}
		game.Spin(mrtp)
		// Clear result and scan
		result = fishing.ShotResult{}
		if game.Scanner(&result) != nil {
			continue
		}
		// Calculate total payout including chain
		totalPay = result.Pay
		for _, ch := range result.Chain {
			totalPay += ch.Pay
		}
		debit = cost - totalPay
		if bank+debit >= 0 || (bank < 0 && debit > 0) {
			break
		}
	}

	// write gain and total bet as transaction
	// debit and totalPay are already set from the loop

	if Cfg.ClubUpdateBuffer > 1 {
		go BankBat[scene.CID].Put(cfg.XormStorage, scene.UID, debit)
	} else if err = BankBat[scene.CID].Put(cfg.XormStorage, scene.UID, debit); err != nil {
		Ret500(c, AEC_fish_fire_sqlbank, err)
		return
	}

	// make changes to memory data
	club.AddBank(debit)
	props.Wallet -= debit

	// write fire result to log and get spin ID
	var sid = SpinCounter.Inc()
	scene.SID = sid
	var rec = Spinlog{
		SID:    sid,
		GID:    arg.GID,
		MRTP:   mrtp,
		Gain:   totalPay,
		Wallet: props.Wallet,
	}
	var b []byte

	if b, err = json.Marshal(scene.Game); err != nil {
		return
	}
	rec.Game = util.B2S(b)

	if b, err = json.Marshal(result); err != nil {
		return
	}
	rec.Wins = util.B2S(b)

	if Cfg.UseSpinLog {
		go func() {
			if err = SpinBuf.Put(cfg.XormSpinlog, rec); err != nil {
				log.Printf("can not write to spin log: %s", err.Error())
			}
		}()
	}

	// prepare result
	ret.SID = sid
	ret.Game = game
	ret.Result = result
	ret.Wallet = props.Wallet

	RetOk(c, ret)
}
