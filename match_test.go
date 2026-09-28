package bipartitematch

import "testing"

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
