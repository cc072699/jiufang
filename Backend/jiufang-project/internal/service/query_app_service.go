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
// It integrates the AI Agent for semantic understanding and SQL generation.
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

	// Create query agent
	queryAgent := agentpkg.NewQueryAgent(llmClient, erpReader, logger)

	// Create conversation manager
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

	// Create query agent
	queryAgent := agentpkg.NewQueryAgent(llmClient, erpReader, logger)

	// Create conversation manager
	conversationManager := agentpkg.NewConversationManager(10)

	// Create visualization service
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
	// Generate snowflake ID for this query
	queryID := s.idGenerator.Generate()

	// Execute query through agent
	result, err := s.agent.Execute(ctx, input, sessionID, userID)
	if err != nil {
		// Log error
		s.logger.Error("Query execution failed",
			zap.String("input", input),
			zap.String("session_id", sessionID),
			zap.Uint("user_id", userID),
			zap.Error(err),
		)

		// Update conversation history with failure
		s.conversationManager.UpdateConversation(sessionID, &agentpkg.QueryHistoryItem{
			Timestamp: time.Now(),
			Input:     input,
			Intent:    agent.IntentTypeStatistics,
			Result:    nil,
			Success:   false,
		})

		return nil, err
	}

	// Set query ID
	result.QueryID = queryID

	// Update conversation history
	s.conversationManager.UpdateConversation(sessionID, &agentpkg.QueryHistoryItem{
		Timestamp: result.Timestamp,
		Input:     input,
		Intent:    agent.IntentTypeStatistics,
		Result:    result,
		Success:   true,
	})

	// Save query record to database if repository is available
	if s.queryRepo != nil {
		// Convert sessionID from string to int64
		var sessionIDInt int64
		if _, err := fmt.Sscanf(sessionID, "%d", &sessionIDInt); err != nil {
			s.logger.Warn("Failed to parse sessionID", zap.Error(err))
			sessionIDInt = 0
		}

		// Convert result.Data to JSON string
		resultDataJSON, err := json.Marshal(result.Data)
		if err != nil {
			s.logger.Warn("Failed to marshal result data", zap.Error(err))
			resultDataJSON = []byte("[]")
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

	// Generate visualization if needed
	if result.VisualizationType != agent.ChartTypeTable && s.visualizationService != nil {
		// Get columns from data
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
	// Get conversation state
	state := s.conversationManager.GetConversation(req.SessionID)

	// Build query context
	queryContext := &agent.QueryContext{
		UserID:    req.UserID,
		SessionID: req.SessionID,
	}

	// Set user role based on last intent
	if state.LastIntent != nil {
		queryContext.UserRole = string(state.LastIntent.Type)
	}

	// If this is a follow-up question, use previous entities
	if req.IsFollowUp && len(state.CurrentEntities) > 0 {
		queryContext.PreviousContext = &agent.EntityCollection{}
		// Note: EntityCollection.ToEntityList() method needs to be implemented
		// For now, we skip adding entities
	}

	// Execute with context
	response, err := s.agent.ExecuteWithContext(ctx, req)
	if err != nil {
		return nil, err
	}

	// Update conversation state
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
	// Execute with permission check
	result, err := s.agent.ExecuteWithPermission(ctx, input, queryContext)
	if err != nil {
		return nil, err
	}

	// Update conversation history
	s.conversationManager.UpdateConversation(queryContext.SessionID, &agentpkg.QueryHistoryItem{
		Timestamp: result.Timestamp,
		Input:     input,
		Result:    result,
		Success:   true,
	})

	return result, nil
}

// StreamQuery executes a query with streaming response.
func (s *QueryAppService) StreamQuery(ctx context.Context, input string, sessionID string, userID uint) (*schema.StreamReader[*agent.QueryResult], error) {
	// Stream query through agent
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
	// Get conversation state
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
	GetSuggestedQuestions(sessionID string) []string
}
