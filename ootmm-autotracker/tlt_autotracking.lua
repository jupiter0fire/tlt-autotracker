-- PJ64 Adapter for ootmm-autotracker
--
-- This script runs inside Project64 and connects to the ootmm-autotracker.
-- It serves memory read requests over a simple binary protocol.
--
-- The autotracker must be started with: --pj64
--
-- Usage: Load this script in PJ64's Lua scripting console while a game
-- is running. The script will connect to the autotracker and stay
-- connected, processing memory read requests.

local HOST = '127.0.0.1'
local PORT = 55190

-- Op codes
local OP_MEMREAD_BULK = 10

-- Adapter greeting.  The autotracker only serves scripts that send the
-- current greeting right after connecting, so outdated Lua scripts are
-- rejected.  Bump PROTOCOL_VERSION whenever the wire protocol changes
-- incompatibly and keep it in sync with adapterGreeting in
-- pj64/server.go ("OAT" .. PROTOCOL_VERSION).
local GREETING_MAGIC = "OAT"
local PROTOCOL_VERSION = 2

-- ComboCtx addresses (virtual) for transition detection.
-- During a game switch OoTMM's Context_Init writes the magic "OoT+MM<3"
-- to the target game's ComboCtx.  While the magic is visible the save
-- data is not yet fully initialized and must not be read.
local COMBOCTX_OOT_ADDR = 0x80006584
local COMBOCTX_MM_ADDR  = 0x80098280
local COMBOCTX_MAGIC    = "OoT+MM<3"

local function is_game_transition()
  local magic_oot = ""
  local magic_mm  = ""
  pcall(function()
    -- Read 8 bytes of magic from OoT ComboCtx
    local parts = {}
    for i = 0, 7 do
      parts[i + 1] = string.char(memory.read_u8(COMBOCTX_OOT_ADDR + i))
    end
    magic_oot = table.concat(parts)
  end)
  pcall(function()
    -- Read 8 bytes of magic from MM ComboCtx
    local parts = {}
    for i = 0, 7 do
      parts[i + 1] = string.char(memory.read_u8(COMBOCTX_MM_ADDR + i))
    end
    magic_mm = table.concat(parts)
  end)
  return magic_oot == COMBOCTX_MAGIC or magic_mm == COMBOCTX_MAGIC
end

function connect()
  local s = socket.tcp(HOST, PORT)
  if s == nil then
    print('Waiting for autotracker at ' .. HOST .. ':' .. PORT .. '...')
    socket.sleep(2)
    return connect()
  end
  return s
end

-- Split a u32 (returned by memory.read_u32, big-endian N64 value) into
-- 4 bytes in big-endian order and append them as a single string.
local function append_u32_be(parts, val)
  local b3 = val % 256
  val = (val - b3) / 256
  local b2 = val % 256
  val = (val - b2) / 256
  local b1 = val % 256
  val = (val - b1) / 256
  local b0 = val % 256
  parts[#parts + 1] = string.char(b0, b1, b2, b3)
end

print("=== OoTMM Autotracker - PJ64 Adapter ===")
print("Connecting to autotracker at " .. HOST .. ":" .. PORT)

local s = connect()

-- Identify this script to the autotracker.  The server rejects
-- connections without a matching greeting, so an old script can never
-- feed the tracker stale data.
local ok_send, send_err = pcall(function()
  s:send(GREETING_MAGIC .. PROTOCOL_VERSION)
end)
if not ok_send then
  print("Failed to send greeting: " .. tostring(send_err))
end

print("Connected to autotracker!")

while true do
  local ok, err = pcall(function()
    local data = s:recv(1)
    if data == nil then
      error("disconnected")
    end

    local op = binary.unpack_u8(data)

    if op == OP_MEMREAD_BULK then
      local addr = binary.unpack_u32(s:recv(4))
      local size = binary.unpack_u32(s:recv(4))

      -- Guard: if the game is mid-transition (ComboCtx magic is still
      -- visible at either address), return zeroes so the autotracker
      -- sees a clean invalid frame instead of partially-written garbage.
      if is_game_transition() then
        s:send(string.char(0):rep(size))
      else
        local parts = {}
        local i = 0

        -- Handle unaligned head bytes
        while i < size and (addr + i) % 4 ~= 0 do
          parts[#parts + 1] = string.char(memory.read_u8(addr + i))
          i = i + 1
        end

        -- Read aligned middle with u32 (4x faster than byte-by-byte)
        while i + 3 < size do
          append_u32_be(parts, memory.read_u32(addr + i))
          i = i + 4
        end

        -- Handle trailing bytes
        while i < size do
          parts[#parts + 1] = string.char(memory.read_u8(addr + i))
          i = i + 1
        end

        s:send(table.concat(parts))
      end
    end
  end)

  if not ok then
    print("Error: " .. tostring(err))
    print("Reconnecting...")
    pcall(function() s:close() end)
    socket.sleep(1)
    s = connect()
    print("Reconnected!")
  end
end
