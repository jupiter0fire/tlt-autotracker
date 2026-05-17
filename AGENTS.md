This project is an autotracker for the Randomizer OOTMM.
The autotracker’s data resides in the `ootmm-autotracker` directory.
In the OOTMM repository (subfolder `OoTMM`) you can find information on how data for obtained items and completed checks is stored in memory.
Only change `ootmm-autotracker`. Never change other subfolders.

There are the following helper scripts:

- analyze_mm_candidates.py: main dump-analysis script for scanning snapshot JSON files, finding plausible OoT foreign-save candidates inside MM payload, and comparing checksum/plausibility/richness metrics. Use this when you need to identify or compare possible foreign-save anchor addresses across one or more dump snapshots.
- export_dump_address_overview.py: exports CSV overviews of addresses from dump snapshots (resolved addresses + regions), split by game. Use this when you want a quick tabular view of address stability or drift across many snapshots.
- decode_save.py: quick byte-offset decoder for save data (currently written as an ad hoc script with embedded base64 sample data). Use this when you need a fast one-off check of a few known fields/offsets during investigation.
- check_bits.py: scaffold script for inspecting specific bits/flags in a dump JSON; currently more of a starting point than a finished tool. Use this when you are validating whether a particular event/check bit is set in a snapshot.
- analyze_conflicts.py: conflict analysis helper for xflag bit mappings (uses OoTMM data tables; useful when correlating dump-observed bits with location conflicts). Use this when observed dump bits seem ambiguous and you need to know which locations share the same underlying bit.
- solution.py: one-off analysis helper script around xflag bit positions/collisions; useful as reference but not a general CLI tool. Use this when reproducing or validating a specific conflict hypothesis without running the full analysis pipeline.
