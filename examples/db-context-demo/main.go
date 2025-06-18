package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	agents "github.com/logkn/agents-go/pkg"
	// To use a real database like SQLite, add:
	// "database/sql"
	// _ "github.com/mattn/go-sqlite3" // SQLite driver
)

// DBContext holds database connection and user information
type DBContext struct {
	DB     *MockDB
	UserID string
}

// QueryTool executes SQL queries with context awareness
type QueryTool struct {
	Query string `json:"query" description:"SQL query to execute (SELECT only for safety)"`
}

func (q QueryTool) Run() any {
	return "Database connection not available"
}

func (q QueryTool) RunWithAnyContext(ctx agents.AnyContext) any {
	if ctx == nil {
		return q.Run()
	}

	dbCtx, err := agents.FromAnyContext[DBContext](ctx)
	if err != nil {
		return fmt.Sprintf("Error accessing database context: %v", err)
	}

	db := dbCtx.Value().DB
	userID := dbCtx.Value().UserID

	// Safety check - only allow SELECT queries
	if len(q.Query) < 6 || q.Query[:6] != "SELECT" {
		return "Only SELECT queries are allowed"
	}

	// Add user context to query (example: user-specific filtering)
	log.Printf("Executing query for user %s: %s", userID, q.Query)

	rows, err := db.Query(q.Query)
	if err != nil {
		return fmt.Sprintf("Query error: %v", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Sprintf("Error getting columns: %v", err)
	}

	// Collect results
	var results []map[string]any
	values := make([]any, len(columns))
	valuePtrs := make([]any, len(columns))

	for rows.Next() {
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return fmt.Sprintf("Scan error: %v", err)
		}

		entry := make(map[string]any)
		for i, col := range columns {
			var v any
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			entry[col] = v
		}
		results = append(results, entry)
	}

	return fmt.Sprintf("Query returned %d rows: %v", len(results), results)
}

// UserInfoTool gets current user information from context
type UserInfoTool struct{}

func (u UserInfoTool) Run() any {
	return "User information not available"
}

func (u UserInfoTool) RunWithAnyContext(ctx agents.AnyContext) any {
	if ctx == nil {
		return u.Run()
	}

	dbCtx, err := agents.FromAnyContext[DBContext](ctx)
	if err != nil {
		return u.Run()
	}

	userID := dbCtx.Value().UserID
	db := dbCtx.Value().DB

	var username, email string
	err = db.QueryRow("SELECT username, email FROM users WHERE id = ?", userID).Scan(&username, &email)
	if err != nil {
		return fmt.Sprintf("Error fetching user info: %v", err)
	}

	return fmt.Sprintf("Current user: %s (ID: %s, Email: %s)", username, userID, email)
}

// MockDB represents a mock database for demonstration purposes
type MockDB struct {
	users    map[string]map[string]string
	products []map[string]any
}

func (m *MockDB) QueryRow(query string, args ...any) *MockRow {
	// Mock user lookup
	if query == "SELECT username, email FROM users WHERE id = ?" && len(args) > 0 {
		userID := args[0].(string)
		if user, exists := m.users[userID]; exists {
			return &MockRow{data: []any{user["username"], user["email"]}, err: nil}
		}
	}
	return &MockRow{err: fmt.Errorf("user not found")}
}

func (m *MockDB) Query(query string) (*MockRows, error) {
	// Mock product queries
	if query == "SELECT * FROM products" {
		return &MockRows{data: m.products}, nil
	}
	if query == "SELECT * FROM products WHERE price < 100" {
		var filtered []map[string]any
		for _, product := range m.products {
			if price, ok := product["price"].(float64); ok && price < 100 {
				filtered = append(filtered, product)
			}
		}
		return &MockRows{data: filtered}, nil
	}
	return nil, fmt.Errorf("query not supported in mock")
}

type MockRow struct {
	data []any
	err  error
}

func (r *MockRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	for i, d := range dest {
		if i < len(r.data) {
			*d.(*string) = r.data[i].(string)
		}
	}
	return nil
}

type MockRows struct {
	data []map[string]any
	pos  int
}

func (r *MockRows) Close() error { return nil }
func (r *MockRows) Next() bool   { return r.pos < len(r.data) }
func (r *MockRows) Columns() ([]string, error) {
	if len(r.data) > 0 {
		var cols []string
		for k := range r.data[0] {
			cols = append(cols, k)
		}
		return cols, nil
	}
	return []string{}, nil
}

func (r *MockRows) Scan(dest ...any) error {
	if r.pos >= len(r.data) {
		return fmt.Errorf("no more rows")
	}

	cols, _ := r.Columns()
	for i, d := range dest {
		if i < len(cols) {
			val := r.data[r.pos][cols[i]]
			*d.(*any) = val
		}
	}
	r.pos++
	return nil
}

func setupDatabase() (*MockDB, error) {
	// Create mock database with sample data
	db := &MockDB{
		users: map[string]map[string]string{
			"user123": {"username": "alice", "email": "alice@example.com"},
			"user456": {"username": "bob", "email": "bob@example.com"},
		},
		products: []map[string]any{
			{"id": 1, "name": "Laptop", "price": 999.99, "stock": 10},
			{"id": 2, "name": "Mouse", "price": 29.99, "stock": 50},
			{"id": 3, "name": "Keyboard", "price": 79.99, "stock": 30},
			{"id": 4, "name": "Monitor", "price": 299.99, "stock": 15},
		},
	}

	return db, nil
}

func main() {
	// Setup database
	db, err := setupDatabase()
	if err != nil {
		log.Fatalf("Failed to setup database: %v", err)
	}
	// No need to close mock database

	// Create database context for a specific user
	dbContext := agents.NewContext(DBContext{
		DB:     db,
		UserID: "user123", // Simulate logged-in user
	})

	// Create lifecycle hooks
	hooks := &agents.LifecycleHooks{
		BeforeRun: func(ctx agents.AnyContext) error {
			fmt.Println("🔧 Initializing database agent...")
			return nil
		},
		BeforeToolCall: func(ctx agents.AnyContext, toolName string, args string) error {
			fmt.Printf("📊 Executing tool: %s with args: %s\n", toolName, args)
			return nil
		},
		AfterToolCall: func(ctx agents.AnyContext, toolName string, result any) error {
			fmt.Printf("✅ Tool %s completed\n", toolName)
			return nil
		},
	}

	// Create tools
	queryTool := agents.NewContextualTool(
		"query_database",
		"Execute SELECT queries on the database",
		&QueryTool{},
		dbContext,
	)

	userInfoTool := agents.NewContextualTool(
		"get_user_info",
		"Get information about the current user",
		&UserInfoTool{},
		dbContext,
	)

	// Create agent
	config := agents.AgentConfig{
		Name: "Database Assistant",
		Instructions: `You are a helpful database assistant. You can:
1. Execute SELECT queries to retrieve data from the database
2. Get information about the current user
3. Help users understand the data

Available tables:
- users (id, username, email)
- products (id, name, price, stock)

Always be helpful and explain the results clearly.`,
		Model: agents.ModelConfig{
			Model: "gpt-4o-mini",
		},
		Logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})),
	}

	agent := agents.NewAgent(config)
	agent = agents.WithTools(agent, queryTool, userInfoTool)
	agent = agents.WithHooks(agent, hooks)

	// Demo queries
	fmt.Println("=== Database Context Demo ===")
	fmt.Println("Connected to mock database")
	fmt.Println("Current user context: user123 (alice)")
	fmt.Println()

	// Run the agent
	queries := []string{
		"Who am I? Can you tell me about my user account?",
		"Show me all the products in the database with their prices and stock levels.",
		"Which products cost less than $100?",
	}

	for i, query := range queries {
		fmt.Printf("\n--- Query %d: %s ---\n", i+1, query)

		response, err := agents.Run(context.Background(), agent, agents.Input{
			OfString: query,
		})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		// Stream the response
		for event := range response.Stream() {
			if token, ok := event.Token(); ok && token != "" {
				fmt.Print(token)
			}
		}
		fmt.Println()
	}
}
