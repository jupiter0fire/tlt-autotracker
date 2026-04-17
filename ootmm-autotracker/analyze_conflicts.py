import csv
import pathlib
import struct
import re

SCENE_TYPES = {"chest": "chest", "collectible": "collect"}
XFLAG_TYPES = {"pot", "crate", "barrel", "grass", "tree", "bush", "rock", "soil", "fairy", "snowball", "hive", "rupee", "heart", "fairy_spot", "wonder", "butterfly", "redboulder", "icicle", "redice"}
XFLAG_TABLE_FILES = {
    "OOT": {
        "scenes": "packages/generator/data/static/xflag_table_oot_scenes.bin",
        "setups": "packages/generator/data/static/xflag_table_oot_setups.bin",
        "rooms": "packages/generator/data/static/xflag_table_oot_rooms.bin",
    }
}
XFLAG_COUNT_RE = re.compile(r"^#define\s+XFLAGS_COUNT_(OOT|MM)\s+0x([0-9a-fA-F]+)\s*$")

def load_symbol_ids(path):
    values = {}
    if not path.exists(): return values
    for line in path.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if not line or line.startswith("#"): continue
        name, sep, value = line.partition(":")
        if not sep: continue
        values[name.strip()] = int(value.strip(), 0)
    return values

def load_xflag_counts(path):
    counts = {}
    for line in path.read_text(encoding="utf-8").splitlines():
        match = XFLAG_COUNT_RE.match(line.strip())
        if match:
            game, value = match.groups()
            counts[game] = int(value, 16)
    return counts

def load_u16_table(path):
    data = path.read_bytes()
    return [v[0] for v in struct.iter_unpack(">H", data)]

def load_i16_table(path):
    data = path.read_bytes()
    return [v[0] for v in struct.iter_unpack(">h", data)]

def xflag_bit_position(scene_id, raw_id, scenes_table, setups_table, rooms_table, bit_limit):
    if scene_id < 0 or scene_id >= len(scenes_table): return None
    setup_id, room_id, slice_id, actor_id = (raw_id >> 14) & 0x3, (raw_id >> 8) & 0x3F, raw_id >> 16, raw_id & 0xFF
    setup_index = scenes_table[scene_id] + setup_id
    if setup_index < 0 or setup_index >= len(setups_table): return None
    room_index = setups_table[setup_index] + room_id * 12 + slice_id
    if room_index < 0 or room_index >= len(rooms_table): return None
    bit_pos = rooms_table[room_index] + actor_id
    return bit_pos if 0 <= bit_pos < bit_limit else None

repo_root = pathlib.Path("../OoTMM").resolve()
data_root = repo_root / "packages/data/src"
defs_root = data_root / "defs"
pool_root = data_root / "pool"
xflags_header = repo_root / "packages/generator/include/combo/xflags_data.h"

scenes = load_symbol_ids(defs_root / "scenes.yml")
xflag_counts = load_xflag_counts(xflags_header)
tables = {k: (load_u16_table(repo_root / v) if "rooms" not in k else load_i16_table(repo_root / v)) for k, v in XFLAG_TABLE_FILES["OOT"].items()}

bit_to_locations = {}
with (pool_root / "pool_oot.csv").open(newline="", encoding="utf-8") as h:
    for row in csv.DictReader(h, skipinitialspace=True):
        loc, typ, scn, val = row["location"].strip(), row["type"].strip(), row["scene"].strip(), row["id"].strip()
        if typ not in XFLAG_TYPES or not scn or not val: continue
        scene_id = scenes.get(f"OOT_{scn}")
        if scene_id is None: continue
        bit_pos = xflag_bit_position(scene_id, int(val, 0), tables["scenes"], tables["setups"], tables["rooms"], xflag_counts["OOT"] * 8)
        if bit_pos is not None:
            bit_to_locations.setdefault(bit_pos, []).append((loc, scn))

# Identify conflicts and map dungeons
dungeon_conflicts = {}
for bit, locs in bit_to_locations.items():
    if len(set(l[0] for l in locs)) > 1:
        print(f"Bit {bit} conflict: {', '.join(sorted(set(l[0] for l in locs)))}")
        for _, scn in locs:
            dungeon_conflicts[scn] = dungeon_conflicts.get(scn, 0) + 1

print("\nPer-Dungeon Summary:")
for scn, count in sorted(dungeon_conflicts.items()):
    print(f"{scn}: {count} conflicts")
