package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Config struct to hold the database configuration
type Config struct {
	Host     string
	Port     string
	Name     string
	Username string
	Password string
}

// Metrics struct to hold the Prometheus metrics
type Metrics struct {
	IndividualCounts *prometheus.GaugeVec
	TotalCount       prometheus.Gauge
}

// ReadConfig reads the database configuration from the specified file
func ReadConfig(filePath string) (Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	config := Config{}
	re := regexp.MustCompile(`database\.0\.(\w+)\s*=\s*(.+)`)

	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindStringSubmatch(line)
		if len(matches) == 3 {
			switch matches[1] {
			case "host":
				config.Host = matches[2]
			case "port":
				config.Port = matches[2]
			case "name":
				config.Name = matches[2]
			case "username":
				config.Username = matches[2]
			case "password":
				config.Password = matches[2]
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return Config{}, err
	}

	return config, nil
}

// NewMetrics creates new Prometheus metrics
func NewMetrics() *Metrics {
	return &Metrics{
		IndividualCounts: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "fusionpbx_individual_caller_destination_count",
				Help: "Count of calls to individual caller destinations",
			},
			[]string{"destination"},
		),
		TotalCount: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "fusionpbx_total_caller_destination_count",
				Help: "Total count of calls to all gateways",
			},
		),
	}
}

// QueryDB queries the database and updates the metrics
func QueryDB(db *sql.DB, metrics *Metrics) error {
	// Query for individual gateways
	rows, err := db.Query("SELECT DISTINCT caller_destination FROM v_xml_cdr WHERE caller_destination LIKE 'gw+%'")
	if err != nil {
		return err
	}
	defer rows.Close()

	destinations := []string{}
	for rows.Next() {
		var destination string
		if err := rows.Scan(&destination); err != nil {
			return err
		}
		destinations = append(destinations, destination)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, destination := range destinations {
		var count int
		query := fmt.Sprintf("SELECT COUNT(*) FROM v_xml_cdr WHERE caller_destination LIKE '%s'", destination)
		if err := db.QueryRow(query).Scan(&count); err != nil {
			return err
		}
		metrics.IndividualCounts.With(prometheus.Labels{"destination": destination}).Set(float64(count))
	}

	// Query for all gateways
	var totalCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM v_xml_cdr WHERE caller_destination LIKE 'gw+%'").Scan(&totalCount); err != nil {
		return err
	}
	metrics.TotalCount.Set(float64(totalCount))

	return nil
}

func main() {
	// Read configuration
	configPath := os.Getenv("FPB_IC_EXP_FUSION_CONFIG_FILE")
	if configPath == "" {
		configPath = "/etc/fusionpbx/config.conf"
	}

	config, err := ReadConfig(configPath)
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	// Open database connection
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s",
		config.Host, config.Port, config.Username, config.Password, config.Name)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	// Register Prometheus metrics
	metrics := NewMetrics()
	prometheus.MustRegister(metrics.IndividualCounts)
	prometheus.MustRegister(metrics.TotalCount)

	// Query database and update metrics periodically
	go func() {
		for {
			if err := QueryDB(db, metrics); err != nil {
				log.Printf("Error querying database: %v", err)
			}
			time.Sleep(3 * time.Second) // Adjust the interval as needed
		}
	}()

	// Expose metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	port := os.Getenv("FPB_IC_EXP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on :%s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
