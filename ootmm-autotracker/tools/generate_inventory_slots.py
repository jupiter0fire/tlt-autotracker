#!/usr/bin/env python3

import argparse
import json
import pathlib
import re
import sys


SLOT_DEFINE_RE = re.compile(r"^#define\s+(ITS_(OOT|MM)_[A-Z0-9_]+)\s+0x([0-9a-fA-F]+)\s*$")

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


def main() -> int:
    args = parse_args()
    repo_root = pathlib.Path(args.ootmm_repo).resolve()
    items_header = repo_root / "packages/generator/include/combo/data/items.h"
    output_path = pathlib.Path(args.output).resolve()

    if not items_header.is_file():
        print(f"items header not found: {items_header}", file=sys.stderr)
        return 1

    mapping = extract_slots(items_header)
    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(json.dumps(mapping, indent=2) + "\n", encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())