// Package bipartitematch 二分图最大匹配与最小顶点覆盖（König 定理），只用标准库。
package bipartitematch

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// Edge 是一条二分图的边，Left 与 Right 是两侧的顶点名。
type Edge struct {
	Left  string
	Right string
}

// ErrFormat 表示边表格式不对。
var ErrFormat = errors.New("bipartitematch: 边表格式不对")

// ParseEdges 解析 "l1-r1,l1-r2" 这样的边表。
// 空项（如结尾多出的逗号）会被跳过；空顶点名、缺 "-" 或重复边都会报错，
// 错误信息里带 1 起始的边序号。
func ParseEdges(text string) ([]Edge, error) {
	edges := []Edge{}
	seen := map[Edge]int{}
	for i, item := range strings.Split(text, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		left, right, found := strings.Cut(item, "-")
		if !found {
			return nil, fmt.Errorf("%w：第 %d 条边 %q 缺少 -", ErrFormat, i+1, item)
		}
		left = strings.TrimSpace(left)
		right = strings.TrimSpace(right)
		if left == "" || right == "" {
			return nil, fmt.Errorf("%w：第 %d 条边 %q 顶点名不能为空", ErrFormat, i+1, item)
		}
		edge := Edge{Left: left, Right: right}
		if first, dup := seen[edge]; dup {
			return nil, fmt.Errorf("%w：第 %d 条边 %q 与第 %d 条重复", ErrFormat, i+1, item, first)
		}
		seen[edge] = i + 1
		edges = append(edges, edge)
	}
	return edges, nil
}

// graph 是二分图的紧凑表示：顶点按下标存储，邻接表升序，保证结果确定。
type graph struct {
	leftNames  []string
	rightNames []string
	adj        [][]int // 左侧下标 -> 右侧下标列表（升序、去重）
}

func buildGraph(edges []Edge) graph {
	leftSet := map[string]bool{}
	rightSet := map[string]bool{}
	for _, edge := range edges {
		leftSet[edge.Left] = true
		rightSet[edge.Right] = true
	}
	g := graph{
		leftNames:  make([]string, 0, len(leftSet)),
		rightNames: make([]string, 0, len(rightSet)),
	}
	for name := range leftSet {
		g.leftNames = append(g.leftNames, name)
	}
	for name := range rightSet {
		g.rightNames = append(g.rightNames, name)
	}
	sort.Strings(g.leftNames)
	sort.Strings(g.rightNames)
	leftIndex := make(map[string]int, len(g.leftNames))
	for i, name := range g.leftNames {
		leftIndex[name] = i
	}
	rightIndex := make(map[string]int, len(g.rightNames))
	for i, name := range g.rightNames {
		rightIndex[name] = i
	}
	adjSets := make([]map[int]bool, len(g.leftNames))
	for _, edge := range edges {
		u := leftIndex[edge.Left]
		if adjSets[u] == nil {
			adjSets[u] = map[int]bool{}
		}
		adjSets[u][rightIndex[edge.Right]] = true
	}
	g.adj = make([][]int, len(g.leftNames))
	for u, set := range adjSets {
		for v := range set {
			g.adj[u] = append(g.adj[u], v)
		}
		sort.Ints(g.adj[u])
	}
	return g
}

// hopcroftKarp 求最大匹配，时间 O(E√V)，内存 O(V+E)。
// 返回 matchL（左 -> 右）与 matchR（右 -> 左），未匹配为 -1。
func hopcroftKarp(g graph) (matchL, matchR []int) {
	nL, nR := len(g.leftNames), len(g.rightNames)
	matchL = make([]int, nL)
	matchR = make([]int, nR)
	for i := range matchL {
		matchL[i] = -1
	}
	for j := range matchR {
		matchR[j] = -1
	}
	dist := make([]int, nL)

	bfs := func() bool {
		queue := make([]int, 0, nL)
		found := false
		for u := 0; u < nL; u++ {
			if matchL[u] < 0 {
				dist[u] = 0
				queue = append(queue, u)
			} else {
				dist[u] = -1
			}
		}
		for head := 0; head < len(queue); head++ {
			u := queue[head]
			for _, v := range g.adj[u] {
				w := matchR[v]
				if w < 0 {
					found = true
				} else if dist[w] < 0 {
					dist[w] = dist[u] + 1
					queue = append(queue, w)
				}
			}
		}
		return found
	}

	var dfs func(u int) bool
	dfs = func(u int) bool {
		for _, v := range g.adj[u] {
			w := matchR[v]
			if w < 0 || (dist[w] == dist[u]+1 && dfs(w)) {
				matchL[u] = v
				matchR[v] = u
				return true
			}
		}
		dist[u] = -1
		return false
	}

	for bfs() {
		for u := 0; u < nL; u++ {
			if matchL[u] < 0 {
				dfs(u)
			}
		}
	}
	return matchL, matchR
}

// MaxMatching 用 Hopcroft-Karp 求最大匹配，返回的配对按字典序排列。
func MaxMatching(edges []Edge) []Edge {
	g := buildGraph(edges)
	matchL, _ := hopcroftKarp(g)
	pairs := []Edge{}
	for u, v := range matchL {
		if v >= 0 {
			pairs = append(pairs, Edge{Left: g.leftNames[u], Right: g.rightNames[v]})
		}
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].Left != pairs[j].Left {
			return pairs[i].Left < pairs[j].Left
		}
		return pairs[i].Right < pairs[j].Right
	})
	return pairs
}

// MinVertexCover 用 König 定理求最小顶点覆盖：从未匹配的左侧顶点出发，
// 沿交替路（左->右走非匹配边，右->左走匹配边）标记可达集 Z，
// 覆盖集为 (L \ Z) ∪ (R ∩ Z)。返回的顶点名按字典序排列。
func MinVertexCover(edges []Edge) []string {
	g := buildGraph(edges)
	matchL, matchR := hopcroftKarp(g)

	visL := make([]bool, len(g.leftNames))
	visR := make([]bool, len(g.rightNames))
	stack := []int{}
	for u := range matchL {
		if matchL[u] < 0 {
			visL[u] = true
			stack = append(stack, u)
		}
	}
	for len(stack) > 0 {
		u := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, v := range g.adj[u] {
			if matchL[u] == v || visR[v] {
				continue
			}
			visR[v] = true
			if w := matchR[v]; w >= 0 && !visL[w] {
				visL[w] = true
				stack = append(stack, w)
			}
		}
	}

	cover := []string{}
	for u := range visL {
		if !visL[u] {
			cover = append(cover, g.leftNames[u])
		}
	}
	for v := range visR {
		if visR[v] {
			cover = append(cover, g.rightNames[v])
		}
	}
	sort.Strings(cover)
	return cover
}
