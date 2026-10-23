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
const AMBIENCE_DIR = path.resolve(__dirname, '../ambiente');
const TEMAS_DIR = path.resolve(__dirname, '../temas');
const RECORDINGS_DIR = path.resolve(__dirname, '../gravacoes');
if (!fs.existsSync(MUSIC_DIR)) fs.mkdirSync(MUSIC_DIR, { recursive: true });
if (!fs.existsSync(AMBIENCE_DIR)) fs.mkdirSync(AMBIENCE_DIR, { recursive: true });
if (!fs.existsSync(TEMAS_DIR)) fs.mkdirSync(TEMAS_DIR, { recursive: true });
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

// Catálogo em memória e cenários de ambiente de Curse of Strahd
let themesCatalogCache = null;
function getThemesCatalog() {
    if (themesCatalogCache && themesCatalogCache.length > 0) return themesCatalogCache;
    try {
        const catPath = path.join(TEMAS_DIR, 'temas_catalog.json');
        if (fs.existsSync(catPath)) {
            themesCatalogCache = JSON.parse(fs.readFileSync(catPath, 'utf8'));
            return themesCatalogCache;
        }
    } catch (e) {
        console.error('[AudioBridge] Erro ao carregar temas_catalog.json:', e);
    }
    return [];
}

const AMBIENCE_SCENARIOS = {
    'cos - svalich woods.mp3': { title: 'Floresta Svalich (Svalich Woods)', emoji: '🌲', key: 'svalich', category: 'Exploração & Regiões' },
    'cos - village of barovia.mp3': { title: 'Vila de Baróvia (Village of Barovia)', emoji: '🏚️', key: 'barovia_village', category: 'Povoados & Civilização' },
    'cos - town of vallaki.mp3': { title: 'Cidade de Vallaki (Town of Vallaki)', emoji: '🏘️', key: 'vallaki', category: 'Povoados & Civilização' },
    'cos - village of krezk.mp3': { title: 'Vila de Krezk (Village of Krezk)', emoji: '🏡', key: 'krezk', category: 'Povoados & Civilização' },
    'cos - castle ravenloft.mp3': { title: 'Castelo Ravenloft (Castle Ravenloft)', emoji: '🏰', key: 'castle_ravenloft', category: 'Masmorras & Covis' },
    'cos - catacombs.mp3': { title: 'Catacumbas de Ravenloft (Catacombs)', emoji: '⚰️', key: 'catacombs', category: 'Masmorras & Covis' },
    'cos - death house.mp3': { title: 'Casa da Morte (Death House)', emoji: '🏚️', key: 'death_house', category: 'Masmorras & Covis' },
    'cos - death house dungeon.mp3': { title: 'Masmorra da Casa da Morte (Death House Dungeon)', emoji: '💀', key: 'death_house_dungeon', category: 'Masmorras & Covis' },
    'cos - old bonegrinder.mp3': { title: 'Velho Moedor de Ossos (Old Bonegrinder)', emoji: '🦴', key: 'old_bonegrinder', category: 'Marcos Notáveis' },
    'cos - argynvostholt.mp3': { title: 'Argynvostholt (Mansão do Dragão)', emoji: '🛡️', key: 'argynvostholt', category: 'Marcos Notáveis' },
    'cos - abbey of st markovia.mp3': { title: 'Abadia de Santa Markovia (Abbey of St. Markovia)', emoji: '⛪', key: 'abbey', category: 'Marcos Notáveis' },
    'cos - madhouse.mp3': { title: 'Asilo dos Loucos (Madhouse)', emoji: '🤪', key: 'madhouse', category: 'Marcos Notáveis' },
    'cos - tsolenka pass.mp3': { title: 'Passo de Tsolenka (Tsolenka Pass)', emoji: '❄️', key: 'tsolenka_pass', category: 'Exploração & Regiões' },
    'cos - van richten\'s tower.mp3': { title: 'Torre de Van Richten (Van Richten\'s Tower)', emoji: '🗼', key: 'van_richtens_tower', category: 'Marcos Notáveis' },
    'cos - ruins of berez.mp3': { title: 'Ruínas de Berez (Ruins of Berez)', emoji: '🌊', key: 'ruins_of_berez', category: 'Exploração & Regiões' },
    'cos - wizard of the wines.mp3': { title: 'Mago dos Vinhos (Wizard of the Wines)', emoji: '🍇', key: 'wizard_of_wines', category: 'Marcos Notáveis' },
    'cos - amber temple.mp3': { title: 'Templo de Âmbar (Amber Temple)', emoji: '🏛️', key: 'amber_temple', category: 'Masmorras & Covis' },
    'cos - mists of ravenloft.mp3': { title: 'Névoas de Ravenloft (Mists of Ravenloft)', emoji: '🌫️', key: 'mists_of_ravenloft', category: 'Exploração & Regiões' },
};

// Varre um diretório em busca de faixas de áudio válidas
function scanDirectoryForAudio(targetDir) {
    if (!fs.existsSync(targetDir)) return [];
    const validExts = ['.mp3', '.ogg', '.wav', '.flac', '.m4a', '.aac'];
    try {
        const files = fs.readdirSync(targetDir);
        const isTemas = (path.resolve(targetDir) === path.resolve(TEMAS_DIR));
        const isAmbience = (path.resolve(targetDir) === path.resolve(AMBIENCE_DIR));
        const catalog = isTemas ? getThemesCatalog() : [];

        return files
            .filter(file => {
                const ext = path.extname(file).toLowerCase();
                if (!validExts.includes(ext)) return false;
                const base = path.basename(file, ext);
                // Evitar ler arquivos parciais que ainda estão sendo gravados pelo ffmpeg/yt-dlp
                if (files.includes(`${base}.webm`) || files.includes(`${base}.webm.part`) || files.includes(`${file}.part`)) {
                    return false;
                }
                return true;
            })
            .map(file => {
                const ext = path.extname(file);
                const baseName = path.basename(file, ext);
                const lowerFile = file.toLowerCase();
                const lowerBase = baseName.toLowerCase();

                let cleanTitle = baseName;
                let emoji = '🎵';
                let key = lowerBase.replace(/[^a-z0-9]+/g, '_');
                let category = 'Geral';
                let duration = durationCache.get(file) || 'Local';

                if (isTemas) {
                    category = 'Baróvia (Temas)';
                    emoji = '🎭';
                    // Tenta achar metadados ricos em temas_catalog.json
                    const catEntry = catalog.find(t => 
                        t.filename.toLowerCase() === lowerFile ||
                        path.basename(t.filename, path.extname(t.filename)).toLowerCase() === lowerBase
                    );
                    if (catEntry) {
                        cleanTitle = catEntry.title;
                        emoji = catEntry.emoji || '🎭';
                        key = catEntry.key || key;
                        category = catEntry.category || category;
                        if (catEntry.duration) duration = catEntry.duration;
                    } else {
                        cleanTitle = baseName.replace(/^CoS\s*-\s*/i, '').replace(/^D&D Ambience\s*-\s*\[CoS\]\s*-\s*/i, '').trim();
                        if (lowerBase.includes('strahd')) emoji = '🩸';
                        else if (lowerBase.includes('battle') || lowerBase.includes('fight') || lowerBase.includes('combate')) emoji = '⚔️';
                    }
                } else if (isAmbience) {
                    category = 'Cenário & Atmosfera';
                    emoji = '🏰';
                    const ambEntry = AMBIENCE_SCENARIOS[lowerFile];
                    if (ambEntry) {
                        cleanTitle = ambEntry.title;
                        emoji = ambEntry.emoji;
                        key = ambEntry.key;
                        category = ambEntry.category;
                    } else {
                        cleanTitle = baseName.replace(/^CoS\s*-\s*/i, '').trim();
                    }
                } else {
                    category = 'Pasta musicas';
                    cleanTitle = baseName.replace(/^CoS\s*-\s*/i, '').trim();
                    if (lowerBase.includes('vistani')) emoji = '🎪';
                    else if (lowerBase.includes('mago')) emoji = '🧙‍♂️';
                }

                const filePath = path.join(targetDir, file);

                return {
                    filename: file,
                    title: cleanTitle,
                    key,
                    emoji,
                    filePath,
                    ext,
                    duration,
                    dir: targetDir,
                    category,
                    isLocal: true
                };
            });
    } catch (err) {
        console.error(`[AudioBridge] Erro ao ler pasta ${targetDir}:`, err);
        return [];
    }
}

// Retorna faixas da pasta /musicas (músicas locais)
function getLocalMusicFiles() {
    return scanDirectoryForAudio(MUSIC_DIR);
}

// Retorna faixas da pasta /ambiente (ambientes góticos de Curse of Strahd)
function getLocalAmbienceFiles() {
    return scanDirectoryForAudio(AMBIENCE_DIR);
}

// Retorna faixas da pasta /temas (temas e trilhas sonoras de Curse of Strahd)
function getLocalThemesFiles() {
    return scanDirectoryForAudio(TEMAS_DIR);
}

// Retorna todas as faixas locais unificadas (músicas + ambientes + temas)
function getAllLocalAudioFiles() {
    return [...getLocalMusicFiles(), ...getLocalAmbienceFiles(), ...getLocalThemesFiles()];
}

// Resolve qualquer query contra o acervo local físico com prioridade máxima
function resolveLocalAudio(query) {
    if (!query) return null;
    const cleanQuery = query.trim();
    if (cleanQuery.startsWith('http://') || cleanQuery.startsWith('https://')) {
        return null; // URLs externas não são locais
    }

    const localTracks = getAllLocalAudioFiles();
    const rawTarget = cleanQuery.replace(/^(local:|theme:|ambience:)/i, '').trim();
    const lowerTarget = rawTarget.toLowerCase();

    // 1. Caminho de arquivo absoluto existente
    if (fs.existsSync(rawTarget) || fs.existsSync(cleanQuery)) {
        const fullPath = fs.existsSync(rawTarget) ? path.resolve(rawTarget) : path.resolve(cleanQuery);
        const found = localTracks.find(t => path.resolve(t.filePath).toLowerCase() === fullPath.toLowerCase());
        if (found) return found;
        const ext = path.extname(fullPath);
        const baseName = path.basename(fullPath, ext);
        return {
            filename: path.basename(fullPath),
            title: baseName,
            key: baseName.toLowerCase().replace(/[^a-z0-9]+/g, '_'),
            emoji: '🎵',
            filePath: fullPath,
            ext,
            duration: durationCache.get(path.basename(fullPath)) || 'Local',
            dir: path.dirname(fullPath),
            isLocal: true
        };
    }

    // 2. Match exato por nome de arquivo (com extensão)
    let match = localTracks.find(t => t.filename.toLowerCase() === lowerTarget);
    if (match) return match;

    // 3. Match exato por nome de arquivo sem extensão
    match = localTracks.find(t => path.basename(t.filename, t.ext).toLowerCase() === lowerTarget);
    if (match) return match;

    // 4. Match por key exata (ex: "into_the_mists", "mists_of_ravenloft", "strahd_von_zarovich")
    match = localTracks.find(t => t.key && t.key.toLowerCase() === lowerTarget);
    if (match) return match;

    // 5. Match transformando underline em espaço e vice-versa
    const spaceTarget = lowerTarget.replace(/_/g, ' ');
    const underTarget = lowerTarget.replace(/\s+/g, '_');
    match = localTracks.find(t => 
        (t.key && (t.key.toLowerCase() === underTarget || t.key.toLowerCase() === spaceTarget)) ||
        path.basename(t.filename, t.ext).toLowerCase() === spaceTarget ||
        path.basename(t.filename, t.ext).toLowerCase() === `cos - ${spaceTarget}`
    );
    if (match) return match;

    // 6. Match por título exato
    match = localTracks.find(t => t.title.toLowerCase() === lowerTarget || t.title.toLowerCase() === spaceTarget);
    if (match) return match;

    // 7. Match no catálogo de temas (temas_catalog.json) verificando arquivo físico no disco
    const catalog = getThemesCatalog();
    const catEntry = catalog.find(t => 
        t.key.toLowerCase() === lowerTarget ||
        t.key.toLowerCase() === underTarget ||
        t.filename.toLowerCase() === lowerTarget ||
        t.title.toLowerCase() === lowerTarget ||
        t.title.toLowerCase() === spaceTarget
    );
    if (catEntry) {
        const themePath = path.join(TEMAS_DIR, catEntry.filename);
        if (fs.existsSync(themePath)) {
            return {
                filename: catEntry.filename,
                title: catEntry.title,
                key: catEntry.key,
                emoji: catEntry.emoji || '🎭',
                filePath: themePath,
                ext: path.extname(catEntry.filename),
                duration: catEntry.duration || 'Local',
                dir: TEMAS_DIR,
                isLocal: true
            };
        }
    }

    // 8. Match no mapa de cenários de ambiente (CoS)
    for (const [ambFile, ambInfo] of Object.entries(AMBIENCE_SCENARIOS)) {
        if (lowerTarget === ambInfo.key || lowerTarget === ambInfo.title.toLowerCase() || lowerTarget === ambFile) {
            const ambPath = path.join(AMBIENCE_DIR, ambFile);
            if (fs.existsSync(ambPath)) {
                return {
                    filename: ambFile,
                    title: ambInfo.title,
                    key: ambInfo.key,
                    emoji: ambInfo.emoji || '🏰',
                    filePath: ambPath,
                    ext: path.extname(ambFile),
                    duration: durationCache.get(ambFile) || 'Local',
                    dir: AMBIENCE_DIR,
                    isLocal: true
                };
            }
        }
    }

    // 9. Match por substring mais específica (mínimo 4 caracteres para evitar falsos positivos)
    if (lowerTarget.length >= 4) {
        match = localTracks.find(t => 
            path.basename(t.filename, t.ext).toLowerCase().includes(lowerTarget) ||
            t.title.toLowerCase().includes(lowerTarget) ||
            (t.key && t.key.toLowerCase().includes(lowerTarget))
        );
        if (match) return match;
    }

    return null;
}

// Inicializa verificação de duração em segundo plano
function probeLocalDurations() {
    const tracks = getAllLocalAudioFiles();
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
                    } else if (state.loopMode !== 'off' && finishedTrack) {
                        playTrack(guildId, finishedTrack);
                    } else {
                        state.isPlaying = false;
                        state.current = null;
                    }
                } else if (state.loopMode === 'track') {
                    // Se novas músicas foram adicionadas à fila durante o loop de faixa:
                    // As novas faixas entram no ciclo do loop e transicionam para loop de fila
                    if (state.queue.length > 0) {
                        console.log(`[AudioBridge] Músicas adicionadas à fila durante loop na guild ${guildId}. Integrando novas faixas ao ciclo de loop...`);
                        state.loopMode = 'queue';
                        if (finishedTrack) {
                            state.queue.push(finishedTrack);
                        }
                        const nextTrack = state.queue.shift();
                        playTrack(guildId, nextTrack);
                    } else if (finishedTrack) {
                        console.log(`[AudioBridge] Repetindo faixa em loop na guild ${guildId}: ${finishedTrack.title}`);
                        playTrack(guildId, finishedTrack);
                    } else {
                        state.isPlaying = false;
                        state.current = null;
                    }
                } else if (state.loopMode === 'queue') {
                    if (finishedTrack) {
                        state.queue.push(finishedTrack);
                    }
                    if (state.queue.length > 0) {
                        const nextTrack = state.queue.shift();
                        console.log(`[AudioBridge] Próxima faixa do loop de fila na guild ${guildId}: ${nextTrack.title} (${state.queue.length} restantes na fila)`);
                        playTrack(guildId, nextTrack);
                    } else if (finishedTrack) {
                        playTrack(guildId, finishedTrack);
                    } else {
                        state.isPlaying = false;
                        state.current = null;
                    }
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
            const state = guildStates.get(guildId);
            if (state) state.isPaused = false;
        });

        player.on(AudioPlayerStatus.Paused, () => {
            console.log(`[AudioBridge] Player está PAUSADO na guild ${guildId}`);
            const state = guildStates.get(guildId);
            if (state) state.isPaused = true;
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
                    state.isPaused = false;
                    state.current = null;
                }
            }
        });

        guildStates.set(guildId, {
            guildId,
            connection: null,
            player,
            queue: [],
            history: [],
            current: null,
            isPlaying: false,
            isPaused: false,
            volume: 1.0,
            resource: null,
            trackStartTime: null,
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

// Fade out suave da faixa atual
async function fadeOutCurrentTrack(state, durationMs = 1500) {
    if (!state || !state.isPlaying || !state.resource || !state.resource.volume) {
        return;
    }
    const targetVol = (typeof state.volume === 'number') ? state.volume : 1.0;
    const steps = 15;
    const stepTime = Math.max(15, Math.floor(durationMs / steps));
    for (let i = 1; i <= steps; i++) {
        await new Promise(r => setTimeout(r, stepTime));
        const factor = (steps - i) / steps;
        try {
            if (state.resource && state.resource.volume) {
                state.resource.volume.setVolume(targetVol * factor);
            }
        } catch (_) {}
    }
}

// Fade in suave da nova faixa
async function fadeInNewTrack(state, durationMs = 1200, targetVol = 1.0) {
    const steps = 15;
    const stepTime = Math.max(15, Math.floor(durationMs / steps));
    for (let i = 1; i <= steps; i++) {
        await new Promise(r => setTimeout(r, stepTime));
        const factor = i / steps;
        try {
            if (state.resource && state.resource.volume) {
                state.resource.volume.setVolume(targetVol * factor);
            }
        } catch (_) {}
    }
}

async function playTrack(guildId, track, options = {}) {
    const state = getOrCreateState(guildId);

    // Se foi solicitada transição suave e já está tocando algo
    if (options.transition && state.isPlaying && state.resource) {
        console.log(`[AudioBridge] Realizando transição suave (fade-out) na guild ${guildId}...`);
        await fadeOutCurrentTrack(state, options.fadeDuration || 1500);
    }

    if (state.current && state.current !== track) {
        if (!state.history) state.history = [];
        state.history.push(state.current);
        if (state.history.length > 50) state.history.shift();
    }
    state.current = track;
    state.isPlaying = true;
    state.isPaused = false;
    state.trackStartTime = Date.now();
    cleanupActiveProcesses(state);

    try {
        console.log(`[AudioBridge] Iniciando reprodução na guild ${guildId}: ${track.title} (Transição: ${Boolean(options.transition)})`);

        let streamUrl = track.url;
        const isLocalFile = (track.isLocal || fs.existsSync(track.url)) && fs.existsSync(track.url);
        // Se não for arquivo local existente ou preset/stream direto, extrai stream HTTPS do YouTube via yt-dlp
        if (!isLocalFile && !track.isPreset && !track.url.includes('zeno.fm')) {
            console.log(`[AudioBridge] Extraindo stream direto para: ${track.url}`);
            streamUrl = await extractDirectStreamUrl(track.url);
        }

        console.log(`[AudioBridge] FFmpeg conectando ao áudio (${isLocalFile ? 'ARQUIVO LOCAL' : 'STREAM WEB'}): ${streamUrl.substring(0, 80)}...`);

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

        state.resource = resource;
        const currentVol = (typeof state.volume === 'number') ? state.volume : 1.0;

        if (resource.volume) {
            if (options.transition) {
                // Inicia com volume zerado e sobe gradualmente (fade-in)
                resource.volume.setVolume(0.0);
                state.player.play(resource);
                fadeInNewTrack(state, options.fadeDuration || 1200, currentVol);
            } else {
                resource.volume.setVolume(currentVol);
                state.player.play(resource);
            }
        } else {
            state.player.play(resource);
        }

        if (state.connection) {
            state.connection.subscribe(state.player);
        }
    } catch (err) {
        console.error(`[AudioBridge] Falha ao iniciar reprodução:`, err);
        state.isPlaying = false;
        state.current = null;
    }
}

// Mapeamento dos Cenários de Curse of Strahd para fallback de streaming imediato
const COS_TRACK_YOUTUBE_MAP = {
    "svalich": { id: "eePl-I8heFc", title: "Floresta Svalich (Svalich Woods)", filename: "CoS - Svalich Woods.mp3", emoji: "🌲" },
    "death house dungeon": { id: "eqUiXz1tcmM", title: "Masmorra da Casa da Morte (Death House Dungeon)", filename: "CoS - Death House Dungeon.mp3", emoji: "💀" },
    "death house": { id: "GOTyzCntzJo", title: "Casa da Morte (Death House)", filename: "CoS - Death House.mp3", emoji: "🏚️" },
    "village of barovia": { id: "nnEGIH1mXwI", title: "Vila de Baróvia (Village of Barovia)", filename: "CoS - Village of Barovia.mp3", emoji: "🏚️" },
    "barovia": { id: "nnEGIH1mXwI", title: "Vila de Baróvia (Village of Barovia)", filename: "CoS - Village of Barovia.mp3", emoji: "🏚️" },
    "vallaki": { id: "2UPkwe_5p_w", title: "Cidade de Vallaki (Town of Vallaki)", filename: "CoS - Town Of Vallaki.mp3", emoji: "🏘️" },
    "bonegrinder": { id: "J63qJ_GFV_k", title: "Velho Moedor de Ossos (Old Bonegrinder)", filename: "CoS - Old Bonegrinder.mp3", emoji: "🦴" },
    "argynvostholt": { id: "K08IQhgXDWQ", title: "Argynvostholt (Mansão do Dragão)", filename: "CoS - Argynvostholt.mp3", emoji: "🛡️" },
    "castle ravenloft": { id: "rfA-pzDkERs", title: "Castelo Ravenloft (Castle Ravenloft)", filename: "CoS - Castle Ravenloft.mp3", emoji: "🏰" },
    "ravenloft": { id: "rfA-pzDkERs", title: "Castelo Ravenloft (Castle Ravenloft)", filename: "CoS - Castle Ravenloft.mp3", emoji: "🏰" },
    "catacombs": { id: "WPpVMmTt74Q", title: "Catacumbas de Ravenloft (Catacombs)", filename: "CoS - Catacombs.mp3", emoji: "⚰️" },
    "krezk": { id: "ZRnLm7BJMnw", title: "Vila de Krezk (Village of Krezk)", filename: "CoS - Village of Krezk.mp3", emoji: "🏡" },
    "markovia": { id: "8AODR94oNMo", title: "Abadia de Santa Markovia (Abbey of St. Markovia)", filename: "CoS - Abbey of St Markovia.mp3", emoji: "⛪" },
    "abbey": { id: "8AODR94oNMo", title: "Abadia de Santa Markovia (Abbey of St. Markovia)", filename: "CoS - Abbey of St Markovia.mp3", emoji: "⛪" },
    "madhouse": { id: "1gDfJHT5n0U", title: "Asilo dos Loucos (Madhouse)", filename: "CoS - Madhouse.mp3", emoji: "🤪" },
    "tsolenka": { id: "L_eIOeXIEfo", title: "Passo de Tsolenka (Tsolenka Pass)", filename: "CoS - Tsolenka Pass.mp3", emoji: "❄️" },
    "van richten": { id: "1u7rJTtlbKw", title: "Torre de Van Richten (Van Richten's Tower)", filename: "CoS - Van Richten's Tower.mp3", emoji: "🗼" },
    "berez": { id: "9MsH9vN1aFQ", title: "Ruínas de Berez (Ruins of Berez)", filename: "CoS - Ruins of Berez.mp3", emoji: "🌊" },
    "wines": { id: "4hT8zvR-nkg", title: "Mago dos Vinhos (Wizard of the Wines)", filename: "CoS - Wizard Of The Wines.mp3", emoji: "🍇" },
    "amber temple": { id: "Dd91LPwFAMk", title: "Templo de Âmbar (Amber Temple)", filename: "CoS - Amber Temple.mp3", emoji: "🏛️" },
    "mists": { id: "J41ipvUvIzI", title: "Névoas de Ravenloft (Mists of Ravenloft)", filename: "CoS - Mists of Ravenloft.mp3", emoji: "🌫️" },
};

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

// Lista de Temas de Curse of Strahd da pasta /temas
app.get('/temas-tracks', (req, res) => {
    const tracks = getLocalThemesFiles().map(t => ({
        filename: t.filename,
        title: t.title,
        emoji: t.emoji,
        duration: t.duration,
        path: t.filePath
    }));
    res.json({ tracks });
});

// Lista de Ambientes de Curse of Strahd da pasta /ambiente
app.get('/ambience-tracks', (req, res) => {
    const tracks = getLocalAmbienceFiles().map(t => ({
        filename: t.filename,
        title: t.title,
        emoji: t.emoji,
        duration: t.duration,
        path: t.filePath,
        category: t.category
    }));
    res.json({ tracks });
});

// Conectar ao canal de voz
app.post('/join', async (req, res) => {
    const { guildId, channelId } = req.body;
    if (!guildId || !channelId) {
        return res.status(400).json({ error: 'guildId e channelId são obrigatórios' });
    }
    try {
        const guild = await client.guilds.fetch(guildId);
        if (!guild) {
            return res.status(404).json({ error: 'Servidor não encontrado' });
        }
        const state = getOrCreateState(guildId);
        if (!state.connection || state.connection.state.status === VoiceConnectionStatus.Destroyed) {
            state.connection = joinVoiceChannel({
                channelId,
                guildId,
                adapterCreator: guild.voiceAdapterCreator,
                selfDeaf: !state.isRecording,
                selfMute: false
            });
            state.connection.subscribe(state.player);
        } else {
            state.connection.rejoin({
                channelId,
                selfDeaf: !state.isRecording,
                selfMute: false
            });
        }
        res.json({ status: 'connected', guildId, channelId });
    } catch (err) {
        console.error('[AudioBridge] Erro ao entrar no canal de voz:', err);
        res.status(500).json({ error: err.message });
    }
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

        // 0. Resolução de áudio local com prioridade absoluta
        const matchedLocal = resolveLocalAudio(query);

        if (matchedLocal) {
            console.log(`[AudioBridge] Faixa local identificada com SUCESSO: ${matchedLocal.filename} (${matchedLocal.filePath})`);
            trackInfo = {
                title: `${matchedLocal.emoji} ${matchedLocal.title}`,
                url: matchedLocal.filePath,
                duration: matchedLocal.duration || durationCache.get(matchedLocal.filename) || 'Local',
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
        } else {
            // Verificar se corresponde a um cenário de Curse of Strahd (checa se o arquivo local existe antes de qualquer streaming)
            let matchedCoS = null;
            for (const [key, cosInfo] of Object.entries(COS_TRACK_YOUTUBE_MAP)) {
                if (lowerQuery === key || lowerQuery === key.replace(/\s+/g, '_') || (cosInfo.filename && lowerQuery === cosInfo.filename.toLowerCase())) {
                    matchedCoS = cosInfo;
                    break;
                }
            }

            if (matchedCoS) {
                // Checa primeiro se o arquivo existe na pasta ambiente/ ou temas/ ou musicas/!
                const possiblePaths = [
                    path.join(AMBIENCE_DIR, matchedCoS.filename),
                    path.join(TEMAS_DIR, matchedCoS.filename),
                    path.join(MUSIC_DIR, matchedCoS.filename)
                ];
                const localFound = possiblePaths.find(p => fs.existsSync(p));
                if (localFound) {
                    console.log(`[AudioBridge] Cenário CoS encontrado no disco: ${localFound}`);
                    trackInfo = {
                        title: `${matchedCoS.emoji} ${matchedCoS.title}`,
                        url: localFound,
                        duration: durationCache.get(matchedCoS.filename) || 'Local',
                        requester: requester || 'Mestre',
                        isPreset: false,
                        isLocal: true,
                        filename: matchedCoS.filename
                    };
                } else {
                    console.log(`[AudioBridge] Cenário de Baróvia identificado (streaming YouTube fallback): ${matchedCoS.title}`);
                    trackInfo = {
                        title: `${matchedCoS.emoji} ${matchedCoS.title}`,
                        url: `https://www.youtube.com/watch?v=${matchedCoS.id}`,
                        duration: 'Online (CoS)',
                        requester: requester || 'Mestre',
                        isPreset: false,
                        isLocal: false
                    };
                }
            } else if (lowerQuery.startsWith('local:') || lowerQuery.startsWith('theme:') || lowerQuery.startsWith('ambience:')) {
                const cleanTarget = query.replace(/^(local:|theme:|ambience:)/i, '').trim();
                const lowerTarget = cleanTarget.toLowerCase();

                // Busca no catálogo de temas para checar arquivo físico e fallback online
                const catalog = getThemesCatalog();
                const matchedTheme = catalog.find(t => 
                    t.filename.toLowerCase() === lowerTarget ||
                    t.key.toLowerCase() === lowerTarget ||
                    t.title.toLowerCase() === lowerTarget
                );

                if (matchedTheme) {
                    const themePath = path.join(TEMAS_DIR, matchedTheme.filename);
                    if (fs.existsSync(themePath)) {
                        console.log(`[AudioBridge] Tema de Baróvia carregado do disco: ${themePath}`);
                        trackInfo = {
                            title: `${matchedTheme.emoji} ${matchedTheme.title}`,
                            url: themePath,
                            duration: matchedTheme.duration || 'Local',
                            requester: requester || 'Mestre',
                            isPreset: false,
                            isLocal: true,
                            filename: matchedTheme.filename
                        };
                    } else if (matchedTheme.youtube_url) {
                        console.log(`[AudioBridge] Tema de Baróvia identificado no catálogo (streaming YouTube fallback): ${matchedTheme.title}`);
                        trackInfo = {
                            title: `${matchedTheme.emoji} ${matchedTheme.title}`,
                            url: matchedTheme.youtube_url,
                            duration: matchedTheme.duration || 'Online (Tema CoS)',
                            requester: requester || 'Mestre',
                            isPreset: false,
                            isLocal: false
                        };
                    }
                }

                if (!trackInfo) {
                    return res.status(404).json({ error: `Faixa local "${cleanTarget}" não foi encontrada nas pastas ambiente, temas ou musicas.` });
                }
            }
        }

        if (!trackInfo) {
            if (query.startsWith('http://') || query.startsWith('https://')) {
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
        }

        if (state.isPlaying) {
            if (req.body.transition) {
                // Transição suave imediata substituindo a faixa atual com crossfade
                // Se o modo de loop estiver ativo, mantém a faixa anterior na fila para continuar no loop
                if (state.loopMode !== 'off' && state.current) {
                    state.queue.push(state.current);
                    if (state.loopMode === 'track') {
                        state.loopMode = 'queue';
                    }
                }
                await playTrack(guildId, trackInfo, {
                    transition: true,
                    fadeDuration: req.body.fadeDuration || 1500
                });
                return res.json({
                    status: 'playing',
                    transition: true,
                    position: 0,
                    track: trackInfo
                });
            } else if (req.body.playNext) {
                // Inserir no topo da fila como a próxima música a tocar
                state.queue.unshift(trackInfo);
                return res.json({
                    status: 'next',
                    position: 1,
                    track: trackInfo
                });
            } else {
                state.queue.push(trackInfo);
                return res.json({
                    status: 'queued',
                    position: state.queue.length,
                    track: trackInfo
                });
            }
        } else {
            await playTrack(guildId, trackInfo, {
                transition: Boolean(req.body.transition),
                fadeDuration: req.body.fadeDuration || 1200
            });
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
        state.isPaused = false;
        cleanupActiveProcesses(state);
        state.player.stop(true);
    }
    res.json({ status: 'stopped' });
});

// Pause
app.post('/pause', (req, res) => {
    const { guildId } = req.body;
    const state = guildStates.get(guildId);
    if (state && state.player) {
        const success = state.player.pause();
        state.isPaused = true;
        console.log(`[AudioBridge] Reprodução pausada na guild ${guildId}`);
        return res.json({ status: 'paused', isPaused: true, success });
    }
    res.status(404).json({ error: 'Nenhuma reprodução ativa para este servidor' });
});

// Resume
app.post('/resume', (req, res) => {
    const { guildId } = req.body;
    const state = guildStates.get(guildId);
    if (state && state.player) {
        const success = state.player.unpause();
        state.isPaused = false;
        console.log(`[AudioBridge] Reprodução retomada na guild ${guildId}`);
        return res.json({ status: 'resumed', isPaused: false, success });
    }
    res.status(404).json({ error: 'Nenhuma reprodução ativa para este servidor' });
});

// Skip (Avançar música) com suporte a transição suave opcional
app.post('/skip', async (req, res) => {
    const { guildId, transition, fadeDuration } = req.body;
    const state = guildStates.get(guildId);
    if (state && state.isPlaying) {
        state.isSkipping = true;
        if (transition && state.resource) {
            console.log(`[AudioBridge] Skip com transição suave na guild ${guildId}...`);
            await fadeOutCurrentTrack(state, fadeDuration || 1500);
        }
        cleanupActiveProcesses(state);
        if (state.current) {
            if (!state.history) state.history = [];
            state.history.push(state.current);
            if (state.history.length > 50) state.history.shift();
        }
        if (state.queue.length > 0) {
            const nextTrack = state.queue.shift();
            // Se loop de fila ou faixa estiver ativo, a faixa que estava tocando vai para o fim da fila para continuar no loop
            if ((state.loopMode === 'queue' || state.loopMode === 'track') && state.current) {
                state.queue.push(state.current);
                if (state.loopMode === 'track') {
                    state.loopMode = 'queue';
                }
            }
            playTrack(guildId, nextTrack, {
                transition: Boolean(transition),
                fadeDuration: fadeDuration || 1200
            });
            state.isSkipping = false;
            return res.json({ status: 'skipped', hasNext: true, track: nextTrack });
        } else if (state.loopMode !== 'off' && state.current) {
            // Se em loop sem outras faixas na fila, reinicia a reprodução da faixa atual no loop
            playTrack(guildId, state.current, {
                transition: Boolean(transition),
                fadeDuration: fadeDuration || 1200
            });
            state.isSkipping = false;
            return res.json({ status: 'skipped', hasNext: true, track: state.current });
        } else {
            state.isPlaying = false;
            state.isPaused = false;
            state.current = null;
            state.player.stop(true);
            state.isSkipping = false;
            return res.json({ status: 'skipped', hasNext: false });
        }
    }
    res.json({ status: 'idle' });
});

// Previous (Voltar música)
app.post('/previous', (req, res) => {
    const { guildId } = req.body;
    const state = guildStates.get(guildId);
    if (!state) {
        return res.status(404).json({ error: 'Nenhuma reprodução ativa para este servidor' });
    }
    cleanupActiveProcesses(state);
    if (state.history && state.history.length > 0) {
        const prevTrack = state.history.pop();
        if (state.current) {
            state.queue.unshift(state.current);
        }
        playTrack(guildId, prevTrack);
        return res.json({ status: 'previous', track: prevTrack });
    } else if (state.current) {
        // Se histórico estiver vazio, reinicia a música atual
        playTrack(guildId, state.current);
        return res.json({ status: 'restarted', track: state.current });
    }
    res.json({ status: 'idle' });
});

// Volume (Ajustar volume de 0.0 a 2.0)
app.post('/volume', (req, res) => {
    const { guildId, volume } = req.body;
    const state = guildStates.get(guildId);
    if (!state) {
        return res.status(404).json({ error: 'Nenhuma reprodução ativa para este servidor' });
    }
    const vol = parseFloat(volume);
    if (isNaN(vol) || vol < 0 || vol > 3.0) {
        return res.status(400).json({ error: 'Volume inválido (use um valor entre 0.0 e 3.0)' });
    }
    state.volume = vol;
    if (state.resource && state.resource.volume) {
        state.resource.volume.setVolume(vol);
    }
    console.log(`[AudioBridge] Volume definido para ${(vol * 100).toFixed(0)}% na guild ${guildId}`);
    res.json({ status: 'ok', volume: vol });
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
        // Alternância cíclica natural: off -> queue -> track -> off
        if (state.loopMode === 'off') {
            state.loopMode = 'queue';
        } else if (state.loopMode === 'queue') {
            state.loopMode = 'track';
        } else {
            state.loopMode = 'off';
        }
    }

    console.log(`[AudioBridge] Modo de loop alterado para ${state.loopMode} na guild ${guildId}`);
    res.json({ loopMode: state.loopMode });
});

// Queue Remove (Remover faixa da fila por índice)
app.post('/queue/remove', (req, res) => {
    const { guildId, index } = req.body;
    const state = guildStates.get(guildId);
    if (!state) {
        return res.status(404).json({ error: 'Nenhuma reprodução ativa para este servidor' });
    }
    const idx = parseInt(index, 10);
    if (isNaN(idx) || idx < 0 || idx >= state.queue.length) {
        return res.status(400).json({ error: 'Índice de fila inválido' });
    }
    const removed = state.queue.splice(idx, 1)[0];
    res.json({ status: 'removed', removed, queue: state.queue });
});

// Queue Clear (Limpar fila)
app.post('/queue/clear', (req, res) => {
    const { guildId } = req.body;
    const state = guildStates.get(guildId);
    if (state) {
        state.queue = [];
    }
    res.json({ status: 'cleared' });
});

// Queue Promote (Mover faixa para a primeira posição da fila - próxima a tocar)
app.post('/queue/promote', (req, res) => {
    const { guildId, index } = req.body;
    const state = guildStates.get(guildId);
    if (!state) {
        return res.status(404).json({ error: 'Nenhuma reprodução ativa para este servidor' });
    }
    const idx = parseInt(index, 10);
    if (isNaN(idx) || idx < 0 || idx >= state.queue.length) {
        return res.status(400).json({ error: 'Índice de fila inválido' });
    }
    const item = state.queue.splice(idx, 1)[0];
    state.queue.unshift(item);
    console.log(`[AudioBridge] Faixa promovida para o topo da fila na guild ${guildId}: ${item.title}`);
    res.json({ status: 'promoted', track: item, queue: state.queue });
});

// Queue Jump (Pular imediatamente para uma música específica da fila, com opção de transição suave)
app.post('/queue/jump', async (req, res) => {
    const { guildId, index, transition, fadeDuration } = req.body;
    const state = guildStates.get(guildId);
    if (!state) {
        return res.status(404).json({ error: 'Nenhuma reprodução ativa para este servidor' });
    }
    const idx = parseInt(index, 10);
    if (isNaN(idx) || idx < 0 || idx >= state.queue.length) {
        return res.status(400).json({ error: 'Índice de fila inválido' });
    }
    const targetTrack = state.queue.splice(idx, 1)[0];
    state.isSkipping = true;
    if (transition && state.isPlaying && state.resource) {
        console.log(`[AudioBridge] Jump com transição suave na guild ${guildId}...`);
        await fadeOutCurrentTrack(state, fadeDuration || 1500);
    }
    cleanupActiveProcesses(state);
    if (state.current) {
        if (!state.history) state.history = [];
        state.history.push(state.current);
        if (state.history.length > 50) state.history.shift();
        if (state.loopMode === 'queue' || state.loopMode === 'track') {
            state.queue.push(state.current);
            if (state.loopMode === 'track') {
                state.loopMode = 'queue';
            }
        }
    }
    playTrack(guildId, targetTrack, {
        transition: Boolean(transition),
        fadeDuration: fadeDuration || 1200
    });
    state.isSkipping = false;
    console.log(`[AudioBridge] Saltou para a faixa na guild ${guildId}: ${targetTrack.title} (Transição: ${Boolean(transition)})`);
    res.json({ status: 'jumped', track: targetTrack, queue: state.queue });
});

// Leave
app.post('/leave', (req, res) => {
    const { guildId } = req.body;
    const state = guildStates.get(guildId);
    if (state) {
        state.queue = [];
        state.current = null;
        state.isPlaying = false;
        state.isPaused = false;
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

// Queue & Status
app.get('/queue/:guildId', (req, res) => {
    const state = guildStates.get(req.params.guildId);
    if (!state) {
        return res.json({
            current: null,
            queue: [],
            history: [],
            isPlaying: false,
            isPaused: false,
            loopMode: 'off',
            volume: 1.0,
            playbackDurationMs: 0
        });
    }
    res.json({
        current: state.current,
        queue: state.queue,
        history: state.history || [],
        isPlaying: state.isPlaying,
        isPaused: Boolean(state.isPaused || (state.player && state.player.state.status === AudioPlayerStatus.Paused)),
        loopMode: state.loopMode || 'off',
        volume: state.volume !== undefined ? state.volume : 1.0,
        trackStartTime: state.trackStartTime || null,
        playbackDurationMs: state.resource ? state.resource.playbackDuration : 0
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

// Preenche arquivo PCM com silêncio (zeros) para manter alinhamento temporal exato
function padSilence(filePath, padBytes) {
    if (padBytes <= 0) return;
    const CHUNK_SIZE = 192000; // 1s de silêncio a 48kHz estéreo s16le (192.000 bytes)
    const fd = fs.openSync(filePath, 'a');
    let remaining = padBytes;
    while (remaining > 0) {
        const toWrite = Math.min(remaining, CHUNK_SIZE);
        fs.writeSync(fd, Buffer.alloc(toWrite));
        remaining -= toWrite;
    }
    fs.closeSync(fd);
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

        const userFile = path.join(sessionDir, `${userId}.pcm`);

        // Sincronização temporal: se o usuário começou a falar após o início da gravação
        // ou após uma pausa, preenche o intervalo com silêncio para sincronizar as faixas
        let currentSize = 0;
        if (fs.existsSync(userFile)) {
            currentSize = fs.statSync(userFile).size;
        }

        const elapsedMs = Date.now() - state.recordingStartTime;
        let expectedBytes = Math.floor((elapsedMs / 1000) * 192000);
        expectedBytes -= (expectedBytes % 4); // Alinhamento para amostras de 4 bytes (2ch * 16bits)

        const padBytes = expectedBytes - currentSize;
        if (padBytes > 0) {
            padSilence(userFile, padBytes);
        }

        const opusStream = receiver.subscribe(userId, {
            end: { behavior: EndBehaviorType.AfterSilence, duration: 2000 },
        });

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

    // Preencher o final de cada arquivo PCM para que todos tenham a duração total da sessão
    const totalBytes = Math.floor((duration / 1000) * 192000);
    try {
        const pcmFiles = fs.readdirSync(sessionDir).filter(f => f.endsWith('.pcm'));
        for (const pcmFile of pcmFiles) {
            const pcmPath = path.join(sessionDir, pcmFile);
            const currentSize = fs.statSync(pcmPath).size;
            let padBytes = totalBytes - currentSize;
            padBytes -= (padBytes % 4);
            if (padBytes > 0) {
                padSilence(pcmPath, padBytes);
            }
        }
    } catch (_) {}

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
        const proc = spawn('python', [scriptPath, '--session-dir', sessionDir, '--model', 'small']);

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
