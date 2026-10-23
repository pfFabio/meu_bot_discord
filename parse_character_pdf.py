#!/usr/bin/env python3
"""
parse_character_pdf.py
Extrator inteligente e adaptativo de fichas de RPG D&D (PDF para JSON).
Suporta fichas oficiais de D&D 5e / D&D 2024 / Roll20 / Beyond / Foundry.
"""

import os
import sys
import json
import re
import unicodedata

# Configurar stdout UTF-8
if hasattr(sys.stdout, "reconfigure"):
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")
if hasattr(sys.stderr, "reconfigure"):
    sys.stderr.reconfigure(encoding="utf-8", errors="replace")

def norm(text: str) -> str:
    """Normaliza texto removendo acentos e caracteres de substituição para comparações seguras"""
    if not text:
        return ""
    text = text.replace("\ufffd", "")
    nfkd = unicodedata.normalize("NFKD", text)
    return "".join(c for c in nfkd if not unicodedata.combining(c)).upper().strip()

def parse_dnd_pdf(pdf_path: str) -> dict:
    import pymupdf

    doc = pymupdf.open(pdf_path)
    total_pages = len(doc)
    pages_lines = []
    for i in range(total_pages):
        raw = doc[i].get_text()
        raw = raw.replace("\ufffd", "")
        lines = [l.strip() for l in raw.splitlines() if l.strip()]
        pages_lines.append(lines)

    sheet = {
        "name": "",
        "player_name": "",
        "class": "",
        "subclass": "",
        "level": 1,
        "race": "",
        "background": "",
        "alignment": "",
        "xp": 0,
        "inspiration": False,
        "proficiency_bonus": 2,
        "attributes": {
            "str": {"score": 10, "modifier": 0},
            "dex": {"score": 10, "modifier": 0},
            "con": {"score": 10, "modifier": 0},
            "int": {"score": 10, "modifier": 0},
            "wis": {"score": 10, "modifier": 0},
            "cha": {"score": 10, "modifier": 0},
        },
        "armor_class": 10,
        "initiative": 0,
        "speed": "9m / 30ft",
        "max_hp": 10,
        "current_hp": 10,
        "temp_hp": 0,
        "hit_dice": "1d8",
        "hit_dice_current": 1,
        "death_saves": {"successes": 0, "failures": 0},
        "saving_throws": {},
        "skills": {},
        "passive_perception": 10,
        "passive_investigation": 10,
        "passive_insight": 10,
        "proficiencies": {
            "armor": [],
            "weapons": [],
            "tools": [],
            "languages": []
        },
        "features": [],
        "limited_uses": [],
        "spellcasting": {
            "ability": "",
            "save_dc": 10,
            "attack_bonus": 0,
            "slots": {},
            "spells": []
        },
        "attacks": [],
        "coins": {
            "cp": 0, "sp": 0, "ep": 0, "gp": 0, "pp": 0
        },
        "inventory": []
    }

    if total_pages == 0:
        return sheet

    p0 = pages_lines[0]

    # 1. Identidade (Nome, Classe, Nível, Antecedente, Raça, Alinhamento)
    labels = ["CLASSE & NIVEL", "CLASSE & NVEL", "ANTECEDENTE", "NOME DO JOGADOR", "RACA", "RAA", "ALINHAMENTO", "PONTOS DE EXPERIENCIA", "PONTOS DE EXPERINCIA"]
    for idx, line in enumerate(p0):
        nline = norm(line)
        if "NOME DO PERSONAGEM" in nline and idx > 0:
            val = p0[idx - 1].strip()
            if not any(lbl in norm(val) for lbl in labels):
                sheet["name"] = val
        elif any(k in nline for k in ["CLASSE & NIVEL", "CLASSE & NVEL"]) and idx > 0:
            raw_class = p0[idx - 1]
            lvl_match = re.search(r"(\d+)$", raw_class)
            if lvl_match:
                sheet["level"] = int(lvl_match.group(1))
                class_part = raw_class[:lvl_match.start()].strip()
            else:
                class_part = raw_class

            parts = class_part.split()
            if len(parts) >= 2:
                sheet["class"] = parts[-1].title()
                raw_sub = " ".join(parts[:-1]).title()
                if "miseric" in raw_sub.lower():
                    raw_sub = "Misericórdia"
                sheet["subclass"] = raw_sub
            else:
                sheet["class"] = class_part.title()
        elif "ANTECEDENTE" in nline and idx > 0:
            val = p0[idx - 1].strip()
            if not any(lbl in norm(val) for lbl in labels) and norm(val) != "NOME DO PERSONAGEM":
                sheet["background"] = val
        elif any(k == nline for k in ["RACA", "RAA"]) and idx > 0:
            val = p0[idx - 1].strip()
            if not any(lbl in norm(val) for lbl in labels):
                sheet["race"] = val.title()
        elif "ALINHAMENTO" in nline and idx > 0:
            val = p0[idx - 1].strip()
            if not any(lbl in norm(val) for lbl in labels):
                sheet["alignment"] = val

    # Se nome não foi detectado pelo label, tenta a primeira linha não vazia
    if not sheet["name"] and len(p0) > 0:
        sheet["name"] = p0[0]

    # 2. Atributos Básicos
    attr_map = [
        ("FORCA", "str"),
        ("DESTREZA", "dex"),
        ("CONSTITUICAO", "con"),
        ("INTELIGENCIA", "int"),
        ("SABEDORIA", "wis"),
        ("CARISMA", "cha")
    ]

    for a_norm, k in attr_map:
        for idx, line in enumerate(p0):
            if norm(line) == a_norm:
                # Linhas seguintes: modificador (-5..+10) e valor (1..30)
                v1, v2 = None, None
                if idx + 1 < len(p0) and re.match(r"^-?\d+$", p0[idx + 1]):
                    v1 = int(p0[idx + 1])
                if idx + 2 < len(p0) and re.match(r"^-?\d+$", p0[idx + 2]):
                    v2 = int(p0[idx + 2])

                if v1 is not None and v2 is not None:
                    if 1 <= v2 <= 30 and -5 <= v1 <= 10:
                        sheet["attributes"][k]["score"] = v2
                        sheet["attributes"][k]["modifier"] = v1
                    elif 1 <= v1 <= 30 and -5 <= v2 <= 10:
                        sheet["attributes"][k]["score"] = v1
                        sheet["attributes"][k]["modifier"] = v2
                    else:
                        sheet["attributes"][k]["score"] = max(v1, v2)
                        sheet["attributes"][k]["modifier"] = min(v1, v2)
                break

    # 3. Bônus de Proficiência e Percepção Passiva
    for idx, line in enumerate(p0):
        nline = norm(line)
        if "BONUS DE PROFICIENCIA" in nline and idx > 0:
            if re.match(r"^\+?\d+$", p0[idx - 1]):
                sheet["proficiency_bonus"] = int(p0[idx - 1].replace("+", ""))
        elif "SABEDORIA PASSIVA (PERCEPCAO)" in nline and idx > 0:
            if re.match(r"^\d+$", p0[idx - 1]):
                sheet["passive_perception"] = int(p0[idx - 1])

    # Encontrar blocos delimitados de Salvaguardas e Perícias
    idx_salv = -1
    idx_peric = -1
    for i, l in enumerate(p0):
        nl = norm(l)
        if "SALVAGUARDAS" in nl and idx_salv == -1:
            idx_salv = i
        elif ("PERICIAS" in nl or "PERCIAS" in nl) and idx_peric == -1:
            idx_peric = i

    # 4. Salvaguardas (Resistências)
    if idx_salv != -1:
        salv_names = [
            ("Forca", "str"), ("Destreza", "dex"), ("Constituicao", "con"),
            ("Inteligencia", "int"), ("Sabedoria", "wis"), ("Carisma", "cha")
        ]
        salv_slice = p0[max(0, idx_salv - 14):idx_salv]
        for back in range(len(salv_slice)):
            val_line = salv_slice[back]
            for an, ak in salv_names:
                if re.search(r"\b" + an[:4], norm(val_line), re.IGNORECASE):
                    if back > 0 and re.match(r"^-?\d+$", salv_slice[back - 1]):
                        mod = int(salv_slice[back - 1])
                        attr_mod = sheet["attributes"][ak]["modifier"]
                        is_prof = mod > attr_mod
                        sheet["saving_throws"][ak] = {
                            "modifier": mod,
                            "proficient": is_prof
                        }

    # 5. Perícias (Skills)
    skill_definitions = [
        ("Acrobacia", r"acrobacia", "dex"),
        ("Adestrar Animais", r"adestrar.*animais", "wis"),
        ("Arcanismo", r"arcanismo", "int"),
        ("Atletismo", r"atletismo", "str"),
        ("Enganação", r"engan", "cha"),
        ("História", r"hist.*ria", "int"),
        ("Intuição", r"intui", "wis"),
        ("Intimidação", r"intimid", "cha"),
        ("Investigação", r"investig", "int"),
        ("Medicina", r"medicina", "wis"),
        ("Natureza", r"natureza", "int"),
        ("Percepção", r"percep", "wis"),
        ("Atuação", r"atua", "cha"),
        ("Persuasão", r"persua", "cha"),
        ("Religião", r"religi", "int"),
        ("Prestidigitação", r"prestidigita", "dex"),
        ("Furtividade", r"furtividade", "dex"),
        ("Sobrevivência", r"sobreviv", "wis")
    ]

    skill_lines = p0[idx_salv + 1:idx_peric] if (idx_salv != -1 and idx_peric != -1 and idx_peric > idx_salv) else p0
    for s_display, s_pat, s_attr in skill_definitions:
        for idx_k, line in enumerate(skill_lines):
            clean = re.sub(r"[\ufffd]", "", line).lower()
            if re.search(s_pat, clean) and "passiva" not in clean:
                if idx_k > 0 and re.match(r"^-?\d+$", skill_lines[idx_k - 1]):
                    mod = int(skill_lines[idx_k - 1])
                    attr_mod = sheet["attributes"].get(s_attr, {}).get("modifier", 0)
                    prof_b = sheet["proficiency_bonus"]
                    is_prof = mod >= attr_mod + prof_b - 1
                    is_exp = mod >= attr_mod + (prof_b * 2) - 1
                    sheet["skills"][s_display] = {
                        "modifier": mod,
                        "proficient": is_prof,
                        "expertise": is_exp,
                        "attribute": s_attr.upper()
                    }
                    break

    # 6. CA, Iniciativa, Deslocamento e HP
    for idx, line in enumerate(p0):
        nline = norm(line)
        if "PONTOS DE VIDA MAXIMO" in nline or "PONTOS DE VIDA" in nline:
            # Encontrar número de HP (linha seguinte ou anterior)
            if idx + 1 < len(p0) and re.match(r"^\d+$", p0[idx + 1]):
                sheet["max_hp"] = int(p0[idx + 1])
                sheet["current_hp"] = int(p0[idx + 1])
            elif idx > 0 and re.match(r"^\d+$", p0[idx - 1]):
                sheet["max_hp"] = int(p0[idx - 1])
                sheet["current_hp"] = int(p0[idx - 1])

            # Os 3 valores antes de Pontos de Vida: CA, Iniciativa, Deslocamento
            candidates = []
            for back in range(max(0, idx - 5), idx):
                val = p0[back].strip()
                if re.match(r"^[+-]?\d+(\s*ft)?$", val, re.IGNORECASE):
                    candidates.append(val)

            if len(candidates) >= 3:
                m_ca = re.search(r"\d+", candidates[-3])
                if m_ca:
                    sheet["armor_class"] = int(m_ca.group(0))
                m_inic = re.search(r"[+-]?\d+", candidates[-2])
                if m_inic:
                    sheet["initiative"] = int(m_inic.group(0))
                m_spd = re.search(r"\d+", candidates[-1])
                if m_spd:
                    spd_num = int(m_spd.group(0))
                    sheet["speed"] = f"{spd_num} ft ({int(spd_num * 0.3)}m)"
            break

    # Hit Dice padrão por nível
    sheet["hit_dice"] = f"{sheet['level']}d8"
    sheet["hit_dice_current"] = sheet["level"]

    # 7. Proficiências e Idiomas (FERRAMENTA, IDIOMA, ARMADURA, ARMA)
    # Delimita texto entre PERÍCIAS e o bloco de CA/HP para não capturar a página toda
    start_prof = idx_peric + 1 if idx_peric != -1 else 0
    end_prof = len(p0)
    for idx_hp, l in enumerate(p0):
        if "PONTOS DE VIDA" in norm(l):
            end_prof = max(0, idx_hp - 3)
            break
    prof_slice = " ".join(p0[start_prof:end_prof])

    for tag, key in [("FERRAMENTA:", "tools"), ("IDIOMA:", "languages"), ("ARMADURA:", "armor"), ("ARMA:", "weapons")]:
        if tag in prof_slice:
            m = re.search(re.escape(tag) + r"\s*(.*?)(?=FERRAMENTA:|IDIOMA:|ARMADURA:|ARMA:|\Z)", prof_slice)
            if m:
                raw_val = m.group(1).strip()
                raw_val = re.sub(r"\s+\d+\s*$", "", raw_val)
                items = [t.strip() for t in raw_val.split(",") if t.strip() and len(t.strip()) < 80]
                sheet["proficiencies"][key] = items

    # 8. Ataques
    weapon_name_expansions = {
        "spear (one-ha": "Spear (One-Handed)",
        "spear (two-ha": "Spear (Two-Handed)",
        "monk unarmed": "Monk Unarmed Strike",
        "adaga": "Adaga",
        "dagger": "Dagger",
        "clava (shillelagh)": "Clava (Shillelagh)",
        "clava": "Clava",
        "shortsword": "Espada Curta",
        "espada curta": "Espada Curta",
        "longsword": "Espada Longa",
        "espada longa": "Espada Longa",
    }

    damage_type_expansions = {
        "bludgeo": "Bludgeoning",
        "pierc": "Piercing",
        "slash": "Slashing",
        "cort": "Cortante",
        "perf": "Perfurante",
        "concus": "Concussão",
    }

    for idx, line in enumerate(p0):
        if "ATAQUES & CONJURACAO" in norm(line) or "ATAQUES E CONJURACAO" in norm(line):
            j = idx + 1
            while j < len(p0) and "EQUIPAMENTO" not in norm(p0[j]):
                if j + 2 < len(p0) and re.match(r"^(\+|\-|\bDC\b)\d+", p0[j+1]):
                    raw_name = p0[j].strip()
                    atk_bonus = p0[j+1].strip()
                    raw_damage = p0[j+2].strip()

                    name_clean = raw_name.replace("…", "").replace("\ufffd", "").strip()
                    name_lower = name_clean.lower()
                    for prefix, expanded in weapon_name_expansions.items():
                        if name_lower.startswith(prefix):
                            name_clean = expanded
                            break

                    dmg_clean = raw_damage.replace("…", "").replace("\ufffd", "").strip()
                    dmg_type = ""
                    dmg_parts = dmg_clean.split(maxsplit=1)
                    if len(dmg_parts) == 2:
                        dmg_val, dmg_t = dmg_parts
                        for t_pre, t_exp in damage_type_expansions.items():
                            if t_pre in dmg_t.lower():
                                dmg_t = t_exp
                                break
                        dmg_clean = f"{dmg_val} {dmg_t}"
                        dmg_type = dmg_t

                    sheet["attacks"].append({
                        "name": name_clean,
                        "attack_bonus": atk_bonus,
                        "damage": dmg_clean,
                        "damage_type": dmg_type
                    })
                    j += 3
                else:
                    j += 1
            break

    # 9. Moedas e Recursos Limitados (Páginas 0 e 1)
    for p_num in range(min(2, total_pages)):
        plines = pages_lines[p_num]
        for idx, line in enumerate(plines):
            nline = norm(line)
            # Moedas
            if nline in ["GP", "PO"] and idx > 0 and re.match(r"^\d+$", plines[idx - 1]):
                sheet["coins"]["gp"] = int(plines[idx - 1])
            elif nline in ["SP", "PP"] and idx > 0 and re.match(r"^\d+$", plines[idx - 1]):
                sheet["coins"]["sp"] = int(plines[idx - 1])
            elif nline in ["CP", "PC"] and idx > 0 and re.match(r"^\d+$", plines[idx - 1]):
                sheet["coins"]["cp"] = int(plines[idx - 1])

            # Recursos Limitados com usos diários (ex: Total: 2, 2, CANALIZAR DIVINDADE)
            if line.startswith("Total:") or line == "Total":
                if idx + 3 < len(plines):
                    v1 = plines[idx + 1]
                    v2 = plines[idx + 2]
                    res_name = plines[idx + 3]
                    if re.match(r"^\d+$", v1) and re.match(r"^\d+$", v2) and not res_name.startswith("Total"):
                        max_u = int(v1)
                        cur_u = int(v2)
                        res_clean = res_name.title()
                        if not any(r["name"] == res_clean for r in sheet["limited_uses"]):
                            sheet["limited_uses"].append({
                                "name": res_clean,
                                "max_uses": max_u,
                                "current_uses": cur_u,
                                "reset_on": "Descanso Longo" if any(w in res_clean.lower() for w in ["sort", "clar", "insp"]) else "Descanso Curto"
                            })

            # Equipamentos
            if "EQUIPAMENTO" in nline:
                j = idx + 1
                while j < len(plines) and not any(norm(plines[j]).startswith(k) for k in ["TOTAL", "CARACTERISTICAS", "TRACOS"]):
                    if re.match(r"^\d+$", plines[j]) and j + 1 < len(plines):
                        qty = int(plines[j])
                        item_name = plines[j + 1]
                        if not item_name.startswith("Total") and not item_name.isupper() and item_name not in ["CP", "SP", "EP", "GP", "PP"]:
                            sheet["inventory"].append({
                                "name": item_name,
                                "quantity": qty,
                                "weight": ""
                            })
                            j += 2
                            continue
                    j += 1

    # 10. Magias e Slots (Geralmente Página 2 em diante)
    for p_num in range(total_pages):
        plines = pages_lines[p_num]
        p_full = " ".join(plines)
        if "SPELLCASTING" in p_full or "TRUQUES" in p_full or "SPELL ATTACK" in p_full:
            # CD e Ataque
            for idx, line in enumerate(plines):
                if "SPELL SAVEDC" in norm(line) and idx + 3 < len(plines):
                    if re.match(r"^\d+$", plines[idx + 2]):
                        sheet["spellcasting"]["save_dc"] = int(plines[idx + 2])
                if "SPELL ATTACK" in norm(line) and idx + 2 < len(plines):
                    if re.match(r"^\+?\d+$", plines[idx + 2]):
                        sheet["spellcasting"]["attack_bonus"] = int(plines[idx + 2].replace("+", ""))

            # Slots e magias
            current_lvl = None
            i = 0
            while i < len(plines):
                line = plines[i]
                nline = norm(line)
                if nline == "TRUQUES":
                    current_lvl = 0
                    i += 1
                    continue
                elif re.match(r"^[1-9]$", line):
                    lvl = int(line)
                    if i + 1 < len(plines) and re.match(r"^\d+$", plines[i + 1]):
                        s_max = int(plines[i + 1])
                        if s_max > 0:
                            sheet["spellcasting"]["slots"][str(lvl)] = {
                                "level": lvl,
                                "max": s_max,
                                "used": 0
                            }
                        current_lvl = lvl
                        i += 2
                        continue

                if current_lvl is not None:
                    if re.match(r"^[1-9]$", line) or nline in ["TRUQUES", "CARACTERISTICAS & TRACOS", "EQUIPAMENTO"]:
                        pass
                    elif len(line) > 2 and not line.isdigit() and nline not in ["SPELLCASTING", "CLASS", "ABILITY", "BONUS"]:
                        sheet["spellcasting"]["spells"].append({
                            "name": line,
                            "level": current_lvl,
                            "prepared": True
                        })
                i += 1

    # Atributo de Conjuração automático caso haja magias
    if sheet["spellcasting"]["spells"] or sheet["spellcasting"]["slots"]:
        cls_lower = sheet["class"].lower()
        if any(c in cls_lower for c in ["cleric", "clerigo", "druid", "druida"]):
            sheet["spellcasting"]["ability"] = "Sabedoria"
        elif any(c in cls_lower for c in ["wizard", "mago", "artificer", "artifice"]):
            sheet["spellcasting"]["ability"] = "Inteligência"
        elif any(c in cls_lower for c in ["bard", "bardo", "sorcerer", "feiticeiro", "warlock", "bruxo", "paladin", "paladino"]):
            sheet["spellcasting"]["ability"] = "Carisma"

    # 11. Características & Traços de Classe e Espécie
    for p_num in range(total_pages):
        plines = pages_lines[p_num]
        for idx, line in enumerate(plines):
            if "CARACTERISTICAS & TRACOS" in norm(line):
                j = idx + 1
                while j < len(plines):
                    title = plines[j]
                    if j + 1 < len(plines) and len(plines[j + 1]) > len(title) and not plines[j+1].startswith("Total"):
                        desc = plines[j + 1]
                        sheet["features"].append({
                            "name": title,
                            "description": desc,
                            "source": "Classe/Espécie"
                        })
                        j += 2
                    else:
                        if len(title) > 2 and not title.isdigit() and not title.startswith("Total"):
                            sheet["features"].append({
                                "name": title,
                                "description": "",
                                "source": "Geral"
                            })
                        j += 1
                break

    # Deduplicar magias, features e equipamentos
    seen_spells = set()
    u_spells = []
    for sp in sheet["spellcasting"]["spells"]:
        if sp["name"].lower() not in seen_spells:
            seen_spells.add(sp["name"].lower())
            u_spells.append(sp)
    sheet["spellcasting"]["spells"] = u_spells

    seen_features = set()
    u_features = []
    for feat in sheet["features"]:
        if feat["name"].lower() not in seen_features:
            seen_features.add(feat["name"].lower())
            u_features.append(feat)
    sheet["features"] = u_features

    return sheet

def main():
    if len(sys.argv) < 2:
        print("Uso: python parse_character_pdf.py <caminho_do_pdf>")
        sys.exit(1)

    pdf_file = sys.argv[1]
    if not os.path.exists(pdf_file):
        print(f"Erro: Arquivo PDF '{pdf_file}' não encontrado.", file=sys.stderr)
        sys.exit(1)

    try:
        sheet_data = parse_dnd_pdf(pdf_file)
        pretty = "--pretty" in sys.argv
        if pretty:
            print(json.dumps(sheet_data, ensure_ascii=False, indent=2))
        else:
            print(json.dumps(sheet_data, ensure_ascii=False))
    except Exception as e:
        print(f"Erro ao processar PDF: {e}", file=sys.stderr)
        import traceback
        traceback.print_exc(file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()
