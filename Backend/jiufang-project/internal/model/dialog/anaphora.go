// Package dialog implements the dialog management model for multi-turn conversations.
// This file implements anaphora resolution (指代消解) for multi-turn dialog.
package dialog

import (
	"fmt"
	"regexp"
	"strings"

	"go.uber.org/zap"
)

// AnaphoraResolver resolves anaphora (指代词) in user input.
// It identifies pronouns and references and resolves them to actual entities from context.
type AnaphoraResolver struct {
	logger *zap.Logger

	// Common anaphora patterns in Chinese
	anaphoraPatterns map[string]*regexp.Regexp
}

// NewAnaphoraResolver creates a new anaphora resolver.
func NewAnaphoraResolver(logger *zap.Logger) *AnaphoraResolver {
	if logger == nil {
		logger = zap.NewNop()
	}

	// Define common anaphora patterns
	patterns := map[string]*regexp.Regexp{
		"这个":   regexp.MustCompile(`这个`),
		"那个":   regexp.MustCompile(`那个`),
		"它":    regexp.MustCompile(`它`),
		"它们":   regexp.MustCompile(`它们`),
		"这条":   regexp.MustCompile(`这条`),
		"那条":   regexp.MustCompile(`那条`),
		"这个单子": regexp.MustCompile(`这个单子`),
		"那个单子": regexp.MustCompile(`那个单子`),
		"上面的":  regexp.MustCompile(`上面的`),
		"前一个":  regexp.MustCompile(`前一个`),
		"刚才的":  regexp.MustCompile(`刚才的`),
		"之前的":  regexp.MustCompile(`之前的`),
		"刚才说的": regexp.MustCompile(`刚才说的`),
		"上次查询": regexp.MustCompile(`上次查询`),
	}

	return &AnaphoraResolver{
		logger:           logger,
		anaphoraPatterns: patterns,
	}
}

// Resolve resolves anaphora in the input text using dialog context.
func (r *AnaphoraResolver) Resolve(input string, context *DialogContext) (string, error) {
	if context == nil || len(context.QueryHistory) == 0 {
		// No context available, return original input
		return input, nil
	}

	// Get last turn for reference
	lastTurn := context.GetLastTurn()
	if lastTurn == nil {
		return input, nil
	}

	resolved := input

	// Resolve each anaphora pattern
	for anaphora, pattern := range r.anaphoraPatterns {
		if pattern.MatchString(resolved) {
			replacement := r.resolveAnaphora(anaphora, lastTurn, context)
			if replacement != "" {
				resolved = pattern.ReplaceAllString(resolved, replacement)
				r.logger.Debug("Anaphora resolved",
					zap.String("anaphora", anaphora),
					zap.String("replacement", replacement),
				)
			}
		}
	}

	return resolved, nil
}

// resolveAnaphora resolves a specific anaphora to its actual reference.
func (r *AnaphoraResolver) resolveAnaphora(anaphora string, lastTurn *QueryTurn, context *DialogContext) string {
	// Different resolution strategies based on anaphora type
	switch anaphora {
	case "这个", "那个", "它", "它们":
		// Resolve to last query subject
		return r.resolveToLastSubject(lastTurn)

	case "这条", "那条", "这个单子", "那个单子":
		// Resolve to last document type
		return r.resolveToLastDocument(lastTurn)

	case "上面的", "前一个", "刚才的", "之前的", "刚才说的":
		// Resolve to last query condition
		return r.resolveToLastCondition(lastTurn)

	case "上次查询":
		// Resolve to last query input
		return r.resolveToLastQuery(lastTurn)

	default:
		return ""
	}
}

// resolveToLastSubject resolves to the subject of the last query.
func (r *AnaphoraResolver) resolveToLastSubject(lastTurn *QueryTurn) string {
	if lastTurn == nil {
		return ""
	}

	// Try to extract subject from understanding
	if lastTurn.Understanding != "" {
		// Simple extraction: return the understanding summary
		return fmt.Sprintf("上次查询的%s", lastTurn.Understanding)
	}

	// Fallback: return last input
	return fmt.Sprintf("上次查询'%s'", lastTurn.Input)
}

// resolveToLastDocument resolves to the document type from last query.
func (r *AnaphoraResolver) resolveToLastDocument(lastTurn *QueryTurn) string {
	if lastTurn == nil {
		return ""
	}

	// Try to get document type from entities
	if lastTurn.Entities != nil {
		if docType, ok := lastTurn.Entities["document_type"]; ok {
			return fmt.Sprintf("%s", docType)
		}
	}

	// Fallback: return "单据"
	return "单据"
}

// resolveToLastCondition resolves to the condition from last query.
func (r *AnaphoraResolver) resolveToLastCondition(lastTurn *QueryTurn) string {
	if lastTurn == nil {
		return ""
	}

	// Build condition description from entities
	if len(lastTurn.Entities) > 0 {
		conditions := []string{}
		for key, value := range lastTurn.Entities {
			conditions = append(conditions, fmt.Sprintf("%s=%v", key, value))
		}
		return strings.Join(conditions, " AND ")
	}

	return ""
}

// resolveToLastQuery resolves to the last query input.
func (r *AnaphoraResolver) resolveToLastQuery(lastTurn *QueryTurn) string {
	if lastTurn == nil {
		return ""
	}

	return lastTurn.Input
}

// DetectAnaphora detects if the input contains anaphora.
func (r *AnaphoraResolver) DetectAnaphora(input string) bool {
	for _, pattern := range r.anaphoraPatterns {
		if pattern.MatchString(input) {
			return true
		}
	}
	return false
}

// GetAnaphoraList returns the list of anaphoras found in the input.
func (r *AnaphoraResolver) GetAnaphoraList(input string) []string {
	anaphoras := []string{}
	for anaphora, pattern := range r.anaphoraPatterns {
		if pattern.MatchString(input) {
			anaphoras = append(anaphoras, anaphora)
		}
	}
	return anaphoras
}
