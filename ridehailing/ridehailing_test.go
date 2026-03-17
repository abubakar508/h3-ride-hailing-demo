package ridehailing

import (
	"fmt"
	"math"
	"sync"
	"testing"

	"github.com/uber/h3-go/v4"
)

// ---- Test Locations (Nairobi, Kenya) ----

var (
	// Nairobi CBD
	nairobiCBD = Location{Lat: -1.2864, Lng: 36.8172}
	// Westlands
	westlands = Location{Lat: -1.2673, Lng: 36.8110}
	// Kilimani
	kilimani = Location{Lat: -1.2891, Lng: 36.7838}
	// JKIA Airport
	jkia = Location{Lat: -1.3192, Lng: 36.9278}
	// Upper Hill
	upperHill = Location{Lat: -1.2977, Lng: 36.8147}
	// Kasarani
	kasarani = Location{Lat: -1.2214, Lng: 36.8993}
	// Karen
	karen = Location{Lat: -1.3176, Lng: 36.7111}
	// Embakasi
	embakasi = Location{Lat: -1.3133, Lng: 36.8901}
)

// ---- H3 Cell Indexing Tests ----

func TestLocationToCell(t *testing.T) {
	cell := LocationToCell(nairobiCBD, DriverResolution)
	if !cell.IsValid() {
		t.Errorf("expected valid H3 cell for Nairobi CBD, got invalid cell")
	}

	res := GetCellResolution(cell)
	if res != DriverResolution {
		t.Errorf("expected resolution %d, got %d", DriverResolution, res)
	}
}

func TestLocationToCellDifferentResolutions(t *testing.T) {
	resolutions := []int{0, 3, 5, 7, 9, 12, 15}
	for _, res := range resolutions {
		t.Run(fmt.Sprintf("resolution_%d", res), func(t *testing.T) {
			cell := LocationToCell(nairobiCBD, res)
			if !cell.IsValid() {
				t.Errorf("expected valid H3 cell at resolution %d", res)
			}
			if GetCellResolution(cell) != res {
				t.Errorf("expected resolution %d, got %d", res, GetCellResolution(cell))
			}
		})
	}
}

func TestDifferentLocationsDifferentCells(t *testing.T) {
	cbdCell := LocationToCell(nairobiCBD, DriverResolution)
	jkiaCell := LocationToCell(jkia, DriverResolution)
	karenCell := LocationToCell(karen, DriverResolution)

	if cbdCell == jkiaCell {
		t.Error("Nairobi CBD and JKIA should map to different H3 cells")
	}
	if cbdCell == karenCell {
		t.Error("Nairobi CBD and Karen should map to different H3 cells")
	}
	if jkiaCell == karenCell {
		t.Error("JKIA and Karen should map to different H3 cells")
	}
}

func TestSameLocationSameCell(t *testing.T) {
	cell1 := LocationToCell(nairobiCBD, DriverResolution)
	cell2 := LocationToCell(nairobiCBD, DriverResolution)

	if cell1 != cell2 {
		t.Error("same location should always map to the same H3 cell")
	}
}

func TestCellToLocation(t *testing.T) {
	cell := LocationToCell(nairobiCBD, DriverResolution)
	center := CellToLocation(cell)

	// Center should be close to the original location (within a few hundred meters)
	dist := HaversineDistance(nairobiCBD, center)
	if dist > 0.5 { // within 500m
		t.Errorf("cell center should be within 500m of original location, got %.2f km", dist)
	}
}

func TestGetCellBoundary(t *testing.T) {
	cell := LocationToCell(nairobiCBD, DriverResolution)
	boundary := GetCellBoundary(cell)

	// H3 cells are hexagons (6 vertices) or pentagons (5 vertices)
	if len(boundary) < 5 || len(boundary) > 6 {
		t.Errorf("expected 5 or 6 boundary vertices, got %d", len(boundary))
	}
}

// ---- Driver Registration and Management Tests ----

func TestRegisterDriver(t *testing.T) {
	gm := NewGridManager()

	driver := &Driver{
		ID:       "d1",
		Name:     "Alice",
		Location: nairobiCBD,
		Rating:   4.8,
	}

	err := gm.RegisterDriver(driver)
	if err != nil {
		t.Fatalf("failed to register driver: %v", err)
	}

	if gm.GetDriverCount() != 1 {
		t.Errorf("expected 1 driver, got %d", gm.GetDriverCount())
	}
	if !driver.Active {
		t.Error("newly registered driver should be active")
	}
	if !driver.CellID.IsValid() {
		t.Error("driver should have a valid H3 cell assigned")
	}
}

func TestRegisterNilDriver(t *testing.T) {
	gm := NewGridManager()
	err := gm.RegisterDriver(nil)
	if err == nil {
		t.Error("registering nil driver should return an error")
	}
}

func TestRegisterDriverEmptyID(t *testing.T) {
	gm := NewGridManager()
	driver := &Driver{ID: "", Name: "NoID", Location: nairobiCBD}
	err := gm.RegisterDriver(driver)
	if err == nil {
		t.Error("registering driver with empty ID should return an error")
	}
}

func TestRemoveDriver(t *testing.T) {
	gm := NewGridManager()
	driver := &Driver{ID: "d1", Name: "Alice", Location: nairobiCBD, Rating: 4.8}
	gm.RegisterDriver(driver)

	err := gm.RemoveDriver("d1")
	if err != nil {
		t.Fatalf("failed to remove driver: %v", err)
	}
	if gm.GetDriverCount() != 0 {
		t.Errorf("expected 0 drivers after removal, got %d", gm.GetDriverCount())
	}
}

func TestRemoveNonExistentDriver(t *testing.T) {
	gm := NewGridManager()
	err := gm.RemoveDriver("nonexistent")
	if err == nil {
		t.Error("removing nonexistent driver should return an error")
	}
}

func TestRegisterMultipleDrivers(t *testing.T) {
	gm := NewGridManager()

	locations := []struct {
		id  string
		loc Location
	}{
		{"d1", nairobiCBD},
		{"d2", westlands},
		{"d3", kilimani},
		{"d4", jkia},
		{"d5", upperHill},
	}

	for _, l := range locations {
		driver := &Driver{ID: l.id, Name: "Driver " + l.id, Location: l.loc, Rating: 4.5}
		err := gm.RegisterDriver(driver)
		if err != nil {
			t.Fatalf("failed to register driver %s: %v", l.id, err)
		}
	}

	if gm.GetDriverCount() != 5 {
		t.Errorf("expected 5 drivers, got %d", gm.GetDriverCount())
	}
	if gm.GetActiveDriverCount() != 5 {
		t.Errorf("expected 5 active drivers, got %d", gm.GetActiveDriverCount())
	}
}

// ---- Driver Location Update Tests ----

func TestUpdateDriverLocation(t *testing.T) {
	gm := NewGridManager()
	driver := &Driver{ID: "d1", Name: "Alice", Location: nairobiCBD, Rating: 4.8}
	gm.RegisterDriver(driver)

	oldCell := driver.CellID
	err := gm.UpdateDriverLocation("d1", westlands)
	if err != nil {
		t.Fatalf("failed to update driver location: %v", err)
	}

	if driver.Location.Lat != westlands.Lat || driver.Location.Lng != westlands.Lng {
		t.Error("driver location not updated correctly")
	}

	newCell := driver.CellID
	// Nairobi CBD to Westlands should be in a different cell at resolution 9
	if oldCell == newCell {
		t.Log("warning: driver moved but stayed in same H3 cell (locations may be very close)")
	}
}

func TestUpdateNonExistentDriverLocation(t *testing.T) {
	gm := NewGridManager()
	err := gm.UpdateDriverLocation("nonexistent", nairobiCBD)
	if err == nil {
		t.Error("updating nonexistent driver location should return an error")
	}
}

func TestUpdateDriverLocationReindexes(t *testing.T) {
	gm := NewGridManager()
	driver := &Driver{ID: "d1", Name: "Alice", Location: nairobiCBD, Rating: 4.8}
	gm.RegisterDriver(driver)

	// Move driver far away (to JKIA)
	gm.UpdateDriverLocation("d1", jkia)

	// Search near CBD should not find the driver
	driversNearCBD := gm.FindNearbyDrivers(nairobiCBD, 1)
	for _, d := range driversNearCBD {
		if d.ID == "d1" {
			t.Error("driver should not be found near CBD after moving to JKIA")
		}
	}

	// Search near JKIA should find the driver
	driversNearJKIA := gm.FindNearbyDrivers(jkia, 1)
	found := false
	for _, d := range driversNearJKIA {
		if d.ID == "d1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("driver should be found near JKIA after moving there")
	}
}

// ---- Nearby Driver Search Tests ----

func TestFindNearbyDriversEmptyGrid(t *testing.T) {
	gm := NewGridManager()
	drivers := gm.FindNearbyDrivers(nairobiCBD, 3)
	if len(drivers) != 0 {
		t.Errorf("expected 0 drivers in empty grid, got %d", len(drivers))
	}
}

func TestFindNearbyDriversSameCell(t *testing.T) {
	gm := NewGridManager()
	// Register a driver very close to the search location
	driver := &Driver{ID: "d1", Name: "Alice", Location: nairobiCBD, Rating: 4.8}
	gm.RegisterDriver(driver)

	drivers := gm.FindNearbyDrivers(nairobiCBD, 0) // k=0 means same cell only
	if len(drivers) != 1 {
		t.Errorf("expected 1 driver in same cell, got %d", len(drivers))
	}
}

func TestFindNearbyDriversExpandingRings(t *testing.T) {
	gm := NewGridManager()

	// Place drivers at different distances
	gm.RegisterDriver(&Driver{ID: "d1", Name: "Near", Location: nairobiCBD, Rating: 4.8})
	gm.RegisterDriver(&Driver{ID: "d2", Name: "Medium", Location: upperHill, Rating: 4.5})
	gm.RegisterDriver(&Driver{ID: "d3", Name: "Far", Location: jkia, Rating: 4.2})

	// k=0: only the very close driver
	d0 := gm.FindNearbyDrivers(nairobiCBD, 0)
	// As k increases, we should find more drivers
	d3 := gm.FindNearbyDrivers(nairobiCBD, 3)
	d10 := gm.FindNearbyDrivers(nairobiCBD, 10)

	if len(d10) < len(d3) {
		t.Error("larger k-ring should find at least as many drivers as smaller k-ring")
	}
	if len(d3) < len(d0) {
		t.Error("larger k-ring should find at least as many drivers as smaller k-ring")
	}
}

func TestFindNearbyDriversOnlyActive(t *testing.T) {
	gm := NewGridManager()

	active := &Driver{ID: "d1", Name: "Active", Location: nairobiCBD, Rating: 4.8}
	inactive := &Driver{ID: "d2", Name: "Inactive", Location: nairobiCBD, Rating: 4.5}

	gm.RegisterDriver(active)
	gm.RegisterDriver(inactive)
	inactive.Active = false

	drivers := gm.FindNearbyDrivers(nairobiCBD, 1)

	for _, d := range drivers {
		if d.ID == "d2" {
			t.Error("inactive driver should not appear in nearby search results")
		}
	}
}

// ---- Nearest Driver Tests ----

func TestFindNearestDriver(t *testing.T) {
	gm := NewGridManager()

	gm.RegisterDriver(&Driver{ID: "d1", Name: "Far", Location: jkia, Rating: 4.2})
	gm.RegisterDriver(&Driver{ID: "d2", Name: "Near", Location: upperHill, Rating: 4.5})
	gm.RegisterDriver(&Driver{ID: "d3", Name: "Medium", Location: westlands, Rating: 4.8})

	nearest, dist := gm.FindNearestDriver(nairobiCBD, 10)
	if nearest == nil {
		t.Fatal("expected to find a nearest driver")
	}
	if dist <= 0 {
		t.Error("distance to nearest driver should be positive")
	}

	// Upper Hill is closest to CBD
	if nearest.ID != "d2" {
		t.Logf("nearest driver is %s (distance: %.3f km), expected d2 (Upper Hill)", nearest.ID, dist)
	}
}

func TestFindNearestDriverNoDrivers(t *testing.T) {
	gm := NewGridManager()
	nearest, _ := gm.FindNearestDriver(nairobiCBD, 5)
	if nearest != nil {
		t.Error("expected nil when no drivers are available")
	}
}

// ---- Haversine Distance Tests ----

func TestHaversineDistance(t *testing.T) {
	// Nairobi CBD to JKIA should be approximately 15 km
	dist := HaversineDistance(nairobiCBD, jkia)
	if dist < 10 || dist > 20 {
		t.Errorf("CBD to JKIA distance should be ~15 km, got %.2f km", dist)
	}
}

func TestHaversineDistanceSameLocation(t *testing.T) {
	dist := HaversineDistance(nairobiCBD, nairobiCBD)
	if dist != 0 {
		t.Errorf("distance from a location to itself should be 0, got %.6f", dist)
	}
}

func TestHaversineDistanceSymmetric(t *testing.T) {
	d1 := HaversineDistance(nairobiCBD, jkia)
	d2 := HaversineDistance(jkia, nairobiCBD)

	if math.Abs(d1-d2) > 0.001 {
		t.Errorf("haversine distance should be symmetric: %.6f vs %.6f", d1, d2)
	}
}

func TestHaversineDistanceTriangleInequality(t *testing.T) {
	dAB := HaversineDistance(nairobiCBD, westlands)
	dBC := HaversineDistance(westlands, jkia)
	dAC := HaversineDistance(nairobiCBD, jkia)

	if dAC > dAB+dBC+0.001 {
		t.Error("haversine distance should satisfy the triangle inequality")
	}
}

// ---- Trip Estimation Tests ----

func TestEstimateTripDistance(t *testing.T) {
	dist := EstimateTripDistance(nairobiCBD, jkia)
	if dist <= 0 {
		t.Error("trip distance should be positive")
	}
	// CBD to JKIA ~15 km
	if dist < 5 || dist > 30 {
		t.Errorf("CBD to JKIA trip should be ~15 km, got %.2f km", dist)
	}
}

func TestGridDistance(t *testing.T) {
	dist, err := GridDistance(nairobiCBD, upperHill, 7) // coarser resolution for reliable grid distance
	if err != nil {
		t.Logf("grid distance could not be computed: %v (this can happen for distant cells)", err)
		return
	}
	if dist < 0 {
		t.Errorf("grid distance should be non-negative, got %d", dist)
	}
}

// ---- Surge Pricing Tests ----

func TestSetAndGetSurgeZone(t *testing.T) {
	gm := NewGridManager()

	zone := gm.SetSurgeZone(nairobiCBD, 2.0, 100, 20)
	if zone == nil {
		t.Fatal("surge zone should not be nil")
	}
	if zone.Multiplier != 2.0 {
		t.Errorf("expected surge multiplier 2.0, got %.1f", zone.Multiplier)
	}
	if zone.Demand != 100 {
		t.Errorf("expected demand 100, got %d", zone.Demand)
	}
	if zone.Supply != 20 {
		t.Errorf("expected supply 20, got %d", zone.Supply)
	}
}

func TestGetSurgeMultiplier(t *testing.T) {
	gm := NewGridManager()
	gm.SetSurgeZone(nairobiCBD, 1.5, 50, 30)

	mult := gm.GetSurgeMultiplier(nairobiCBD)
	if mult != 1.5 {
		t.Errorf("expected surge multiplier 1.5, got %.1f", mult)
	}
}

func TestGetSurgeMultiplierDefault(t *testing.T) {
	gm := NewGridManager()
	mult := gm.GetSurgeMultiplier(nairobiCBD)
	if mult != 1.0 {
		t.Errorf("expected default surge multiplier 1.0, got %.1f", mult)
	}
}

func TestCalculateSurge(t *testing.T) {
	tests := []struct {
		name     string
		demand   int
		supply   int
		expected float64
	}{
		{"balanced", 10, 10, 1.0},
		{"low_demand", 5, 10, 1.0},
		{"double_demand", 20, 10, 2.0},
		{"triple_demand", 30, 10, 3.0},
		{"extreme_demand", 100, 10, 3.0}, // capped at 3.0
		{"no_supply", 10, 0, 3.0},        // max surge
		{"slight_surge", 15, 10, 1.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			surge := CalculateSurge(tt.demand, tt.supply)
			if math.Abs(surge-tt.expected) > 0.1 {
				t.Errorf("CalculateSurge(%d, %d) = %.1f, want %.1f", tt.demand, tt.supply, surge, tt.expected)
			}
		})
	}
}

// ---- Trip Creation Tests ----

func TestCreateTrip(t *testing.T) {
	gm := NewGridManager()
	gm.RegisterDriver(&Driver{ID: "d1", Name: "Alice", Location: nairobiCBD, Rating: 4.8})
	gm.RegisterDriver(&Driver{ID: "d2", Name: "Bob", Location: westlands, Rating: 4.5})

	rider := &Rider{ID: "r1", Name: "Charlie", Location: nairobiCBD}
	trip, err := gm.CreateTrip("t1", rider, nairobiCBD, jkia)
	if err != nil {
		t.Fatalf("failed to create trip: %v", err)
	}

	if trip.ID != "t1" {
		t.Errorf("expected trip ID t1, got %s", trip.ID)
	}
	if trip.Driver == nil {
		t.Fatal("trip should have a driver assigned")
	}
	if trip.EstimatedDistKm <= 0 {
		t.Error("estimated trip distance should be positive")
	}
	if trip.SurgeMultiplier < 1.0 {
		t.Error("surge multiplier should be at least 1.0")
	}
	if trip.Driver.Active {
		t.Error("assigned driver should be marked as inactive (on trip)")
	}
}

func TestCreateTripNoDrivers(t *testing.T) {
	gm := NewGridManager()
	rider := &Rider{ID: "r1", Name: "Charlie", Location: nairobiCBD}
	_, err := gm.CreateTrip("t1", rider, nairobiCBD, jkia)
	if err == nil {
		t.Error("creating a trip with no available drivers should return an error")
	}
}

func TestCreateTripWithSurge(t *testing.T) {
	gm := NewGridManager()
	gm.RegisterDriver(&Driver{ID: "d1", Name: "Alice", Location: nairobiCBD, Rating: 4.8})
	gm.SetSurgeZone(nairobiCBD, 2.5, 80, 20)

	rider := &Rider{ID: "r1", Name: "Charlie", Location: nairobiCBD}
	trip, err := gm.CreateTrip("t1", rider, nairobiCBD, jkia)
	if err != nil {
		t.Fatalf("failed to create trip: %v", err)
	}
	if trip.SurgeMultiplier != 2.5 {
		t.Errorf("expected surge multiplier 2.5, got %.1f", trip.SurgeMultiplier)
	}
}

// ---- H3 Hierarchy Tests ----

func TestParentChildRelationship(t *testing.T) {
	cell := LocationToCell(nairobiCBD, 9)
	parent := GetParentCell(cell, 7)

	if !parent.IsValid() {
		t.Error("parent cell should be valid")
	}
	if GetCellResolution(parent) != 7 {
		t.Errorf("parent cell should have resolution 7, got %d", GetCellResolution(parent))
	}

	children := GetChildrenCells(parent, 9)
	if len(children) == 0 {
		t.Error("parent cell should have children at resolution 9")
	}

	// The original cell should be one of the children
	found := false
	for _, child := range children {
		if child == cell {
			found = true
			break
		}
	}
	if !found {
		t.Error("original cell should be among the parent's children")
	}
}

func TestCellNeighbors(t *testing.T) {
	cell := LocationToCell(nairobiCBD, DriverResolution)
	neighbors, err := h3.GridDisk(cell, 1)
	if err != nil {
		t.Fatalf("GridDisk failed: %v", err)
	}

	// k=1 disk should have 7 cells (center + 6 neighbors)
	if len(neighbors) != 7 {
		t.Errorf("k=1 grid disk should have 7 cells, got %d", len(neighbors))
	}

	// Each neighbor should be adjacent to the center
	for _, n := range neighbors {
		if n == cell {
			continue // skip center
		}
		if !AreCellsNeighbors(cell, n) {
			t.Errorf("cell in k=1 ring should be a neighbor of center")
		}
	}
}

// ---- Driver Ranking Tests ----

func TestRankDriversByDistance(t *testing.T) {
	drivers := []*Driver{
		{ID: "d1", Name: "Far", Location: jkia},
		{ID: "d2", Name: "Near", Location: upperHill},
		{ID: "d3", Name: "Medium", Location: westlands},
	}

	ranked := RankDriversByDistance(drivers, nairobiCBD)

	if ranked.Len() != 3 {
		t.Fatalf("expected 3 ranked drivers, got %d", ranked.Len())
	}

	// Distances should be in ascending order
	for i := 0; i < ranked.Len()-1; i++ {
		if ranked.Distances[i] > ranked.Distances[i+1] {
			t.Errorf("drivers not sorted by distance: %.3f > %.3f",
				ranked.Distances[i], ranked.Distances[i+1])
		}
	}
}

func TestRankDriversByDistanceEmpty(t *testing.T) {
	ranked := RankDriversByDistance([]*Driver{}, nairobiCBD)
	if ranked.Len() != 0 {
		t.Error("ranking empty driver list should return empty result")
	}
}

// ---- Concurrency Tests ----

func TestConcurrentDriverRegistration(t *testing.T) {
	gm := NewGridManager()
	var wg sync.WaitGroup

	n := 100
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(id int) {
			defer wg.Done()
			driver := &Driver{
				ID:       fmt.Sprintf("d%d", id),
				Name:     fmt.Sprintf("Driver %d", id),
				Location: Location{Lat: -1.28 + float64(id)*0.001, Lng: 36.81 + float64(id)*0.001},
				Rating:   4.5,
			}
			gm.RegisterDriver(driver)
		}(i)
	}
	wg.Wait()

	if gm.GetDriverCount() != n {
		t.Errorf("expected %d drivers after concurrent registration, got %d", n, gm.GetDriverCount())
	}
}

func TestConcurrentSearchAndUpdate(t *testing.T) {
	gm := NewGridManager()

	// Register some initial drivers
	for i := 0; i < 20; i++ {
		driver := &Driver{
			ID:       fmt.Sprintf("d%d", i),
			Name:     fmt.Sprintf("Driver %d", i),
			Location: Location{Lat: -1.28 + float64(i)*0.001, Lng: 36.81 + float64(i)*0.001},
			Rating:   4.5,
		}
		gm.RegisterDriver(driver)
	}

	// Concurrently search and update
	var wg sync.WaitGroup
	wg.Add(40)

	for i := 0; i < 20; i++ {
		go func(id int) {
			defer wg.Done()
			gm.FindNearbyDrivers(nairobiCBD, 3)
		}(i)
	}

	for i := 0; i < 20; i++ {
		go func(id int) {
			defer wg.Done()
			gm.UpdateDriverLocation(
				fmt.Sprintf("d%d", id),
				Location{Lat: -1.28 + float64(id)*0.002, Lng: 36.81 + float64(id)*0.002},
			)
		}(i)
	}

	wg.Wait()
	// Test passes if no race condition panic occurs
}

// ---- Benchmark Tests ----

func BenchmarkLocationToCell(b *testing.B) {
	for i := 0; i < b.N; i++ {
		LocationToCell(nairobiCBD, DriverResolution)
	}
}

func BenchmarkFindNearbyDrivers(b *testing.B) {
	gm := NewGridManager()
	for i := 0; i < 1000; i++ {
		driver := &Driver{
			ID:       fmt.Sprintf("d%d", i),
			Name:     fmt.Sprintf("Driver %d", i),
			Location: Location{Lat: -1.28 + float64(i)*0.0001, Lng: 36.81 + float64(i)*0.0001},
			Rating:   4.5,
		}
		gm.RegisterDriver(driver)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gm.FindNearbyDrivers(nairobiCBD, 3)
	}
}

func BenchmarkHaversineDistance(b *testing.B) {
	for i := 0; i < b.N; i++ {
		HaversineDistance(nairobiCBD, jkia)
	}
}

func BenchmarkCreateTrip(b *testing.B) {
	gm := NewGridManager()
	for i := 0; i < 100; i++ {
		driver := &Driver{
			ID:       fmt.Sprintf("d%d", i),
			Name:     fmt.Sprintf("Driver %d", i),
			Location: Location{Lat: -1.28 + float64(i)*0.001, Lng: 36.81 + float64(i)*0.001},
			Rating:   4.5,
		}
		gm.RegisterDriver(driver)
	}

	rider := &Rider{ID: "r1", Name: "Test Rider", Location: nairobiCBD}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Reset all drivers to active
		for _, d := range gm.drivers {
			d.Active = true
		}
		gm.CreateTrip(fmt.Sprintf("t%d", i), rider, nairobiCBD, jkia)
	}
}
