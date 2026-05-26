package fishing

import (
	"encoding/xml"
	"errors"
	"math/rand/v2"
	"sync"
	"time"
)

// RoomPlayer holds per-player state in a fishing room.
type RoomPlayer struct {
	UID    uint64      `json:"uid" yaml:"uid" xml:"uid,attr"`
	Bet    float64     `json:"bet" yaml:"bet" xml:"bet"`
	Cannon CannonLevel `json:"cannon" yaml:"cannon" xml:"cannon"`
	Aim    int         `json:"aim" yaml:"aim" xml:"aim"`
	Wallet float64     `json:"wallet" yaml:"wallet" xml:"wallet"`
}

// FishingRoom represents a multiplayer fishing room with a shared pool.
type FishingRoom struct {
	ID        string                 `json:"id" yaml:"id" xml:"id,attr"`
	Pool      Pool                   `json:"pool" yaml:"pool" xml:"pool"`
	players   map[uint64]*RoomPlayer `json:"-" yaml:"-" xml:"-"` // UID -> player (private, access via methods)
	CreatedAt time.Time              `json:"createdAt" yaml:"createdAt" xml:"createdAt"`
	mu        sync.Mutex
}

// ShotInRoom represents the result of a shot fired in a room.
type ShotInRoom struct {
	XMLName  xml.Name     `json:"-" yaml:"-" xml:"result"`
	Player   *RoomPlayer  `json:"player" yaml:"player" xml:"player"`
	Pool     Pool         `json:"pool" yaml:"pool" xml:"pool"`
	Result   ShotResult   `json:"result" yaml:"result" xml:"result"`
	TotalPay float64      `json:"totalPay" yaml:"totalPay" xml:"totalPay"`
}

// FishingRoomManager manages all active fishing rooms.
type FishingRoomManager struct {
	mu    sync.Mutex
	rooms map[string]*FishingRoom
}

// Global room manager instance.
var RoomManager = &FishingRoomManager{
	rooms: make(map[string]*FishingRoom),
}

// CreateRoom creates a new empty fishing room with a shared pool.
func (mgr *FishingRoomManager) CreateRoom() *FishingRoom {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	var id = generateRoomID()
	// Ensure unique ID
	for {
		if _, ok := mgr.rooms[id]; !ok {
			break
		}
		id = generateRoomID()
	}

	var room = &FishingRoom{
		ID:        id,
		Pool:      make(Pool, PoolSize),
		players:   make(map[uint64]*RoomPlayer),
		CreatedAt: time.Now(),
	}

	// Fill the pool with initial fish
	room.Pool.Fill()

	mgr.rooms[id] = room
	return room
}

// GetRoom returns a room by ID.
func (mgr *FishingRoomManager) GetRoom(id string) *FishingRoom {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	return mgr.rooms[id]
}

// RemoveRoom removes a room by ID.
func (mgr *FishingRoomManager) RemoveRoom(id string) {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	delete(mgr.rooms, id)
}

// JoinRoom adds a player to a room. Returns the room.
func (room *FishingRoom) JoinPlayer(uid uint64, bet float64, wallet float64) *RoomPlayer {
	room.mu.Lock()
	defer room.mu.Unlock()

	var player = &RoomPlayer{
		UID:    uid,
		Bet:    bet,
		Cannon: Cannon1,
		Aim:    0,
		Wallet: wallet,
	}
	room.players[uid] = player
	return player
}

// LeaveRoom removes a player from a room.
func (room *FishingRoom) LeavePlayer(uid uint64) {
	room.mu.Lock()
	defer room.mu.Unlock()
	delete(room.players, uid)

	// Clean up empty rooms
	if len(room.players) == 0 {
		RoomManager.RemoveRoom(room.ID)
	}
}

// FireInRoom fires a shot at the shared pool for a specific player.
// Returns the shot result and updated pool.
func (room *FishingRoom) FireInRoom(uid uint64, aim int, cannon CannonLevel, bet float64) (*ShotInRoom, error) {
	room.mu.Lock()
	defer room.mu.Unlock()

	var player, ok = room.players[uid]
	if !ok {
		return nil, ErrNoPlayer
	}

	// Update player's aim, cannon, bet
	player.Aim = aim
	player.Cannon = cannon
	player.Bet = bet

	// Resolve the shot
	var result ShotResult
	room.resolveShotInPool(player.Aim, player.Cannon, player.Bet, &result)

	// Calculate total payout including chain
	var totalPay = result.Pay
	for _, ch := range result.Chain {
		totalPay += ch.Pay
	}

	// Apply swim away after the shot
	room.Pool.ApplySwimAway()

	return &ShotInRoom{
		Player:   player,
		Pool:     room.Pool,
		Result:   result,
		TotalPay: totalPay,
	}, nil
}

// GetPlayer returns a copy of a player's state by UID, or nil if not found.
func (room *FishingRoom) GetPlayer(uid uint64) *RoomPlayer {
	room.mu.Lock()
	defer room.mu.Unlock()
	if p, ok := room.players[uid]; ok {
		var pc = *p
		return &pc
	}
	return nil
}

// GetRoomState returns the current room state (pool + players).
func (room *FishingRoom) GetRoomState() (Pool, []*RoomPlayer) {
	room.mu.Lock()
	defer room.mu.Unlock()

	// Copy pool
	var pool = make(Pool, len(room.Pool))
	copy(pool, room.Pool)

	// Copy players list
	var players []*RoomPlayer
	for _, p := range room.players {
		var pc = *p
		players = append(players, &pc)
	}

	return pool, players
}

// resolveShotInPool attempts to catch the fish at the given position in the shared pool.
func (room *FishingRoom) resolveShotInPool(pos int, cannon CannonLevel, bet float64, result *ShotResult) {
	result.Pos = pos
	if pos < 0 || pos >= len(room.Pool) {
		result.Hit = false
		result.Catch = false
		result.Pay = 0
		return
	}

	var fish = room.Pool[pos]
	if fish.Type == FishNone {
		result.Hit = false
		result.Catch = false
		result.Pay = 0
		return
	}

	result.Fish = fish.Type

	// Determine catch based on cannon level and fish type
	var catchProb float64
	if cannon >= Cannon1 && cannon <= Cannon5 {
		catchProb = CannonCatch[cannon-1][fish.Type-1]
	} else {
		catchProb = 0
	}

	var caught = rand.Float64() < catchProb
	result.Hit = true
	result.Catch = caught

	if caught {
		result.Pay = FishMult[fish.Type-1] * bet
		room.Pool.ReplaceFish(pos)

		// Octopus (bomb): chain-catch 2 random small fish
		if fish.Type == FishOctopus {
			result.Chain = room.doBombChainInPool(bet)
		}
	} else {
		result.Pay = 0
		// Decrease HP visually on a miss
		if room.Pool[pos].HP > 0 {
			room.Pool[pos].HP--
		}
	}
}

// doBombChainInPool catches 2 random Guppy/Perch fish from the shared pool.
func (room *FishingRoom) doBombChainInPool(bet float64) []ShotResult {
	var chain []ShotResult
	for range 2 {
		var targets []int
		for i, f := range room.Pool {
			if f.Type == FishGuppy || f.Type == FishPerch {
				targets = append(targets, i)
			}
		}
		if len(targets) == 0 {
			break
		}
		var idx = targets[rand.IntN(len(targets))]
		var f = room.Pool[idx]
		var pay = FishMult[f.Type-1] * bet
		chain = append(chain, ShotResult{
			Fish:  f.Type,
			Hit:   true,
			Catch: true,
			Pay:   pay,
			Pos:   idx,
		})
		room.Pool.ReplaceFish(idx)
	}
	return chain
}

// generateRoomID creates a short random room ID.
func generateRoomID() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	var b = make([]byte, 6)
	for i := range b {
		b[i] = chars[rand.IntN(len(chars))]
	}
	return string(b)
}

// ErrNoPlayer is returned when a player is not in a room.
var ErrNoPlayer = errors.New("player not in room")
