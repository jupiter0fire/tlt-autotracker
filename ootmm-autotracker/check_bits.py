import json
import base64
import sys

def get_bit(data, offset, bit_mask):
    if offset >= len(data):
        return None
    return (data[offset] & bit_mask) != 0

def check_dump(filename):
    with open(filename, 'r') as f:
        data = json.load(f)
    
    # The JSON structure might have many chunks. We need to find the one containing MmSaveCtx.
    # We need to know where AddrMmSaveCtx is. Let's look for it in the JSON.
    # It might be in 'regions' or 'memory'.
    
    regions = data.get('regions', {})
    mm_save_ctx_addr = regions.get('AddrMmSaveCtx')
    if mm_save_ctx_addr is None:
        # Try to find it in meta or something?
        # Actually, let's just look at the JSON structure again.
        pass

    # Let's search for the actual memory data.
    # Usually it's in data['chunks'] or data['memory']
    chunks = data.get('chunks', [])
    
    # Values we are looking for (relative to mmSaveContext):
    # week event byte 8 bit 0x80
    # collectible perm bits for scenes:
    # 0x40 (Stone Tower Temple)
    # 0x6e (Moon - Deku)
    # 0x6f (Moon - Goron)
    
    # Save context structure (approximate from OoTMM source):
    # WeekEventReg is offset 0xda0
    # SceneFlags starts at 0xd4? No, MM save context is different.
    # In MM:
    # weekEventReg starts at 0xda0 (100 bytes)
    # cycleSceneFlags starts at 0xde4
    #   each scene has:
    #     chest: 4 bytes
    #     switch: 4 bytes
    #     collectible: 4 bytes
    #     cleared: 4 bytes
    #   Total 16 bytes per scene.
    
    print(f"File: {filename}")
    # print(data.keys())
    # print(regions)
    
    # Need to find the memory buffer.
    # The JSON format seems to have a 'memory' field which is base64 encoded?
    # Or chunks?
    
    # Let's try to locate the memory and the address.
    # For now, let's just print the available keys and a bit of data.

if __name__ == "__main__":
    check_dump(sys.argv[1])
