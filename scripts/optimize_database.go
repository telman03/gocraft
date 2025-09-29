//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/telman03/ai-backend-generator/internal/database"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Initialize database
	database.InitDB()

	fmt.Println("Optimizing database indexes for better performance...")

	// Add composite index for maintenance queries
	queries := []string{
		// Composite index for cleanup queries (created_at + zip_file_status)
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_project_history_cleanup 
		 ON project_history (created_at, zip_file_status) 
		 WHERE zip_file_status IN ('expired', 'deleted')`,

		// Index for archival queries
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_project_history_archival 
		 ON project_history (created_at) 
		 WHERE zip_file_status = 'expired'`,

		// Index for user queries (already exists but ensure it's optimal)
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_project_history_user_created 
		 ON project_history (user_id, created_at DESC)`,

		// Analyze tables to update statistics
		`ANALYZE project_history`,
		`ANALYZE users`,
	}

	for i, query := range queries {
		fmt.Printf("Executing optimization %d/%d...\n", i+1, len(queries))

		if err := database.DB.Exec(query).Error; err != nil {
			log.Printf("Warning: Failed to execute query: %v\nQuery: %s", err, query)
		} else {
			fmt.Printf("✅ Optimization %d completed\n", i+1)
		}
	}

	fmt.Println("\n🚀 Database optimization completed!")
	fmt.Println("Expected improvements:")
	fmt.Println("  - Faster maintenance cleanup queries")
	fmt.Println("  - Improved archival performance")
	fmt.Println("  - Better user history query performance")
	fmt.Println("  - Updated table statistics for query planner")
}
