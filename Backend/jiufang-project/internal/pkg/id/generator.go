package id

// SnowflakeGenerator wraps Snowflake to provide a simple interface for ID generation.
type SnowflakeGenerator struct {
	snowflake *Snowflake
}

// NewSnowflakeGenerator creates a new snowflake generator with the given node ID.
func NewSnowflakeGenerator(node int64) (*SnowflakeGenerator, error) {
	snowflake, err := NewSnowflake(node)
	if err != nil {
		return nil, err
	}
	return &SnowflakeGenerator{snowflake: snowflake}, nil
}

// Generate generates a new unique snowflake ID.
func (g *SnowflakeGenerator) Generate() int64 {
	return g.snowflake.Generate()
}
