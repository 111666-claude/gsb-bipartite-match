package bipartitematch

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestParseEdges(t *testing.T) {
	edges, err := ParseEdges("l1-r1, l1-r2")
	if err != nil {
		t.Fatalf("ParseEdges 出错：%v", err)
	}
	if len(edges) != 2 || edges[0] != (Edge{Left: "l1", Right: "r1"}) || edges[1] != (Edge{Left: "l1", Right: "r2"}) {
		t.Fatalf("解析结果不对：%+v", edges)
	}
}

func TestParseEdgesRejectsMissingDash(t *testing.T) {
	if _, err := ParseEdges("l1r1"); err == nil {
		t.Fatal("缺 - 应该报错")
	}
}

func TestSingleEdge(t *testing.T) {
	edges, _ := ParseEdges("l1-r1")
	pairs := MaxMatching(edges)
	if len(pairs) != 1 || pairs[0] != (Edge{Left: "l1", Right: "r1"}) {
		t.Fatalf("单条边应给一个配对：%+v", pairs)
	}
}

func TestDisjointEdges(t *testing.T) {
	edges, _ := ParseEdges("l1-r1,l2-r2")
	if pairs := MaxMatching(edges); len(pairs) != 2 {
		t.Fatalf("两条不相交的边应给两个配对：%+v", pairs)
	}
}

func TestMatchingIsValid(t *testing.T) {
	edges, _ := ParseEdges("l1-r1,l1-r2,l2-r1,l2-r2")
	lefts := map[string]bool{}
	rights := map[string]bool{}
	for _, pair := range MaxMatching(edges) {
		if lefts[pair.Left] || rights[pair.Right] {
			t.Fatalf("配对里出现了重复顶点：%+v", pair)
		}
		lefts[pair.Left] = true
		rights[pair.Right] = true
	}
}

func TestMatchingSizeIsAtMostSide(t *testing.T) {
	edges, _ := ParseEdges("l1-r1,l1-r2,l1-r3")
	if pairs := MaxMatching(edges); len(pairs) > 1 {
		t.Fatalf("左侧只有一个顶点，匹配不能超过 1：%+v", pairs)
	}
}

func TestParseEdgesSkipsBlankItems(t *testing.T) {
	edges, err := ParseEdges("l1-r1,")
	if err != nil {
		t.Fatalf("ParseEdges 出错：%v", err)
	}
	if len(edges) != 1 {
		t.Fatalf("空项应该被跳过：%+v", edges)
	}
}

// 贪心按边的书写顺序只能给 matching=1，增广路必须给到 2。
func TestAugmentingPathBeatsGreedy(t *testing.T) {
	edges, err := ParseEdges("l1-r1,l1-r2,l2-r1")
	if err != nil {
		t.Fatalf("ParseEdges 出错：%v", err)
	}
	pairs := MaxMatching(edges)
	want := []Edge{{Left: "l1", Right: "r2"}, {Left: "l2", Right: "r1"}}
	if len(pairs) != len(want) {
		t.Fatalf("matching 应为 2，实际 %+v", pairs)
	}
	for i := range want {
		if pairs[i] != want[i] {
			t.Fatalf("pairs 应为 %+v，实际 %+v", want, pairs)
		}
	}
}

func TestMinVertexCoverSample(t *testing.T) {
	edges, _ := ParseEdges("l1-r1,l1-r2,l2-r1")
	cover := MinVertexCover(edges)
	if len(cover) != len(MaxMatching(edges)) {
		t.Fatalf("|cover| 必须等于 matching：cover=%v", cover)
	}
	assertCoversAllEdges(t, edges, cover)
	if !sort.StringsAreSorted(cover) {
		t.Fatalf("cover 应按字典序输出：%v", cover)
	}
}

func TestParseEdgesInvalidInput(t *testing.T) {
	cases := []struct {
		name  string
		text  string
		index int
	}{
		{"缺 -", "l1r1", 1},
		{"第二条缺 -", "l1-r1,l2r2", 2},
		{"空左顶点", "-r1", 1},
		{"空右顶点", "l1-", 1},
		{"第三条空顶点", "l1-r1,l2-r2,-r3", 3},
		{"重复边", "l1-r1,l1-r1", 2},
		{"隔开的重复边", "l1-r1,l2-r2,l1-r1", 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseEdges(tc.text)
			if err == nil {
				t.Fatalf("%q 应该报错", tc.text)
			}
			if !strings.Contains(err.Error(), fmt.Sprintf("第 %d 条", tc.index)) {
				t.Fatalf("错误应指出第 %d 条边，实际：%v", tc.index, err)
			}
		})
	}
}

func TestDeterministic(t *testing.T) {
	edges, _ := ParseEdges("l3-r2,l1-r3,l2-r1,l1-r2,l3-r1,l2-r3")
	firstPairs := MaxMatching(edges)
	firstCover := MinVertexCover(edges)
	for i := 0; i < 20; i++ {
		pairs := MaxMatching(edges)
		cover := MinVertexCover(edges)
		if fmt.Sprint(pairs) != fmt.Sprint(firstPairs) || fmt.Sprint(cover) != fmt.Sprint(firstCover) {
			t.Fatalf("重复运行结果不一致：%v / %v", pairs, cover)
		}
	}
}

func assertCoversAllEdges(t *testing.T, edges []Edge, cover []string) {
	t.Helper()
	inCover := map[string]bool{}
	for _, name := range cover {
		inCover[name] = true
	}
	for _, edge := range edges {
		if !inCover[edge.Left] && !inCover[edge.Right] {
			t.Fatalf("cover %v 没盖住边 %+v", cover, edge)
		}
	}
}

func assertValidMatching(t *testing.T, edges, pairs []Edge) {
	t.Helper()
	isEdge := map[Edge]bool{}
	lefts := map[string]bool{}
	rights := map[string]bool{}
	for _, edge := range edges {
		isEdge[edge] = true
		lefts[edge.Left] = true
		rights[edge.Right] = true
	}
	usedLeft := map[string]bool{}
	usedRight := map[string]bool{}
	for _, pair := range pairs {
		if !isEdge[pair] {
			t.Fatalf("配对 %+v 不是输入的边", pair)
		}
		if usedLeft[pair.Left] || usedRight[pair.Right] {
			t.Fatalf("配对里顶点重复：%+v", pair)
		}
		usedLeft[pair.Left] = true
		usedRight[pair.Right] = true
	}
	if len(pairs) > len(lefts) || len(pairs) > len(rights) {
		t.Fatalf("匹配大小 %d 超过某一侧顶点数", len(pairs))
	}
}

// 穷举所有匹配取最大（仅用于小规模对照）。
func bruteMaxMatching(edges []Edge) int {
	leftSet := map[string]bool{}
	adj := map[string][]string{}
	for _, edge := range edges {
		leftSet[edge.Left] = true
		adj[edge.Left] = append(adj[edge.Left], edge.Right)
	}
	lefts := make([]string, 0, len(leftSet))
	for name := range leftSet {
		lefts = append(lefts, name)
	}
	sort.Strings(lefts)
	usedRight := map[string]bool{}
	best := 0
	var search func(i, count int)
	search = func(i, count int) {
		if count+len(lefts)-i <= best {
			return
		}
		if i == len(lefts) {
			best = count
			return
		}
		search(i+1, count)
		for _, right := range adj[lefts[i]] {
			if !usedRight[right] {
				usedRight[right] = true
				search(i+1, count+1)
				usedRight[right] = false
			}
		}
	}
	search(0, 0)
	return best
}

// 枚举顶点子集取最小覆盖（仅用于小规模对照）。
func bruteMinVertexCover(edges []Edge) int {
	vertexSet := map[string]bool{}
	for _, edge := range edges {
		vertexSet[edge.Left] = true
		vertexSet[edge.Right] = true
	}
	vertices := make([]string, 0, len(vertexSet))
	for name := range vertexSet {
		vertices = append(vertices, name)
	}
	best := len(vertices) + 1
	for mask := 0; mask < 1<<len(vertices); mask++ {
		size := 0
		inCover := map[string]bool{}
		for i, name := range vertices {
			if mask&(1<<i) != 0 {
				size++
				inCover[name] = true
			}
		}
		if size >= best {
			continue
		}
		ok := true
		for _, edge := range edges {
			if !inCover[edge.Left] && !inCover[edge.Right] {
				ok = false
				break
			}
		}
		if ok {
			best = size
		}
	}
	return best
}

// 两侧各 ≤ 6 个顶点的随机图上，与穷举对照，差异必须为 0。
func TestAgainstBruteForce(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	for trial := 0; trial < 300; trial++ {
		numLeft := 1 + rng.Intn(6)
		numRight := 1 + rng.Intn(6)
		edges := []Edge{}
		for li := 0; li < numLeft; li++ {
			for ri := 0; ri < numRight; ri++ {
				if rng.Float64() < 0.4 {
					edges = append(edges, Edge{
						Left:  fmt.Sprintf("l%d", li),
						Right: fmt.Sprintf("r%d", ri),
					})
				}
			}
		}
		pairs := MaxMatching(edges)
		cover := MinVertexCover(edges)

		assertValidMatching(t, edges, pairs)
		assertCoversAllEdges(t, edges, cover)

		if got, want := len(pairs), bruteMaxMatching(edges); got != want {
			t.Fatalf("第 %d 组：匹配大小 %d，穷举是 %d，边表 %+v", trial, got, want, edges)
		}
		if got, want := len(cover), bruteMinVertexCover(edges); got != want {
			t.Fatalf("第 %d 组：cover 大小 %d，穷举是 %d，边表 %+v", trial, got, want, edges)
		}
		if len(cover) != len(pairs) {
			t.Fatalf("第 %d 组：|cover|=%d != matching=%d", trial, len(cover), len(pairs))
		}
	}
}

// 两侧各 1 万点、10 万条边要在 3 秒内出结果。
func TestLargeGraphWithinTimeLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("短模式跳过规模测试")
	}
	rng := rand.New(rand.NewSource(7))
	const side = 10000
	const numEdges = 100000
	edges := make([]Edge, 0, numEdges)
	seen := map[Edge]bool{}
	for len(edges) < numEdges {
		edge := Edge{
			Left:  fmt.Sprintf("l%d", rng.Intn(side)),
			Right: fmt.Sprintf("r%d", rng.Intn(side)),
		}
		if seen[edge] {
			continue
		}
		seen[edge] = true
		edges = append(edges, edge)
	}
	start := time.Now()
	pairs := MaxMatching(edges)
	cover := MinVertexCover(edges)
	elapsed := time.Since(start)
	assertValidMatching(t, edges, pairs)
	assertCoversAllEdges(t, edges, cover)
	if len(cover) != len(pairs) {
		t.Fatalf("|cover|=%d != matching=%d", len(cover), len(pairs))
	}
	if elapsed > 3*time.Second {
		t.Fatalf("耗时 %v 超过 3 秒", elapsed)
	}
	t.Logf("1 万 x 1 万点、10 万条边耗时 %v，matching=%d", elapsed, len(pairs))
}
