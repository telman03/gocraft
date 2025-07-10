package supabase

// import (
// 	"context"
// 	"encoding/json"
// 	"log"

// 	"github.com/supabase-community/supabase-go"
// )

// var supabaseClient *supabase.Client

// func Init(client *supabase.Client) {
// 	supabaseClient = client
// }
// func CheckUsageLimit(userID string) bool {
// 	resp, err := supabaseClient.From("public").Select("check_daily_usage", supabase.Map{
// 		"user_id": userID,
// 	}).Execute()
// 	if err != nil {
// 		log.Println("RPC failed:", err)
// 		return false
// 	}

// 	var usageCount int
// 	err = json.NewDecoder(resp.Body).Decode(&usageCount)
// 	if err != nil {
// 		log.Println("Decode error:", err)
// 		return false
// 	}

// 	return usageCount < 5 // Limit can be made configurable
// }
