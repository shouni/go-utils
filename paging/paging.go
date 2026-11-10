// Package paging builds the metadata a listing page needs to render pagination.
//
// It lives here because a PageMeta of the same JSON shape is returned both by GCS-backed
// listings (go-job-kit) and by Firestore-backed listings (gcp-kit/jobstatus), and an M2M
// client receives both through one type. With two definitions the edge values (total_pages
// for an empty list, prev_page on page 1) evolve separately, and responses that should share
// one contract differ from service to service.
// Reading the listing itself is out of scope.
package paging

// PageMeta is the metadata a listing page needs to render pagination.
//
// The JSON tags match the existing responses each service returns. Both pages and M2M
// clients depend on them, so changing them requires consumers to follow.
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

// New builds the metadata for one page from the total item count.
//
// Page numbers start at 1. When perPage is 0 or less, nothing is paged: all items form one
// page and PerPage is set to total. An out-of-range page is clamped to the last page
// (jobs can be deleted while a listing is being browsed, and the page being viewed can vanish).
//
// Every value is clamped into range: TotalPages is 1 even for an empty list, and PrevPage
// and NextPage are between 1 and TotalPages. Templates are expected to switch on HasPrev /
// HasNext and use PrevPage / NextPage only as link targets. Decide whether to show a pager
// by whether TotalPages is greater than 1, not by whether it is 0.
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

// Offset is the number of items to skip before the start of this page.
func (m PageMeta) Offset() int {
	if m.PerPage <= 0 {
		return 0
	}
	return (m.Page - 1) * m.PerPage
}

// WithItemCount returns a copy with From / To adjusted to the number of items actually fetched.
//
// A listing orders the IDs first and then reads each item's metadata, so when some reads
// fail, fewer items are shown than New assumed. Use it to avoid saying "showing items 1-10"
// while only 8 are listed. Total is left unchanged.
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
