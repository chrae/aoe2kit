package scenario

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const blankScenarioSeedSHA256 = "8399e4aa78eb0c34caa225d0d2b15ac507350e8b8eaf1c38ee94b104830dbd9d"
const defaultBlankDummyUnit = 598

type BlankOptions struct {
	PlayerCount       int
	HumanSlots        int
	Timestamp         int
	ClearTriggers     bool
	KeepSeedTriggers  bool
	DummyStarters     bool
	DummyUnit         int
	DummyStartX       float64
	DummyStartY       float64
	DummySpacing      float64
	InactiveRestHuman bool
	GaiaActive        bool
}

type BlankReport struct {
	Output             string `json:"output"`
	SeedSHA256         string `json:"seed_sha256"`
	Version            string `json:"version"`
	PlayerCount        int    `json:"player_count"`
	HumanSlots         int    `json:"human_slots"`
	ClearTriggers      bool   `json:"clear_triggers"`
	DummyStarters      bool   `json:"dummy_starters"`
	DummyUnit          int    `json:"dummy_unit,omitempty"`
	GaiaActive         bool   `json:"gaia_active"`
	TriggerCountBefore int    `json:"trigger_count_before"`
	TriggerCountAfter  int    `json:"trigger_count_after"`
	UnitCountBefore    int    `json:"unit_count_before"`
	UnitCountAfter     int    `json:"unit_count_after"`
	RebuildOK          bool   `json:"rebuild_ok"`
	InvariantOK        bool   `json:"invariant_ok"`
	Verification       string `json:"verification"`
}

func BlankScenarioSeedBytes() ([]byte, error) {
	data, err := hex.DecodeString(blankScenarioSeedHex)
	if err != nil {
		return nil, fmt.Errorf("decode embedded blank scenario seed: %w", err)
	}
	sum := sha256.Sum256(data)
	if got := hex.EncodeToString(sum[:]); got != blankScenarioSeedSHA256 {
		return nil, fmt.Errorf("embedded blank scenario seed sha256=%s, want %s", got, blankScenarioSeedSHA256)
	}
	return data, nil
}

func WriteBlankScenarioFile(output string, opts BlankOptions) (BlankReport, error) {
	opts = normalizeBlankOptions(opts)
	if opts.PlayerCount < 1 || opts.PlayerCount > 8 {
		return BlankReport{}, fmt.Errorf("blank scenario player count %d out of range 1..8", opts.PlayerCount)
	}
	if opts.HumanSlots < 0 || opts.HumanSlots > opts.PlayerCount {
		return BlankReport{}, fmt.Errorf("blank scenario human slots %d out of range 0..player_count", opts.HumanSlots)
	}
	seed, err := BlankScenarioSeedBytes()
	if err != nil {
		return BlankReport{}, err
	}
	file, err := Parse(seed)
	if err != nil {
		return BlankReport{}, fmt.Errorf("parse embedded blank scenario seed: %w", err)
	}
	beforeTriggers := 0
	if file.Triggers != nil {
		beforeTriggers = file.Triggers.Count
	}
	beforeUnits := 0
	if file.Units != nil {
		beforeUnits = file.Units.Total
	}
	recipe := blankScenarioRecipe(opts)
	if err := file.ApplyRecipe(recipe); err != nil {
		return BlankReport{}, err
	}
	if dir := filepath.Dir(output); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return BlankReport{}, err
		}
	}
	if err := file.Write(output); err != nil {
		return BlankReport{}, err
	}
	verified, err := Open(output)
	if err != nil {
		return BlankReport{}, err
	}
	if err := verified.VerifyRebuild(); err != nil {
		return BlankReport{}, err
	}
	afterTriggers := 0
	invariantOK := false
	if verified.Triggers != nil {
		afterTriggers = verified.Triggers.Count
		invariantOK = verified.Triggers.InvariantOK
	}
	afterUnits := 0
	if verified.Units != nil {
		afterUnits = verified.Units.Total
	}
	if !invariantOK {
		return BlankReport{}, fmt.Errorf("blank scenario trigger invariants failed: %s", verified.Triggers.InvariantNote)
	}
	return BlankReport{
		Output:             output,
		SeedSHA256:         blankScenarioSeedSHA256,
		Version:            verified.Version,
		PlayerCount:        opts.PlayerCount,
		HumanSlots:         opts.HumanSlots,
		ClearTriggers:      opts.ClearTriggers,
		DummyStarters:      opts.DummyStarters,
		DummyUnit:          opts.DummyUnit,
		GaiaActive:         opts.GaiaActive,
		TriggerCountBefore: beforeTriggers,
		TriggerCountAfter:  afterTriggers,
		UnitCountBefore:    beforeUnits,
		UnitCountAfter:     afterUnits,
		RebuildOK:          true,
		InvariantOK:        invariantOK,
		Verification:       scenarioWriteVerification().Label,
	}, nil
}

func normalizeBlankOptions(opts BlankOptions) BlankOptions {
	if opts.PlayerCount == 0 {
		opts.PlayerCount = 2
	}
	if opts.HumanSlots == 0 {
		opts.HumanSlots = 1
	}
	if opts.Timestamp == 0 {
		opts.Timestamp = int(time.Now().Unix())
	}
	if opts.DummyUnit == 0 {
		opts.DummyUnit = defaultBlankDummyUnit
	}
	if opts.DummyStartX == 0 {
		opts.DummyStartX = 5.5
	}
	if opts.DummyStartY == 0 {
		opts.DummyStartY = 5.5
	}
	if opts.DummySpacing == 0 {
		opts.DummySpacing = 6
	}
	opts.ClearTriggers = !opts.KeepSeedTriggers
	return opts
}

func blankScenarioRecipe(opts BlankOptions) Recipe {
	active := true
	inactive := false
	human := true
	computer := false
	timestamp := opts.Timestamp
	playerCount := opts.PlayerCount
	recipe := Recipe{
		Scenario: &ScenarioRecipe{
			PlayerCount:         &playerCount,
			TimestampOfLastSave: &timestamp,
		},
	}
	if opts.ClearTriggers {
		recipe.Triggers = append(recipe.Triggers, TriggerRecipe{Op: "clear_triggers"})
	}
	for player := 0; player <= 15; player++ {
		playerActive := &inactive
		if player == 0 && opts.GaiaActive {
			playerActive = &active
		} else if player > 0 && player <= opts.PlayerCount {
			playerActive = &active
		}
		playerHuman := &computer
		if player > 0 && (player <= opts.HumanSlots || (!*playerActive && opts.InactiveRestHuman)) {
			playerHuman = &human
		}
		recipe.Players = append(recipe.Players, PlayerRecipe{
			Player: player,
			Active: playerActive,
			Human:  playerHuman,
		})
	}
	if opts.DummyStarters {
		status := 2
		rotation := 0.0
		z := 0.0
		for player := 1; player <= opts.PlayerCount; player++ {
			x := opts.DummyStartX + float64(player-1)*opts.DummySpacing
			y := opts.DummyStartY
			caption := fmt.Sprintf("A2K_BLANK_DUMMY_P%d", player)
			recipe.Units = append(recipe.Units, UnitRecipe{
				Op:            "add_unit",
				Player:        player,
				UnitConst:     opts.DummyUnit,
				X:             &x,
				Y:             &y,
				Z:             &z,
				Status:        &status,
				Rotation:      &rotation,
				CaptionString: caption,
			})
		}
	}
	return recipe
}

const blankScenarioSeedHex = "" +
	"312e3538000000000600000051f85e6a010000000002000000e8030000010000001100000002000000030000000400000005000000060000000700000008000000090000000a0000000c0000000d0000000e0000000f000000100000001100000012000000130000000600000063687261650003000000ecdbcb6e1b551cc0e1" +
	"71d3264d4a010989cb6e508560d15a2aeb229a4b059128ad48056297a9336e863ae3321e0bb24442acd982d447e01d102bc443f00c3c0166ced811cec553da58aa52be9f74f4d53e93f3b7c7092b1c557dfdc11f1fbe5a79313c9024499224492f7c7f8f46a3d3ac56754658db2b2bd167ab9f6edcb97d6d7df3f3c38fce4fae" +
	"f92fd7855ce7baa7b9ee79165e63dd1bd15a2fc91fc6f7d241d94efae9fb834e9a2745d69ff59773d2690b279dac19b58ef478c65d7de2390d7be71af6a63fada335fd6e5e68d85b6cd85b6ad8bbd8b0b7dcb0b7d2b077a961efa586bdcb0d7b2f37ecbd52adf0193ee937fffbeac6bf77657cff2dcbfa7f2e49922449922449" +
	"9224499224499224493a2b3debf75d0ed56afcaaca992b7c5fe7342bdcd3831e9f707ef8ffceb7ab2fcd6caf2c451b693719f6caf0308aa67eeeb914fd5ebd88d3acd0e5bf7ef9b3fec6d3b793357963000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000" +
	"00000000000000000000000000000000000000000000000000000000000000000000000000301a010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000080d108000000000000000000000000c09961b972761fafbf087be13dbe196d75d23c29b27e7c6b272bfb" +
	"457c7737c9cbfe5ef4eb8fdfadfef0dbf2da8de846142d470bd1c2b8f0b3ad6a9d6f58a16e77f1e6b1a1e1d6ce1e19458fd6c23a18793033fc6438b6696cd89bc7c8c9cc83914d63cfcd6be478e6f4c85963c345f31959cf3c3a72d6d8398d0c334f1a7974ec85398eac66ce1a393d76719e23eb898d7f2061ec52d3c8634f9e" +
	"d0cf3f855ebb59bfc370e28cdeaad666f9ee20fe6a9875f6e3224d7abdfdf1837674a9dafc2229d3622fedf5f3fa45dd4b0665b5335dab153e9356d86d855b359ed67ae7e0e424ee16c3ac8ccbdda48c77d241a7c8eea7e1e94ebfd72fdaf5b1778a247f90d6fffc72b8b73735a1e1b55ff977c2fd61ded96dc79b6575e85e75" +
	"78968f9f4a07edfae3fba8481ea5f5fbb99d3c4ce36fb23c6dd74787dbb355ad705f6715ae6bdad7e156abfb75abdb4d3be1b3f9241b9457e3ad34df89d7c32f40b7fa58caac9f5f8d07fd61d149e3bbbd643f2de2ebd5136591e50fe2cd8df8daf5e39f7b78e2f56abd3d59a77985e1bce0cae4ec6739e3d00b0c0fc27ff225" +
	"e969fb6700"
