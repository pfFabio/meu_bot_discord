// Kenobi Bot — Painel de Controle do Mestre (Dashboard Script)
// Arquitetura moderna em JavaScript Vanilla para controle total e independente do Bot

document.addEventListener('DOMContentLoaded', () => {
  // Estado Global da Aplicação
  const state = {
    currentGuildId: '',
    currentVoiceChannelId: '',
    selectedTextChannelId: '',
    guilds: [],
    audioStatus: null,
    audioLibrary: { temas: [], ambience: [], local: [], presets: [] },
    activeLibCategory: 'temas',
    activeCompCategory: 'spells',
    compendiumData: [],
    diceHistory: [],
    selectedModifier: 0,
    isSuccessMode: false,
    recordingTimer: null,
    recordingSeconds: 0,
    isRecording: false,
    selectedDNDItem: null,
    characters: [],
    selectedCharacter: null,
  };

  // ═══════════════════════════════════════════════════════════════════
  // INICIALIZAÇÃO GERAL
  // ═══════════════════════════════════════════════════════════════════

  initTabs();
  initSSE();
  fetchStatus();
  fetchGuilds();
  fetchAudioLibrary();
  fetchRecordings();
  initAudioControls();
  initDiceControls();
  initCompendium();
  initCharacters();
  initRecorderControls();
  initMasterChat();
  initSystemLogs();
  initKeyboardShortcuts();

  // Polling periódico para manter dados sempre sincronizados
  setInterval(fetchStatus, 3000);
  setInterval(fetchAudioStatus, 1500);

  // ═══════════════════════════════════════════════════════════════════
  // TABS DE NAVEGAÇÃO
  // ═══════════════════════════════════════════════════════════════════
  function initTabs() {
    const navButtons = document.querySelectorAll('.nav-tab');
    const tabPanes = document.querySelectorAll('.tab-pane');

    navButtons.forEach(btn => {
      btn.addEventListener('click', () => {
        const targetId = btn.getAttribute('data-tab');
        navButtons.forEach(b => b.classList.remove('active'));
        tabPanes.forEach(p => p.classList.remove('active'));

        btn.classList.add('active');
        const targetPane = document.getElementById(targetId);
        if (targetPane) {
          targetPane.classList.add('active');
        }
      });
    });
  }

  // ═══════════════════════════════════════════════════════════════════
  // STATUS DO BOT & SERVIDORES (GUILDS)
  // ═══════════════════════════════════════════════════════════════════
  async function fetchStatus() {
    try {
      const res = await fetch('/api/status');
      if (!res.ok) return;
      const data = await res.json();

      const ind = document.getElementById('status-indicator');
      const botName = document.getElementById('bot-name');
      const pingElem = document.getElementById('stat-ping');
      const uptimeElem = document.getElementById('stat-uptime');
      const avatarElem = document.getElementById('bot-avatar');

      if (data.online) {
        ind.className = 'status-indicator online';
        botName.textContent = data.username || 'Kenobi D&D Bot';
        pingElem.textContent = `${data.pingMs} ms`;
        uptimeElem.textContent = formatUptime(data.uptimeSeconds);
        if (data.avatarUrl) avatarElem.src = data.avatarUrl;
      } else {
        ind.className = 'status-indicator offline';
        botName.textContent = 'Kenobi Bot (Conectando...)';
      }
    } catch (_) {}
  }

  function formatUptime(seconds) {
    if (!seconds || seconds <= 0) return '00:00:00';
    const hrs = Math.floor(seconds / 3600);
    const mins = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;
    return `${hrs.toString().padStart(2, '0')}:${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  }

  async function fetchGuilds() {
    try {
      const res = await fetch('/api/guilds');
      if (!res.ok) return;
      state.guilds = await res.json();

      const serverSelect = document.getElementById('server-select');
      serverSelect.innerHTML = '';

      if (state.guilds.length === 0) {
        serverSelect.innerHTML = '<option value="">Nenhum servidor encontrado</option>';
        return;
      }

      state.guilds.forEach((g, idx) => {
        const opt = document.createElement('option');
        opt.value = g.id;
        opt.textContent = g.name;
        if (idx === 0 && !state.currentGuildId) {
          state.currentGuildId = g.id;
        }
        serverSelect.appendChild(opt);
      });

      if (state.currentGuildId) {
        serverSelect.value = state.currentGuildId;
      }

      updateChannelsForGuild(state.currentGuildId);

      serverSelect.addEventListener('change', () => {
        state.currentGuildId = serverSelect.value;
        updateChannelsForGuild(state.currentGuildId);
        fetchAudioStatus();
      });
    } catch (err) {
      console.error('Erro ao buscar servidores:', err);
    }
  }

  function updateChannelsForGuild(guildId) {
    const guild = state.guilds.find(g => g.id === guildId);
    if (!guild) return;

    const voiceSelect = document.getElementById('voice-select');
    const diceChannelSelect = document.getElementById('dice-channel-select');
    const chatChannelSelect = document.getElementById('chat-channel-select');

    // Preenche canais de voz
    voiceSelect.innerHTML = '<option value="">Selecione um canal de voz...</option>';
    if (guild.voiceChannels && guild.voiceChannels.length > 0) {
      guild.voiceChannels.forEach(ch => {
        const opt = document.createElement('option');
        opt.value = ch.id;
        opt.textContent = `🔊 ${ch.name}`;
        if (guild.currentVoiceChannel === ch.id) {
          opt.textContent += ' (Bot conectado)';
          opt.selected = true;
          state.currentVoiceChannelId = ch.id;
        }
        voiceSelect.appendChild(opt);
      });
      if (!state.currentVoiceChannelId && guild.voiceChannels.length > 0) {
        state.currentVoiceChannelId = guild.voiceChannels[0].id;
        voiceSelect.value = state.currentVoiceChannelId;
      }
    }

    voiceSelect.onchange = () => {
      state.currentVoiceChannelId = voiceSelect.value;
    };

    // Preenche canais de texto
    const fillTextChannels = (selectElem) => {
      selectElem.innerHTML = '<option value="">Selecione um canal de texto...</option>';
      if (guild.textChannels && guild.textChannels.length > 0) {
        guild.textChannels.forEach((ch, idx) => {
          const opt = document.createElement('option');
          opt.value = ch.id;
          opt.textContent = `# ${ch.name}`;
          if (idx === 0) {
            opt.selected = true;
            state.selectedTextChannelId = ch.id;
          }
          selectElem.appendChild(opt);
        });
      }
      selectElem.onchange = () => {
        state.selectedTextChannelId = selectElem.value;
      };
    };

    fillTextChannels(diceChannelSelect);
    fillTextChannels(chatChannelSelect);
  }

  // ═══════════════════════════════════════════════════════════════════
  // CONTROLES DE VOZ (CONNECT / DISCONNECT)
  // ═══════════════════════════════════════════════════════════════════
  document.getElementById('btn-voice-connect').addEventListener('click', async () => {
    if (!state.currentGuildId || !state.currentVoiceChannelId) {
      showToast('⚠️ Selecione um servidor e um canal de voz primeiro!');
      return;
    }
    try {
      const res = await fetch('/api/voice/connect', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ guildId: state.currentGuildId, channelId: state.currentVoiceChannelId })
      });
      const data = await res.json();
      if (res.ok) {
        showToast('🔊 Bot conectado ao canal de voz!');
        fetchGuilds();
      } else {
        showToast('❌ Erro: ' + (data.error || 'Falha ao conectar'));
      }
    } catch (e) {
      showToast('❌ Erro de comunicação com o bot');
    }
  });

  document.getElementById('btn-voice-disconnect').addEventListener('click', async () => {
    if (!state.currentGuildId) return;
    try {
      const res = await fetch('/api/voice/disconnect', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ guildId: state.currentGuildId })
      });
      if (res.ok) {
        showToast('🚪 Bot desconectado do canal de voz.');
        fetchGuilds();
      }
    } catch (_) {}
  });

  // ═══════════════════════════════════════════════════════════════════
  // REPRODUÇÃO & MÚSICA (TAB 1)
  // ═══════════════════════════════════════════════════════════════════
  async function fetchAudioStatus() {
    if (!state.currentGuildId) return;
    try {
      const res = await fetch(`/api/audio/status?guildId=${state.currentGuildId}`);
      if (!res.ok) return;
      state.audioStatus = await res.json();
      updatePlayerUI(state.audioStatus);
    } catch (_) {}
  }

  function updatePlayerUI(data) {
    if (!data) return;

    const vinyl = document.getElementById('vinyl-disc');
    const badge = document.getElementById('playback-status-badge');
    const loopBadge = document.getElementById('playback-loop-badge');
    const title = document.getElementById('current-track-title');
    const requester = document.getElementById('current-track-requester');
    const btnPlayPause = document.getElementById('ctrl-play-pause');
    const svgPlay = document.getElementById('svg-play');
    const svgPause = document.getElementById('svg-pause');
    const timeTotal = document.getElementById('time-total');
    const timeCurrent = document.getElementById('time-current');
    const scrubFill = document.getElementById('scrub-fill');
    const volumeSlider = document.getElementById('volume-slider');
    const volumeLabel = document.getElementById('volume-label');
    const loopBtn = document.getElementById('ctrl-loop');

    // Volume
    if (data.volume !== undefined && document.activeElement !== volumeSlider) {
      const volPercent = Math.round(data.volume * 100);
      volumeSlider.value = volPercent;
      volumeLabel.textContent = `${volPercent}%`;
    }

    // Modo de Loop
    loopBtn.className = 'btn-ctrl btn-loop';
    if (data.loopMode === 'track') {
      loopBadge.textContent = 'Loop: Faixa';
      loopBtn.classList.add('active-track');
    } else if (data.loopMode === 'queue') {
      loopBadge.textContent = 'Loop: Fila';
      loopBtn.classList.add('active-queue');
    } else {
      loopBadge.textContent = 'Loop: Off';
    }

    if (data.isPlaying && data.current) {
      title.textContent = data.current.title || 'Faixa de Áudio';
      requester.textContent = `Pedido por: ${data.current.requester || 'Mestre'}`;
      timeTotal.textContent = data.current.duration || '--:--';

      if (data.isPaused) {
        badge.textContent = 'Pausado';
        badge.className = 'badge badge-paused';
        vinyl.className = 'vinyl-disc paused';
        svgPlay.classList.remove('hidden');
        svgPause.classList.add('hidden');
      } else {
        badge.textContent = 'Tocando';
        badge.className = 'badge badge-playing';
        vinyl.className = 'vinyl-disc spinning';
        svgPlay.classList.add('hidden');
        svgPause.classList.remove('hidden');
      }

      // Cálculo de Progresso
      if (data.playbackDurationMs > 0) {
        const elapsedSecs = Math.floor(data.playbackDurationMs / 1000);
        timeCurrent.textContent = formatDurationSeconds(elapsedSecs);

        const totalSecs = parseDurationToSeconds(data.current.duration);
        if (totalSecs > 0) {
          const pct = Math.min(100, (elapsedSecs / totalSecs) * 100);
          scrubFill.style.width = `${pct}%`;
        }
      }
    } else {
      title.textContent = 'Nenhuma música tocando';
      requester.textContent = 'Selecione uma faixa abaixo ou cole um link';
      badge.textContent = 'Parado';
      badge.className = 'badge badge-accent';
      vinyl.className = 'vinyl-disc';
      svgPlay.classList.remove('hidden');
      svgPause.classList.add('hidden');
      timeCurrent.textContent = '00:00';
      timeTotal.textContent = '00:00';
      scrubFill.style.width = '0%';
    }

    // Atualiza Fila
    renderQueue(data.queue || []);
    // Atualiza Histórico
    renderHistory(data.history || []);
  }

  function renderQueue(queue) {
    const container = document.getElementById('queue-items-container');
    const countBadge = document.getElementById('queue-count');
    countBadge.textContent = queue.length;

    if (queue.length === 0) {
      container.innerHTML = '<p class="empty-state">Fila de reprodução vazia.</p>';
      return;
    }

    container.innerHTML = queue.map((t, idx) => `
      <div class="queue-item" data-idx="${idx}">
        <div class="queue-item-header">
          <span class="queue-item-idx">${idx + 1}</span>
          <div class="queue-item-content">
            <div class="queue-item-title" title="${escapeHtml(t.title)}">${escapeHtml(t.title)}</div>
            <div class="queue-item-meta-sub">${escapeHtml(t.duration || 'Local')}</div>
          </div>
        </div>
        <div class="queue-actions">
          ${idx > 0 
            ? `<button class="btn btn-warning btn-xs btn-promote-queue" data-idx="${idx}" title="Promover: Tornar a PRÓXIMA música da fila">⬆️ A Seguir</button>` 
            : `<span class="badge badge-next" title="Próxima faixa a tocar assim que a atual acabar">⏭️ Próxima</span>`
          }
          <button class="btn btn-accent-purple btn-xs btn-jump-queue" data-idx="${idx}" title="Pular imediatamente para esta música com Transição Suave (Fade)">✨ Fade</button>
          <button class="btn btn-ghost-danger btn-xs btn-remove-queue" data-idx="${idx}" title="Remover da fila">✕</button>
        </div>
      </div>
    `).join('');

    container.querySelectorAll('.btn-promote-queue').forEach(btn => {
      btn.addEventListener('click', (e) => {
        e.stopPropagation();
        const idx = btn.getAttribute('data-idx');
        promoteQueueIndex(idx);
      });
    });

    container.querySelectorAll('.btn-jump-queue').forEach(btn => {
      btn.addEventListener('click', (e) => {
        e.stopPropagation();
        const idx = btn.getAttribute('data-idx');
        jumpQueueIndex(idx);
      });
    });

    container.querySelectorAll('.btn-remove-queue').forEach(btn => {
      btn.addEventListener('click', (e) => {
        e.stopPropagation();
        const idx = btn.getAttribute('data-idx');
        removeQueueIndex(idx);
      });
    });
  }

  async function promoteQueueIndex(idx) {
    if (!state.currentGuildId) return;
    try {
      const res = await fetch('/api/audio/queue/promote', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ guildId: state.currentGuildId, index: parseInt(idx, 10) })
      });
      if (res.ok) {
        showToast('⬆️ Música definida como a PRÓXIMA a tocar!');
        fetchAudioStatus();
      }
    } catch (_) {}
  }

  async function jumpQueueIndex(idx) {
    if (!state.currentGuildId) return;
    const isSmooth = document.getElementById('chk-smooth-transition')?.checked ?? true;
    const duration = parseInt(document.getElementById('sel-fade-duration')?.value, 10) || 1500;
    try {
      showToast('✨ Saltando com transição suave...');
      const res = await fetch('/api/audio/queue/jump', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          guildId: state.currentGuildId,
          index: parseInt(idx, 10),
          transition: isSmooth,
          fadeDuration: duration
        })
      });
      if (res.ok) {
        showToast('▶️ Reproduzindo faixa selecionada com transição suave!');
        fetchAudioStatus();
      }
    } catch (_) {}
  }

  function renderHistory(history) {
    const container = document.getElementById('history-items-container');
    if (!history || history.length === 0) {
      container.innerHTML = '<p class="empty-state">Nenhum histórico recente.</p>';
      return;
    }

    const recent = history.slice(-6).reverse();
    container.innerHTML = recent.map(t => `
      <div class="queue-item history-item" data-query="${escapeHtml(t.url || t.filename || t.title)}">
        <span class="queue-item-title" title="${escapeHtml(t.title)}">${escapeHtml(t.title)}</span>
        <button class="btn btn-ghost-cyan btn-xs" title="Tocar novamente">▶️</button>
      </div>
    `).join('');

    container.querySelectorAll('.history-item').forEach(item => {
      item.addEventListener('click', () => {
        const query = item.getAttribute('data-query');
        playAudio(query);
      });
    });
  }

  async function removeQueueIndex(idx) {
    if (!state.currentGuildId) return;
    try {
      await fetch('/api/audio/queue/remove', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ guildId: state.currentGuildId, index: parseInt(idx, 10) })
      });
      fetchAudioStatus();
    } catch (_) {}
  }

  function initAudioControls() {
    // Play / Pause Toggle
    document.getElementById('ctrl-play-pause').addEventListener('click', async () => {
      if (!state.currentGuildId) return;
      const status = state.audioStatus;

      if (status && status.isPlaying) {
        if (status.isPaused) {
          await fetch('/api/audio/resume', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ guildId: state.currentGuildId })
          });
          showToast('▶️ Música retomada');
        } else {
          await fetch('/api/audio/pause', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ guildId: state.currentGuildId })
          });
          showToast('⏸️ Música pausada');
        }
      } else {
        // Toca a primeira da lista ou taverna preset
        playAudio('taverna');
      }
      fetchAudioStatus();
    });

    // Skip (Avançar Música) com suporte a Transição Suave
    document.getElementById('ctrl-skip').addEventListener('click', async () => {
      if (!state.currentGuildId) return;
      const isSmooth = document.getElementById('chk-smooth-transition')?.checked ?? true;
      const duration = parseInt(document.getElementById('sel-fade-duration')?.value, 10) || 1500;
      if (isSmooth) {
        showToast('✨ Avançando com transição suave (fade-out)...');
      }
      try {
        const res = await fetch('/api/audio/skip', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            guildId: state.currentGuildId,
            transition: isSmooth,
            fadeDuration: duration
          })
        });
        const d = await res.json();
        showToast(d.success ? '⏭️ Música avançada com sucesso!' : '⏭️ Fim da fila.');
        fetchAudioStatus();
      } catch (_) {}
    });

    // Previous (Voltar Música)
    document.getElementById('ctrl-prev').addEventListener('click', async () => {
      if (!state.currentGuildId) return;
      try {
        const res = await fetch('/api/audio/previous', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ guildId: state.currentGuildId })
        });
        const d = await res.json();
        showToast(d.track ? `⏮️ Tocando: ${d.track.title}` : '⏮️ Reiniciando faixa.');
        fetchAudioStatus();
      } catch (_) {}
    });

    // Stop (Parar)
    document.getElementById('ctrl-stop').addEventListener('click', async () => {
      if (!state.currentGuildId) return;
      await fetch('/api/audio/stop', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ guildId: state.currentGuildId })
      });
      showToast('⏹️ Reprodução interrompida.');
      fetchAudioStatus();
    });

    // Loop
    document.getElementById('ctrl-loop').addEventListener('click', async () => {
      if (!state.currentGuildId) return;
      const res = await fetch('/api/audio/loop', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ guildId: state.currentGuildId })
      });
      const data = await res.json();
      showToast(`🔁 Loop: ${data.loopMode.toUpperCase()}`);
      fetchAudioStatus();
    });

    // Clear Queue
    document.getElementById('btn-clear-queue').addEventListener('click', async () => {
      if (!state.currentGuildId) return;
      await fetch('/api/audio/queue/clear', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ guildId: state.currentGuildId })
      });
      showToast('🗑️ Fila limpa com sucesso!');
      fetchAudioStatus();
    });

    // Volume Slider
    const volumeSlider = document.getElementById('volume-slider');
    const volumeLabel = document.getElementById('volume-label');
    let volumeDebounce = null;

    volumeSlider.addEventListener('input', () => {
      const val = volumeSlider.value;
      volumeLabel.textContent = `${val}%`;
      clearTimeout(volumeDebounce);
      volumeDebounce = setTimeout(async () => {
        if (!state.currentGuildId) return;
        const volFloat = parseFloat(val) / 100.0;
        await fetch('/api/audio/volume', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ guildId: state.currentGuildId, volume: volFloat })
        });
      }, 150);
    });

    // Mute Button
    document.getElementById('ctrl-mute').addEventListener('click', async () => {
      if (!state.currentGuildId) return;
      const isMuted = volumeSlider.value === '0';
      const newVol = isMuted ? 100 : 0;
      volumeSlider.value = newVol;
      volumeLabel.textContent = `${newVol}%`;
      await fetch('/api/audio/volume', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ guildId: state.currentGuildId, volume: newVol / 100.0 })
      });
    });

    // Universal Play Input
    const playSearchInput = document.getElementById('play-search-input');
    const btnPlayNow = document.getElementById('btn-play-now');
    const btnPlayNext = document.getElementById('btn-play-next');
    const btnPlayFade = document.getElementById('btn-play-fade');
    const btnPlayQueue = document.getElementById('btn-play-queue');

    btnPlayNow.addEventListener('click', () => {
      const q = playSearchInput.value.trim();
      if (!q) return;
      const isSmooth = document.getElementById('chk-smooth-transition')?.checked ?? true;
      playAudio(q, { transition: isSmooth });
      playSearchInput.value = '';
    });

    if (btnPlayNext) {
      btnPlayNext.addEventListener('click', () => {
        const q = playSearchInput.value.trim();
        if (!q) return;
        playAudio(q, { playNext: true });
        playSearchInput.value = '';
      });
    }

    if (btnPlayFade) {
      btnPlayFade.addEventListener('click', () => {
        const q = playSearchInput.value.trim();
        if (!q) return;
        playAudio(q, { transition: true });
        playSearchInput.value = '';
      });
    }

    btnPlayQueue.addEventListener('click', () => {
      const q = playSearchInput.value.trim();
      if (!q) return;
      playAudio(q, { playNext: false, transition: false });
      playSearchInput.value = '';
    });

    playSearchInput.addEventListener('keydown', (e) => {
      if (e.key === 'Enter') {
        const q = playSearchInput.value.trim();
        if (!q) return;
        const isSmooth = document.getElementById('chk-smooth-transition')?.checked ?? true;
        playAudio(q, { transition: isSmooth });
        playSearchInput.value = '';
      }
    });

    // Library Subtabs
    document.querySelectorAll('.lib-tab').forEach(tab => {
      tab.addEventListener('click', () => {
        document.querySelectorAll('.lib-tab').forEach(t => t.classList.remove('active'));
        tab.classList.add('active');
        state.activeLibCategory = tab.getAttribute('data-lib');
        renderLibrary();
      });
    });

    // Library Filter Input
    document.getElementById('lib-filter-input').addEventListener('input', () => {
      renderLibrary();
    });
  }

  async function playAudio(query, options = {}) {
    if (!state.currentGuildId) {
      showToast('⚠️ Selecione um servidor no topo!');
      return;
    }
    if (!state.currentVoiceChannelId) {
      showToast('⚠️ Selecione um canal de voz para o bot entrar!');
      return;
    }

    const isSmooth = options.transition !== undefined 
      ? Boolean(options.transition) 
      : (document.getElementById('chk-smooth-transition')?.checked ?? true);

    const fadeDuration = options.fadeDuration 
      || parseInt(document.getElementById('sel-fade-duration')?.value, 10) 
      || 1500;

    const playNext = Boolean(options.playNext);

    if (playNext) {
      showToast(`⏭️ Definindo como PRÓXIMA música: ${query}...`);
    } else if (isSmooth) {
      showToast(`✨ Iniciando transição suave para: ${query}...`);
    } else {
      showToast(`🎵 Enviando áudio: ${query}...`);
    }

    try {
      const res = await fetch('/api/audio/play', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          guildId: state.currentGuildId,
          channelId: state.currentVoiceChannelId,
          query: query,
          requester: 'Mestre (Painel)',
          playNext: playNext,
          transition: isSmooth && !playNext,
          fadeDuration: fadeDuration
        })
      });

      const data = await res.json();
      if (!res.ok) {
        showToast(`❌ Erro: ${data.error || 'Falha ao reproduzir'}`);
        return;
      }

      if (data.status === 'next') {
        showToast(`⏭️ Definida como a PRÓXIMA música a tocar!`);
      } else if (data.status === 'playing') {
        const transText = data.transition ? ' (com Transição Suave ✨)' : '';
        showToast(`▶️ Tocando: ${data.track ? data.track.title : query}${transText}`);
      } else if (data.status === 'queued') {
        showToast(`➕ Adicionado à fila na posição #${data.position}`);
      } else if (data.status === 'playlist') {
        showToast(`📜 Playlist de ${data.count} faixas iniciada!`);
      }
      fetchAudioStatus();
    } catch (err) {
      showToast('❌ Erro de conexão com o bot');
    }
  }

  async function fetchAudioLibrary() {
    try {
      const res = await fetch('/api/audio/library');
      if (!res.ok) return;
      state.audioLibrary = await res.json();
      renderLibrary();
    } catch (_) {}
  }

  function renderLibrary() {
    const container = document.getElementById('library-cards-container');
    const filter = (document.getElementById('lib-filter-input').value || '').toLowerCase();
    const cat = state.activeLibCategory;

    let items = [];

    if (cat === 'temas') {
      items = (state.audioLibrary.themes || []).map(t => ({
        title: t.title,
        query: t.filename ? 'local:' + t.filename : (t.key || t.title),
        duration: t.duration || 'Local',
        category: t.category || 'Baróvia',
        emoji: t.emoji || '🎭'
      }));
    } else if (cat === 'ambience') {
      items = (state.audioLibrary.ambience || []).map(a => ({
        title: a.title,
        query: a.filename ? 'local:' + a.filename : (a.key || a.title),
        duration: a.duration || 'Local',
        category: a.category || 'Cenário',
        emoji: a.emoji || '🏰'
      }));
    } else if (cat === 'local') {
      items = (state.audioLibrary.local || []).map(m => ({
        title: m.title,
        query: m.filename ? 'local:' + m.filename : m.title,
        duration: m.duration || 'Local',
        category: 'Pasta musicas',
        emoji: m.emoji || '📁'
      }));
    } else if (cat === 'presets') {
      items = (state.audioLibrary.presets || []).map(p => ({
        title: p.title,
        query: p.key,
        duration: 'Stream 24/7',
        category: p.description,
        emoji: '🍻'
      }));
    }

    if (filter) {
      items = items.filter(it => 
        it.title.toLowerCase().includes(filter) || 
        it.category.toLowerCase().includes(filter) ||
        it.query.toLowerCase().includes(filter)
      );
    }

    if (items.length === 0) {
      container.innerHTML = '<p class="empty-state">Nenhuma trilha encontrada para este filtro.</p>';
      return;
    }

    container.innerHTML = items.map(item => `
      <div class="track-card" data-query="${escapeHtml(item.query)}">
        <div class="track-card-header">
          <span class="track-card-icon">${item.emoji}</span>
          <div class="track-card-info">
            <div class="track-card-title" title="${escapeHtml(item.title)}">${escapeHtml(item.title)}</div>
            <div class="track-card-meta">
              <span class="meta-category">${escapeHtml(item.category)}</span>
              <span class="meta-dot">•</span>
              <span class="meta-duration">${escapeHtml(item.duration)}</span>
            </div>
          </div>
        </div>
        <div class="track-card-actions">
          <button class="btn btn-primary btn-xs btn-card-play" data-query="${escapeHtml(item.query)}" title="Tocar Agora">▶️ Tocar</button>
          <button class="btn btn-warning btn-xs btn-card-next" data-query="${escapeHtml(item.query)}" title="Tocar a Seguir (próxima da fila)">⏭️ A Seguir</button>
          <button class="btn btn-accent-purple btn-xs btn-card-fade" data-query="${escapeHtml(item.query)}" title="Transição Suave Imediata (Fade)">✨ Fade</button>
          <button class="btn btn-outline-cyan btn-xs btn-card-queue" data-query="${escapeHtml(item.query)}" title="Adicionar ao final da fila">➕ Fila</button>
        </div>
      </div>
    `).join('');

    container.querySelectorAll('.btn-card-play').forEach(btn => {
      btn.addEventListener('click', (e) => {
        e.stopPropagation();
        const q = btn.getAttribute('data-query');
        playAudio(q, { transition: false });
      });
    });

    container.querySelectorAll('.btn-card-next').forEach(btn => {
      btn.addEventListener('click', (e) => {
        e.stopPropagation();
        const q = btn.getAttribute('data-query');
        playAudio(q, { playNext: true });
      });
    });

    container.querySelectorAll('.btn-card-fade').forEach(btn => {
      btn.addEventListener('click', (e) => {
        e.stopPropagation();
        const q = btn.getAttribute('data-query');
        playAudio(q, { transition: true });
      });
    });

    container.querySelectorAll('.btn-card-queue').forEach(btn => {
      btn.addEventListener('click', (e) => {
        e.stopPropagation();
        const q = btn.getAttribute('data-query');
        playAudio(q, { playNext: false, transition: false });
      });
    });

    container.querySelectorAll('.track-card').forEach(card => {
      card.addEventListener('click', () => {
        const query = card.getAttribute('data-query');
        const isSmooth = document.getElementById('chk-smooth-transition')?.checked ?? true;
        playAudio(query, { transition: isSmooth });
      });
    });
  }

  // ═══════════════════════════════════════════════════════════════════
  // ROLADOR DE DADOS & RPG (TAB 2)
  // ═══════════════════════════════════════════════════════════════════
  function initDiceControls() {
    const d20Elem = document.getElementById('interactive-d20');
    const d20Val = document.getElementById('d20-display-val');

    // Click on Giant D20
    d20Elem.addEventListener('click', () => {
      rollFormula('1d20');
    });

    // Quick Dice buttons
    document.querySelectorAll('.btn-quick-die').forEach(btn => {
      btn.addEventListener('click', () => {
        const die = btn.getAttribute('data-die');
        const mod = state.selectedModifier;
        let formula = `1${die}`;
        if (mod > 0) formula += `+${mod}`;
        else if (mod < 0) formula += `${mod}`;
        rollFormula(formula);
      });
    });

    // Modifiers bar
    document.querySelectorAll('.btn-mod').forEach(btn => {
      btn.addEventListener('click', () => {
        document.querySelectorAll('.btn-mod').forEach(b => b.classList.remove('active'));
        btn.classList.add('active');
        state.selectedModifier = parseInt(btn.getAttribute('data-mod'));
      });
    });

    // Formula chips
    document.querySelectorAll('.btn-formula-chip').forEach(chip => {
      chip.addEventListener('click', () => {
        const formula = chip.getAttribute('data-formula');
        document.getElementById('dice-notation-input').value = formula;
        rollFormula(formula);
      });
    });

    // Mode Switcher (Standard vs Successes)
    document.getElementById('mode-btn-standard').addEventListener('click', () => {
      document.getElementById('mode-btn-standard').classList.add('active');
      document.getElementById('mode-btn-success').classList.remove('active');
      document.getElementById('standard-roll-section').classList.remove('hidden');
      document.getElementById('success-roll-section').classList.add('hidden');
      state.isSuccessMode = false;
    });

    document.getElementById('mode-btn-success').addEventListener('click', () => {
      document.getElementById('mode-btn-success').classList.add('active');
      document.getElementById('mode-btn-standard').classList.remove('active');
      document.getElementById('success-roll-section').classList.remove('hidden');
      document.getElementById('standard-roll-section').classList.add('hidden');
      state.isSuccessMode = true;
    });

    // Button Roll Standard
    document.getElementById('btn-roll-formula').addEventListener('click', () => {
      const formula = document.getElementById('dice-notation-input').value.trim();
      if (formula) rollFormula(formula);
    });

    // Button Roll Successes
    document.getElementById('btn-roll-successes').addEventListener('click', () => {
      const dice = document.getElementById('success-dice-input').value.trim();
      const threshold = parseInt(document.getElementById('success-threshold-input').value) || 5;
      rollSuccesses(dice, threshold);
    });

    // Clear History
    document.getElementById('btn-clear-dice-history').addEventListener('click', () => {
      state.diceHistory = [];
      renderDiceHistory();
    });
  }

  async function rollFormula(notation) {
    const d20Elem = document.getElementById('interactive-d20');
    const d20Val = document.getElementById('d20-display-val');

    d20Elem.classList.add('rolling');
    setTimeout(() => d20Elem.classList.remove('rolling'), 400);

    const sendToDiscord = document.getElementById('chk-send-discord').checked;
    const rerollOn1 = document.getElementById('chk-reroll-1s').checked;
    const channelId = state.selectedTextChannelId;

    try {
      const res = await fetch('/api/dice/roll', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          notation,
          rerollOn1,
          sendToDiscord,
          guildId: state.currentGuildId,
          channelId: channelId
        })
      });

      const data = await res.json();
      if (!res.ok) {
        showToast('❌ Erro na rolagem: ' + (data.error || 'Inválido'));
        return;
      }

      d20Val.textContent = data.total;

      // Atualiza Hero Card
      document.getElementById('last-roll-formula').textContent = data.notation;
      document.getElementById('last-roll-total').textContent = data.total;
      document.getElementById('last-roll-breakdown').textContent = 
        `Dados: [${data.rolls.join(', ')}] ${data.modifier ? (data.modifier > 0 ? '+ ' + data.modifier : data.modifier) : ''}`;

      state.diceHistory.unshift({
        formula: data.notation,
        total: data.total,
        rolls: data.rolls,
        time: new Date().toLocaleTimeString()
      });
      renderDiceHistory();

      if (sendToDiscord) {
        showToast(`🎲 Rolagem enviada para o canal do Discord!`);
      }
    } catch (_) {
      showToast('❌ Falha ao rolar dados');
    }
  }

  async function rollSuccesses(notation, threshold) {
    const sendToDiscord = document.getElementById('chk-send-discord').checked;
    const rerollOn1 = document.getElementById('chk-reroll-1s').checked;
    const channelId = state.selectedTextChannelId;

    try {
      const res = await fetch('/api/dice/roll', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          notation,
          threshold,
          isSuccessMode: true,
          rerollOn1,
          sendToDiscord,
          guildId: state.currentGuildId,
          channelId: channelId
        })
      });

      const data = await res.json();
      if (!res.ok) {
        showToast('❌ ' + (data.error || 'Erro na rolagem'));
        return;
      }

      document.getElementById('last-roll-formula').textContent = `${data.notation} (Alvo ≥ ${threshold})`;
      document.getElementById('last-roll-total').textContent = `${data.successes} Sucessos`;
      document.getElementById('last-roll-breakdown').textContent = 
        `Dados: [${data.rolls.join(', ')}] • Falhas: ${data.failures}`;

      state.diceHistory.unshift({
        formula: `${data.notation} (≥${threshold})`,
        total: `${data.successes} S / ${data.failures} F`,
        rolls: data.rolls,
        time: new Date().toLocaleTimeString()
      });
      renderDiceHistory();

      if (sendToDiscord) {
        showToast('🎯 Rolagem de sucessos enviada para o Discord!');
      }
    } catch (_) {}
  }

  function renderDiceHistory() {
    const container = document.getElementById('dice-history-container');
    if (state.diceHistory.length === 0) {
      container.innerHTML = '<p class="empty-state">Nenhuma rolagem realizada nesta sessão.</p>';
      return;
    }

    container.innerHTML = state.diceHistory.slice(0, 30).map(item => `
      <div class="history-roll-item">
        <div>
          <div class="history-formula">${escapeHtml(item.formula)}</div>
          <div style="font-size: 0.7rem; color: var(--text-dim);">[${item.rolls.join(', ')}] • ${item.time}</div>
        </div>
        <div class="history-total">${item.total}</div>
      </div>
    `).join('');
  }

  // ═══════════════════════════════════════════════════════════════════
  // COMPÊNDIO D&D 2024 (TAB 3)
  // ═══════════════════════════════════════════════════════════════════
  function initCompendium() {
    document.querySelectorAll('.comp-cat-btn').forEach(btn => {
      btn.addEventListener('click', () => {
        document.querySelectorAll('.comp-cat-btn').forEach(b => b.classList.remove('active'));
        btn.classList.add('active');
        state.activeCompCategory = btn.getAttribute('data-cat');
        fetchCompendiumCategory();
      });
    });

    document.getElementById('comp-search-input').addEventListener('input', () => {
      fetchCompendiumCategory();
    });

    fetchCompendiumCategory();
  }

  async function fetchCompendiumCategory() {
    const cat = state.activeCompCategory;
    const query = document.getElementById('comp-search-input').value.trim();
    const titleElem = document.getElementById('compendium-list-title');

    const titles = {
      spells: 'Grimório de Magias (PHB 2024)',
      monsters: 'Bestiário de Monstros & Criaturas',
      classes: 'Classes & Subclasses D&D 2024',
      items: 'Itens Mágicos & Maestria de Armas',
      species: 'Espécies, Linhagens & Ravenloft',
      rules: 'Glossário de Regras & Condições'
    };
    titleElem.textContent = titles[cat] || 'Compêndio D&D';

    try {
      const res = await fetch(`/api/dnd/${cat}?q=${encodeURIComponent(query)}`);
      if (!res.ok) return;
      state.compendiumData = await res.json();
      renderCompendiumList();
    } catch (_) {}
  }

  function renderCompendiumList() {
    const container = document.getElementById('compendium-cards-container');
    const badge = document.getElementById('comp-count-badge');
    const items = state.compendiumData || [];
    badge.textContent = `${items.length} itens`;

    if (items.length === 0) {
      container.innerHTML = '<p class="empty-state">Nenhum item encontrado.</p>';
      return;
    }

    const cat = state.activeCompCategory;

    container.innerHTML = items.slice(0, 100).map((item, idx) => {
      let subtext = '';
      if (cat === 'spells') subtext = `Nível ${item.level === 0 ? 'Truque' : item.level} • ${item.school}`;
      else if (cat === 'monsters') subtext = `ND ${item.cr} • ${item.type}`;
      else if (cat === 'classes') subtext = `Dado Vida: ${item.hit_die || 'd8'}`;
      else if (cat === 'items') subtext = `${item.category} • ${item.rarity || 'Comum'}`;
      else if (cat === 'species') subtext = `${item.category || 'Espécie'}`;
      else if (cat === 'rules') subtext = `${item.category || 'Regra Oficial'}`;

      return `
        <div class="dnd-card" data-idx="${idx}">
          <div class="dnd-card-name">${escapeHtml(item.name || item.title)}</div>
          <div class="dnd-card-badge">${escapeHtml(subtext)}</div>
        </div>
      `;
    }).join('');

    container.querySelectorAll('.dnd-card').forEach(card => {
      card.addEventListener('click', () => {
        const idx = card.getAttribute('data-idx');
        const item = items[idx];
        showCompendiumDetail(item, cat);
      });
    });
  }

  function showCompendiumDetail(item, cat) {
    state.selectedDNDItem = { item, cat };

    const ph = document.getElementById('detail-placeholder');
    const content = document.getElementById('detail-content');

    ph.classList.add('hidden');
    content.classList.remove('hidden');

    let bodyHTML = '';
    const name = item.name || item.title;

    if (cat === 'spells') {
      bodyHTML = `
        <div class="detail-header-block">
          <div class="detail-title">✨ ${escapeHtml(name)}</div>
          <div class="detail-subtitle">Nível ${item.level === 0 ? 'Truque' : item.level} • ${escapeHtml(item.school)} ${item.is_ritual ? '• (Ritual)' : ''} ${item.is_concentration ? '• (Concentração)' : ''}</div>
        </div>
        <div class="detail-body">
          <p><strong>Tempo de Conjuração:</strong> ${escapeHtml(item.casting_time || '1 ação')}</p>
          <p><strong>Alcance:</strong> ${escapeHtml(item.range || 'Pessoal')}</p>
          <p><strong>Componentes:</strong> ${escapeHtml(item.components || 'V, S')}</p>
          <p><strong>Duração:</strong> ${escapeHtml(item.duration || 'Instantânea')}</p>
          <p><strong>Classes:</strong> ${escapeHtml(item.classes ? item.classes.join(', ') : 'Todas')}</p>
          <br>
          <p>${escapeHtml(item.description || '').replace(/\n/g, '<br>')}</p>
        </div>
      `;
    } else if (cat === 'monsters') {
      bodyHTML = `
        <div class="detail-header-block">
          <div class="detail-title">🐉 ${escapeHtml(name)}</div>
          <div class="detail-subtitle">ND ${escapeHtml(item.cr)} • ${escapeHtml(item.size)} ${escapeHtml(item.type)}</div>
        </div>
        <div class="detail-body">
          <p><strong>Classe de Armadura (CA):</strong> ${escapeHtml(item.ac)}</p>
          <p><strong>Pontos de Vida (PV):</strong> ${escapeHtml(item.hp)}</p>
          <p><strong>Deslocamento:</strong> ${escapeHtml(item.speed)}</p>
          <br>
          <p>${escapeHtml(item.description || item.actions || '').replace(/\n/g, '<br>')}</p>
        </div>
      `;
    } else {
      bodyHTML = `
        <div class="detail-header-block">
          <div class="detail-title">${escapeHtml(name)}</div>
          <div class="detail-subtitle">${escapeHtml(item.category || 'Compêndio')}</div>
        </div>
        <div class="detail-body">
          <p>${escapeHtml(item.description || item.text || '').replace(/\n/g, '<br>')}</p>
        </div>
      `;
    }

    bodyHTML += `
      <div class="detail-footer">
        <button id="btn-send-dnd-discord" class="btn btn-primary btn-block">
          <span class="btn-icon">🚀</span> Postar Ficha no Canal do Discord
        </button>
      </div>
    `;

    content.innerHTML = bodyHTML;

    document.getElementById('btn-send-dnd-discord').addEventListener('click', async () => {
      if (!state.selectedTextChannelId) {
        showToast('⚠️ Selecione um canal de texto de destino no topo!');
        return;
      }

      let typeSlug = cat;
      if (cat === 'spells') typeSlug = 'spell';
      else if (cat === 'monsters') typeSlug = 'monster';
      else if (cat === 'classes') typeSlug = 'class';
      else if (cat === 'items') typeSlug = 'item';
      else if (cat === 'rules') typeSlug = 'rule';
      else if (cat === 'species') typeSlug = 'species';

      try {
        const res = await fetch('/api/dnd/send', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            type: typeSlug,
            name: name,
            channelId: state.selectedTextChannelId
          })
        });

        if (res.ok) {
          showToast(`✅ Ficha de "${name}" enviada para o canal do Discord!`);
        } else {
          showToast('❌ Falha ao enviar para o Discord');
        }
      } catch (_) {}
    });
  }

  // ═══════════════════════════════════════════════════════════════════
  // FICHAS DE PERSONAGEM (TAB: CHARACTERS)
  // ═══════════════════════════════════════════════════════════════════
  function initCharacters() {
    fetchCharacters();

    const searchInput = document.getElementById('character-search-input');
    if (searchInput) {
      searchInput.addEventListener('input', () => {
        renderCharactersList();
      });
    }

    // Controles do Modal de Edição de Ficha
    document.getElementById('btn-close-edit-modal')?.addEventListener('click', closeEditModal);
    document.getElementById('btn-cancel-edit-char')?.addEventListener('click', closeEditModal);

    document.getElementById('btn-save-edit-char')?.addEventListener('click', async () => {
      const charId = document.getElementById('edit-char-id').value;
      if (!charId) return;

      const payload = {
        characterId: charId,
        userId: (document.getElementById('edit-char-userid')?.value || '').trim(),
        name: document.getElementById('edit-char-name').value.trim(),
        race: document.getElementById('edit-char-race').value.trim(),
        class: document.getElementById('edit-char-class').value.trim(),
        subclass: document.getElementById('edit-char-subclass').value.trim(),
        level: parseInt(document.getElementById('edit-char-level').value, 10) || 1,
        armor_class: parseInt(document.getElementById('edit-char-ac').value, 10) || 10,
        speed: document.getElementById('edit-char-speed').value.trim(),
        max_hp: parseInt(document.getElementById('edit-char-hp-max').value, 10) || 10,
        current_hp: parseInt(document.getElementById('edit-char-hp-curr').value, 10) || 0,
        temp_hp: parseInt(document.getElementById('edit-char-hp-temp').value, 10) || 0,
        str: parseInt(document.getElementById('edit-char-str').value, 10) || 10,
        dex: parseInt(document.getElementById('edit-char-dex').value, 10) || 10,
        con: parseInt(document.getElementById('edit-char-con').value, 10) || 10,
        int: parseInt(document.getElementById('edit-char-int').value, 10) || 10,
        wis: parseInt(document.getElementById('edit-char-wis').value, 10) || 10,
        cha: parseInt(document.getElementById('edit-char-cha').value, 10) || 10,
        gp: parseInt(document.getElementById('edit-char-gp').value, 10) || 0,
        sp: parseInt(document.getElementById('edit-char-sp').value, 10) || 0,
        cp: parseInt(document.getElementById('edit-char-cp').value, 10) || 0,
      };

      try {
        const res = await fetch('/api/character/update', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload)
        });
        const data = await res.json();
        if (res.ok && data.character) {
          showToast(`💾 Ficha de "${data.character.name}" atualizada com sucesso!`);
          closeEditModal();
          // Atualiza lista local
          const idx = state.characters.findIndex(c => (c.id === charId || c.name === charId));
          if (idx !== -1) {
            state.characters[idx] = data.character;
          }
          state.selectedCharacter = data.character;
          renderCharactersList();
          renderCharacterSheet(data.character);
        } else {
          showToast(`❌ Falha ao salvar: ${data.error || 'Erro desconhecido'}`);
        }
      } catch (err) {
        showToast('❌ Erro de conexão ao salvar alterações da ficha');
      }
    });
  }

  function openEditModal(c) {
    const modal = document.getElementById('modal-edit-character');
    if (!modal) return;
    document.getElementById('edit-char-id').value = c.id || c.name;
    document.getElementById('edit-char-userid').value = c.user_id || '';
    document.getElementById('edit-char-name').value = c.name || '';
    document.getElementById('edit-char-race').value = c.race || '';
    document.getElementById('edit-char-class').value = c.class || '';
    document.getElementById('edit-char-subclass').value = c.subclass || '';
    document.getElementById('edit-char-level').value = c.level || 1;
    document.getElementById('edit-char-ac').value = c.armor_class || 10;
    document.getElementById('edit-char-speed').value = c.speed || '';
    document.getElementById('edit-char-hp-max').value = c.max_hp || 10;
    document.getElementById('edit-char-hp-curr').value = c.current_hp || 10;
    document.getElementById('edit-char-hp-temp').value = c.temp_hp || 0;

    const attrs = c.attributes || {};
    document.getElementById('edit-char-str').value = attrs.str?.score || 10;
    document.getElementById('edit-char-dex').value = attrs.dex?.score || 10;
    document.getElementById('edit-char-con').value = attrs.con?.score || 10;
    document.getElementById('edit-char-int').value = attrs.int?.score || 10;
    document.getElementById('edit-char-wis').value = attrs.wis?.score || 10;
    document.getElementById('edit-char-cha').value = attrs.cha?.score || 10;

    const coins = c.coins || {};
    document.getElementById('edit-char-gp').value = coins.gp || 0;
    document.getElementById('edit-char-sp').value = coins.sp || 0;
    document.getElementById('edit-char-cp').value = coins.cp || 0;

    modal.classList.remove('hidden');
  }

  function closeEditModal() {
    const modal = document.getElementById('modal-edit-character');
    if (modal) modal.classList.add('hidden');
  }

  async function fetchCharacters() {
    try {
      const res = await fetch('/api/characters');
      if (!res.ok) return;
      const data = await res.json();
      state.characters = data.characters || [];
      const badge = document.getElementById('characters-count-badge');
      if (badge) {
        badge.textContent = `${state.characters.length} Ficha${state.characters.length !== 1 ? 's' : ''}`;
      }
      renderCharactersList();

      if (state.characters.length > 0 && !state.selectedCharacter) {
        selectCharacter(state.characters[0].id || state.characters[0].name);
      }
    } catch (err) {
      console.error('Erro ao buscar fichas:', err);
    }
  }

  function getCharacterClassEmoji(className) {
    if (!className) return '🛡️';
    const c = className.toLowerCase();
    if (c.includes('rogue') || c.includes('ladino')) return '🗡️';
    if (c.includes('monk') || c.includes('monge')) return '🥋';
    if (c.includes('cleric') || c.includes('clérigo') || c.includes('clerigo')) return '☀️';
    if (c.includes('bard') || c.includes('bardo')) return '🎻';
    if (c.includes('wizard') || c.includes('mago') || c.includes('sorcerer') || c.includes('feiticeiro')) return '🧙‍♂️';
    if (c.includes('fighter') || c.includes('guerreiro')) return '🛡️';
    if (c.includes('paladin') || c.includes('paladino')) return '⚔️';
    if (c.includes('barbarian') || c.includes('bárbaro') || c.includes('barbaro')) return '🪓';
    if (c.includes('ranger') || c.includes('patrulheiro')) return '🏹';
    if (c.includes('druid') || c.includes('druida')) return '🌿';
    if (c.includes('warlock') || c.includes('bruxo')) return '👁️';
    return '🛡️';
  }

  function renderCharactersList() {
    const container = document.getElementById('characters-list-container');
    if (!container) return;

    const query = (document.getElementById('character-search-input')?.value || '').toLowerCase().trim();
    let chars = state.characters || [];

    if (query) {
      chars = chars.filter(c => 
        (c.name && c.name.toLowerCase().includes(query)) ||
        (c.class && c.class.toLowerCase().includes(query)) ||
        (c.subclass && c.subclass.toLowerCase().includes(query)) ||
        (c.race && c.race.toLowerCase().includes(query))
      );
    }

    if (chars.length === 0) {
      container.innerHTML = '<p class="empty-state">Nenhum herói encontrado com este filtro.</p>';
      return;
    }

    container.innerHTML = chars.map(c => {
      const charId = c.id || c.name;
      const isSel = (state.selectedCharacter && (state.selectedCharacter.id === charId || state.selectedCharacter.name === c.name));
      const emoji = getCharacterClassEmoji(c.class);
      const sub = c.subclass ? ` (${c.subclass})` : '';

      return `
        <div class="character-card ${isSel ? 'active' : ''}" data-char-id="${escapeHtml(charId)}">
          <div class="char-card-top">
            <div class="char-card-avatar">${emoji}</div>
            <div class="char-card-info">
              <div class="char-card-name">${escapeHtml(c.name)}</div>
              <div class="char-card-class">${escapeHtml(c.class || 'Classe')}${escapeHtml(sub)} • Nív ${c.level || 1}</div>
            </div>
          </div>
          <div class="char-card-stats">
            <div class="char-stat-pill char-stat-hp">
              <span>❤️</span> <span>${c.current_hp}/${c.max_hp} PV</span>
            </div>
            <div class="char-stat-pill char-stat-ca">
              <span>🛡️</span> <span>CA ${c.armor_class || 10}</span>
            </div>
            <div class="char-stat-pill">
              <span>⚡</span> <span>Inic ${c.initiative >= 0 ? '+' : ''}${c.initiative || 0}</span>
            </div>
          </div>
        </div>
      `;
    }).join('');

    container.querySelectorAll('.character-card').forEach(card => {
      card.addEventListener('click', () => {
        const id = card.getAttribute('data-char-id');
        selectCharacter(id);
      });
    });
  }

  function selectCharacter(charId) {
    const char = state.characters.find(c => (c.id === charId || c.name === charId));
    if (!char) return;
    state.selectedCharacter = char;
    renderCharactersList();
    renderCharacterSheet(char);
  }

  function renderCharacterSheet(c) {
    const emptyBox = document.getElementById('character-sheet-empty');
    const content = document.getElementById('character-sheet-content');
    if (!content) return;

    if (emptyBox) emptyBox.classList.add('hidden');
    content.classList.remove('hidden');

    const emoji = getCharacterClassEmoji(c.class);
    const sub = c.subclass ? ` • Especialização: ${c.subclass}` : '';
    const race = c.race ? c.race : 'Humanoide';
    const bg = c.background ? ` • Antecedente: ${c.background}` : '';
    const align = c.alignment ? ` • Tendência: ${c.alignment}` : '';

    // Atributos D&D com modificadores
    const attrs = [
      { key: 'str', name: 'Força (FOR)', abbr: 'FOR' },
      { key: 'dex', name: 'Destreza (DES)', abbr: 'DES' },
      { key: 'con', name: 'Constituição (CON)', abbr: 'CON' },
      { key: 'int', name: 'Inteligência (INT)', abbr: 'INT' },
      { key: 'wis', name: 'Sabedoria (SAB)', abbr: 'SAB' },
      { key: 'cha', name: 'Carisma (CAR)', abbr: 'CAR' }
    ];

    const attrCardsHTML = attrs.map(a => {
      const data = (c.attributes && c.attributes[a.key]) || { score: 10, modifier: 0 };
      const modStr = data.modifier >= 0 ? `+${data.modifier}` : `${data.modifier}`;
      const rollFormulaStr = `1d20${data.modifier >= 0 ? '+' : ''}${data.modifier}`;
      return `
        <div class="attr-card" data-attr="${a.abbr}">
          <div class="attr-name">${a.abbr}</div>
          <div class="attr-mod">${modStr}</div>
          <div class="attr-score">Base ${data.score}</div>
          <button class="btn-roll-attr" data-formula="${rollFormulaStr}" data-label="${a.name}">🎲 Rolar</button>
        </div>
      `;
    }).join('');

    // Testes de Resistência
    const savesHTML = attrs.map(a => {
      const save = (c.saving_throws && c.saving_throws[a.key]) || { modifier: 0, proficient: false };
      const modStr = save.modifier >= 0 ? `+${save.modifier}` : `${save.modifier}`;
      const rollFormulaStr = `1d20${save.modifier >= 0 ? '+' : ''}${save.modifier}`;
      return `
        <div class="sheet-list-item">
          <div class="item-left-info">
            <span class="prof-dot ${save.proficient ? 'proficient' : ''}"></span>
            <span>${a.name}</span>
          </div>
          <button class="item-roll-btn" data-formula="${rollFormulaStr}" data-label="Resistência: ${a.abbr}">${modStr}</button>
        </div>
      `;
    }).join('');

    // Perícias D&D
    const skillList = [
      { key: 'acrobatics', name: 'Acrobacia', attr: 'DES' },
      { key: 'animal_handling', name: 'Adestrar Animais', attr: 'SAB' },
      { key: 'arcana', name: 'Arcanismo', attr: 'INT' },
      { key: 'athletics', name: 'Atletismo', attr: 'FOR' },
      { key: 'deception', name: 'Enganação', attr: 'CAR' },
      { key: 'history', name: 'História', attr: 'INT' },
      { key: 'insight', name: 'Intuição', attr: 'SAB' },
      { key: 'intimidation', name: 'Intimidação', attr: 'CAR' },
      { key: 'investigation', name: 'Investigação', attr: 'INT' },
      { key: 'medicine', name: 'Medicina', attr: 'SAB' },
      { key: 'nature', name: 'Natureza', attr: 'INT' },
      { key: 'perception', name: 'Percepção', attr: 'SAB' },
      { key: 'performance', name: 'Atuação', attr: 'CAR' },
      { key: 'persuasion', name: 'Persuasão', attr: 'CAR' },
      { key: 'religion', name: 'Religião', attr: 'INT' },
      { key: 'sleight_of_hand', name: 'Prestidigitação', attr: 'DES' },
      { key: 'stealth', name: 'Furtividade', attr: 'DES' },
      { key: 'survival', name: 'Sobrevivência', attr: 'SAB' },
    ];

    const skillsHTML = skillList.map(sk => {
      const s = (c.skills && (c.skills[sk.key] || c.skills[sk.name])) || { modifier: 0, proficient: false, expertise: false };
      const modStr = s.modifier >= 0 ? `+${s.modifier}` : `${s.modifier}`;
      const rollFormulaStr = `1d20${s.modifier >= 0 ? '+' : ''}${s.modifier}`;
      let dotClass = '';
      if (s.expertise) dotClass = 'expertise';
      else if (s.proficient) dotClass = 'proficient';

      return `
        <div class="sheet-list-item">
          <div class="item-left-info">
            <span class="prof-dot ${dotClass}" title="${s.expertise ? 'Especialização' : (s.proficient ? 'Proficiência' : 'Sem bônus')}"></span>
            <span>${sk.name} <small style="color:var(--text-muted)">(${sk.attr})</small></span>
          </div>
          <button class="item-roll-btn" data-formula="${rollFormulaStr}" data-label="Perícia: ${sk.name}">${modStr}</button>
        </div>
      `;
    }).join('');

    // Ataques & Armas
    let attacksHTML = '<p class="empty-state" style="padding:10px;">Nenhum ataque registrado.</p>';
    if (c.attacks && c.attacks.length > 0) {
      attacksHTML = c.attacks.map(atk => {
        const rawBonus = atk.attack_bonus !== undefined ? atk.attack_bonus : (atk.bonus !== undefined ? atk.bonus : '+0');
        let bonusStr = String(rawBonus).trim();
        if (!bonusStr.startsWith('+') && !bonusStr.startsWith('-') && !bonusStr.startsWith('DC')) {
          bonusStr = '+' + bonusStr;
        }
        const rollFormula = bonusStr.startsWith('DC') ? '1d20' : `1d20${bonusStr}`;
        return `
          <div class="attack-card">
            <div class="attack-header">
              <span class="attack-name">⚔️ ${escapeHtml(atk.name)}</span>
              <button class="btn btn-xs btn-outline-cyan item-roll-btn" data-formula="${rollFormula}" data-label="Ataque: ${escapeHtml(atk.name)}">${bonusStr} para Acertar</button>
            </div>
            <div class="attack-damage">Dano: <strong>${escapeHtml(atk.damage || '1d4')}</strong> ${atk.damage_type ? `• ${escapeHtml(atk.damage_type)}` : ''}</div>
            ${atk.notes ? `<div style="font-size:0.75rem; color:var(--text-muted);">${escapeHtml(atk.notes)}</div>` : ''}
          </div>
        `;
      }).join('');
    }

    // Habilidades & Características
    let featuresHTML = '<p class="empty-state" style="padding:10px;">Nenhuma habilidade registrada.</p>';
    if (c.features && c.features.length > 0) {
      featuresHTML = c.features.map(f => `
        <div style="background:rgba(255,255,255,0.03); border:1px solid rgba(255,255,255,0.06); padding:10px; border-radius:6px; margin-bottom:8px;">
          <div style="font-weight:700; color:var(--text-main); font-size:0.86rem; margin-bottom:4px;">✨ ${escapeHtml(f.name)} ${f.source ? `<small style="color:var(--text-muted);">(${escapeHtml(f.source)})</small>` : ''}</div>
          <div style="font-size:0.8rem; color:var(--text-muted); line-height:1.4;">${escapeHtml(f.description)}</div>
        </div>
      `).join('');
    }

    // Equipamentos / Inventário
    const inv = (c.inventory && c.inventory.length > 0) ? c.inventory : (c.equipment || []);
    let equipHTML = '<p class="empty-state" style="padding:10px;">Mochila vazia.</p>';
    if (inv.length > 0) {
      equipHTML = `<ul style="margin:0; padding-left:18px; font-size:0.82rem; color:var(--text-muted);">` +
        inv.map(eq => `<li><strong>${escapeHtml(eq.name)}</strong> ${eq.quantity > 1 ? `x${eq.quantity}` : ''}</li>`).join('') +
        `</ul>`;
    }

    content.innerHTML = `
      <!-- Header do Herói -->
      <div class="sheet-hero-header">
        <div class="sheet-hero-title">
          <h2>${emoji} ${escapeHtml(c.name)} ${c.inspiration ? '⭐' : ''}</h2>
          <div class="sheet-hero-subtitle">
            <span><strong>${escapeHtml(race)}</strong></span>
            <span>•</span>
            <span>${escapeHtml(c.class)} Nível ${c.level}</span>
            <span>${escapeHtml(sub)}</span>
            <span>${escapeHtml(bg)}</span>
            <span>${escapeHtml(align)}</span>
          </div>
        </div>

        <div class="sheet-hero-actions">
          <!-- Gestor de PVs -->
          <div class="sheet-hp-card">
            <div class="sheet-hp-info">
              <span class="sheet-hp-label">Pontos de Vida</span>
              <span class="sheet-hp-value" id="sheet-hp-display">${c.current_hp} / ${c.max_hp} PV ${c.temp_hp > 0 ? `<small style="font-size:0.75rem; color:#fde047;">(+${c.temp_hp} Temp)</small>` : ''}</span>
            </div>
            <div class="sheet-hp-btns">
              <button class="btn-hp btn-hp-dmg" id="btn-char-dmg-5" title="Levar 5 de Dano">-5</button>
              <button class="btn-hp btn-hp-dmg" id="btn-char-dmg-1" title="Levar 1 de Dano">-1</button>
              <button class="btn-hp btn-hp-heal" id="btn-char-heal-1" title="Curar 1 PV">+1</button>
              <button class="btn-hp btn-hp-heal" id="btn-char-heal-5" title="Curar 5 PV">+5</button>
              <button class="btn-hp btn-hp-heal" id="btn-char-heal-all" title="Curar Tudo">Max</button>
            </div>
          </div>

          <!-- Ação: Editar Ficha -->
          <button id="btn-edit-char" class="btn btn-outline-cyan btn-sm" title="Editar dados da ficha">
            <span class="btn-icon">✏️</span> Editar
          </button>

          <!-- Ação: Deletar Ficha -->
          <button id="btn-delete-char" class="btn btn-outline-danger btn-sm" title="Excluir permanentemente esta ficha">
            <span class="btn-icon">🗑️</span> Excluir
          </button>

          <!-- Ação: Postar Ficha no Discord -->
          <button id="btn-post-char-discord" class="btn btn-primary btn-sm" title="Enviar ficha formatada para o canal de texto do Discord selecionado no topo">
            <span class="btn-icon">🚀</span> Postar no Discord
          </button>
        </div>
      </div>

      <!-- Barra de Vitais Principais -->
      <div class="sheet-vitals-row">
        <div class="vital-box">
          <div class="vital-label">Classe de Armadura</div>
          <div class="vital-value" style="color:var(--accent-cyan);">🛡️ ${c.armor_class || 10}</div>
        </div>
        <div class="vital-box">
          <div class="vital-label">Iniciativa</div>
          <div class="vital-value" style="color:var(--accent-amber);">${c.initiative >= 0 ? '+' : ''}${c.initiative || 0}</div>
        </div>
        <div class="vital-box">
          <div class="vital-label">Deslocamento</div>
          <div class="vital-value">${escapeHtml(c.speed || '9m')}</div>
        </div>
        <div class="vital-box">
          <div class="vital-label">Proficiência</div>
          <div class="vital-value" style="color:var(--accent-purple);">+${c.proficiency_bonus || 2}</div>
        </div>
        <div class="vital-box">
          <div class="vital-label">Dados de Vida</div>
          <div class="vital-value">${c.hit_dice_current || c.level}/${c.hit_dice || `${c.level}d8`}</div>
        </div>
        <div class="vital-box">
          <div class="vital-label">Inspiração</div>
          <div class="vital-value">${c.inspiration ? '⭐ Sim' : 'Não'}</div>
        </div>
      </div>

      <!-- Grade dos 6 Atributos D&D -->
      <div class="sheet-attributes-grid">
        ${attrCardsHTML}
      </div>

      <!-- Colunas de Testes de Resistência e Perícias -->
      <div class="sheet-columns-row">
        <div class="sheet-section-box">
          <div class="sheet-section-title">
            <span>🛡️ Testes de Resistência</span>
            <small style="color:var(--text-muted);">Clique para Rolar</small>
          </div>
          <div class="sheet-list-items">
            ${savesHTML}
          </div>
        </div>

        <div class="sheet-section-box">
          <div class="sheet-section-title">
            <span>🎯 Perícias Treinadas</span>
            <small style="color:var(--text-muted);"><span style="color:var(--accent-purple);">●</span> Treinada <span style="color:var(--accent-amber);">●</span> Especialista</small>
          </div>
          <div class="sheet-list-items">
            ${skillsHTML}
          </div>
        </div>
      </div>

      <!-- Ataques & Armas -->
      <div class="sheet-section-box" style="margin-bottom:24px;">
        <div class="sheet-section-title">
          <span>⚔️ Ações de Combate & Armas</span>
          <small style="color:var(--text-muted);">${(c.attacks || []).length} ataque(s) catalogado(s)</small>
        </div>
        <div class="attacks-grid">
          ${attacksHTML}
        </div>
      </div>

      <!-- Habilidades de Classe & Equipamentos -->
      <div class="sheet-columns-row">
        <div class="sheet-section-box">
          <div class="sheet-section-title">
            <span>✨ Habilidades & Características</span>
          </div>
          <div>${featuresHTML}</div>
        </div>

        <div class="sheet-section-box">
          <div class="sheet-section-title">
            <span>🎒 Equipamento & Itens</span>
          </div>
          <div>${equipHTML}</div>
        </div>
      </div>
    `;

    // Listeners de rolagem em todos os botões de atributo, perícia, save e arma
    content.querySelectorAll('.btn-roll-attr, .item-roll-btn').forEach(btn => {
      btn.addEventListener('click', (e) => {
        e.stopPropagation();
        const formula = btn.getAttribute('data-formula');
        const label = btn.getAttribute('data-label') || 'Teste';
        if (formula) {
          showToast(`🎲 Rolando ${label} (${formula})...`);
          rollFormula(formula);
        }
      });
    });

    // Gestão de PVs
    const updateHP = async (newHP) => {
      const charId = c.id || c.name;
      try {
        const res = await fetch('/api/character/hp', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            characterId: charId,
            currentHp: newHP,
            tempHp: c.temp_hp || 0
          })
        });
        const d = await res.json();
        if (res.ok && d.character) {
          c.current_hp = d.character.current_hp;
          document.getElementById('sheet-hp-display').innerHTML = 
            `${c.current_hp} / ${c.max_hp} PV ${c.temp_hp > 0 ? `<small style="font-size:0.75rem; color:#fde047;">(+${c.temp_hp} Temp)</small>` : ''}`;
          renderCharactersList();
          showToast(`❤️ PV de ${c.name} atualizado: ${c.current_hp}/${c.max_hp}`);
        } else {
          showToast('❌ Erro ao atualizar PV');
        }
      } catch (_) {
        showToast('❌ Falha na conexão ao atualizar PV');
      }
    };

    document.getElementById('btn-char-dmg-1')?.addEventListener('click', () => updateHP(c.current_hp - 1));
    document.getElementById('btn-char-dmg-5')?.addEventListener('click', () => updateHP(c.current_hp - 5));
    document.getElementById('btn-char-heal-1')?.addEventListener('click', () => updateHP(c.current_hp + 1));
    document.getElementById('btn-char-heal-5')?.addEventListener('click', () => updateHP(c.current_hp + 5));
    document.getElementById('btn-char-heal-all')?.addEventListener('click', () => updateHP(c.max_hp));

    // Editar Ficha
    document.getElementById('btn-edit-char')?.addEventListener('click', () => {
      openEditModal(c);
    });

    // Deletar Ficha
    document.getElementById('btn-delete-char')?.addEventListener('click', async () => {
      const charId = c.id || c.name;
      if (!confirm(`Tem certeza que deseja excluir permanentemente a ficha de "${c.name}"? Esta ação não pode ser desfeita.`)) {
        return;
      }

      try {
        const res = await fetch('/api/character/delete', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ characterId: charId })
        });
        const data = await res.json();
        if (res.ok && data.status === 'deleted') {
          showToast(`🗑️ Ficha de "${c.name}" excluída com sucesso!`);
          state.selectedCharacter = null;
          const emptyBox = document.getElementById('character-sheet-empty');
          const content = document.getElementById('character-sheet-content');
          if (content) content.classList.add('hidden');
          if (emptyBox) emptyBox.classList.remove('hidden');
          await fetchCharacters();
        } else {
          showToast(`❌ Falha ao excluir: ${data.error || 'Erro desconhecido'}`);
        }
      } catch (_) {
        showToast('❌ Falha de conexão ao excluir ficha');
      }
    });

    // Postar Ficha no Discord
    document.getElementById('btn-post-char-discord')?.addEventListener('click', async () => {
      if (!state.selectedTextChannelId) {
        showToast('⚠️ Selecione um canal de texto no topo antes de postar a ficha!');
        return;
      }

      const charId = c.id || c.name;
      try {
        const res = await fetch('/api/character/send', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            characterId: charId,
            channelId: state.selectedTextChannelId,
            view: 'full'
          })
        });
        if (res.ok) {
          showToast(`🚀 Ficha de "${c.name}" postada no canal do Discord com sucesso!`);
        } else {
          const errData = await res.json();
          showToast(`❌ Falha: ${errData.error || 'Não foi possível postar'}`);
        }
      } catch (_) {
        showToast('❌ Erro de conexão ao enviar ficha');
      }
    });
  }

  // ═══════════════════════════════════════════════════════════════════
  // GRAVAÇÃO DE SESSÃO (TAB 4)
  // ═══════════════════════════════════════════════════════════════════
  function initRecorderControls() {
    const btnStart = document.getElementById('btn-start-record');
    const btnStop = document.getElementById('btn-stop-record');
    const dot = document.getElementById('record-pulse-dot');
    const title = document.getElementById('record-state-title');
    const sub = document.getElementById('record-state-subtitle');

    btnStart.addEventListener('click', async () => {
      if (!state.currentGuildId || !state.currentVoiceChannelId) {
        showToast('⚠️ Selecione um servidor e um canal de voz no topo!');
        return;
      }

      try {
        const res = await fetch('/api/record/start', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            guildId: state.currentGuildId,
            channelId: state.currentVoiceChannelId
          })
        });

        const data = await res.json();
        if (!res.ok) {
          showToast('❌ Erro: ' + (data.error || 'Falha ao iniciar gravação'));
          return;
        }

        state.isRecording = true;
        dot.className = 'record-pulse-dot recording';
        title.textContent = '🔴 Gravando Sessão de RPG ao Vivo';
        sub.textContent = 'Gravando áudio de todos os jogadores com criptografia DAVE';
        btnStart.disabled = true;
        btnStop.disabled = false;

        // Cronômetro
        state.recordingSeconds = 0;
        clearInterval(state.recordingTimer);
        state.recordingTimer = setInterval(() => {
          state.recordingSeconds++;
          document.getElementById('recorder-stopwatch').textContent = formatUptime(state.recordingSeconds);
        }, 1000);

        showToast('🔴 Gravação de sessão iniciada!');
      } catch (_) {
        showToast('❌ Falha de comunicação com o serviço de gravação');
      }
    });

    btnStop.addEventListener('click', async () => {
      if (!state.currentGuildId) return;

      try {
        const res = await fetch('/api/record/stop', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ guildId: state.currentGuildId })
        });

        const data = await res.json();
        state.isRecording = false;
        clearInterval(state.recordingTimer);
        dot.className = 'record-pulse-dot';
        title.textContent = 'Gravação Concluída';
        sub.textContent = 'Arquivo salvo na pasta gravacoes/';
        btnStart.disabled = false;
        btnStop.disabled = true;

        showToast('⏹️ Sessão gravada com sucesso!');
        fetchRecordings();
      } catch (_) {
        showToast('❌ Falha ao finalizar gravação');
      }
    });

    document.getElementById('btn-refresh-recordings').addEventListener('click', fetchRecordings);
  }

  async function fetchRecordings() {
    try {
      const res = await fetch('/api/record/files');
      if (!res.ok) return;
      const files = await res.json();

      const container = document.getElementById('recordings-list-container');
      if (files.length === 0) {
        container.innerHTML = '<p class="empty-state">Nenhum arquivo gravado.</p>';
        return;
      }

      container.innerHTML = files.map(f => `
        <div class="recording-item" data-path="${escapeHtml(f.name)}">
          <div>
            <div style="font-weight: 600; color: #fff;">${escapeHtml(f.name)}</div>
            <div style="font-size: 0.72rem; color: var(--text-dim);">${f.formattedSize} • ${f.modifiedAt}</div>
          </div>
          <button class="btn btn-outline-cyan btn-xs btn-play-recording">🎧 Ouvir</button>
        </div>
      `).join('');

      container.querySelectorAll('.recording-item').forEach(item => {
        item.addEventListener('click', () => {
          const name = item.getAttribute('data-path');
          const audioPlayer = document.getElementById('recording-audio-preview');
          const label = document.getElementById('preview-filename');

          audioPlayer.src = `/api/record/download/${encodeURIComponent(name)}`;
          label.textContent = name;
          audioPlayer.play().catch(() => {});
        });
      });
    } catch (_) {}
  }

  // ═══════════════════════════════════════════════════════════════════
  // CHAT DO MESTRE (TAB 5)
  // ═══════════════════════════════════════════════════════════════════
  function initMasterChat() {
    const btnSend = document.getElementById('btn-send-chat');
    const msgInput = document.getElementById('chat-message-input');

    btnSend.addEventListener('click', async () => {
      const channelId = document.getElementById('chat-channel-select').value || state.selectedTextChannelId;
      const message = msgInput.value.trim();

      if (!channelId || !message) {
        showToast('⚠️ Digite uma mensagem e selecione um canal!');
        return;
      }

      try {
        const res = await fetch('/api/chat/send', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ channelId, message })
        });

        if (res.ok) {
          showToast('🚀 Mensagem enviada pelo bot!');
          msgInput.value = '';
        } else {
          showToast('❌ Falha ao enviar mensagem');
        }
      } catch (_) {}
    });

    // NPC Presets
    document.querySelectorAll('.btn-chip').forEach(chip => {
      chip.addEventListener('click', () => {
        const npc = chip.getAttribute('data-npc');
        if (npc === 'strahd') {
          msgInput.value = '🧛 *"Eu sou a Terra Antiga. Eu sou a Terra. O sangue desta terra corre em minhas veias, e a Baróvia é meu lar para todo o sempre."*';
        } else if (npc === 'madam_eva') {
          msgInput.value = '🔮 *"As cartas Tarokka não mentem, andarilhos. O destino os trouxe até estas névoas, e apenas cinco relíquias podem libertá-los das garras do Diabo."*';
        } else if (npc === 'van_richten') {
          msgInput.value = '🏹 *"Para caçar um monstro como Strahd von Zarovich, não basta ter coragem. É preciso ter paciência, aço afiado e conhecimento de cada uma de suas fraquezas."*';
        } else if (npc === 'baba_lysaga') {
          msgInput.value = '🧙‍♀️ *"Malditos corvos e forasteiros! Vocês nunca tocarão no meu amado menino Strahd! Que as trevas dos pântanos os devorem vivos!"*';
        }
      });
    });
  }

  // ═══════════════════════════════════════════════════════════════════
  // DIAGNÓSTICOS & TERMINAL DE LOGS (TAB 6)
  // ═══════════════════════════════════════════════════════════════════
  function initSystemLogs() {
    document.getElementById('btn-run-doctor').addEventListener('click', runDoctor);
    document.getElementById('btn-clear-logs').addEventListener('click', () => {
      document.getElementById('terminal-logs').innerHTML = '';
    });
    runDoctor();
  }

  async function runDoctor() {
    const list = document.getElementById('health-checks-container');
    list.innerHTML = '<div class="loading-spinner">Executando suite de testes de diagnóstico...</div>';

    try {
      const res = await fetch('/api/diagnostics');
      if (!res.ok) return;
      const report = await res.json();

      document.getElementById('health-pass-count').textContent = `${report.total_pass} Passou`;
      document.getElementById('health-warn-count').textContent = `${report.total_warn} Avisos`;
      document.getElementById('health-fail-count').textContent = `${report.total_fail} Falhas`;

      list.innerHTML = (report.checks || []).map(c => {
        let badgeColor = 'var(--accent-green)';
        let icon = '✅';
        if (c.status === 'WARN') { badgeColor = 'var(--accent-amber)'; icon = '⚠️'; }
        else if (c.status === 'FAIL') { badgeColor = 'var(--accent-crimson)'; icon = '❌'; }

        return `
          <div class="health-check-row">
            <span>${icon}</span>
            <div style="flex: 1;">
              <strong style="color: #fff;">${escapeHtml(c.name)}</strong>
              <div style="color: var(--text-muted); font-size: 0.72rem;">${escapeHtml(c.message)}</div>
            </div>
          </div>
        `;
      }).join('');
    } catch (_) {
      list.innerHTML = '<p class="empty-state">Erro ao rodar diagnóstico.</p>';
    }
  }

  // ═══════════════════════════════════════════════════════════════════
  // SERVER-SENT EVENTS (SSE) PARA LOGS E ATUALIZAÇÕES EM TEMPO REAL
  // ═══════════════════════════════════════════════════════════════════
  function initSSE() {
    const term = document.getElementById('terminal-logs');

    try {
      const evtSource = new EventSource('/api/events');
      evtSource.onmessage = (e) => {
        try {
          const msg = JSON.parse(e.data);
          if (msg.type === 'log') {
            const line = document.createElement('div');
            line.className = `log-line ${(msg.data.level || 'info').toLowerCase()}`;
            line.textContent = `[${msg.data.timestamp}] [${msg.data.level}] ${msg.data.message}`;
            term.appendChild(line);
            term.scrollTop = term.scrollHeight;
          }
        } catch (_) {}
      };
    } catch (_) {}
  }

  // ═══════════════════════════════════════════════════════════════════
  // ATALHOS DE TECLADO
  // ═══════════════════════════════════════════════════════════════════
  function initKeyboardShortcuts() {
    window.addEventListener('keydown', (e) => {
      // Ignora se estiver digitando em um input ou textarea
      if (['INPUT', 'TEXTAREA', 'SELECT'].includes(document.activeElement.tagName)) {
        return;
      }

      if (e.code === 'Space') {
        e.preventDefault();
        document.getElementById('ctrl-play-pause').click();
      } else if (e.code === 'ArrowRight' && e.altKey) {
        e.preventDefault();
        document.getElementById('ctrl-skip').click();
      } else if (e.code === 'ArrowLeft' && e.altKey) {
        e.preventDefault();
        document.getElementById('ctrl-prev').click();
      } else if (e.key === 'r' || e.key === 'R') {
        rollFormula('1d20');
      } else if (e.key === 'm' || e.key === 'M') {
        document.getElementById('ctrl-mute').click();
      }
    });
  }

  // ═══════════════════════════════════════════════════════════════════
  // UTILITÁRIOS (TOAST, PARSERS, ESCAPE)
  // ═══════════════════════════════════════════════════════════════════
  function showToast(message) {
    const container = document.getElementById('toast-container');
    const toast = document.createElement('div');
    toast.className = 'toast';
    toast.textContent = message;
    container.appendChild(toast);
    setTimeout(() => {
      toast.style.opacity = '0';
      toast.style.transform = 'translateY(10px)';
      toast.style.transition = '0.3s ease';
      setTimeout(() => toast.remove(), 300);
    }, 3200);
  }

  function escapeHtml(text) {
    if (!text) return '';
    return String(text)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#039;');
  }

  function formatDurationSeconds(seconds) {
    const s = Math.round(Number(seconds));
    if (isNaN(s) || s <= 0) return '00:00';
    const mins = Math.floor(s / 60);
    const secs = s % 60;
    return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  }

  function parseDurationToSeconds(durationStr) {
    if (!durationStr || typeof durationStr !== 'string') return 0;
    const parts = durationStr.split(':').map(p => parseInt(p, 10));
    if (parts.length === 2) {
      return (parts[0] * 60) + parts[1];
    } else if (parts.length === 3) {
      return (parts[0] * 3600) + (parts[1] * 60) + parts[2];
    }
    return 0;
  }
});
