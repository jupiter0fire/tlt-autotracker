#!/usr/bin/env node

/*
 * Compare OOT-Auto-Tracker Event checks (eventsItem range only) with
 * ootmm-autotracker OoT special-location eventsItem mappings.
 */

const path = require('path');

const repoRoot = path.resolve(__dirname, '..', '..');
const oatChecksPath = path.join(repoRoot, 'OOT-Auto-Tracker', 'src', 'main', 'checks.js');
const ootSpecialPath = path.join(repoRoot, 'ootmm-autotracker', 'ootmm', 'oot_special_locations.json');

// OOT-Auto-Tracker encodes OoT event bits as one 768-bit string:
// 0..223 = eventsChk, 224..287 = eventsItem, 288..767 = eventsMisc.
const EVENT_ITEM_START = 224;
const EVENT_ITEM_END = 288;

function requireOrFail(modulePath) {
  try {
    return require(modulePath);
  } catch (err) {
    console.error(`Failed to load ${modulePath}: ${err.message}`);
    process.exit(1);
  }
}

function parseJsonOrFail(modulePath) {
  try {
    return require(modulePath);
  } catch (err) {
    console.error(`Failed to parse JSON ${modulePath}: ${err.message}`);
    process.exit(1);
  }
}

function eventBitIndexToFlag(index) {
  const local = index - EVENT_ITEM_START;
  return Math.floor(local / 16) * 16 + (15 - (local % 16));
}

function findSingleSetBit(bits) {
  let idx = -1;
  for (let i = 0; i < bits.length; i++) {
    if (bits[i] !== '1') {
      continue;
    }
    if (idx !== -1) {
      return -2;
    }
    idx = i;
  }
  return idx;
}

function addMapArrayValue(map, key, value) {
  if (!map.has(key)) {
    map.set(key, []);
  }
  map.get(key).push(value);
}

function sortedNumericKeys(map) {
  return Array.from(map.keys()).sort((a, b) => a - b);
}

function extractOatEventItemChecks(checksRoot) {
  const out = new Map();
  const skipped = {
    notEventType: 0,
    noBits: 0,
    noSingleBit: 0,
    outsideEventItemRange: 0,
  };

  for (const region of checksRoot) {
    const areas = Array.isArray(region.areas) ? region.areas : [];
    for (const area of areas) {
      const checks = Array.isArray(area.checks) ? area.checks : [];
      for (const check of checks) {
        if (check.type !== 'Event') {
          skipped.notEventType++;
          continue;
        }
        if (typeof check.bits !== 'string' || check.bits.length === 0) {
          skipped.noBits++;
          continue;
        }

        const bitIndex = findSingleSetBit(check.bits);
        if (bitIndex < 0) {
          skipped.noSingleBit++;
          continue;
        }
        if (bitIndex < EVENT_ITEM_START || bitIndex >= EVENT_ITEM_END) {
          skipped.outsideEventItemRange++;
          continue;
        }

        const flag = eventBitIndexToFlag(bitIndex);
        addMapArrayValue(out, flag, {
          check: check.check,
          area: area.area,
          location: check.location,
          bitIndex,
        });
      }
    }
  }

  return { byFlag: out, skipped };
}

function extractOotSpecialEventItemMappings(ootSpecialEntries) {
  const out = new Map();
  let ignoredSources = 0;

  for (const entry of ootSpecialEntries) {
    if (!entry || !Array.isArray(entry.sources)) {
      continue;
    }

    for (const src of entry.sources) {
      if (src.group !== 'eventsItem') {
        ignoredSources++;
        continue;
      }
      if (typeof src.flag !== 'number') {
        continue;
      }

      addMapArrayValue(out, src.flag, {
        symbol: entry.symbol || '(no symbol)',
        note: entry.note || '',
      });
    }
  }

  return { byFlag: out, ignoredSources };
}

function printFlagSection(title, flags, leftMap, rightMap) {
  console.log(`\n${title}: ${flags.length}`);
  for (const flag of flags) {
    console.log(`  - flag ${flag}`);
    const left = leftMap.get(flag) || [];
    const right = rightMap.get(flag) || [];
    for (const item of left) {
      console.log(`    OOT-Auto-Tracker: ${item.check} (area: ${item.area}, bitIndex: ${item.bitIndex})`);
    }
    for (const item of right) {
      const suffix = item.note ? `; note: ${item.note}` : '';
      console.log(`    ootmm-autotracker: ${item.symbol}${suffix}`);
    }
  }
}

function main() {
  const oatChecksModule = requireOrFail(oatChecksPath);
  const oatRoot = oatChecksModule.file_checks;
  if (!Array.isArray(oatRoot)) {
    console.error('OOT-Auto-Tracker checks.js does not expose file_checks as an array.');
    process.exit(1);
  }

  const ootSpecial = parseJsonOrFail(ootSpecialPath);
  if (!Array.isArray(ootSpecial)) {
    console.error('oot_special_locations.json is not an array.');
    process.exit(1);
  }

  const oat = extractOatEventItemChecks(oatRoot);
  const oot = extractOotSpecialEventItemMappings(ootSpecial);

  const oatFlags = new Set(oat.byFlag.keys());
  const ootFlags = new Set(oot.byFlag.keys());

  const sharedFlags = sortedNumericKeys(new Map([...oat.byFlag].filter(([flag]) => ootFlags.has(flag))));
  const oatOnlyFlags = sortedNumericKeys(new Map([...oat.byFlag].filter(([flag]) => !ootFlags.has(flag))));
  const ootOnlyFlags = sortedNumericKeys(new Map([...oot.byFlag].filter(([flag]) => !oatFlags.has(flag))));

  console.log('OOT eventsItem mapping audit (eventsChk intentionally excluded)');
  console.log(`OOT-Auto-Tracker eventsItem flags: ${oat.byFlag.size}`);
  console.log(`ootmm-autotracker eventsItem flags: ${oot.byFlag.size}`);
  console.log(`Shared flags: ${sharedFlags.length}`);
  console.log(`OOT-Auto-Tracker only flags: ${oatOnlyFlags.length}`);
  console.log(`ootmm-autotracker only flags: ${ootOnlyFlags.length}`);

  console.log('\nOOT-Auto-Tracker extraction summary:');
  console.log(`  Event checks skipped because not Event type: ${oat.skipped.notEventType}`);
  console.log(`  Event checks skipped because bits missing: ${oat.skipped.noBits}`);
  console.log(`  Event checks skipped because not single-bit: ${oat.skipped.noSingleBit}`);
  console.log(`  Event checks skipped because outside eventsItem range: ${oat.skipped.outsideEventItemRange}`);

  printFlagSection('Shared flag details', sharedFlags, oat.byFlag, oot.byFlag);
  printFlagSection('OOT-Auto-Tracker-only flags', oatOnlyFlags, oat.byFlag, oot.byFlag);
  printFlagSection('ootmm-autotracker-only flags', ootOnlyFlags, oat.byFlag, oot.byFlag);
}

main();
