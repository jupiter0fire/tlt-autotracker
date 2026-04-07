#!/usr/bin/env python3

import argparse
import asyncio
import json
from datetime import datetime

import websockets


def parse_args() -> argparse.Namespace:
	parser = argparse.ArgumentParser(description="Debug the OoTMM autotracker WebSocket stream.")
	parser.add_argument("--url", default="ws://127.0.0.1:17026/", help="WebSocket URL")
	parser.add_argument(
		"--features",
		default="items,checks,gps",
		help="Comma-separated Magpie feature list to advertise in the handshake",
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


async def recv_loop(ws: websockets.ClientConnection, raw: bool) -> None:
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
			print(f"[{now}] item diff={payload.get('diff')} refresh={refresh} count={len(payload.get('items', []))}")
		elif msg_type == "check":
			print(f"[{now}] check diff={payload.get('diff')} refresh={refresh} count={len(payload.get('checks', []))}")
		elif msg_type == "location":
			print(
				f"[{now}] location refresh={refresh} game={payload.get('game')} scene=0x{int(payload.get('sceneId', 0)):02X}"
			)
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