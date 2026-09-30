import os
import re
import json
import time
import hashlib
import urllib.request
import urllib.parse
import sys

CACHE_FILE = "translation_cache.json"
INPUT_FILE = "BetteR20 Backup - 2026-09-16 03_02_33.json"
OUTPUT_FILE = "BetteR20_Curse_of_Strahd_PTBR.json"
PROGRESS_FILE = "translation_progress.json"

# Carregar cache existente
cache = {}
if os.path.exists(CACHE_FILE):
    try:
        with open(CACHE_FILE, "r", encoding="utf-8") as f:
            cache = json.load(f)
        print(f"[CACHE] Carregado {len(cache)} traduções pré-existentes.")
    except Exception as e:
        print(f"[CACHE] Aviso: {e}")

def save_cache():
    try:
        with open(CACHE_FILE, "w", encoding="utf-8") as f:
            json.dump(cache, f, ensure_ascii=False, indent=1)
    except Exception as e:
        print(f"[CACHE] Erro ao salvar: {e}")

def js_escape(s):
    if not s:
        return ""
    res = []
    for c in s:
        code = ord(c)
        if code < 0x10000:
            if (48 <= code <= 57) or (65 <= code <= 90) or (97 <= code <= 122) or c in '@*_+-./':
                res.append(c)
            elif code < 256:
                res.append(f'%{code:02X}')
            else:
                res.append(f'%u{code:04X}')
        else:
            code -= 0x10000
            hi = 0xD800 + (code >> 10)
            lo = 0xDC00 + (code & 0x3FF)
            res.append(f'%u{hi:04X}%u{lo:04X}')
    return ''.join(res)

def js_unescape(s):
    if not s:
        return ""
    def repl(m):
        if m.group(1):
            return chr(int(m.group(1), 16))
        elif m.group(2):
            return chr(int(m.group(2), 16))
        return m.group(0)
    text = re.sub(r'%u([0-9a-fA-F]{4})|%([0-9a-fA-F]{2})', repl, s)
    try:
        return text.encode('utf-16', 'surrogatepass').decode('utf-16')
    except Exception:
        return text

def request_translation(text, source="en", target="pt"):
    if not text or not text.strip():
        return text
    
    key = hashlib.md5(text.encode('utf-8')).hexdigest()
    if key in cache:
        return cache[key]
    
    url = "https://translate.googleapis.com/translate_a/single?client=gtx&sl=en&tl=pt&dt=t"
    data = urllib.parse.urlencode({'q': text}).encode('utf-8')
    req = urllib.request.Request(
        url,
        data=data,
        headers={
            'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36'
        }
    )
    
    for attempt in range(4):
        try:
            with urllib.request.urlopen(req, timeout=15) as res:
                raw_json = json.loads(res.read().decode('utf-8'))
                translated = "".join([segment[0] for segment in raw_json[0] if segment and segment[0]])
                cache[key] = translated
                return translated
        except Exception as e:
            wait_time = (attempt + 1) * 2
            time.sleep(wait_time)
            
    return text

def split_html_chunks(html, max_len=3000):
    if len(html) <= max_len:
        return [html]
    
    pattern = r'(</p>|</div>|</h2>|</h3>|</h4>|</tr>|</table>|<br\s*/?>)'
    parts = re.split(pattern, html, flags=re.IGNORECASE)
    
    chunks = []
    current = ""
    i = 0
    while i < len(parts):
        token = parts[i]
        tag = parts[i+1] if i + 1 < len(parts) else ""
        piece = token + tag
        i += 2
        
        if len(current) + len(piece) > max_len and current:
            chunks.append(current)
            current = piece
        else:
            current += piece
            
    if current:
        chunks.append(current)
        
    return chunks

def translate_html_content(raw_encoded_blob):
    if not raw_encoded_blob or len(raw_encoded_blob) < 10:
        return raw_encoded_blob
    
    decoded = js_unescape(raw_encoded_blob)
    if not decoded.strip():
        return raw_encoded_blob
    
    chunks = split_html_chunks(decoded, max_len=3200)
    translated_chunks = []
    
    for chunk in chunks:
        if not chunk.strip():
            translated_chunks.append(chunk)
            continue
        tr = request_translation(chunk)
        translated_chunks.append(tr)
        time.sleep(0.12)
        
    full_translated = "".join(translated_chunks)
    return js_escape(full_translated)

FOLDER_TRANSLATIONS = {
    "Curse of Strahd": "A Maldição de Strahd",
    "START HERE": "COMECE AQUI",
    "Roll20 Information": "Informações do Roll20",
    "Introduction": "Introdução",
    "Chapter 1 - Into the Mists": "Capítulo 1 - Nas Brumas",
    "Fortunes of Ravenloft": "Destinos de Ravenloft (Tarokka)",
    "Adventure Hooks": "Ganchos de Aventura",
    "Chapter 2 - The Lands of Barovia": "Capítulo 2 - As Terras de Baróvia",
    "Chapter 3 - The Village of Barovia": "Capítulo 3 - O Vilarejo de Baróvia",
    "Chapter 4 - Castle Ravenloft": "Capítulo 4 - O Castelo Ravenloft",
    "Chapter 5 - The Town of Vallaki": "Capítulo 5 - A Cidade de Vallaki",
    "Areas of Vallaki": "Áreas de Vallaki",
    "Special Events": "Eventos Especiais",
    "Chapter 6 - Old Bonegrinder": "Capítulo 6 - O Moinho Velho",
    "Chapter 7 - Argynvostholt": "Capítulo 7 - Argynvostholt",
    "Chapter 8 - The Village of Krezk": "Capítulo 8 - A Vila de Krezk",
    "Chapter 9 - Tsolenka Pass": "Capítulo 9 - O Passo de Tsolenka",
    "Chapter 10 - The Ruins of Berez": "Capítulo 10 - As Ruínas de Berez",
    "Chapter 11 - Van Richten's Tower": "Capítulo 11 - A Torre de Van Richten",
    "Chapter 12 - The Wizard of Wines": "Capítulo 12 - Os Magos do Vinho",
    "Chapter 13 - The Amber Temple": "Capítulo 13 - O Templo de Âmbar",
    "Chapter 14 - Yester Hill": "Capítulo 14 - A Colina de Yester",
    "Chapter 15 - Werewolf Den": "Capítulo 15 - O Covil dos Lobisomens",
    "Epilogue": "Epílogo",
    "Appendix A - Character Options": "Apêndice A - Opções de Personagem",
    "Appendix B - Death House": "Apêndice B - A Casa da Morte",
    "Appendix C - Treasures": "Apêndice C - Tesouros",
    "Appendix D - Monsters and NPCs": "Apêndice D - Monstros e NPCs",
    "Appendix E - The Tarokka Deck": "Apêndice E - O Baralho de Tarokka",
    "Appendix F - Handouts": "Apêndice F - Documentos e Cartas",
    "Player Handouts": "Documentos dos Jogadores (Handouts)",
    "Treasure": "Tesouros e Itens Mágicos",
    "Supplemental Material": "Material Suplementar"
}

def translate_path(path_list):
    if not path_list:
        return path_list
    res = []
    for item in path_list:
        res.append(FOLDER_TRANSLATIONS.get(item, item))
    return res

def main():
    print(f"[INÍCIO] Carregando arquivo de backup: {INPUT_FILE}")
    with open(INPUT_FILE, "r", encoding="utf-8") as f:
        data = json.load(f)
        
    total_handouts = len(data.get("handouts", []))
    print(f"[CAMPANHA] Total de Handouts para traduzir: {total_handouts}")
    
    # 1. Traduzir Folders da Journal
    print("[1/3] Traduzindo estrutura de pastas da campanha...")
    for j in data.get("journal", []):
        if "path" in j and j["path"]:
            j["path"] = translate_path(j["path"])
            
    # 2. Traduzir Handouts
    print("[2/3] Iniciando tradução contínua dos Handouts...")
    start_time = time.time()
    
    for idx, h in enumerate(data.get("handouts", [])):
        orig_name = h["attributes"].get("name", "")
        trans_name = request_translation(orig_name)
        h["attributes"]["name"] = trans_name
        
        # Traduzir Notes
        if "blobNotes" in h and h["blobNotes"]:
            h["blobNotes"] = translate_html_content(h["blobNotes"])
            
        # Traduzir GM Notes
        if "blobGmNotes" in h and h["blobGmNotes"]:
            h["blobGmNotes"] = translate_html_content(h["blobGmNotes"])
            
        pct = ((idx + 1) / total_handouts) * 100
        elapsed = time.time() - start_time
        avg_per_item = elapsed / (idx + 1)
        remaining_s = avg_per_item * (total_handouts - (idx + 1))
        
        print(f"[{idx+1}/{total_handouts}] ({pct:.1f}%) Traduzido: {orig_name} -> {trans_name} | Restante: ~{remaining_s/60:.1f} min")
        
        # Salvar cache e progresso a cada 10 handouts
        if (idx + 1) % 10 == 0 or idx + 1 == total_handouts:
            save_cache()
            # Salvar checkpoint no output intermediário
            with open(OUTPUT_FILE, "w", encoding="utf-8") as out_f:
                json.dump(data, out_f, ensure_ascii=False)
            print(f"  --> Checkpoint salvo com sucesso no arquivo {OUTPUT_FILE}!")

    # 3. Traduzir Bios dos Personagens/Monstros principais
    print("[3/3] Traduzindo bios e descrições dos Personagens...")
    for c in data.get("characters", []):
        if "blobBio" in c and c["blobBio"]:
            c["blobBio"] = translate_html_content(c["blobBio"])
        if "blobGmNotes" in c and c["blobGmNotes"]:
            c["blobGmNotes"] = translate_html_content(c["blobGmNotes"])
            
    # Salvar arquivo final consolidado
    print(f"[CONCLUÍDO] Salvando arquivo final traduzido em: {OUTPUT_FILE}")
    with open(OUTPUT_FILE, "w", encoding="utf-8") as out_f:
        json.dump(data, out_f, ensure_ascii=False)
        
    save_cache()
    print("🎉 TODA A CAMPANHA FOI TRADUZIDA COM SUCESSO PARA O PORTUGUÊS!")

if __name__ == "__main__":
    main()
