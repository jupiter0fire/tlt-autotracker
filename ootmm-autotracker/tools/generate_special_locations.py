#!/usr/bin/env python3

import argparse
import csv
import json
import pathlib
import re
import sys
from collections import defaultdict
from typing import Any


SUPPORTED_EXTRA_GROUPS = {"gMmExtraFlags", "gMmExtraFlags2", "gMmExtraFlags3"}
SUPPORTED_GROUPS = SUPPORTED_EXTRA_GROUPS | {"weekEventReg", "gMmOwlFlags", "sharedNpcBitmap", "inventoryQuest", "gMmExtraBoss"}
SUPPORTED_OOT_GROUPS = {"gOotExtraFlags", "inventoryQuest", "gOotTradeSave", "eventsChk", "eventsItem", "eventsMisc"}

EXTRA_STRUCTS = {
    "gMmExtraFlags": "MmExtraFlags",
    "gMmExtraFlags2": "MmExtraFlags2",
    "gMmExtraFlags3": "MmExtraFlags3",
}

QUEST_SYMBOL_FIELDS = {
    "MM_SONG_AWAKENING": "songAwakening",
    "MM_SONG_ZORA": "songNewWave",
    "MM_SKULL_KID_SONG": "songTime",
}

BOSS_SYMBOL_BITS = {
    "MM_REMAINS_ODOLWA": 0,
    "MM_REMAINS_GOHT": 1,
    "MM_REMAINS_GYORG": 2,
    "MM_REMAINS_TWINMOLD": 3,
}

OOT_EXTRA_SOURCES = [
    ("FAIRY_MAGIC_UPGRADE", "greatFairies", 25, "Great Fairy magic upgrade"),
    ("FAIRY_MAGIC_UPGRADE2", "greatFairies", 26, "Great Fairy magic upgrade 2"),
    ("FAIRY_DEFENSE_UPGRADE", "greatFairies", 27, "Great Fairy defense upgrade"),
    ("FAIRY_SPELL_WIND", "greatFairies", 28, "Great Fairy wind spell"),
    ("FAIRY_SPELL_FIRE", "greatFairies", 29, "Great Fairy fire spell"),
    ("FAIRY_SPELL_LOVE", "greatFairies", 30, "Great Fairy love spell"),
    ("FISH_CHILD", "fishingChild", 24, "Child fishing pond completed"),
    ("FISH_ADULT", "fishingAdult", 23, "Adult fishing pond completed"),
    ("GORON_LINK_TUNIC", "tunicGoron", 22, "Goron tunic trade item"),
    ("TRADE_BIGGORON_SWORD", "biggoron", 21, "Biggoron Sword trade item"),
    ("ZORA_KING_TUNIC", "tunicZora", 20, "Zora tunic trade item"),
    ("FIRE_ARROW", "fireArrow", 19, "Fire Arrow reward"),
    ("CHEST_GAME_KEY", "chestGameKey", 6, "Treasure Chest Game skeleton key"),
    ("TRADE_ODD_POTION", "oddPotion", 2, "Odd Potion trade item"),
]

OOT_EVENT_CHK_SOURCES = [
    ("MALON_EGG", 0x12, "Malon egg event"),
    ("MALON_SONG", 0x62, "Epona's Song event"),
    ("MASTER_SWORD", 0x4F, "Master Sword pulled event"),
    ("LIGHT_MEDALLION", 0x45, "Light Medallion obtained"),
    ("OCARINA_TIME_ITEM", 0x43, "Ocarina of Time item obtained"),
    ("OCARINA_TIME_SONG", 0xA9, "Song of Time learned"),
    ("ROYAL_TOMB_SONG", 0x5A, "Song of Storms (Royal Tomb) event"),
    ("RUTO_LETTER", 0x31, "Ruto's Letter event"),
    ("FROGS_GAME", 0xD0, "Frogs game started"),
    ("FROGS_ZL", 0xD1, "Frogs Zelda's Song"),
    ("FROGS_EPONA", 0xD2, "Frogs Epona's Song"),
    ("FROGS_SARIA", 0xD4, "Frogs Saria's Song"),
    ("FROGS_SUNS", 0xD3, "Frogs Sun's Song"),
    ("FROGS_SOT", 0xD5, "Frogs Song of Time"),
    ("FROGS_STORMS", 0xD6, "Frogs Song of Storms"),
    ("SARIA_SONG", 0x58, "Saria's Song event"),
    ("ZORA_DIVING_GAME", 0x38, "Zora Diving Game completed"),
    ("GS_10", 0xDA, "10 Golden Skulltulas collected"),
    ("GS_20", 0xDB, "20 Golden Skulltulas collected"),
    ("GS_30", 0xDC, "30 Golden Skulltulas collected"),
    ("GS_40", 0xDD, "40 Golden Skulltulas collected"),
    ("GS_50", 0xDE, "50 Golden Skulltulas collected"),
    ("SARIA_OCARINA", 0xC1, "Saria's Ocarina obtained"),
    ("SHEIK_FOREST", 0x50, "Forest Temple boss defeated"),
    ("SHEIK_FIRE", 0x51, "Fire Temple boss defeated"),
    ("SHEIK_WATER", 0x52, "Water Temple boss defeated"),
    ("SHEIK_SHADOW", 0x54, "Shadow Temple boss defeated"),
    ("SHEIK_LIGHT", 0x55, "Light Temple boss defeated"),
    ("SHEIK_SPIRIT", 0xAC, "Spirit Temple boss defeated"),
    ("SONG_STORMS", 0x5B, "Song of Storms learned"),
    ("ZELDA_LETTER_EVENT", 0x40, "Zelda's Letter event flag"),
    ("ZELDA_LIGHT_ARROW", 0xC4, "Light Arrows obtained"),
    ("ZELDA_SONG", 0x59, "Zelda's Song learned"),
]

OOT_EVENT_ITEM_SOURCES = [
    ("ANJU_BOTTLE", 0x0C, "Anju's Bottle obtained"),
    ("TALON_BOTTLE", 0x02, "Talon Bottle obtained"),
    ("SHOOTING_GAME_ADULT", 0x0E, "Shooting Gallery (Adult) reward"),
    ("GERUDO_ARCHERY_2", 0x0F, "Gerudo Archery second reward"),
    ("LABORATORY_DIVE", 0x10, "Laboratory Dive reward"),
    ("BOMBCHU_BOWLING_1", 0x11, "Bombchu Bowling first reward"),
    ("BOMBCHU_BOWLING_2", 0x12, "Bombchu Bowling second reward"),
    ("KAKARIKO_ROOF_MAN", 0x15, "Kakariko Roof Man reward"),
    ("LOST_WOODS_TARGET", 0x1D, "Lost Woods Target reward"),
    ("DARUNIA_BRACELET", 0x20, "Darunia's Bracelet obtained"),
    ("TRADE_POCKET_EGG", 0x2C, "Trade sequence Pocket Egg event"),
    ("POCKET_EGG_EVENT", 0x2C, "Pocket Egg event item"),
    ("MASK_SELL_KEATON", 0x38, "Keaton Mask sold"),
    ("MASK_SELL_SKULL", 0x39, "Skull Mask sold"),
    ("MASK_SELL_SPOOKY", 0x3A, "Spooky Mask sold"),
    ("MASK_SELL_BUNNY", 0x3B, "Bunny Hood sold"),
]

OOT_EVENT_MISC_SOURCES = [
    ("GORON_BOMB_BAG", 0x11E, "Goron Bomb Bag obtained"),
    ("GERUDO_ARCHERY_1", 0x190, "Gerudo Archery first reward"),
    ("DOG_LADY", 0x191, "Dog Lady Heart Piece"),
    ("MEDIGORON", 0xB2, "Medigoron obtained"),
]

NPC_DEFINE_RE = re.compile(r"^(MM_[A-Z0-9_]+):\s*(0x[0-9a-fA-F]+|\d+)\s*$")
BITFIELD_RE = re.compile(r"\bu32\s+([A-Za-z_][A-Za-z0-9_]*)\s*:\s*(\d+)\s*;")
STRUCT_RE_TEMPLATE = r"typedef\s+struct\s*\{(?P<body>[^{}]*)\}\s*%s\s*;"
MM_EV_RE = re.compile(r"#define\s+(EV_MM_WEEK_[A-Z0-9_]+)\s+MM_EV\((\d+),\s*(\d+)\)")
MM_SET_EVENT_RE = re.compile(r"MM_SET_EVENT_WEEK\(([^)]+)\)")
EXTRA_ASSIGN_RE = re.compile(r"\b(gMmExtraFlags2|gMmExtraFlags3|gMmExtraFlags)\.([A-Za-z_][A-Za-z0-9_]*)\s*(?:=|\|=)")
NPC_REF_RE = re.compile(r"\bNPC_(MM_[A-Z0-9_]+)\b")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Generate special location metadata for the autotracker from an OoTMM checkout."
    )
    parser.add_argument("--ootmm-repo", required=True, help="Path to the OoTMM repository root.")
    parser.add_argument("--output", required=True, help="Path to write MM special_locations.json.")
    parser.add_argument("--oot-output", help="Optional path to write OoT oot_special_locations.json.")
    parser.add_argument(
        "--hints",
        help=(
            "Optional existing special_locations.json to preserve manual source hints/notes. "
            "Defaults to --output when that file exists."
        ),
    )
    parser.add_argument(
        "--oot-hints",
        help=(
            "Optional existing oot_special_locations.json to preserve manual source hints/notes. "
            "Defaults to --oot-output when that file exists."
        ),
    )
    parser.add_argument(
        "--no-existing-hints",
        action="store_true",
        help="Do not read output files as hint files when explicit hints are omitted.",
    )
    return parser.parse_args()


def load_npc_ids(path: pathlib.Path) -> dict[str, int]:
    ids: dict[str, int] = {}
    for raw_line in path.read_text(encoding="utf-8").splitlines():
        line = raw_line.split("#", 1)[0].strip()
        match = NPC_DEFINE_RE.match(line)
        if not match:
            continue
        symbol, value = match.groups()
        if symbol.startswith("MM_"):
            ids[symbol] = int(value, 0)
    return ids


def load_pool_names(path: pathlib.Path) -> dict[str, str]:
    names: dict[str, str] = {}
    with path.open(newline="", encoding="utf-8") as handle:
        reader = csv.DictReader(handle, skipinitialspace=True)
        for row in reader:
            if row["type"].strip() != "npc":
                continue
            location = row["location"].strip()
            raw_id = row["id"].strip()
            if not location or not raw_id:
                continue
            symbol = raw_id if raw_id.startswith("MM_") else f"MM_{raw_id}"
            names.setdefault(symbol, location)
    return names


def load_hints(path: pathlib.Path | None) -> dict[str, dict[str, Any]]:
    if path is None or not path.is_file():
        return {}
    entries = json.loads(path.read_text(encoding="utf-8"))
    return {entry["symbol"]: entry for entry in entries if entry.get("symbol")}


def parse_bitfield_structs(header: pathlib.Path) -> dict[str, dict[str, dict[str, Any]]]:
    text = header.read_text(encoding="utf-8")
    structs: dict[str, dict[str, dict[str, Any]]] = {}

    for group, struct_name in EXTRA_STRUCTS.items():
        match = re.search(STRUCT_RE_TEMPLATE % re.escape(struct_name), text, re.S)
        if not match:
            raise ValueError(f"failed to find {struct_name} in {header}")

        offset = 0
        fields: dict[str, dict[str, Any]] = {}
        for field, width_text in BITFIELD_RE.findall(match.group("body")):
            width = int(width_text)
            logical_bits = list(range(offset, offset + width))
            raw_bits = [31 - (offset + width - 1 - idx) for idx in range(width)]
            if not field.startswith("unused"):
                fields[field] = {
                    "width": width,
                    "offset": offset,
                    "logical_bits": logical_bits,
                    "raw_bits": raw_bits,
                    "logical_mask": sum(1 << bit for bit in logical_bits),
                }
            offset += width
        structs[group] = fields
    return structs


def parse_quest_fields(header: pathlib.Path) -> dict[str, dict[str, Any]]:
    text = header.read_text(encoding="utf-8")
    match = re.search(r"typedef\s+union\s*\{(?P<body>.*?)\}\s*MmQuestItems\s*;", text, re.S)
    if not match:
        raise ValueError(f"failed to find MmQuestItems in {header}")
    body = match.group("body")
    fields: dict[str, dict[str, Any]] = {}
    offset = 0
    for field, width_text in BITFIELD_RE.findall(body):
        width = int(width_text)
        logical_bits = list(range(offset, offset + width))
        if not field.startswith("unused"):
            fields[field] = {
                "width": width,
                "offset": offset,
                "logical_bits": logical_bits,
                "logical_mask": sum(1 << bit for bit in logical_bits),
            }
        offset += width
    return fields


def parse_week_events(path: pathlib.Path) -> dict[str, tuple[int, int]]:
    events: dict[str, tuple[int, int]] = {}
    for name, byte_text, bit_text in MM_EV_RE.findall(path.read_text(encoding="utf-8")):
        byte_index = int(byte_text)
        bit = int(bit_text)
        events[name] = (byte_index, 1 << bit)
    return events


def extra_source(group: str, field: str, bitfields: dict[str, dict[str, dict[str, Any]]]) -> dict[str, Any] | None:
    info = bitfields.get(group, {}).get(field)
    if info is None:
        return None
    return {
        "group": group,
        "field": f"{group}.{field}",
        "mask": f"0x{info['logical_mask']:08x}",
    }


def enrich_source(source: dict[str, Any], bitfields: dict[str, dict[str, dict[str, Any]]]) -> tuple[dict[str, Any], list[int], int | None, int | None]:
    source = dict(source)
    bits: list[int] = []
    byte_index: int | None = None
    mask: int | None = None

    group = source.get("group", "")
    field = source.get("field", "")
    source_mask = source.get("mask")

    if group in SUPPORTED_EXTRA_GROUPS:
        field_name = field.rsplit(".", 1)[-1]
        field_info = bitfields.get(group, {}).get(field_name)
        if field_info is not None:
            if not source_mask:
                source["mask"] = f"0x{field_info['logical_mask']:08x}"
                bits = list(field_info["raw_bits"])
            else:
                logical_mask = int(str(source_mask), 0)
                raw_bits: list[int] = []
                for logical_bit, raw_bit in zip(field_info["logical_bits"], field_info["raw_bits"]):
                    if logical_mask & (1 << logical_bit):
                        raw_bits.append(raw_bit)
                bits = raw_bits or list(field_info["raw_bits"])
    elif group == "weekEventReg":
        if source_mask:
            mask = int(str(source_mask), 0)
        match = re.search(r"weekEventReg\[(\d+)\]", field)
        if match:
            byte_index = int(match.group(1))
        if byte_index is not None and mask is not None:
            bits = [byte_index * 8 + bit for bit in range(8) if mask & (1 << bit)]
    elif group in {"gMmOwlFlags", "sharedNpcBitmap"}:
        if source_mask:
            logical_mask = int(str(source_mask), 0)
            bits = [bit for bit in range(32) if logical_mask & (1 << bit)]

    return source, bits, byte_index, mask


def discover_simple_sources(repo_root: pathlib.Path, bitfields: dict[str, dict[str, dict[str, Any]]], week_events: dict[str, tuple[int, int]]) -> dict[str, list[dict[str, Any]]]:
    src_root = repo_root / "packages/generator/src"
    sources: dict[str, list[dict[str, Any]]] = defaultdict(list)

    for path in list(src_root.rglob("*.c")) + list(src_root.rglob("*.S")):
        text = path.read_text(encoding="utf-8", errors="ignore")
        symbols = sorted(set(NPC_REF_RE.findall(text)))
        if len(symbols) != 1:
            continue

        symbol = symbols[0]
        for group, field in EXTRA_ASSIGN_RE.findall(text):
            source = extra_source(group, field, bitfields)
            if source is not None:
                append_unique_source(sources[symbol], source)

        for expr in MM_SET_EVENT_RE.findall(text):
            expr = expr.strip()
            if expr in week_events:
                byte_index, mask = week_events[expr]
                append_unique_source(
                    sources[symbol],
                    {
                        "group": "weekEventReg",
                        "field": f"gMmSave.info.weekEventReg[{byte_index}]",
                        "mask": f"0x{mask:02x}",
                    },
                )
            elif re.fullmatch(r"0x[0-9a-fA-F]+|\d+", expr):
                event = int(expr, 0)
                byte_index = event >> 3
                mask = 1 << (event & 7)
                append_unique_source(
                    sources[symbol],
                    {
                        "group": "weekEventReg",
                        "field": f"gMmSave.info.weekEventReg[{byte_index}]",
                        "mask": f"0x{mask:02x}",
                    },
                )

    return sources


def append_unique_source(sources: list[dict[str, Any]], source: dict[str, Any]) -> None:
    key = (
        source.get("group"),
        source.get("field"),
        source.get("mask"),
        source.get("bit"),
        source.get("flag"),
    )
    if all(
        (item.get("group"), item.get("field"), item.get("mask"), item.get("bit"), item.get("flag")) != key
        for item in sources
    ):
        sources.append(source)


def oot_bit_source(group: str, field: str, bit: int) -> dict[str, Any]:
    return {
        "group": group,
        "field": field,
        "bit": bit,
    }


def oot_flag_source(group: str, field: str, flag: int) -> dict[str, Any]:
    return {
        "group": group,
        "field": field,
        "flag": flag,
    }


def oot_mask_source(group: str, field: str, mask: int) -> dict[str, Any]:
    return {
        "group": group,
        "field": field,
        "mask": f"0x{mask:04X}",
    }


def apply_oot_hint(entry: dict[str, Any], hint: dict[str, Any]) -> dict[str, Any]:
    if not hint:
        return entry

    hinted_sources: list[dict[str, Any]] = []
    for source in hint.get("sources", []):
        if source.get("group") in SUPPORTED_OOT_GROUPS:
            append_unique_source(hinted_sources, source)
    if hinted_sources:
        entry["sources"] = hinted_sources

    for key in ("name", "note"):
        if hint.get(key):
            entry[key] = hint[key]

    return entry


def build_oot_entries(hints: dict[str, dict[str, Any]]) -> list[dict[str, Any]]:
    entries: list[dict[str, Any]] = []

    def add(symbol: str, sources: list[dict[str, Any]], note: str) -> None:
        entry = {
            "symbol": symbol,
            "sources": sources,
            "note": note,
        }
        entries.append(apply_oot_hint(entry, hints.get(symbol, {})))

    for symbol, field, bit, note in OOT_EXTRA_SOURCES:
        add(symbol, [oot_bit_source("gOotExtraFlags", f"gOotExtraFlags.{field}", bit)], note)

    add("GERUDO_CARD", [oot_bit_source("inventoryQuest", "gOotSave.inventory.quest.value", 22)], "Gerudo Card quest item")
    add("WEIRD_EGG", [oot_mask_source("gOotTradeSave", "gOotTradeSave.child", 0x1FFE)], "Weird Egg child trade progression")
    add("ZELDA_LETTER", [oot_mask_source("gOotTradeSave", "gOotTradeSave.child", 0x1FFC)], "Zelda's Letter child trade progression")
    add(
        "POCKET_EGG",
        [
            oot_mask_source("gOotTradeSave", "gOotTradeSave.adult", 0x07FE),
            oot_flag_source("eventsItem", "gOotSave.context.eventsItem", 0x2C),
        ],
        "Pocket Egg adult trade progression",
    )

    for symbol, flag, note in OOT_EVENT_CHK_SOURCES:
        add(symbol, [oot_flag_source("eventsChk", "gOotSave.context.eventsChk", flag)], note)
    for symbol, flag, note in OOT_EVENT_ITEM_SOURCES:
        add(symbol, [oot_flag_source("eventsItem", "gOotSave.context.eventsItem", flag)], note)
    for symbol, flag, note in OOT_EVENT_MISC_SOURCES:
        add(symbol, [oot_flag_source("eventsMisc", "gOotSave.context.eventsMisc", flag)], note)

    known_symbols = {entry["symbol"] for entry in entries}
    for symbol, hint in hints.items():
        if symbol in known_symbols:
            continue
        sources: list[dict[str, Any]] = []
        for source in hint.get("sources", []):
            if source.get("group") in SUPPORTED_OOT_GROUPS:
                append_unique_source(sources, source)
        if not sources:
            continue
        entry: dict[str, Any] = {
            "symbol": symbol,
            "sources": sources,
        }
        for key in ("name", "note"):
            if hint.get(key):
                entry[key] = hint[key]
        entries.append(entry)

    return entries


def quest_source(symbol: str, quest_fields: dict[str, dict[str, Any]]) -> dict[str, Any] | None:
    field = QUEST_SYMBOL_FIELDS.get(symbol)
    if field is None:
        return None
    info = quest_fields.get(field)
    if info is None:
        return None
    return {
        "group": "inventoryQuest",
        "field": "gMmSave.info.inventory.quest.value",
        "mask": f"0x{info['logical_mask']:08x}",
    }


def boss_source(symbol: str) -> dict[str, Any] | None:
    bit = BOSS_SYMBOL_BITS.get(symbol)
    if bit is None:
        return None
    return {
        "group": "gMmExtraBoss",
        "field": "gMmExtraBoss",
        "mask": f"0x{1 << bit:08x}",
    }


def build_entries(repo_root: pathlib.Path, hints: dict[str, dict[str, Any]]) -> tuple[list[dict[str, Any]], list[str]]:
    data_root = repo_root / "packages/data/src"
    npc_ids = load_npc_ids(data_root / "defs/npc.yml")
    pool_names = load_pool_names(data_root / "pool/pool_mm.csv")
    mm_save_header = repo_root / "packages/generator/include/combo/mm/save.h"
    bitfields = parse_bitfield_structs(mm_save_header)
    quest_fields = parse_quest_fields(mm_save_header)
    week_events = parse_week_events(repo_root / "packages/generator/include/combo/common/events.h")
    discovered = discover_simple_sources(repo_root, bitfields, week_events)

    warnings: list[str] = []
    entries: list[dict[str, Any]] = []

    for symbol, code in sorted(npc_ids.items(), key=lambda item: item[1]):
        hint = hints.get(symbol, {})
        sources: list[dict[str, Any]] = []
        for source in hint.get("sources", []):
            if source.get("group") in SUPPORTED_GROUPS:
                append_unique_source(sources, source)
        if not sources:
            for source in discovered.get(symbol, []):
                append_unique_source(sources, source)
            q_source = quest_source(symbol, quest_fields)
            if q_source is not None:
                append_unique_source(sources, q_source)
            b_source = boss_source(symbol)
            if b_source is not None:
                append_unique_source(sources, b_source)

        entry: dict[str, Any] = {
            "code": f"0x{code:02x}",
            "symbol": symbol,
        }
        if sources:
            enriched_sources: list[dict[str, Any]] = []
            all_bits: list[int] = []
            byte_index: int | None = None
            mask: int | None = None
            for source in sources:
                enriched, bits, source_byte_index, source_mask = enrich_source(source, bitfields)
                append_unique_source(enriched_sources, enriched)
                all_bits.extend(bits)
                if byte_index is None and source_byte_index is not None:
                    byte_index = source_byte_index
                if mask is None and source_mask is not None:
                    mask = source_mask
            entry["sources"] = enriched_sources
            if hint.get("note"):
                entry["note"] = hint["note"]
            name = pool_names.get(symbol) or hint.get("name")
            if name:
                entry["name"] = name
            unique_bits = sorted(set(all_bits))
            if unique_bits:
                entry["bits"] = unique_bits
            if byte_index is not None and mask is not None:
                entry["byteIndex"] = byte_index
                entry["mask"] = mask
        elif symbol in pool_names or hint.get("name"):
            entry["name"] = pool_names.get(symbol) or hint["name"]
            warnings.append(f"{symbol}: no source found")

        entries.append(entry)

    return entries, warnings


def main() -> int:
    args = parse_args()
    repo_root = pathlib.Path(args.ootmm_repo).resolve()
    output_path = pathlib.Path(args.output).resolve()
    oot_output_path = pathlib.Path(args.oot_output).resolve() if args.oot_output else None

    if not repo_root.is_dir():
        print(f"OoTMM repository not found: {repo_root}", file=sys.stderr)
        return 1

    hint_path: pathlib.Path | None = None
    if args.hints:
        hint_path = pathlib.Path(args.hints).resolve()
    elif not args.no_existing_hints and output_path.is_file():
        hint_path = output_path

    oot_hint_path: pathlib.Path | None = None
    if args.oot_hints:
        oot_hint_path = pathlib.Path(args.oot_hints).resolve()
    elif oot_output_path is not None and not args.no_existing_hints and oot_output_path.is_file():
        oot_hint_path = oot_output_path

    try:
        entries, warnings = build_entries(repo_root, load_hints(hint_path))
        oot_entries = build_oot_entries(load_hints(oot_hint_path)) if oot_output_path is not None else []
    except Exception as exc:
        print(f"failed to generate special locations: {exc}", file=sys.stderr)
        return 1

    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(json.dumps(entries, indent=2) + "\n", encoding="utf-8")
    if oot_output_path is not None:
        oot_output_path.parent.mkdir(parents=True, exist_ok=True)
        oot_output_path.write_text(json.dumps(oot_entries, indent=2) + "\n", encoding="utf-8")

    if warnings:
        print(f"generated {len(entries)} entries with {len(warnings)} missing source hints", file=sys.stderr)
        for warning in warnings[:25]:
            print(f"warning: {warning}", file=sys.stderr)
        if len(warnings) > 25:
            print(f"warning: ... {len(warnings) - 25} more", file=sys.stderr)
    if oot_output_path is not None:
        print(f"generated {len(oot_entries)} OoT special location entries", file=sys.stderr)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
