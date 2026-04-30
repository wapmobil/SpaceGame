package game

import (
	"fmt"
	"math/rand"
)

// DrillConfig holds configuration for drill world generation
type DrillConfig struct {
	Seed int64
}

// Depth zones for geological layer system
type DepthZone struct {
	Name       string
	MinDepth   int
	MaxDepth   int
	Dirt       float64
	Stone      float64
	Metal      float64
	Mithril    float64
}

var depthZones = []DepthZone{
	{Name: "surface", MinDepth: 0, MaxDepth: 64, Dirt: 0.60, Stone: 0.25, Metal: 0.12, Mithril: 0.03},
	{Name: "shallow", MinDepth: 64, MaxDepth: 128, Dirt: 0.40, Stone: 0.35, Metal: 0.20, Mithril: 0.05},
	{Name: "forest", MinDepth: 128, MaxDepth: 192, Dirt: 0.20, Stone: 0.45, Metal: 0.30, Mithril: 0.05},
	{Name: "mid", MinDepth: 192, MaxDepth: 256, Dirt: 0.10, Stone: 0.35, Metal: 0.40, Mithril: 0.15},
	{Name: "deep", MinDepth: 256, MaxDepth: 320, Dirt: 0.02, Stone: 0.20, Metal: 0.45, Mithril: 0.33},
	{Name: "abyss", MinDepth: 320, MaxDepth: 9999, Dirt: 0.00, Stone: 0.05, Metal: 0.25, Mithril: 0.70},
}

const (
	noiseGridSize = 16
	noiseScale    = float64(noiseGridSize)
	caveThreshold = 0.85
)

// initNoiseGrids generates random noise grids for terrain, caves, and veins
func (g *DrillGame) initNoiseGrids() {
	r := rand.New(rand.NewSource(g.config.Seed))

	for x := 0; x < noiseGridSize; x++ {
		for y := 0; y < noiseGridSize; y++ {
			g.terrainNoiseGrid[x][y] = r.Float64()
			g.caveNoiseGrid[x][y] = r.Float64()
			g.veinNoiseGrid[x][y] = r.Float64()
		}
	}

	g.applyCaveSmoothing()
}

func (g *DrillGame) applyCaveSmoothing() {
	newGrid := make([][]float64, noiseGridSize)
	for i := range newGrid {
		newGrid[i] = make([]float64, noiseGridSize)
	}

	for x := 0; x < noiseGridSize; x++ {
		for y := 0; y < noiseGridSize; y++ {
			caveCount := 0
			for dx := -1; dx <= 1; dx++ {
				for dy := -1; dy <= 1; dy++ {
					nx := (x + dx + noiseGridSize) % noiseGridSize
					ny := (y + dy + noiseGridSize) % noiseGridSize
					if g.caveNoiseGrid[nx][ny] > caveThreshold {
						caveCount++
					}
				}
			}
			if g.caveNoiseGrid[x][y] > caveThreshold {
				if caveCount >= 5 {
					newGrid[x][y] = 1.0
				} else {
					newGrid[x][y] = 0.0
				}
			} else {
				if caveCount >= 6 {
					newGrid[x][y] = 1.0
				} else {
					newGrid[x][y] = 0.0
				}
			}
		}
	}

	for x := 0; x < noiseGridSize; x++ {
		for y := 0; y < noiseGridSize; y++ {
			g.caveNoiseGrid[x][y] = newGrid[x][y]
		}
	}
}

func (g *DrillGame) sampleNoise(grid [noiseGridSize][noiseGridSize]float64, fx, fy float64) float64 {
	gx0 := ((int(fx) % noiseGridSize) + noiseGridSize) % noiseGridSize
	gy0 := ((int(fy) % noiseGridSize) + noiseGridSize) % noiseGridSize
	gx1 := (gx0 + 1) % noiseGridSize
	gy1 := (gy0 + 1) % noiseGridSize

	fx0 := fx - float64(int(fx))
	fy0 := fy - float64(int(fy))

	sx := fx0 * fx0 * (3 - 2*fx0)
	sy := fy0 * fy0 * (3 - 2*fy0)

	n00 := grid[gx0][gy0]
	n10 := grid[gx1][gy0]
	n01 := grid[gx0][gy1]
	n11 := grid[gx1][gy1]

	x1 := n00 + (n10-n00)*sx
	x2 := n01 + (n11-n01)*sx

	return x1 + (x2-x1)*sy
}

func (g *DrillGame) getTerrainNoise(x, y int) float64 {
	fx := float64(x) / noiseScale
	fy := float64(y) / noiseScale
	return g.sampleNoise(g.terrainNoiseGrid, fx, fy)
}

func (g *DrillGame) getCaveNoise(x, y int) float64 {
	fx := float64(x) / noiseScale
	fy := float64(y) / noiseScale
	return g.sampleNoise(g.caveNoiseGrid, fx, fy)
}

func (g *DrillGame) getVeinNoise(x, y int) float64 {
	fx := float64(x) / noiseScale
	fy := float64(y) / noiseScale
	return g.sampleNoise(g.veinNoiseGrid, fx, fy)
}

func (g *DrillGame) getDepthZone(depth int) *DepthZone {
	for i := range depthZones {
		z := &depthZones[i]
		if depth >= z.MinDepth && depth < z.MaxDepth {
			return z
		}
	}
	return &depthZones[len(depthZones)-1]
}

func (g *DrillGame) selectCellType(zone *DepthZone, noise float64, depth int) string {
	noiseBias := (noise - 0.5) * 0.35
	dirt := zone.Dirt + noiseBias
	stone := zone.Stone + noiseBias*0.5
	metal := zone.Metal + noiseBias*0.3
	mithril := zone.Mithril - noiseBias*0.5

	if mithril < 0 {
		metal += mithril
		mithril = 0
	}
	if metal < 0 {
		stone += metal
		metal = 0
	}
	if stone < 0 {
		dirt += stone
		stone = 0
	}
	if dirt < 0 {
		dirt = 0
	}

	rVal := noise
	if rVal < dirt {
		return CellDirt
	} else if rVal < dirt+stone {
		return CellStone
	} else if rVal < dirt+stone+metal {
		return CellMetal
	}
	return CellMithril
}

func (g *DrillGame) isCave(x, y int) bool {
	caveNoise := g.getCaveNoise(x, y)
	return caveNoise > caveThreshold
}

func (g *DrillGame) getVeinMultiplier(x, y int) float64 {
	veinNoise := g.getVeinNoise(x, y)
	if veinNoise > 0.7 {
		return 1.8
	} else if veinNoise > 0.5 {
		return 1.3
	} else if veinNoise > 0.3 {
		return 1.0
	}
	return 0.5
}

// generateInitialWorld creates the initial 5x5 world centered at depth 0
func (g *DrillGame) generateInitialWorld() {
	g.session.World = g.buildWorldAt(g.session.Depth, g.session.DrillX)
}

// buildWorldAt builds a 5x5 world with drill at top center
func (g *DrillGame) buildWorldAt(depth, drillX int) [][]Cell {
	world := make([][]Cell, DefaultWorldWidth)
	for dy := 0; dy < DefaultWorldWidth; dy++ {
		rowIdx := dy
		world[rowIdx] = make([]Cell, DefaultWorldWidth)
		for dx := -2; dx <= 2; dx++ {
			colIdx := dx + 2
			x := drillX + dx
			y := depth + dy
			cell := g.getCellWithState(x, y)
			world[rowIdx][colIdx] = cell
		}
	}
	return world
}

// GetChunk generates a chunk of the world centered at given coordinates
func (g *DrillGame) GetChunk(centerX, centerY, width, height int) [][]Cell {
	chunk := make([][]Cell, height)
	for dy := 0; dy < height; dy++ {
		chunk[dy] = make([]Cell, width)
		for dx := 0; dx < width; dx++ {
			x := centerX - width/2 + dx
			y := centerY - height/2 + dy
			chunk[dy][dx] = g.getCellWithState(x, y)
		}
	}
	return chunk
}

// getCellWithState returns a cell with tracked extraction state
func (g *DrillGame) getCellWithState(x, y int) Cell {
	cell := g.getCellAt(x, y)
	cellKey := fmt.Sprintf("%d,%d", x, y)
	if extracted, ok := g.session.ExtractedCells[cellKey]; ok && extracted {
		cell.Extracted = true
		cell.ResourceType = ""
		cell.ResourceValue = 0
	}
	return cell
}

// regenerateWorld rebuilds the 5x5 world centered on current drill position
func (g *DrillGame) regenerateWorld() {
	g.session.World = g.buildWorldAt(g.session.Depth, g.session.DrillX)
	g.cleanupCache()
}

// cleanupCache removes cells from the cache that are far above the drill
func (g *DrillGame) cleanupCache() {
	if len(g.cellsCache) == 0 {
		return
	}
	cutoff := g.session.Depth - 20
	for key := range g.cellsCache {
		var x, y int
		fmt.Sscanf(key, "%d,%d", &x, &y)
		if y < cutoff {
			delete(g.cellsCache, key)
		}
	}
}
