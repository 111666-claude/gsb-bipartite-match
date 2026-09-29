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

// ParseEdges 解析 "l1-r1,l1-r2"。空顶点名、缺 "-" 与重复边都会报错，
// 错误信息里带 1 起始的边序号。
func ParseEdges(text string) ([]Edge, error) {
	edges := []Edge{}
	seen := map[Edge]int{}
	for _, item := range strings.Split(text, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		index := len(edges) + 1
		left, right, found := strings.Cut(item, "-")
		if !found {
			return nil, fmt.Errorf("%w：第 %d 条边 %q 缺少 \"-\"", ErrFormat, index, item)
		}
		left = strings.TrimSpace(left)
		right = strings.TrimSpace(right)
		if left == "" || right == "" {
			return nil, fmt.Errorf("%w：第 %d 条边 %q 的顶点名为空", ErrFormat, index, item)
		}
		edge := Edge{Left: left, Right: right}
		if first, dup := seen[edge]; dup {
			return nil, fmt.Errorf("%w：第 %d 条边 %q 与第 %d 条重复", ErrFormat, index, item, first)
		}
		seen[edge] = index
		edges = append(edges, edge)
	}
	return edges, nil
}

// graph 是二分图的邻接表表示，左右顶点名各自按字典序编号，
// 邻接点也按编号升序，保证算法结果确定。内存 O(V+E)。
type graph struct {
	lefts  []string
	rights []string
	adj    [][]int // 左侧下标 -> 右侧下标列表（升序）
}

func buildGraph(edges []Edge) graph {
	leftSet := map[string]bool{}
	rightSet := map[string]bool{}
	for _, edge := range edges {
		leftSet[edge.Left] = true
		rightSet[edge.Right] = true
	}
	g := graph{
		lefts:  make([]string, 0, len(leftSet)),
		rights: make([]string, 0, len(rightSet)),
	}
	for name := range leftSet {
		g.lefts = append(g.lefts, name)
	}
	for name := range rightSet {
		g.rights = append(g.rights, name)
	}
	sort.Strings(g.lefts)
	sort.Strings(g.rights)
	leftIndex := make(map[string]int, len(g.lefts))
	for i, name := range g.lefts {
		leftIndex[name] = i
	}
	rightIndex := make(map[string]int, len(g.rights))
	for i, name := range g.rights {
		rightIndex[name] = i
	}
	g.adj = make([][]int, len(g.lefts))
	for _, edge := range edges {
		u := leftIndex[edge.Left]
		g.adj[u] = append(g.adj[u], rightIndex[edge.Right])
	}
	for u := range g.adj {
		sort.Ints(g.adj[u])
	}
	return g
}

// hopcroftKarp 求最大匹配，复杂度 O(E√V)。
// 返回 pairLeft（左 -> 右，-1 表示未匹配）与 pairRight（右 -> 左）。
// 顶点与邻接表都有序，结果确定。
func hopcroftKarp(g graph) (pairLeft, pairRight []int) {
	n := len(g.lefts)
	pairLeft = make([]int, n)
	pairRight = make([]int, len(g.rights))
	for i := range pairLeft {
		pairLeft[i] = -1
	}
	for j := range pairRight {
		pairRight[j] = -1
	}
	dist := make([]int, n)

	bfs := func() bool {
		queue := make([]int, 0, n)
		reachableFreeRight := false
		for u := 0; u < n; u++ {
			if pairLeft[u] == -1 {
				dist[u] = 0
				queue = append(queue, u)
			} else {
				dist[u] = -1
			}
		}
		for head := 0; head < len(queue); head++ {
			u := queue[head]
			for _, v := range g.adj[u] {
				w := pairRight[v]
				if w == -1 {
					reachableFreeRight = true
				} else if dist[w] == -1 {
					dist[w] = dist[u] + 1
					queue = append(queue, w)
				}
			}
		}
		return reachableFreeRight
	}

	var dfs func(u int) bool
	dfs = func(u int) bool {
		for _, v := range g.adj[u] {
			w := pairRight[v]
			if w == -1 || (dist[w] == dist[u]+1 && dfs(w)) {
				pairLeft[u] = v
				pairRight[v] = u
				return true
			}
		}
		dist[u] = -1
		return false
	}

	for bfs() {
		for u := 0; u < n; u++ {
			if pairLeft[u] == -1 {
				dfs(u)
			}
		}
	}
	return pairLeft, pairRight
}

// MaxMatching 用 Hopcroft-Karp 增广路求最大匹配。
// 返回的配对按左侧顶点名字典序排列，同一输入结果确定。
func MaxMatching(edges []Edge) []Edge {
	g := buildGraph(edges)
	pairLeft, _ := hopcroftKarp(g)
	pairs := make([]Edge, 0)
	for u, v := range pairLeft {
		if v != -1 {
			pairs = append(pairs, Edge{Left: g.lefts[u], Right: g.rights[v]})
		}
	}
	return pairs
}

// MinVertexCover 用 König 定理求最小顶点覆盖：
// 从未匹配的左侧顶点出发，沿「非匹配边 左->右、匹配边 右->左」走交替路径，
// 覆盖 =（左侧未访问点）∪（右侧已访问点）。大小等于最大匹配。
// 返回的顶点名按字典序排列，同一输入结果确定。
func MinVertexCover(edges []Edge) []string {
	g := buildGraph(edges)
	pairLeft, pairRight := hopcroftKarp(g)

	visitedLeft := make([]bool, len(g.lefts))
	visitedRight := make([]bool, len(g.rights))
	stack := make([]int, 0, len(g.lefts))
	for u := range pairLeft {
		if pairLeft[u] == -1 {
			visitedLeft[u] = true
			stack = append(stack, u)
		}
	}
	for len(stack) > 0 {
		u := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, v := range g.adj[u] {
			if pairLeft[u] == v || visitedRight[v] {
				continue
			}
			visitedRight[v] = true
			if w := pairRight[v]; w != -1 && !visitedLeft[w] {
				visitedLeft[w] = true
				stack = append(stack, w)
			}
		}
	}

	cover := make([]string, 0)
	for u, name := range g.lefts {
		if !visitedLeft[u] {
			cover = append(cover, name)
		}
	}
	for v, name := range g.rights {
		if visitedRight[v] {
			cover = append(cover, name)
		}
	}
	sort.Strings(cover)
	return cover
}
