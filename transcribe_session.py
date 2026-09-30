#!/usr/bin/env python3
"""
transcribe_session.py
Transcreve os áudios de uma sessão de RPG gravada pelo Kenobi Bot.
Utiliza faster-whisper localmente (sem APIs externas, 100% grátis e offline).
Aproveita a arquitetura multitrack para identificar com precisão quem está falando.
"""

import os
import sys
import json
import argparse
import glob
from pathlib import Path

# Configurar saída do console para UTF-8 no Windows
if hasattr(sys.stdout, "reconfigure"):
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")
if hasattr(sys.stderr, "reconfigure"):
    sys.stderr.reconfigure(encoding="utf-8", errors="replace")

def format_timestamp(seconds: float) -> str:
    total_sec = int(round(seconds))
    mins = total_sec // 60
    secs = total_sec % 60
    hours = mins // 60
    if hours > 0:
        return f"{hours:02d}:{mins%60:02d}:{secs:02d}"
    return f"{mins:02d}:{secs:02d}"

import subprocess
import numpy as np

def load_audio(file_path: str, sr: int = 16000) -> np.ndarray:
    """Decodifica qualquer formato de áudio para PCM float32 16kHz mono via ffmpeg"""
    cmd = [
        "ffmpeg",
        "-nostdin",
        "-threads", "0",
        "-i", str(file_path),
        "-f", "s16le",
        "-ac", "1",
        "-ar", str(sr),
        "-"
    ]
    out = subprocess.check_output(cmd, stderr=subprocess.DEVNULL)
    return np.frombuffer(out, np.int16).flatten().astype(np.float32) / 32768.0

def load_session_metadata(session_dir: Path) -> dict:
    meta_path = session_dir / "session.json"
    if meta_path.exists():
        try:
            with open(meta_path, "r", encoding="utf-8") as f:
                return json.load(f)
        except Exception as e:
            print(f"[Aviso] Falha ao ler session.json: {e}", file=sys.stderr)
    return {}

def transcribe_session(session_dir_str: str, model_size: str = "base"):
    session_dir = Path(session_dir_str).resolve()
    if not session_dir.exists():
        print(f"Erro: Pasta da sessão não encontrada: {session_dir}", file=sys.stderr)
        sys.exit(1)

    meta = load_session_metadata(session_dir)
    participants = meta.get("participants", {})

    # Mapear userId -> Nome do Jogador
    user_names = {}
    for uid, pdata in participants.items():
        if isinstance(pdata, dict):
            user_names[uid] = pdata.get("username", uid)
        elif isinstance(pdata, str):
            user_names[uid] = pdata

    # Encontrar faixas individuais (.ogg) excluindo sessao_completa.ogg
    ogg_files = [f for f in session_dir.glob("*.ogg") if f.name != "sessao_completa.ogg"]

    print(f"Iniciando transcrição na pasta: {session_dir}")
    print(f"Carregando modelo Faster-Whisper '{model_size}' (CPU / int8)...")

    from faster_whisper import WhisperModel
    # int8 na CPU é rápido, leve e tem ótima precisão
    model = WhisperModel(model_size, device="cpu", compute_type="int8")

    all_segments = []

    if ogg_files:
        print(f"Encontradas {len(ogg_files)} faixa(s) individual(is) de jogadores.")
        for ogg in ogg_files:
            uid = ogg.stem
            speaker_name = user_names.get(uid, f"Jogador_{uid[:4]}")
            print(f"Transcrevendo fala de [{speaker_name}] ({ogg.name})...")

            try:
                audio_data = load_audio(str(ogg))
                segments, info = model.transcribe(
                    audio_data,
                    language="pt",
                    beam_size=5,
                    vad_filter=True, # filtra silêncios
                    vad_parameters=dict(min_silence_duration_ms=500)
                )

                for seg in segments:
                    text = seg.text.strip()
                    if text:
                        all_segments.append({
                            "speaker": speaker_name,
                            "user_id": uid,
                            "start": seg.start,
                            "end": seg.end,
                            "text": text
                        })
            except Exception as e:
                print(f"[Erro] Falha ao transcrever {ogg.name}: {e}", file=sys.stderr)

    else:
        # Se só tiver sessao_completa.ogg
        fallback = session_dir / "sessao_completa.ogg"
        if fallback.exists():
            print("Nenhuma faixa individual encontrada, transcrevendo sessao_completa.ogg...")
            audio_data = load_audio(str(fallback))
            segments, info = model.transcribe(
                audio_data,
                language="pt",
                beam_size=5,
                vad_filter=True
            )
            for seg in segments:
                text = seg.text.strip()
                if text:
                    all_segments.append({
                        "speaker": "Mesa",
                        "user_id": "all",
                        "start": seg.start,
                        "end": seg.end,
                        "text": text
                    })
        else:
            print("Nenhum arquivo .ogg encontrado para transcrição.", file=sys.stderr)
            sys.exit(1)

    if not all_segments:
        print("Nenhuma fala detectada nos áudios gravados.")
        # Criar arquivo indicando que não houve fala audível
        md_path = session_dir / "transcricao.md"
        with open(md_path, "w", encoding="utf-8") as f:
            f.write("# 📜 Transcrição da Sessão de RPG\n\n*Nenhuma fala audível detectada durante a gravação.*")
        return

    # Ordenar todos os segmentos cronologicamente pelo timestamp de início
    all_segments.sort(key=lambda s: s["start"])

    # Agrupar falas consecutivas do mesmo falante se o intervalo for pequeno (< 2.5s)
    grouped = []
    for seg in all_segments:
        if grouped and grouped[-1]["speaker"] == seg["speaker"] and (seg["start"] - grouped[-1]["end"]) <= 2.5:
            grouped[-1]["text"] += " " + seg["text"]
            grouped[-1]["end"] = seg["end"]
        else:
            grouped.append(dict(seg))

    # Gerar transcricao.md (Markdown rico para leitura e Discord)
    md_lines = [
        "# 📜 Transcrição da Sessão de RPG",
        "",
        f"**Pasta:** `{session_dir.name}`",
    ]
    if meta.get("startTime"):
        md_lines.append(f"**Data de Início:** {meta.get('startTime')}")
    if meta.get("durationMs"):
        dur_sec = meta.get("durationMs") / 1000
        md_lines.append(f"**Duração:** {format_timestamp(dur_sec)}")
    
    unique_speakers = sorted(list(set(s["speaker"] for s in grouped)))
    md_lines.append(f"**Participantes:** {', '.join(unique_speakers)}")
    md_lines.append("")
    md_lines.append("---")
    md_lines.append("")

    # Gerar transcricao.txt (Texto puro)
    txt_lines = [
        f"=== Transcricao da Sessao: {session_dir.name} ===",
        f"Participantes: {', '.join(unique_speakers)}",
        "--------------------------------------------------",
        ""
    ]

    for item in grouped:
        time_tag = format_timestamp(item["start"])
        spk = item["speaker"]
        txt = item["text"]

        md_lines.append(f"`[{time_tag}]` **{spk}:** {txt}\n")
        txt_lines.append(f"[{time_tag}] {spk}: {txt}")

    # Salvar arquivos na pasta da sessão
    md_file = session_dir / "transcricao.md"
    txt_file = session_dir / "transcricao.txt"
    json_file = session_dir / "transcricao.json"

    with open(md_file, "w", encoding="utf-8") as f:
        f.write("\n".join(md_lines))

    with open(txt_file, "w", encoding="utf-8") as f:
        f.write("\n".join(txt_lines))

    with open(json_file, "w", encoding="utf-8") as f:
        json.dump({
            "sessionDir": str(session_dir),
            "speakers": unique_speakers,
            "turns": grouped
        }, f, ensure_ascii=False, indent=2)

    # Atualizar session.json com referência da transcrição
    meta["transcription"] = "transcricao.md"
    meta_path = session_dir / "session.json"
    with open(meta_path, "w", encoding="utf-8") as f:
        json.dump(meta, f, ensure_ascii=False, indent=2)

    print(f"\n✅ Transcrição concluída com sucesso!")
    print(f"📄 Arquivo Markdown: {md_file}")
    print(f"📄 Arquivo Texto:    {txt_file}")
    print(f"📄 Arquivo JSON:     {json_file}")
    print(f"\n--- Prévia da Transcrição ---")
    for line in txt_lines[4:14]:
        print(line)

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Transcreve uma sessão de RPG gravada")
    parser.add_argument("--session-dir", required=True, help="Caminho da pasta da sessão")
    parser.add_argument("--model", default="base", help="Tamanho do modelo (tiny, base, small)")
    args = parser.parse_args()

    transcribe_session(args.session_dir, args.model)
