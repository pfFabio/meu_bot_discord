import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';

import {
  writeRecordingState,
  readRecordingState,
  clearRecordingState,
  recoverPendingRecordings,
  shouldWarnForRecordingSilence,
  isRecordingHeartbeatStale,
} from './recording_recovery.js';

test('gravação pendente é persistida e recuperada após reinício', () => {
  const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'rec-'));

  const sessionDir = path.join(tempDir, 'guild-123_session');
  fs.mkdirSync(sessionDir, { recursive: true });
  fs.writeFileSync(path.join(sessionDir, 'session.json'), JSON.stringify({ guildId: 'guild-123', participants: {} }));

  writeRecordingState({
    guildId: 'guild-123',
    sessionDir,
    channelId: 'voice-chan-1',
    recordingStartTime: 1700000000000,
    participants: { user1: { username: 'gandalf' } },
  }, tempDir);

  const persisted = readRecordingState('guild-123', tempDir);
  assert.ok(persisted);
  assert.equal(persisted.guildId, 'guild-123');
  assert.equal(persisted.sessionDir, sessionDir);

  const recovered = recoverPendingRecordings(tempDir);
  assert.equal(recovered.length, 1);
  assert.equal(recovered[0].guildId, 'guild-123');

  const cleared = clearRecordingState('guild-123', tempDir);
  assert.equal(cleared, true);
  assert.equal(readRecordingState('guild-123', tempDir), null);

  fs.rmSync(tempDir, { recursive: true, force: true });
});

test('avisa quando a gravação fica inativa por 1 minuto sem áudio', () => {
  const state = {
    isRecording: true,
    lastAudioActivityAt: Date.now() - 65000,
    recordingSilenceWarningSentAt: null,
  };

  assert.equal(shouldWarnForRecordingSilence(state), true);

  state.recordingSilenceWarningSentAt = Date.now();
  assert.equal(shouldWarnForRecordingSilence(state), false);

  state.recordingSilenceWarningSentAt = null;
  state.isRecording = false;
  assert.equal(shouldWarnForRecordingSilence(state), false);
});

test('detecta heartbeat estagnado e exige recuperação da gravação', () => {
  const state = {
    isRecording: true,
    lastHeartbeat: Date.now() - 120000,
    lastAudioActivityAt: Date.now() - 120000,
  };

  assert.equal(isRecordingHeartbeatStale(state), true);
  assert.equal(isRecordingHeartbeatStale({ ...state, isRecording: false }), false);
});
