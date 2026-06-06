// Package agent implements the AI Agent for semantic understanding and SQL generation.
package agent

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"

	"jiufang/internal/infrastructure/erp"
	"jiufang/internal/infrastructure/llm"
	"jiufang/internal/model/agent"
	agentModel "jiufang/internal/model/agent"
)

// QueryAgent is the main agent for natural language query processing.
// It implements the ReAct pattern: Reason -> Act -> Observe.
type QueryAgent struct {
	intentParser    *IntentParser
	entityExtractor *EntityExtractor
	sqlGenerator    *SQLGenerator
	sqlValidator    *SQLValidator
	sqlExecutor     *SQLExecutor
	resultFormatter *ResultFormatter
	llmClient       llm.LLMClientInterface
	logger          *zap.Logger
}

// NewQueryAgent creates a new query agent.
func NewQueryAgent(llmClient llm.LLMClientInterface, erpReader erp.ERPReaderInterface, logger *zap.Logger) *QueryAgent {
	if logger == nil {
		logger = zap.NewNop()
	}

	validator := NewSQLValidator()

	return &QueryAgent{
		intentParser:    NewIntentParser(llmClient),
		entityExtractor: NewEntityExtractor(llmClient),
		sqlGenerator:    NewSQLGenerator(llmClient),
		sqlValidator:    validator,
		sqlExecutor:     NewSQLExecutor(erpReader, validator),
		resultFormatter: NewResultFormatter(),
		llmClient:       llmClient,
		logger:          logger,
	}
}

// Execute executes a natural language query.
func (a *QueryAgent) Execute(ctx context.Context, input string, sessionID string, userID uint) (*agent.QueryResult, error) {
	// Step 1: Parse intent
	intent, err := a.intentParser.Parse(ctx, input)
	if err != nil {
		a.logger.Error("Failed to parse intent", zap.Error(err), zap.String("input", input))
		return nil, err
	}

	a.logger.Info("Intent parsed",
		zap.String("intent_type", string(intent.Type)),
		zap.Float64("confidence", intent.Confidence),
		zap.String("input", input),
	)

	// Step 2: Check if clarification is needed
	if intent.NeedsClarification() {
		a.logger.Warn("Intent needs clarification", zap.String("intent_type", string(intent.Type)))
		return nil, a.createClarificationError(intent, input, sessionID)
	}

	// Step 3: Extract entities
	entities, err := a.entityExtractor.Extract(ctx, input)
	if err != nil {
		a.logger.Error("Failed to extract entities", zap.Error(err))
		return nil, err
	}

	a.logger.Info("Entities extracted", zap.Int("entity_count", len(entities)))

	// Step 4: Generate SQL
	sql, err := a.sqlGenerator.Generate(ctx, intent, entities)
	if err != nil {
		a.logger.Error("Failed to generate SQL", zap.Error(err))
		return nil, err
	}

	a.logger.Info("SQL generated", zap.String("sql", sql))

	// Step 5: Validate SQL
	if err := a.sqlValidator.Validate(sql); err != nil {
		a.logger.Error("SQL validation failed", zap.Error(err), zap.String("sql", sql))
		return nil, err
	}

	// Step 6: Execute SQL
	result, err := a.sqlExecutor.Execute(ctx, sql)
	if err != nil {
		a.logger.Error("SQL execution failed", zap.Error(err))
		return nil, err
	}

	// Step 7: Format result
	result.GeneratedSQL = sql
	result.Understanding = a.generateUnderstanding(intent, entities)
	result.VisualizationType = a.resultFormatter.DetermineVisualizationType(result, intent)

	a.logger.Info("Query completed",
		zap.Int("total_rows", result.TotalRows),
		zap.Int64("execution_time", result.ExecutionTime),
	)

	return result, nil
}

// ExecuteWithContext executes a query with full context.
func (a *QueryAgent) ExecuteWithContext(ctx context.Context, req *agent.QueryRequest) (*agent.QueryResponse, error) {
	// Handle clarification response
	if req.ClarificationResponse != nil {
		return a.handleClarificationResponse(ctx, req)
	}

	// Execute query
	result, err := a.Execute(ctx, req.Input, req.SessionID, req.UserID)
	if err != nil {
		// Check if error is clarification request
		if clarificationErr, ok := err.(*ClarificationError); ok {
			return &agent.QueryResponse{
				Clarification:      clarificationErr.Clarification,
				NeedsClarification: true,
			}, nil
		}
		return &agent.QueryResponse{
			Error: err.Error(),
		}, nil
	}

	// Generate suggested questions
	intent, _ := a.intentParser.Parse(ctx, req.Input)
	suggestedQuestions := a.resultFormatter.GenerateSuggestedQuestions(result, intent)

	return &agent.QueryResponse{
		Result:             result,
		NeedsClarification: false,
		SuggestedQuestions: suggestedQuestions,
	}, nil
}

// ExecuteWithPermission executes a query with permission filtering.
func (a *QueryAgent) ExecuteWithPermission(ctx context.Context, input string, queryContext *agent.QueryContext) (*agent.QueryResult, error) {
	// Parse intent
	intent, err := a.intentParser.Parse(ctx, input)
	if err != nil {
		return nil, err
	}

	if intent.NeedsClarification() {
		return nil, a.createClarificationError(intent, input, queryContext.SessionID)
	}

	// Extract entities
	entities, err := a.entityExtractor.Extract(ctx, input)
	if err != nil {
		return nil, err
	}

	// Generate SQL with permission filter
	sql, err := a.sqlGenerator.GenerateWithPermissionFilter(ctx, intent, entities, queryContext.PermissionFilter)
	if err != nil {
		return nil, err
	}

	// Execute with permission check
	result, err := a.sqlExecutor.ExecuteWithPermission(ctx, sql, queryContext)
	if err != nil {
		return nil, err
	}

	// Format result
	result.GeneratedSQL = sql
	result.Understanding = a.generateUnderstanding(intent, entities)
	result.VisualizationType = a.resultFormatter.DetermineVisualizationType(result, intent)

	return result, nil
}

// Stream executes a query with streaming response.
func (a *QueryAgent) Stream(ctx context.Context, input string, sessionID string, userID uint) (*schema.StreamReader[*agent.QueryResult], error) {
	// For streaming, we use a channel to send progress updates
	msgChan := make(chan *agent.QueryResult, 10)

	go func() {
		defer close(msgChan)

		// Execute query
		result, err := a.Execute(ctx, input, sessionID, userID)
		if err != nil {
			msgChan <- &agent.QueryResult{
				Understanding: fmt.Sprintf("查询失败：%v", err),
				IsEmpty:       true,
			}
			return
		}

		// Send result
		msgChan <- result
	}()

	// Convert channel to StreamReader
	// Note: StreamReader from channel requires custom implementation
	// For now, we collect all results and return as array-based stream
	results := make([]*agentModel.QueryResult, 0)
	for result := range msgChan {
		results = append(results, result)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no results generated")
	}

	reader := schema.StreamReaderFromArray(results)
	return reader, nil
}

// handleClarificationResponse handles user's response to clarification.
func (a *QueryAgent) handleClarificationResponse(ctx context.Context, req *agent.QueryRequest) (*agent.QueryResponse, error) {
	// Re-parse with clarification input
	clarifiedInput := req.Input
	if req.ClarificationResponse.SelectedOption != "" {
		clarifiedInput = req.ClarificationResponse.SelectedOption + " " + req.Input
	}
	if req.ClarificationResponse.AdditionalInput != "" {
		clarifiedInput = clarifiedInput + " " + req.ClarificationResponse.AdditionalInput
	}

	// Execute with clarified input
	result, err := a.Execute(ctx, clarifiedInput, req.SessionID, req.UserID)
	if err != nil {
		return &agent.QueryResponse{
			Error: err.Error(),
		}, nil
	}

	return &agent.QueryResponse{
		Result:             result,
		NeedsClarification: false,
	}, nil
}

// createClarificationError creates a clarification error.
func (a *QueryAgent) createClarificationError(intent *agent.Intent, input string, sessionID string) *ClarificationError {
	options := []agent.ClarificationOption{
		{ID: "statistics", Label: "统计查询", Value: "statistics", Description: "查询统计数据（总数、总金额等）"},
		{ID: "detail", Label: "明细查询", Value: "detail", Description: "查询详细记录"},
		{ID: "trend", Label: "趋势查询", Value: "trend", Description: "查询趋势变化"},
		{ID: "comparison", Label: "对比查询", Value: "comparison", Description: "对比不同数据"},
	}

	clarification := agent.NewIntentClarification(input, options)

	return &ClarificationError{
		Clarification: clarification,
	}
}

// generateUnderstanding generates understanding summary.
func (a *QueryAgent) generateUnderstanding(intent *agent.Intent, entities []agent.Entity) string {
	understanding := fmt.Sprintf("您想查询%s。", intent.Description)

	if len(entities) > 0 {
		understanding += "查询条件："
		for _, e := range entities {
			understanding += fmt.Sprintf(" %s=%s", e.Type, e.RawText)
		}
	}

	return understanding
}

// ClarificationError represents an error that requires clarification.
type ClarificationError struct {
	Clarification *agent.ClarificationRequest
}

// Error implements the error interface.
func (e *ClarificationError) Error() string {
	return "需要用户澄清：" + e.Clarification.Question
}

// AgentInterface defines the interface for query agents.
type AgentInterface interface {
	Execute(ctx context.Context, input string, sessionID string, userID uint) (*agent.QueryResult, error)
	ExecuteWithContext(ctx context.Context, req *agent.QueryRequest) (*agent.QueryResponse, error)
	ExecuteWithPermission(ctx context.Context, input string, queryContext *agent.QueryContext) (*agent.QueryResult, error)
	Stream(ctx context.Context, input string, sessionID string, userID uint) (*schema.StreamReader[*agent.QueryResult], error)
}

// Ensure QueryAgent implements AgentInterface
var _ AgentInterface = (*QueryAgent)(nil)
