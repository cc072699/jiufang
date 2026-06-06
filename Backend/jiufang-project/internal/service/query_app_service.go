// Package service implements application services for the system.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"

	agentpkg "jiufang/internal/agent"
	"jiufang/internal/infrastructure/erp"
	"jiufang/internal/infrastructure/llm"
	"jiufang/internal/model/agent"
	querymodel "jiufang/internal/model/query"
	"jiufang/internal/pkg/id"
	"jiufang/internal/repository"
)

// QueryAppService handles natural language query operations.
type QueryAppService struct {
	agent                agentpkg.AgentInterface
	conversationManager  *agentpkg.ConversationManager
	logger               *zap.Logger
	queryRepo            repository.QueryRepositoryInterface
	idGenerator          id.SnowflakeGeneratorInterface
	visualizationService VisualizationServiceInterface
}

// NewQueryAppService creates a new query application service.
func NewQueryAppService(llmClient llm.LLMClientInterface, erpReader erp.ERPReaderInterface, logger *zap.Logger) *QueryAppService {
	if logger == nil {
		logger = zap.NewNop()
	}

	queryAgent := agentpkg.NewQueryAgent(llmClient, erpReader, logger)
	conversationManager := agentpkg.NewConversationManager(10)

	return &QueryAppService{
		agent:               queryAgent,
		conversationManager: conversationManager,
		logger:              logger,
	}
}

// NewQueryAppServiceWithHistory creates a new query application service with history support.
func NewQueryAppServiceWithHistory(llmClient llm.LLMClientInterface, erpReader erp.ERPReaderInterface, queryRepo repository.QueryRepositoryInterface, idGenerator id.SnowflakeGeneratorInterface, logger *zap.Logger) *QueryAppService {
	if logger == nil {
		logger = zap.NewNop()
	}

	queryAgent := agentpkg.NewQueryAgent(llmClient, erpReader, logger)
	conversationManager := agentpkg.NewConversationManager(10)
	visualizationService := NewVisualizationService(logger)

	return &QueryAppService{
		agent:                queryAgent,
		conversationManager:  conversationManager,
		logger:               logger,
		queryRepo:            queryRepo,
		idGenerator:          idGenerator,
		visualizationService: visualizationService,
	}
}

// ExecuteQuery executes a natural language query.
func (s *QueryAppService) ExecuteQuery(ctx context.Context, input string, sessionID string, userID uint) (*agent.QueryResult, error) {
	var queryID int64

	if s.idGenerator != nil {
		queryID = s.idGenerator.Generate()
	}

	result, err := s.agent.Execute(ctx, input, sessionID, userID)
	if err != nil {
		s.logger.Error("Query execution failed",
			zap.String("input", input),
			zap.String("session_id", sessionID),
			zap.Uint("user_id", userID),
			zap.Error(err),
		)

		s.conversationManager.UpdateConversation(sessionID, &agentpkg.QueryHistoryItem{
			Timestamp: time.Now(),
			Input:     input,
			Intent:    agent.IntentTypeStatistics,
			Result:    nil,
			Success:   false,
		})

		return nil, err
	}

	result.QueryID = queryID

	s.conversationManager.UpdateConversation(sessionID, &agentpkg.QueryHistoryItem{
		Timestamp: result.Timestamp,
		Input:     input,
		Intent:    agent.IntentTypeStatistics,
		Result:    result,
		Success:   true,
	})

	if s.queryRepo != nil {
		var sessionIDInt int64
		if _, err := fmt.Sscanf(sessionID, "%d", &sessionIDInt); err != nil {
			s.logger.Warn("Failed to parse sessionID", zap.Error(err))
			sessionIDInt = 0
		}

		// Build column definitions from data
		columns := buildColumnDefs(result.Data)

		// Store full result structure for frontend history display
		fullResult := map[string]interface{}{
			"session_id":   sessionID,
			"understanding": result.Understanding,
			"result_type":  "table",
			"sql":          result.GeneratedSQL,
			"columns":      columns,
			"rows":         result.Data,
			"can_export":   true,
		}
		if result.VisualizationType != "" && result.VisualizationType != "table" {
			fullResult["result_type"] = "chart"
		}
		if len(result.Data) == 0 {
			fullResult["result_type"] = "empty"
		}

		resultDataJSON, err := json.Marshal(fullResult)
		if err != nil {
			s.logger.Warn("Failed to marshal result data", zap.Error(err))
			resultDataJSON = []byte("{}")
		}

		queryRecord := &querymodel.QueryRecord{
			SnowflakeID: queryID,
			UserID:      int64(userID),
			SessionID:   sessionIDInt,
			Input:       input,
			SQL:         result.GeneratedSQL,
			Status:      querymodel.QueryStatusSuccess,
			ResultData:  string(resultDataJSON),
			ResultCount: len(result.Data),
			CreatedAt:   result.Timestamp,
		}

		if err := s.queryRepo.CreateQueryRecord(ctx, queryRecord); err != nil {
			s.logger.Warn("Failed to save query record", zap.Error(err))
		}
	}

	if result.VisualizationType != agent.ChartTypeTable && s.visualizationService != nil {
		var columns []string
		if len(result.Data) > 0 {
			for key := range result.Data[0] {
				columns = append(columns, key)
			}
		}

		chartConfig, err := s.visualizationService.GenerateChartConfig(result.Data, result.VisualizationType, columns, "")
		if err != nil {
			s.logger.Warn("Failed to generate chart config", zap.Error(err))
		} else {
			result.ChartConfig = chartConfig
		}
	}

	return result, nil
}

// ExecuteQueryWithContext executes a query with full context support.
func (s *QueryAppService) ExecuteQueryWithContext(ctx context.Context, req *agent.QueryRequest) (*agent.QueryResponse, error) {
	state := s.conversationManager.GetConversation(req.SessionID)
	queryContext := &agent.QueryContext{
		UserID:    req.UserID,
		SessionID: req.SessionID,
	}
	if state.LastIntent != nil {
		queryContext.UserRole = string(state.LastIntent.Type)
	}

	if req.IsFollowUp && len(state.CurrentEntities) > 0 {
		queryContext.PreviousContext = &agent.EntityCollection{}
	}

	response, err := s.agent.ExecuteWithContext(ctx, req)
	if err != nil {
		return nil, err
	}

	if response.Result != nil {
		s.conversationManager.UpdateConversation(req.SessionID, &agentpkg.QueryHistoryItem{
			Timestamp: response.Result.Timestamp,
			Input:     req.Input,
			Result:    response.Result,
			Success:   true,
		})
	}

	return response, nil
}

// ExecuteQueryWithPermission executes a query with permission filtering.
func (s *QueryAppService) ExecuteQueryWithPermission(ctx context.Context, input string, queryContext *agent.QueryContext) (*agent.QueryResult, error) {
	var queryID int64
	if s.idGenerator != nil {
		queryID = s.idGenerator.Generate()
	}

	result, err := s.agent.ExecuteWithPermission(ctx, input, queryContext)
	if err != nil {
		return nil, err
	}

	result.QueryID = queryID

	s.conversationManager.UpdateConversation(queryContext.SessionID, &agentpkg.QueryHistoryItem{
		Timestamp: result.Timestamp,
		Input:     input,
		Result:    result,
		Success:   true,
	})

	// Save history record
	if s.queryRepo != nil {
		var sessionIDInt int64
		if _, err := fmt.Sscanf(queryContext.SessionID, "%d", &sessionIDInt); err != nil {
			s.logger.Warn("Failed to parse sessionID", zap.Error(err))
			sessionIDInt = 0
		}

		columns := buildColumnDefs(result.Data)
		fullResult := map[string]interface{}{
			"session_id":   queryContext.SessionID,
			"understanding": result.Understanding,
			"result_type":  "table",
			"sql":          result.GeneratedSQL,
			"columns":      columns,
			"rows":         result.Data,
			"can_export":   true,
		}
		if len(result.Data) == 0 {
			fullResult["result_type"] = "empty"
		}

		resultDataJSON, err := json.Marshal(fullResult)
		if err != nil {
			resultDataJSON = []byte("{}")
		}

		queryRecord := &querymodel.QueryRecord{
			SnowflakeID: queryID,
			UserID:      int64(queryContext.UserID),
			SessionID:   sessionIDInt,
			Input:       input,
			SQL:         result.GeneratedSQL,
			Status:      querymodel.QueryStatusSuccess,
			ResultData:  string(resultDataJSON),
			ResultCount: len(result.Data),
			CreatedAt:   result.Timestamp,
		}

		if err := s.queryRepo.CreateQueryRecord(ctx, queryRecord); err != nil {
			s.logger.Warn("Failed to save query record", zap.Error(err))
		}
	}

	return result, nil
}

// StreamQuery executes a query with streaming response.
func (s *QueryAppService) StreamQuery(ctx context.Context, input string, sessionID string, userID uint) (*schema.StreamReader[*agent.QueryResult], error) {
	stream, err := s.agent.Stream(ctx, input, sessionID, userID)
	if err != nil {
		return nil, err
	}
	return stream, nil
}

// GetConversationHistory retrieves conversation history for a session.
func (s *QueryAppService) GetConversationHistory(sessionID string) []agentpkg.QueryHistoryItem {
	state := s.conversationManager.GetConversation(sessionID)
	if state == nil {
		return []agentpkg.QueryHistoryItem{}
	}
	return state.QueryHistory
}

// GetSuggestedQuestions generates suggested follow-up questions.
func (s *QueryAppService) GetSuggestedQuestions(sessionID string) []string {
	state := s.conversationManager.GetConversation(sessionID)
	if state.LastResult == nil || state.LastIntent == nil {
		return []string{}
	}
	formatter := agentpkg.NewResultFormatter()
	return formatter.GenerateSuggestedQuestions(state.LastResult, state.LastIntent)
}

// QueryAppServiceInterface defines the interface for query application service.
type QueryAppServiceInterface interface {
	ExecuteQuery(ctx context.Context, input string, sessionID string, userID uint) (*agent.QueryResult, error)
	ExecuteQueryWithContext(ctx context.Context, req *agent.QueryRequest) (*agent.QueryResponse, error)
	ExecuteQueryWithPermission(ctx context.Context, input string, queryContext *agent.QueryContext) (*agent.QueryResult, error)
	StreamQuery(ctx context.Context, input string, sessionID string, userID uint) (*schema.StreamReader[*agent.QueryResult], error)
	GetConversationHistory(sessionID string) []agentpkg.QueryHistoryItem
}

// buildColumnDefs builds column definitions from result data for frontend display.
func buildColumnDefs(data []map[string]interface{}) []map[string]string {
	if len(data) == 0 {
		return []map[string]string{}
	}
	var cols []map[string]string
	for name := range data[0] {
		colType := "string"
		switch data[0][name].(type) {
		case int, int8, int16, int32, int64, float32, float64:
			colType = "number"
		}
		cols = append(cols, map[string]string{"name": name, "type": colType})
	}
	return cols
}