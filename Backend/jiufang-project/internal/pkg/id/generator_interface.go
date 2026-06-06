package id

// SnowflakeGeneratorInterface defines the interface for snowflake ID generation.
type SnowflakeGeneratorInterface interface {
	Generate() int64
}

// Ensure SnowflakeGenerator implements SnowflakeGeneratorInterface
var _ SnowflakeGeneratorInterface = (*SnowflakeGenerator)(nil)