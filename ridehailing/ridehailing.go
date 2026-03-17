// Package ridehailing demonstrates H3 hexagonal spatial indexing
// for a ride-hailing application. It provides functionality for
// mapping drivers and riders to H3 cells, finding nearby drivers,
// computing trip distances, and managing surge pricing zones.
package ridehailing

import (
	"fmt"
	"math"
	"sort"
	"sync"

	"github.com/uber/h3-go/v4"
)

// Resolution constants for different use cases.
const (
	// DriverResolution is the H3 resolution used for driver indexing.
	// Resolution 9 corresponds to ~0.1 km^2 hexagons.
	DriverResolution = 9

	// SurgeResolution is the H3 resolution used for surge pricing zones.
	// Resolution 7 corresponds to ~5 km^2 hexagons.
	SurgeResolution = 7

	// EarthRadiusKm is the mean radius of the Earth in kilometers.
	EarthRadiusKm = 6371.0
)

// Location represents a geographic coordinate.
type Location struct {
	Lat float64
	Lng float64
}

// Driver represents a driver in the ride-hailing system.
type Driver struct {
	ID       string
	Name     string
	Location Location
	CellID   h3.Cell
	Rating   float64
	Active   bool
}

// Rider represents a rider requesting a ride.
type Rider struct {
	ID       string
	Name     string
	Location Location
	CellID   h3.Cell
}

// Trip represents a ride from pickup to dropoff.
type Trip struct {
	ID              string
	Driver          *Driver
	Rider           *Rider
	Pickup          Location
	Dropoff         Location
	PickupCell      h3.Cell
	DropoffCell     h3.Cell
	EstimatedDistKm float64
	SurgeMultiplier float64
}

// SurgeZone represents a surge pricing zone.
type SurgeZone struct {
	CellID     h3.Cell
	Multiplier float64
	Demand     int
	Supply     int
}

// GridManager manages the H3-based spatial index for the ride-hailing system.
type GridManager struct {
	mu         sync.RWMutex
	drivers    map[string]*Driver
	cellIndex  map[h3.Cell][]*Driver
	surgeZones map[h3.Cell]*SurgeZone
}

// NewGridManager creates a new GridManager instance.
func NewGridManager() *GridManager {
	return &GridManager{
		drivers:    make(map[string]*Driver),
		cellIndex:  make(map[h3.Cell][]*Driver),
		surgeZones: make(map[h3.Cell]*SurgeZone),
	}
}

// LocationToCell converts a location to an H3 cell at the given resolution.
func LocationToCell(loc Location, resolution int) h3.Cell {
	latLng := h3.NewLatLng(loc.Lat, loc.Lng)
	cell, _ := h3.LatLngToCell(latLng, resolution)
	return cell
}

// CellToLocation returns the center of an H3 cell as a Location.
func CellToLocation(cell h3.Cell) Location {
	latLng, _ := h3.CellToLatLng(cell)
	return Location{Lat: latLng.Lat, Lng: latLng.Lng}
}

// RegisterDriver adds a driver to the grid manager.
func (gm *GridManager) RegisterDriver(driver *Driver) error {
	if driver == nil {
		return fmt.Errorf("driver cannot be nil")
	}
	if driver.ID == "" {
		return fmt.Errorf("driver ID cannot be empty")
	}

	gm.mu.Lock()
	defer gm.mu.Unlock()

	cell := LocationToCell(driver.Location, DriverResolution)
	driver.CellID = cell
	driver.Active = true

	gm.drivers[driver.ID] = driver
	gm.cellIndex[cell] = append(gm.cellIndex[cell], driver)

	return nil
}

// RemoveDriver removes a driver from the grid manager.
func (gm *GridManager) RemoveDriver(driverID string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	driver, ok := gm.drivers[driverID]
	if !ok {
		return fmt.Errorf("driver %s not found", driverID)
	}

	// Remove from cell index
	cell := driver.CellID
	drivers := gm.cellIndex[cell]
	for i, d := range drivers {
		if d.ID == driverID {
			gm.cellIndex[cell] = append(drivers[:i], drivers[i+1:]...)
			break
		}
	}

	delete(gm.drivers, driverID)
	return nil
}

// UpdateDriverLocation updates a driver's position and re-indexes them.
func (gm *GridManager) UpdateDriverLocation(driverID string, newLoc Location) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	driver, ok := gm.drivers[driverID]
	if !ok {
		return fmt.Errorf("driver %s not found", driverID)
	}

	oldCell := driver.CellID
	newCell := LocationToCell(newLoc, DriverResolution)

	driver.Location = newLoc
	driver.CellID = newCell

	// If driver moved to a new cell, update the index
	if oldCell != newCell {
		// Remove from old cell
		oldDrivers := gm.cellIndex[oldCell]
		for i, d := range oldDrivers {
			if d.ID == driverID {
				gm.cellIndex[oldCell] = append(oldDrivers[:i], oldDrivers[i+1:]...)
				break
			}
		}
		// Add to new cell
		gm.cellIndex[newCell] = append(gm.cellIndex[newCell], driver)
	}

	return nil
}

// FindNearbyDrivers returns drivers within k-ring distance of the given location.
// The kRing parameter controls how many rings of hexagons to search.
func (gm *GridManager) FindNearbyDrivers(loc Location, kRing int) []*Driver {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	cell := LocationToCell(loc, DriverResolution)
	neighbors, _ := h3.GridDisk(cell, kRing)

	var result []*Driver
	for _, neighbor := range neighbors {
		if drivers, ok := gm.cellIndex[neighbor]; ok {
			for _, d := range drivers {
				if d.Active {
					result = append(result, d)
				}
			}
		}
	}

	return result
}

// FindNearestDriver finds the closest active driver to the given location.
func (gm *GridManager) FindNearestDriver(loc Location, maxKRing int) (*Driver, float64) {
	for k := 1; k <= maxKRing; k++ {
		drivers := gm.FindNearbyDrivers(loc, k)
		if len(drivers) == 0 {
			continue
		}

		var nearest *Driver
		minDist := math.MaxFloat64

		for _, d := range drivers {
			dist := HaversineDistance(loc, d.Location)
			if dist < minDist {
				minDist = dist
				nearest = d
			}
		}

		if nearest != nil {
			return nearest, minDist
		}
	}
	return nil, 0
}

// HaversineDistance calculates the great-circle distance between two locations in km.
func HaversineDistance(a, b Location) float64 {
	lat1 := a.Lat * math.Pi / 180
	lat2 := b.Lat * math.Pi / 180
	dLat := (b.Lat - a.Lat) * math.Pi / 180
	dLng := (b.Lng - a.Lng) * math.Pi / 180

	h := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(h), math.Sqrt(1-h))

	return EarthRadiusKm * c
}

// EstimateTripDistance estimates the distance of a trip using H3 grid distance
// and haversine calculation.
func EstimateTripDistance(pickup, dropoff Location) float64 {
	return HaversineDistance(pickup, dropoff)
}

// GridDistance returns the H3 grid distance between two locations at a given resolution.
func GridDistance(a, b Location, resolution int) (int, error) {
	cellA := LocationToCell(a, resolution)
	cellB := LocationToCell(b, resolution)
	dist, err := h3.GridDistance(cellA, cellB)
	if err != nil {
		return 0, fmt.Errorf("grid distance could not be computed: %w", err)
	}
	return int(dist), nil
}

// SetSurgeZone sets or updates a surge pricing zone.
func (gm *GridManager) SetSurgeZone(loc Location, multiplier float64, demand, supply int) *SurgeZone {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	cell := LocationToCell(loc, SurgeResolution)
	zone := &SurgeZone{
		CellID:     cell,
		Multiplier: multiplier,
		Demand:     demand,
		Supply:     supply,
	}
	gm.surgeZones[cell] = zone
	return zone
}

// GetSurgeMultiplier returns the surge multiplier for a given location.
// Returns 1.0 if no surge zone is set.
func (gm *GridManager) GetSurgeMultiplier(loc Location) float64 {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	cell := LocationToCell(loc, SurgeResolution)
	if zone, ok := gm.surgeZones[cell]; ok {
		return zone.Multiplier
	}
	return 1.0
}

// CalculateSurge computes the surge multiplier based on demand and supply.
func CalculateSurge(demand, supply int) float64 {
	if supply <= 0 {
		return 3.0 // max surge
	}
	ratio := float64(demand) / float64(supply)
	if ratio <= 1.0 {
		return 1.0
	}
	// Cap at 3x surge
	surge := math.Min(ratio, 3.0)
	// Round to nearest 0.1
	return math.Round(surge*10) / 10
}

// CreateTrip creates a new trip with estimated distance and surge pricing.
func (gm *GridManager) CreateTrip(tripID string, rider *Rider, pickup, dropoff Location) (*Trip, error) {
	driver, dist := gm.FindNearestDriver(pickup, 5)
	if driver == nil {
		return nil, fmt.Errorf("no available drivers nearby")
	}

	_ = dist // driver distance, could be used for ETA

	tripDist := EstimateTripDistance(pickup, dropoff)
	surge := gm.GetSurgeMultiplier(pickup)

	trip := &Trip{
		ID:              tripID,
		Driver:          driver,
		Rider:           rider,
		Pickup:          pickup,
		Dropoff:         dropoff,
		PickupCell:      LocationToCell(pickup, DriverResolution),
		DropoffCell:     LocationToCell(dropoff, DriverResolution),
		EstimatedDistKm: tripDist,
		SurgeMultiplier: surge,
	}

	// Mark driver as inactive (on a trip)
	driver.Active = false

	return trip, nil
}

// GetDriverCount returns the total number of registered drivers.
func (gm *GridManager) GetDriverCount() int {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	return len(gm.drivers)
}

// GetActiveDriverCount returns the number of active (available) drivers.
func (gm *GridManager) GetActiveDriverCount() int {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	count := 0
	for _, d := range gm.drivers {
		if d.Active {
			count++
		}
	}
	return count
}

// GetCellBoundary returns the boundary vertices of an H3 cell.
func GetCellBoundary(cell h3.Cell) []Location {
	boundary, _ := h3.CellToBoundary(cell)
	locs := make([]Location, len(boundary))
	for i, ll := range boundary {
		locs[i] = Location{Lat: ll.Lat, Lng: ll.Lng}
	}
	return locs
}

// GetCellResolution returns the resolution of an H3 cell.
func GetCellResolution(cell h3.Cell) int {
	return cell.Resolution()
}

// AreCellsNeighbors checks if two H3 cells are adjacent.
func AreCellsNeighbors(a, b h3.Cell) bool {
	result, _ := a.IsNeighbor(b)
	return result
}

// DriversByDistance sorts drivers by distance from a given location.
type DriversByDistance struct {
	Drivers   []*Driver
	Origin    Location
	Distances []float64
}

// RankDriversByDistance returns drivers sorted by distance from the origin.
func RankDriversByDistance(drivers []*Driver, origin Location) *DriversByDistance {
	dbd := &DriversByDistance{
		Drivers:   make([]*Driver, len(drivers)),
		Origin:    origin,
		Distances: make([]float64, len(drivers)),
	}
	copy(dbd.Drivers, drivers)
	for i, d := range dbd.Drivers {
		dbd.Distances[i] = HaversineDistance(origin, d.Location)
	}
	sort.Sort(dbd)
	return dbd
}

func (d *DriversByDistance) Len() int           { return len(d.Drivers) }
func (d *DriversByDistance) Swap(i, j int)      {
	d.Drivers[i], d.Drivers[j] = d.Drivers[j], d.Drivers[i]
	d.Distances[i], d.Distances[j] = d.Distances[j], d.Distances[i]
}
func (d *DriversByDistance) Less(i, j int) bool { return d.Distances[i] < d.Distances[j] }

// GetParentCell returns the parent cell at a coarser resolution.
func GetParentCell(cell h3.Cell, parentRes int) h3.Cell {
	parent, _ := cell.Parent(parentRes)
	return parent
}

// GetChildrenCells returns the children cells at a finer resolution.
func GetChildrenCells(cell h3.Cell, childRes int) []h3.Cell {
	children, _ := cell.Children(childRes)
	return children
}
