This project is an autotracker for the Randomizer OOTMM.
The autotracker’s data resides in the `ootmm-autotracker` directory.
In the OOTMM repository (subfolder `OoTMM`) you can find information on how data for obtained items and completed checks is stored in memory.
Only change `ootmm-autotracker`. Never change other subfolders.

The dump-analysis helpers now live in ../scripts/autotracker:

- analyze_conflicts.py: inspect conflicting OoT xflag bit mappings against the current OoTMM data tables.
- check_bits.py: inspect specific byte masks inside migrated dump fixtures.
- decode_save.py: decode selected save-context fields from migrated dump fixtures.
- solution.py: probe specific OoT xflag raw IDs and print their collision set.

The old embedded-repo copies were removed as part of the Phase 4 migration.
