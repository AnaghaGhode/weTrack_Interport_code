package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"
)

type Event struct {
	EventType string                 `json:"event_type"`
	Timestamp string                 `json:"timestamp"`
	Location  string                 `json:"location"`
	Metadata  map[string]interface{} `json:"metadata"`
}

type Shipment struct {
	ContainerID string  `json:"container_id"`
	Events      []Event `json:"events"`
}

type TimelineEvent struct {
	EventType    string `json:"event_type"`
	Timestamp    string `json:"timestamp"`
	Location     string `json:"location"`
	DelayMinutes int    `json:"delay_minutes,omitempty"`
}

type Anomaly struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type Report struct {
	ContainerID     string          `json:"container_id"`
	CurrentStatus   string          `json:"current_status"`
	CurrentLocation string          `json:"current_location"`
	LastEventTime   time.Time       `json:"last_event_time"`
	Timeline        []TimelineEvent `json:"timeline"`
	Anomalies       []Anomaly       `json:"anomalies"`
	JourneyProgress int             `json:"journey_progress"`
}

var statusMap = map[string]string{
	"port_arrival":          "arrived_at_port",
	"transshipment_arrival": "arrived_transshipment",
	"customs_hold":          "held_by_customs",
	"customs_clearance":     "cleared",
	"port_departure":        "departed_port",
	"in_transit":            "in_transit",
	"road_checkpoint":       "on_road",
	"lcl_pickup":            "picked_up",
}

var eventOrder = map[string]int{
	"port_arrival":          1,
	"transshipment_arrival": 2,
	"customs_hold":          3,
	"customs_clearance":     4,
	"port_departure":        5,
	"in_transit":            6,
	"road_checkpoint":       7,
	"lcl_pickup":            8,
}

func main() {

	data, err := os.ReadFile("20_shipments_detailed_input_wetrack_developer_assignment.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read file: %v\n", err)
		os.Exit(1)
	}

	var shipments []Shipment
	json.Unmarshal(data, &shipments)

	jobs := make(chan Shipment, len(shipments))
	results := make(chan Report, len(shipments))

	for i := 0; i < 4; i++ {
		go worker(jobs, results)
	}

	for _, s := range shipments {
		jobs <- s
	}
	close(jobs)

	for i := 0; i < len(shipments); i++ {
		r := <-results
		out, _ := json.MarshalIndent(r, "", " ")
		fmt.Println(string(out))
	}
}

func worker(jobs <-chan Shipment, results chan<- Report) {
	for j := range jobs {
		results <- processShipment(j)
	}
}

func processShipment(s Shipment) Report {

	report := Report{ContainerID: s.ContainerID}

	sort.Slice(s.Events, func(i, j int) bool {
		t1, _ := time.Parse(time.RFC3339, s.Events[i].Timestamp)
		t2, _ := time.Parse(time.RFC3339, s.Events[j].Timestamp)
		return t1.Before(t2)
	})

	var prevTime time.Time
	var prevEvent string

	for _, e := range s.Events {

		if err := validateEvent(s.ContainerID, e); err != nil {
			report.Anomalies = append(report.Anomalies,
				Anomaly{"validation_error", err.Error()})
			continue
		}

		currTime, _ := time.Parse(time.RFC3339, e.Timestamp)

		// unusual gap
		if !prevTime.IsZero() && currTime.Sub(prevTime).Hours() > 24 {
			report.Anomalies = append(report.Anomalies,
				Anomaly{"unusual_gap", "gap > 24 hours"})
		}

		// duplicate
		if prevEvent == e.EventType &&
			currTime.Sub(prevTime).Minutes() < 60 {
			report.Anomalies = append(report.Anomalies,
				Anomaly{"duplicate_event", "same event within 1 hour"})
		}

		// out of sequence
		if prevEvent != "" &&
			eventOrder[e.EventType] < eventOrder[prevEvent] {
			report.Anomalies = append(report.Anomalies,
				Anomaly{"out_of_sequence", "incorrect event order"})
		}

		delayMinutes := 0

		if e.Metadata != nil && e.Metadata["expected_arrival"] != nil {

			expTime, err := time.Parse(
				time.RFC3339,
				e.Metadata["expected_arrival"].(string),
			)

			if err == nil {
				delayMinutes = int(currTime.Sub(expTime).Minutes())

				if delayMinutes > 120 {
					report.Anomalies = append(report.Anomalies,
						Anomaly{
							"late_arrival",
							fmt.Sprintf("Container arrived %d minutes late", delayMinutes),
						})
				}
			}
		}

		report.Timeline = append(report.Timeline,
			TimelineEvent{
				EventType:    e.EventType,
				Timestamp:    e.Timestamp,
				Location:     e.Location,
				DelayMinutes: delayMinutes,
			})

		if mapped, ok := statusMap[e.EventType]; ok {
			report.CurrentStatus = mapped
		} else {
			report.CurrentStatus = e.EventType
		}
		report.CurrentLocation = e.Location
		report.LastEventTime = currTime

		prevEvent = e.EventType
		prevTime = currTime
	}

	totalStages := len(eventOrder)
	if totalStages > 0 {
		report.JourneyProgress =
			(len(report.Timeline) * 100) / totalStages
	}

	return report
}

func validateEvent(containerID string, e Event) error {

	if containerID == "" {
		return fmt.Errorf("container id is empty")
	}

	if _, ok := eventOrder[e.EventType]; !ok {
		return fmt.Errorf("invalid event type %s", e.EventType)
	}

	if _, err := time.Parse(time.RFC3339, e.Timestamp); err != nil {
		return fmt.Errorf("invalid timestamp")
	}

	switch e.EventType {

	case "port_arrival":
		if e.Metadata == nil || e.Metadata["expected_arrival"] == nil {
			return fmt.Errorf("expected_arrival missing")
		}

	case "customs_clearance":
		if e.Metadata == nil || e.Metadata["clearance_status"] == nil {
			return fmt.Errorf("clearance_status missing")
		}

	case "road_checkpoint":
		if e.Metadata == nil || e.Metadata["checkpoint_id"] == nil {
			return fmt.Errorf("checkpoint_id missing")
		}

	case "transshipment_arrival":
		if e.Metadata == nil || e.Metadata["next_vessel"] == nil {
			return fmt.Errorf("next_vessel missing")
		}
	}

	return nil
}
