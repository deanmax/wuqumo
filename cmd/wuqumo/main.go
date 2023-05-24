package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/deanmax/wuqumo/internal/pkg/helper"
	"github.com/go-chi/chi/v5"
)

type TranscodeObject struct {
	ID     uint32 `json:"id,omitempty"`
	LCP    string `json:"lcp,omitempty"`
	Wuqumo string `json:"wuqumo,omitempty"`
}

var (
	transcodeObjectCache []TranscodeObject
	cacheExpires         time.Time
	cacheLock            sync.RWMutex
)

func main() {
	router := chi.NewRouter()

	// Define routes
	router.Get("/mapping", getTranscodeMapping)
	router.Get("/bywuqumo/{id}", getCodebyWuqumo)
	router.Get("/bylcp/{id}", getCodebyLCP)

	// Periodically refresh cache
	go refreshCache()

	log.Fatal(http.ListenAndServe(":8000", router))
}

// Get all LineColumePage(LCP) to Wuqumo code mapping
func getTranscodeMapping(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(transcodeObjectCache)
}

// Get transocde by LCP
func getCodebyLCP(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	cacheLock.RLock()
	defer cacheLock.RUnlock()

	for _, object := range transcodeObjectCache {
		if object.LCP == id {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(object)
			return
		}
	}
	// Not found
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "LCP ID %s not found!", id)
}

// Get transocde by Wuqumo
func getCodebyWuqumo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	cacheLock.RLock()
	defer cacheLock.RUnlock()

	for _, object := range transcodeObjectCache {
		if object.Wuqumo == id {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(object)
			return
		}
	}
	// Not found
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "Wuqumo ID %s not found!", id)
}

// Refresh cache periodically
func refreshCache() {
	for {
		// Wait for cache to expire
		time.Sleep(time.Until(cacheExpires))

		// Cache has expired, refresh data and update cache
		file, err := os.Open(helper.GetEnvDefault("CSV_FILE", "./transcode_sheet.csv"))
		if err != nil {
			log.Fatalf("Unable to read CSV file: %v", err)
		}
		defer file.Close()

		// Parse CSV file
		reader := csv.NewReader(file)
		reader.FieldsPerRecord = 11 // Total 11 columns, only first 3 are useful
		records, err := reader.ReadAll()
		if err != nil {
			log.Fatalf("Unable to parse CSV file: %v", err)
		}

		// Map data to User struct
		var objects []TranscodeObject
		for idx, row := range records {
			if row[2] == "" {
				continue
			}

			record := TranscodeObject{
				ID:     uint32(idx) + 1,
				Wuqumo: row[0],
				LCP:    row[2],
			}
			objects = append(objects, record)
		}

		// Update cache
		cacheLock.Lock()
		transcodeObjectCache = objects
		cacheExpires = time.Now().Add(time.Minute * 5) // Cache expires in 5 minutes
		cacheLock.Unlock()
	}
}
