package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HousingPriceIndex struct {
	ID                   int       `json:"id"`
	Tarih                time.Time `json:"tarih"`
	IstanbulTurkiye      string    `json:"istanbul_turkiye"`
	YeniYeniOlmayanKonut string    `json:"yeni_yeni_olmayan_konut"`
	FiyatEndeksi         float64   `json:"fiyat_endeksi"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type HousingStats struct {
	LastMonthIndex      float64 `json:"last_month_index"`
	ChangeFromStart     float64 `json:"change_from_start_percentage"`
	LastYearIncrease    float64 `json:"last_year_increase_percentage"`
	MaxValue            float64 `json:"max_value"`
	MinValue            float64 `json:"min_value"`
	LastMonthDate       string  `json:"last_month_date"`
}

// handleGetHousingData handles filtering and retrieving housing data
func handleGetHousingData(w http.ResponseWriter, r *http.Request, dbPool *pgxpool.Pool) {
	ctx := context.Background()

	// Parse query parameters
	location := r.URL.Query().Get("location")
	konutType := r.URL.Query().Get("type")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	// Build query
	query := "SELECT id, tarih, istanbul_turkiye, yeni_yeni_olmayan_konut, fiyat_endeksi, created_at, updated_at FROM housing_price_index WHERE 1=1"
	var args []interface{}
	argCount := 1

	if location != "" {
		query += fmt.Sprintf(" AND istanbul_turkiye = $%d", argCount)
		args = append(args, location)
		argCount++
	}

	if konutType != "" {
		query += fmt.Sprintf(" AND yeni_yeni_olmayan_konut = $%d", argCount)
		args = append(args, konutType)
		argCount++
	}

	if startDate != "" {
		query += fmt.Sprintf(" AND tarih >= $%d", argCount)
		args = append(args, startDate)
		argCount++
	}

	if endDate != "" {
		query += fmt.Sprintf(" AND tarih <= $%d", argCount)
		args = append(args, endDate)
		argCount++
	}

	query += " ORDER BY tarih DESC, istanbul_turkiye, yeni_yeni_olmayan_konut"

	rows, err := dbPool.Query(ctx, query, args...)
	if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var results []HousingPriceIndex
	for rows.Next() {
		var h HousingPriceIndex
		err := rows.Scan(&h.ID, &h.Tarih, &h.IstanbulTurkiye, &h.YeniYeniOlmayanKonut, &h.FiyatEndeksi, &h.CreatedAt, &h.UpdatedAt)
		if err != nil {
			http.Error(w, fmt.Sprintf("Scan error: %v", err), http.StatusInternalServerError)
			return
		}
		results = append(results, h)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count": len(results),
		"data":  results,
	})
}

// handleGetHousingStats calculates KPIs
func handleGetHousingStats(w http.ResponseWriter, r *http.Request, dbPool *pgxpool.Pool) {
	ctx := context.Background()

	location := r.URL.Query().Get("location")
	konutType := r.URL.Query().Get("type")

	if location == "" || konutType == "" {
		http.Error(w, "location and type parameters are required for stats", http.StatusBadRequest)
		return
	}

	// Get latest record
	var lastRecord HousingPriceIndex
	err := dbPool.QueryRow(ctx, `
		SELECT tarih, fiyat_endeksi 
		FROM housing_price_index 
		WHERE istanbul_turkiye = $1 AND yeni_yeni_olmayan_konut = $2 
		ORDER BY tarih DESC LIMIT 1`, location, konutType).Scan(&lastRecord.Tarih, &lastRecord.FiyatEndeksi)

	if err == pgx.ErrNoRows {
		http.Error(w, "No data found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}

	// Get first record
	var firstRecord HousingPriceIndex
	err = dbPool.QueryRow(ctx, `
		SELECT fiyat_endeksi 
		FROM housing_price_index 
		WHERE istanbul_turkiye = $1 AND yeni_yeni_olmayan_konut = $2 
		ORDER BY tarih ASC LIMIT 1`, location, konutType).Scan(&firstRecord.FiyatEndeksi)
	if err != nil {
		http.Error(w, fmt.Sprintf("Database error (first record): %v", err), http.StatusInternalServerError)
		return
	}

	// Get record from 1 year ago
	oneYearAgo := lastRecord.Tarih.AddDate(-1, 0, 0)
	var yearAgoRecord HousingPriceIndex
	// Find closest record to 1 year ago
	err = dbPool.QueryRow(ctx, `
		SELECT fiyat_endeksi 
		FROM housing_price_index 
		WHERE istanbul_turkiye = $1 AND yeni_yeni_olmayan_konut = $2 AND tarih <= $3
		ORDER BY tarih DESC LIMIT 1`, location, konutType, oneYearAgo).Scan(&yearAgoRecord.FiyatEndeksi)
	
	// If no record exactly 1 year ago or before, try to find the earliest one after
	if err == pgx.ErrNoRows {
		 err = dbPool.QueryRow(ctx, `
		SELECT fiyat_endeksi 
		FROM housing_price_index 
		WHERE istanbul_turkiye = $1 AND yeni_yeni_olmayan_konut = $2
		ORDER BY tarih ASC LIMIT 1`, location, konutType).Scan(&yearAgoRecord.FiyatEndeksi)
	}

	if err != nil && err != pgx.ErrNoRows {
		http.Error(w, fmt.Sprintf("Database error (year ago): %v", err), http.StatusInternalServerError)
		return
	}

	// Get Max/Min
	var maxVal, minVal float64
	err = dbPool.QueryRow(ctx, `
		SELECT MAX(fiyat_endeksi), MIN(fiyat_endeksi)
		FROM housing_price_index 
		WHERE istanbul_turkiye = $1 AND yeni_yeni_olmayan_konut = $2`, location, konutType).Scan(&maxVal, &minVal)
	if err != nil {
		http.Error(w, fmt.Sprintf("Database error (min/max): %v", err), http.StatusInternalServerError)
		return
	}

	stats := HousingStats{
		LastMonthIndex:   lastRecord.FiyatEndeksi,
		LastMonthDate:    lastRecord.Tarih.Format("2006-01-02"),
		MaxValue:         maxVal,
		MinValue:         minVal,
	}

	if firstRecord.FiyatEndeksi > 0 {
		stats.ChangeFromStart = ((lastRecord.FiyatEndeksi - firstRecord.FiyatEndeksi) / firstRecord.FiyatEndeksi) * 100
	}

	if yearAgoRecord.FiyatEndeksi > 0 {
		stats.LastYearIncrease = ((lastRecord.FiyatEndeksi - yearAgoRecord.FiyatEndeksi) / yearAgoRecord.FiyatEndeksi) * 100
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// handleGetHousingCharts returns data formatted for charts
func handleGetHousingCharts(w http.ResponseWriter, r *http.Request, dbPool *pgxpool.Pool) {
	// For now, reuse handleGetHousingData logic but maybe format differently if needed.
	// The frontend can likely use the raw data from handleGetHousingData to build charts.
	// But let's provide a specific endpoint if we want to pre-aggregate or format.
	
	// Requirement: 
	// 1. Time Series: Date vs Price Index (filtered by location/type)
	// 2. Comparison: Istanbul vs Turkey (same type)
	// 3. Comparison: New vs Not New (same location)

	chartType := r.URL.Query().Get("chart_type")
	
	if chartType == "comparison_location" {
		// Compare Istanbul vs Turkey for a specific type
		konutType := r.URL.Query().Get("type")
		if konutType == "" {
			http.Error(w, "type parameter is required for comparison_location", http.StatusBadRequest)
			return
		}
		
		// Fetch all data for this type
		handleGetHousingData(w, r, dbPool) // Reuse for now, frontend filters/groups
		return
	}

	if chartType == "comparison_type" {
		// Compare New vs Not New for a specific location
		location := r.URL.Query().Get("location")
		if location == "" {
			http.Error(w, "location parameter is required for comparison_type", http.StatusBadRequest)
			return
		}
		
		// Fetch all data for this location
		handleGetHousingData(w, r, dbPool) // Reuse for now
		return
	}

	// Default: just return data
	handleGetHousingData(w, r, dbPool)
}
