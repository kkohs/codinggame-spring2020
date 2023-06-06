package main

import (
	"container/heap"
	"fmt"
	"time"
)
import "os"
import "bufio"

// debug logging method
func log(a ...any) {
	_, _ = fmt.Fprintln(os.Stderr, a)
}

// Pac structs
type Pac struct {
	Id               int
	Mine             bool
	X                int
	Y                int
	TypeId           string
	SpeedTurnsLeft   int
	AbilityCooldown  int
	TargetX          int
	TargetY          int
	TargetPelletId   int
	TargetPelletDist int
}

// Pellet structs
type Pellet struct {
	X        int
	Y        int
	Value    int
	Consumed bool
	Targeted bool
}

// String
func (p Pellet) String() string {
	return fmt.Sprintf("Pellet (%d, %d) %d %v", p.X, p.Y, p.Value, p.Consumed)
}

// Cell type struct
type Type string

// Cell type constants
const (
	Empty Type = " "
	Wall  Type = "#"
)

// Cell structs
type Cell struct {
	x, y    int
	isWall  bool
	g, h, f int
	parent  *Cell
	index   int // index in the heap
	// Neighbors
	Neighbors []*Cell
}

// Initialize neighbors for cell
func (c *Cell) InitNeighbors(grid [][]*Cell) {
	c.Neighbors = getNeighbors(c, grid)
}

type PriorityQueue []*Cell

// PriorityQueue methods

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].f < pq[j].f
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Cell)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

func (pq *PriorityQueue) update(item *Cell, g, h int) {
	item.g = g
	item.h = h
	item.f = g + h
	heap.Fix(pq, item.index)
}

func getNeighbors(cell *Cell, grid [][]*Cell) []*Cell {
	neighbors := []*Cell{}
	x, y := cell.x, cell.y
	if x > 0 {
		neighbors = append(neighbors, grid[y][x-1])
	}
	if x < len(grid[0])-1 {
		neighbors = append(neighbors, grid[y][x+1])
	}
	if y > 0 {
		neighbors = append(neighbors, grid[y-1][x])
	}
	if y < len(grid)-1 {
		neighbors = append(neighbors, grid[y+1][x])
	}
	return neighbors
}

func manhattanDistance(a, b *Cell) int {
	return abs(a.x-b.x) + abs(a.y-b.y)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Clone grid of cells
func cloneGrid(grid [][]*Cell) [][]*Cell {
	newGrid := make([][]*Cell, len(grid))
	for y, row := range grid {
		newGrid[y] = make([]*Cell, len(row))
		for x, cell := range row {
			newGrid[y][x] = &Cell{
				x:      cell.x,
				y:      cell.y,
				isWall: cell.isWall,
			}
		}
	}
	return newGrid
}

func AStar(startX, startY, endX, endY int, grid [][]*Cell) []*Cell {
	openSet := &PriorityQueue{}
	clone := cloneGrid(grid)
	heap.Init(openSet)
	start := GetCell(startX, startY, clone)
	goal := GetCell(endX, endY, clone)
	heap.Push(openSet, start)

	closedSet := make(map[*Cell]bool)
	for openSet.Len() > 0 {
		current := heap.Pop(openSet).(*Cell)

		if current == goal {
			var path []*Cell
			for current != nil {
				path = append([]*Cell{current}, path...)
				current = current.parent
			}
			for _, cell := range path {
				log(cell.x, cell.y)
			}
			return path
		}
		closedSet[current] = true

		for _, neighbor := range current.Neighbors {
			if neighbor.isWall || closedSet[neighbor] {
				continue
			}

			tentativeGScore := current.g + 1
			if !contains(openSet, neighbor) {
				heap.Push(openSet, neighbor)
			} else if tentativeGScore >= neighbor.g {
				continue
			}

			neighbor.parent = current
			openSet.update(neighbor, tentativeGScore, manhattanDistance(neighbor, goal))
		}
	}

	return nil
}

func contains(pq *PriorityQueue, cell *Cell) bool {
	for _, c := range *pq {
		if c == cell {
			return true
		}
	}
	return false
}

// Game state structs
type Game struct {
	Width               int
	Height              int
	MyPacs              []*Pac
	OpponentPacs        []*Pac
	Pellet              []*Pellet
	Grid                [][]*Cell
	MyScore             int
	OpponentScore       int
	VisiblePacCount     int
	VisiblePalleteCount int
}

// Get cell pointer at x, y
func GetCell(x, y int, grid [][]*Cell) *Cell {
	return grid[y][x]
}

// Add pac or update existing pac location data to state mine or opponent
func (g *Game) AddPac(id, mine, x, y int, typeId string, speedTurnsLeft, abilityCooldown int) {
	var pacs []*Pac
	if mine == 1 {
		pacs = g.MyPacs
	} else {
		pacs = g.OpponentPacs
	}
	for _, pac := range pacs {
		if pac.Id == id {
			pac.X = x
			pac.Y = y
			pac.TypeId = typeId
			pac.SpeedTurnsLeft = speedTurnsLeft
			pac.AbilityCooldown = abilityCooldown
			return
		}
	}
	pacs = append(pacs, &Pac{
		Id:              id,
		Mine:            mine == 1,
		X:               x,
		Y:               y,
		TypeId:          typeId,
		SpeedTurnsLeft:  speedTurnsLeft,
		AbilityCooldown: abilityCooldown,
		TargetX:         x,
		TargetY:         y,
	})
	if mine == 1 {
		g.MyPacs = pacs
	} else {
		g.OpponentPacs = pacs
	}
}

// Add pellet or update existing pellet location data to state
func (g *Game) AddPellet(id, x, y, value int) {
	for _, pellet := range g.Pellet {
		if pellet.X == x && pellet.Y == y {
			pellet.X = x
			pellet.Y = y
			pellet.Value = value
			pellet.Consumed = false
			return
		}
	}
	g.Pellet = append(g.Pellet, &Pellet{
		X:        x,
		Y:        y,
		Value:    value,
		Consumed: false,
	})
}

// Get the closest super pallet to pac using a star
func (g *Game) GetClosestSuperPallet(pac *Pac) *Pellet {
	var closest *Pellet
	var closestDist int
	for _, pallet := range g.Pellet {
		if pallet.Value == 10 && !pallet.Consumed && !pallet.Targeted {
			path := AStar(pac.X, pac.Y, pallet.X, pallet.Y, g.Grid)
			if closest == nil || len(path) < closestDist {
				closest = pallet
				closestDist = len(path)
			}
		}
	}
	return closest
}

// Get closest regular pallet to pac
func (g *Game) GetClosestRegularPallet(pac *Pac) *Pellet {
	var closest *Pellet
	var closestDist int
	for _, pallet := range g.Pellet {
		if pallet.Value == 1 && !pallet.Consumed && !pallet.Targeted {
			path := AStar(pac.X, pac.Y, pallet.X, pallet.Y, g.Grid)
			if closest == nil || len(path) < closestDist {
				closest = pallet
				closestDist = len(path)
			}
		}
	}
	return closest
}

// Get pallet by cordinates
func (g *Game) GetPallet(x, y int) *Pellet {
	log("Getting pallet", x, y)
	log("total pallets", len(g.Pellet))
	for _, pallet := range g.Pellet {
		if pallet.X == 19 && pallet.Y == 2 {
			log("Pallet 19,2", pallet)
		}
		if pallet.X == x && pallet.Y == y {
			return pallet
		}
	}
	return nil
}

// Check if pac target has been eaten already and remove current target if so
func (g *Game) CheckTargetEaten(pac *Pac) {
	pallet := g.GetPallet(pac.TargetX, pac.TargetY)
	log("Checking target", pac.TargetX, pac.TargetY, "pallet", pallet)
	if pallet != nil {
		log("Target eaten", pac.TargetX, pac.TargetY, pallet.Value)
		if pallet.Consumed {
			pac.TargetX = pac.X
			pac.TargetY = pac.Y
			pac.TargetPelletDist = -1
		}
	} else {
		pac.TargetX = pac.X
		pac.TargetY = pac.Y
		pac.TargetPelletDist = -1

	}
}

// Remove pallet from game o  current Pac cordinates
func (g *Game) RemovePallet(pac *Pac) {
	pallet := g.GetPallet(pac.X, pac.Y)
	if pallet != nil {
		log("Pac", pac.Id, "ate pallet", pallet.X, pallet.Y, pallet.Value)
		pallet.Consumed = true
	}
}

// Play a turn
func (g *Game) PlayTurn() {
	startTime := time.Now()
	log(len(g.MyPacs))
	for _, pac := range g.MyPacs {
		log("Pac", pac.Id, "x", pac.X, "y", pac.Y)
		g.RemovePallet(pac)
		g.CheckTargetEaten(pac)
	}
	for _, pac := range g.OpponentPacs {
		g.RemovePallet(pac)
	}
	moves := ""
	for _, pac := range g.MyPacs {
		log("Pac", pac.Id, "x", pac.X, "y", pac.Y, "type", pac.TypeId, "speed turns left", pac.SpeedTurnsLeft, "ability cooldown", pac.AbilityCooldown, "target x", pac.TargetX, "target y", pac.TargetY, "target pellet dist", pac.TargetPelletDist)
		if pac.X == pac.TargetX && pac.Y == pac.TargetY {
			log("Pac", pac.Id, "reached target", pac.TargetX, pac.TargetY)
			old := g.GetPallet(pac.TargetX, pac.TargetY)
			if old != nil {
				old.Value = 0
				log("Pac", pac.Id, "ate pallet", pac.TargetX, pac.TargetY)
				pac.TargetX = -1
				pac.TargetY = -1
				pac.TargetPelletDist = -1
			}

			pallet := g.GetClosestSuperPallet(pac)
			if pallet != nil {
				moves += fmt.Sprintf("MOVE %d %d %d|", pac.Id, pallet.X, pallet.Y)
				pac.TargetX = pallet.X
				pac.TargetY = pallet.Y
				pac.TargetPelletDist = len(AStar(pac.X, pac.Y, pallet.X, pallet.Y, g.Grid))
				pallet.Targeted = true
			} else {
				pallet = g.GetClosestRegularPallet(pac)
				if pallet != nil {
					moves += fmt.Sprintf("MOVE %d %d %d|", pac.Id, pallet.X, pallet.Y)
					pac.TargetX = pallet.X
					pac.TargetY = pallet.Y
					pac.TargetPelletDist = len(AStar(pac.X, pac.Y, pallet.X, pallet.Y, g.Grid))
					pallet.Targeted = true
				} else {
					moves += fmt.Sprintf("MOVE %d %d %d|", pac.Id, pallet.X, pallet.Y)
					pac.TargetX = pac.X
					pac.TargetY = pac.Y
					pac.TargetPelletDist = 0
				}
			}
		} else {
			moves += fmt.Sprintf("MOVE %d %d %d|", pac.Id, pac.TargetX, pac.TargetY)
		}
	}
	fmt.Println(moves)
	log("Turn took", time.Since(startTime))
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1000000), 1000000)

	// game: game state
	var game Game
	game.MyPacs = make([]*Pac, 0)
	game.OpponentPacs = make([]*Pac, 0)
	game.Pellet = make([]*Pellet, 0)
	// width: size of the grid
	// height: top left corner is (x=0, y=0)
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &game.Width, &game.Height)
	game.Grid = make([][]*Cell, game.Height)
	for i := range game.Grid {
		scanner.Scan()
		row := scanner.Text()
		game.Grid[i] = make([]*Cell, game.Width)
		for j, c := range row {
			game.Grid[i][j] = &Cell{
				x:      j,
				y:      i,
				isWall: c == '#',
			}
		}
	}

	for _, cells := range game.Grid {
		for _, cell := range cells {
			cell.InitNeighbors(game.Grid)
		}
	}
	for {
		var myScore, opponentScore int
		scanner.Scan()
		fmt.Sscan(scanner.Text(), &myScore, &opponentScore)
		game.MyScore = myScore
		game.OpponentScore = opponentScore
		// remove all pallets
		for _, pallet := range game.Pellet {
			pallet.Consumed = true
		}
		// visiblePacCount: all your pacs and enemy pacs in sight
		var visiblePacCount int
		scanner.Scan()
		fmt.Sscan(scanner.Text(), &visiblePacCount)
		game.VisiblePacCount = visiblePacCount
		log("Visible pac count", visiblePacCount)
		for i := 0; i < visiblePacCount; i++ {
			// pacId: pac number (unique within a team)
			// mine: true if this pac is yours
			// x: position in the grid
			// y: position in the grid
			// typeId: unused in wood leagues
			// speedTurnsLeft: unused in wood leagues
			// abilityCooldown: unused in wood leagues
			var pacId int
			var _mine int
			var x, y int
			var typeId string
			var speedTurnsLeft, abilityCooldown int
			scanner.Scan()
			fmt.Sscan(scanner.Text(), &pacId, &_mine, &x, &y, &typeId, &speedTurnsLeft, &abilityCooldown)
			log("pac id", pacId, "mine", _mine, "x", x, "y", y, "type id", typeId, "speed turns left",
				speedTurnsLeft, "ability cooldown", abilityCooldown)
			game.AddPac(pacId, _mine, x, y, typeId, speedTurnsLeft, abilityCooldown)
		}
		// visiblePelletCount: all pellets in sight
		var visiblePelletCount int
		scanner.Scan()
		fmt.Sscan(scanner.Text(), &visiblePelletCount)
		game.VisiblePalleteCount = visiblePelletCount
		for i := 0; i < visiblePelletCount; i++ {
			// value: amount of points this pellet is worth
			var x, y, value int
			scanner.Scan()
			fmt.Sscan(scanner.Text(), &x, &y, &value)
			game.AddPellet(i, x, y, value)
			if x == 19 && y == 9 {
				log("Pellet", i, "x", x, "y", y, "value", value)
			}
		}

		pellets := ""
		for _, pellet := range game.Pellet {
			pellets += pellet.String() + " "
		}
		//log(pellets)

		game.PlayTurn()
	}
}
