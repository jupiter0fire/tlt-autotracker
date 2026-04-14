#!/usr/bin/env python3

import argparse
import csv
import json
import pathlib
import re
import struct
import sys


SCENE_TYPES = {
    "chest": "chest",
    "collectible": "collect",
}

XFLAG_TYPES = {
    "pot",
    "crate",
    "barrel",
    "grass",
    "tree",
    "bush",
    "rock",
    "soil",
    "fairy",
    "snowball",
    "hive",
    "rupee",
    "heart",
    "fairy_spot",
    "wonder",
    "butterfly",
    "redboulder",
    "icicle",
    "redice",
}

BITMAP_SPECS = {
    ("OOT", "npc"): "npcOot",
    ("MM", "npc"): "npcMm",
    ("OOT", "shop"): "shopsOot",
    ("MM", "shop"): "shopsMm",
    ("OOT", "scrub"): "scrubsOot",
    ("OOT", "sr"): "srOot",
}

XFLAG_TABLE_FILES = {
    "OOT": {
        "scenes": "packages/generator/data/static/xflag_table_oot_scenes.bin",
        "setups": "packages/generator/data/static/xflag_table_oot_setups.bin",
        "rooms": "packages/generator/data/static/xflag_table_oot_rooms.bin",
    },
    "MM": {
        "scenes": "packages/generator/data/static/xflag_table_mm_scenes.bin",
        "setups": "packages/generator/data/static/xflag_table_mm_setups.bin",
        "rooms": "packages/generator/data/static/xflag_table_mm_rooms.bin",
    },
}

XFLAG_COUNT_RE = re.compile(r"^#define\s+XFLAGS_COUNT_(OOT|MM)\s+0x([0-9a-fA-F]+)\s*$")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Generate autotracker location mappings from an OoTMM checkout."
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


def load_symbol_ids(path: pathlib.Path) -> dict[str, int]:
    values: dict[str, int] = {}
    for raw_line in path.read_text(encoding="utf-8").splitlines():
        line = raw_line.strip()
        if not line or line.startswith("#"):
            continue
        name, sep, value = line.partition(":")
        if not sep:
            continue
        name = name.strip()
        if not name.startswith(("OOT_", "MM_")):
            continue
        try:
            values[name] = int(value.strip(), 0)
        except ValueError:
            continue
    return values


def load_xflag_counts(path: pathlib.Path) -> dict[str, int]:
    counts: dict[str, int] = {}
    for raw_line in path.read_text(encoding="utf-8").splitlines():
        match = XFLAG_COUNT_RE.match(raw_line.strip())
        if not match:
            continue
        game, value = match.groups()
        counts[game] = int(value, 16)
    if set(counts) != {"OOT", "MM"}:
        raise ValueError(f"failed to read XFLAG counts from {path}")
    return counts


def load_u16_table(path: pathlib.Path) -> list[int]:
    data = path.read_bytes()
    if len(data) % 2 != 0:
        raise ValueError(f"expected even-sized u16 table: {path}")
    return [value[0] for value in struct.iter_unpack(">H", data)]


def load_i16_table(path: pathlib.Path) -> list[int]:
    data = path.read_bytes()
    if len(data) % 2 != 0:
        raise ValueError(f"expected even-sized i16 table: {path}")
    return [value[0] for value in struct.iter_unpack(">h", data)]


def scene_check_key(game: str, scene_id: int, kind: str, bit: int) -> str:
    return f"{game}_{kind}_{scene_id}_{bit}"


def xflag_bit_position(
    scene_id: int,
    raw_id: int,
    scenes_table: list[int],
    setups_table: list[int],
    rooms_table: list[int],
    bit_limit: int,
) -> int | None:
    if scene_id < 0 or scene_id >= len(scenes_table):
        return None

    setup_id = (raw_id >> 14) & 0x3
    room_id = (raw_id >> 8) & 0x3F
    slice_id = raw_id >> 16
    actor_id = raw_id & 0xFF

    setup_index = scenes_table[scene_id] + setup_id
    if setup_index < 0 or setup_index >= len(setups_table):
        return None

    room_index = setups_table[setup_index] + room_id * 12 + slice_id
    if room_index < 0 or room_index >= len(rooms_table):
        return None

    bit_pos = rooms_table[room_index] + actor_id
    if bit_pos < 0 or bit_pos >= bit_limit:
        return None
    return bit_pos


def add_unique_mapping(table: dict[object, str], conflicts: set[object], key: object, name: str) -> None:
    existing = table.get(key)
    if existing is None:
        table[key] = name
        return
    if existing != name:
        conflicts.add(key)


def finalize_mapping(table: dict[object, str], conflicts: set[object]) -> dict[object, str]:
    return {key: value for key, value in table.items() if key not in conflicts}


def build_location_mapping(repo_root: pathlib.Path) -> dict[str, object]:
    data_root = repo_root / "packages/data/src"
    defs_root = data_root / "defs"
    pool_root = data_root / "pool"
    xflags_header = repo_root / "packages/generator/include/combo/xflags_data.h"

    scenes = load_symbol_ids(defs_root / "scenes.yml")
    npcs = load_symbol_ids(defs_root / "npc.yml")
    xflag_counts = load_xflag_counts(xflags_header)

    xflag_tables: dict[str, dict[str, list[int]]] = {}
    for game, files in XFLAG_TABLE_FILES.items():
        xflag_tables[game] = {
            "scenes": load_u16_table(repo_root / files["scenes"]),
            "setups": load_u16_table(repo_root / files["setups"]),
            "rooms": load_i16_table(repo_root / files["rooms"]),
        }

    scene_checks_raw: dict[str, str] = {}
    scene_conflicts: set[str] = set()
    bitmap_checks_raw: dict[tuple[str, int], str] = {}
    bitmap_conflicts: set[tuple[str, int]] = set()
    symbol_checks_raw: dict[tuple[str, str], str] = {}

    for game, pool_name in (("OOT", "pool_oot.csv"), ("MM", "pool_mm.csv")):
        with (pool_root / pool_name).open(newline="", encoding="utf-8") as handle:
            reader = csv.DictReader(handle, skipinitialspace=True)
            for row in reader:
                location = row["location"].strip()
                check_type = row["type"].strip()
                scene_name = row["scene"].strip()
                value = row["id"].strip()
                if not location or not scene_name or not value:
                    continue

                if check_type == "npc":
                    symbol_checks_raw.setdefault((game, value), location)

                scene_kind = SCENE_TYPES.get(check_type)
                if scene_kind is not None:
                    scene_id = scenes.get(f"{game}_{scene_name}")
                    if scene_id is None:
                        continue
                    try:
                        bit = int(value, 0)
                    except ValueError:
                        continue
                    if bit < 0 or bit >= 32:
                        continue
                    key = scene_check_key(game, scene_id, scene_kind, bit)
                    add_unique_mapping(scene_checks_raw, scene_conflicts, key, location)
                    continue

                bitmap_block = BITMAP_SPECS.get((game, check_type))
                if bitmap_block is not None:
                    if check_type == "npc":
                        npc_id = npcs.get(f"{game}_{value}")
                        if npc_id is None:
                            continue
                        add_unique_mapping(bitmap_checks_raw, bitmap_conflicts, (bitmap_block, npc_id), location)
                        continue

                    try:
                        bit = int(value, 0)
                    except ValueError:
                        continue
                    if bit < 0:
                        continue
                    add_unique_mapping(bitmap_checks_raw, bitmap_conflicts, (bitmap_block, bit), location)
                    continue

                if check_type not in XFLAG_TYPES:
                    continue

                scene_id = scenes.get(f"{game}_{scene_name}")
                if scene_id is None:
                    continue
                try:
                    raw_id = int(value, 0)
                except ValueError:
                    continue
                tables = xflag_tables[game]
                bit_pos = xflag_bit_position(
                    scene_id,
                    raw_id,
                    tables["scenes"],
                    tables["setups"],
                    tables["rooms"],
                    xflag_counts[game] * 8,
                )
                if bit_pos is None:
                    continue
                block = "xflagsOot" if game == "OOT" else "xflagsMm"
                add_unique_mapping(bitmap_checks_raw, bitmap_conflicts, (block, bit_pos), location)

    scene_checks = [
        {"key": key, "name": name}
        for key, name in sorted(finalize_mapping(scene_checks_raw, scene_conflicts).items())
    ]
    bitmap_checks = [
        {"block": block, "bit": bit, "name": name}
        for (block, bit), name in sorted(finalize_mapping(bitmap_checks_raw, bitmap_conflicts).items())
    ]
    symbol_checks = [
        {"game": game, "symbol": symbol, "name": name}
        for (game, symbol), name in sorted(symbol_checks_raw.items())
    ]

    return {
        "scene": scene_checks,
        "bitmap": bitmap_checks,
        "symbols": symbol_checks,
    }


def main() -> int:
    args = parse_args()
    repo_root = pathlib.Path(args.ootmm_repo).resolve()
    output_path = pathlib.Path(args.output).resolve()

    if not repo_root.is_dir():
        print(f"OoTMM repository not found: {repo_root}", file=sys.stderr)
        return 1

    mapping = build_location_mapping(repo_root)
    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(json.dumps(mapping, indent=2) + "\n", encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())