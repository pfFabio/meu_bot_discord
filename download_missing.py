import os
import sys
import json
import time
import subprocess

if hasattr(sys.stdout, "reconfigure"):
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")

TEMAS_DIR = os.path.abspath("temas")
CATALOG_PATH = os.path.join(TEMAS_DIR, "temas_catalog.json")

def main():
    with open(CATALOG_PATH, "r", encoding="utf-8") as f:
        tracks = json.load(f)

    missing = []
    for t in tracks:
        p = os.path.join(TEMAS_DIR, t["filename"])
        if not (os.path.exists(p) and os.path.getsize(p) > 50 * 1024):
            missing.append(t)

    print(f"Total de faixas restantes para baixar: {len(missing)}")
    if not missing:
        print("Todas as faixas já estão baixadas!")
        return

    for idx, t in enumerate(missing, 1):
        filename = t["filename"]
        out_template = os.path.join(TEMAS_DIR, f"{os.path.splitext(filename)[0]}.%(ext)s")
        print(f"[{idx}/{len(missing)}] Baixando com Node runtime: {t['emoji']} {t['title']} -> {filename}...")

        cmd = [
            "yt-dlp",
            "--js-runtimes", "node",
            "-f", "ba",
            "-x",
            "--audio-format", "mp3",
            "--audio-quality", "128K",
            "-N", "4",
            "--no-playlist",
            "-o", out_template,
            t["youtube_url"]
        ]

        res = subprocess.run(cmd, capture_output=True, text=True)
        final_path = os.path.join(TEMAS_DIR, filename)
        if res.returncode == 0 and os.path.exists(final_path):
            print(f"  ✅ Concluído: {filename} ({os.path.getsize(final_path) // (1024*1024)} MB)")
        else:
            print(f"  ❌ Falha em {filename}: {res.stderr[-200:]}")
        time.sleep(2)

    print("Processamento das faixas pendentes concluído!")

if __name__ == "__main__":
    main()
