#!/usr/bin/env python3

import argparse
import asyncio
import json
from datetime import datetime
from typing import Any

import websockets


def parse_args() -> argparse.Namespace:
	parser = argparse.ArgumentParser(description="Debug the OoTMM autotracker WebSocket stream.")
	parser.add_argument("--url", default="ws://127.0.0.1:17026/", help="WebSocket URL")
	parser.add_argument(
		"--protocol",
		choices=("auto", "legacy", "raw"),
		default="auto",
		help="Handshake protocol override. 'auto' leaves mode selection to the advertised features.",
	)
	parser.add_argument(
		"--features",
		default="items,checks,gps",
		help="Comma-separated feature list to advertise in the handshake",
	)
	parser.add_argument(
		"--send-full",
		action="store_true",
		help="Request a full state snapshot after the handshake",
	)
	parser.add_argument(
		"--raw",
		action="store_true",
		help="Print full JSON frames instead of short summaries",
	)
	parser.add_argument(
		"--timeout",
		type=float,
		default=0,
		help="Stop after N seconds; 0 waits forever",
	)
	return parser.parse_args()


def format_raw_snapshot(payload: dict[str, Any]) -> str:
	chunk_entries = payload.get("chunks", [])
	if not isinstance(chunk_entries, list):
		return "raw snapshot with invalid chunks payload"

	total_length = 0
	chunk_names: list[str] = []
	for entry in chunk_entries:
		if not isinstance(entry, dict):
			continue
		name = entry.get("name")
		length = entry.get("length")
		if isinstance(name, str):
			chunk_names.append(name)
		if isinstance(length, int):
			total_length += length

	preview = ", ".join(chunk_names[:4]) if chunk_names else "no chunks"
	if len(chunk_names) > 4:
		preview = f"{preview}, +{len(chunk_names) - 4} more"

	return (
		f"raw seq={payload.get('sequence')} game={payload.get('game')} "
		f"saveIndex={payload.get('saveIndex')} chunks={len(chunk_names)} "
		f"bytes={total_length} [{preview}]"
	)


def format_inventory(items: dict[str, int]) -> str:
	entries = [f"{item_id} x{qty}" for item_id, qty in sorted(items.items()) if qty > 0]
	return ", ".join(entries) if entries else "no items"


def format_checks(current_checks: set[str], limit: int = 12) -> str:
	entries = sorted(current_checks)
	if not entries:
		return "no checks"
	if len(entries) <= limit:
		return ", ".join(entries)
	preview = ", ".join(entries[:limit])
	return f"{preview} (+{len(entries) - limit} more)"


def get_item_updates(payload: dict[str, Any], current_items: dict[str, int]) -> list[str]:
	updates: list[str] = []
	item_entries = payload.get("items", [])
	if not isinstance(item_entries, list):
		return updates

	diff = bool(payload.get("diff"))
	if diff:
		for entry in item_entries:
			if not isinstance(entry, dict):
				continue
			item_id = entry.get("id")
			qty = entry.get("qty")
			if not isinstance(item_id, str) or not isinstance(qty, int):
				continue

			previous_qty = current_items.get(item_id, 0)
			new_qty = previous_qty + qty
			if new_qty > 0:
				current_items[item_id] = new_qty
			else:
				current_items.pop(item_id, None)

			if qty > 0 and previous_qty == 0:
				updates.append(f"new item: {item_id} (+{qty}, now {new_qty})")
			elif qty > 0:
				updates.append(f"Item update: {item_id} (+{qty}, now {new_qty})")
			elif qty < 0:
				updates.append(f"Item update: {item_id} ({qty}, now {max(new_qty, 0)})")
	else:
		current_items.clear()
		for entry in item_entries:
			if not isinstance(entry, dict):
				continue
			item_id = entry.get("id")
			qty = entry.get("qty")
			if isinstance(item_id, str) and isinstance(qty, int) and qty > 0:
				current_items[item_id] = qty
		updates.append(f"current inventory: {format_inventory(current_items)}")

	return updates


def get_check_updates(payload: dict[str, Any], current_checks: set[str]) -> list[str]:
	updates: list[str] = []
	check_entries = payload.get("checks", [])
	if not isinstance(check_entries, list):
		return updates

	diff = bool(payload.get("diff"))
	if diff:
		for entry in check_entries:
			if not isinstance(entry, dict):
				continue
			name = entry.get("name") or entry.get("id")
			checked = entry.get("checked")
			if not isinstance(name, str) or not isinstance(checked, bool):
				continue

			if checked:
				if name not in current_checks:
					updates.append(f"new check: {name}")
				current_checks.add(name)
			else:
				if name in current_checks:
					updates.append(f"check removed: {name}")
				current_checks.discard(name)
	else:
		current_checks.clear()
		for entry in check_entries:
			if not isinstance(entry, dict):
				continue
			name = entry.get("name") or entry.get("id")
			checked = entry.get("checked")
			if isinstance(name, str) and checked is True:
				current_checks.add(name)
		updates.append(f"current checks: {format_checks(current_checks)}")

	return updates


async def recv_loop(ws: websockets.ClientConnection, raw: bool) -> None:
	current_items: dict[str, int] = {}
	current_checks: set[str] = set()

	async for message in ws:
		now = datetime.now().strftime("%H:%M:%S")
		try:
			payload = json.loads(message)
		except json.JSONDecodeError:
			print(f"[{now}] non-json: {message}")
			continue

		if raw:
			print(f"[{now}] {json.dumps(payload, ensure_ascii=True, sort_keys=True)}")
			continue

		msg_type = payload.get("type", "?")
		refresh = payload.get("refresh")
		if msg_type == "item":
			for update in get_item_updates(payload, current_items):
				print(f"[{now}] {update}")
			print(f"[{now}] item diff={payload.get('diff')} refresh={refresh} count={len(payload.get('items', []))}")
		elif msg_type == "check":
			for update in get_check_updates(payload, current_checks):
				print(f"[{now}] {update}")
			print(
				f"[{now}] check diff={payload.get('diff')} refresh={refresh} count={len(payload.get('checks', []))} totalChecked={len(current_checks)}"
			)
		elif msg_type == "location":
			print(
				f"[{now}] location refresh={refresh} game={payload.get('game')} scene=0x{int(payload.get('sceneId', 0)):02X}"
			)
		elif msg_type == "raw":
			print(f"[{now}] {format_raw_snapshot(payload)}")
		else:
			print(f"[{now}] {msg_type} refresh={refresh} payload={payload}")


async def main() -> None:
	args = parse_args()
	features = [feature.strip() for feature in args.features.split(",") if feature.strip()]

	async with websockets.connect(args.url, max_size=10 * 1024 * 1024) as ws:
		handshake = {
			"type": "handshake",
			"features": features,
			"flags": {},
		}
		if args.protocol != "auto":
			handshake["flags"]["protocol"] = args.protocol
		await ws.send(json.dumps(handshake))
		print(f"sent handshake: {handshake}")

		if args.send_full:
			request = {"type": "sendFull"}
			await ws.send(json.dumps(request))
			print(f"sent request: {request}")

		if args.timeout > 0:
			await asyncio.wait_for(recv_loop(ws, args.raw), timeout=args.timeout)
		else:
			await recv_loop(ws, args.raw)


if __name__ == "__main__":
	try:
		asyncio.run(main())
	except asyncio.TimeoutError:
		pass