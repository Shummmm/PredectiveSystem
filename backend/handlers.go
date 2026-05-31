package main

import (
    "encoding/json"
    "net/http"
    "strconv"
    "strings"
)

func enableCORS(w http.ResponseWriter) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func handleLatestData(w http.ResponseWriter, r *http.Request) {
    enableCORS(w)
    if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }

    minutes := 15
    if m := r.URL.Query().Get("minutes"); m != "" {
        if val, err := strconv.Atoi(m); err == nil {
            minutes = val
        }
    }

    readings, err := getRecentReadings(minutes)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "data":  readings,
        "count": len(readings),
    })
}

func handleLatestPrediction(w http.ResponseWriter, r *http.Request) {
    enableCORS(w)
    if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }

    pred, err := getLatestPrediction()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    if pred == nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "rul":    nil,
            "status": "Нет данных",
        })
        return
    }
    json.NewEncoder(w).Encode(pred)
}

func handlePredictionsHistory(w http.ResponseWriter, r *http.Request) {
    enableCORS(w)
    if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }

    limit := 20
    if l := r.URL.Query().Get("limit"); l != "" {
        if val, err := strconv.Atoi(l); err == nil {
            limit = val
        }
    }

    predictions, err := getPredictionsHistory(limit)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(predictions)
}

func handleNotifications(w http.ResponseWriter, r *http.Request) {
    enableCORS(w)
    if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }

    notifications, err := getUnreadNotifications()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(notifications)
}

func handleMarkNotification(w http.ResponseWriter, r *http.Request) {
    enableCORS(w)
    if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }

    parts := strings.Split(r.URL.Path, "/")
    if len(parts) < 4 {
        http.Error(w, "Invalid path", http.StatusBadRequest)
        return
    }
    id, err := strconv.Atoi(parts[3])
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    if err := markNotificationAsRead(id); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func handleFeatureImportance(w http.ResponseWriter, r *http.Request) {
    enableCORS(w)
    if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }

    importance, err := getFeatureImportance()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(importance)
}

func handlePredict(w http.ResponseWriter, r *http.Request) {
    enableCORS(w)
    if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }

    var req struct {
        Features [][]float64 `json:"features"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    predictions, err := callMLService(req.Features)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string][]float64{"predictions": predictions})
}

func handleStatic(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path == "/" {
        http.ServeFile(w, r, "../static/index.html")
        return
    }
    http.ServeFile(w, r, "../static"+r.URL.Path)
}

func handlePredictAndSave(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		Features [][]float64 `json:"features"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	predictions, err := callMLService(req.Features)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(predictions) == 0 {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Нет предсказаний"})
		return
	}

	currentRUL := int(predictions[len(predictions)-1])

	// ПРАВИЛЬНОЕ ОПРЕДЕЛЕНИЕ СТАТУСА
	var status string
	if currentRUL > 200 {
		status = "Норма"
	} else if currentRUL > 50 {
		status = "Внимание"
	} else {
		status = "Критично"
	}

	// Сохраняем в БД
	if err := savePrediction(currentRUL, status, ""); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"predictions": predictions,
		"saved":       true,
		"rul":         currentRUL,
		"status":      status,
	})
}
