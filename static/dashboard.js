let activeCards = new Set();
function hideEmptyState() {
  document.getElementById('emptyState').style.display = 'none';
}
function showEmptyState() {
  if (activeCards.size === 0) {
    document.getElementById('emptyState').style.display = 'flex';
  }
}
function createResultCard(id, title, content, isError = false) {
  const resultsContainer = document.querySelector('.results');
  let card = document.getElementById(id);
  if (!card) {
    card = document.createElement('div');
    card.id = id;
    card.className = 'result-card';
    resultsContainer.appendChild(card);
  }
  const timestamp = new Date().toLocaleTimeString();
  card.innerHTML = `
    <div class="result-header">
      <div class="result-title">${title}</div>
      <div class="result-timestamp">${timestamp}</div>
    </div>
    <div class="result-content ${isError ? 'error-message' : ''}">${content}</div>
  `;
  activeCards.add(id);
  hideEmptyState();
}
function formatUserData(data) {
  if (!data || !data.data || data.data.length === 0) {
    return '<div class="error-message">User not found</div>';
  }
  const user = data.data[0];
  const createdDate = user.created_at ? new Date(user.created_at).toLocaleDateString() : 'Unknown';
  return `
    <div class="data-grid">
      <span class="data-label">Display Name:</span>
      <span class="data-value">${user.display_name || 'N/A'}</span>
      <span class="data-label">User ID:</span>
      <span class="data-value">${user.id || 'N/A'}</span>
      <span class="data-label">Account Type:</span>
      <span class="data-value">${user.broadcaster_type || 'Regular User'}</span>
      <span class="data-label">Total Views:</span>
      <span class="data-value">${user.view_count?.toLocaleString() || '0'}</span>
      <span class="data-label">Created:</span>
      <span class="data-value">${createdDate}</span>
      <span class="data-label">Description:</span>
      <span class="data-value">${user.description || 'No description'}</span>
    </div>
  `;
}
function formatStreamData(data) {
  if (!data || !data.data || data.data.length === 0) {
    return `
      <div style="text-align: center; padding: 20px;">
        <div class="status-badge status-offline">Offline</div>
        <div style="margin-top: 8px; color: #7d8590;">Stream is currently
offline</div>
      </div>
    `;
  }
  const stream = data.data[0];
  const startTime = stream.started_at ? new Date(stream.started_at).toLocaleString() : 'Unknown';
  return `
    <div style="text-align: center; margin-bottom: 12px;">
      <div class="status-badge status-live">🔴 Live</div>
    </div>
    <div class="data-grid">
      <span class="data-label">Title:</span>
      <span class="data-value">${stream.title || 'No title'}</span>
      <span class="data-label">Game:</span>
      <span class="data-value">${stream.game_name || 'No category'}</span>
      <span class="data-label">Viewers:</span>
      <span class="data-value">${stream.viewer_count?.toLocaleString() || '0'}</span>
      <span class="data-label">Language:</span>
      <span class="data-value">${stream.language || 'N/A'}</span>
      <span class="data-label">Started:</span>
      <span class="data-value">${startTime}</span>
      <span class="data-label">Mature:</span>
      <span class="data-value">${stream.is_mature ? 'Yes' : 'No'}</span>
    </div>
  `;
}
function formatGamesData(data) {
  if (!data || !data.data || data.data.length === 0) {
    return '<div class="error-message">No games data available</div>';
  }
  const gamesList = data.data.slice(0, 12).map((game, i) => `
    <div class="game-item">
      <span class="game-rank">${i + 1}</span>
      <span class="game-name">${game.name}</span>
    </div>
  `).join('');
  return `<div class="games-list">${gamesList}</div>`;
}
async function getUserInfo() {
  const username = document.getElementById('usernameInput').value.trim();
  const button = event.target;
  if (!username) {
    alert('Please enter a username');
    return;
  }
  button.classList.add('loading');
  try {
    const response = await fetch(`/user/${username}`);
    const data = await response.json();
    createResultCard('userCard', `👤 User: ${username}`, formatUserData(data));
  } catch (error) {
    createResultCard('userCard', `👤 User: ${username}`, `Error: ${error.
message}`, true);
  } finally {
    button.classList.remove('loading');
  }
}
async function getStreamInfo() {
  const username = document.getElementById('streamUsernameInput').value.trim();
  const button = event.target;
  if (!username) {
    alert('Please enter a username');
    return;
  }
  button.classList.add('loading');
  try {
    const response = await fetch(`/stream/${username}`);
    const data = await response.json();
    createResultCard('streamCard', `📺 Stream: ${username}`, formatStreamData(data));
  } catch (error) {
    createResultCard('streamCard', `📺 Stream: ${username}`, `Error: ${error.
message}`, true);
  } finally {
    button.classList.remove('loading');
  }
}
async function getTopGames() {
  const button = event.target;
  button.classList.add('loading');
  try {
    const response = await fetch(`/games/top`);
    const data = await response.json();
    createResultCard('gamesCard', '🎮 Top Games on Twitch', formatGamesData(data));
  } catch (error) {
    createResultCard('gamesCard', '🎮 Top Games on Twitch', `Error: ${error.
message}`, true);
  } finally {
    button.classList.remove('loading');
  }
}
// WebSocket for chat
let ws = new WebSocket(`ws://${window.location.host}/ws`);
let chatConnected = false;
let monitoredChannels = [];
let noticesInitialized = false;

function initNoticesCard() {
  if (document.getElementById('ircNoticesCard')) {
    noticesInitialized = true;
    return;
  }
  const results = document.querySelector('.results');
  const card = document.createElement('div');
  card.id = 'ircNoticesCard';
  card.className = 'result-card';
  card.innerHTML = `
    <div class="result-header">
      <div class="result-title">🔔 Notices</div>
      <div class="result-timestamp" id="ircNoticesCount">0</div>
    </div>
    <div class="result-content">
      <div id="ircNoticesList" style="display:flex;flex-direction:column;gap:8px;max-height:240px;overflow-y:auto;"></div>
      <div style="margin-top:8px;display:flex;gap:6px;">
        <button class="btn btn-secondary" onclick="clearNotices()">Clear</button>
      </div>
    </div>
  `;
  results.prepend(card);
  hideEmptyState();
  noticesInitialized = true;
}

function clearNotices() {
  const list = document.getElementById('ircNoticesList');
  if (list) list.innerHTML = '';
  const count = document.getElementById('ircNoticesCount');
  if (count) count.textContent = '0';
}

function addNoticeEntry(channel, kind, text) {
  if (!noticesInitialized) initNoticesCard();
  const list = document.getElementById('ircNoticesList');
  if (!list) return;
  const when = new Date().toLocaleTimeString();
  const item = document.createElement('div');
  item.style.padding = '8px';
  item.style.background = '#0d1117';
  item.style.border = '1px solid #21262d';
  item.style.borderRadius = '4px';
  item.innerHTML = `
    <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:4px;">
      <div style="display:flex;gap:6px;align-items:center;">
        <span style="display:inline-block;padding:2px 6px;border-radius:10px;background:#21262d;color:#c9d1d9;font-size:0.7rem;text-transform:uppercase;">${kind}</span>
        <span class="data-label">#${channel}</span>
      </div>
      <span style="color:#7d8590;font-size:0.75rem;">${when}</span>
    </div>
    <div class="data-value">${text}</div>
  `;
  // prepend new entries so latest is visible without scrolling
  if (list.firstChild) {
    list.insertBefore(item, list.firstChild);
  } else {
    list.appendChild(item);
  }
  // cap list size
  while (list.children.length > 200) {
    list.removeChild(list.lastChild);
  }
  const count = document.getElementById('ircNoticesCount');
  if (count) count.textContent = String(list.children.length);
}
function updateIrcPrefs() {
  if (!ws || ws.readyState !== WebSocket.OPEN) return;
  const prefs = {
    notice: document.getElementById('prefNotice')?.checked ?? true,
    usernotice: document.getElementById('prefUserNotice')?.checked ?? true,
    clearchat: document.getElementById('prefClearChat')?.checked ?? true,
    roomstate: document.getElementById('prefRoomState')?.checked ?? true,
  };
  ws.send(JSON.stringify({ action: 'setPreferences', prefs }));
  addNoticeEntry('-', 'prefs', `Updated: ${Object.entries(prefs).map(([k,v]) => `${k}:${v?'on':'off'}`).join(', ')}`);
}
ws.onopen = () => {
  console.log('WebSocket connected');
  chatConnected = true;
  updateChatStatus('Connected', true);
  renderMonitoredChannels();
  createResultCard(`ws-connected-${Date.now()}`, '🔌 WebSocket', '<div>Connected to server.</div>');
  // send initial preferences
  setTimeout(updateIrcPrefs, 0);
};
ws.onmessage = (event) => {
  let data;
  try {
    data = JSON.parse(event.data);
  } catch (e) {
    return;
  }
  if (data.user && data.message && data.channel) {
    // Only show messages from monitored channels
    if (!monitoredChannels.includes(data.channel)) return;
    const chatContainer = document.getElementById('ircChatMessages');
    // Clear placeholder
    const placeholder = chatContainer.querySelector('.chat-placeholder');
    if (placeholder) {
      chatContainer.innerHTML = '';
    }
    const messageDiv = document.createElement('div');
    messageDiv.className = 'chat-message';
    messageDiv.innerHTML = `
      <div class="chat-meta">
        <span class="chat-user">${data.user}</span>
        <span class="chat-channel">#${data.channel}</span>
      </div>
      <div class="chat-text">${data.message}</div>
    `;
    chatContainer.appendChild(messageDiv);
    chatContainer.scrollTop = chatContainer.scrollHeight;
    // Keep only last 100 messages
    while (chatContainer.children.length > 100) {
      chatContainer.removeChild(chatContainer.firstChild);
    }
  }
  // Render USERNOTICE and other room notices into results (only for monitored channels)
  if (data.type === 'notice' && data.channel && monitoredChannels.includes(data.channel)) {
    addNoticeEntry(data.channel, 'notice', data.system || 'Notification');
  }
  if (data.type === 'usernotice' && data.channel && monitoredChannels.includes(data.channel)) {
    addNoticeEntry(data.channel, 'usernotice', data.system || 'User notice');
  }
  // Subscription acknowledgement
  if (data.type === 'subscribed' && data.channel) {
    addNoticeEntry(data.channel, 'joined', `Now monitoring #${data.channel}. Room notices will appear here.`);
  }
  if (data.type === 'clearchat' && data.channel && monitoredChannels.includes(data.channel)) {
    addNoticeEntry(data.channel, 'clearchat', 'Chat cleared / moderation action triggered');
  }
  if (data.type === 'roomstate' && data.channel && monitoredChannels.includes(data.channel)) {
    addNoticeEntry(data.channel, 'roomstate', 'Room state changed');
  }
};
ws.onclose = () => {
  console.log('WebSocket disconnected');
  chatConnected = false;
  updateChatStatus('Disconnected', false);
  monitoredChannels = [];
  renderMonitoredChannels();
  createResultCard(`ws-closed-${Date.now()}`, '🔌 WebSocket', '<div class="error-message">Disconnected from server.</div>');
};
ws.onerror = (error) => {
  console.error('WebSocket error:', error);
  chatConnected = false;
  updateChatStatus('Error', false);
  monitoredChannels = [];
  renderMonitoredChannels();
  createResultCard(`ws-error-${Date.now()}`, '🔌 WebSocket', '<div class="error-message">WebSocket error occurred. Check server logs.</div>', true);
};
function updateChatStatus(status, connected) {
  const statusElement = document.getElementById('chatStatus');
  const dot = statusElement.querySelector('.dot');
  const text = statusElement.querySelector('span');
  if (connected) {
    dot.style.background = '#238636';
    text.style.color = '#238636';
  } else {
    dot.style.background = '#7d8590';
    text.style.color = '#7d8590';
  }
  text.textContent = status;
}
function renderMonitoredChannels() {
  const container = document.getElementById('monitoredChannels');
  if (monitoredChannels.length === 0) {
    container.innerHTML = '<span style="color:#7d8590;font-size:0.8rem;">No channels monitored</span>';
  } else {
    container.innerHTML = monitoredChannels.map(channel => `
      <span style="display:inline-block;background:#21262d;color:#58a6ff;padding:2px 8px;border-radius:12px;margin:2px 4px 2px 0;font-size:0.8rem;">
        #${channel}
        <button onclick="unsubscribeFromIRCChat('${channel}')" style="margin-left:6px;background:none;border:none;color:#f85149;cursor:pointer;font-size:0.9em;">✕</button>
      </span>
    `).join('');
  }
  // Update chatSendChannel dropdown
  const select = document.getElementById('chatSendChannel');
  if (select) {
    select.innerHTML = monitoredChannels.map(c => `<option value="${c}">${c}</option>`).join('');
    select.disabled = monitoredChannels.length === 0;
  }
}
function subscribeToIRCChat() {
  const channel = document.getElementById('ircChannelInput').value.trim().toLowerCase();
  const joinBtn = document.getElementById('joinChatBtn');
  if (!channel) {
    alert('Please enter a channel name');
    return;
  }
  if (!chatConnected) {
    alert('WebSocket not connected. Please refresh and try again.');
    return;
  }
  if (monitoredChannels.includes(channel)) {
    alert('Already monitoring this channel.');
    return;
  }
  joinBtn.classList.add('loading');
  ws.send(JSON.stringify({ action: 'subscribe', channel }));
  monitoredChannels.push(channel);
  renderMonitoredChannels();
  joinBtn.classList.remove('loading');
  document.getElementById('ircChannelInput').value = '';
  // Optionally update chat status
  updateChatStatus(`Monitoring ${monitoredChannels.map(c => `#${c}`).join(',')}`, true);
  // Add a notice entry so UI responds immediately
  addNoticeEntry(channel, 'joined', `Now monitoring #${channel}. Room notices will appear here.`);
}
function unsubscribeFromIRCChat(channel) {
  ws.send(JSON.stringify({ action: 'unsubscribe', channel }));
  monitoredChannels = monitoredChannels.filter(c => c !== channel);
  renderMonitoredChannels();
  // If no channels left, reset chat UI
  if (monitoredChannels.length === 0) {
    const chatContainer = document.getElementById('ircChatMessages');
    chatContainer.innerHTML = `
      <div class="chat-placeholder">
        <div class="icon">💬</div>
        <div>Join a channel to monitor chat</div>
      </div>
    `;
    updateChatStatus('Disconnected', false);
  } else {
    updateChatStatus(`Monitoring ${monitoredChannels.map(c => `#${c}`).join(' , ')}`, true);
  }
}
function sendChatMessage(event) {
  event.preventDefault();
  const channel = document.getElementById('chatSendChannel').value;
  const input = document.getElementById('chatSendInput');
  const message = input.value.trim();
  const btn = document.getElementById('chatSendBtn');
  if (!channel) {
    alert('Select a channel to send message');
    return;
  }
  if (!message) {
    return;
  }
  btn.classList.add('loading');
  fetch('/irc/send', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ channel, message })
  })
    .then(res => res.json())
    .then(data => {
      if (!data.success) {
        alert('Failed to send: ' + data.message);
      } else {
        input.value = '';
      }
    })
    .catch(() => alert('Failed to send message'))
    .finally(() => btn.classList.remove('loading'));
}
