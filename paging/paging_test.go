package paging_test

import (
	"testing"

	"github.com/shouni/go-utils/paging"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name                 string
		page, perPage, total int
		want                 paging.PageMeta
	}{
		{
			name: "途中のページ", page: 2, perPage: 10, total: 25,
			want: paging.PageMeta{Page: 2, PerPage: 10, Total: 25, TotalPages: 3, HasPrev: true, HasNext: true, PrevPage: 1, NextPage: 3, From: 11, To: 20},
		},
		{
			name: "最終ページは To が総件数で止まる", page: 3, perPage: 10, total: 25,
			want: paging.PageMeta{Page: 3, PerPage: 10, Total: 25, TotalPages: 3, HasPrev: true, HasNext: false, PrevPage: 2, NextPage: 3, From: 21, To: 25},
		},
		{
			// 1 ページ目の prev_page が 0 になると、リンク先が存在しないページになる。
			name: "1 ページ目の PrevPage は 1", page: 1, perPage: 10, total: 25,
			want: paging.PageMeta{Page: 1, PerPage: 10, Total: 25, TotalPages: 3, HasPrev: false, HasNext: true, PrevPage: 1, NextPage: 2, From: 1, To: 10},
		},
		{
			// 空一覧でも TotalPages は 1。0 だと「1 / 0 ページ」と出る。
			name: "空一覧", page: 1, perPage: 10, total: 0,
			want: paging.PageMeta{Page: 1, PerPage: 10, Total: 0, TotalPages: 1, PrevPage: 1, NextPage: 1},
		},
		{
			// 空一覧に page>1 を渡しても HasPrev が立たない。
			name: "空一覧の範囲外ページは 1 へ丸める", page: 5, perPage: 10, total: 0,
			want: paging.PageMeta{Page: 1, PerPage: 10, Total: 0, TotalPages: 1, PrevPage: 1, NextPage: 1},
		},
		{
			name: "範囲外のページは最終ページへ丸める", page: 99, perPage: 10, total: 25,
			want: paging.PageMeta{Page: 3, PerPage: 10, Total: 25, TotalPages: 3, HasPrev: true, HasNext: false, PrevPage: 2, NextPage: 3, From: 21, To: 25},
		},
		{
			name: "0 以下のページは 1 へ丸める", page: 0, perPage: 10, total: 25,
			want: paging.PageMeta{Page: 1, PerPage: 10, Total: 25, TotalPages: 3, HasPrev: false, HasNext: true, PrevPage: 1, NextPage: 2, From: 1, To: 10},
		},
		{
			name: "perPage が 0 なら全件 1 ページ", page: 1, perPage: 0, total: 7,
			want: paging.PageMeta{Page: 1, PerPage: 7, Total: 7, TotalPages: 1, PrevPage: 1, NextPage: 1, From: 1, To: 7},
		},
		{
			name: "負の総件数は 0 として扱う", page: 1, perPage: 10, total: -3,
			want: paging.PageMeta{Page: 1, PerPage: 10, Total: 0, TotalPages: 1, PrevPage: 1, NextPage: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := paging.New(tt.page, tt.perPage, tt.total); got != tt.want {
				t.Errorf("New(%d, %d, %d) = %+v, want %+v", tt.page, tt.perPage, tt.total, got, tt.want)
			}
		})
	}
}

func TestOffset(t *testing.T) {
	if got := paging.New(3, 10, 25).Offset(); got != 20 {
		t.Errorf("Offset() = %d, want 20", got)
	}
	if got := paging.New(1, 0, 25).Offset(); got != 0 {
		t.Errorf("Offset() with perPage=0 = %d, want 0", got)
	}
}

func TestWithItemCount(t *testing.T) {
	meta := paging.New(1, 10, 25)

	if got := meta.WithItemCount(8); got.From != 1 || got.To != 8 || got.Total != 25 {
		t.Errorf("WithItemCount(8) = From %d To %d Total %d, want 1 8 25", got.From, got.To, got.Total)
	}
	if got := meta.WithItemCount(0); got.From != 0 || got.To != 0 {
		t.Errorf("WithItemCount(0) = From %d To %d, want 0 0", got.From, got.To)
	}
	if got := paging.New(1, 0, 7).WithItemCount(5); got.From != 1 || got.To != 5 {
		t.Errorf("WithItemCount(5) with perPage=0 = From %d To %d, want 1 5", got.From, got.To)
	}
	if got := paging.New(1, 10, 0).WithItemCount(3); got != paging.New(1, 10, 0) {
		t.Errorf("WithItemCount on an empty list must be a no-op: %+v", got)
	}
}
