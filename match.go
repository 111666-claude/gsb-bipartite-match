// Package bipartitematch 二分图最大匹配与最小顶点覆盖，只用标准库。
package bipartitematch

import (
	"errors"
	"strings"
)

// Edge 是一条二分图的边，Left 与 Right 是两侧的顶点名。
type Edge struct {
	Left  string
	Right string
}

// ErrNoCover 表示最小顶点覆盖还没有实现。
var ErrNoCover = errors.New("bipartitematch: 最小顶点覆盖还没实现")

// ErrFormat 表示边表格式不对。
var ErrFormat = errors.New("bipartitematch: 边表格式不对")

// ParseEdges 解析 "l1-r1,l1-r2"。缺陷：不校验顶点名，报错也不给边的下标。
func ParseEdges(text string) ([]Edge, error) {
	edges := []Edge{}
	for _, item := range strings.Split(text, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		left, right, found := strings.Cut(item, "-")
		if !found {
			return nil, ErrFormat
		}
		edges = append(edges, Edge{Left: left, Right: right})
	}
	return edges, nil
}

// MaxMatching 返回一组匹配。
// 缺陷：按边的书写顺序给每个左侧顶点贪心挑第一条可用的边，不做增广。
func MaxMatching(edges []Edge) []Edge {
	usedRight := map[string]bool{}
	usedLeft := map[string]bool{}
	pairs := []Edge{}
	for _, edge := range edges {
		if usedLeft[edge.Left] || usedRight[edge.Right] {
			continue
		}
		usedLeft[edge.Left] = true
		usedRight[edge.Right] = true
		pairs = append(pairs, edge)
	}
	return pairs
}

// MinVertexCover 缺陷：还没实现。
func MinVertexCover(edges []Edge) ([]string, error) {
	return nil, ErrNoCover
}
