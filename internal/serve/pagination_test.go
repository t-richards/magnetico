package serve_test

import (
	"testing"

	"github.com/t-richards/magnetico/internal/serve"
)

func TestPaginationDisablesPrev(t *testing.T) {
	p := serve.Paginate(1, 100)
	if p.Prev != nil {
		t.Errorf("expected prev to be disabled")
	}
}

func TestPaginationDisablesNext(t *testing.T) {
	p := serve.Paginate(100, 100)
	if p.Next != nil {
		t.Errorf("expected next to be disabled")
	}
}

func TestPaginationDisablesPrevAndNext(t *testing.T) {
	p := serve.Paginate(1, 1)
	if p.Prev != nil {
		t.Errorf("expected prev to be disabled")
	}
	if p.Next != nil {
		t.Errorf("expected next to be disabled")
	}
}

func TestPaginationItems(t *testing.T) {
	cases := []struct {
		current int
		max     int
		items   []any
	}{
		// One page.
		{1, 1, []any{1}},

		// Three pages.
		{1, 3, []any{1, 2, 3}},
		{2, 3, []any{1, 2, 3}},
		{3, 3, []any{1, 2, 3}},

		// Five pages.
		{1, 5, []any{1, 2, 3, nil, 5}},
		{2, 5, []any{1, 2, 3, 4, 5}},
		{3, 5, []any{1, 2, 3, 4, 5}},
		{4, 5, []any{1, 2, 3, 4, 5}},
		{5, 5, []any{1, nil, 3, 4, 5}},

		// Seven pages.
		{1, 7, []any{1, 2, 3, nil, 7}},
		{2, 7, []any{1, 2, 3, 4, nil, 7}},
		{3, 7, []any{1, 2, 3, 4, 5, nil, 7}},
		{4, 7, []any{1, 2, 3, 4, 5, 6, 7}},
		{5, 7, []any{1, nil, 3, 4, 5, 6, 7}},
		{6, 7, []any{1, nil, 4, 5, 6, 7}},
		{7, 7, []any{1, nil, 5, 6, 7}},

		// Nine pages
		{1, 9, []any{1, 2, 3, nil, 9}},
		{2, 9, []any{1, 2, 3, 4, nil, 9}},
		{3, 9, []any{1, 2, 3, 4, 5, nil, 9}},
		{4, 9, []any{1, 2, 3, 4, 5, 6, nil, 9}},
		{5, 9, []any{1, nil, 3, 4, 5, 6, 7, nil, 9}},
		{6, 9, []any{1, nil, 4, 5, 6, 7, 8, 9}},
		{7, 9, []any{1, nil, 5, 6, 7, 8, 9}},
		{8, 9, []any{1, nil, 6, 7, 8, 9}},
		{9, 9, []any{1, nil, 7, 8, 9}},
	}

	for _, c := range cases {
		p := serve.Paginate(c.current, c.max)
		if len(p.Items) != len(c.items) {
			t.Errorf("expected %d items, got %d: %+v", len(c.items), len(p.Items), c)
		}
		for i := range p.Items {
			if p.Items[i] != c.items[i] {
				t.Errorf("expected %v, got %v: %+v", c.items[i], p.Items[i], c)
			}
		}
	}
}
