This is an autotracker for [thelasttracker.org](https://www.thelasttracker.org/).

## Autotracking Setup for Project64-EM

Only stable randomizer versions v30.1 to v32.3 are supported (no dev seeds). Autotracking only works after you import a spoiler log.

Download the autotracker for Windows or Linux, and also download the adapter Lua file from [the latest releases](https://github.com/jupiter0fire/tlt-autotracker/releases/latest). Put the adapter Lua in the same folder as the Multiworld/Coop Lua script, inside the Scripts folder of Project64-EM.

1. Generate a seed and open it in Project64-EM.
2. Start the autotracker.
3. In Project64-EM, open File -> Lua Scripts and double-click `pj64-adapter.lua`.
4. Upload the spoiler log to thelasttracker.org.
5. If The Last Tracker does not connect automatically, click Auto. Check if the autotracker shows that it is connected to Project64-EM and the tracker.
6. Potentially, your browser will display a popup regarding access to the autotracker. You need to grant access there once.
7. The autotracker currently tracks items and locations, but not entrances, song events, or similar.

A green outline around the Auto button means the autotracker connected successfully. An orange outline means no connection to the autotracker has been established yet.

## Autotracking Setup for RetroArch

Only stable randomizer versions v30.1 to v32.3 are supported (no dev seeds). Autotracking only works after you import a spoiler log.

Download the autotracker for Windows or Linux from [the latest releases](https://github.com/jupiter0fire/tlt-autotracker/releases/latest). If this is your first time using autotracking, enable Show Advanced Settings under Settings -> User Interface. Then enable Network Commands under Settings -> Network and leave the Network Command Port set to 55355.

1. Generate a seed and open it in RetroArch.
2. Start the autotracker.
3. Upload the spoiler log to thelasttracker.org.
4. If The Last Tracker does not connect automatically, click Auto. Check if the autotracker shows that it is connected to RetroArch and the tracker.
5. Potentially, your browser will display a popup regarding access to the autotracker. You need to grant access there once.
6. The autotracker currently tracks items and locations, but not entrances, song events, or similar.

A green outline around the Auto button means the autotracker connected successfully. An orange outline means no connection to the autotracker has been established yet.

## Autotracking Setup for Ares

Only randomizer versions v30.1 to v32.3 are supported (no dev seeds). Autotracking only works after you import a spoiler log.

Download the autotracker for Windows or Linux from [the latest releases](https://github.com/jupiter0fire/tlt-autotracker/releases/latest). If this is your first time using autotracking, you have to enable GDB debugging in Ares. Toggle "Enabled" and "Use IPv4" under Settings -> Debug. Leave the Port set to 9123.

1. Generate a seed and open it in Ares.
2. Start the autotracker.
3. Upload the spoiler log to thelasttracker.org.
4. If The Last Tracker does not connect automatically, click Auto. Check if the autotracker shows that it is connected to Ares and the tracker.
5. Potentially, your browser will display a popup regarding access to the autotracker. You need to grant access there once.
6. The autotracker currently tracks items and locations, but not entrances, song events, or similar.

A green outline around the Auto button means the autotracker connected successfully. An orange outline means no connection to the autotracker has been established yet. You cannot use the multiclient and autotracking at the same time, because Ares only allows one external connection

## Credits

- The first version, and especially the integration of RetroArch, was based on the [Magpie LADX autotracker](https://github.com/kbranch/Magpie-LADX-Autotracker).
- Information about the storage locations of items and checks comes from [OoTMM](https://github.com/OoTMM/OoTMM).

## Licensing

[MIT License](./LICENSE.txt)
