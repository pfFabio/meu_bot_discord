import fs from 'fs';
import path from 'path';

export function getRecordingStatePath(guildId, recordingsDir = path.resolve(process.cwd(), 'gravacoes')) {
    return path.join(recordingsDir, `${guildId}.recording.json`);
}

export function readRecordingState(guildId, recordingsDir = path.resolve(process.cwd(), 'gravacoes')) {
    const filePath = getRecordingStatePath(guildId, recordingsDir);
    if (!fs.existsSync(filePath)) {
        return null;
    }

    try {
        const raw = fs.readFileSync(filePath, 'utf8');
        const parsed = JSON.parse(raw);
        if (!parsed || !parsed.guildId || !parsed.sessionDir) {
            return null;
        }
        return parsed;
    } catch (error) {
        console.warn(`[Recorder] Estado de gravação corrompido para guild ${guildId}:`, error.message);
        return null;
    }
}

export function writeRecordingState({ guildId, sessionDir, channelId = null, recordingStartTime = Date.now(), participants = {} }, recordingsDir = path.resolve(process.cwd(), 'gravacoes')) {
    const filePath = getRecordingStatePath(guildId, recordingsDir);
    const payload = {
        guildId,
        sessionDir,
        channelId,
        recordingStartTime,
        lastHeartbeat: Date.now(),
        updatedAt: new Date().toISOString(),
        participants,
        status: 'recording',
    };

    fs.mkdirSync(recordingsDir, { recursive: true });
    fs.writeFileSync(filePath, JSON.stringify(payload, null, 2));
    return payload;
}

export function touchRecordingState(guildId, overrides = {}, recordingsDir = path.resolve(process.cwd(), 'gravacoes')) {
    const current = readRecordingState(guildId, recordingsDir) || {};
    const next = {
        ...current,
        ...overrides,
        guildId,
        lastHeartbeat: Date.now(),
        updatedAt: new Date().toISOString(),
    };

    if (!next.sessionDir) {
        return null;
    }

    const filePath = getRecordingStatePath(guildId, recordingsDir);
    fs.writeFileSync(filePath, JSON.stringify(next, null, 2));
    return next;
}

export function clearRecordingState(guildId, recordingsDir = path.resolve(process.cwd(), 'gravacoes')) {
    const filePath = getRecordingStatePath(guildId, recordingsDir);
    if (!fs.existsSync(filePath)) {
        return false;
    }

    fs.unlinkSync(filePath);
    return true;
}

export function getRecordingCheckpointPath(guildId, recordingsDir = path.resolve(process.cwd(), 'gravacoes')) {
    return path.join(recordingsDir, `${guildId}.recording.checkpoint.json`);
}

export function writeRecordingCheckpoint(guildId, payload, recordingsDir = path.resolve(process.cwd(), 'gravacoes')) {
    const filePath = getRecordingCheckpointPath(guildId, recordingsDir);
    fs.mkdirSync(recordingsDir, { recursive: true });
    const checkpoint = {
        ...payload,
        guildId,
        lastHeartbeat: Date.now(),
        updatedAt: new Date().toISOString(),
        status: 'recording',
    };
    fs.writeFileSync(filePath, JSON.stringify(checkpoint, null, 2));
    return checkpoint;
}

export function clearRecordingCheckpoint(guildId, recordingsDir = path.resolve(process.cwd(), 'gravacoes')) {
    const filePath = getRecordingCheckpointPath(guildId, recordingsDir);
    if (!fs.existsSync(filePath)) {
        return false;
    }
    fs.unlinkSync(filePath);
    return true;
}

export function shouldWarnForRecordingSilence(state, now = Date.now()) {
    if (!state || !state.isRecording || !state.lastAudioActivityAt) {
        return false;
    }

    const inactivityMs = now - state.lastAudioActivityAt;
    if (inactivityMs < 60000) {
        return false;
    }

    if (state.recordingSilenceWarningSentAt && now - state.recordingSilenceWarningSentAt < 60000) {
        return false;
    }

    return true;
}

export function isRecordingHeartbeatStale(state, now = Date.now(), thresholdMs = 90000) {
    if (!state || !state.isRecording || !state.lastHeartbeat) {
        return false;
    }
    return now - state.lastHeartbeat > thresholdMs;
}

export function recoverPendingRecordings(recordingsDir = path.resolve(process.cwd(), 'gravacoes')) {
    if (!fs.existsSync(recordingsDir)) {
        return [];
    }

    const files = fs.readdirSync(recordingsDir)
        .filter((name) => name.endsWith('.recording.json'))
        .sort();

    return files
        .map((fileName) => {
            const filePath = path.join(recordingsDir, fileName);
            try {
                const parsed = JSON.parse(fs.readFileSync(filePath, 'utf8'));
                if (!parsed || !parsed.guildId || !parsed.sessionDir) {
                    return null;
                }
                return parsed;
            } catch (error) {
                console.warn(`[Recorder] Falha ao ler sessão pendente ${filePath}:`, error.message);
                return null;
            }
        })
        .filter(Boolean);
}
