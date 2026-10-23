import os
import subprocess
import concurrent.futures

import sys

# Garante suporte UTF-8 no terminal Windows
if hasattr(sys.stdout, "reconfigure"):
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")

AMBIENCE_DIR = os.path.abspath("ambiente")

TRACKS = [
    ("eePl-I8heFc", "CoS - Svalich Woods.mp3"),
    ("GOTyzCntzJo", "CoS - Death House.mp3"),
    ("eqUiXz1tcmM", "CoS - Death House Dungeon.mp3"),
    ("nnEGIH1mXwI", "CoS - Village of Barovia.mp3"),
    ("2UPkwe_5p_w", "CoS - Town Of Vallaki.mp3"),
    ("J63qJ_GFV_k", "CoS - Old Bonegrinder.mp3"),
    ("K08IQhgXDWQ", "CoS - Argynvostholt.mp3"),
    ("rfA-pzDkERs", "CoS - Castle Ravenloft.mp3"),
    ("WPpVMmTt74Q", "CoS - Catacombs.mp3"),
    ("ZRnLm7BJMnw", "CoS - Village of Krezk.mp3"),
    ("8AODR94oNMo", "CoS - Abbey of St Markovia.mp3"),
    ("1gDfJHT5n0U", "CoS - Madhouse.mp3"),
    ("L_eIOeXIEfo", "CoS - Tsolenka Pass.mp3"),
    ("1u7rJTtlbKw", "CoS - Van Richten's Tower.mp3"),
    ("9MsH9vN1aFQ", "CoS - Ruins of Berez.mp3"),
    ("4hT8zvR-nkg", "CoS - Wizard Of The Wines.mp3"),
    ("Dd91LPwFAMk", "CoS - Amber Temple.mp3"),
    ("J41ipvUvIzI", "CoS - Mists of Ravenloft.mp3"),
]

def download_track(video_id, filename):
    out_path = os.path.join(AMBIENCE_DIR, filename)
    if os.path.exists(out_path) and os.path.getsize(out_path) > 10 * 1024 * 1024:
        print(f"[OK] Já existe: {filename} ({os.path.getsize(out_path) // (1024*1024)} MB)")
        return True

    print(f"[BAIXANDO] {filename} (ID: {video_id})...")
    url = f"https://www.youtube.com/watch?v={video_id}"
    base_name = os.path.splitext(filename)[0]
    out_template = os.path.join(AMBIENCE_DIR, f"{base_name}.%(ext)s")

    cmd = [
        "yt-dlp",
        "-x",
        "--audio-format", "mp3",
        "--audio-quality", "128K",
        "-N", "4",
        "--no-playlist",
        "-o", out_template,
        url
    ]

    res = subprocess.run(cmd, capture_output=True, text=True)
    if res.returncode == 0 and os.path.exists(out_path):
        print(f"[CONCLUÍDO] {filename} ({os.path.getsize(out_path) // (1024*1024)} MB)")
        return True
    else:
        print(f"[ERRO] Falha ao baixar {filename}: {res.stderr[:200]}")
        return False

def main():
    os.makedirs(AMBIENCE_DIR, exist_ok=True)
    print(f"Iniciando download de {len(TRACKS)} faixas da playlist Curse of Strahd para pasta ambiente...")

    with concurrent.futures.ThreadPoolExecutor(max_workers=3) as executor:
        futures = {executor.submit(download_track, vid, fn): fn for vid, fn in TRACKS}
        for future in concurrent.futures.as_completed(futures):
            fn = futures[future]
            try:
                success = future.result()
                if not success:
                    print(f"⚠️ Atenção para {fn}")
            except Exception as e:
                print(f"❌ Exceção em {fn}: {e}")

    print("Processo de download finalizado!")

if __name__ == "__main__":
    main()
