package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	if err := initDB(); err != nil {
		log.Fatal("Ошибка инициализации БД:", err)
	}
	fmt.Println("База данных подключена")

	go backgroundPredictor()

	http.HandleFunc("/api/latest_data", handleLatestData)
	http.HandleFunc("/api/latest_prediction", handleLatestPrediction)
	http.HandleFunc("/api/predictions_history", handlePredictionsHistory)
	http.HandleFunc("/api/notifications", handleNotifications)
	http.HandleFunc("/api/mark_notification/", handleMarkNotification)
	http.HandleFunc("/api/feature_importance", handleFeatureImportance)
	http.HandleFunc("/api/predict", handlePredict)
	http.HandleFunc("/", handleStatic)

	fmt.Println("Go сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func backgroundPredictor() {
	for {
		time.Sleep(1 * time.Minute)

		readings, err := getRecentReadings(15)
		if err != nil {
			log.Println("Ошибка получения данных:", err)
			continue
		}

		if len(readings) < 10 {
			log.Printf("Недостаточно данных для прогноза (%d записей)", len(readings))
			continue
		}

		features := make([][]float64, len(readings))
		for i, r := range readings {
			features[i] = []float64{r.Current, r.Voltage, r.Pressure, r.Temperature, r.Vibration}
		}

		predictions, err := callMLService(features)
		if err != nil {
			log.Println("Ошибка ML сервиса:", err)
			continue
		}

		if len(predictions) == 0 {
			continue
		}

		currentRUL := int(predictions[len(predictions)-1])

		var status string
		if currentRUL > 250 {
			status = "Норма"
		} else if currentRUL > 50 {
			status = "Внимание"
		} else {
			status = "Критично"
		}

		if err := savePrediction(currentRUL, status, ""); err != nil {
			log.Println("Ошибка сохранения прогноза:", err)
			continue
		}

		if status == "Критично" {
			addNotification("danger", fmt.Sprintf("КРИТИЧНО! Остаточный ресурс: %d циклов", currentRUL))
		} else if status == "Внимание" {
			addNotification("warning", fmt.Sprintf("Внимание! Остаточный ресурс: %d циклов", currentRUL))
		}

		log.Printf("Прогноз обновлен: RUL=%d, статус=%s", currentRUL, status)
	}
}
