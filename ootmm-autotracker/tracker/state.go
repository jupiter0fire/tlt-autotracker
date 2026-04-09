package tracker

import (
	"github.com/ootmm-autotracker/ootmm"
)

// State tracks the current and previous game state, computing deltas.
type State struct {
	prevItems  map[string]int
	prevChecks map[string]bool
	prevGame   ootmm.ActiveGame
}

func NewState() *State {
	return &State{
		prevItems:  make(map[string]int),
		prevChecks: make(map[string]bool),
	}
}

// ItemDiff represents a changed item.
type ItemDiff struct {
	ID  string `json:"id"`
	Qty int    `json:"qty"`
}

// CheckDiff represents a changed check.
type CheckDiff struct {
	ID      string `json:"id"`
	Checked bool   `json:"checked"`
}

// Update computes what changed since the last update.
// Returns changed items, changed checks, and whether the active game changed.
func (s *State) Update(gs *ootmm.GameState) (changedItems []ItemDiff, changedChecks []CheckDiff, gameChanged bool) {
	items := ootmm.ExtractItems(gs)
	checks := ootmm.ExtractChecks(gs)

	// Items diff
	currentItems := make(map[string]int, len(items))
	for _, it := range items {
		currentItems[it.ID] = it.Qty
		prev, existed := s.prevItems[it.ID]
		delta := it.Qty - prev
		if !existed {
			delta = it.Qty
		}
		if delta != 0 {
			changedItems = append(changedItems, ItemDiff{ID: it.ID, Qty: delta})
		}
	}
	// Emit negative deltas for items that disappeared
	for id, prevQty := range s.prevItems {
		if _, exists := currentItems[id]; !exists && prevQty != 0 {
			changedItems = append(changedItems, ItemDiff{ID: id, Qty: -prevQty})
		}
	}
	s.prevItems = currentItems

	// Checks diff
	currentChecks := make(map[string]bool, len(checks))
	for _, ch := range checks {
		currentChecks[ch.ID] = ch.Checked
		prev, existed := s.prevChecks[ch.ID]
		if !existed || prev != ch.Checked {
			changedChecks = append(changedChecks, CheckDiff{ID: ch.ID, Checked: ch.Checked})
		}
	}
	// Emit unchecked for checks that disappeared
	for id, prevChecked := range s.prevChecks {
		if _, exists := currentChecks[id]; !exists && prevChecked {
			changedChecks = append(changedChecks, CheckDiff{ID: id, Checked: false})
		}
	}
	s.prevChecks = currentChecks

	// Game change
	gameChanged = gs.ActiveGame != s.prevGame
	s.prevGame = gs.ActiveGame

	return
}

// FullState returns all items and checks as if everything changed (for initial sync).
func (s *State) FullState(gs *ootmm.GameState) ([]ItemDiff, []CheckDiff) {
	items := ootmm.ExtractItems(gs)
	checks := ootmm.ExtractChecks(gs)

	allItems := make([]ItemDiff, len(items))
	for i, it := range items {
		allItems[i] = ItemDiff{ID: it.ID, Qty: it.Qty}
	}

	allChecks := make([]CheckDiff, len(checks))
	for i, ch := range checks {
		allChecks[i] = CheckDiff{ID: ch.ID, Checked: ch.Checked}
	}

	return allItems, allChecks
}
