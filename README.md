# H3 Ride-Hailing Demo

A Go project demonstrating [Uber's H3](https://h3geo.org/) hexagonal spatial indexing system for a ride-hailing application. This project uses H3 to efficiently index driver locations, find nearby drivers, estimate trip distances, and manage surge pricing zones.

## Features

- **H3 Spatial Indexing**: Map geographic coordinates to H3 hexagonal cells for efficient spatial queries
- **Driver Management**: Register, update, and remove drivers with automatic H3 cell indexing
- **Nearby Driver Search**: Find available drivers using H3 k-ring expansion
- **Nearest Driver Matching**: Match riders with the closest available driver
- **Trip Estimation**: Estimate trip distances using Haversine formula
- **Surge Pricing**: Zone-based surge pricing using coarser H3 resolution
- **H3 Hierarchy**: Parent/child cell relationships for multi-resolution analysis
- **Concurrency Safe**: Thread-safe operations with mutex-based synchronization

## Prerequisites

- Go 1.21+
- GCC (for CGO, required by h3-go)

## Getting Started

```bash
# Clone the repo
git clone https://github.com/abubakar508/h3-ride-hailing-demo.git
cd h3-ride-hailing-demo

# Install dependencies
go mod tidy

# Run all tests
go test -v ./...

# Run benchmarks
go test -bench=. -benchmem ./...

# Run with race detector
go test -race -v ./...
```

## Project Structure

```
h3-ride-hailing-demo/
├── go.mod
├── go.sum
├── README.md
└── ridehailing/
    ├── ridehailing.go       # Core ride-hailing H3 functionality
    └── ridehailing_test.go  # Comprehensive test suite
```

## Test Coverage

The test suite covers:

| Category | Tests |
|---|---|
| H3 Cell Indexing | Location-to-cell mapping, resolution handling, boundary extraction |
| Driver Management | Registration, removal, location updates, re-indexing |
| Nearby Search | Empty grid, same cell, expanding rings, active-only filtering |
| Nearest Driver | Closest driver matching, no-driver edge case |
| Distance Calculations | Haversine accuracy, symmetry, triangle inequality |
| Trip Operations | Trip creation, driver assignment, surge integration |
| Surge Pricing | Zone management, multiplier calculation, demand/supply ratios |
| H3 Hierarchy | Parent/child relationships, cell neighbors |
| Concurrency | Concurrent registration, concurrent search + update |
| Benchmarks | Cell conversion, nearby search, distance calc, trip creation |

## H3 Resolutions Used

| Resolution | Hex Area | Use Case |
|---|---|---|
| 9 | ~0.1 km² | Driver location indexing |
| 7 | ~5 km² | Surge pricing zones |

## License

MIT
