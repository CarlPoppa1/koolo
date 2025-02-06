package action

import (
	"fmt"

	"github.com/hectorgimenez/d2go/pkg/data/skill"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/context"
)

var SkillWeights = map[stat.ID]float64{
	stat.AllSkills:      200.0,
	stat.AddClassSkills: 175.0,
	stat.AddSkillTab:    125.0,
}

var ResistWeights = map[stat.ID]float64{
	stat.FireResist:             3.0,
	stat.ColdResist:             2.0,
	stat.LightningResist:        3.0,
	stat.PoisonResist:           1.0,
	stat.MaxFireResist:          8.0,
	stat.MaxLightningResist:     8.0,
	stat.MaxColdResist:          6.0,
	stat.MaxPoisonResist:        4.0,
	stat.AbsorbFire:             2.0,
	stat.AbsorbLightning:        2.0,
	stat.AbsorbMagic:            2.0,
	stat.AbsorbCold:             2.0,
	stat.AbsorbFirePercent:      4.0,
	stat.AbsorbLightningPercent: 4.0,
	stat.AbsorbMagicPercent:     4.0,
	stat.AbsorbColdPercent:      4.0,
	stat.DamageReduced:          2.0,
	stat.DamagePercent:          3.0,
	stat.MagicDamageReduction:   2.0,
	stat.MagicResist:            2.0,
}

var MercWeights = map[stat.ID]float64{
	stat.IncreasedAttackSpeed:   3.5,
	stat.MinDamage:              3.0,
	stat.MaxDamage:              3.0,
	stat.TwoHandedMinDamage:     3.0,
	stat.TwoHandedMaxDamage:     3.0,
	stat.AttackRating:           0.1,
	stat.CrushingBlow:           3.0,
	stat.OpenWounds:             3.0,
	stat.LifeSteal:              8.0,
	stat.ReplenishLife:          2.0,
	stat.FasterHitRecovery:      3.0,
	stat.Defense:                0.05,
	stat.Strength:               1.5,
	stat.Dexterity:              1.5,
	stat.FireResist:             2.0,
	stat.ColdResist:             1.5,
	stat.LightningResist:        2.0,
	stat.PoisonResist:           1.0,
	stat.DamageReduced:          2.0,
	stat.MagicResist:            3.0,
	stat.AbsorbFirePercent:      2.7,
	stat.AbsorbColdPercent:      2.7,
	stat.AbsorbLightningPercent: 2.7,
	stat.AbsorbMagicPercent:     2.7,
}

type MercCTCWeights struct {
	StatID stat.ID
	Weight float64
	Layer  int
}

// Example usage:
var MercCTCWeight = []MercCTCWeights{
	{StatID: stat.SkillOnAttack, Weight: 5.0, Layer: 4227},     // Amp Damage
	{StatID: stat.SkillOnAttack, Weight: 10.0, Layer: 5572},    // Decrepify
	{StatID: stat.SkillOnHit, Weight: 3.0, Layer: 4225},        // Amp Damage
	{StatID: stat.SkillOnHit, Weight: 8.0, Layer: 5572},        // Decrepify
	{StatID: stat.SkillOnGetHit, Weight: 1000.0, Layer: 17103}, // Fade
	{StatID: stat.Aura, Weight: 1000.0, Layer: 123},            // Infinity
	{StatID: stat.Aura, Weight: 100.0, Layer: 120},             // Insight
}

var GeneralWeights = map[stat.ID]float64{
	stat.CannotBeFrozen:    25.0,
	stat.FasterHitRecovery: 3.0,
	stat.FasterRunWalk:     2.0,
	stat.FasterBlockRate:   2.0,
	stat.FasterCastRate:    4.0,
	stat.ChanceToBlock:     2.5,
	stat.MagicFind:         1.0,
	stat.GoldFind:          0.1,
	stat.Defense:           0.05,
	stat.ManaRecovery:      2.0,
	stat.Strength:          1.0,
	stat.Dexterity:         1.0,
	stat.Vitality:          1.5,
	stat.Energy:            0.5,
	stat.MaxLife:           0.5,
	stat.MaxMana:           0.25,
	stat.ReplenishQuantity: 2.0,
	stat.ReplenishLife:     2.0,
	stat.LifePerLevel:      3.0,
	stat.ManaPerLevel:      2.0,
}

var beltSizes = map[string]int{
	"lbl": 2,
	"vbl": 2,
	"mbl": 3,
	"tbl": 3,
}

// TierScore calculates overall item tier score
func TierScore(item data.Item) float64 {
	if item.IsFromQuest() {
		return -1
	}

	score := 1.0

	score += calculateGeneralScore(item)
	score += calculateResistScore(item)
	//score += calculateBuildScore(item)
	score += calculateSkillScore(item)

	// Apply final tier adjustments
	//if score > 1 && score < MaxTierScore {
	//	if IsWantedItem(item) {
	//		score += MaxTierScore
	//	}
	//}

	return score
}

func calculateGeneralScore(newitem data.Item) float64 {
	ctx := context.Get()

	score := 0.0

	// Handle Cannot Be Frozen
	//if !ctx.Data.CanTeleport() && item.HasStat(stat.CannotbeFrozen) {
	//	if !char.HasStat(stat.CannotbeFrozen) {
	//		score += GeneralWeights[stat.CannotbeFrozen]
	//	}
	//}

	// Handle belt slots
	if newitem.Desc().Type == "belt" {
		beltSize := beltSizes[newitem.Desc().Code]
		if beltSize == 0 {
			beltSize = 4
		}

		currentBeltSize := 0
		for _, eqItem := range ctx.Data.Inventory.ByLocation(item.LocationEquipped) {
			if eqItem.Desc().Type == "belt" {
				currentBeltSize = beltSizes[eqItem.Desc().Code]
			}
		}

		if currentBeltSize > beltSize {
			score -= 50
		} else {
			score += float64(beltSize * 4 * 2)
		}
	}

	// Handle sockets
	//if !item.IsRuneword {
	//	score += float64(item.Desc().MaxSockets * 10)
	//}

	// Per level calculations
	charLevel, _ := ctx.Data.PlayerUnit.FindStat(stat.Level, 0)
	lifeperlvl, _ := newitem.FindStat(stat.LifePerLevel, 0)
	manaperlvl, _ := newitem.FindStat(stat.ManaPerLevel, 0)
	score += (float64(lifeperlvl.Value/2048) * float64(charLevel.Value)) * GeneralWeights[stat.LifePerLevel]
	score += (float64(manaperlvl.Value/2048) * float64(charLevel.Value)) * GeneralWeights[stat.ManaPerLevel]

	// Process other stats
	otherStats := []stat.ID{
		stat.FasterHitRecovery, stat.FasterRunWalk, stat.FasterBlockRate, stat.FasterCastRate, stat.ChanceToBlock,
		stat.MagicFind, stat.GoldFind, stat.Defense, stat.ManaRecovery,
		stat.Strength, stat.Dexterity, stat.Vitality, stat.Energy,
		stat.MaxLife, stat.MaxMana, stat.ReplenishQuantity, stat.ReplenishLife,
	}

	for _, statID := range otherStats {
		statData, _ := newitem.FindStat(statID, 0)
		score += float64(statData.Value) * GeneralWeights[statID]
	}

	return score
}

func calculateResistScore(newitem data.Item) float64 {
	ctx := context.Get()
	score := 0.0

	// Get new item resists
	newFR, _ := newitem.FindStat(stat.FireResist, 0)
	newCR, _ := newitem.FindStat(stat.ColdResist, 0)
	newLR, _ := newitem.FindStat(stat.LightningResist, 0)
	newPR, _ := newitem.FindStat(stat.PoisonResist, 0)

	if newFR.Value > 0 || newCR.Value > 0 || newLR.Value > 0 || newPR.Value > 0 {
		// Calculate max resists based on difficulty
		maxRes := 75

		// Get equipped item resists
		var oldFR, oldCR, oldLR, oldPR int
		for _, eqItem := range ctx.Data.Inventory.ByLocation(item.LocationEquipped) {
			if eqItem.Desc().Type == newitem.Desc().Type {
				oldFRStat, _ := eqItem.FindStat(stat.FireResist, 0)
				oldFR = oldFRStat.Value
				oldCRStat, _ := eqItem.FindStat(stat.ColdResist, 0)
				oldCR = oldCRStat.Value
				oldLRStat, _ := eqItem.FindStat(stat.LightningResist, 0)
				oldLR = oldLRStat.Value
				oldPRStat, _ := eqItem.FindStat(stat.PoisonResist, 0)
				oldPR = oldPRStat.Value
			}

			// Calculate base resists without equipped item
			baseFRStat, _ := ctx.Data.PlayerUnit.BaseStats.FindStat(stat.FireResist, 0)
			baseFR := baseFRStat.Value - oldFR
			baseCRStat, _ := ctx.Data.PlayerUnit.BaseStats.FindStat(stat.ColdResist, 0)
			baseCR := baseCRStat.Value - oldCR
			baseLRStat, _ := ctx.Data.PlayerUnit.BaseStats.FindStat(stat.LightningResist, 0)
			baseLR := baseLRStat.Value - oldLR
			basePRStat, _ := ctx.Data.PlayerUnit.BaseStats.FindStat(stat.PoisonResist, 0)
			basePR := basePRStat.Value - oldPR

			// Calculate effective resists up to max
			frLimit := max(maxRes-baseFR, 0)
			crLimit := max(maxRes-baseCR, 0)
			lrLimit := max(maxRes-baseLR, 0)
			prLimit := max(maxRes-basePR, 0)

			effectiveFR := min(newFR.Value, frLimit)
			effectiveCR := min(newCR.Value, crLimit)
			effectiveLR := min(newLR.Value, lrLimit)
			effectivePR := min(newPR.Value, prLimit)

			// Calculate resist score
			score += float64(effectiveFR) * ResistWeights[stat.FireResist]
			score += float64(effectiveCR) * ResistWeights[stat.ColdResist]
			score += float64(effectiveLR) * ResistWeights[stat.LightningResist]
			score += float64(effectivePR) * ResistWeights[stat.PoisonResist]
		}
	}

	// Add special resist stats
	otherStats := []stat.ID{
		stat.MaxFireResist, stat.MaxLightningResist,
		stat.MaxColdResist, stat.MaxPoisonResist,
		stat.AbsorbFire, stat.AbsorbLightning,
		stat.AbsorbMagic, stat.AbsorbCold,
		stat.AbsorbFirePercent, stat.AbsorbLightningPercent,
		stat.AbsorbMagicPercent, stat.AbsorbColdPercent,
		stat.DamageReduced, stat.DamagePercent,
		stat.MagicDamageReduction, stat.MagicResist,
	}

	for _, statID := range otherStats {
		statData, _ := newitem.FindStat(statID, 0)
		score += float64(statData.Value) * GeneralWeights[statID]
	}

	return score
}

func calculateSkillScore(item data.Item) float64 {
	ctx := context.Get()
	score := 0.0

	if statData, found := item.FindStat(stat.AllSkills, 0); found {
		score += float64(statData.Value) * SkillWeights[statData.ID]
	}
	if classSkillsStat, found := item.FindStat(stat.AddClassSkills, int(ctx.Data.PlayerUnit.Class)); found {
		score += float64(classSkillsStat.Value) * SkillWeights[classSkillsStat.ID]
	}

	tabskill := int(ctx.Data.PlayerUnit.Class)*8 + (getMaxSkillTabPage() - 1)
	if tabSkillsStat, found := item.FindStat(stat.AddSkillTab, tabskill); found {
		score += float64(tabSkillsStat.Value) * SkillWeights[tabSkillsStat.ID]
	}

	usedSkills := make([]skill.ID, 0)

	//Let's ignore 1 point wonders unless we're above level 2
	for sk, pts := range ctx.Data.PlayerUnit.Skills {
		if pts.Level > 1 {
			usedSkills = append(usedSkills, sk)
		} else if lvl, _ := ctx.Data.PlayerUnit.FindStat(stat.Level, 0); lvl.Value < 3 {
			usedSkills = append(usedSkills, sk)
		}
	}
	for _, usedSkill := range usedSkills {
		if _, found := item.FindStat(stat.SingleSkill, int(usedSkill)); found {
			score += 40
		}
	}

	return score
}

// sumElementalDamage calculates total elemental damage for an item
func sumElementalDamage(item data.Item) float64 {
	score := 0.0

	fireMin, _ := item.FindStat(stat.FireMinDamage, 0)
	fireMax, _ := item.FindStat(stat.FireMaxDamage, 0)
	score += float64(fireMin.Value + fireMax.Value)

	lightMin, _ := item.FindStat(stat.LightningMinDamage, 0)
	lightMax, _ := item.FindStat(stat.LightningMaxDamage, 0)
	score += float64(lightMin.Value + lightMax.Value)

	coldMin, _ := item.FindStat(stat.ColdMinDamage, 0)
	coldMax, _ := item.FindStat(stat.ColdMaxDamage, 0)
	score += float64(coldMin.Value + coldMax.Value)

	magicMin, _ := item.FindStat(stat.MagicMinDamage, 0)
	magicMax, _ := item.FindStat(stat.MagicMaxDamage, 0)
	score += float64(magicMin.Value + magicMax.Value)

	// PSN damage adjusted for damage per frame (125/256)
	poisonMin, _ := item.FindStat(stat.PoisonMinDamage, 0)
	score += float64(poisonMin.Value) * 125.0 / 256.0

	return score
}

func CalculateMercScore(item data.Item) float64 {
	ctx := context.Get()
	score := 1.0
	score += sumElementalDamage(item) * 2.0
	ctx.Logger.Debug(fmt.Sprintf("Elemental Damage: %v", score))
	for statID, weight := range MercWeights {
		if statData, found := item.FindStat(statID, 0); found {
			ctx.Logger.Debug(fmt.Sprintf("Stat: %v", statData))
			score += float64(statData.Value) * weight
		}
	}
	for _, CTCstat := range MercCTCWeight {
		if statData, found := item.FindStat(CTCstat.StatID, CTCstat.Layer); found {
			score += CTCstat.Weight
			ctx.Logger.Debug(fmt.Sprintf("Stat: %v", statData))
		}
	}

	return score
}

func getMaxSkillTabPage() int {
	ctx := context.Get()

	tabCounts := make(map[int]int)
	maxCount := 0
	maxPage := 0
	for pskill, pts := range ctx.Data.PlayerUnit.Skills {
		if page := pskill.Desc().Page; page > 0 {
			tabCounts[page] += int(pts.Level)
			if tabCounts[page] > maxCount {
				maxCount = tabCounts[page]
				maxPage = page
			}
		}
	}

	return maxPage
}
