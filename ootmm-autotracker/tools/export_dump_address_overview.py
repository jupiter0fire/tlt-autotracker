#!/usr/bin/env python3

import argparse
import csv
import json
import pathlib
import sys
from collections import defaultdict
from typing import Dict, Iterable, List, Tuple


GAME_OUTPUT_NAMES = {
    "OoT": "oot-address-overview.csv",
    "MM": "mm-address-overview.csv",
}


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Export one CSV per game with snapshot address columns from resolvedAddresses and regions."
    )
    parser.add_argument(
        "dump_dir",
        nargs="?",
        default="memory-dumps",
        help="Directory containing debug snapshot JSON files (default: memory-dumps)",
    )
    return parser.parse_args()


def load_snapshot(path: pathlib.Path) -> dict:
    with path.open("r", encoding="utf-8") as handle:
        return json.load(handle)


def iter_address_columns(snapshot: dict) -> Iterable[Tuple[str, str]]:
    for entry in snapshot.get("resolvedAddresses") or []:
        logical_id = entry.get("logical_id")
        if not logical_id:
            continue
        yield f"resolved:{logical_id}", entry.get("address", "")

    for entry in snapshot.get("regions") or []:
        name = entry.get("name")
        if not name:
            continue
        yield f"region:{name}", entry.get("address", "")


def snapshot_game(snapshot: dict) -> str:
    summary = snapshot.get("summary") or {}
    game = summary.get("activeGame")
    if not game:
        raise ValueError("snapshot is missing summary.activeGame")
    return game


def collect_rows(dump_dir: pathlib.Path) -> Dict[str, List[dict]]:
    rows_by_game: Dict[str, List[dict]] = defaultdict(list)

    for path in sorted(dump_dir.glob("*.json")):
        snapshot = load_snapshot(path)
        game = snapshot_game(snapshot)
        if game not in GAME_OUTPUT_NAMES:
            print(f"Skipping {path.name}: unsupported activeGame {game!r}", file=sys.stderr)
            continue

        row = {
            "file": path.name,
            "createdAt": snapshot.get("createdAt", ""),
            "activeGame": game,
        }
        row.update(dict(iter_address_columns(snapshot)))
        rows_by_game[game].append(row)

    return rows_by_game


def write_csv(path: pathlib.Path, rows: List[dict]) -> None:
    if not rows:
        return

    fixed_columns = ["file", "createdAt", "activeGame"]
    dynamic_columns = sorted({key for row in rows for key in row.keys()} - set(fixed_columns))
    columns = fixed_columns + dynamic_columns

    with path.open("w", encoding="utf-8", newline="") as handle:
        writer = csv.DictWriter(handle, fieldnames=columns)
        writer.writeheader()
        for row in rows:
            writer.writerow({column: row.get(column, "") for column in columns})


def main() -> int:
    args = parse_args()
    dump_dir = pathlib.Path(args.dump_dir)

    if not dump_dir.is_dir():
        print(f"Dump directory not found: {dump_dir}", file=sys.stderr)
        return 1

    rows_by_game = collect_rows(dump_dir)
    for game, output_name in GAME_OUTPUT_NAMES.items():
        write_csv(dump_dir / output_name, rows_by_game.get(game, []))

    written = [name for game, name in GAME_OUTPUT_NAMES.items() if rows_by_game.get(game)]
    if written:
        print("Wrote:")
        for name in written:
            print(f"  {dump_dir / name}")
    else:
        print("No supported OoT/MM snapshots found.", file=sys.stderr)
        return 1

    return 0


if __name__ == "__main__":
    raise SystemExit(main())