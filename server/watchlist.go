package main

import (
	"github.com/gin-gonic/gin"
    "net/http"
	"gorm.io/gorm"

)
func exportWatchedData(db *gorm.DB, userID uint, c *gin.Context) ([]Watched, error) {
    watchedData, err := getWatched(db, userID)

	if err != nil {
        return nil, err
    }

    format := c.DefaultQuery("format", "json") // Default format is JSON, implement CSV later
    
    // if format == "csv" {
    //     csvData, err := convertToCSV(watchedData)
    //     if err != nil {
    //         c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert data to CSV"})
    //         return
    //     }
    //     c.Header("Content-Type", "text/csv")
    //     c.Header("Content-Disposition", "attachment; filename=watchedData.csv")
    //     return csvData, nil
    // } else {  // Default is JSON
        c.Header("Content-Type", "application/json")
        c.Header("Content-Disposition", "attachment; filename=watchedData.json")
		
		return watchedData, nil
    // }
}

func getWatched(db *gorm.DB, userId uint) ([]Watched, error) {
	watched := new([]Watched)
	res := db.Model(&Watched{}).Preload("Content").Preload("Activity").Where("user_id = ?", userId).Find(&watched)
	if res.Error != nil {
		return nil, res.Error
	}
	return *watched, nil
}

func convertToCSV(watchedData []WatchedData) (string, error) {
    // Convert your data to CSV format and return it
    // Libraries like "encoding/csv" can be helpful
}
