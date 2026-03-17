// main.go provides an interactive CLI to demonstrate and test H3
// ride-hailing functionality: location indexing, distance calculations,
// nearby driver search, routing, surge pricing, and cell hierarchy.
package main

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	rh "github.com/abubakar508/h3-ride-hailing-demo/ridehailing"
)

func main() {
	fmt.Println("==========================================================")
	fmt.Println("  H3 Ride-Hailing Demo - Interactive Test Suite")
	fmt.Println("==========================================================")
	fmt.Println()

	gm := rh.NewGridManager()

	demoLocationIndexing()
	demoDistanceCalculations()
	demoDriverManagement(gm)
	demoNearbySearch(gm)
	demoRouting()
	demoSurgePricing(gm)
	demoCellHierarchy()
	demoCellBoundaries()
	demoTripCreation(gm)
	demoDriverRanking(gm)

	fmt.Println()
	fmt.Println("==========================================================")
	fmt.Println("  All demos completed successfully!")
	fmt.Println("==========================================================")
}

// ──────────────────────────────────────────────────────────────────────
// 1. Location Indexing
// ──────────────────────────────────────────────────────────────────────

func demoLocationIndexing() {
	printHeader("1. Location -> H3 Cell Indexing")

	locations := []struct {
		name string
		loc  rh.Location
	}{
		{"Nairobi CBD", rh.Location{Lat: -1.2921, Lng: 36.8219}},
		{"Westlands", rh.Location{Lat: -1.2673, Lng: 36.8114}},
		{"Jomo Kenyatta Airport (JKIA)", rh.Location{Lat: -1.3192, Lng: 36.9278}},
		{"Karen", rh.Location{Lat: -1.3187, Lng: 36.7114}},
		{"Times Square, NYC", rh.Location{Lat: 40.7580, Lng: -73.9855}},
		{"Eiffel Tower, Paris", rh.Location{Lat: 48.8584, Lng: 2.2945}},
	}

	for _, l := range locations {
		cellRes9 := rh.LocationToCell(l.loc, rh.DriverResolution)
		cellRes7 := rh.LocationToCell(l.loc, rh.SurgeResolution)
		center := rh.CellToLocation(cellRes9)

		fmt.Printf("  %-30s  (%.4f, %.4f)\n", l.name, l.loc.Lat, l.loc.Lng)
		fmt.Printf("    H3 Cell (res 9 / driver):  %v\n", cellRes9)
		fmt.Printf("    H3 Cell (res 7 / surge):   %v\n", cellRes7)
		fmt.Printf("    Cell center:               (%.6f, %.6f)\n", center.Lat, center.Lng)
		fmt.Printf("    Cell resolution:            %d\n", rh.GetCellResolution(cellRes9))
		fmt.Println()
	}
}

// ──────────────────────────────────────────────────────────────────────
// 2. Distance Calculations
// ──────────────────────────────────────────────────────────────────────

func demoDistanceCalculations() {
	printHeader("2. Distance Calculations (Haversine & Grid)")

	pairs := []struct {
		nameA, nameB string
		a, b         rh.Location
	}{
		{
			"Nairobi CBD", "JKIA",
			rh.Location{Lat: -1.2921, Lng: 36.8219},
			rh.Location{Lat: -1.3192, Lng: 36.9278},
		},
		{
			"Nairobi CBD", "Westlands",
			rh.Location{Lat: -1.2921, Lng: 36.8219},
			rh.Location{Lat: -1.2673, Lng: 36.8114},
		},
		{
			"Nairobi CBD", "Karen",
			rh.Location{Lat: -1.2921, Lng: 36.8219},
			rh.Location{Lat: -1.3187, Lng: 36.7114},
		},
		{
			"Nairobi CBD", "Times Square NYC",
			rh.Location{Lat: -1.2921, Lng: 36.8219},
			rh.Location{Lat: 40.7580, Lng: -73.9855},
		},
	}

	for _, p := range pairs {
		haversineDist := rh.HaversineDistance(p.a, p.b)
		gridDist, err := rh.GridDistance(p.a, p.b, rh.DriverResolution)

		fmt.Printf("  %s -> %s\n", p.nameA, p.nameB)
		fmt.Printf("    Haversine distance:  %.3f km\n", haversineDist)
		if err == nil {
			fmt.Printf("    Grid distance (res 9): %d cells\n", gridDist)
		} else {
			fmt.Printf("    Grid distance (res 9): N/A (too far apart for grid path)\n")
		}

		// Show symmetry
		reverseDist := rh.HaversineDistance(p.b, p.a)
		fmt.Printf("    Symmetric check:     %.3f km (reverse) -- match: %v\n",
			reverseDist, math.Abs(haversineDist-reverseDist) < 0.001)
		fmt.Println()
	}
}

// ──────────────────────────────────────────────────────────────────────
// 3. Driver Management
// ──────────────────────────────────────────────────────────────────────

func demoDriverManagement(gm *rh.GridManager) {
	printHeader("3. Driver Registration & Management")

	// Nairobi area drivers
	drivers := []*rh.Driver{
		{ID: "D001", Name: "James Mwangi", Location: rh.Location{Lat: -1.2921, Lng: 36.8219}, Rating: 4.8},
		{ID: "D002", Name: "Faith Wanjiku", Location: rh.Location{Lat: -1.2935, Lng: 36.8230}, Rating: 4.9},
		{ID: "D003", Name: "Peter Ochieng", Location: rh.Location{Lat: -1.2673, Lng: 36.8114}, Rating: 4.5},
		{ID: "D004", Name: "Mary Akinyi", Location: rh.Location{Lat: -1.2800, Lng: 36.8150}, Rating: 4.7},
		{ID: "D005", Name: "John Kamau", Location: rh.Location{Lat: -1.3000, Lng: 36.8300}, Rating: 4.6},
		{ID: "D006", Name: "Sarah Njeri", Location: rh.Location{Lat: -1.2750, Lng: 36.8050}, Rating: 4.4},
		{ID: "D007", Name: "David Mutua", Location: rh.Location{Lat: -1.3100, Lng: 36.8400}, Rating: 4.3},
		{ID: "D008", Name: "Grace Wambui", Location: rh.Location{Lat: -1.2600, Lng: 36.8200}, Rating: 4.9},
	}

	for _, d := range drivers {
		err := gm.RegisterDriver(d)
		if err != nil {
			fmt.Printf("  ERROR registering %s: %v\n", d.Name, err)
		} else {
			fmt.Printf("  Registered: %-18s (%.4f, %.4f) -> Cell %v\n",
				d.Name, d.Location.Lat, d.Location.Lng, d.CellID)
		}
	}

	fmt.Printf("\n  Total drivers:  %d\n", gm.GetDriverCount())
	fmt.Printf("  Active drivers: %d\n", gm.GetActiveDriverCount())

	// Update a driver location
	newLoc := rh.Location{Lat: -1.2800, Lng: 36.8300}
	err := gm.UpdateDriverLocation("D003", newLoc)
	if err != nil {
		fmt.Printf("  ERROR updating D003: %v\n", err)
	} else {
		fmt.Printf("\n  Updated D003 (Peter Ochieng) location to (%.4f, %.4f)\n", newLoc.Lat, newLoc.Lng)
	}

	// Remove a driver
	err = gm.RemoveDriver("D007")
	if err != nil {
		fmt.Printf("  ERROR removing D007: %v\n", err)
	} else {
		fmt.Println("  Removed D007 (David Mutua)")
	}

	fmt.Printf("  Total drivers after removal: %d\n", gm.GetDriverCount())
	fmt.Println()
}

// ──────────────────────────────────────────────────────────────────────
// 4. Nearby Driver Search
// ──────────────────────────────────────────────────────────────────────

func demoNearbySearch(gm *rh.GridManager) {
	printHeader("4. Nearby Driver Search (K-Ring Expansion)")

	riderLoc := rh.Location{Lat: -1.2921, Lng: 36.8219} // Nairobi CBD
	fmt.Printf("  Rider location: Nairobi CBD (%.4f, %.4f)\n\n", riderLoc.Lat, riderLoc.Lng)

	for k := 1; k <= 5; k++ {
		nearby := gm.FindNearbyDrivers(riderLoc, k)
		fmt.Printf("  K-Ring %d: Found %d driver(s)\n", k, len(nearby))
		for _, d := range nearby {
			dist := rh.HaversineDistance(riderLoc, d.Location)
			fmt.Printf("    - %-18s  %.3f km away  (rating: %.1f)\n", d.Name, dist, d.Rating)
		}
	}

	// Find nearest driver
	fmt.Println()
	nearest, dist := gm.FindNearestDriver(riderLoc, 5)
	if nearest != nil {
		fmt.Printf("  Nearest driver: %s (%.3f km away, rating: %.1f)\n",
			nearest.Name, dist, nearest.Rating)
	} else {
		fmt.Println("  No nearby drivers found!")
	}
	fmt.Println()
}

// ──────────────────────────────────────────────────────────────────────
// 5. Routing (Multi-Stop Trip Distance)
// ──────────────────────────────────────────────────────────────────────

func demoRouting() {
	printHeader("5. Routing & Multi-Stop Trip Distance")

	// Simulate a multi-stop route through Nairobi
	stops := []struct {
		name string
		loc  rh.Location
	}{
		{"Nairobi CBD", rh.Location{Lat: -1.2921, Lng: 36.8219}},
		{"Westlands", rh.Location{Lat: -1.2673, Lng: 36.8114}},
		{"Parklands", rh.Location{Lat: -1.2590, Lng: 36.8180}},
		{"Gigiri (UN)", rh.Location{Lat: -1.2340, Lng: 36.8016}},
		{"Village Market", rh.Location{Lat: -1.2290, Lng: 36.8035}},
	}

	fmt.Println("  Route stops:")
	for i, s := range stops {
		cell := rh.LocationToCell(s.loc, rh.DriverResolution)
		fmt.Printf("    %d. %-20s  (%.4f, %.4f)  Cell: %v\n",
			i+1, s.name, s.loc.Lat, s.loc.Lng, cell)
	}

	fmt.Println("\n  Leg distances:")
	totalDist := 0.0
	for i := 0; i < len(stops)-1; i++ {
		legDist := rh.EstimateTripDistance(stops[i].loc, stops[i+1].loc)
		totalDist += legDist
		fmt.Printf("    %s -> %s: %.3f km\n", stops[i].name, stops[i+1].name, legDist)
	}

	// Direct distance for comparison
	directDist := rh.EstimateTripDistance(stops[0].loc, stops[len(stops)-1].loc)
	fmt.Printf("\n  Total route distance:  %.3f km\n", totalDist)
	fmt.Printf("  Direct distance:       %.3f km\n", directDist)
	fmt.Printf("  Route overhead:        %.1f%%\n", (totalDist/directDist-1)*100)

	// Check if consecutive stops share neighbor cells
	fmt.Println("\n  Cell adjacency between stops:")
	for i := 0; i < len(stops)-1; i++ {
		cellA := rh.LocationToCell(stops[i].loc, rh.DriverResolution)
		cellB := rh.LocationToCell(stops[i+1].loc, rh.DriverResolution)
		isNeighbor := rh.AreCellsNeighbors(cellA, cellB)
		sameCell := cellA == cellB
		fmt.Printf("    %s <-> %s: neighbor=%v, sameCell=%v\n",
			stops[i].name, stops[i+1].name, isNeighbor, sameCell)
	}
	fmt.Println()
}

// ──────────────────────────────────────────────────────────────────────
// 6. Surge Pricing
// ──────────────────────────────────────────────────────────────────────

func demoSurgePricing(gm *rh.GridManager) {
	printHeader("6. Surge Pricing Zones")

	// Set up surge zones in high-demand areas
	surgeAreas := []struct {
		name                 string
		loc                  rh.Location
		demand, supply       int
	}{
		{"Nairobi CBD (rush hour)", rh.Location{Lat: -1.2921, Lng: 36.8219}, 50, 20},
		{"Westlands (evening)", rh.Location{Lat: -1.2673, Lng: 36.8114}, 30, 25},
		{"JKIA (flight arrivals)", rh.Location{Lat: -1.3192, Lng: 36.9278}, 40, 10},
		{"Karen (normal)", rh.Location{Lat: -1.3187, Lng: 36.7114}, 10, 15},
	}

	baseFarePerKm := 50.0 // KES per km

	for _, area := range surgeAreas {
		surge := rh.CalculateSurge(area.demand, area.supply)
		zone := gm.SetSurgeZone(area.loc, surge, area.demand, area.supply)
		cell := rh.LocationToCell(area.loc, rh.SurgeResolution)

		fmt.Printf("  %-28s\n", area.name)
		fmt.Printf("    Demand: %d | Supply: %d | Ratio: %.1f\n",
			area.demand, area.supply, float64(area.demand)/float64(area.supply))
		fmt.Printf("    Surge multiplier: %.1fx\n", zone.Multiplier)
		fmt.Printf("    Surge cell (res 7): %v\n", cell)

		// Price example for a 5km trip
		price := baseFarePerKm * 5.0 * surge
		fmt.Printf("    5 km trip price: KES %.0f (base: KES %.0f)\n", price, baseFarePerKm*5.0)
		fmt.Println()
	}

	// Query surge at specific locations
	fmt.Println("  Surge lookup at locations:")
	queryLocs := []struct {
		name string
		loc  rh.Location
	}{
		{"Nairobi CBD", rh.Location{Lat: -1.2921, Lng: 36.8219}},
		{"Suburb (no surge)", rh.Location{Lat: -1.3500, Lng: 36.7500}},
	}
	for _, q := range queryLocs {
		mult := gm.GetSurgeMultiplier(q.loc)
		fmt.Printf("    %-25s -> surge: %.1fx\n", q.name, mult)
	}
	fmt.Println()
}

// ──────────────────────────────────────────────────────────────────────
// 7. Cell Hierarchy (Parent / Children)
// ──────────────────────────────────────────────────────────────────────

func demoCellHierarchy() {
	printHeader("7. H3 Cell Hierarchy (Parent & Children)")

	loc := rh.Location{Lat: -1.2921, Lng: 36.8219} // Nairobi CBD
	cell := rh.LocationToCell(loc, rh.DriverResolution)

	fmt.Printf("  Base cell (res %d): %v\n\n", rh.DriverResolution, cell)

	// Show parent cells at coarser resolutions
	fmt.Println("  Parent cells (coarser resolutions):")
	for res := rh.DriverResolution - 1; res >= 0; res-- {
		parent := rh.GetParentCell(cell, res)
		parentCenter := rh.CellToLocation(parent)
		fmt.Printf("    Res %d: %v  center=(%.4f, %.4f)\n",
			res, parent, parentCenter.Lat, parentCenter.Lng)
	}

	// Show children cells at one finer resolution
	fmt.Printf("\n  Children cells at res %d:\n", rh.DriverResolution+1)
	children := rh.GetChildrenCells(cell, rh.DriverResolution+1)
	fmt.Printf("    Count: %d\n", len(children))
	for i, child := range children {
		childCenter := rh.CellToLocation(child)
		fmt.Printf("    [%d] %v  center=(%.6f, %.6f)\n",
			i, child, childCenter.Lat, childCenter.Lng)
	}

	// Show that parent of child == original cell
	if len(children) > 0 {
		parentBack := rh.GetParentCell(children[0], rh.DriverResolution)
		fmt.Printf("\n  Verify: Parent of child[0] == original cell? %v\n", parentBack == cell)
	}
	fmt.Println()
}

// ──────────────────────────────────────────────────────────────────────
// 8. Cell Boundaries
// ──────────────────────────────────────────────────────────────────────

func demoCellBoundaries() {
	printHeader("8. H3 Cell Boundaries (Polygon Vertices)")

	loc := rh.Location{Lat: -1.2921, Lng: 36.8219}
	cell := rh.LocationToCell(loc, rh.DriverResolution)

	boundary := rh.GetCellBoundary(cell)
	fmt.Printf("  Cell %v has %d boundary vertices:\n", cell, len(boundary))
	for i, v := range boundary {
		fmt.Printf("    Vertex %d: (%.8f, %.8f)\n", i, v.Lat, v.Lng)
	}

	// Show that boundary forms a closed polygon
	if len(boundary) > 0 {
		first := boundary[0]
		last := boundary[len(boundary)-1]
		dist := rh.HaversineDistance(first, last)
		fmt.Printf("\n  Distance between first and last vertex: %.6f km\n", dist)
		fmt.Printf("  (These are adjacent vertices of the hexagon, not the same point)\n")
	}

	// Also show boundary at surge resolution for comparison
	surgeCell := rh.LocationToCell(loc, rh.SurgeResolution)
	surgeBoundary := rh.GetCellBoundary(surgeCell)
	fmt.Printf("\n  Surge cell (res %d) has %d boundary vertices:\n",
		rh.SurgeResolution, len(surgeBoundary))
	for i, v := range surgeBoundary {
		fmt.Printf("    Vertex %d: (%.8f, %.8f)\n", i, v.Lat, v.Lng)
	}
	fmt.Println()
}

// ──────────────────────────────────────────────────────────────────────
// 9. Trip Creation (End-to-End)
// ──────────────────────────────────────────────────────────────────────

func demoTripCreation(gm *rh.GridManager) {
	printHeader("9. End-to-End Trip Creation")

	rider := &rh.Rider{
		ID:       "R001",
		Name:     "Alice Muthoni",
		Location: rh.Location{Lat: -1.2921, Lng: 36.8219},
	}

	pickup := rh.Location{Lat: -1.2921, Lng: 36.8219}   // Nairobi CBD
	dropoff := rh.Location{Lat: -1.3192, Lng: 36.9278}   // JKIA

	fmt.Printf("  Rider:   %s\n", rider.Name)
	fmt.Printf("  Pickup:  Nairobi CBD   (%.4f, %.4f)\n", pickup.Lat, pickup.Lng)
	fmt.Printf("  Dropoff: JKIA          (%.4f, %.4f)\n", dropoff.Lat, dropoff.Lng)
	fmt.Println()

	trip, err := gm.CreateTrip("T001", rider, pickup, dropoff)
	if err != nil {
		fmt.Printf("  ERROR creating trip: %v\n", err)
		return
	}

	baseFare := 50.0 // KES per km
	tripCost := trip.EstimatedDistKm * baseFare * trip.SurgeMultiplier

	fmt.Printf("  Trip created!\n")
	fmt.Printf("    Trip ID:          %s\n", trip.ID)
	fmt.Printf("    Assigned driver:  %s (rating: %.1f)\n", trip.Driver.Name, trip.Driver.Rating)
	fmt.Printf("    Estimated dist:   %.3f km\n", trip.EstimatedDistKm)
	fmt.Printf("    Surge multiplier: %.1fx\n", trip.SurgeMultiplier)
	fmt.Printf("    Estimated fare:   KES %.0f\n", tripCost)
	fmt.Printf("    Pickup cell:      %v\n", trip.PickupCell)
	fmt.Printf("    Dropoff cell:     %v\n", trip.DropoffCell)
	fmt.Printf("    Driver now active: %v\n", trip.Driver.Active)

	fmt.Printf("\n  Active drivers after trip: %d / %d\n", gm.GetActiveDriverCount(), gm.GetDriverCount())
	fmt.Println()
}

// ──────────────────────────────────────────────────────────────────────
// 10. Driver Ranking
// ──────────────────────────────────────────────────────────────────────

func demoDriverRanking(gm *rh.GridManager) {
	printHeader("10. Driver Ranking by Distance")

	origin := rh.Location{Lat: -1.2921, Lng: 36.8219} // Nairobi CBD
	nearby := gm.FindNearbyDrivers(origin, 5)

	if len(nearby) == 0 {
		fmt.Println("  No nearby drivers to rank.")
		return
	}

	ranked := rh.RankDriversByDistance(nearby, origin)
	fmt.Printf("  Ranking %d drivers from Nairobi CBD:\n\n", len(ranked.Drivers))
	fmt.Printf("  %-5s %-18s %-12s %-8s %s\n", "Rank", "Name", "Distance", "Rating", "Active")
	fmt.Printf("  %s\n", strings.Repeat("-", 60))
	for i, d := range ranked.Drivers {
		status := "Yes"
		if !d.Active {
			status = "No (on trip)"
		}
		fmt.Printf("  %-5d %-18s %-12.3f %-8.1f %s\n",
			i+1, d.Name, ranked.Distances[i], d.Rating, status)
	}

	// Simulate random driver positions and re-rank
	fmt.Println("\n  Simulating 5 random drivers for ranking demo:")
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomDrivers := make([]*rh.Driver, 5)
	for i := 0; i < 5; i++ {
		randomDrivers[i] = &rh.Driver{
			ID:   fmt.Sprintf("RAND-%d", i),
			Name: fmt.Sprintf("RandomDriver-%d", i),
			Location: rh.Location{
				Lat: origin.Lat + (rng.Float64()-0.5)*0.05,
				Lng: origin.Lng + (rng.Float64()-0.5)*0.05,
			},
			Rating: 3.0 + rng.Float64()*2.0,
			Active: true,
		}
	}

	randomRanked := rh.RankDriversByDistance(randomDrivers, origin)
	fmt.Printf("\n  %-5s %-18s %-12s %-8s\n", "Rank", "Name", "Distance", "Rating")
	fmt.Printf("  %s\n", strings.Repeat("-", 50))
	for i, d := range randomRanked.Drivers {
		fmt.Printf("  %-5d %-18s %-12.3f %-8.1f\n",
			i+1, d.Name, randomRanked.Distances[i], d.Rating)
	}
	fmt.Println()
}

// ──────────────────────────────────────────────────────────────────────

func printHeader(title string) {
	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("  %s\n", title)
	fmt.Println(strings.Repeat("─", 60))
}
