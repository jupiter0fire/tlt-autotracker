import json

def analyze_json(filename):
    with open(filename, 'r') as f:
        data = json.load(f)
    summary = data.get('summary', {})
    items = summary.get('items', [])
    # Summary checks might be list of dicts or list of strings
    checks = summary.get('checks', [])
    return {
        'item_count': len(items),
        'check_count': len(checks),
        'items': {item['id']: item['qty'] for item in items},
        'checks': [json.dumps(c, sort_keys=True) if isinstance(c, dict) else c for c in checks]
    }

before = analyze_json('/home/silke/stuff/repos/tlt_autotracker/ootmm-autotracker/memory-dumps/before-grass-20260419-192606.json')
after = analyze_json('/home/silke/stuff/repos/tlt_autotracker/ootmm-autotracker/memory-dumps/after-grass-20260419-192638.json')

print(f"Before: Items={before['item_count']}, Checks={before['check_count']}")
print(f"After:  Items={after['item_count']}, Checks={after['check_count']}")

# Items differ
diff_items = []
all_item_ids = set(before['items'].keys()) | set(after['items'].keys())
for item_id in all_item_ids:
    qty_before = before['items'].get(item_id, 0)
    qty_after = after['items'].get(item_id, 0)
    if qty_before != qty_after:
        diff_items.append((item_id, qty_before, qty_after))

# Checks differ
set_before = set(before['checks'])
set_after = set(after['checks'])
only_before = set_before - set_after
only_after = set_after - set_before

print(f"Different Items: {len(diff_items)}")
for item in diff_items:
    print(f"  {item[0]}: {item[1]} -> {item[2]}")

print(f"Checks only in Before: {len(only_before)}")
for check in only_before:
    print(f"  {check}")
print(f"Checks only in After: {len(only_after)}")
for check in only_after:
    print(f"  {check}")
