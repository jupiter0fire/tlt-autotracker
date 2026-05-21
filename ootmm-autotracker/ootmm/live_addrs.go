package ootmm

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type liveAddrFile struct {
	SchemaVersion int              `json:"schemaVersion"`
	Oot           liveAddrGameFile `json:"oot"`
	Mm            liveAddrGameFile `json:"mm"`
}

type liveAddrGameFile struct {
	ComboCtx string `json:"comboCtx"`
	SaveCtx  string `json:"saveCtx"`
	Payload  string `json:"payload"`
}

type resolvedLiveAddrs struct {
	ComboCtxOot uint32
	ComboCtxMm  uint32
	OotSaveCtx  uint32
	MmSaveCtx   uint32
	OotPayload  uint32
	MmPayload   uint32
}

//go:embed live_addrs.json
var embeddedLiveAddrs []byte

var loadedLiveAddrs = mustLoadLiveAddrs()

var (
	AddrComboCtxOot = loadedLiveAddrs.ComboCtxOot
	AddrComboCtxMm  = loadedLiveAddrs.ComboCtxMm
	AddrOotSaveCtx  = loadedLiveAddrs.OotSaveCtx
	AddrMmSaveCtx   = loadedLiveAddrs.MmSaveCtx
	AddrOotPayload  = loadedLiveAddrs.OotPayload
	AddrMmPayload   = loadedLiveAddrs.MmPayload
)

func mustLoadLiveAddrs() resolvedLiveAddrs {
	var file liveAddrFile
	if err := json.Unmarshal(embeddedLiveAddrs, &file); err != nil {
		panic(fmt.Sprintf("parse embedded live address mapping: %v", err))
	}
	if file.SchemaVersion != 1 {
		panic(fmt.Sprintf("unsupported live address schema version %d", file.SchemaVersion))
	}

	return resolvedLiveAddrs{
		ComboCtxOot: mustParseHexUint32("oot.comboCtx", file.Oot.ComboCtx),
		ComboCtxMm:  mustParseHexUint32("mm.comboCtx", file.Mm.ComboCtx),
		OotSaveCtx:  mustParseHexUint32("oot.saveCtx", file.Oot.SaveCtx),
		MmSaveCtx:   mustParseHexUint32("mm.saveCtx", file.Mm.SaveCtx),
		OotPayload:  mustParseHexUint32("oot.payload", file.Oot.Payload),
		MmPayload:   mustParseHexUint32("mm.payload", file.Mm.Payload),
	}
}

func mustParseHexUint32(field string, raw string) uint32 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		panic(fmt.Sprintf("live address field %s is empty", field))
	}
	value, err := strconv.ParseUint(raw, 0, 32)
	if err != nil {
		panic(fmt.Sprintf("parse live address field %s: %v", field, err))
	}
	return uint32(value)
}
