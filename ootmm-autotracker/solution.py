import pathlib
import sys

# Add tools to path
sys.path.append('/home/silke/stuff/repos/tlt_autotracker/ootmm-autotracker/tools')
import generate_locations

repo_root = pathlib.Path('/home/silke/stuff/repos/tlt_autotracker/OoTMM')

# Load tables
oot_tables = {
    "scenes": generate_locations.load_u16_table(repo_root / "packages/generator/data/static/xflag_table_oot_scenes.bin"),
    "setups": generate_locations.load_u16_table(repo_root / "packages/generator/data/static/xflag_table_oot_setups.bin"),
    "rooms": generate_locations.load_i16_table(repo_root / "packages/generator/data/static/xflag_table_oot_rooms.bin"),
}
bit_limit = 0x2e8 # From XFLAGS_COUNT_OOT

# Scene DODONGO_CAVERN is 0x01
scene_id = 0x01
raw_ids = [0x00305, 0x00306, 0x00307]

results = {}
for rid in raw_ids:
    bit = generate_locations.xflag_bit_position(
        scene_id, rid, 
        oot_tables["scenes"], oot_tables["setups"], oot_tables["rooms"], 
        bit_limit
    )
    results[rid] = bit

# Now check for collisions
mapping = generate_locations.build_location_mapping(repo_root)
# We want to see if any location has scene_check with game='OOT', kind='xflag', bit=bit
# The mapping from build_location_mapping returns location name -> data
# Data contains scene_check: { "game": ..., "kind": ..., "scene": ..., "bit": ... } or similar structure?
# Let's inspect mapping structure.

reverse_mapping = {}
for loc_name, loc_data in mapping.items():
    checks = loc_data.get('checks', [])
    for check in checks:
        if check.get('game') == 'OOT' and check.get('type') == 'xflag':
            b = check.get('bit')
            if b is not None:
                if b not in reverse_mapping:
                    reverse_mapping[b] = []
                reverse_mapping[b].append(loc_name)

for rid, bit in results.items():
    collisions = reverse_mapping.get(bit, [])
    print(f"Raw ID 0x{rid:05x}: Bit {bit}, Collisions: {', '.join(collisions)}")

