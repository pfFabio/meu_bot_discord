import express from 'express';
import { 
    joinVoiceChannel, 
    createAudioPlayer, 
    createAudioResource, 
    AudioPlayerStatus, 
    VoiceConnectionStatus, 
    StreamType,
    NoSubscriberBehavior,
    EndBehaviorType
} from '@discordjs/voice';
import { Client, GatewayIntentBits } from 'discord.js';
import { spawn } from 'child_process';
import { pipeline } from 'stream';
import dotenv from 'dotenv';
import path from 'path';
import fs from 'fs';
import { createWriteStream } from 'fs';
import { fileURLToPath } from 'url';
import prism from 'prism-media';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const MUSIC_DIR = path.resolve(__dirname, '../musicas');
const RECORDINGS_DIR = path.resolve(__dirname, '../gravacoes');
if (!fs.existsSync(RECORDINGS_DIR)) fs.mkdirSync(RECORDINGS_DIR, { recursive: true });

// Cache em memória para duração das faixas locais
const durationCache = new Map();

function formatDurationSeconds(seconds) {
    const s = Math.round(Number(seconds));
    if (isNaN(s) || s <= 0) return 'Local';
    const mins = Math.floor(s / 60);
    const secs = s % 60;
    return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
}

// Retorna todas as faixas encontradas na pasta /musicas
function getLocalMusicFiles() {
    if (!fs.existsSync(MUSIC_DIR)) return [];
    const validExts = ['.mp3', '.ogg', '.wav', '.flac', '.m4a', '.aac'];
    try {
        const files = fs.readdirSync(MUSIC_DIR);
        return files
            .filter(file => validExts.includes(path.extname(file).toLowerCase()))
            .map(file => {
                const ext = path.extname(file);
                const baseName = path.basename(file, ext);
                let cleanTitle = baseName;
                let emoji = '🎵';

                const lowerBase = baseName.toLowerCase();
                if (lowerBase.includes('mago') && lowerBase.includes('strahd')) {
                    cleanTitle = 'Mago e Strahd (História de Baróvia)';
                    emoji = '🧙‍♂️';
                } else if (lowerBase.includes('vistani') && (lowerBase.includes('introdução') || lowerBase.includes('introducao'))) {
                    cleanTitle = 'Vistani (Canção de Introdução)';
                    emoji = '🎪';
                } else if (lowerBase.includes('vistani') && lowerBase.includes('festa')) {
                    cleanTitle = 'Vistani (Festa de Abertura)';
                    emoji = '🎪';
                } else if (lowerBase.includes('vistani') || lowerBase.includes('gypsy') || lowerBase.includes('balkan') || lowerBase.includes('cigano')) {
                    cleanTitle = cleanTitle.charAt(0).toUpperCase() + cleanTitle.slice(1);
                    emoji = '🎪';
                } else if (lowerBase.includes('strahd') || lowerBase.includes('vampiro')) {
                    cleanTitle = cleanTitle.charAt(0).toUpperCase() + cleanTitle.slice(1);
                    emoji = '🩸';
                } else if (lowerBase.includes('combate') || lowerBase.includes('batalha')) {
                    cleanTitle = cleanTitle.charAt(0).toUpperCase() + cleanTitle.slice(1);
                    emoji = '⚔️';
                } else {
                    cleanTitle = cleanTitle.charAt(0).toUpperCase() + cleanTitle.slice(1);
                }

                let duration = durationCache.get(file) || 'Local';
                const filePath = path.join(MUSIC_DIR, file);

                return {
                    filename: file,
                    title: cleanTitle,
                    emoji,
                    filePath,
                    ext,
                    duration
                };
            });
    } catch (err) {
        console.error('[AudioBridge] Erro ao ler pasta musicas:', err);
        return [];
    }
}

// Inicializa verificação de duração em segundo plano
function probeLocalDurations() {
    const tracks = getLocalMusicFiles();
    for (const track of tracks) {
        if (!durationCache.has(track.filename)) {
            const proc = spawn('ffprobe', [
                '-i', track.filePath,
                '-show_entries', 'format=duration',
                '-v', 'quiet',
                '-of', 'csv=p=0'
            ]);
            let out = '';
            proc.stdout.on('data', d => out += d.toString());
            proc.on('close', code => {
                if (code === 0 && out.trim()) {
                    const dur = formatDurationSeconds(out.trim());
                    durationCache.set(track.filename, dur);
                }
            });
            proc.on('error', () => {});
        }
    }
}
setTimeout(probeLocalDurations, 1000);

// Carregar .env
dotenv.config({ path: path.resolve(__dirname, '../.env') });

const TOKEN = process.env.DISCORD_TOKEN;
const PORT = process.env.AUDIO_BRIDGE_PORT || 2389;

if (!TOKEN) {
    console.error('[AudioBridge] ERRO: DISCORD_TOKEN não encontrado no ambiente.');
    process.exit(1);
}

// Inicializa cliente Discord exclusivo para voz com DAVE E2EE
const client = new Client({
    intents: [
        GatewayIntentBits.Guilds,
        GatewayIntentBits.GuildVoiceStates,
    ]
});

// Mapa de servidores -> Estado de áudio
const guildStates = new Map();

function getOrCreateState(guildId) {
    if (!guildStates.has(guildId)) {
        const player = createAudioPlayer({
            behaviors: {
                noSubscriber: NoSubscriberBehavior.Play,
                maxMissedFrames: 50
            }
        });
        
        player.on(AudioPlayerStatus.Idle, () => {
            const state = guildStates.get(guildId);
            if (state) {
                cleanupActiveProcesses(state);
                const finishedTrack = state.current;

                if (state.isSkipping) {
                    state.isSkipping = false;
                    if (state.queue.length > 0) {
                        const nextTrack = state.queue.shift();
                        playTrack(guildId, nextTrack);
                    } else {
                        state.isPlaying = false;
                        state.current = null;
                    }
                } else if (state.loopMode === 'track' && finishedTrack) {
                    console.log(`[AudioBridge] Repetindo faixa em loop na guild ${guildId}: ${finishedTrack.title}`);
                    playTrack(guildId, finishedTrack);
                } else if (state.loopMode === 'queue' && finishedTrack) {
                    state.queue.push(finishedTrack);
                    const nextTrack = state.queue.shift();
                    playTrack(guildId, nextTrack);
                } else if (state.queue.length > 0) {
                    const nextTrack = state.queue.shift();
                    playTrack(guildId, nextTrack);
                } else {
                    state.isPlaying = false;
                    state.current = null;
                }
            }
        });

        player.on(AudioPlayerStatus.Playing, () => {
            console.log(`[AudioBridge] Player está TOCANDO na guild ${guildId}`);
        });

        player.on(AudioPlayerStatus.Buffering, () => {
            console.log(`[AudioBridge] Player em BUFFERING na guild ${guildId}`);
        });

        player.on('error', (error) => {
            console.error(`[AudioBridge] Erro no player (guild ${guildId}):`, error.message);
            const state = guildStates.get(guildId);
            if (state) {
                cleanupActiveProcesses(state);
                if (state.queue.length > 0) {
                    const nextTrack = state.queue.shift();
                    playTrack(guildId, nextTrack);
                } else {
                    state.isPlaying = false;
                    state.current = null;
                }
            }
        });

        guildStates.set(guildId, {
            guildId,
            connection: null,
            player,
            queue: [],
            current: null,
            isPlaying: false,
            activeProcesses: [],
            loopMode: 'off',
            isSkipping: false,
            // Estado de gravação de sessão
            isRecording: false,
            recordingStreams: new Map(),
            recordingDir: null,
            recordingStartTime: null,
            speakingHandler: null,
        });
    }
    return guildStates.get(guildId);
}

function cleanupActiveProcesses(state) {
    if (state.activeProcesses && state.activeProcesses.length > 0) {
        for (const proc of state.activeProcesses) {
            try {
                proc.kill('SIGKILL');
            } catch (_) {}
        }
        state.activeProcesses = [];
    }
}

// Extrai a URL direta de áudio HTTPS usando yt-dlp
async function extractDirectStreamUrl(targetUrl) {
    return new Promise((resolve) => {
        const proc = spawn('yt-dlp', [
            '-f', 'bestaudio/best',
            '-g',
            '--no-playlist',
            '--no-warnings',
            targetUrl
        ]);

        let out = '';
        let errData = '';
        proc.stdout.on('data', (d) => out += d.toString());
        proc.stderr.on('data', (d) => errData += d.toString());
        proc.on('close', (code) => {
            if (code === 0 && out.trim()) {
                const stream = out.trim().split('\n')[0].trim();
                resolve(stream);
            } else {
                console.warn(`[AudioBridge] yt-dlp stream extraction warning/error (code ${code}):`, errData.trim());
                resolve(targetUrl); // Fallback para URL original
            }
        });
        proc.on('error', (err) => {
            console.error('[AudioBridge] Falha ao invocar yt-dlp:', err);
            resolve(targetUrl);
        });
    });
}

async function playTrack(guildId, track) {
    const state = getOrCreateState(guildId);
    state.current = track;
    state.isPlaying = true;
    cleanupActiveProcesses(state);

    try {
        console.log(`[AudioBridge] Iniciando reprodução na guild ${guildId}: ${track.title}`);

        let streamUrl = track.url;
        // Não tentar extrair stream do YouTube se for arquivo local, preset ou stream direto .mp3
        if (!track.isLocal && !track.isPreset && !track.url.includes('zeno.fm') && !track.url.endsWith('.mp3')) {
            console.log(`[AudioBridge] Extraindo stream direto para: ${track.url}`);
            streamUrl = await extractDirectStreamUrl(track.url);
        }

        console.log(`[AudioBridge] FFmpeg conectando ao áudio: ${streamUrl.substring(0, 80)}...`);

        const ffmpegArgs = [];
        if (streamUrl.startsWith('http://') || streamUrl.startsWith('https://')) {
            ffmpegArgs.push('-reconnect', '1', '-reconnect_streamed', '1', '-reconnect_delay_max', '5');
        }
        ffmpegArgs.push('-i', streamUrl, '-f', 's16le', '-ar', '48000', '-ac', '2', 'pipe:1');

        const ffmpeg = spawn('ffmpeg', ffmpegArgs, { stdio: ['ignore', 'pipe', 'pipe'] });
        ffmpeg.stderr.on('data', (chunk) => {
            const msg = chunk.toString();
            if (msg.includes('Error') || msg.includes('Server returned') || msg.includes('Failed')) {
                console.error(`[AudioBridge] FFmpeg erro:`, msg.trim());
            }
        });

        state.activeProcesses.push(ffmpeg);

        const resource = createAudioResource(ffmpeg.stdout, {
            inputType: StreamType.Raw,
            inlineVolume: true
        });

        if (resource.volume) {
            resource.volume.setVolume(1.0);
        }

        state.player.play(resource);

        if (state.connection) {
            state.connection.subscribe(state.player);
        }
    } catch (err) {
        console.error(`[AudioBridge] Falha ao iniciar reprodução:`, err);
        state.isPlaying = false;
        state.current = null;
    }
}

// Presets de Áudio D&D
const AMBIENT_PRESETS = {
    taverna: { title: "🍻 Taverna Medieval", url: "https://stream.zeno.fm/f3wvbbqmdg8uv", duration: "Stream 24/7", isPreset: true },
    combate: { title: "⚔️ Combate Épico", url: "https://stream.zeno.fm/87v2x8yvh7zuv", duration: "Stream 24/7", isPreset: true },
    masmorra: { title: "🕯️ Masmorra & Caverna", url: "https://stream.zeno.fm/0r0xa792kwzuv", duration: "Stream 24/7", isPreset: true },
    floresta: { title: "🌿 Floresta Feérica", url: "https://stream.zeno.fm/6wz252u1x1zuv", duration: "Stream 24/7", isPreset: true },
};

// Express REST API
const app = express();
app.use(express.json());

// Inicia o servidor HTTP imediatamente
app.listen(PORT, '127.0.0.1', () => {
    console.log(`[AudioBridge] Microserviço de áudio ouvindo em http://127.0.0.1:${PORT}`);
});

// Health Check
app.get('/health', (req, res) => {
    res.json({
        status: 'ok',
        dave: true,
        clientReady: client.isReady(),
        activeGuilds: guildStates.size,
        uptime: process.uptime()
    });
});

// Lista de Músicas Locais da pasta /musicas
app.get('/local-tracks', (req, res) => {
    const tracks = getLocalMusicFiles().map(t => ({
        filename: t.filename,
        title: t.title,
        emoji: t.emoji,
        duration: t.duration,
        path: t.filePath
    }));
    res.json({ tracks });
});

// Play / Enqueue
app.post('/play', async (req, res) => {
    const { guildId, channelId, query, requester } = req.body;
    if (!guildId || !channelId || !query) {
        return res.status(400).json({ error: 'guildId, channelId e query são obrigatórios' });
    }

    try {
        const guild = await client.guilds.fetch(guildId);
        if (!guild) {
            return res.status(404).json({ error: 'Servidor não encontrado' });
        }

        const state = getOrCreateState(guildId);

        // Conectar ao canal de voz com suporte a DAVE E2EE
        if (!state.connection || state.connection.state.status === VoiceConnectionStatus.Destroyed) {
            state.connection = joinVoiceChannel({
                channelId,
                guildId,
                adapterCreator: guild.voiceAdapterCreator,
                selfDeaf: !state.isRecording,  // desativa surdez se estiver gravando
                selfMute: false
            });

            state.connection.on('stateChange', (oldState, newState) => {
                console.log(`[AudioBridge] Conexão de voz (guild ${guildId}): ${oldState.status} ➔ ${newState.status}`);
            });

            state.connection.subscribe(state.player);
        }

// Extrai faixas de uma playlist do YouTube de forma rápida com --flat-playlist
async function extractPlaylistTracks(playlistUrl, requester) {
    return new Promise((resolve) => {
        const proc = spawn('yt-dlp', [
            '--flat-playlist',
            '--playlist-end', '50',
            '--print', '%(title)s|%(id)s|%(duration_string)s',
            playlistUrl
        ]);
        let out = '';
        proc.stdout.on('data', (d) => out += d.toString());
        proc.on('close', () => {
            const lines = out.trim().split('\n');
            const tracks = [];
            for (const line of lines) {
                if (!line.trim()) continue;
                const parts = line.split('|');
                const title = parts[0] ? parts[0].trim() : 'Faixa de Playlist';
                const id = parts[1] ? parts[1].trim() : '';
                const duration = parts[2] ? parts[2].trim() : 'Desconhecida';
                if (id) {
                    tracks.push({
                        title,
                        url: `https://www.youtube.com/watch?v=${id}`,
                        duration,
                        requester: requester || 'Jogador',
                        isPreset: false
                    });
                }
            }
            resolve(tracks);
        });
        proc.on('error', () => resolve([]));
    });
}

        // Resolver faixa ou playlist
        let trackInfo = null;
        const lowerQuery = query.toLowerCase().trim();

        // 1. Detectar Playlist do YouTube
        const isPlaylist = query.includes('/playlist?list=') || 
            (query.includes('list=') && (query.includes('&playlist') || !query.includes('watch?v=')));

        if (isPlaylist) {
            console.log(`[AudioBridge] Extraindo playlist: ${query}`);
            const playlistTracks = await extractPlaylistTracks(query, requester);
            if (playlistTracks.length > 0) {
                if (!state.isPlaying) {
                    const first = playlistTracks[0];
                    const remaining = playlistTracks.slice(1);
                    state.queue.push(...remaining);
                    playTrack(guildId, first);
                    return res.json({
                        status: 'playlist',
                        count: playlistTracks.length,
                        track: first
                    });
                } else {
                    state.queue.push(...playlistTracks);
                    return res.json({
                        status: 'playlist_queued',
                        count: playlistTracks.length,
                        track: playlistTracks[0]
                    });
                }
            }
        }

        // 0. Detectar se é uma música local da pasta /musicas
        const localTracks = getLocalMusicFiles();
        let matchedLocal = null;

        if (lowerQuery.startsWith('local:')) {
            const cleanTarget = query.slice(6).trim().toLowerCase();
            matchedLocal = localTracks.find(t => 
                t.filename.toLowerCase() === cleanTarget || 
                path.basename(t.filename, t.ext).toLowerCase() === cleanTarget
            );
        } else if (!query.startsWith('http://') && !query.startsWith('https://')) {
            // Match exato por nome de arquivo ou sem extensão
            matchedLocal = localTracks.find(t => 
                t.filename.toLowerCase() === lowerQuery ||
                path.basename(t.filename, t.ext).toLowerCase() === lowerQuery
            );
            // Se não encontrou exato, busca por substring
            if (!matchedLocal) {
                matchedLocal = localTracks.find(t => 
                    t.filename.toLowerCase().includes(lowerQuery) ||
                    t.title.toLowerCase().includes(lowerQuery) ||
                    lowerQuery.includes(path.basename(t.filename, t.ext).toLowerCase())
                );
            }
        }

        if (matchedLocal) {
            console.log(`[AudioBridge] Faixa local identificada: ${matchedLocal.filename}`);
            trackInfo = {
                title: `${matchedLocal.emoji} ${matchedLocal.title}`,
                url: matchedLocal.filePath,
                duration: matchedLocal.duration || 'Local',
                requester: requester || 'Mestre',
                isPreset: false,
                isLocal: true,
                filename: matchedLocal.filename
            };
        } else if (AMBIENT_PRESETS[lowerQuery]) {
            trackInfo = {
                ...AMBIENT_PRESETS[lowerQuery],
                requester: requester || 'Mestre'
            };
        } else if (query.startsWith('http://') || query.startsWith('https://')) {
            // Extrair título e duração via yt-dlp
            let title = 'Áudio Web';
            let duration = 'Online';
            try {
                const infoProc = spawn('yt-dlp', [
                    '--no-warnings',
                    '--print', '%(title)s|%(id)s|%(duration_string)s',
                    '--no-playlist',
                    query
                ]);
                let outData = '';
                infoProc.stdout.on('data', (d) => outData += d.toString());
                await new Promise((resolve) => infoProc.on('close', resolve));
                const line = outData.trim().split('\n')[0];
                if (line) {
                    const parts = line.split('|');
                    if (parts[0] && parts[0].trim()) title = parts[0].trim();
                    if (parts[2] && parts[2].trim() && parts[2].trim() !== 'NA') duration = parts[2].trim();
                }
            } catch (_) {}

            trackInfo = {
                title,
                url: query,
                duration,
                requester: requester || 'Jogador',
                isPreset: false
            };
        } else {
            // Busca no YouTube via yt-dlp search
            let title = query;
            let targetUrl = `ytsearch1:${query}`;
            let duration = 'Desconhecida';

            try {
                const infoProc = spawn('yt-dlp', [
                    '--no-warnings',
                    '--print', '%(title)s|%(id)s|%(duration_string)s',
                    '--no-playlist',
                    `ytsearch1:${query}`
                ]);
                let outData = '';
                infoProc.stdout.on('data', (d) => outData += d.toString());
                await new Promise((resolve) => infoProc.on('close', resolve));
                const line = outData.trim().split('\n')[0];
                if (line) {
                    const parts = line.split('|');
                    if (parts[0] && parts[0].trim()) title = parts[0].trim();
                    const videoId = parts[1] ? parts[1].trim() : '';
                    if (parts[2] && parts[2].trim() && parts[2].trim() !== 'NA') duration = parts[2].trim();
                    if (videoId) {
                        targetUrl = `https://www.youtube.com/watch?v=${videoId}`;
                    }
                }
            } catch (err) {
                console.error('[AudioBridge] Erro na busca do YouTube:', err);
            }

            trackInfo = {
                title,
                url: targetUrl,
                duration,
                requester: requester || 'Jogador',
                isPreset: false
            };
        }

        if (state.isPlaying) {
            state.queue.push(trackInfo);
            return res.json({
                status: 'queued',
                position: state.queue.length,
                track: trackInfo
            });
        } else {
            playTrack(guildId, trackInfo);
            return res.json({
                status: 'playing',
                position: 0,
                track: trackInfo
            });
        }
    } catch (err) {
        console.error('[AudioBridge] Erro no /play:', err);
        return res.status(500).json({ error: err.message });
    }
});

// Stop
app.post('/stop', (req, res) => {
    const { guildId } = req.body;
    const state = guildStates.get(guildId);
    if (state) {
        state.queue = [];
        state.current = null;
        state.isPlaying = false;
        cleanupActiveProcesses(state);
        state.player.stop(true);
    }
    res.json({ status: 'stopped' });
});

// Skip
app.post('/skip', (req, res) => {
    const { guildId } = req.body;
    const state = guildStates.get(guildId);
    if (state && state.isPlaying) {
        cleanupActiveProcesses(state);
        state.isSkipping = true;
        if (state.queue.length > 0) {
            const nextTrack = state.queue.shift();
            // Se loop de fila estiver ativo, a faixa que estava tocando vai para o fim da fila
            if (state.loopMode === 'queue' && state.current) {
                state.queue.push(state.current);
            }
            playTrack(guildId, nextTrack);
            state.isSkipping = false;
            return res.json({ status: 'skipped', hasNext: true });
        } else {
            state.isPlaying = false;
            state.current = null;
            state.player.stop(true);
            state.isSkipping = false;
            return res.json({ status: 'skipped', hasNext: false });
        }
    }
    res.json({ status: 'idle' });
});

// Loop (Música Atual / Fila / Desativado)
app.post('/loop', (req, res) => {
    const { guildId, mode } = req.body;
    const state = guildStates.get(guildId);
    if (!state) {
        return res.status(404).json({ error: 'Nenhuma reprodução ativa para este servidor' });
    }

    if (mode === 'track' || mode === 'musica') {
        state.loopMode = 'track';
    } else if (mode === 'queue' || mode === 'fila') {
        state.loopMode = 'queue';
    } else if (mode === 'off' || mode === 'desativar') {
        state.loopMode = 'off';
    } else {
        // Alternância cíclica: off -> track -> queue -> off
        if (state.loopMode === 'off') {
            state.loopMode = 'track';
        } else if (state.loopMode === 'track') {
            state.loopMode = 'queue';
        } else {
            state.loopMode = 'off';
        }
    }

    console.log(`[AudioBridge] Modo de loop alterado para ${state.loopMode} na guild ${guildId}`);
    res.json({ loopMode: state.loopMode });
});

// Leave
app.post('/leave', (req, res) => {
    const { guildId } = req.body;
    const state = guildStates.get(guildId);
    if (state) {
        state.queue = [];
        state.current = null;
        state.isPlaying = false;
        state.loopMode = 'off';
        cleanupActiveProcesses(state);
        state.player.stop(true);
        if (state.connection) {
            state.connection.destroy();
            state.connection = null;
        }
        guildStates.delete(guildId);
    }
    res.json({ status: 'disconnected' });
});

// Queue
app.get('/queue/:guildId', (req, res) => {
    const state = guildStates.get(req.params.guildId);
    if (!state) {
        return res.json({ current: null, queue: [], isPlaying: false, loopMode: 'off' });
    }
    res.json({
        current: state.current,
        queue: state.queue,
        isPlaying: state.isPlaying,
        loopMode: state.loopMode || 'off'
    });
});

// ═══════════════════════════════════════════════════════════════
// ═══════════════ GRAVAÇÃO DE SESSÃO DE RPG ═══════════════════
// ═══════════════════════════════════════════════════════════════

// Reconecta ao canal de voz alterando selfDeaf sem destruir a conexão
async function reconnectForRecording(guildId, shouldRecord) {
    const state = guildStates.get(guildId);
    if (!state || !state.connection) return;

    try {
        const channelId = state.connection.joinConfig.channelId;
        state.connection.rejoin({
            channelId,
            selfDeaf: !shouldRecord,
            selfMute: false,
        });
        console.log(`[Recorder] selfDeaf alterado para ${!shouldRecord} na guild ${guildId}`);
    } catch (err) {
        console.error(`[Recorder] Erro ao reconectar (rejoin):`, err.message);
        // Fallback: destruir e reconectar
        try {
            const guild = await client.guilds.fetch(guildId);
            const channelId = state.connection.joinConfig.channelId;
            state.connection.destroy();
            state.connection = joinVoiceChannel({
                channelId,
                guildId,
                adapterCreator: guild.voiceAdapterCreator,
                selfDeaf: !shouldRecord,
                selfMute: false
            });
            state.connection.subscribe(state.player);
            console.log(`[Recorder] Reconexão completa (fallback) na guild ${guildId}`);
        } catch (fallbackErr) {
            console.error(`[Recorder] Falha no fallback de reconexão:`, fallbackErr.message);
        }
    }
}

// Atualiza o JSON de metadados com o participante
function updateSessionParticipant(sessionDir, userId, username) {
    const metaPath = path.join(sessionDir, 'session.json');
    try {
        const meta = JSON.parse(fs.readFileSync(metaPath, 'utf-8'));
        if (!meta.participants[userId]) {
            meta.participants[userId] = {
                username: username || userId,
                joinedAt: new Date().toISOString(),
            };
            fs.writeFileSync(metaPath, JSON.stringify(meta, null, 2));
        }
    } catch (_) {}
}

// Inicia a captura de áudio de todos os usuários no canal de voz
function startRecording(guildId) {
    const state = guildStates.get(guildId);
    if (!state || !state.connection) throw new Error('Sem conexão de voz ativa');

    const receiver = state.connection.receiver;
    const sessionId = new Date().toISOString().replace(/[:.]/g, '-');
    const sessionDir = path.join(RECORDINGS_DIR, `${guildId}_${sessionId}`);
    fs.mkdirSync(sessionDir, { recursive: true });

    state.isRecording = true;
    state.recordingDir = sessionDir;
    state.recordingStartTime = Date.now();

    // Salvar metadados iniciais da sessão
    fs.writeFileSync(path.join(sessionDir, 'session.json'), JSON.stringify({
        guildId,
        startTime: new Date().toISOString(),
        participants: {},
    }, null, 2));

    // Handler para capturar cada usuário que fala
    const speakingHandler = (userId) => {
        // Ignorar se já temos um stream ativo para este usuário
        if (state.recordingStreams.has(userId)) return;

        console.log(`[Recorder] Usuário ${userId} começou a falar na guild ${guildId}`);

        const opusStream = receiver.subscribe(userId, {
            end: { behavior: EndBehaviorType.AfterSilence, duration: 2000 },
        });

        const userFile = path.join(sessionDir, `${userId}.pcm`);
        const writeStream = createWriteStream(userFile, { flags: 'a' });

        const decoder = new prism.opus.Decoder({
            rate: 48000,
            channels: 2,
            frameSize: 960,
        });

        pipeline(opusStream, decoder, writeStream, (err) => {
            if (err && err.code !== 'ERR_STREAM_PREMATURE_CLOSE') {
                console.error(`[Recorder] Erro no pipeline do user ${userId}:`, err.message);
            }
            state.recordingStreams.delete(userId);
        });

        state.recordingStreams.set(userId, { writeStream, decoder, opusStream });

        // Buscar nome legível do usuário no Discord (apelido no servidor ou username)
        client.guilds.fetch(guildId).then(g => g.members.fetch(userId)).then(m => {
            const name = m.displayName || m.user.username;
            updateSessionParticipant(sessionDir, userId, name);
        }).catch(() => {
            updateSessionParticipant(sessionDir, userId, userId);
        });
    };

    receiver.speaking.on('start', speakingHandler);
    state.speakingHandler = speakingHandler;

    console.log(`[Recorder] Gravação iniciada na guild ${guildId} → ${sessionDir}`);
    return { sessionDir, sessionId };
}

// Para a gravação e fecha todos os streams
function stopRecording(guildId) {
    const state = guildStates.get(guildId);
    if (!state || !state.isRecording) return null;

    state.isRecording = false;
    const sessionDir = state.recordingDir;
    const duration = Date.now() - state.recordingStartTime;

    // Fechar todos os streams de gravação
    for (const [userId, streams] of state.recordingStreams) {
        try {
            streams.opusStream.destroy();
            streams.decoder.destroy();
            streams.writeStream.end();
        } catch (_) {}
    }
    state.recordingStreams.clear();

    // Remover o handler de speaking
    if (state.speakingHandler && state.connection && state.connection.receiver) {
        state.connection.receiver.speaking.removeListener('start', state.speakingHandler);
    }
    state.speakingHandler = null;
    state.recordingDir = null;
    state.recordingStartTime = null;

    // Atualizar metadados com hora de fim
    const metaPath = path.join(sessionDir, 'session.json');
    try {
        const meta = JSON.parse(fs.readFileSync(metaPath, 'utf-8'));
        meta.endTime = new Date().toISOString();
        meta.durationMs = duration;
        fs.writeFileSync(metaPath, JSON.stringify(meta, null, 2));
    } catch (_) {}

    console.log(`[Recorder] Gravação encerrada na guild ${guildId} — duração: ${Math.round(duration / 1000)}s`);
    return { sessionDir, durationMs: duration };
}

// Converte cada arquivo .pcm para .ogg via ffmpeg
async function convertPcmToOgg(sessionDir) {
    const files = fs.readdirSync(sessionDir).filter(f => f.endsWith('.pcm'));
    const results = [];

    for (const file of files) {
        const pcmPath = path.join(sessionDir, file);
        const oggPath = pcmPath.replace('.pcm', '.ogg');

        // Verificar se o PCM tem conteúdo
        const stats = fs.statSync(pcmPath);
        if (stats.size === 0) {
            fs.unlinkSync(pcmPath);
            continue;
        }

        await new Promise((resolve, reject) => {
            const proc = spawn('ffmpeg', [
                '-y',
                '-f', 's16le', '-ar', '48000', '-ac', '2',
                '-i', pcmPath,
                '-c:a', 'libopus', '-b:a', '64k',
                oggPath
            ]);
            proc.on('close', (code) => {
                if (code === 0) {
                    fs.unlinkSync(pcmPath); // apagar PCM bruto
                    results.push(oggPath);
                    resolve();
                } else {
                    console.error(`[Recorder] ffmpeg falhou ao converter ${file} (exit ${code})`);
                    resolve(); // continuar mesmo se falhar
                }
            });
            proc.on('error', (err) => {
                console.error(`[Recorder] ffmpeg não encontrado:`, err.message);
                resolve();
            });
        });
    }
    return results;
}

// Mescla todas as faixas OGG em um único arquivo
async function mergeAllTracks(sessionDir) {
    const oggFiles = fs.readdirSync(sessionDir).filter(f => f.endsWith('.ogg') && f !== 'sessao_completa.ogg');
    if (oggFiles.length === 0) return null;
    if (oggFiles.length === 1) {
        // Apenas 1 participante, copiar direto
        const src = path.join(sessionDir, oggFiles[0]);
        const dest = path.join(sessionDir, 'sessao_completa.ogg');
        fs.copyFileSync(src, dest);
        return dest;
    }

    const mergedPath = path.join(sessionDir, 'sessao_completa.ogg');
    const inputs = oggFiles.flatMap(f => ['-i', path.join(sessionDir, f)]);

    await new Promise((resolve, reject) => {
        const proc = spawn('ffmpeg', [
            '-y',
            ...inputs,
            '-filter_complex', `amix=inputs=${oggFiles.length}:duration=longest`,
            '-c:a', 'libopus', '-b:a', '96k',
            mergedPath
        ]);
        proc.on('close', (code) => {
            if (code === 0) {
                resolve();
            } else {
                console.error(`[Recorder] ffmpeg merge falhou (exit ${code})`);
                resolve(); // continuar sem merge
            }
        });
        proc.on('error', (err) => {
            console.error(`[Recorder] ffmpeg não encontrado para merge:`, err.message);
            resolve();
        });
    });

    return fs.existsSync(mergedPath) ? mergedPath : null;
}

// ══════════ ENDPOINTS REST DE GRAVAÇÃO ══════════

// POST /record/start - Inicia gravação da sessão
app.post('/record/start', async (req, res) => {
    const { guildId, channelId } = req.body;
    if (!guildId || !channelId) {
        return res.status(400).json({ error: 'guildId e channelId são obrigatórios' });
    }

    try {
        const state = getOrCreateState(guildId);

        if (state.isRecording) {
            return res.status(409).json({ error: 'Já existe uma gravação em andamento neste servidor' });
        }

        // Garantir conexão ao canal de voz (com selfDeaf: false para poder escutar)
        if (!state.connection || state.connection.state.status === VoiceConnectionStatus.Destroyed) {
            const guild = await client.guilds.fetch(guildId);
            state.connection = joinVoiceChannel({
                channelId,
                guildId,
                adapterCreator: guild.voiceAdapterCreator,
                selfDeaf: false,
                selfMute: false
            });
            state.connection.on('stateChange', (oldState, newState) => {
                console.log(`[Recorder] Conexão de voz (guild ${guildId}): ${oldState.status} → ${newState.status}`);
            });
            state.connection.subscribe(state.player);
        } else {
            // Já conectado — reconectar sem surdez para poder gravar
            await reconnectForRecording(guildId, true);
        }

        const result = startRecording(guildId);
        res.json({ status: 'recording', ...result });
    } catch (err) {
        console.error('[Recorder] Erro ao iniciar gravação:', err);
        res.status(500).json({ error: err.message });
    }
});

// Executa a transcrição local da sessão via Python (faster-whisper multitrack)
async function transcribeSession(sessionDir) {
    return new Promise((resolve) => {
        const scriptPath = path.resolve(__dirname, '../transcribe_session.py');
        if (!fs.existsSync(scriptPath)) {
            console.warn('[Transcriber] transcribe_session.py não encontrado');
            return resolve(null);
        }

        console.log(`[Transcriber] Iniciando transcrição local para ${sessionDir}...`);
        const proc = spawn('python', [scriptPath, '--session-dir', sessionDir]);

        proc.stdout.on('data', (d) => {
            console.log(`[Transcriber] ${d.toString().trim()}`);
        });

        proc.stderr.on('data', (d) => {
            console.error(`[Transcriber Err] ${d.toString().trim()}`);
        });

        proc.on('close', (code) => {
            if (code === 0) {
                const mdFile = path.join(sessionDir, 'transcricao.md');
                if (fs.existsSync(mdFile)) {
                    console.log(`[Transcriber] Transcrição concluída: ${mdFile}`);
                    return resolve(mdFile);
                }
            }
            console.warn(`[Transcriber] Processo finalizou com código ${code}`);
            resolve(null);
        });

        proc.on('error', (err) => {
            console.error('[Transcriber] Falha ao executar python:', err.message);
            resolve(null);
        });
    });
}

// POST /record/stop - Para a gravação e converte os arquivos
app.post('/record/stop', async (req, res) => {
    const { guildId } = req.body;

    try {
        const result = stopRecording(guildId);
        if (!result) {
            return res.status(404).json({ error: 'Nenhuma gravação ativa neste servidor' });
        }

        // Reconectar como surdo novamente (economiza banda)
        await reconnectForRecording(guildId, false);

        // Converter PCM → OGG
        const oggFiles = await convertPcmToOgg(result.sessionDir);
        const mergedFile = await mergeAllTracks(result.sessionDir);

        // Executar transcrição automática com Whisper local
        const transcriptionFile = await transcribeSession(result.sessionDir);

        res.json({
            status: 'stopped',
            sessionDir: result.sessionDir,
            durationMs: result.durationMs,
            files: oggFiles,
            mergedFile,
            transcriptionFile,
        });
    } catch (err) {
        console.error('[Recorder] Erro ao parar gravação:', err);
        res.status(500).json({ error: err.message });
    }
});

// GET /record/status/:guildId - Status da gravação atual
app.get('/record/status/:guildId', (req, res) => {
    const state = guildStates.get(req.params.guildId);
    if (!state || !state.isRecording) {
        return res.json({ isRecording: false });
    }
    res.json({
        isRecording: true,
        sessionDir: state.recordingDir,
        durationMs: Date.now() - state.recordingStartTime,
        participants: state.recordingStreams.size,
    });
});

client.once('clientReady', () => {
    console.log(`[AudioBridge] Conectado ao Discord como ${client.user.tag} (DAVE E2EE Pronto)`);
});

client.login(TOKEN).catch((err) => {
    console.error('[AudioBridge] Falha ao logar no Discord:', err);
    process.exit(1);
});
