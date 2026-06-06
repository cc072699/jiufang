// Package dialog implements the dialog management model for multi-turn conversations.
// This file implements condition merging for multi-turn dialog.
package dialog

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// ConditionMerger merges query conditions from multiple turns.
// It handles conflicts and ensures conditions are consistent.
type ConditionMerger struct {
	logger *zap.Logger
}

// NewConditionMerger creates a new condition merger.
func NewConditionMerger(logger *zap.Logger) *ConditionMerger {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &ConditionMerger{logger: logger}
}

// Merge merges new conditions with existing conditions.
// It handles conflicts by preferring new conditions (latest input).
func (m *ConditionMerger) Merge(existing map[string]interface{}, new map[string]interface{}) map[string]interface{} {
	if existing == nil {
		existing = make(map[string]interface{})
	}

	if new == nil {
		return existing
	}

	merged := make(map[string]interface{})

	// Copy existing conditions
	for key, value := range existing {
		merged[key] = value
	}

	// Merge new conditions (overwrite conflicts)
	for key, value := range new {
		// Check for conflicts
		if existingValue, exists := merged[key]; exists {
			// Handle conflict: prefer new value but log it
			m.logger.Debug("Condition conflict detected, using new value",
				zap.String("key", key),
				zap.Any("existing", existingValue),
				zap.Any("new", value),
			)
		}
		merged[key] = value
	}

	return merged
}

// MergeWithStrategy merges conditions with a specific strategy.
type MergeStrategy string

const (
	StrategyPreferNew    MergeStrategy = "prefer_new"   // Prefer new conditions (default)
	StrategyPreferOld    MergeStrategy = "prefer_old"   // Prefer existing conditions
	StrategyIntersection MergeStrategy = "intersection" // Only keep conditions that exist in both
	StrategyUnion        MergeStrategy = "union"        // Keep all conditions from both
)

// MergeWithStrategy merges conditions using a specific strategy.
func (m *ConditionMerger) MergeWithStrategy(existing map[string]interface{}, new map[string]interface{}, strategy MergeStrategy) map[string]interface{} {
	switch strategy {
	case StrategyPreferNew:
		return m.Merge(existing, new)

	case StrategyPreferOld:
		// Prefer existing conditions
		merged := make(map[string]interface{})
		for key, value := range existing {
			merged[key] = value
		}
		// Only add new conditions that don't exist
		for key, value := range new {
			if _, exists := merged[key]; !exists {
				merged[key] = value
			}
		}
		return merged

	case StrategyIntersection:
		// Only keep conditions that exist in both
		merged := make(map[string]interface{})
		for key, value := range existing {
			if _, exists := new[key]; exists {
				merged[key] = value
			}
		}
		return merged

	case StrategyUnion:
		// Keep all conditions from both
		return m.Merge(existing, new)

	default:
		return m.Merge(existing, new)
	}
}

// DetectConflict detects conflicts between existing and new conditions.
func (m *ConditionMerger) DetectConflict(existing map[string]interface{}, new map[string]interface{}) []string {
	conflicts := []string{}

	if existing == nil || new == nil {
		return conflicts
	}

	for key, newValue := range new {
		if existingValue, exists := existing[key]; exists {
			// Check if values are different
			if !m.valuesEqual(existingValue, newValue) {
				conflicts = append(conflicts, key)
			}
		}
	}

	return conflicts
}

// valuesEqual checks if two values are equal.
func (m *ConditionMerger) valuesEqual(a, b interface{}) bool {
	// Simple equality check
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// BuildConditionString builds a condition string from merged conditions.
func (m *ConditionMerger) BuildConditionString(conditions map[string]interface{}) string {
	if len(conditions) == 0 {
		return ""
	}

	parts := []string{}
	for key, value := range conditions {
		parts = append(parts, fmt.Sprintf("%s=%v", key, value))
	}

	return strings.Join(parts, " AND ")
}

// ExtractTimeRange extracts time range from conditions.
func (m *ConditionMerger) ExtractTimeRange(conditions map[string]interface{}) (start, end string, ok bool) {
	if conditions == nil {
		return "", "", false
	}

	// Try to extract time range
	startValue, hasStart := conditions["time_start"]
	endValue, hasEnd := conditions["time_end"]

	if hasStart && hasEnd {
		return fmt.Sprintf("%v", startValue), fmt.Sprintf("%v", endValue), true
	}

	// Try alternative keys
	startValue, hasStart = conditions["start_date"]
	endValue, hasEnd = conditions["end_date"]

	if hasStart && hasEnd {
		return fmt.Sprintf("%v", startValue), fmt.Sprintf("%v", endValue), true
	}

	return "", "", false
}

// ClearCondition clears a specific condition from the merged conditions.
func (m *ConditionMerger) ClearCondition(conditions map[string]interface{}, key string) map[string]interface{} {
	if conditions == nil {
		return make(map[string]interface{})
	}

	merged := make(map[string]interface{})
	for k, v := range conditions {
		if k != key {
			merged[k] = v
		}
	}

	return merged
}
