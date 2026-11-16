package api

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
)

// UpdateTargetTemperature safely updates room target temperature parameters with RLS validation.
func UpdateTargetTemperature(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        if db == nil {
            log.Println("❌ [ROOM UPDATE] Database connection is nil")
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
            return
        }

        userID, exists := getUserIDFromContext(c)
        if !exists || userID == nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: user context missing"})
            return
        }

        roomIDStr := c.Param("id")
        var roomID int
        if _, err := fmt.Sscanf(roomIDStr, "%d", &roomID); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID format"})
            return
        }

        var input struct {
            TargetTemperature float64 `json:"target_temperature"`
        }

        if err := c.ShouldBindJSON(&input); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload parameters"})
            return
        }

        ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
        defer cancel()

        tx, err := db.BeginTx(ctx, nil)
        if err != nil {
            log.Printf("❌ [ROOM UPDATE TX ERROR] BeginTx failed: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize transaction"})
            return
        }
        defer func() {
            _ = tx.Rollback()
        }()

        userIDStr := fmt.Sprintf("%v", userID)
        if _, err := tx.ExecContext(ctx, "SELECT set_config('app.current_user_id', $1, true)", userIDStr); err != nil {
            log.Printf("❌ [ROOM UPDATE RLS ERROR] Failed to set app.current_user_id: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to configure session context"})
            return
        }

        query := `UPDATE rooms SET target_temperature = $1 WHERE id = $2`
        result, err := tx.ExecContext(ctx, query, input.TargetTemperature, roomID)
        if err != nil {
            log.Printf("❌ [ROOM UPDATE DB ERROR] Query failed: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update room parameters"})
            return
        }

        rowsAffected, err := result.RowsAffected()
        if err != nil || rowsAffected == 0 {
            c.JSON(http.StatusNotFound, gin.H{"error": "Room not found or modification unauthorized"})
            return
        }

        if err := tx.Commit(); err != nil {
            log.Printf("❌ [ROOM UPDATE COMMIT ERROR] Transaction commit failed: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to finalize transaction"})
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "message":            "Target temperature updated successfully",
            "room_id":            roomID,
            "target_temperature": input.TargetTemperature,
        })
    }
}