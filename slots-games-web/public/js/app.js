// === State ===
const state = {
  token: null,
  user: null,
  uid: null,
  gid: null,
  games: [],
  currentGame: null,
  currentProvider: 'all',
  currentCategory: 'all',
  bet: 1,
  spinning: false,
  activities: [],
  // Game-specific state
  bj: { dealer: [], player: [], bet: 0, phase: 'betting' },
  bac: { betType: 'player', bet: 0, phase: 'betting' },
  vp: { hand: [], held: [], phase: 'betting', payout: 0 },
  dt: { betType: 'dragon', bet: 0, phase: 'betting' },
  rou: { betType: 'red', bet: 0, phase: 'betting' },
  avi: { multiplier: 1.0, phase: 'idle', cashOut: 0 },
  fish: { cannon: 1, aim: 0, phase: 'idle', pool: [] },
  keno: { selected: [], drawn: [], phase: 'betting' },
  adminAccess: false,
  adminData: {
    settings: null,
    payments: [],
    users: [],
    analytics: {}
  }
};

// ===== SYMBOLS & CONSTANTS =====
const SYMBOLS = {
  1: '⭐', 2: '💎', 3: '🍒', 4: '🍋', 5: '🍊',
  6: '🍇', 7: '🔔', 8: '🍀', 9: '👑', 10: '💟',
  11: '🔥', 12: '💫', 13: '🎯', 14: '🌈', 15: '🦁',
};

const CARD_RANKS = ['A','2','3','4','5','6','7','8','9','10','J','Q','K'];
const CARD_SUITS = ['♠','♥','♦','♣'];
const SUIT_COLORS = { '♠': 'black', '♥': 'red', '♦': 'red', '♣': 'black' };
const KENO_NUMBERS = 80;
const KENO_MAX_SELECT = 15;

// ===== API =====
async function api(method, path, body = null) {
  const headers = { 'Content-Type': 'application/json' };
  if (state.token) headers['Authorization'] = `Bearer ${state.token}`;
  const opts = { method, headers };
  if (body) opts.body = JSON.stringify(body);
  const res = await fetch(`/api${path}`, opts);
  const ct = res.headers.get('content-type') || '';
  let data;
  if (ct.includes('json')) {
    data = await res.json();
  } else if (ct.includes('xml')) {
    const text = await res.text();
    try {
      const parser = new DOMParser();
      const xmlDoc = parser.parseFromString(text, 'text/xml');
      data = xmlDoc;
    } catch(e) {
      data = text;
    }
  } else {
    data = await res.text();
  }
  if (!res.ok) throw new Error(data?.what || data?.error || `HTTP ${res.status}`);
  return data;
}

// ===== TOASTS =====
function toast(message, type = 'info') {
  const container = document.getElementById('toast-container');
  const el = document.createElement('div');
  el.className = `toast ${type}`;
  const icons = { success: '✅', error: '❌', info: 'ℹ️' };
  el.innerHTML = `<span>${icons[type] || 'ℹ️'}</span><span>${message}</span>`;
  container.appendChild(el);
  setTimeout(() => { el.style.opacity = '0'; el.style.transition = 'opacity 0.3s'; setTimeout(() => el.remove(), 300); }, 3000);
}

// ===== AUTH =====
async function login() {
  const email = document.getElementById('login-email').value.trim();
  const secret = document.getElementById('login-password').value;
  const btn = document.getElementById('login-btn');
  const errorEl = document.getElementById('login-error');
  btn.disabled = true;
  btn.querySelector('.btn-text').style.display = 'none';
  btn.querySelector('.btn-loader').style.display = 'inline';
  errorEl.textContent = '';
  try {
    const data = await api('POST', '/signin', { email, secret });
    state.token = data.access || data.token;
    state.uid = data.uid;
    state.user = email;
    localStorage.setItem('slotopol_token', state.token);
    localStorage.setItem('slotopol_uid', state.uid);
    localStorage.setItem('slotopol_user', email);
    enterApp();
  } catch (e) {
    errorEl.textContent = e.message || 'Login failed';
  } finally {
    btn.disabled = false;
    btn.querySelector('.btn-text').style.display = 'inline';
    btn.querySelector('.btn-loader').style.display = 'none';
  }
}

async function signup() {
  const email = document.getElementById('signup-email').value.trim();
  const secret = document.getElementById('signup-password').value;
  const btn = document.getElementById('signup-btn');
  const errorEl = document.getElementById('signup-error');
  btn.disabled = true;
  btn.querySelector('.btn-text').style.display = 'none';
  btn.querySelector('.btn-loader').style.display = 'inline';
  errorEl.textContent = '';
  try {
    await api('POST', '/signup', { email, secret, cid: 1 });
    toast('Account created! You can now sign in.', 'success');
    document.getElementById('login-email').value = email;
    document.getElementById('login-password').value = secret;
    showScreen('login');
  } catch (e) {
    errorEl.textContent = e.message || 'Signup failed';
  } finally {
    btn.disabled = false;
    btn.querySelector('.btn-text').style.display = 'inline';
    btn.querySelector('.btn-loader').style.display = 'none';
  }
}

function showSignup() {
  document.getElementById('signup-error').textContent = '';
  showScreen('signup');
}

function logout() {
  state.token = null;
  state.user = null;
  state.uid = null;
  state.gid = null;
  localStorage.removeItem('slotopol_token');
  localStorage.removeItem('slotopol_uid');
  localStorage.removeItem('slotopol_user');
  showScreen('login');
}

// ===== SCREEN MANAGEMENT =====
function showScreen(name) {
  document.querySelectorAll('.screen').forEach(s => s.classList.remove('active'));
  document.getElementById(`${name}-screen`).classList.add('active');
}

function showView(name) {
  document.querySelectorAll('.view').forEach(v => v.classList.remove('active'));
  document.querySelectorAll('.nav-btn, .bnav-btn').forEach(b => b.classList.remove('active'));
  if (document.getElementById(`nav-${name}`)) document.getElementById(`nav-${name}`).classList.add('active');
  if (document.getElementById(`bnav-${name}`)) document.getElementById(`bnav-${name}`).classList.add('active');
  if (document.getElementById(`view-${name}`)) document.getElementById(`view-${name}`).classList.add('active');
}

// ===== APP ENTRY =====
async function enterApp() {
  showScreen('app');
  document.getElementById('user-email').textContent = state.user;
  document.getElementById('profile-name').textContent = state.user;
  document.getElementById('profile-id').textContent = `UID: ${state.uid}`;
  document.getElementById('user-badge').textContent = `#${state.uid}`;    // Check admin access
  try {
    await api('GET', '/admin/analytics');
    state.adminAccess = true;
    document.querySelectorAll('.admin-only').forEach(el => el.style.display = '');
    document.getElementById('nav-admin').style.display = '';
    document.getElementById('bnav-admin').style.display = '';
    await refreshAdminDashboard();
  } catch(e) {
    state.adminAccess = false;
  }
  
  await loadGames();
  await loadAdminSettings();
  await refreshWallet();
  addActivity('🎰', 'Welcome to Slotopol Casino!', '');
}

function refreshWallet() {
  // Try to get wallet info
  if (!state.gid) return;
  api('POST', '/prop/wallet/get', { gid: state.gid, cid: 1 })
    .then(d => updateWallet(d.wallet))
    .catch(() => {});
}

function updateWallet(amount) {
  const el = document.getElementById('wallet-value');
  const el2 = document.getElementById('gameplay-balance');
  if (amount !== undefined && amount !== null) {
    const val = Number(amount).toFixed(2);
    el.textContent = val;
    if (el2) el2.textContent = val;
  }
}

function addActivity(icon, text, amount, cls = '') {
  state.activities.unshift({ icon, text, amount, cls, time: new Date().toLocaleTimeString() });
  renderActivities();
}

function renderActivities() {
  const container = document.getElementById('recent-activity');
  if (!container) return;
  if (state.activities.length === 0) {
    container.innerHTML = '<div class="activity-empty">No recent activity. Start playing!</div>';
    return;
  }
  container.innerHTML = state.activities.slice(0, 10).map(a => `
    <div class="activity-item">
      <span class="act-icon">${a.icon}</span>
      <span class="act-text">${a.text}</span>
      <span class="act-time">${a.time}</span>
      ${a.amount ? `<span class="act-amount ${a.cls}">${a.amount}</span>` : ''}
    </div>
  `).join('');
}

// ===== GAME BROWSER =====
async function loadGames() {
  try {
    const data = await api('GET', '/game/algs');
    state.games = Array.isArray(data) ? data : (data.aliases || []);
    renderCategoryTabs(state.games);
    renderProviderTabs(state.games);
    renderGames(state.games);
    renderHomeCarousel(state.games);
  } catch (e) {
    document.getElementById('games-grid').innerHTML = `<div class="loading">Error loading games: ${e.message}</div>`;
  }
}

function renderCategoryTabs(games) {
  const cats = {};
  games.forEach(alg => {
    (alg.aliases || [alg]).forEach(alias => {
      const name = alias.name || alias;
      const cat = classifyGame(typeof name === 'string' ? name : '');
      cats[cat] = (cats[cat] || 0) + 1;
    });
  });
  const sorted = Object.entries(cats).sort((a, b) => b[1] - a[1]);
  const container = document.getElementById('category-tabs');
  if (!container) return;
  container.innerHTML = '<button class="active" onclick="filterByCategory(\'all\')">🎯 All</button>' +
    sorted.map(([cat, count]) =>
      `<button onclick="filterByCategory('${cat}')" data-cat="${cat}">${getCatIcon(cat)} ${cat} (${count})</button>`
    ).join('');
}

function renderProviderTabs(games) {
  const provs = [...new Set(games.flatMap(g => (g.aliases || [g]).map(a => a.prov || '')))].filter(Boolean).sort();
  const container = document.getElementById('provider-tabs');
  if (!container) return;
  container.innerHTML = provs.map(p =>
    `<button onclick="filterByProvider('${p}')" data-prov="${p}">${p}</button>`
  ).join('');
  // Click first provider to set default
  if (provs.length > 0 && !state.currentProvider || state.currentProvider === 'all') {
    // Keep "All" selected by default
  }
}

function renderGames(games) {
  const grid = document.getElementById('games-grid');
  if (!grid) return;
  const items = [];
  games.forEach(alg => {
    (alg.aliases || [alg]).forEach(alias => {
      items.push({
        prov: alias.prov || '',
        name: alias.name || alias,
        category: classifyGame(alias.name || alias),
        grid: alg.sx && alg.sy ? `${alg.sx}x${alg.sy}` : '',
        lines: alg.ln || 0,
        rtp: alg.rtp ? `${Math.min(...alg.rtp).toFixed(1)}-${Math.max(...alg.rtp).toFixed(1)}%` : '',
      });
    });
  });
  if (items.length === 0) { grid.innerHTML = '<div class="loading">No games found</div>'; return; }
  grid.innerHTML = items.map(gi =>
    `<div class="game-card" data-prov="${gi.prov}" data-name="${gi.name}" data-category="${gi.category}" onclick="selectSlotGame('${gi.prov}','${gi.name.replace(/'/g,"\\'")}')">
      <div class="game-icon">${getCatIcon(gi.category)}</div>
      <div class="game-name">${gi.name}</div>
      <div class="game-prov">${gi.prov} • ${gi.category}</div>
      ${gi.grid ? `<div class="game-meta"><span>${gi.grid}</span>${gi.lines ? `<span>${gi.lines}L</span>` : ''}</div>` : ''}
    </div>`
  ).join('');
}

function renderHomeCarousel(games) {
  const container = document.getElementById('home-carousel');
  if (!container) return;
  const items = [];
  games.forEach(alg => {
    (alg.aliases || [alg]).slice(0, 3).forEach(alias => {
      items.push({ prov: alias.prov || '', name: alias.name || alias });
    });
  });
  const shuffled = items.sort(() => Math.random() - 0.5).slice(0, 12);
  if (shuffled.length === 0) { container.innerHTML = '<div class="loading">No games</div>'; return; }
  container.innerHTML = shuffled.map(gi =>
    `<div class="game-card-mini" onclick="selectSlotGame('${gi.prov}','${gi.name.replace(/'/g,"\\'")}')">
      <div class="gcm-icon">${getCatIcon(classifyGame(gi.name))}</div>
      <div class="gcm-name">${gi.name}</div>
      <div class="gcm-prov">${gi.prov}</div>
    </div>`
  ).join('');
}

function classifyGame(name) {
  const n = (name || '').toLowerCase();
  if (n.includes('keno')) return 'keno';
  if (['fish','dolphin','pearl','ocean','mermaid','kraken','reef','whale'].some(w => n.includes(w))) return 'fishing';
  if (['egypt','pharaoh','pyramid','cleopatra','sphinx'].some(w => n.includes(w)) || n.startsWith('book of')) return 'egypt';
  if (['fruit','cherry','lemon','orange','grape'].some(w => n.includes(w))) return 'fruit';
  if (n.includes('joker')) return 'joker';
  if (['lion','tiger','panda','horse','wolf','bear','panther','eagle'].some(w => n.includes(w))) return 'animals';
  if (['magic','wizard','fairy','unicorn','fantasy'].some(w => n.includes(w))) return 'fantasy';
  if (['fire','burning','flame','hot'].some(w => n.includes(w))) return 'hot';
  if (['lucky','fortune','charm'].some(w => n.includes(w))) return 'lucky';
  if (['viking','nordic','odin'].some(w => n.includes(w))) return 'vikings';
  if (['treasure','gold','diamond','crown'].some(w => n.includes(w))) return 'classic';
  if (['dragon'].some(w => n.includes(w))) return 'fantasy';
  return 'slots';
}

function getCatIcon(cat) {
  const icons = { fishing:'🎣', egypt:'🏛️', fruit:'🍒', joker:'🃏', animals:'🐾', fantasy:'✨', hot:'🔥', lucky:'🍀', vikings:'⚔️', classic:'💎', slots:'🎰', keno:'🎱' };
  return icons[cat] || '🎰';
}

function filterByCategory(cat) {
  state.currentCategory = cat;
  document.querySelectorAll('#category-tabs button').forEach(b => b.classList.toggle('active', b.dataset.cat === cat || (!b.dataset.cat && cat === 'all')));
  filterGames();
}

function filterByProvider(prov) {
  state.currentProvider = prov;
  document.querySelectorAll('#provider-tabs button').forEach(b => b.classList.toggle('active', b.dataset.prov === prov));
  filterGames();
}

function filterGames() {
  const search = (document.getElementById('game-search')?.value || '').toLowerCase();
  const filtered = state.games.filter(alg => {
    return (alg.aliases || [alg]).some(alias => {
      const name = alias.name || alias;
      const cat = classifyGame(typeof name === 'string' ? name : '');
      const mp = state.currentProvider === 'all' || (alias.prov || '').toLowerCase() === state.currentProvider.toLowerCase();
      const mc = state.currentCategory === 'all' || cat === state.currentCategory;
      const ms = !search || name.toLowerCase().includes(search) || (alias.prov || '').toLowerCase().includes(search);
      return mp && mc && ms;
    });
  });
  renderGames(filtered);
}

// ===== SLOT GAME =====
async function selectSlotGame(prov, name) {
  showView('home'); // We'll handle gameplay in a modal/section
  try {
    const gid = await createGameSession(prov, name);
    if (!gid) return;
    
    // Set bet
    await api('POST', '/slot/bet/set', { gid, bet: state.bet });
    
    // Spin
    const data = await api('POST', '/slot/spin', { gid, bet: state.bet });
    updateWallet(data.wallet);
    
    if (data.game && data.game.grid) {
      renderSlotResult(data.game.grid, data.wins);
    }
    
    let gain = 0;
    if (data.wins) {
      if (Array.isArray(data.wins)) gain = data.wins.reduce((s,w) => s + (w.gain||0), 0);
      else if (typeof data.wins === 'object') gain = data.wins.gain || 0;
    }
    
    if (gain > 0) {
      addActivity('🎰', `${name}`, `+${gain.toFixed(2)}`, 'positive');
      toast(`You won ${gain.toFixed(2)}!`, 'success');
    } else {
      addActivity('🎰', `${name}`, `${(-state.bet).toFixed(2)}`, 'negative');
    }
  } catch(e) {
    toast(`Game error: ${e.message}`, 'error');
  }
}

async function createGameSession(prov, name) {
  const sanitize = s => s.toLowerCase().replace(/[^a-z0-9_/]/g, '');
  const aliasId = `${sanitize(prov)}/${sanitize(name)}`;
  try {
    const data = await api('POST', '/game/new', { cid: 1, uid: state.uid, alias: aliasId });
    state.gid = data.gid;
    updateWallet(data.wallet);
    return data.gid;
  } catch(e) {
    toast(`Couldn't create game: ${e.message}`, 'error');
    return null;
  }
}

function renderSlotResult(grid, wins) {
  const container = document.getElementById('reel-container');
  if (!container) return;
  if (!grid || !Array.isArray(grid) || grid.length === 0) return;
  
  container.innerHTML = grid.map((col, ri) =>
    `<div class="reel" id="reel-${ri}">${col.map(sym =>
      `<div class="symbol" data-sym="${sym}">${SYMBOLS[sym] || '❓'}</div>`
    ).join('')}</div>`
  ).join('');
  
  // Add win highlights
  if (wins && Array.isArray(wins)) {
    wins.forEach(w => {
      if (w.reel && w.pos !== undefined) {
        const reel = document.getElementById(`reel-${w.reel}`);
        if (reel) {
          const symbols = reel.querySelectorAll('.symbol');
          if (symbols[w.pos]) symbols[w.pos].classList.add('win');
        }
      }
    });
  }
}

// ===== CASINO GAMES =====
function launchCasinoGame(type) {
  state.currentGame = type;
  state.gid = null; // Clear GID so a new session is created for this game type
  document.getElementById('gameplay-title').textContent = getGameTitle(type);
  showView('gameplay');
  const area = document.getElementById('gameplay-area');
  
  switch(type) {
    case 'blackjack': renderBlackjack(); break;
    case 'baccarat': renderBaccarat(); break;
    case 'poker': renderVideoPoker(); break;
    case 'dragontiger': renderDragonTiger(); break;
    case 'roulette': renderRoulette(); break;
    case 'aviator': renderAviator(); break;
  }
}

function getGameTitle(type) {
  const titles = { blackjack: '🃏 Blackjack', baccarat: '🎴 Baccarat', poker: '♠️ Video Poker', dragontiger: '🐉 Dragon Tiger', roulette: '🎡 Roulette', aviator: '✈️ Aviator' };
  return titles[type] || type;
}

// ===== BLACKJACK =====
function renderBlackjack() {
  state.bj = { dealer: [], player: [], bet: state.bet, phase: 'betting' };
  const area = document.getElementById('gameplay-area');
  area.innerHTML = `
    <div class="bj-table">
      <div class="bj-dealer">
        <div class="bj-label">Dealer's Hand <span id="bj-dealer-value"></span></div>
        <div class="bj-hand" id="bj-dealer-hand"></div>
      </div>
      <div class="bj-player">
        <div class="bj-label">Your Hand <span id="bj-player-value"></span></div>
        <div class="bj-hand" id="bj-player-hand"></div>
      </div>
      <div class="info-item" style="margin:0 auto">
        <span class="info-label">Bet</span>
        <div class="info-control">
          <button class="btn-icon" onclick="adjBjBet(-1)">−</button>
          <span class="info-value" id="bj-bet">${state.bet.toFixed(2)}</span>
          <button class="btn-icon" onclick="adjBjBet(1)">+</button>
        </div>
      </div>
      <div class="bj-actions" id="bj-actions">
        <button class="btn btn-primary" onclick="bjDeal()">🃏 Deal</button>
      </div>
      <div class="bj-result" id="bj-result"></div>
    </div>
  `;
}

function adjBjBet(delta) {
  state.bj.bet = Math.max(0.25, Math.min(25, state.bj.bet + delta));
  document.getElementById('bj-bet').textContent = state.bj.bet.toFixed(2);
}

async function bjDeal() {
  try {
    const alias = getCasinoAlias('blackjack', 'Blackjack');
    const gid = await ensureGameSession(alias);
    if (!gid) return;
    
    await api('POST', '/blackjack/bet/set', { gid, bet: state.bj.bet });
    const data = await api('POST', '/blackjack/deal', { gid });
    updateWallet(data.wallet);
    
    const game = data.game || data;
    state.bj.dealer = game.dealerHand || [];
    state.bj.player = game.playerHand || [];
    state.bj.phase = 'playing';
    
    renderBjHands();
    
    // Check for blackjack or natural
    if (game.result) {
      state.bj.phase = 'done';
      document.getElementById('bj-actions').innerHTML = `<button class="btn btn-primary" onclick="bjDeal()">🃏 Deal Again</button>`;
      document.getElementById('bj-result').textContent = game.result;
      document.getElementById('bj-result').className = `bj-result ${game.result.toLowerCase().includes('win') || game.result.toLowerCase().includes('blackjack') ? 'win' : game.result.toLowerCase().includes('lose') ? 'lose' : 'push'}`;
      if (game.gain > 0) {
        addActivity('🃏', 'Blackjack', `+${game.gain.toFixed(2)}`, 'positive');
        toast(`Blackjack win! ${game.gain.toFixed(2)}`, 'success');
      }
      return;
    }
    
    document.getElementById('bj-actions').innerHTML = `
      <button class="btn btn-primary" onclick="bjHit()">Hit</button>
      <button class="btn btn-secondary" onclick="bjStand()">Stand</button>
      <button class="btn btn-outline" onclick="bjDouble()">Double</button>
    `;
  } catch(e) {
    toast(`Error: ${e.message}`, 'error');
  }
}

async function bjHit() {
  try {
    const data = await api('POST', '/blackjack/hit', { gid: state.gid });
    updateWallet(data.wallet);
    const game = data.game || data;
    state.bj.player = game.playerHand || [];
    renderBjHands();
    
    if (game.result) {
      state.bj.phase = 'done';
      document.getElementById('bj-actions').innerHTML = `<button class="btn btn-primary" onclick="bjDeal()">🃏 Deal Again</button>`;
      document.getElementById('bj-result').textContent = game.result;
      document.getElementById('bj-result').className = `bj-result ${game.result.includes('win') ? 'win' : 'lose'}`;
      if (game.gain > 0) addActivity('🃏', 'Blackjack hit', `+${game.gain.toFixed(2)}`, 'positive');
    }
  } catch(e) { toast(`Error: ${e.message}`, 'error'); }
}

async function bjStand() {
  try {
    const data = await api('POST', '/blackjack/stand', { gid: state.gid });
    updateWallet(data.wallet);
    const game = data.game || data;
    state.bj.dealer = game.dealerHand || [];
    state.bj.player = game.playerHand || [];
    renderBjHands();
    
    state.bj.phase = 'done';
    document.getElementById('bj-actions').innerHTML = `<button class="btn btn-primary" onclick="bjDeal()">🃏 Deal Again</button>`;
    document.getElementById('bj-result').textContent = game.result;
    document.getElementById('bj-result').className = `bj-result ${game.result.includes('win') ? 'win' : 'lose'}`;
    if (game.gain > 0) addActivity('🃏', 'Blackjack stand', `+${game.gain.toFixed(2)}`, 'positive');
  } catch(e) { toast(`Error: ${e.message}`, 'error'); }
}

async function bjDouble() {
  try {
    const data = await api('POST', '/blackjack/double', { gid: state.gid });
    updateWallet(data.wallet);
    const game = data.game || data;
    state.bj.dealer = game.dealerHand || [];
    state.bj.player = game.playerHand || [];
    renderBjHands();
    
    state.bj.phase = 'done';
    document.getElementById('bj-actions').innerHTML = `<button class="btn btn-primary" onclick="bjDeal()">🃏 Deal Again</button>`;
    document.getElementById('bj-result').textContent = game.result;
    document.getElementById('bj-result').className = `bj-result ${game.result.includes('win') ? 'win' : 'lose'}`;
    if (game.gain > 0) addActivity('🃏', 'Blackjack double', `+${game.gain.toFixed(2)}`, 'positive');
  } catch(e) { toast(`Error: ${e.message}`, 'error'); }
}

function renderBjHands() {
  renderBjHand('dealer', state.bj.dealer, state.bj.phase === 'betting' || state.bj.phase === 'playing');
  renderBjHand('player', state.bj.player, false);
  document.getElementById('bj-dealer-value').textContent = state.bj.phase === 'playing' ? '(?)' : `= ${calcHandValue(state.bj.dealer)}`;
  document.getElementById('bj-player-value').textContent = `= ${calcHandValue(state.bj.player)}`;
}

function renderBjHand(who, cards, hideFirst) {
  const container = document.getElementById(`bj-${who}-hand`);
  if (!container) return;
  container.innerHTML = cards.map((c, i) => {
    if (hideFirst && i === 0) return `<div class="bj-card hidden">?</div>`;
    const rank = typeof c === 'object' ? (c.rank || c) : c;
    const suit = typeof c === 'object' ? (c.suit || '') : '';
    const isRed = suit === '♥' || suit === '♦';
    return `<div class="bj-card ${isRed ? 'red' : ''}">${rank}${suit}</div>`;
  }).join('');
}

function calcHandValue(cards) {
  if (!cards || cards.length === 0) return 0;
  let total = 0, aces = 0;
  cards.forEach(c => {
    const rank = typeof c === 'object' ? (c.rank || c) : c;
    const r = String(rank);
    if (r === 'A') { aces++; total += 11; }
    else if (['K','Q','J'].includes(r)) total += 10;
    else total += parseInt(r) || 0;
  });
  while (total > 21 && aces > 0) { total -= 10; aces--; }
  return total;
}

// ===== BACCARAT =====
function renderBaccarat() {
  state.bac = { betType: 'player', bet: state.bet, phase: 'betting' };
  const area = document.getElementById('gameplay-area');
  area.innerHTML = `
    <div class="bac-table">
      <div class="bac-bet-options">
        <button class="bac-bet-btn player active" onclick="selectBacBet('player')">Player <span class="bac-odds">1:1</span></button>
        <button class="bac-bet-btn banker" onclick="selectBacBet('banker')">Banker <span class="bac-odds">0.95:1</span></button>
        <button class="bac-bet-btn tie" onclick="selectBacBet('tie')">Tie <span class="bac-odds">8:1</span></button>
      </div>
      <div class="info-item" style="margin:0 auto 16px">
        <span class="info-label">Bet</span>
        <div class="info-control">
          <button class="btn-icon" onclick="adjBacBet(-1)">−</button>
          <span class="info-value" id="bac-bet">${state.bet.toFixed(2)}</span>
          <button class="btn-icon" onclick="adjBacBet(1)">+</button>
        </div>
      </div>
      <div class="bj-dealer">
        <div class="bj-label">Player Hand</div>
        <div class="bj-hand" id="bac-player-hand"></div>
      </div>
      <div class="bj-player">
        <div class="bj-label">Banker Hand</div>
        <div class="bj-hand" id="bac-banker-hand"></div>
      </div>
      <div class="bj-actions">
        <button class="btn btn-primary" onclick="bacDeal()">🎴 Deal</button>
      </div>
      <div class="bj-result" id="bac-result"></div>
    </div>
  `;
}

function selectBacBet(type) {
  state.bac.betType = type;
  document.querySelectorAll('.bac-bet-btn').forEach(b => b.classList.remove('active'));
  document.querySelector(`.bac-bet-btn.${type}`).classList.add('active');
}

function adjBacBet(d) { state.bac.bet = Math.max(0.25, Math.min(25, state.bac.bet + d)); document.getElementById('bac-bet').textContent = state.bac.bet.toFixed(2); }

async function bacDeal() {
  try {
    const alias = getCasinoAlias('baccarat', 'Baccarat');
    const gid = await ensureGameSession(alias);
    if (!gid) return;
    
    await api('POST', '/baccarat/bet/set', { gid, bet: state.bac.bet });
    const data = await api('POST', '/baccarat/deal', { gid, betTarget: state.bac.betType });
    updateWallet(data.wallet);
    
    const game = data.game || data;
    renderBjHand('player', game.playerHand || [], false);
    renderBjHand('banker', game.bankerHand || [], false);
    
    const result = document.getElementById('bac-result');
    if (game.result) {
      result.textContent = game.result;
      result.className = `bj-result ${game.gain > 0 ? 'win' : 'lose'}`;
      if (game.gain > 0) {
        addActivity('🎴', `Baccarat ${state.bac.betType}`, `+${game.gain.toFixed(2)}`, 'positive');
        toast(`Baccarat win! ${game.gain.toFixed(2)}`, 'success');
      }
    }
  } catch(e) { toast(`Error: ${e.message}`, 'error'); }
}

// ===== VIDEO POKER =====
function renderVideoPoker() {
  state.vp = { hand: [], held: [], phase: 'betting', payout: 0 };
  const area = document.getElementById('gameplay-area');
  area.innerHTML = `
    <div class="bac-table">
      <div class="info-item" style="margin:0 auto 16px">
        <span class="info-label">Bet</span>
        <div class="info-control">
          <button class="btn-icon" onclick="adjVpBet(-1)">−</button>
          <span class="info-value" id="vp-bet">${state.bet.toFixed(2)}</span>
          <button class="btn-icon" onclick="adjVpBet(1)">+</button>
        </div>
      </div>
      <div class="vp-hand" id="vp-hand"></div>
      <div class="bj-actions" id="vp-actions">
        <button class="btn btn-primary" onclick="vpDeal()">♠️ Deal</button>
      </div>
      <div class="bj-result" id="vp-result"></div>
      <div class="vp-paytable" id="vp-paytable"></div>
    </div>
  `;
  renderVpPaytable();
}

function adjVpBet(d) { state.vp.bet = Math.max(0.25, Math.min(5, state.vp.bet + d)); document.getElementById('vp-bet').textContent = state.vp.bet.toFixed(2); }

function renderVpPaytable() {
  const container = document.getElementById('vp-paytable');
  if (!container) return;
  container.innerHTML = `
    <table>
      <tr><th>Hand</th><th>Payout</th></tr>
      <tr><td>Royal Flush</td><td>800</td></tr>
      <tr><td>Straight Flush</td><td>50</td></tr>
      <tr><td>Four of a Kind</td><td>25</td></tr>
      <tr><td>Full House</td><td>9</td></tr>
      <tr><td>Flush</td><td>6</td></tr>
      <tr><td>Straight</td><td>4</td></tr>
      <tr><td>Three of a Kind</td><td>3</td></tr>
      <tr><td>Two Pair</td><td>2</td></tr>
      <tr><td>Jacks or Better</td><td>1</td></tr>
    </table>
  `;
}

function toggleHold(idx) {
  if (state.vp.phase !== 'holding') return;
  const i = state.vp.held.indexOf(idx);
  if (i >= 0) state.vp.held.splice(i, 1);
  else state.vp.held.push(idx);
  renderVpHand();
}

async function vpDeal() {
  try {
    const alias = getCasinoAlias('poker', 'VideoPoker');
    const gid = await ensureGameSession(alias);
    if (!gid) return;
    
    await api('POST', '/poker/bet/set', { gid, bet: state.vp.bet });
    const data = await api('POST', '/poker/deal', { gid });
    updateWallet(data.wallet);
    
    state.vp.hand = data.hand || [];
    state.vp.phase = 'holding';
    state.vp.held = [];
    renderVpHand();
    
    document.getElementById('vp-actions').innerHTML = `<button class="btn btn-primary" onclick="vpDraw()">🎯 Draw</button>`;
    document.getElementById('vp-result').textContent = '';
  } catch(e) { toast(`Error: ${e.message}`, 'error'); }
}

async function vpDraw() {
  try {
    const holdMask = state.vp.held.reduce((m, i) => m | (1 << i), 0);
    const data = await api('POST', '/poker/draw', { gid: state.gid, hold: holdMask });
    updateWallet(data.wallet);
    
    state.vp.hand = data.hand || [];
    state.vp.phase = 'done';
    renderVpHand(false);
    
    const result = document.getElementById('vp-result');
    if (data.payout > 0) {
      result.textContent = `You won ${data.payout.toFixed(2)}!`;
      result.className = 'bj-result win';
      addActivity('♠️', 'Video Poker', `+${data.payout.toFixed(2)}`, 'positive');
      toast(`Video Poker win! ${data.payout.toFixed(2)}`, 'success');
    } else {
      result.textContent = 'No win. Try again!';
      result.className = 'bj-result lose';
    }
    document.getElementById('vp-actions').innerHTML = `<button class="btn btn-primary" onclick="vpDeal()">♠️ Deal Again</button>`;
  } catch(e) { toast(`Error: ${e.message}`, 'error'); }
}

function renderVpHand(canHold = true) {
  const container = document.getElementById('vp-hand');
  if (!container) return;
  container.innerHTML = state.vp.hand.map((c, i) => {
    const held = state.vp.held.includes(i);
    const rank = typeof c === 'object' ? (c.rank || c) : c;
    const suit = typeof c === 'object' ? (c.suit || '') : '';
    const isRed = suit === '♥' || suit === '♦';
    return `<div class="vp-card ${held ? 'held' : ''}" onclick="${canHold ? `toggleHold(${i})` : ''}">
      <div class="vp-card-inner ${isRed ? 'red' : ''}">${rank}${suit}</div>
      <span class="vp-card-held">HELD</span>
    </div>`;
  }).join('');
}

// ===== DRAGON TIGER =====
function renderDragonTiger() {
  state.dt = { betType: 'dragon', bet: state.bet, phase: 'betting' };
  const area = document.getElementById('gameplay-area');
  area.innerHTML = `
    <div class="dt-table">
      <div class="dt-bet-options">
        <button class="dt-bet-btn active" onclick="selectDtBet('dragon')">🐉 Dragon <small>1:1</small></button>
        <button class="dt-bet-btn" onclick="selectDtBet('tiger')">🐯 Tiger <small>1:1</small></button>
        <button class="dt-bet-btn" onclick="selectDtBet('tie')">🤝 Tie <small>8:1</small></button>
      </div>
      <div class="info-item" style="margin:0 auto 16px">
        <span class="info-label">Bet</span>
        <div class="info-control">
          <button class="btn-icon" onclick="adjDtBet(-1)">−</button>
          <span class="info-value" id="dt-bet">${state.bet.toFixed(2)}</span>
          <button class="btn-icon" onclick="adjDtBet(1)">+</button>
        </div>
      </div>
      <div class="dt-cards">
        <div class="dt-side dragon">
          <div class="dt-side-label">🐉 Dragon</div>
          <div class="dt-card" id="dt-dragon-card">?</div>
        </div>
        <div class="dt-side tiger">
          <div class="dt-side-label">🐯 Tiger</div>
          <div class="dt-card" id="dt-tiger-card">?</div>
        </div>
      </div>
      <div class="bj-actions">
        <button class="btn btn-primary" onclick="dtDeal()">🎴 Deal</button>
      </div>
      <div class="bj-result" id="dt-result"></div>
    </div>
  `;
}

function selectDtBet(t) { state.dt.betType = t; document.querySelectorAll('.dt-bet-btn').forEach(b => b.classList.remove('active')); document.querySelectorAll('.dt-bet-btn')[['dragon','tiger','tie'].indexOf(t)].classList.add('active'); }
function adjDtBet(d) { state.dt.bet = Math.max(0.25, Math.min(25, state.dt.bet + d)); document.getElementById('dt-bet').textContent = state.dt.bet.toFixed(2); }

async function dtDeal() {
  try {
    const alias = getCasinoAlias('dragontiger', 'DragonTiger');
    const gid = await ensureGameSession(alias);
    if (!gid) return;
    await api('POST', '/dragontiger/bet/set', { gid, bet: state.dt.bet });
    const data = await api('POST', '/dragontiger/deal', { gid, betTarget: state.dt.betType });
    updateWallet(data.wallet);
    const game = data.game || data;
    document.getElementById('dt-dragon-card').textContent = game.dragonCard || '?';
    document.getElementById('dt-tiger-card').textContent = game.tigerCard || '?';
    const result = document.getElementById('dt-result');
    if (game.result) { result.textContent = game.result; result.className = `bj-result ${game.gain > 0 ? 'win' : 'lose'}`; }
    if (game.gain > 0) { addActivity('🐉', 'Dragon Tiger', `+${game.gain.toFixed(2)}`, 'positive'); toast(`Dragon Tiger win! ${game.gain.toFixed(2)}`, 'success'); }
  } catch(e) { toast(`Error: ${e.message}`, 'error'); }
}

// ===== ROULETTE =====
function renderRoulette() {
  state.rou = { betType: 'red', bet: state.bet, phase: 'betting' };
  const area = document.getElementById('gameplay-area');
  const betTypes = [
    { id: 'red', label: '🔴 Red', class: 'red', payout: '2:1' },
    { id: 'black', label: '⚫ Black', class: 'black', payout: '2:1' },
    { id: 'even', label: '✅ Even', class: 'even', payout: '2:1' },
    { id: 'odd', label: '❌ Odd', class: 'odd', payout: '2:1' },
    { id: '1-18', label: '1-18', class: '', payout: '2:1' },
    { id: '19-36', label: '19-36', class: '', payout: '2:1' },
    { id: '1-12', label: '1st 12', class: '', payout: '3:1' },
    { id: '13-24', label: '2nd 12', class: '', payout: '3:1' },
    { id: '25-36', label: '3rd 12', class: '', payout: '3:1' },
  ];
  area.innerHTML = `
    <div class="roulette-table">
      <div class="roulette-wheel" id="roulette-wheel">
        <div class="roulette-wheel-inner"></div>
      </div>
      <div class="roulette-result" id="roulette-result">?</div>
      <div class="roulette-bet-options" id="roulette-bets">
        ${betTypes.map(bt => `<button class="roulette-bet-chip ${bt.class}" onclick="selectRouBet('${bt.id}')">${bt.label} <small>${bt.payout}</small></button>`).join('')}
      </div>
      <div class="info-item" style="margin:0 auto 16px">
        <span class="info-label">Bet</span>
        <div class="info-control">
          <button class="btn-icon" onclick="adjRouBet(-1)">−</button>
          <span class="info-value" id="rou-bet">${state.bet.toFixed(2)}</span>
          <button class="btn-icon" onclick="adjRouBet(1)">+</button>
        </div>
      </div>
      <div class="bj-actions">
        <button class="btn btn-primary" onclick="rouSpin()">🎡 Spin</button>
      </div>
      <div class="bj-result" id="rou-result"></div>
    </div>
  `;
  document.querySelector(`.roulette-bet-chip[onclick*="'red'"]`)?.classList.add('active');
}

function selectRouBet(t) {
  state.rou.betType = t;
  document.querySelectorAll('.roulette-bet-chip').forEach(b => {
    b.classList.remove('active');
    if (b.getAttribute('onclick') && b.getAttribute('onclick').includes(`'${t}'`)) b.classList.add('active');
  });
}
function adjRouBet(d) { state.rou.bet = Math.max(0.25, Math.min(25, state.rou.bet + d)); document.getElementById('rou-bet').textContent = state.rou.bet.toFixed(2); }

async function rouSpin() {
  try {
    document.getElementById('roulette-wheel').classList.add('spinning');
    const alias = getCasinoAlias('roulette', 'Roulette');
    const gid = await ensureGameSession(alias);
    if (!gid) return;
    await api('POST', '/roulette/bet/set', { gid, bet: state.rou.bet });
    await api('POST', '/roulette/bettype/set', { gid, betType: state.rou.betType });
    const data = await api('POST', '/roulette/spin', { gid });
    await new Promise(r => setTimeout(r, 1000));
    document.getElementById('roulette-wheel').classList.remove('spinning');
    updateWallet(data.wallet);
    const game = data.game || data;
    document.getElementById('roulette-result').textContent = game.number || '?';
    const result = document.getElementById('rou-result');
    if (game.gain > 0) {
      result.textContent = `You won ${game.gain.toFixed(2)}!`;
      result.className = 'bj-result win';
      addActivity('🎡', 'Roulette', `+${game.gain.toFixed(2)}`, 'positive');
      toast(`Roulette win! ${game.gain.toFixed(2)}`, 'success');
    } else {
      result.textContent = 'No win. Try again!';
      result.className = 'bj-result lose';
    }
  } catch(e) { document.getElementById('roulette-wheel').classList.remove('spinning'); toast(`Error: ${e.message}`, 'error'); }
}

// ===== AVIATOR =====
function renderAviator() {
  state.avi = { multiplier: 1.0, phase: 'idle', cashOut: 0 };
  const area = document.getElementById('gameplay-area');
  area.innerHTML = `
    <div class="aviator-game">
      <div class="aviator-screen">
        <div class="aviator-multiplier" id="avi-multiplier">1.00x</div>
        <canvas id="aviator-canvas" class="aviator-canvas"></canvas>
      </div>
      <div class="aviator-info">
        <span class="aviator-stat">Target: <strong id="avi-target">2.00x</strong></span>
        <span class="aviator-stat">Cash Out: <strong id="avi-cashout">0.00</strong></span>
      </div>
      <div class="info-item" style="margin:0 auto 16px">
        <span class="info-label">Bet</span>
        <div class="info-control">
          <button class="btn-icon" onclick="adjAviBet(-1)">−</button>
          <span class="info-value" id="avi-bet">${state.bet.toFixed(2)}</span>
          <button class="btn-icon" onclick="adjAviBet(1)">+</button>
        </div>
      </div>
      <div class="bj-actions" id="avi-actions">
        <button class="btn btn-primary" onclick="aviLaunch()">✈️ Launch</button>
      </div>
      <div class="bj-result" id="avi-result"></div>
    </div>
  `;
}

function adjAviBet(d) { state.avi.bet = Math.max(0.25, Math.min(25, (state.avi.bet || state.bet) + d)); document.getElementById('avi-bet').textContent = (state.avi.bet || state.bet).toFixed(2); }

async function aviLaunch() {
  try {
    const alias = getCasinoAlias('aviator', 'Aviator');
    const gid = await ensureGameSession(alias);
    if (!gid) return;
    await api('POST', '/aviator/bet/set', { gid, bet: state.avi.bet || state.bet });
    await api('POST', '/aviator/launch', { gid });
    state.avi.phase = 'flying';
    state.avi.multiplier = 1.0;
    document.getElementById('avi-actions').innerHTML = `<button class="btn btn-success" onclick="aviCashOut()">💸 Cash Out!</button>`;
    document.getElementById('avi-result').textContent = '';
    
    // Simulate the flight
    state.avi.interval = setInterval(() => {
      state.avi.multiplier += 0.01 + Math.random() * 0.03;
      document.getElementById('avi-multiplier').textContent = state.avi.multiplier.toFixed(2) + 'x';
      renderAviatorGraph();
    }, 100);
    
    // Check server state periodically
    setTimeout(() => checkAviatorState(gid), 2000);
  } catch(e) { toast(`Error: ${e.message}`, 'error'); }
}

async function aviCashOut() {
  try {
    const data = await api('POST', '/aviator/cashout', { gid: state.gid });
    clearInterval(state.avi.interval);
    state.avi.phase = 'idle';
    updateWallet(data.wallet);
    const gain = data.gain || 0;
    document.getElementById('avi-result').textContent = `Cashed out at ${(data.multiplier || 1).toFixed(2)}x for ${gain.toFixed(2)}!`;
    document.getElementById('avi-result').className = 'bj-result win';
    document.getElementById('avi-actions').innerHTML = `<button class="btn btn-primary" onclick="aviLaunch()">✈️ Fly Again</button>`;
    if (gain > 0) { addActivity('✈️', 'Aviator cash out', `+${gain.toFixed(2)}`, 'positive'); toast(`Aviator cash out! ${gain.toFixed(2)}`, 'success'); }
  } catch(e) { toast(`Error: ${e.message}`, 'error'); }
}

async function checkAviatorState(gid) {
  try {
    const data = await api('POST', '/aviator/state', { gid });
    if (data.crashed) {
      clearInterval(state.avi.interval);
      state.avi.phase = 'idle';
      document.getElementById('avi-multiplier').textContent = '💥 CRASHED';
      document.getElementById('avi-multiplier').className = 'aviator-multiplier crashed';
      document.getElementById('avi-actions').innerHTML = `<button class="btn btn-primary" onclick="aviLaunch()">✈️ Try Again</button>`;
      document.getElementById('avi-result').textContent = `Crashed at ${(data.multiplier || 1).toFixed(2)}x`;
      document.getElementById('avi-result').className = 'bj-result lose';
    } else if (state.avi.phase === 'flying') {
      setTimeout(() => checkAviatorState(gid), 2000);
    }
  } catch(e) { /* ignore */ }
}

function renderAviatorGraph() {
  const canvas = document.getElementById('aviator-canvas');
  if (!canvas) return;
  const ctx = canvas.getContext('2d');
  canvas.width = canvas.offsetWidth;
  canvas.height = canvas.offsetHeight;
  ctx.clearRect(0, 0, canvas.width, canvas.height);
  
  // Draw curve
  ctx.beginPath();
  ctx.strokeStyle = '#00d4ff';
  ctx.lineWidth = 3;
  const points = 50;
  for (let i = 0; i < points; i++) {
    const x = (i / points) * canvas.width;
    const y = canvas.height - (Math.pow(i / points, 0.5) * canvas.height * 0.8);
    if (i === 0) ctx.moveTo(x, y); else ctx.lineTo(x, y);
  }
  ctx.stroke();
  
  // Draw plane
  const px = canvas.width * 0.85;
  const py = canvas.height - (Math.pow(0.85, 0.5) * canvas.height * 0.8);
  ctx.font = '24px sans-serif';
  ctx.fillText('✈️', px, py);
}

// ===== HELPER: Create game session for casino games =====
function getCasinoAlias(prov, name) {
  return `${prov}/${name.toLowerCase()}`;
}

async function ensureGameSession(alias) {
  if (state.gid) return state.gid;
  return await createGameSession(alias.split('/')[0], alias.split('/')[1]);
}

// ===== DEPOSIT / WITHDRAWAL =====
function switchPayTab(tab) {
  document.querySelectorAll('.pay-tab').forEach(t => t.classList.remove('active'));
  document.querySelectorAll('.payview').forEach(v => v.classList.remove('active'));
  document.getElementById(`paytab-${tab}`).classList.add('active');
  document.getElementById(`payview-${tab}`).classList.add('active');
}

function selectPayment(el) {
  document.querySelectorAll('#payment-methods .pay-method').forEach(m => m.classList.remove('selected'));
  el.classList.add('selected');
}

function selectWithdrawMethod(el) {
  document.querySelectorAll('#withdraw-methods .pay-method').forEach(m => m.classList.remove('selected'));
  el.classList.add('selected');
}

function setDepositAmount(amt) {
  document.getElementById('deposit-amount').value = amt;
}

function processDeposit() {
  const method = document.querySelector('#payment-methods .pay-method.selected')?.dataset?.method || 'Easypaisa';
  const amount = parseFloat(document.getElementById('deposit-amount').value) || 0;
  const minDep = parseFloat(document.getElementById('min-deposit').textContent.replace(',','')) || 100;
  const maxDep = parseFloat(document.getElementById('max-deposit').textContent.replace(',','')) || 100000;
  
  if (amount < minDep) { showPayResult('deposit', `Minimum deposit is ${minDep} PKR`, true); return; }
  if (amount > maxDep) { showPayResult('deposit', `Maximum deposit is ${maxDep} PKR`, true); return; }
  
  showPayResult('deposit', `Processing ${amount} PKR via ${method}...`, false);
  
  // Try to add wallet via API, or simulate
  if (state.gid && state.token) {
    api('POST', '/prop/wallet/add', { gid: state.gid, cid: 1, sum: amount })
      .then(data => {
        updateWallet(data.wallet);
        showPayResult('deposit', `✅ Successfully deposited ${amount.toLocaleString()} PKR via ${method}!`, false, true);
        addActivity('💳', `Deposit via ${method}`, `+${amount.toFixed(2)}`, 'positive');
        toast(`Deposit of ${amount} PKR successful!`, 'success');
      })
      .catch(() => {
        // Fallback - simulate deposit
        const current = parseFloat(document.getElementById('wallet-value').textContent) || 0;
        updateWallet(current + amount);
        showPayResult('deposit', `✅ Successfully deposited ${amount.toLocaleString()} PKR via ${method}!`, false, true);
        addActivity('💳', `Deposit via ${method}`, `+${amount.toFixed(2)}`, 'positive');
        toast(`Deposit of ${amount} PKR successful!`, 'success');
      });
  } else {
    const current = parseFloat(document.getElementById('wallet-value').textContent) || 0;
    updateWallet(current + amount);
    showPayResult('deposit', `✅ Successfully deposited ${amount.toLocaleString()} PKR via ${method}!`, false, true);
    addActivity('💳', `Deposit via ${method}`, `+${amount.toFixed(2)}`, 'positive');
    toast(`Deposit of ${amount} PKR successful!`, 'success');
  }
}

function processWithdraw() {
  const method = document.querySelector('#withdraw-methods .pay-method.selected')?.dataset?.method || 'Easypaisa';
  const account = document.getElementById('withdraw-account').value.trim();
  const amount = parseFloat(document.getElementById('withdraw-amount').value) || 0;
  const balance = parseFloat(document.getElementById('wallet-value').textContent) || 0;
  const minWd = parseFloat(document.getElementById('min-withdraw').textContent.replace(',','')) || 500;
  const maxWd = parseFloat(document.getElementById('max-withdraw').textContent.replace(',','')) || 50000;
  
  if (!account) { showPayResult('withdraw', 'Please enter your account number', true); return; }
  if (amount < minWd) { showPayResult('withdraw', `Minimum withdrawal is ${minWd} PKR`, true); return; }
  if (amount > maxWd) { showPayResult('withdraw', `Maximum withdrawal is ${maxWd} PKR`, true); return; }
  if (amount > balance) { showPayResult('withdraw', 'Insufficient balance', true); return; }
  
  showPayResult('withdraw', `Processing withdrawal of ${amount} PKR to ${method} (${account})...`, false);
  
  // Simulate withdrawal
  updateWallet(balance - amount);
  showPayResult('withdraw', `✅ Withdrawal of ${amount.toLocaleString()} PKR to ${method} (${account}) initiated!`, false, true);
  addActivity('💸', `Withdrawal via ${method}`, `-${amount.toFixed(2)}`, 'negative');
  toast(`Withdrawal of ${amount} PKR initiated!`, 'success');
}

function showPayResult(tab, msg, isError, isSuccess) {
  const el = document.getElementById(`${tab}-result`);
  if (!el) return;
  el.textContent = msg;
  el.className = 'pay-result';
  if (isError) el.classList.add('error');
  if (isSuccess) el.classList.add('success');
}

// ===== PROFILE =====
async function updateProfile() {
  const name = document.getElementById('profile-display-name').value.trim();
  if (name) {
    try {
      await api('POST', '/user/rename', { uid: state.uid, name });
      toast('Profile updated!', 'success');
    } catch(e) { toast(`Error: ${e.message}`, 'error'); }
  } else {
    toast('Please enter a display name', 'info');
  }
  
  // Load commission
  try {
    const data = await api('POST', '/admin/user/commission/get', { uid: state.uid });
    const rate = data.commission || 5;
    document.getElementById('profile-commission').textContent = rate + '%';
    document.getElementById('profile-comm-value').textContent = rate + '%';
  } catch(e) {
    document.getElementById('profile-commission').textContent = state.adminData?.settings?.defaultCommission + '%' || '5.0%';
  }
}

function copyReferral() {
  const code = document.getElementById('referral-code')?.textContent || 'SLOTOPOL';
  navigator.clipboard?.writeText(code).then(() => toast('Referral code copied!', 'success')).catch(() => toast('Could not copy', 'error'));
}

async function claimBonus() {
  if (!state.gid) {
    toast('Start a game first to claim bonus', 'info');
    return;
  }
  try {
    const data = await api('POST', '/admin/bonus/claim', { gid: state.gid });
    updateWallet(data.wallet);
    toast(`Bonus of ${data.bonus} claimed!`, 'success');
    addActivity('🎁', 'Registration Bonus', `+${data.bonus}`, 'positive');
  } catch(e) {
    toast(`Bonus error: ${e.message}`, 'error');
  }
}

// ===== ADMIN =====
async function loadAdminSettings() {
  try {
    const data = await api('GET', '/admin/settings');
    state.adminData.settings = data;
    // Update UI
    if (data.minDeposit) document.getElementById('min-deposit').textContent = data.minDeposit.toLocaleString();
    if (data.maxDeposit) document.getElementById('max-deposit').textContent = data.maxDeposit.toLocaleString();
    if (data.minWithdrawal) document.getElementById('min-withdraw').textContent = data.minWithdrawal.toLocaleString();
    if (data.maxWithdrawal) document.getElementById('max-withdraw').textContent = data.maxWithdrawal.toLocaleString();
    if (data.winSchedule) document.getElementById('win-schedule').textContent = data.winSchedule;
    if (data.siteName) document.getElementById('adm-site-name').value = data.siteName;
    if (data.welcomeMessage) document.getElementById('adm-welcome-msg').value = data.welcomeMessage;
    if (data.minDeposit) document.getElementById('adm-min-dep').value = data.minDeposit;
    if (data.maxDeposit) document.getElementById('adm-max-dep').value = data.maxDeposit;
    if (data.minWithdrawal) document.getElementById('adm-min-wd').value = data.minWithdrawal;
    if (data.maxWithdrawal) document.getElementById('adm-max-wd').value = data.maxWithdrawal;
    if (data.defaultCommission) document.getElementById('adm-commission').value = data.defaultCommission;
    if (data.winSchedule) document.getElementById('adm-schedule').value = data.winSchedule;
    if (data.registrationBonus !== undefined) document.getElementById('adm-reg-bonus').value = data.registrationBonus;
    if (data.depositBonus !== undefined) document.getElementById('adm-dep-bonus').value = data.depositBonus;
    if (data.registrationBonus) document.getElementById('signup-bonus-amount').textContent = data.registrationBonus;
    
    // Load payments
    loadAdminPayments();
  } catch(e) { /* not admin or settings not loaded */ }
}

function switchAdminTab(tab) {
  document.querySelectorAll('.admin-tab').forEach(t => t.classList.remove('active'));
  document.querySelectorAll('.admin-view').forEach(v => v.classList.remove('active'));
  const tabs = ['dashboard','settings','payments','users','bonuses','upload'];
  const idx = tabs.indexOf(tab);
  document.querySelectorAll('.admin-tab')[idx]?.classList.add('active');
  document.getElementById(`admin-${tab}`)?.classList.add('active');
  if (tab === 'users') loadAdminUsers();
  if (tab === 'dashboard') refreshAdminDashboard();
}

async function refreshAdminDashboard() {
  try {
    const data = await api('GET', '/admin/analytics');
    state.adminData.analytics = data;
    document.getElementById('stat-users').textContent = data.totalUsers || 0;
    document.getElementById('stat-games').textContent = data.totalGames || 0;
    document.getElementById('stat-active').textContent = data.activeGames || 0;
  } catch(e) { /* */ }
}

async function saveAdminSettings() {
  const settings = {
    minDeposit: parseFloat(document.getElementById('adm-min-dep').value) || 100,
    maxDeposit: parseFloat(document.getElementById('adm-max-dep').value) || 100000,
    minWithdrawal: parseFloat(document.getElementById('adm-min-wd').value) || 500,
    maxWithdrawal: parseFloat(document.getElementById('adm-max-wd').value) || 50000,
    defaultCommission: parseFloat(document.getElementById('adm-commission').value) || 5,
    winSchedule: document.getElementById('adm-schedule').value || '08:00-23:00',
    siteName: document.getElementById('adm-site-name').value || 'Slotopol Casino',
    welcomeMessage: document.getElementById('adm-welcome-msg').value || 'Welcome!',
    registrationBonus: parseFloat(document.getElementById('adm-reg-bonus').value) || 0,
    depositBonus: parseFloat(document.getElementById('adm-dep-bonus').value) || 0,
  };
  
  try {
    const data = await api('POST', '/admin/settings', settings);
    state.adminData.settings = data;
    document.getElementById('admin-settings-result').textContent = '✅ Settings saved!';
    document.getElementById('admin-settings-result').className = 'pay-result success';
    toast('Settings saved successfully!', 'success');
    // Update displayed values
    document.getElementById('min-deposit').textContent = settings.minDeposit.toLocaleString();
    document.getElementById('max-deposit').textContent = settings.maxDeposit.toLocaleString();
    document.getElementById('min-withdraw').textContent = settings.minWithdrawal.toLocaleString();
    document.getElementById('max-withdraw').textContent = settings.maxWithdrawal.toLocaleString();
    document.getElementById('win-schedule').textContent = settings.winSchedule;
    document.getElementById('signup-bonus-amount').textContent = settings.registrationBonus;
  } catch(e) {
    document.getElementById('admin-settings-result').textContent = `❌ ${e.message}`;
    document.getElementById('admin-settings-result').className = 'pay-result error';
    toast(`Error: ${e.message}`, 'error');
  }
}

async function loadAdminPayments() {
  try {
    const data = await api('GET', '/admin/payments');
    state.adminData.payments = Array.isArray(data) ? data : [];
    renderAdminPayments();
  } catch(e) { /* */ }
}

function renderAdminPayments() {
  const container = document.getElementById('admin-payment-list');
  if (!container) return;
  container.innerHTML = state.adminData.payments.map(p =>
    `<div class="admin-payment-item">
      <span>${p.logo ? `<img src="${p.logo}" style="width:24px;height:24px;object-fit:contain" onerror="this.outerHTML='💳'">` : '💳'}</span>
      <span class="api-name">${p.name}</span>
      <span class="api-status ${p.active ? 'active' : 'inactive'}">${p.active ? 'Active' : 'Inactive'}</span>
      <button class="btn btn-sm btn-outline" onclick="togglePayment('${p.name}')">Toggle</button>
      <button class="btn btn-sm btn-danger" onclick="removePayment('${p.name}')">✕</button>
    </div>`
  ).join('');
}

async function addPaymentMethod() {
  const name = document.getElementById('adm-new-payment').value.trim();
  if (!name) { toast('Enter a payment method name', 'info'); return; }
  try {
    const data = await api('POST', '/admin/payments/add', { name });
    state.adminData.payments = Array.isArray(data) ? data : [];
    renderAdminPayments();
    document.getElementById('adm-new-payment').value = '';
    toast(`Added ${name}`, 'success');
  } catch(e) { toast(`Error: ${e.message}`, 'error'); }
}

async function togglePayment(name) {
  try {
    const data = await api('POST', '/admin/payments/toggle', { name });
    state.adminData.payments = Array.isArray(data) ? data : [];
    renderAdminPayments();
  } catch(e) { toast(`Error: ${e.message}`, 'error'); }
}

async function removePayment(name) {
  try {
    const data = await api('POST', '/admin/payments/remove', { name });
    state.adminData.payments = Array.isArray(data) ? data : [];
    renderAdminPayments();
    toast(`Removed ${name}`, 'info');
  } catch(e) { toast(`Error: ${e.message}`, 'error'); }
}

async function loadAdminUsers() {
  const container = document.getElementById('admin-user-list');
  if (!container) return;
  container.innerHTML = '<div class="loading">Loading users...</div>';
  try {
    const data = await api('POST', '/admin/users/list');
    const users = data.users || [];
    state.adminData.users = users;
    renderAdminUsers(users);
  } catch(e) {
    container.innerHTML = '<div class="loading">Could not load users</div>';
  }
}

function renderAdminUsers(users) {
  const container = document.getElementById('admin-user-list');
  if (!container) return;
  if (users.length === 0) { container.innerHTML = '<div class="loading">No users found</div>'; return; }
  container.innerHTML = users.map(u =>
    `<div class="admin-user-item">
      <span class="aui-uid">#${u.uid}</span>
      <span class="aui-email">${u.email || 'N/A'}</span>
      <span class="aui-commission">${u.commission || state.adminData.settings?.defaultCommission || 5}%</span>
      <span class="aui-actions">
        <button class="btn btn-sm btn-outline" onclick="blockUser(${u.uid})">${u.status === 0 ? '🔒' : '🔓'}</button>
      </span>
    </div>`
  ).join('');
}

function filterAdminUsers() {
  const search = (document.getElementById('admin-user-search').value || '').toLowerCase();
  const filtered = state.adminData.users.filter(u => 
    (u.email || '').toLowerCase().includes(search) || String(u.uid).includes(search)
  );
  renderAdminUsers(filtered);
}

async function blockUser(uid) {
  try {
    await api('POST', '/admin/user/block', { uid, blocked: true });
    toast(`User #${uid} blocked`, 'info');
    loadAdminUsers();
  } catch(e) { toast(`Error: ${e.message}`, 'error'); }
}

async function setUserCommission() {
  const uid = parseInt(document.getElementById('adm-comm-uid').value);
  const rate = parseFloat(document.getElementById('adm-comm-rate').value);
  if (!uid || !rate) { toast('Enter valid UID and commission rate', 'info'); return; }
  try {
    await api('POST', '/admin/user/commission/set', { uid, commission: rate });
    toast(`Commission set to ${rate}% for user #${uid}`, 'success');
    loadAdminUsers();
  } catch(e) { toast(`Error: ${e.message}`, 'error'); }
}

async function claimBonusForUser() {
  if (!state.gid) { toast('Start a game first', 'info'); return; }
  try {
    const data = await api('POST', '/admin/bonus/claim', { gid: state.gid });
    updateWallet(data.wallet);
    toast(`Bonus of ${data.bonus} claimed!`, 'success');
  } catch(e) { toast(`Error: ${e.message}`, 'error'); }
}

// ===== FILE UPLOAD =====
function handleFileSelect(event) {
  const files = event.target.files;
  uploadFiles(files);
}

function handleDrop(event) {
  event.preventDefault();
  const files = event.dataTransfer.files;
  uploadFiles(files);
}

async function uploadFiles(files) {
  if (!files || files.length === 0) return;
  const progress = document.getElementById('upload-progress');
  const results = document.getElementById('upload-results');
  progress.style.display = 'block';
  progress.textContent = `Uploading ${files.length} file(s)...`;
  results.innerHTML = '';
  
  for (const file of files) {
    const formData = new FormData();
    formData.append('file', file);
    
    try {
      const res = await fetch('/api/admin/upload', {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${state.token}` },
        body: formData
      });
      const data = await res.json();
      if (res.ok) {
        results.innerHTML += `<div class="upload-result-item">
          <img src="${data.url}" alt="${file.name}">
          <div class="uri-name">${file.name}</div>
        </div>`;
        toast(`Uploaded ${file.name}`, 'success');
      } else {
        results.innerHTML += `<div class="upload-result-item">❌ ${file.name}: ${data.what || 'Error'}</div>`;
      }
    } catch(e) {
      results.innerHTML += `<div class="upload-result-item">❌ ${file.name}: ${e.message}</div>`;
    }
  }
  progress.style.display = 'none';
}

// ===== INIT =====
document.addEventListener('DOMContentLoaded', () => {
  const token = localStorage.getItem('slotopol_token');
  const uid = localStorage.getItem('slotopol_uid');
  const user = localStorage.getItem('slotopol_user');
  
  if (token && uid && user) {
    state.token = token;
    state.uid = parseInt(uid);
    state.user = user;
    enterApp();
  }
  
  // Event listeners
  document.getElementById('login-password')?.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') login();
  });
  document.getElementById('login-email')?.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') document.getElementById('login-password')?.focus();
  });

  // Mobile nav sync
  document.querySelectorAll('.bnav-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      const view = btn.id.replace('bnav-', '');
      showView(view);
    });
  });
});
