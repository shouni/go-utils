// Package paging は、一覧画面がページネーションを描画するためのメタデータを組み立てます。
//
// 置き場所がここなのは、同じ JSON 形の PageMeta が GCS 系の一覧（go-job-kit）と
// Firestore 系の一覧（gcp-kit/jobstatus）の両方から返り、それを 1 つの型で受ける
// M2M クライアントがあるためです。定義が 2 つあると端の値（空一覧の total_pages、
// 1 ページ目の prev_page）が別々に育ち、同じ契約のはずの応答がサービスごとに変わります。
// 一覧の読み込みそのものは扱いません。
package paging

// PageMeta は、一覧画面がページネーションを描画するために必要なメタデータです。
//
// JSON タグは各サービスが返している既存のレスポンスと同じ形です。画面と M2M
// クライアントの双方が依存しているため、変更するときは利用側の追随が要ります。
type PageMeta struct {
	Page       int  `json:"page"`
	PerPage    int  `json:"per_page"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasPrev    bool `json:"has_prev"`
	HasNext    bool `json:"has_next"`
	PrevPage   int  `json:"prev_page"`
	NextPage   int  `json:"next_page"`
	From       int  `json:"from"`
	To         int  `json:"to"`
}

// New は、総件数から 1 ページ分のメタデータを組み立てます。
//
// ページ番号は 1 始まりです。perPage が 0 以下のときはページングせず全件を 1 ページと
// して扱い、PerPage には total を入れます。範囲外のページは最終ページへ丸めます
// （一覧の途中でジョブが削除され、見ていたページが消えることがあるためです）。
//
// 値はすべて範囲内に収めます。TotalPages は空一覧でも 1、PrevPage と NextPage は
// 1 以上 TotalPages 以下です。テンプレートは HasPrev / HasNext で表示を切り替え、
// PrevPage / NextPage はリンク先にだけ使う前提です。ページャを出すかどうかは
// TotalPages が 0 かどうかではなく、1 より大きいかどうかで判定してください。
func New(page, perPage, total int) PageMeta {
	if total < 0 {
		total = 0
	}
	if perPage <= 0 {
		perPage = total
	}

	totalPages := 1
	if perPage > 0 {
		totalPages = max((total+perPage-1)/perPage, 1)
	}
	page = min(max(page, 1), totalPages)

	from, to := 0, 0
	if total > 0 && perPage > 0 {
		from = (page-1)*perPage + 1
		to = min(from+perPage-1, total)
	}

	return PageMeta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
		HasPrev:    page > 1,
		HasNext:    page < totalPages,
		PrevPage:   max(page-1, 1),
		NextPage:   min(page+1, totalPages),
		From:       from,
		To:         to,
	}
}

// Offset は、このページの先頭までに読み飛ばす件数です。
func (m PageMeta) Offset() int {
	if m.PerPage <= 0 {
		return 0
	}
	return (m.Page - 1) * m.PerPage
}

// WithItemCount は、実際に取得できた件数に合わせて From / To を補正した複製を返します。
//
// 一覧は ID を並べてからメタデータ本体を読みにいくため、一部の読み込みに失敗すると
// 表示件数が New の想定より少なくなります。「1〜10 件目を表示」と出しながら 8 件しか
// 並ばない、というズレを防ぐために使います。Total は変えません。
func (m PageMeta) WithItemCount(itemCount int) PageMeta {
	if m.Total == 0 {
		return m
	}
	if itemCount <= 0 {
		m.From, m.To = 0, 0
		return m
	}
	if m.PerPage > 0 {
		m.To = m.From + itemCount - 1
		return m
	}
	m.To = itemCount
	return m
}
