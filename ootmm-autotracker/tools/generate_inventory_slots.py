#!/usr/bin/env python3

import argparse
import json
import pathlib
import re
import sys


SLOT_DEFINE_RE = re.compile(r"^#define\s+(ITS_(OOT|MM)_[A-Z0-9_]+)\s+0x([0-9a-fA-F]+)\s*$")
GI_ID_RE = re.compile(r"^-\s+\{\s+id:\s+([A-Z0-9_]+),")

OOT_OVERRIDES = {
    "STICKS": "STICK",
    "NUTS": "DEKU_NUTS",
    "ARROW_FIRE": "ARROW_FIRE",
    "SPELL_FIRE": "SPELL_FIRE",
    "BOMBCHU": "BOMBCHUS",
    "ARROW_ICE": "ARROW_ICE",
    "SPELL_WIND": "SPELL_WIND",
    "MAGIC_BEAN": "MAGIC_BEAN",
    "HAMMER": "HAMMER",
    "ARROW_LIGHT": "ARROW_LIGHT",
    "SPELL_LOVE": "SPELL_LOVE",
    "BOTTLE": "BOTTLE_1",
    "BOTTLE2": "BOTTLE_2",
    "BOTTLE3": "BOTTLE_3",
    "BOTTLE4": "BOTTLE_4",
    "TRADE_ADULT": "ADULT_TRADE",
    "TRADE_CHILD": "CHILD_TRADE",
}

MM_OVERRIDES = {
    "ARROW_FIRE": "ARROW_FIRE",
    "ARROW_ICE": "ARROW_ICE",
    "ARROW_LIGHT": "ARROW_LIGHT",
    "TRADE1": "TRADE_1",
    "BOMBCHU": "BOMBCHU",
    "STICKS": "STICK",
    "NUTS": "NUT",
    "BEANS": "MAGIC_BEAN",
    "TRADE2": "TRADE_2",
    "KEG": "POWDER_KEG",
    "PICTOBOX": "PICTOGRAPH_BOX",
    "TRADE3": "TRADE_3",
    "BOTTLE": "BOTTLE_1",
    "BOTTLE2": "BOTTLE_2",
    "BOTTLE3": "BOTTLE_3",
    "BOTTLE4": "BOTTLE_4",
    "BOTTLE5": "BOTTLE_5",
    "BOTTLE6": "BOTTLE_6",
}

EXPECTED_SLOT_COUNTS = {
    "OOT": 24,
    "MM": 48,
}

SLOT_QUANTITY_RULES = {
    "ITS_OOT_OCARINA": {"stages": [0x07, 0x08]},
    "ITS_OOT_HOOKSHOT": {"stages": [0x0A, 0x0B]},
    "ITS_OOT_MAGIC_BEAN": {"useBeansCount": True},
    "ITS_OOT_TRADE_ADULT": {
        "stages": [0x2D, 0x2E, 0x2F, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x14],
        "maxWithBottle": True,
    },
    "ITS_OOT_TRADE_CHILD": {
        "stages": [0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2A, 0x2B, 0x9C, 0x9D, 0x14],
        "maxWithBottle": True,
    },
    "ITS_MM_OCARINA": {"stages": [0x05, 0x00]},
    "ITS_MM_TRADE1": {"stages": [0xB0, 0x28, 0x29, 0x2A, 0x2B, 0x2C]},
    "ITS_MM_TRADE2": {"stages": [0xAE, 0xB1, 0xB3, 0x2D, 0x2E]},
    "ITS_MM_HOOKSHOT": {"stages": [0x11, 0x0F]},
    "ITS_MM_GREAT_FAIRY_SWORD": {"stages": [0x10, 0xB5]},
    "ITS_MM_TRADE3": {"stages": [0xAF, 0xB2, 0xB4, 0x2F, 0x30]},
}

SHARED_STORAGE = {
    "baseOffset": 0x18000,
    "stride": 0x4000,
    "trackedSize": 0x800,
    "bitmaps": [
        {"name": "soulsEnemyOot", "offset": 0x7CC, "size": 8},
        {"name": "soulsEnemyMm", "offset": 0x7D4, "size": 8},
        {"name": "soulsBossOot", "offset": 0x7DC, "size": 2},
        {"name": "soulsBossMm", "offset": 0x7DE, "size": 1},
        {"name": "soulsNpcOot", "offset": 0x7DF, "size": 8},
        {"name": "soulsNpcMm", "offset": 0x7E7, "size": 8},
        {"name": "soulsAnimalOot", "offset": 0x7EF, "size": 2},
        {"name": "soulsAnimalMm", "offset": 0x7F1, "size": 2},
        {"name": "soulsMiscOot", "offset": 0x7F3, "size": 1},
        {"name": "soulsMiscMm", "offset": 0x7F4, "size": 1},
    ],
}

SOUL_SOURCE_SPECS = [
    {"prefix": "OOT_SOUL_ENEMY_", "block": "soulsEnemyOot"},
    {"prefix": "OOT_SOUL_BOSS_", "block": "soulsBossOot"},
    {"prefix": "OOT_SOUL_NPC_", "block": "soulsNpcOot"},
    {"prefix": "OOT_SOUL_ANIMAL_", "block": "soulsAnimalOot"},
    {"prefix": "OOT_SOUL_MISC_", "block": "soulsMiscOot"},
    {"prefix": "MM_SOUL_ENEMY_", "block": "soulsEnemyMm"},
    {"prefix": "MM_SOUL_BOSS_", "block": "soulsBossMm"},
    {"prefix": "MM_SOUL_NPC_", "block": "soulsNpcMm"},
    {"prefix": "MM_SOUL_ANIMAL_", "block": "soulsAnimalMm"},
    {"prefix": "MM_SOUL_MISC_", "block": "soulsMiscMm"},
]

SPECIAL_ITEM_SOURCES = [
    {"itemId": "MM_HAMMER", "source": {"kind": "oot-extra-bit", "record": 4, "bit": 6}},
    {"itemId": "MM_SPELL_FIRE", "source": {"kind": "oot-extra-bit", "record": 5, "bit": 16}},
    {"itemId": "MM_MOON_TEAR", "source": {"kind": "oot-extra-bit", "record": 5, "bit": 17}},
    {"itemId": "MM_DEED_LAND", "source": {"kind": "oot-extra-bit", "record": 5, "bit": 18}},
    {"itemId": "MM_DEED_SWAMP", "source": {"kind": "oot-extra-bit", "record": 5, "bit": 19}},
    {"itemId": "MM_DEED_MOUNTAIN", "source": {"kind": "oot-extra-bit", "record": 5, "bit": 20}},
    {"itemId": "MM_DEED_OCEAN", "source": {"kind": "oot-extra-bit", "record": 5, "bit": 21}},
    {"itemId": "MM_SPELL_WIND", "source": {"kind": "oot-extra-bit", "record": 5, "bit": 22}},
    {"itemId": "MM_BOOTS_IRON", "source": {"kind": "oot-extra-bit", "record": 5, "bit": 23}},
    {"itemId": "MM_TUNIC_GORON", "source": {"kind": "oot-extra-bit", "record": 5, "bit": 24}},
    {"itemId": "MM_ROOM_KEY", "source": {"kind": "oot-extra-bit", "record": 5, "bit": 25}},
    {"itemId": "MM_LETTER_TO_MAMA", "source": {"kind": "oot-extra-bit", "record": 5, "bit": 26}},
    {"itemId": "MM_SPELL_LOVE", "source": {"kind": "oot-extra-bit", "record": 5, "bit": 27}},
    {"itemId": "MM_BOOTS_HOVER", "source": {"kind": "oot-extra-bit", "record": 5, "bit": 28}},
    {"itemId": "MM_TUNIC_ZORA", "source": {"kind": "oot-extra-bit", "record": 5, "bit": 29}},
    {"itemId": "MM_LETTER_TO_KAFEI", "source": {"kind": "oot-extra-bit", "record": 5, "bit": 30}},
    {"itemId": "MM_PENDANT_OF_MEMORIES", "source": {"kind": "oot-extra-bit", "record": 5, "bit": 31}},
    {"itemId": "MM_WALLET5", "source": {"kind": "oot-extra-bit", "record": 13, "bit": 0}},
    {"itemId": "MM_STONE_OF_AGONY", "source": {"kind": "oot-extra-bit", "record": 13, "bit": 1}},
]


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Generate autotracker inventory slot mappings from an OoTMM checkout."
    )
    parser.add_argument(
        "--ootmm-repo",
        required=True,
        help="Path to the OoTMM repository root.",
    )
    parser.add_argument(
        "--output",
        required=True,
        help="Path to the output JSON file.",
    )
    return parser.parse_args()


def tracker_id_for(slot_name: str, game: str) -> str:
    suffix = slot_name.removeprefix(f"ITS_{game}_")
    overrides = OOT_OVERRIDES if game == "OOT" else MM_OVERRIDES
    suffix = overrides.get(suffix, suffix)
    return f"{game}_{suffix}"


def extract_slots(items_header: pathlib.Path) -> dict[str, list[dict[str, object]]]:
    slots: dict[str, list[dict[str, object]]] = {"OOT": [], "MM": []}

    for line in items_header.read_text(encoding="utf-8").splitlines():
        match = SLOT_DEFINE_RE.match(line)
        if not match:
            continue

        slot_name, game, raw_index = match.groups()
        index = int(raw_index, 16)
        entry = {
            "index": index,
            "slot": slot_name,
            "itemId": tracker_id_for(slot_name, game),
        }
        quantity = SLOT_QUANTITY_RULES.get(slot_name)
        if quantity is not None:
            entry["quantity"] = quantity
        slots[game].append(entry)

    for game, entries in slots.items():
        entries.sort(key=lambda entry: int(entry["index"]))
        expected = EXPECTED_SLOT_COUNTS[game]
        if len(entries) != expected:
            raise ValueError(
                f"expected {expected} {game} slots in {items_header}, found {len(entries)}"
            )

        indices = [int(entry["index"]) for entry in entries]
        if indices != list(range(expected)):
            raise ValueError(f"{game} slot indices are not contiguous: {indices}")

    return {"oot": slots["OOT"], "mm": slots["MM"]}


def extract_gi_ids(gi_defs: pathlib.Path) -> list[str]:
    gi_ids: list[str] = []

    for line in gi_defs.read_text(encoding="utf-8").splitlines():
        match = GI_ID_RE.match(line)
        if match:
            gi_ids.append(match.group(1))

    return gi_ids


def collect_prefixed_ids(gi_ids: list[str], prefix: str) -> list[str]:
    return [item_id for item_id in gi_ids if item_id.startswith(prefix)]


def ensure_ids_exist(gi_ids: list[str], required_ids: list[str], label: str) -> None:
    available = set(gi_ids)
    missing = [item_id for item_id in required_ids if item_id not in available]
    if missing:
        raise ValueError(f"missing {label} IDs in gi.yml: {', '.join(missing)}")


def build_catalog(gi_defs: pathlib.Path) -> dict[str, object]:
    gi_ids = extract_gi_ids(gi_defs)
    bitmap_sizes = {bitmap["name"]: bitmap["size"] for bitmap in SHARED_STORAGE["bitmaps"]}

    items: list[dict[str, object]] = []
    for spec in SOUL_SOURCE_SPECS:
        soul_ids = collect_prefixed_ids(gi_ids, spec["prefix"])
        max_bits = bitmap_sizes[spec["block"]] * 8
        if len(soul_ids) > max_bits:
            raise ValueError(
                f"{spec['block']} only has space for {max_bits} bits, found {len(soul_ids)} items"
            )
        for bit, item_id in enumerate(soul_ids):
            items.append(
                {
                    "itemId": item_id,
                    "source": {
                        "kind": "shared-bitmap-bit",
                        "block": spec["block"],
                        "bit": bit,
                    },
                }
            )

    ensure_ids_exist(gi_ids, [entry["itemId"] for entry in SPECIAL_ITEM_SOURCES], "special item")
    items.extend(SPECIAL_ITEM_SOURCES)

    return {
        "shared": SHARED_STORAGE,
        "items": items,
    }


def main() -> int:
    args = parse_args()
    repo_root = pathlib.Path(args.ootmm_repo).resolve()
    items_header = repo_root / "packages/generator/include/combo/data/items.h"
    gi_defs = repo_root / "packages/data/src/defs/gi.yml"
    output_path = pathlib.Path(args.output).resolve()

    if not items_header.is_file():
        print(f"items header not found: {items_header}", file=sys.stderr)
        return 1
    if not gi_defs.is_file():
        print(f"gi definitions not found: {gi_defs}", file=sys.stderr)
        return 1

    mapping = extract_slots(items_header)
    mapping["catalog"] = build_catalog(gi_defs)
    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(json.dumps(mapping, indent=2) + "\n", encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
