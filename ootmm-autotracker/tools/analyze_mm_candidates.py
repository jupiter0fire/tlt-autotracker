#!/usr/bin/env python3

import argparse
import base64
import json
import pathlib
from typing import Dict, Iterable, List, Optional


ADDR_MM_PAYLOAD = 0x80730000
MM_PAYLOAD_SIZE = 0x50000
OOT_SAVE_SIZE = 0x1354
SHARED_CUSTOM_SAVE_SIZE = 0x870
MAX_FOREIGN_OOT_CHECKSUM_DELTA = 0x1000
MIN_PLAUSIBLE_OOT_EMPTY_INVENTORY_SLOTS = 4
EMPTY_INVENTORY_ITEM = 0xFF

OOT_OFF_AGE = 0x04
OOT_OFF_SCENE_ID = 0x66
OOT_OFF_INV_ITEMS = 0x74
OOT_OFF_UPGRADES = 0xA0
OOT_OFF_GOLD_TOKENS = 0xD0
OOT_OFF_DUNGEON_KEYS = 0xBC
OOT_OFF_CHECKSUM = 0x1352
OOT_PERM_COUNT = 124

CHECK_BITMAP_NAMES = (
    "xflagsOot",
    "npcOot",
    "shopsOot",
    "scrubsOot",
    "srOot",
    "xflagsMm",
    "npcMm",
    "shopsMm",
)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Compare plausible OoT foreign-save candidates inside MM payload dumps."
    )
    parser.add_argument("snapshots", nargs="+", help="Debug snapshot JSON files to analyze")
    parser.add_argument(
        "--top",
        type=int,
        default=5,
        help="How many exact candidates to show in the richness table (default: 5)",
    )
    return parser.parse_args()


def load_shared_layout(repo_root: pathlib.Path) -> Dict[str, Dict[str, int]]:
    layout_path = repo_root / "ootmm" / "inventory_slots.json"
    with layout_path.open("r", encoding="utf-8") as handle:
        data = json.load(handle)
    bitmaps = data["catalog"]["shared"]["bitmaps"]
    return {
        bitmap["name"]: {"offset": bitmap["offset"], "size": bitmap["size"]}
        for bitmap in bitmaps
        if bitmap["name"] in CHECK_BITMAP_NAMES
    }


def load_snapshot(path: pathlib.Path) -> dict:
    with path.open("r", encoding="utf-8") as handle:
        return json.load(handle)


def region_map(snapshot: dict) -> Dict[str, dict]:
    return {region["name"]: region for region in snapshot.get("regions", [])}


def resolved_address_map(snapshot: dict) -> Dict[str, int]:
    addresses = {}
    for entry in snapshot.get("resolvedAddresses", []):
        address = entry.get("address")
        if address:
            addresses[entry["logical_id"]] = int(address, 16)
    return addresses


def decode_region(region: dict) -> bytes:
    encoding = region.get("encoding")
    if encoding != "base64":
        raise ValueError(f"unsupported encoding for region {region.get('name')}: {encoding!r}")
    return base64.b64decode(region["data"])


def be_u16(data: bytes, offset: int) -> int:
    return int.from_bytes(data[offset : offset + 2], "big")


def be_u32(data: bytes, offset: int) -> int:
    return int.from_bytes(data[offset : offset + 4], "big")


def oot_checksum(data: bytes) -> int:
    checksum = 0
    for offset in range(0, OOT_SAVE_SIZE, 2):
        if offset == OOT_OFF_CHECKSUM:
            continue
        checksum = (checksum + be_u16(data, offset)) & 0xFFFF
    return checksum


def oot_checksum_delta(data: bytes) -> Optional[int]:
    if len(data) < OOT_SAVE_SIZE:
        return None
    expected = be_u16(data, OOT_OFF_CHECKSUM)
    if expected == 0:
        return None
    checksum = oot_checksum(data)
    delta = abs(checksum - expected)
    if delta > 0x8000:
        delta = 0x10000 - delta
    return delta


def is_plausible_oot_save(data: bytes) -> bool:
    if len(data) < OOT_SAVE_SIZE:
        return False

    age = be_u32(data, OOT_OFF_AGE)
    if age > 1:
        return False

    scene_id = be_u16(data, OOT_OFF_SCENE_ID)
    if scene_id >= OOT_PERM_COUNT:
        return False

    empty_slots = sum(
        1 for item_id in data[OOT_OFF_INV_ITEMS : OOT_OFF_INV_ITEMS + 24] if item_id == EMPTY_INVENTORY_ITEM
    )
    if empty_slots < MIN_PLAUSIBLE_OOT_EMPTY_INVENTORY_SLOTS:
        return False

    gold_tokens = be_u16(data, OOT_OFF_GOLD_TOKENS)
    if gold_tokens > 100:
        return False

    for index in range(19):
        keys = int.from_bytes(data[OOT_OFF_DUNGEON_KEYS + index : OOT_OFF_DUNGEON_KEYS + index + 1], "big", signed=True)
        if keys < -1 or keys > 9:
            return False

    return True


def count_bits(blob: bytes) -> int:
    return sum(byte.bit_count() for byte in blob)


def shared_check_counts(payload: bytes, offset: int, shared_layout: Dict[str, Dict[str, int]]) -> Optional[Dict[str, int]]:
    if offset < SHARED_CUSTOM_SAVE_SIZE:
        return None
    start = offset - SHARED_CUSTOM_SAVE_SIZE
    end = start + SHARED_CUSTOM_SAVE_SIZE
    if end > len(payload):
        return None

    shared = payload[start:end]
    counts = {}
    for name in CHECK_BITMAP_NAMES:
        bitmap = shared_layout.get(name)
        if bitmap is None:
            continue
        bitmap_start = bitmap["offset"]
        bitmap_end = bitmap_start + bitmap["size"]
        counts[name] = count_bits(shared[bitmap_start:bitmap_end])
    return counts


def candidate_report(
    payload: bytes,
    offset: int,
    shared_layout: Dict[str, Dict[str, int]],
) -> Optional[dict]:
    if offset < 0 or offset + OOT_SAVE_SIZE > len(payload):
        return None

    data = payload[offset : offset + OOT_SAVE_SIZE]
    delta = oot_checksum_delta(data)
    plausible = is_plausible_oot_save(data)
    shared_counts = shared_check_counts(payload, offset, shared_layout)
    non_empty_items = sum(
        1 for item_id in data[OOT_OFF_INV_ITEMS : OOT_OFF_INV_ITEMS + 24] if item_id != EMPTY_INVENTORY_ITEM
    )

    report = {
        "offset": offset,
        "address": ADDR_MM_PAYLOAD + offset,
        "plausible": plausible,
        "checksum_delta": delta,
        "checksum_exact": delta == 0 if delta is not None else False,
        "age": be_u32(data, OOT_OFF_AGE),
        "scene_id": be_u16(data, OOT_OFF_SCENE_ID),
        "non_empty_items": non_empty_items,
        "gold_tokens": be_u16(data, OOT_OFF_GOLD_TOKENS),
        "upgrades": be_u32(data, OOT_OFF_UPGRADES),
        "shared_counts": shared_counts or {},
    }
    report["shared_total"] = sum(report["shared_counts"].values())
    report["richness"] = report["shared_total"] * 100 + report["non_empty_items"] * 10 + report["gold_tokens"]
    return report


def scan_oot_candidates(payload: bytes, shared_layout: Dict[str, Dict[str, int]]) -> List[dict]:
    candidates: List[dict] = []
    for offset in range(0, len(payload) - OOT_SAVE_SIZE + 1, 16):
        report = candidate_report(payload, offset, shared_layout)
        if report is None or not report["plausible"]:
            continue
        delta = report["checksum_delta"]
        if delta is None or delta > MAX_FOREIGN_OOT_CHECKSUM_DELTA:
            continue
        candidates.append(report)
    return candidates


def format_candidate(report: Optional[dict]) -> str:
    if report is None:
        return "unavailable"
    delta = report["checksum_delta"]
    delta_text = "n/a" if delta is None else f"0x{delta:04x}"
    shared_counts = ", ".join(
        f"{name}={count}" for name, count in report["shared_counts"].items() if count
    )
    if not shared_counts:
        shared_counts = "none"
    return (
        f"addr=0x{report['address']:08x} off=0x{report['offset']:05x} "
        f"exact={report['checksum_exact']} delta={delta_text} plausible={report['plausible']} "
        f"scene=0x{report['scene_id']:02x} items={report['non_empty_items']} gold={report['gold_tokens']} "
        f"shared={report['shared_total']} [{shared_counts}] upgrades=0x{report['upgrades']:08x}"
    )


def analyze_snapshot(path: pathlib.Path, shared_layout: Dict[str, Dict[str, int]], comparison_addresses: Iterable[int]) -> None:
    snapshot = load_snapshot(path)
    regions = region_map(snapshot)
    resolved = resolved_address_map(snapshot)
    mm_payload_region = regions.get("mmPayload")
    if mm_payload_region is None:
        raise ValueError(f"snapshot {path} is missing mmPayload region")
    payload = decode_region(mm_payload_region)
    if len(payload) != MM_PAYLOAD_SIZE:
        raise ValueError(f"snapshot {path} has mmPayload size {len(payload):#x}, expected {MM_PAYLOAD_SIZE:#x}")

    candidates = scan_oot_candidates(payload, shared_layout)
    exact_candidates = [candidate for candidate in candidates if candidate["checksum_exact"]]
    selected_addr = resolved.get("oot.foreign_save_in_mm_payload")
    selected_report = None
    if selected_addr is not None:
        selected_report = candidate_report(payload, selected_addr - ADDR_MM_PAYLOAD, shared_layout)

    first_exact = exact_candidates[0] if exact_candidates else None
    richest_exact = None
    if exact_candidates:
        richest_exact = max(
            exact_candidates,
            key=lambda candidate: (candidate["richness"], -candidate["offset"]),
        )

    summary = snapshot.get("summary", {})
    print(f"== {path.name} ==")
    print(
        "summary: "
        f"active={summary.get('activeGame')} saveIndex={summary.get('saveIndex')} "
        f"items={len(summary.get('items', []))} checks={len(summary.get('checks', []))}"
    )
    if selected_addr is not None:
        print(f"resolved oot.foreign_save_in_mm_payload: 0x{selected_addr:08x}")
    else:
        print("resolved oot.foreign_save_in_mm_payload: unavailable")
    print(f"plausible OoT candidates in mmPayload: {len(candidates)} total, {len(exact_candidates)} exact checksum")
    print(f"selected candidate: {format_candidate(selected_report)}")
    print(f"first exact by offset: {format_candidate(first_exact)}")
    print(f"richest exact candidate: {format_candidate(richest_exact)}")

    for address in sorted(set(comparison_addresses)):
        if address == 0:
            continue
        label = "selected" if address == selected_addr else "reference"
        report = candidate_report(payload, address - ADDR_MM_PAYLOAD, shared_layout)
        print(f"{label} probe 0x{address:08x}: {format_candidate(report)}")

    if exact_candidates:
        print("top exact candidates by richness:")
        ranked = sorted(
            exact_candidates,
            key=lambda candidate: (candidate["richness"], candidate["checksum_exact"], -candidate["offset"]),
            reverse=True,
        )
        for candidate in ranked[: TOP_COUNT]:
            print(f"  {format_candidate(candidate)}")
    print()


if __name__ == "__main__":
    args = parse_args()
    TOP_COUNT = args.top
    repo_root = pathlib.Path(__file__).resolve().parents[1]
    shared_layout = load_shared_layout(repo_root)

    snapshot_paths = [pathlib.Path(path).resolve() for path in args.snapshots]
    snapshots = [load_snapshot(path) for path in snapshot_paths]
    comparison_addresses = []
    for snapshot in snapshots:
        resolved = resolved_address_map(snapshot)
        address = resolved.get("oot.foreign_save_in_mm_payload")
        if address is not None:
            comparison_addresses.append(address)

    for path in snapshot_paths:
        analyze_snapshot(path, shared_layout, comparison_addresses)