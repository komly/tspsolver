package main

import (
	"log"
	"math/rand"
	"sort"
	"fmt"
	"strings"
	"strconv"

	"math"
)

func init() {
	rand.Seed(42)//time.Now().Unix())
}

type edge struct {
	from, to int
}

type instance struct {
	edges []edge
	way []int
	cachedFitness int
}

func (s instance) bestPath() []int {
	bestN := 1
	n := 1
	seen := make(map[int]struct{})
	seen[s.edges[s.way[0]].from] = struct{}{}
	path := []int{
		s.edges[s.way[0]].from,
	}
	bestPath := path
	for i := 0; i < len(s.way) - 1; i++ {
		if _, ok := seen[s.edges[s.way[i]].to]; !ok && s.edges[s.way[i]].to == s.edges[s.way[i + 1]].from {
			n++
			seen[s.edges[s.way[i]].to] = struct{}{}
			path = append(path, s.edges[s.way[i]].to)

			if n > bestN {
				bestN = n
				bestPath = path
			}
		} else {
			n = 1
			path = []int{
				s.edges[s.way[i]].to,
			}
		}
	}
	return bestPath
}

func (s instance) fitness() int {
	if s.cachedFitness != 0 {
		return s.cachedFitness
	}

	fitness :=  len(estimate(s.edges, s.way))

	s.cachedFitness = fitness
	return s.cachedFitness
}
func (s instance) copy() instance {
	way := make([]int, len(s.way))
	copy(way, s.way)

	return instance{
		way: way,
		edges: s.edges,
	}
}

func pick(population []instance, totalFitness int) instance {
	k := rand.Intn(totalFitness)
	for i := 0; i < len(population);i++ {
		k -= population[i].fitness()
		if k <= 0 {
			return population[i]
		}
	}
	return population[len(population) - 1]
}

func cross(p1, p2 instance) instance {
	si, sj := rand.Intn(len(p1.way)), rand.Intn(len(p1.way))
	if si > sj {
		sj, si = si, sj
	}

	seen := make(map[int]struct{})
	way := make([]int, len(p1.way))

	for i := si; i < sj; i++ {
		seen[p2.way[i]] = struct{}{}
		way[i] = p2.way[i]
	}

	for i := 0; i < si; i++ {
		j := i
		for {
			if _, ok := seen[p1.way[j]]; !ok {
				break
			}
			j = (j + 1) % len(p1.way)
		}
		way[i] = p1.way[j]
		seen[p1.way[j]] = struct{}{}
	}

	for i := sj; i < len(p2.way); i++ {
		j := i
		for {
			if _, ok := seen[p2.way[j]]; !ok {
				break
			}
			j = (j + 1) % len(p2.way)

		}
		way[i] = p2.way[j]
		seen[p2.way[j]] = struct{}{}

	}

	res :=  instance{
		way: way,
		edges: p1.edges,
	}

	return res
}

func solveAnnealingGA(inst instance) []int {
	edges := inst.edges
	way := inst.way

	fit := len(estimate(edges, way))
	bestFit := fit
	best := way
	T := 1.0
	for T > 0.001 {
		i, j := rand.Intn(len(way)), rand.Intn(len(way))
		way[i], way[j] = way[j], way[i]
		newFit := len(estimate(edges, way))
		if newFit < fit && rand.Float64() > math.Exp(float64(newFit - fit) / T){
			way[i], way[j] = way[j], way[i]
		} else {
			fit = newFit
			if fit > bestFit {
				best = make([]int, len(way))
				copy(best, way)
				bestFit = fit
			}
		}

		T *= 0.999//9
	}

	return  best
}



func solve2opt(inst instance) {
	fit := inst.fitness()
	for i := 0;i < len(inst.way); i++ {
		for j := 0; j < len(inst.way); j++ {
			inst.way[i], inst.way[j] = inst.way[j], inst.way[i]
			inst.cachedFitness = 0
			if inst.fitness() < fit {
				inst.way[i], inst.way[j] = inst.way[j], inst.way[i]
			} else {
				fit = inst.fitness()
			}
		}
	}
}

func mutate(inst instance) instance {
	i, j := rand.Intn(len(inst.way)),rand.Intn(len(inst.way))
	way := make([]int, len(inst.way))
	copy(way, inst.way)

	way[i], way[j] = way[j], way[i]

	i = rand.Intn(len(inst.way))
	to := inst.edges[inst.way[i]].to

	for _, j := range  inst.way[1:] {
		if inst.edges[inst.way[j]].from == to {
			way[i - 1], way[j] = way[j], way[i - 1]
			break
		}

	}

	res := instance{
		way: way,
		edges: inst.edges,
	}
	//solve2opt(res)


	return res
}

func solveGenetic(edges []edge) []int {
	popSize := 100

	population := make([]instance, 0, popSize)
	for i := 0; i < popSize; i++ {
		inst := instance{
			edges: edges,
			way: rand.Perm(len(edges)),
		}
		solve2opt(inst)
		population = append(population, inst)

	}
	totalFitness := 0
	for _, inst := range population {
		totalFitness += inst.fitness()
	}

	best := population[0]
	bestFitness := population[0].fitness()
	//log.Printf("best fitness: %d", bestFitness)

	notChangedCount := 0
	for {

			newTotalFitness := 0
			newPopulation := make([]instance, 0, len(population))
			for i := 0; i < popSize; i++ {
				parent1 := pick(population, totalFitness)
				parent2 := pick(population, totalFitness)
				child := cross(parent1, parent2)
				if rand.Float64() < 0.08 {
					child = mutate(child)
				}
				//if rand.Float64() < 0.01 {
				//	child = instance{
				//		edges: edges,
				//		way: rand.Perm(len(edges)),
				//	}
				//	solve2opt(child)
				//}
				//if rand.Float64() < 0.01 {
				//	child = instance{
				//		edges: edges,
				//		way: rand.Perm(len(edges)),
				//	}
				//	child.way = solveAnnealingGA(child)
				//}

					newPopulation = append(newPopulation, child)
				newTotalFitness += child.fitness()

			}
			sort.SliceStable(population, func(i, j int) bool {
				return population[i].fitness() < population[j].fitness()
			})
			population = newPopulation
			totalFitness = newTotalFitness
			populationBest := population[0]
			populationBestFitness := populationBest.fitness()

			if populationBestFitness > bestFitness {
				bestFitness = populationBestFitness
				best = populationBest.copy()
				log.Printf("record updated: %d", bestFitness)
				//notChangedCount = 0
			} else {
				notChangedCount++
			}

			if notChangedCount > 10000 {
				//break
			}


	}

	return best.bestPath()
}

func estimate(edges []edge, way []int) []int {
	n := 1
	bestN := 1
	seen := make(map[int]struct{})
	current := edges[way[0]].from
	seen[current] = struct{}{}
	path := []int{current}
	bestPath := path

	for i := 1; i < len(way); i++ {
		if _, ok := seen[edges[way[i]].to]; !ok && edges[way[i]].from == current {
			n++
			seen[edges[way[i]].to] = struct{}{}
			path = append(path, edges[way[i]].to)

			if n > bestN {
				bestN = n
				bestPath = path
			}
			current = edges[way[i]].to
		} else {
			n = 1
			path = []int{current}

		}
	}
	return bestPath
}


func solve2optComplete(edges []edge) []int {
	way := rand.Perm(len(edges))

	fit := len(estimate(edges, way))
	for i := 0;i < len(way); i++ {
		for j := 0; j < len(way); j++ {
			way[i], way[j] = way[j], way[i]
			newFit := len(estimate(edges, way))
			if newFit < fit {
				way[i], way[j] = way[j], way[i]
			} else {
				fit = newFit
			}
		}
	}
	return estimate(edges, way)
}



func solveAnnealing(edges []edge) []int {
	way := rand.Perm(len(edges))
	fit := len(estimate(edges, way))
	bestFit := fit
	best := way
	T := 1.0
	for T > 0.001 {
		i, j := rand.Intn(len(way)), rand.Intn(len(way))
		way[i], way[j] = way[j], way[i]
		newFit := len(estimate(edges, way))
		if newFit < fit && rand.Float64() > math.Exp(float64(newFit - fit) / T){
			way[i], way[j] = way[j], way[i]
		} else {
			fit = newFit
			if fit > bestFit {
				best = make([]int, len(way))
				copy(best, way)
				bestFit = fit
			}
		}

		T *= 0.99999
	}

	return estimate(edges, way)
}


func main() {


	//best := solve2optComplete(testGraph(20, 50))
	g := testGraph(100, 200)
	//g := readGraphFromCli()

	//best := solveAnnealing(g)
	best := solveGenetic(g)
	fmt.Printf("%d\n", len(best))

	path := make([]string, 0, len(best))
	for _, p := range best {
		path = append(path, strconv.Itoa(p))

	}
	fmt.Print(strings.Join(path, " ") + "\n")


	best = solveAnnealing(g)
	fmt.Printf("%d\n", len(best))

	path = make([]string, 0, len(best))
	for _, p := range best {
		path = append(path, strconv.Itoa(p))

	}
	fmt.Print(strings.Join(path, " ") + "\n")
	//g := testGraph(25, 75)
	//solve(g)

}
func readGraphFromCli() []edge {
	var n, m int
	if _, err := fmt.Scanf("%d %d", &n, &m); err != nil {
		log.Fatalf("can'r read input: %s", err)
	}
	var from, to int
	edges := make([]edge, 0, m)
	for i := 0; i < m; i++ {
		if _, err := fmt.Scanf("%d %d", &from, &to); err != nil {
			log.Fatalf("can't read input: %s", err)
		}
		edges = append(edges, edge{
			from: from,
			to: to,
		})
		edges = append(edges, edge{
			from: to,
			to: from,
		})
	}
	return edges
}

func testGraph(n, m int) []edge {
	edges := make([]edge,0)
	way := rand.Perm(n)
	for i := 0; i < len(way) - 1; i++ {
		edges = append(edges, edge{
			from: way[i],
			to: way[i + 1],
		})
		edges = append(edges, edge{
			from: way[i + 1],
			to: way[i],
		})
	}

	for i := 0; i < m - n; i++ {
		edges = append(edges, edge{
			from: rand.Intn(n),
			to: rand.Intn(n),
		})
	}
	return  edges
}

