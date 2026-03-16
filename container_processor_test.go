package main

import (
	"testing"
)

func TestHappyPath(t *testing.T) {

	shipment := Shipment{
		ContainerID: "CONT001",
		Events: []Event{
			{
				EventType: "port_arrival",
				Timestamp: "2024-11-15T08:30:00Z",
				Location:  "Port Singapore",
				Metadata: map[string]interface{}{
					"expected_arrival": "2024-11-15T08:00:00Z",
				},
			},
			{
				EventType: "customs_clearance",
				Timestamp: "2024-11-15T10:00:00Z",
				Location:  "Customs Singapore",
				Metadata: map[string]interface{}{
					"clearance_status": "approved",
				},
			},
		},
	}

	report := processShipment(shipment)

	if report.CurrentStatus != "cleared" {
		t.Errorf("expected status cleared got %s", report.CurrentStatus)
	}

	if len(report.Timeline) != 2 {
		t.Errorf("expected timeline length 2")
	}

}

func TestValidationFailure(t *testing.T) {

	shipment := Shipment{
		ContainerID: "",
		Events: []Event{
			{
				EventType: "invalid_event",
				Timestamp: "bad-time",
				Location:  "Unknown",
			},
		},
	}

	report := processShipment(shipment)

	if len(report.Anomalies) == 0 {
		t.Errorf("expected validation anomaly")
	}

}

func TestAnomalyDetection_UnusualGap(t *testing.T) {

	shipment := Shipment{
		ContainerID: "CONT002",
		Events: []Event{
			{
				EventType: "port_arrival",
				Timestamp: "2024-11-10T08:00:00Z",
				Location:  "Shanghai",
				Metadata: map[string]interface{}{
					"expected_arrival": "2024-11-10T07:00:00Z",
				},
			},
			{
				EventType: "port_departure",
				Timestamp: "2024-11-12T10:00:00Z",
				Location:  "Shanghai",
				Metadata:  map[string]interface{}{},
			},
		},
	}

	report := processShipment(shipment)

	found := false

	for _, a := range report.Anomalies {
		if a.Type == "unusual_gap" {
			found = true
		}
	}

	if !found {
		t.Errorf("expected unusual gap anomaly")
	}

}
