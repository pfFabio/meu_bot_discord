import os
import sys
import json
import re
import subprocess
import concurrent.futures

# Suporte UTF-8 no terminal Windows
if hasattr(sys.stdout, "reconfigure"):
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")

TEMAS_DIR = os.path.abspath("temas")
PLAYLIST_INFO_FILE = os.path.abspath("playlist_info.json")

def sanitize_filename(name):
    # Remove sufixos como "| Unofficial Curse of Strahd..."
    clean = name.split("|")[0].strip()
    # Remove termos redundantes entre parênteses
    clean = re.sub(r"\s*\((1\s*h|loop|ep|no ai|bonus track)\)", "", clean, flags=re.IGNORECASE)
    # Remove caracteres inválidos no Windows
    clean = re.sub(r'[<>:"/\\|?*]', "", clean)
    clean = re.sub(r"\s+", " ", clean).strip()
    return f"CoS - {clean}.mp3"

def categorize_track(title):
    lt = title.lower()
    if any(k in lt for k in ["strahd battle", "encounter", "fight", "onslaught", "skirmish", "wrath", "vermin", "dread", "spawn", "hags", "revenant", "battle"]):
        return "Combates & Encontros", "⚔️"
    if any(k in lt for k in ["strahd", "ireena", "ismark", "rahadin", "tatyana", "brides", "izek", "rictavio", "godfrey", "horngaard", "abbot", "clovin", "vasilka", "ezmerelda", "baba lysaga", "van richten", "scholar", "father", "hunter", "exethanter", "donavich", "doru", "rose and thorn", "hanged one"]):
        return "Personagens & NPCs", "👑"
    if any(k in lt for k in ["sunsword", "holy symbol", "tome of strahd", "icon of ravenloft", "beacon", "tarokka", "dark gifts"]):
        return "Relíquias & Destino", "✨"
    if any(k in lt for k in ["keepers of the feather", "silver dragon", "martikov", "vistani", "werewolf"]):
        return "Facções & Ordens", "🦅"
    return "Cenários & Mistério", "🕯️"

def load_playlist_tracks():
    if not os.path.exists(PLAYLIST_INFO_FILE):
        print(f"Arquivo {PLAYLIST_INFO_FILE} não encontrado. Baixando metadados da playlist...")
        cmd = [
            "yt-dlp",
            "--flat-playlist",
            "-J",
            "https://www.youtube.com/playlist?list=PLQBqJdSe3H4E7e-Vibw_RZ2bKTH03Z-yX"
        ]
        res = subprocess.run(cmd, capture_output=True, text=True, encoding="utf-8")
        if res.returncode != 0:
            raise RuntimeError(f"Erro ao obter metadados da playlist: {res.stderr}")
        data = json.loads(res.stdout)
    else:
        try:
            with open(PLAYLIST_INFO_FILE, "r", encoding="utf-16") as f:
                data = json.load(f)
        except Exception:
            with open(PLAYLIST_INFO_FILE, "r", encoding="utf-8-sig") as f:
                data = json.load(f)

    entries = data.get("entries", [])
    tracks = []
    seen_filenames = set()

    for idx, e in enumerate(entries):
        # Pula compilações completas (álbum completo no início e compilações no final)
        if idx in [0, 95, 96, 97, 98]:
            continue

        vid = e.get("id")
        raw_title = e.get("title", "")
        clean_title = raw_title.split("|")[0].strip()
        filename = sanitize_filename(raw_title)

        # Trata possíveis colisões
        if filename in seen_filenames:
            base, ext = os.path.splitext(filename)
            filename = f"{base} (v2){ext}"
        seen_filenames.add(filename)

        category, emoji = categorize_track(clean_title)
        dur = int(e.get("duration") or 0)
        dur_str = f"{dur//60:02d}:{dur%60:02d}" if dur > 0 else "1h"

        key = re.sub(r"[^a-zA-Z0-9_]+", "_", clean_title.lower()).strip("_")

        tracks.append({
            "index": idx + 1,
            "id": vid,
            "key": key,
            "title": clean_title,
            "filename": filename,
            "category": category,
            "emoji": emoji,
            "duration": dur_str,
            "youtube_url": f"https://www.youtube.com/watch?v={vid}"
        })

    return tracks

def download_track(track, total, counter_dict, lock):
    vid = track["id"]
    filename = track["filename"]
    out_path = os.path.join(TEMAS_DIR, filename)

    # Verifica se já foi baixado com sucesso
    if os.path.exists(out_path) and os.path.getsize(out_path) > 1 * 1024 * 1024:
        with lock:
            counter_dict["done"] += 1
            print(f"[{counter_dict['done']}/{total}] [JÁ EXISTE] {filename} ({os.path.getsize(out_path) // (1024*1024)} MB)")
        return True

    with lock:
        print(f"[BAIXANDO] {track['emoji']} {track['title']} -> {filename}...")

    # Template de saída
    out_template = os.path.join(TEMAS_DIR, f"{os.path.splitext(filename)[0]}.%(ext)s")

    cmd = [
        "yt-dlp",
        "-f", "ba[ext=m4a]/ba",
        "-x",
        "--audio-format", "mp3",
        "--audio-quality", "128K",
        "-N", "4",
        "--no-playlist",
        "-o", out_template,
        track["youtube_url"]
    ]

    res = subprocess.run(cmd, capture_output=True, text=True)
    if res.returncode == 0 and os.path.exists(out_path):
        size_mb = os.path.getsize(out_path) // (1024 * 1024)
        with lock:
            counter_dict["done"] += 1
            print(f"[{counter_dict['done']}/{total}] [CONCLUÍDO] {filename} ({size_mb} MB)")
        return True
    else:
        err_msg = res.stderr[-200:] if res.stderr else "Erro desconhecido"
        with lock:
            print(f"[ERRO] Falha em {filename}: {err_msg}")
        return False

def main():
    os.makedirs(TEMAS_DIR, exist_ok=True)
    tracks = load_playlist_tracks()
    print(f"Carregadas {len(tracks)} faixas individuais de Curse of Strahd.")

    # Salva catálogo JSON em temas/temas_catalog.json
    catalog_path = os.path.join(TEMAS_DIR, "temas_catalog.json")
    with open(catalog_path, "w", encoding="utf-8") as f:
        json.dump(tracks, f, ensure_ascii=False, indent=2)
    print(f"Catálogo salvo em: {catalog_path}")

    import threading
    lock = threading.Lock()
    counter_dict = {"done": 0}

    print(f"Iniciando download concorrente de {len(tracks)} faixas com 3 workers...")
    with concurrent.futures.ThreadPoolExecutor(max_workers=3) as executor:
        futures = {executor.submit(download_track, t, len(tracks), counter_dict, lock): t for t in tracks}
        for future in concurrent.futures.as_completed(futures):
            t = futures[future]
            try:
                future.result()
            except Exception as e:
                print(f"Exceção em {t['filename']}: {e}")

    print("Todos os downloads foram concluídos com sucesso!")

if __name__ == "__main__":
    main()
