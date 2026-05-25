This is an autotracker for [thelasttracker.org](https://www.thelasttracker.org/).

## Autotracking Setup for Project64-EM

Only stable randomizer versions (starting from v30.1) are supported (no dev seeds). Autotracking only works after you import a spoiler log.

Download the autotracker for Windows or Linux, and also download the adapter Lua file from [the latest releases](https://github.com/jupiter0fire/tlt-autotracker/releases/latest). Put the adapter Lua in the same folder as the Multiworld/Coop Lua script, inside the Scripts folder of Project64-EM.

1. Generate a seed and open it in Project64-EM.
2. Start the autotracker.
3. In Project64-EM, open File -> Lua Scripts and double-click `pj64-adapter.lua`.
4. Upload the spoiler log to thelasttracker.org.
5. If The Last Tracker does not connect automatically, click Auto.

A green outline around the Auto button means the autotracker connected successfully. An orange outline means no connection to the autotracker has been established yet.

## Autotracking Setup for RetroArch

Only randomizer version v30.1 is supported (no dev seeds). Autotracking only works after you import a spoiler log.

Download the autotracker for Windows or Linux from [the latest releases](https://github.com/jupiter0fire/tlt-autotracker/releases/latest). If this is your first time using autotracking, enable Show Advanced Settings under Settings -> User Interface. Then enable Network Commands under Settings -> Network and leave the Network Command Port set to 55355.

1. Generate a seed and open it in RetroArch.
2. Start the autotracker.
3. Upload the spoiler log to thelasttracker.org.
4. If The Last Tracker does not connect automatically, click Auto.

A green outline around the Auto button means the autotracker connected successfully. An orange outline means no connection to the autotracker has been established yet.

## Credits

- The first version, and especially the integration of RetroArch, was based on the [Magpie LADX autotracker](https://github.com/kbranch/Magpie-LADX-Autotracker).
- Information about the storage locations of items and checks comes from [OoTMM](https://github.com/OoTMM/OoTMM).

## Licensing

[MIT License](./LICENSE.txt)
