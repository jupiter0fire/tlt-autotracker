#!/usr/bin/env python3

import argparse
import json
import pathlib
import re
import sys


SLOT_DEFINE_RE = re.compile(r"^#define\s+(ITS_(OOT|MM)_[A-Z0-9_]+)\s+0x([0-9a-fA-F]+)\s*$")
GI_ID_RE = re.compile(r"^-\s+\{\s+id:\s+([A-Z0-9_]+),")

SOUL_GROUP_PREFIXES = {
    "oot": {
        "enemy": "OOT_SOUL_ENEMY_",
        "boss": "OOT_SOUL_BOSS_",
        "npc": "OOT_SOUL_NPC_",
        "animal": "OOT_SOUL_ANIMAL_",
        "misc": "OOT_SOUL_MISC_",
    },
    "mm": {
        "enemy": "MM_SOUL_ENEMY_",
        "boss": "MM_SOUL_BOSS_",
        "npc": "MM_SOUL_NPC_",
        "animal": "MM_SOUL_ANIMAL_",
        "misc": "MM_SOUL_MISC_",
    },
}

MM_SPECIAL_IDS = {
    "mmItems": [
        "MM_HAMMER",
    ],
    "mmTrade1": [
        "MM_SPELL_FIRE",
        "MM_MOON_TEAR",
        "MM_DEED_LAND",
        "MM_DEED_SWAMP",
        "MM_DEED_MOUNTAIN",
        "MM_DEED_OCEAN",
    ],
    "mmTrade2": [
        "MM_SPELL_WIND",
        "MM_BOOTS_IRON",
        "MM_TUNIC_GORON",
        "MM_ROOM_KEY",
        "MM_LETTER_TO_MAMA",
    ],
    "mmTrade3": [
        "MM_SPELL_LOVE",
        "MM_BOOTS_HOVER",
        "MM_TUNIC_ZORA",
        "MM_LETTER_TO_KAFEI",
        "MM_PENDANT_OF_MEMORIES",
    ],
    "mmFlags3": [
        "MM_WALLET5",
        "MM_STONE_OF_AGONY",
    ],
}

OOT_OVERRIDES = {
    "STICKS": "DEKU_STICKS",
    "NUTS": "DEKU_NUTS",
    "ARROW_FIRE": "FIRE_ARROWS",
    "SPELL_FIRE": "DINS_FIRE",
    "BOMBCHU": "BOMBCHUS",
    "ARROW_ICE": "ICE_ARROWS",
    "SPELL_WIND": "FARORES_WIND",
    "MAGIC_BEAN": "MAGIC_BEANS",
    "HAMMER": "MEGATON_HAMMER",
    "ARROW_LIGHT": "LIGHT_ARROWS",
    "SPELL_LOVE": "NAYRUS_LOVE",
    "BOTTLE": "BOTTLE_1",
    "BOTTLE2": "BOTTLE_2",
    "BOTTLE3": "BOTTLE_3",
    "BOTTLE4": "BOTTLE_4",
    "TRADE_ADULT": "ADULT_TRADE",
    "TRADE_CHILD": "CHILD_TRADE",
}

MM_OVERRIDES = {
    "ARROW_FIRE": "FIRE_ARROWS",
    "ARROW_ICE": "ICE_ARROWS",
    "ARROW_LIGHT": "LIGHT_ARROWS",
    "TRADE1": "TRADE_1",
    "BOMBCHU": "BOMBCHUS",
    "STICKS": "DEKU_STICKS",
    "NUTS": "DEKU_NUTS",
    "BEANS": "MAGIC_BEANS",
    "TRADE2": "TRADE_2",
    "KEG": "POWDER_KEG",
    "PICTOBOX": "PICTOGRAPH",
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
        slots[game].append(
            {
                "index": index,
                "slot": slot_name,
                "itemId": tracker_id_for(slot_name, game),
            }
        )

    for game, entries in slots.items():
        entries.sort(key=lambda entry: int(entry["index"]))
        expected = EXPECTED_SLOT_COUNTS[game]
        if len(entries) != expected:
            raise ValueError(
                f"expected {expected} {game} slots in {items_header}, found {len(entries)}"
            )

        indices = [int(entry["index"]) for entry in entries]
        if indices != list(range(expected)):
            raise ValueError(
                f"{game} slot indices are not contiguous: {indices}"
            )

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

    souls: dict[str, dict[str, list[str]]] = {}
    for game, groups in SOUL_GROUP_PREFIXES.items():
        souls[game] = {}
        for group_name, prefix in groups.items():
            souls[game][group_name] = collect_prefixed_ids(gi_ids, prefix)

    special: dict[str, list[str]] = {}
    for label, item_ids in MM_SPECIAL_IDS.items():
        ensure_ids_exist(gi_ids, item_ids, label)
        special[label] = item_ids

    return {
        "souls": souls,
        "special": special,
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