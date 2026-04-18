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
