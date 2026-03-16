# Container Event Processor (Go)

## Overview

This project implements a container tracking event processing engine for logistics visibility systems.
It reads shipment tracking events from a JSON input file, validates them, detects anomalies in container movement, and generates a structured summary report for each container.

The solution is designed to simulate real-world backend processing logic used in supply chain tracking platforms.



## Features

### Event Validation

Each event is validated before processing:

* Container ID must not be empty
* Event type must be one of the supported logistics events
* Timestamp must follow ISO-8601 (RFC3339) format
* Required metadata fields must be present based on event type

Validation errors are captured and reported as anomalies without stopping processing.

### Container Status Tracking

For every container, the processor maintains:

* Current shipment status (mapped from system event to business status)
* Current container location
* Last processed event timestamp
* Chronological timeline of all valid events
* Delay calculation based on expected vs actual arrival

### Anomaly Detection

The system flags the following anomalies:

* **Late Arrival** – Container arrives more than 2 hours after expected arrival
* **Unusual Gap** – More than 24 hours between consecutive events
* **Out of Sequence** – Events occurring in an incorrect logical order
* **Duplicate Event** – Same event type occurring within 1 hour

### Performance Considerations

* Worker pool concurrency implemented for parallel shipment processing
* Efficient in-memory structures used for fast container status lookup
* Suitable for processing large batches (e.g., 10,000 events)



## Supported Event Types

* port_arrival
* port_departure
* customs_clearance
* customs_hold
* road_checkpoint
* lcl_pickup
* transshipment_arrival
* in_transit



## How to Run

### 1. Prepare Input File

Place the shipment JSON file in the same directory as the program.

Example file name:


20_shipments_detailed_input_wetrack_developer_assignment.json


### 2. Run Application

go run container_processor.go


The program will print a structured JSON summary report for each container.



## How to Run Tests
go test


Test cases include:

* Happy path shipment flow
* Event validation failure
* Anomaly detection scenarios


## Design Decisions

* Used Go standard library only to keep the solution lightweight and portable
* Implemented worker pool to simulate scalable batch event processing
* Introduced event-to-business status mapping layer for clearer domain representation
* Chronological sorting ensures correct anomaly detection and timeline accuracy
* Graceful error handling prevents invalid events from crashing the system


## Assumptions

* Input timestamps are provided in UTC RFC3339 format
* Expected arrival metadata is relevant mainly for port arrival events
* Journey progress is approximated based on number of processed event stages


## Limitations

* No persistent storage (in-memory processing only)
* Duplicate detection considers only consecutive event comparison window
* Shipment route prediction and ETA forecasting not implemented
* Input currently read from local JSON file instead of API or streaming source


## Future Improvements

* REST API or streaming ingestion support
* Persistent database integration
* Configurable anomaly detection thresholds

